package permission

import (
	"context"

	"klubok/internal/db/sqlc"
)

type Querier interface {
	DiagramCollaboratorExist(ctx context.Context, diagramID string, userID string) (sqlc.DiagramCollaborator, error)
}

var _ Querier = (*sqlc.Queries)(nil)
