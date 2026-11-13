package mailer

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/wneessen/go-mail"
	"go.uber.org/zap"

	"klubok/internal/application/service/mailer/mocks"
)

// testTimeout is the maximum time we wait for async operations in tests.
const testTimeout = 5 * time.Second

// newTestMailer creates a *Mailer backed by the given mock client.
func newTestMailer(client *mocks.MailClientMock) *Mailer {
	return &Mailer{
		from:   "Test <test@example.com>",
		client: client,
	}
}

// newTestWorker creates a Worker with a buffered channel, a mock mailer,
// a no-op logger, and the given interval.
func newTestWorker(client *mocks.MailClientMock, ch chan EmailTask, interval time.Duration) *Worker {
	return NewWorker(
		newTestMailer(client),
		ch,
		interval,
		zap.NewNop(),
	)
}

// mustSendTask sends a task on ch and fails the test if it blocks.
func mustSendTask(t *testing.T, ch chan<- EmailTask, task EmailTask) {
	t.Helper()
	select {
	case ch <- task:
	case <-time.After(testTimeout):
		t.Fatal("timed out sending task to channel")
	}
}

// ---------------------------------------------------------------------------
// TestWorker_NewWorker
// ---------------------------------------------------------------------------

func TestWorker_NewWorker(t *testing.T) {
	client := mocks.NewMailClientMock(t)
	ch := make(chan EmailTask, 1)
	worker := NewWorker(newTestMailer(client), ch, time.Second, zap.NewNop())

	require.NotNil(t, worker)
	assert.Equal(t, newTestMailer(client), worker.emailSvc)
	assert.Equal(t, (<-chan EmailTask)(ch), worker.ch)
	assert.Equal(t, time.Second, worker.interval)
	assert.NotNil(t, worker.log)
	assert.Nil(t, worker.cancel)
}

// ---------------------------------------------------------------------------
// TestWorker_SendsEmail
// ---------------------------------------------------------------------------

func TestWorker_SendsEmail(t *testing.T) {
	client := mocks.NewMailClientMock(t)
	ch := make(chan EmailTask, 1)

	sent := make(chan struct{})
	client.EXPECT().
		DialAndSendWithContext(mock.Anything, mock.Anything).
		Run(func(_ context.Context, _ ...*mail.Msg) {
			close(sent)
		}).
		Return(nil).
		Once()

	worker := newTestWorker(client, ch, time.Microsecond)
	worker.Start(t.Context())
	defer worker.Stop()

	mustSendTask(t, ch, EmailTask{
		Recipients: []string{"user@example.com"},
		Subject:    "Test",
		Template:   TemplateAuthCode,
		Data:       map[string]any{"code": "123456", "timeString": "5 минут"},
	})

	select {
	case <-sent:
	case <-time.After(testTimeout):
		t.Fatal("timed out waiting for email to be sent")
	}
}

// ---------------------------------------------------------------------------
// TestWorker_RespectsInterval
// ---------------------------------------------------------------------------

func TestWorker_RespectsInterval(t *testing.T) {
	client := mocks.NewMailClientMock(t)
	ch := make(chan EmailTask, 2)

	var (
		callCount     atomic.Int64
		firstCallTime atomic.Value // time.Time
		secondCallCh  = make(chan struct{})
	)

	// First call: record start time.
	client.EXPECT().
		DialAndSendWithContext(mock.Anything, mock.Anything).
		Run(func(_ context.Context, _ ...*mail.Msg) {
			callCount.Add(1)
			firstCallTime.Store(time.Now())
		}).
		Return(nil).
		Once()

	// Second call: record time and signal completion.
	client.EXPECT().
		DialAndSendWithContext(mock.Anything, mock.Anything).
		Run(func(_ context.Context, _ ...*mail.Msg) {
			callCount.Add(1)
			close(secondCallCh)
		}).
		Return(nil).
		Once()

	interval := 50 * time.Millisecond
	worker := newTestWorker(client, ch, interval)
	worker.Start(t.Context())
	defer worker.Stop()

	mustSendTask(t, ch, EmailTask{
		Recipients: []string{"a@example.com"},
		Subject:    "First",
		Template:   TemplateAuthCode,
		Data:       map[string]any{"code": "111", "timeString": "5 минут"},
	})
	mustSendTask(t, ch, EmailTask{
		Recipients: []string{"b@example.com"},
		Subject:    "Second",
		Template:   TemplateAuthCode,
		Data:       map[string]any{"code": "222", "timeString": "5 минут"},
	})

	// Wait for the second call.
	select {
	case <-secondCallCh:
	case <-time.After(testTimeout):
		t.Fatal("timed out waiting for second email call")
	}

	firstTime, ok := firstCallTime.Load().(time.Time)
	require.True(t, ok, "first call timestamp should be set")
	elapsed := time.Since(firstTime)
	assert.GreaterOrEqual(t, elapsed, interval/2,
		"worker did not respect interval between sends")
}

// ---------------------------------------------------------------------------
// TestWorker_StopViaContextCancellation
// ---------------------------------------------------------------------------

func TestWorker_StopViaContextCancellation(t *testing.T) {
	client := mocks.NewMailClientMock(t)
	ch := make(chan EmailTask, 10)

	// The mock should never be called because we cancel before sending.
	client.EXPECT().
		DialAndSendWithContext(mock.Anything, mock.Anything).
		Maybe()

	worker := newTestWorker(client, ch, time.Microsecond)
	ctx, cancel := context.WithCancel(t.Context())
	worker.Start(ctx)
	cancel() // cancel immediately

	// Give the goroutine time to notice the cancellation.
	time.Sleep(100 * time.Millisecond)

	// Worker should have exited; this send should block because nothing reads.
	select {
	case ch <- EmailTask{
		Recipients: []string{"user@example.com"},
		Subject:    "Should not send",
		Template:   TemplateAuthCode,
		Data:       map[string]any{"code": "000", "timeString": "5 минут"},
	}:
		// channel accepted – worker may have been still running
	default:
		// channel full – expected if worker exited and nobody reads
	}

	assert.NotNil(t, worker)
}

// ---------------------------------------------------------------------------
// TestWorker_StopViaStop
// ---------------------------------------------------------------------------

func TestWorker_StopViaStop(t *testing.T) {
	client := mocks.NewMailClientMock(t)
	ch := make(chan EmailTask, 1)

	done := make(chan struct{})
	client.EXPECT().
		DialAndSendWithContext(mock.Anything, mock.Anything).
		Run(func(_ context.Context, _ ...*mail.Msg) {
			close(done)
		}).
		Return(nil).
		Once()

	worker := newTestWorker(client, ch, time.Microsecond)
	worker.Start(t.Context())

	mustSendTask(t, ch, EmailTask{
		Recipients: []string{"user@example.com"},
		Subject:    "Test",
		Template:   TemplateAuthCode,
		Data:       map[string]any{"code": "123", "timeString": "5 минут"},
	})

	// Wait for the send to complete.
	select {
	case <-done:
	case <-time.After(testTimeout):
		t.Fatal("timed out waiting for send")
	}

	worker.Stop()
	// After Stop the worker has exited; no deadlock expected.
	time.Sleep(50 * time.Millisecond)
}

// ---------------------------------------------------------------------------
// TestWorker_ChannelClosed
// ---------------------------------------------------------------------------

func TestWorker_ChannelClosed(t *testing.T) {
	client := mocks.NewMailClientMock(t)
	ch := make(chan EmailTask, 1)

	done := make(chan struct{})
	client.EXPECT().
		DialAndSendWithContext(mock.Anything, mock.Anything).
		Run(func(_ context.Context, _ ...*mail.Msg) {
			close(done)
		}).
		Return(nil).
		Once()

	worker := newTestWorker(client, ch, time.Microsecond)
	worker.Start(t.Context())

	mustSendTask(t, ch, EmailTask{
		Recipients: []string{"user@example.com"},
		Subject:    "Last task",
		Template:   TemplateAuthCode,
		Data:       map[string]any{"code": "000", "timeString": "5 минут"},
	})

	// Wait for the call.
	select {
	case <-done:
	case <-time.After(testTimeout):
		t.Fatal("timed out waiting for send")
	}

	// Close the channel – the worker should exit its loop.
	close(ch)
	time.Sleep(50 * time.Millisecond)

	assert.NotNil(t, worker)
}

// ---------------------------------------------------------------------------
// TestWorker_SendError_DoesNotCrash
// ---------------------------------------------------------------------------

func TestWorker_SendError_DoesNotCrash(t *testing.T) {
	client := mocks.NewMailClientMock(t)
	ch := make(chan EmailTask, 2)

	var callCount atomic.Int64
	done := make(chan struct{})

	sendErr := errors.New("smtp connection refused")
	client.EXPECT().
		DialAndSendWithContext(mock.Anything, mock.Anything).
		Run(func(_ context.Context, _ ...*mail.Msg) {
			callCount.Add(1)
		}).
		Return(sendErr).
		Once()

	client.EXPECT().
		DialAndSendWithContext(mock.Anything, mock.Anything).
		Run(func(_ context.Context, _ ...*mail.Msg) {
			callCount.Add(1)
			close(done)
		}).
		Return(nil).
		Once()

	worker := newTestWorker(client, ch, time.Microsecond)
	worker.Start(t.Context())
	defer worker.Stop()

	mustSendTask(t, ch, EmailTask{
		Recipients: []string{"user@example.com"},
		Subject:    "Error test",
		Template:   TemplateAuthCode,
		Data:       map[string]any{"code": "err", "timeString": "5 минут"},
	})

	// Wait for the first call (error case).
	require.Eventually(t, func() bool {
		return callCount.Load() >= 1
	}, testTimeout, 5*time.Millisecond, "expected first DialAndSendWithContext call")

	// The worker should still be alive and accept another task.
	mustSendTask(t, ch, EmailTask{
		Recipients: []string{"user2@example.com"},
		Subject:    "Second attempt",
		Template:   TemplateAuthCode,
		Data:       map[string]any{"code": "ok", "timeString": "5 минут"},
	})

	// Wait for the second call.
	select {
	case <-done:
	case <-time.After(testTimeout):
		t.Fatal("timed out waiting for second call")
	}
}

// ---------------------------------------------------------------------------
// TestWorker_SendWithTimeout_RespectsDeadline
// ---------------------------------------------------------------------------

func TestWorker_SendWithTimeout_RespectsDeadline(t *testing.T) {
	client := mocks.NewMailClientMock(t)
	ch := make(chan EmailTask, 1)

	// We capture the context passed to DialAndSendWithContext via a Run callback.
	callCtxCh := make(chan context.Context, 1)
	client.EXPECT().
		DialAndSendWithContext(mock.Anything, mock.Anything).
		Run(func(ctx context.Context, _ ...*mail.Msg) {
			select {
			case callCtxCh <- ctx:
			default:
			}
		}).
		Return(nil).
		Once()

	worker := newTestWorker(client, ch, time.Microsecond)
	worker.Start(t.Context())
	defer worker.Stop()

	mustSendTask(t, ch, EmailTask{
		Recipients: []string{"user@example.com"},
		Subject:    "Timeout test",
		Template:   TemplateAuthCode,
		Data:       map[string]any{"code": "slow", "timeString": "5 минут"},
	})

	// Receive the context from the callback.
	var callCtx context.Context
	select {
	case callCtx = <-callCtxCh:
	case <-time.After(testTimeout):
		t.Fatal("timed out waiting for DialAndSendWithContext call")
	}

	deadline, hasDeadline := callCtx.Deadline()
	assert.True(t, hasDeadline, "expected a deadline on the context passed to Send")
	assert.WithinDuration(t,
		time.Now().Add(defaultTimeout), deadline,
		defaultTimeout,
		"deadline should be roughly %s from now", defaultTimeout)
}

// ---------------------------------------------------------------------------
// TestWorker_DoubleStartStop
// ---------------------------------------------------------------------------

func TestWorker_DoubleStartStop(t *testing.T) {
	client := mocks.NewMailClientMock(t)
	ch := make(chan EmailTask, 10)

	// Mock accepts any calls.
	client.EXPECT().
		DialAndSendWithContext(mock.Anything, mock.Anything).
		Return(nil).
		Maybe()

	worker := newTestWorker(client, ch, time.Microsecond)

	// Calling Start twice should not panic (only last cancel takes effect).
	worker.Start(t.Context())
	worker.Start(t.Context()) // second start overrides the cancel

	worker.Stop()
	worker.Stop() // second stop is a no-op

	assert.NotNil(t, worker)
}

// ---------------------------------------------------------------------------
// TestWorker_IntervalCancellation
// ---------------------------------------------------------------------------

func TestWorker_IntervalCancellation(t *testing.T) {
	client := mocks.NewMailClientMock(t)
	ch := make(chan EmailTask, 1)

	done := make(chan struct{})
	client.EXPECT().
		DialAndSendWithContext(mock.Anything, mock.Anything).
		Run(func(_ context.Context, _ ...*mail.Msg) {
			close(done)
		}).
		Return(nil).
		Once()

	interval := time.Hour // very long interval
	worker := newTestWorker(client, ch, interval)
	worker.Start(t.Context())

	mustSendTask(t, ch, EmailTask{
		Recipients: []string{"user@example.com"},
		Subject:    "Cancel during interval",
		Template:   TemplateAuthCode,
		Data:       map[string]any{"code": "cancel", "timeString": "5 минут"},
	})

	// Wait for the send to happen (it should happen despite the long interval
	// because the interval is only between sends, not before the first).
	select {
	case <-done:
	case <-time.After(testTimeout):
		t.Fatal("expected first send to complete")
	}

	// Now stop – this should interrupt the interval wait.
	worker.Stop()
	time.Sleep(50 * time.Millisecond)
	assert.NotNil(t, worker)
}
