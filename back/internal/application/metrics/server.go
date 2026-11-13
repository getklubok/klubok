package metrics

import (
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

type ServerConfig struct {
	Metrics *Metrics
	DB      ServerConfigDB
	HTTP    ServerConfigHTTP
}

type ServerConfigHTTP struct {
	Port int

	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
}

type ServerConfigDB struct {
	Pool *pgxpool.Pool
	Name string
}

type Server struct {
	httpServer *http.Server

	log *zap.Logger
}

func NewServer(cfg ServerConfig, log *zap.Logger) *Server {
	log = log.Named("metrics_server")

	reg := prometheus.NewRegistry()

	reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		newPgxpoolStatsCollector(cfg.DB.Pool, cfg.DB.Name),
		cfg.Metrics.HTTPRequestsTotal,
		cfg.Metrics.HTTPRequestsCurrent,
		cfg.Metrics.HTTPRequestDurationSeconds,
	)

	httpServer := &http.Server{
		Addr: fmt.Sprintf(":%d", cfg.HTTP.Port),

		ReadTimeout:       cfg.HTTP.ReadTimeout,
		ReadHeaderTimeout: cfg.HTTP.ReadHeaderTimeout,
		WriteTimeout:      cfg.HTTP.WriteTimeout,
		IdleTimeout:       cfg.HTTP.IdleTimeout,
	}

	http.DefaultServeMux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))

	go func() {
		log.Info("Metrics started", zap.String("address", "http://"+httpServer.Addr+"/metrics"))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Metrics listen", zap.Error(err))
		}
	}()

	return &Server{
		httpServer: httpServer,

		log: log,
	}
}

func (s *Server) Stop() {
	if s.httpServer == nil {
		return
	}

	if err := s.httpServer.Close(); err != nil {
		s.log.Error("Metrics stop", zap.Error(err))
	}
}
