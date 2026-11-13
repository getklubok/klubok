package metrics

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestMetrics_HTTPRequestsTotalInc(t *testing.T) {
	tests := []struct {
		name   string
		m      *Metrics
		method string
		path   string
		status string
	}{
		{
			name:   "Normal counter",
			m:      New(zap.NewNop()),
			method: http.MethodGet,
			path:   "/api/users",
			status: "200",
		},
		{
			name:   "Nil counter does not panic",
			m:      &Metrics{log: zap.NewNop()},
			method: http.MethodPost,
			path:   "/api/users",
			status: "201",
		},
		{
			name:   "Different methods and statuses",
			m:      New(zap.NewNop()),
			method: http.MethodDelete,
			path:   "/api/users/:id",
			status: "204",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				tt.m.HTTPRequestsTotalInc(tt.method, tt.path, tt.status)
			})
		})
	}
}

func TestMetrics_HTTPRequestsCurrentInc(t *testing.T) {
	tests := []struct {
		name string
		m    *Metrics
	}{
		{
			name: "Normal gauge",
			m:    New(zap.NewNop()),
		},
		{
			name: "Nil gauge does not panic",
			m:    &Metrics{log: zap.NewNop()},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				tt.m.HTTPRequestsCurrentInc()
			})
		})
	}
}

func TestMetrics_HTTPRequestsCurrentDec(t *testing.T) {
	tests := []struct {
		name string
		m    *Metrics
	}{
		{
			name: "Normal gauge",
			m:    New(zap.NewNop()),
		},
		{
			name: "Nil gauge does not panic",
			m:    &Metrics{log: zap.NewNop()},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				tt.m.HTTPRequestsCurrentDec()
			})
		})
	}
}

func TestMetrics_HTTPRequestDurationSecondsObserve(t *testing.T) {
	tests := []struct {
		name    string
		m       *Metrics
		method  string
		path    string
		elapsed float64
	}{
		{
			name:    "Normal histogram",
			m:       New(zap.NewNop()),
			method:  http.MethodGet,
			path:    "/api/users",
			elapsed: 0.123,
		},
		{
			name:    "Nil histogram does not panic",
			m:       &Metrics{log: zap.NewNop()},
			method:  http.MethodGet,
			path:    "/api/users",
			elapsed: 0.5,
		},
		{
			name:    "Zero duration",
			m:       New(zap.NewNop()),
			method:  http.MethodPost,
			path:    "/api/data",
			elapsed: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				tt.m.HTTPRequestDurationSecondsObserve(tt.method, tt.path, tt.elapsed)
			})
		})
	}
}

func TestMetrics_HTTPRequestsCurrentIncDec(t *testing.T) {
	m := New(zap.NewNop())

	assert.NotPanics(t, func() {
		m.HTTPRequestsCurrentInc()
		m.HTTPRequestsCurrentInc()
		m.HTTPRequestsCurrentDec()
	})
}
