package application

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ConfigDB struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string

	MaxOpenConns    int32
	MaxIdleConns    int32
	ConnMaxLifetime time.Duration
}

func NewDB(ctx context.Context, cfg ConfigDB) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.Database,
	)
	connConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	connConfig.MaxConnLifetime = cfg.ConnMaxLifetime
	connConfig.MaxConns = cfg.MaxOpenConns
	connConfig.MinIdleConns = cfg.MaxIdleConns

	dbpool, err := pgxpool.NewWithConfig(ctx, connConfig)
	if err != nil {
		return nil, fmt.Errorf("create database connection: %w", err)
	}

	return dbpool, nil
}
