package zapctx

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestGetFields(t *testing.T) {
	tests := []struct {
		name     string
		setupCtx func() context.Context
		wantNil  bool
		wantLen  int
	}{
		{
			name:     "empty context returns nil",
			setupCtx: context.Background,
			wantNil:  true,
		},
		{
			name: "context with fields returns slice",
			setupCtx: func() context.Context {
				return WithFields(context.Background(), zap.String("key", "val"))
			},
			wantLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fields := GetFields(tt.setupCtx())

			if tt.wantNil {
				assert.Nil(t, fields)
			} else {
				assert.Len(t, fields, tt.wantLen)
			}
		})
	}
}

func TestWithFields(t *testing.T) {
	tests := []struct {
		name      string
		setupCtx  func() context.Context
		wantLen   int
		wantCheck func(t *testing.T, fields []zap.Field)
	}{
		{
			name: "add single field",
			setupCtx: func() context.Context {
				return WithFields(context.Background(), zap.String("key1", "value1"))
			},
			wantLen: 1,
			wantCheck: func(t *testing.T, fields []zap.Field) {
				t.Helper()

				assert.Equal(t, "key1", fields[0].Key)
				assert.Equal(t, "value1", fields[0].String)
			},
		},
		{
			name: "add multiple fields",
			setupCtx: func() context.Context {
				return WithFields(context.Background(),
					zap.String("key1", "value1"),
					zap.Int("key2", 42),
				)
			},
			wantLen: 2,
			wantCheck: func(t *testing.T, fields []zap.Field) {
				t.Helper()

				assert.Equal(t, "key1", fields[0].Key)
				assert.Equal(t, "value1", fields[0].String)
				assert.Equal(t, "key2", fields[1].Key)
				assert.Equal(t, int64(42), fields[1].Integer)
			},
		},
		{
			name: "duplicate key overwrites value",
			setupCtx: func() context.Context {
				ctx := WithFields(context.Background(), zap.String("key", "old"))

				return WithFields(ctx, zap.String("key", "new"))
			},
			wantLen: 1,
			wantCheck: func(t *testing.T, fields []zap.Field) {
				t.Helper()

				assert.Equal(t, "key", fields[0].Key)
				assert.Equal(t, "new", fields[0].String)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fields := GetFields(tt.setupCtx())

			assert.Len(t, fields, tt.wantLen)
			tt.wantCheck(t, fields)
		})
	}
}

func TestWithFields_Immutable(t *testing.T) {
	ctx1 := context.Background()
	ctx2 := WithFields(ctx1, zap.String("key1", "value1"))
	ctx3 := WithFields(ctx2, zap.String("key2", "value2"))

	assert.Nil(t, GetFields(ctx1))
	assert.Len(t, GetFields(ctx2), 1)
	assert.Len(t, GetFields(ctx3), 2)
}

func TestGrasp(t *testing.T) {
	tests := []struct {
		name        string
		setupCtx    func() context.Context
		wantFields  map[string]string
		wantNoExtra bool
	}{
		{
			name: "merges context fields into logger",
			setupCtx: func() context.Context {
				return WithFields(context.Background(), zap.String("request_id", "abc-123"))
			},
			wantFields: map[string]string{"request_id": "abc-123"},
		},
		{
			name:        "empty context adds no fields",
			setupCtx:    context.Background,
			wantNoExtra: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, logs := observer.New(zapcore.DebugLevel)
			log := zap.New(core)

			Grasp(tt.setupCtx(), log).Info("test")

			assert.Equal(t, 1, logs.Len())
			entry := logs.All()[0]

			if tt.wantNoExtra {
				assert.Empty(t, entry.Context)

				return
			}

			got := make(map[string]string)
			for _, f := range entry.Context {
				got[f.Key] = f.String
			}

			for k, v := range tt.wantFields {
				assert.Equal(t, v, got[k])
			}
		})
	}
}
