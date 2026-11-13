package zapctx

import (
	"context"
	"maps"

	"go.uber.org/zap"
)

type contextKeyFields struct{}

type contextFields struct {
	fields []zap.Field
	keys   map[string]int
}

// WithFields returns a context with fields stored into it.
func WithFields(ctx context.Context, fields ...zap.Field) context.Context {
	ecf, ok := ctx.Value(contextKeyFields{}).(contextFields)
	if !ok {
		ecf = contextFields{}
	}

	cf := contextFields{
		fields: make([]zap.Field, len(ecf.fields)),
		keys:   make(map[string]int),
	}

	copy(cf.fields, ecf.fields)
	maps.Copy(cf.keys, ecf.keys)

	for _, field := range fields {
		if idx, exists := cf.keys[field.Key]; exists {
			cf.fields[idx] = field
			continue
		}

		cf.fields = append(cf.fields, field)
		cf.keys[field.Key] = len(cf.fields) - 1
	}

	return context.WithValue(ctx, contextKeyFields{}, cf)
}

// GetFields returns fields stored in the context.
func GetFields(ctx context.Context) []zap.Field {
	cf, ok := ctx.Value(contextKeyFields{}).(contextFields)
	if !ok {
		return nil
	}

	return cf.fields
}

// Grasp makes the logger understand the context. It extracts zap fields from
// the context, if any, and returns a logger with that fields.
func Grasp(ctx context.Context, log *zap.Logger) *zap.Logger {
	return log.With(GetFields(ctx)...)
}
