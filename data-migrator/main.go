package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"

	"data-migrator/db"
)

type dbConfig struct {
	Host     string
	Port     string
	DBName   string
	User     string
	Password string
}

func main() {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg := dbConfig{
		Host:     getEnv("POSTGRES_HOST", "klubok_db"),
		Port:     getEnv("POSTGRES_PORT", "5432"),
		DBName:   getEnv("POSTGRES_DB", "postgres"),
		User:     getEnv("POSTGRES_USER", "postgres"),
		Password: getEnv("POSTGRES_PASSWORD", "postgres"),
	}

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=disable",
		cfg.User,
		cfg.Password,
		net.JoinHostPort(cfg.Host, cfg.Port),
		cfg.DBName,
	)

	logger.Info("connecting to database", "db_name", cfg.DBName)

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		return
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		logger.Error("failed to ping database", "error", err)
		return
	}

	logger.Info("connected to database")

	fillOrderShapes := db.NewFillOrderShapes(pool, logger)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /fill-order-shapes", func(w http.ResponseWriter, r *http.Request) {
		if err := fillOrderShapes.Run(r.Context()); err != nil {
			logger.Error("migration failed", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		logger.Info("migration completed successfully")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})

	port := "3000"
	logger.Info("starting migrator HTTP server", "port", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		logger.Error("server error", "error", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
