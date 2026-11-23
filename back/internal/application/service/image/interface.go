package image

import (
	"context"

	"klubok/internal/db/sqlc"
)

// Querier описывает запросы к таблице images, нужные сервису.
type Querier interface {
	ImageCreate(ctx context.Context, arg sqlc.ImageCreateParams) (sqlc.Image, error)
	ImageExistsByBlobHash(ctx context.Context, blobHash string) (bool, error)
	ImageGet(ctx context.Context, iD string, diagramID string) (sqlc.Image, error)
}

var _ Querier = (*sqlc.Queries)(nil)
