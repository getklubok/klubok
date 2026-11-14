package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

type Metrics struct {
	log *zap.Logger

	HTTPRequestsTotal          *prometheus.CounterVec
	HTTPRequestsCurrent        *prometheus.GaugeVec
	HTTPRequestDurationSeconds *prometheus.HistogramVec

	EmailChannelSize     prometheus.GaugeFunc
	WSConnectionsCurrent prometheus.GaugeFunc
}

func New(log *zap.Logger, emailChannelSize, wsConnections func() int) *Metrics {
	if emailChannelSize == nil {
		emailChannelSize = func() int { return 0 }
	}
	if wsConnections == nil {
		wsConnections = func() int { return 0 }
	}

	m := &Metrics{
		log: log.Named("metrics"),

		HTTPRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),
		HTTPRequestsCurrent: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "http_requests_inflight_current",
				Help: "Number of HTTP requests in flight",
			},
			[]string{},
		),
		HTTPRequestDurationSeconds: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "Duration of HTTP requests",
				Buckets: prometheus.ExponentialBuckets(0.001, 2, 15), //nolint:mnd // 1ms to ~32s
			},
			[]string{"method", "path"},
		),
		EmailChannelSize: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Name: "klubok_email_channel_size",
				Help: "Number of emails waiting in the sending channel",
			},
			func() float64 {
				return float64(emailChannelSize())
			},
		),
		WSConnectionsCurrent: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Name: "klubok_ws_connections_current",
				Help: "Number of active WebSocket connections",
			},
			func() float64 {
				return float64(wsConnections())
			},
		),
	}

	return m
}

func (m *Metrics) HTTPRequestsTotalInc(method string, path string, status string) {
	if m.HTTPRequestsTotal == nil {
		return
	}

	cnt, err := m.HTTPRequestsTotal.GetMetricWithLabelValues(method, path, status)
	if err != nil {
		m.log.Error("failed to get metric", zap.Error(err))
		return
	}

	cnt.Inc()
}

func (m *Metrics) HTTPRequestsCurrentInc() {
	if m.HTTPRequestsCurrent == nil {
		return
	}

	cnt, err := m.HTTPRequestsCurrent.GetMetricWithLabelValues()
	if err != nil {
		m.log.Error("failed to get metric", zap.Error(err))
		return
	}

	cnt.Inc()
}

func (m *Metrics) HTTPRequestsCurrentDec() {
	if m.HTTPRequestsCurrent == nil {
		return
	}

	cnt, err := m.HTTPRequestsCurrent.GetMetricWithLabelValues()
	if err != nil {
		m.log.Error("failed to get metric", zap.Error(err))
		return
	}

	cnt.Dec()
}

func (m *Metrics) HTTPRequestDurationSecondsObserve(method string, path string, elapsedSeconds float64) {
	if m.HTTPRequestDurationSeconds == nil {
		return
	}

	cnt, err := m.HTTPRequestDurationSeconds.GetMetricWithLabelValues(method, path)
	if err != nil {
		m.log.Error("failed to get metric", zap.Error(err))
		return
	}

	cnt.Observe(elapsedSeconds)
}
