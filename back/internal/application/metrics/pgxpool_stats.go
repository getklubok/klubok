package metrics

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
)

type pgxpoolStatsCollector struct {
	pool *pgxpool.Pool

	maxOpenConnections *prometheus.Desc
	openConnections    *prometheus.Desc
	inUseConnections   *prometheus.Desc
	idleConnections    *prometheus.Desc
	waitCount          *prometheus.Desc
	waitDuration       *prometheus.Desc
	maxIdleClosed      *prometheus.Desc
	maxLifetimeClosed  *prometheus.Desc
}

func newPgxpoolStatsCollector(pool *pgxpool.Pool, dbName string) *pgxpoolStatsCollector {
	fqName := func(name string) string {
		return "go_sql_" + name
	}
	return &pgxpoolStatsCollector{
		pool: pool,

		maxOpenConnections: prometheus.NewDesc(
			fqName("max_open_connections"),
			"Maximum number of open connections to the database.",
			nil, prometheus.Labels{"db_name": dbName},
		),
		openConnections: prometheus.NewDesc(
			fqName("open_connections"),
			"The number of established connections both in use and idle.",
			nil, prometheus.Labels{"db_name": dbName},
		),
		inUseConnections: prometheus.NewDesc(
			fqName("in_use_connections"),
			"The number of connections currently in use.",
			nil, prometheus.Labels{"db_name": dbName},
		),
		idleConnections: prometheus.NewDesc(
			fqName("idle_connections"),
			"The number of idle connections.",
			nil, prometheus.Labels{"db_name": dbName},
		),
		waitCount: prometheus.NewDesc(
			fqName("wait_count_total"),
			"The total number of connections waited for.",
			nil, prometheus.Labels{"db_name": dbName},
		),
		waitDuration: prometheus.NewDesc(
			fqName("wait_duration_seconds_total"),
			"The total time blocked waiting for a new connection.",
			nil, prometheus.Labels{"db_name": dbName},
		),
		maxIdleClosed: prometheus.NewDesc(
			fqName("max_idle_closed_total"),
			"The total number of connections closed due to SetMaxIdleConns.",
			nil, prometheus.Labels{"db_name": dbName},
		),
		maxLifetimeClosed: prometheus.NewDesc(
			fqName("max_lifetime_closed_total"),
			"The total number of connections closed due to SetConnMaxLifetime.",
			nil, prometheus.Labels{"db_name": dbName},
		),
	}
}

func (c *pgxpoolStatsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.maxOpenConnections
	ch <- c.openConnections
	ch <- c.inUseConnections
	ch <- c.idleConnections
	ch <- c.waitCount
	ch <- c.waitDuration
	ch <- c.maxIdleClosed
	ch <- c.maxLifetimeClosed
}

func (c *pgxpoolStatsCollector) Collect(ch chan<- prometheus.Metric) {
	stat := c.pool.Stat()

	ch <- prometheus.MustNewConstMetric(c.maxOpenConnections, prometheus.GaugeValue, float64(stat.MaxConns()))
	ch <- prometheus.MustNewConstMetric(c.openConnections, prometheus.GaugeValue, float64(stat.TotalConns()))
	ch <- prometheus.MustNewConstMetric(c.inUseConnections, prometheus.GaugeValue, float64(stat.AcquiredConns()))
	ch <- prometheus.MustNewConstMetric(c.idleConnections, prometheus.GaugeValue, float64(stat.IdleConns()))
	ch <- prometheus.MustNewConstMetric(c.waitCount, prometheus.CounterValue, float64(stat.EmptyAcquireCount()))
	ch <- prometheus.MustNewConstMetric(c.waitDuration, prometheus.CounterValue, stat.EmptyAcquireWaitTime().Seconds())
	ch <- prometheus.MustNewConstMetric(c.maxIdleClosed, prometheus.CounterValue, float64(stat.MaxIdleDestroyCount()))
	ch <- prometheus.MustNewConstMetric(c.maxLifetimeClosed, prometheus.CounterValue, float64(stat.MaxLifetimeDestroyCount()))
}
