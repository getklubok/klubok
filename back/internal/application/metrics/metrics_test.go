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
			m:      New(zap.NewNop(), nil, nil),
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
			m:      New(zap.NewNop(), nil, nil),
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
			m:    New(zap.NewNop(), nil, nil),
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
			m:    New(zap.NewNop(), nil, nil),
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
			m:       New(zap.NewNop(), nil, nil),
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
			m:       New(zap.NewNop(), nil, nil),
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
			m := New(zap.NewNop(), tt.size, nil)

			name, got := gatherGauge(t, m.EmailChannelSize)
			assert.Equal(t, "klubok_email_channel_size", name)
			assert.InDelta(t, tt.want, got, 0)
		})
	}
}

func TestMetrics_WSConnectionsCurrent(t *testing.T) {
	tests := []struct {
		name        string
		connections func() int
		want        float64
	}{
		{
			name:        "No connections",
			connections: func() int { return 0 },
			want:        0,
		},
		{
			name:        "Active connections",
			connections: func() int { return 3 },
			want:        3,
		},
		{
			name:        "Nil connections func",
			connections: nil,
			want:        0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New(zap.NewNop(), nil, tt.connections)

			name, got := gatherGauge(t, m.WSConnectionsCurrent)
			assert.Equal(t, "klubok_ws_connections_current", name)
			assert.InDelta(t, tt.want, got, 0)
		})
	}
}

func gatherGauge(t *testing.T, gauge prometheus.Collector) (string, float64) {
	t.Helper()

	reg := prometheus.NewRegistry()
	reg.MustRegister(gauge)

	mfs, err := reg.Gather()
	require.NoError(t, err)
	require.Len(t, mfs, 1)
	require.Len(t, mfs[0].GetMetric(), 1)

	return mfs[0].GetName(), mfs[0].GetMetric()[0].GetGauge().GetValue()
}

func TestMetrics_HTTPRequestsCurrentIncDec(t *testing.T) {
	m := New(zap.NewNop(), nil, nil)

	assert.NotPanics(t, func() {
		m.HTTPRequestsCurrentInc()
		m.HTTPRequestsCurrentInc()
		m.HTTPRequestsCurrentDec()
	})
}
