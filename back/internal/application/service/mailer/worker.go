package mailer

import (
	"context"
	"time"

	"go.uber.org/zap"
)

const defaultTimeout = 30 * time.Second

// EmailTask represents an email to be sent asynchronously.
type EmailTask struct {
	Recipients []string
	Subject    string
	Template   Template
	Data       map[string]any
}

// Worker sends emails from a channel with a configured interval between sends.
type Worker struct {
	emailSvc *Mailer
	ch       <-chan EmailTask
	interval time.Duration
	log      *zap.Logger
	cancel   context.CancelFunc
}

// NewWorker creates a new email worker.
func NewWorker(emailSvc *Mailer, ch <-chan EmailTask, interval time.Duration, log *zap.Logger) *Worker {
	return &Worker{
		emailSvc: emailSvc,
		ch:       ch,
		interval: interval,
		log:      log.Named("email_worker"),
	}
}

// Start launches the worker goroutine. Call Stop to shut it down.
func (w *Worker) Start(ctx context.Context) {
	ctx, w.cancel = context.WithCancel(ctx)
	go w.run(ctx)
}

// Stop signals the worker to shut down and waits for completion.
func (w *Worker) Stop() {
	if w.cancel != nil {
		w.cancel()
	}
}

func (w *Worker) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case task, ok := <-w.ch:
			if !ok {
				return
			}

			w.sendWithTimeout(ctx, task)

			// Wait interval before next send, unless context is cancelled.
			select {
			case <-ctx.Done():
				return
			case <-time.After(w.interval):
			}
		}
	}
}

func (w *Worker) sendWithTimeout(ctx context.Context, task EmailTask) {
	sendCtx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	if err := w.emailSvc.Send(sendCtx, task.Recipients, task.Subject, task.Template, task.Data); err != nil {
		w.log.Error("Failed to send email",
			zap.Strings("recipients", task.Recipients),
			zap.String("subject", task.Subject),
			zap.Error(err),
		)
	}
}
