//go:build integration && postgres

package db

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
)

func TestIntegration_FillOrderShapes_Run(t *testing.T) {
	ctx := t.Context()

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=disable",
		getEnv("TEST_INTEGRATION_PG_USER", "postgres"),
		getEnv("TEST_INTEGRATION_PG_PASSWORD", "postgres"),
		net.JoinHostPort("127.0.0.1", getEnv("TEST_INTEGRATION_PG_PORT", "5433")),
		getEnv("TEST_INTEGRATION_PG_DATABASE", "postgres"),
	)

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Skipf("skipping: cannot connect to database: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("ping: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	tests := []struct {
		name       string
		setupSQL   string
		cleanupSQL string
		want       []shape
	}{
		{
			name: "fills order for shapes without order",
			setupSQL: `
				CREATE TABLE IF NOT EXISTS diagram_shapes (
					id TEXT PRIMARY KEY,
					diagram_id TEXT NOT NULL,
					data JSONB NOT NULL,
					updated_at TIMESTAMPTZ DEFAULT NOW()
				);
				DELETE FROM diagram_shapes WHERE diagram_id = 'd1';
				INSERT INTO diagram_shapes (id, diagram_id, data) VALUES
					('1', 'd1', '{"id":"1","tp":"SQUARE","l":10,"t":10}'::jsonb),
					('2', 'd1', '{"id":"2","tp":"ELLIPSE","l":20,"t":20}'::jsonb),
					('3', 'd1', '{"id":"3","tp":"DIAMOND","l":30,"t":30}'::jsonb);
			`,
			cleanupSQL: `DELETE FROM diagram_shapes WHERE diagram_id = 'd1';`,
			want: []shape{
				{
					ID:    "1",
					Type:  "SQUARE",
					Left:  10,
					Top:   10,
					Order: "a0",
				},
				{
					ID:    "2",
					Type:  "ELLIPSE",
					Left:  20,
					Top:   20,
					Order: "a1",
				},
				{
					ID:    "3",
					Type:  "DIAMOND",
					Left:  30,
					Top:   30,
					Order: "a2",
				},
			},
		},
		{
			name: "skips shapes that already have order",
			setupSQL: `
				CREATE TABLE IF NOT EXISTS diagram_shapes (
					id TEXT PRIMARY KEY,
					diagram_id TEXT NOT NULL,
					data JSONB NOT NULL,
					updated_at TIMESTAMPTZ DEFAULT NOW()
				);
				DELETE FROM diagram_shapes WHERE diagram_id = 'd2';
				INSERT INTO diagram_shapes (id, diagram_id, data) VALUES
					('4', 'd2', '{"id":"4","tp":"SQUARE","o":"a0"}'::jsonb),
					('5', 'd2', '{"id":"5","tp":"ELLIPSE","o":"a1"}'::jsonb);
			`,
			cleanupSQL: `DELETE FROM diagram_shapes WHERE diagram_id = 'd2';`,
			want: []shape{
				{
					ID:    "4",
					Type:  "SQUARE",
					Order: "a0",
				},
				{
					ID:    "5",
					Type:  "ELLIPSE",
					Order: "a1",
				},
			},
		},
		{
			name: "fills order for mixed shapes",
			setupSQL: `
				CREATE TABLE IF NOT EXISTS diagram_shapes (
					id TEXT PRIMARY KEY,
					diagram_id TEXT NOT NULL,
					data JSONB NOT NULL,
					updated_at TIMESTAMPTZ DEFAULT NOW()
				);
				DELETE FROM diagram_shapes WHERE diagram_id = 'd3';
				INSERT INTO diagram_shapes (id, diagram_id, data) VALUES
					('6', 'd3', '{"id":"6","tp":"SQUARE","o":"a3"}'::jsonb),
					('7', 'd3', '{"id":"7","tp":"ELLIPSE"}'::jsonb),
					('8', 'd3', '{"id":"8","tp":"DIAMOND","o":"Zz"}'::jsonb);
			`,
			cleanupSQL: `DELETE FROM diagram_shapes WHERE diagram_id = 'd3';`,
			want: []shape{
				{
					ID:    "6",
					Type:  "SQUARE",
					Order: "a0",
				},
				{
					ID:    "7",
					Type:  "ELLIPSE",
					Order: "a1",
				},
				{
					ID:    "8",
					Type:  "DIAMOND",
					Order: "a2",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := pool.Exec(ctx, tt.setupSQL)
			if err != nil {
				t.Fatalf("setup: %v", err)
			}
			defer pool.Exec(ctx, tt.cleanupSQL)

			fillOrderShapes := NewFillOrderShapes(pool, logger)
			if err := fillOrderShapes.Run(ctx); err != nil {
				t.Fatalf("Run: %v", err)
			}

			rows, err := pool.Query(ctx, "SELECT data FROM diagram_shapes")
			if err != nil {
				t.Fatalf("query shapes: %v", err)
			}
			defer rows.Close()

			got := make([]shape, 0)
			for rows.Next() {
				var data []byte
				if err := rows.Scan(&data); err != nil {
					t.Fatalf("scan shape: %v", err)
				}
				var s shape
				if err := json.Unmarshal(data, &s); err != nil {
					t.Fatalf("unmarshal shape: %v", err)
				}
				got = append(got, s)
			}
			if err := rows.Err(); err != nil {
				t.Fatalf("iterate shapes: %v", err)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
