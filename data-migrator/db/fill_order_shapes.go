package db

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"roci.dev/fracdex"
)

type FillOrderShapes struct {
	pool   dbTX
	logger *slog.Logger
}

type shapeRow struct {
	ID    string
	Shape shape
}

type shape struct {
	ID     string    `json:"id"`
	Type   shapeType `json:"tp"`
	Left   float64   `json:"l"`
	Top    float64   `json:"t"`
	Width  float64   `json:"w"`
	Height float64   `json:"h"`
	ScaleX float64   `json:"sx"`
	ScaleY float64   `json:"sy"`
	Fill   string    `json:"fc"`
	Stroke string    `json:"sc"`
	Text   string    `json:"txt"`
	Order  string    `json:"o"`
}

type shapeType string

func NewFillOrderShapes(pool dbTX, logger *slog.Logger) *FillOrderShapes {
	return &FillOrderShapes{
		pool:   pool,
		logger: logger,
	}
}

func (m *FillOrderShapes) Run(ctx context.Context) error {
	rows, err := m.pool.Query(ctx, "SELECT DISTINCT diagram_id FROM diagram_shapes WHERE data->>'o' IS NULL")
	if err != nil {
		return fmt.Errorf("query distinct diagram_ids: %w", err)
	}
	defer rows.Close()

	var diagramIDs []string
	for rows.Next() {
		var diagramID string
		if err := rows.Scan(&diagramID); err != nil {
			return fmt.Errorf("scan diagram_id: %w", err)
		}
		diagramIDs = append(diagramIDs, diagramID)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate diagram_ids: %w", err)
	}

	if len(diagramIDs) == 0 {
		m.logger.Info("no shapes need order fill")
		return nil
	}

	m.logger.Info("found diagrams with shapes needing order", "count", len(diagramIDs))

	for _, diagramID := range diagramIDs {
		if err := m.processDiagram(ctx, diagramID); err != nil {
			return fmt.Errorf("process diagram %s: %w", diagramID, err)
		}
	}

	return nil
}

func (m *FillOrderShapes) processDiagram(ctx context.Context, diagramID string) error {
	var count int
	err := m.pool.QueryRow(ctx, "SELECT COUNT(*) FROM diagram_shapes WHERE diagram_id = $1 AND data->>'o' IS NULL", diagramID).Scan(&count)
	if err != nil {
		return fmt.Errorf("count shapes without order for diagram %s: %w", diagramID, err)
	}

	if count == 0 {
		return nil
	}

	m.logger.Info("processing diagram", "diagram_id", diagramID, "shapes_without_order", count)

	rows, err := m.pool.Query(ctx, "SELECT id, data FROM diagram_shapes WHERE diagram_id = $1", diagramID)
	if err != nil {
		return fmt.Errorf("query shapes for diagram %s: %w", diagramID, err)
	}
	defer rows.Close()

	shapes := make([]shapeRow, 0, count)
	for rows.Next() {
		var s shapeRow
		var data []byte
		if err := rows.Scan(&s.ID, &data); err != nil {
			return fmt.Errorf("scan shape: %w", err)
		}
		if err := json.Unmarshal(data, &s.Shape); err != nil {
			return fmt.Errorf("unmarshal shape %s: %w", s.ID, err)
		}
		shapes = append(shapes, s)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate shapes for diagram %s: %w", diagramID, err)
	}

	for i := range shapes {
		shapes[i].Shape.Order = generateOrderKey(i, shapes)
	}

	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer tx.Rollback(ctx)

	updateStmt := "UPDATE diagram_shapes SET data = $1, updated_at = NOW() WHERE id = $2"
	for _, s := range shapes {
		data, err := json.Marshal(s.Shape)
		if err != nil {
			return fmt.Errorf("marshal shape %s: %w", s.ID, err)
		}
		if _, err := tx.Exec(ctx, updateStmt, data, s.ID); err != nil {
			return fmt.Errorf("update shape %s: %w", s.ID, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	m.logger.Info("order shapes filled for diagram", "diagram_id", diagramID, "count", len(shapes))

	return nil
}

func generateOrderKey(idx int, shapes []shapeRow) string {
	if idx == 0 {
		key, _ := fracdex.KeyBetween("", "")
		return key
	}

	key, _ := fracdex.KeyBetween(shapes[idx-1].Shape.Order, "")
	return key
}
