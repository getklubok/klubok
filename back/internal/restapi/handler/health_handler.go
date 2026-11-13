package handler

import (
	"context"
	"time"

	"klubok/internal/restapi/oas"
)

// healthReadyTimeout ограничивает ping БД, чтобы проба укладывалась в таймаут куба.
const healthReadyTimeout = 2 * time.Second

type Health struct {
	db DBHealthChecker
}

func NewHealth(db DBHealthChecker) *Health {
	return &Health{db: db}
}

// Check liveness.
//
// GET /health/live.
func (*Health) HealthLive(_ context.Context) (oas.HealthLiveRes, error) {
	return &oas.HealthLiveNoContent{}, nil
}

// Check readiness.
//
// GET /health/ready.
func (h *Health) HealthReady(ctx context.Context) (oas.HealthReadyRes, error) {
	ctx, cancel := context.WithTimeout(ctx, healthReadyTimeout)
	defer cancel()

	if err := h.db.Ping(ctx); err != nil {
		return &oas.HealthReadyServiceUnavailable{}, nil
	}

	return &oas.HealthReadyNoContent{}, nil
}
