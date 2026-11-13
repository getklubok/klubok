package handler

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"klubok/internal/restapi/oas"
)

type fakeDBHealthChecker struct {
	err    error
	called bool
}

func (f *fakeDBHealthChecker) Ping(context.Context) error {
	f.called = true

	return f.err
}

func TestHealth_HealthLive(t *testing.T) {
	db := &fakeDBHealthChecker{err: errors.New("db down")}
	handler := NewHealth(db)

	got, err := handler.HealthLive(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, &oas.HealthLiveNoContent{}, got)
	assert.False(t, db.called, "liveness не должен проверять БД")
}

func TestHealth_HealthReady(t *testing.T) {
	tests := []struct {
		name    string
		pingErr error
		want    oas.HealthReadyRes
	}{
		{
			name: "БД доступна",
			want: &oas.HealthReadyNoContent{},
		},
		{
			name:    "БД недоступна",
			pingErr: errors.New("db down"),
			want:    &oas.HealthReadyServiceUnavailable{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := &fakeDBHealthChecker{err: tt.pingErr}
			handler := NewHealth(db)

			got, err := handler.HealthReady(context.Background())

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
			assert.True(t, db.called)
		})
	}
}
