package session

import (
	"context"
	"time"

	"klubok/internal/db/sqlc"
)

type Querier interface {
	SessionCancelOld(ctx context.Context, checkedAt time.Time) error
	SessionCreate(ctx context.Context, arg sqlc.SessionCreateParams) (sqlc.Session, error)
	SessionDeleteByID(ctx context.Context, id string) error
	SessionFindByCID(ctx context.Context, cid string) (sqlc.Session, error)
	SessionFindByID(ctx context.Context, id string) (sqlc.Session, error)
	SessionFindByUID(ctx context.Context, uid string) ([]sqlc.Session, error)
	SessionUpdateCheckedAt(ctx context.Context, id string, checkedAt time.Time) (sqlc.Session, error)

	UserUpdateLastLoginAt(ctx context.Context, id string) error
}

var _ Querier = (*sqlc.Queries)(nil)
