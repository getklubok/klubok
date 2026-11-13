package embed

import (
	"context"

	"klubok/internal/db/sqlc"
)

type Querier interface {
	EmbedCreate(ctx context.Context, arg sqlc.EmbedCreateParams) (sqlc.DiagramEmbed, error)
	EmbedGetByToken(ctx context.Context, token string) (sqlc.DiagramEmbed, error)
	EmbedGetByID(ctx context.Context, id string) (sqlc.DiagramEmbed, error)
	EmbedListByDiagram(ctx context.Context, diagramID string) ([]sqlc.DiagramEmbed, error)
	EmbedSoftDelete(ctx context.Context, id string) error
	EmbedIncrementViews(ctx context.Context, id string) error
}

var _ Querier = (*sqlc.Queries)(nil)
