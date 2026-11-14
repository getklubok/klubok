package metrics

import (
	"net/http"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			m:      New(zap.NewNop(), nil),
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
			m:      New(zap.NewNop(), nil),
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
			m:    New(zap.NewNop(), nil),
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
			m:    New(zap.NewNop(), nil),
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
			m:       New(zap.NewNop(), nil),
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
			m:       New(zap.NewNop(), nil),
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

func TestMetrics_EmailChannelSize(t *testing.T) {
	tests := []struct {
		name string
		size func() int
		want float64
	}{
		{
			name: "Empty channel",
			size: func() int { return 0 },
			want: 0,
		},
		{
			name: "Non-empty channel",
			size: func() int { return 42 },
			want: 42,
		},
		{
			name: "Nil size func",
			size: nil,
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New(zap.NewNop(), tt.size)

			reg := prometheus.NewRegistry()
			reg.MustRegister(m.EmailChannelSize)

			mfs, err := reg.Gather()
			require.NoError(t, err)
			require.Len(t, mfs, 1)
			require.Len(t, mfs[0].GetMetric(), 1)
			assert.Equal(t, "klubok_email_channel_size", mfs[0].GetName())
			assert.InDelta(t, tt.want, mfs[0].GetMetric()[0].GetGauge().GetValue(), 0)
		})
	}
}

func TestMetrics_HTTPRequestsCurrentIncDec(t *testing.T) {
	m := New(zap.NewNop(), nil)

	assert.NotPanics(t, func() {
		m.HTTPRequestsCurrentInc()
		m.HTTPRequestsCurrentInc()
		m.HTTPRequestsCurrentDec()
	})
}
