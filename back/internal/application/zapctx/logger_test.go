package zapctx

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestNewLogger(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	log := zap.New(core)

	ctxLog := NewLogger(log)

	assert.NotNil(t, ctxLog)
	assert.NotNil(t, ctxLog.Bare())
}

func TestLogger_Bare(t *testing.T) {
	tests := []struct {
		name     string
		setupLog func() *zap.Logger
		wantName string
	}{
		{
			name:     "returns underlying logger",
			setupLog: func() *zap.Logger { return zap.NewNop().Named("test") },
			wantName: "test",
		},
		{
			name:     "returns unnamed logger",
			setupLog: zap.NewNop,
			wantName: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctxLog := NewLogger(tt.setupLog())

			assert.Equal(t, tt.wantName, ctxLog.Bare().Name())
		})
	}
}

func TestLogger_Named(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantSame bool
		wantName string
	}{
		{
			name:     "returns named logger",
			input:    "handler",
			wantName: "handler",
		},
		{
			name:     "empty string returns same instance",
			input:    "",
			wantSame: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctxLog := NewLogger(zap.NewNop())
			result := ctxLog.Named(tt.input)

			if tt.wantSame {
				assert.Equal(t, ctxLog, result)
			} else {
				assert.Equal(t, tt.wantName, result.Bare().Name())
			}
		})
	}
}

func TestLogger_With(t *testing.T) {
	tests := []struct {
		name      string
		fields    []zap.Field
		wantSame  bool
		wantCheck func(t *testing.T, logs *observer.ObservedLogs)
	}{
		{
			name:   "returns logger with fields",
			fields: []zap.Field{zap.String("key", "value")},
			wantCheck: func(t *testing.T, logs *observer.ObservedLogs) {
				t.Helper()

				assert.Equal(t, 1, logs.Len())
				entry := logs.All()[0]
				assert.Len(t, entry.Context, 1)
				assert.Equal(t, "key", entry.Context[0].Key)
				assert.Equal(t, "value", entry.Context[0].String)
			},
		},
		{
			name:     "no fields returns same instance",
			fields:   []zap.Field{},
			wantSame: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, logs := observer.New(zapcore.DebugLevel)
			ctxLog := NewLogger(zap.New(core))

			result := ctxLog.With(tt.fields...)

			if tt.wantSame {
				assert.Equal(t, ctxLog, result)
			} else {
				result.Bare().Info("test")
				tt.wantCheck(t, logs)
			}
		})
	}
}

func TestLoggerWithContext(t *testing.T) {
	tests := []struct {
		name       string
		call       func(log Logging, ctx context.Context)
		wantLevel  zapcore.Level
		wantFields map[string]string
	}{
		{
			name: "InfoWithContext logs with context fields",
			call: func(log Logging, ctx context.Context) {
				log.InfoWithContext(ctx, "info msg", zap.String("extra", "field"))
			},
			wantLevel:  zapcore.InfoLevel,
			wantFields: map[string]string{"request_id": "req-1", "extra": "field"},
		},
		{
			name: "ErrorWithContext logs with context fields",
			call: func(log Logging, ctx context.Context) {
				log.ErrorWithContext(ctx, "error msg")
			},
			wantLevel:  zapcore.ErrorLevel,
			wantFields: map[string]string{"user_id": "u-1"},
		},
		{
			name: "WarnWithContext logs with context fields",
			call: func(log Logging, ctx context.Context) {
				log.WarnWithContext(ctx, "warn msg")
			},
			wantLevel:  zapcore.WarnLevel,
			wantFields: map[string]string{"key": "val"},
		},
		{
			name: "DebugWithContext logs with context fields",
			call: func(log Logging, ctx context.Context) {
				log.DebugWithContext(ctx, "debug msg")
			},
			wantLevel:  zapcore.DebugLevel,
			wantFields: map[string]string{"key": "val"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, logs := observer.New(zapcore.DebugLevel)
			ctxLog := NewLogger(zap.New(core))

			var ctx context.Context
			switch tt.wantLevel {
			case zapcore.InfoLevel:
				ctx = WithFields(context.Background(), zap.String("request_id", "req-1"))
			case zapcore.ErrorLevel:
				ctx = WithFields(context.Background(), zap.String("user_id", "u-1"))
			default:
				ctx = WithFields(context.Background(), zap.String("key", "val"))
			}

			tt.call(ctxLog, ctx)

			assert.Equal(t, 1, logs.Len())
			entry := logs.All()[0]
			assert.Equal(t, tt.wantLevel, entry.Level)

			got := make(map[string]string)
			for _, f := range entry.Context {
				got[f.Key] = f.String
			}

			for k, v := range tt.wantFields {
				assert.Equal(t, v, got[k])
			}

			assert.Contains(t, got, "caller")
		})
	}
}

func TestLoggerWithContext_EmptyContext(t *testing.T) {
	core, logs := observer.New(zapcore.DebugLevel)
	ctxLog := NewLogger(zap.New(core))

	ctxLog.InfoWithContext(context.Background(), "no context fields")

	assert.Equal(t, 1, logs.Len())

	got := make(map[string]string)
	for _, f := range logs.All()[0].Context {
		got[f.Key] = f.String
	}

	assert.Contains(t, got, "caller")
}
