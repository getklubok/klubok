package zapctx

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"go.uber.org/zap"
)

// Logging describes a context logger.
type Logging interface {
	Bare() *zap.Logger

	Named(s string) *Logger
	With(fields ...zap.Field) *Logger

	DebugWithContext(ctx context.Context, msg string, fields ...zap.Field)
	InfoWithContext(ctx context.Context, msg string, fields ...zap.Field)
	WarnWithContext(ctx context.Context, msg string, fields ...zap.Field)
	ErrorWithContext(ctx context.Context, msg string, fields ...zap.Field)
}

var _ Logging = (*Logger)(nil)

// Logger is a concrete context logger.
type Logger struct {
	logger *zap.Logger
}

// NewLogger returns a new context logger.
func NewLogger(log *zap.Logger) *Logger {
	return &Logger{
		logger: log.WithOptions(
			zap.WithCaller(false),
		),
	}
}

func (l *Logger) Bare() *zap.Logger {
	return l.logger
}

func (l *Logger) Named(s string) *Logger {
	if s == "" {
		return l
	}

	return NewLogger(l.logger.Named(s))
}

func (l *Logger) With(fields ...zap.Field) *Logger {
	if len(fields) == 0 {
		return l
	}

	return NewLogger(l.logger.With(fields...))
}

func (l *Logger) DebugWithContext(ctx context.Context, msg string, fields ...zap.Field) {
	Grasp(ctx, l.logger).Debug(msg, append(fields, getCallerField())...)
}

func (l *Logger) InfoWithContext(ctx context.Context, msg string, fields ...zap.Field) {
	Grasp(ctx, l.logger).Info(msg, append(fields, getCallerField())...)
}

func (l *Logger) WarnWithContext(ctx context.Context, msg string, fields ...zap.Field) {
	Grasp(ctx, l.logger).Warn(msg, append(fields, getCallerField())...)
}

func (l *Logger) ErrorWithContext(ctx context.Context, msg string, fields ...zap.Field) {
	Grasp(ctx, l.logger).Error(msg, append(fields, getCallerField())...)
}

func getCallerField() zap.Field {
	const callerSkipFrames = 2

	_, file, line, _ := runtime.Caller(callerSkipFrames)

	return zap.String("caller", fmt.Sprintf("%s:%d", getShortPath(file), line))
}

func getShortPath(file string) string {
	idx := strings.LastIndexByte(file, '/')
	if idx == -1 {
		return file
	}

	idx = strings.LastIndexByte(file[:idx], '/')
	if idx == -1 {
		return file
	}

	return file[idx+1:]
}
