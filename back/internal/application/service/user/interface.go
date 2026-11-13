package user

import (
	"context"
	"time"

	sess "klubok/internal/application/service/session"
	"klubok/internal/db"
	"klubok/internal/db/sqlc"
	"klubok/internal/domain"
)

type Querier interface {
	AuthCodeUpsert(ctx context.Context, arg sqlc.AuthCodeUpsertParams) error
	AuthCodeFind(ctx context.Context, userEmail string, code string) (sqlc.AuthCode, error)
	AuthCodeDelete(ctx context.Context, id string) error
	AuthCodeCancelOld(ctx context.Context, createdAt time.Time) error

	UserUpdate(ctx context.Context, arg sqlc.UserUpdateParams) (sqlc.User, error)
	UserFindByID(ctx context.Context, id string) (sqlc.User, error)
	UserFindByEmail(ctx context.Context, email string) (sqlc.User, error)
	UserFindByName(ctx context.Context, username string) (sqlc.User, error)
	UserDeleteNotActivated(ctx context.Context, createdAt time.Time) error
	UserCancelActivateCode(ctx context.Context, id string) error

	UserSSOIdentityFind(ctx context.Context, issuer string, sub string) (sqlc.UserSsoIdentity, error)
	UserSSOIdentityCreate(ctx context.Context, arg sqlc.UserSSOIdentityCreateParams) (sqlc.UserSsoIdentity, error)

	SessionCreate(ctx context.Context, arg sqlc.SessionCreateParams) (sqlc.Session, error)
}

var _ Querier = (*sqlc.Queries)(nil)

type DBUseCase interface {
	UserCreate(ctx context.Context, user sqlc.UserCreateParams, privacyLog sqlc.PrivacyLogCreateParams) (sqlc.User, error)
	UserCreateWithSSO(
		ctx context.Context,
		user sqlc.UserCreateParams,
		privacyLog sqlc.PrivacyLogCreateParams,
		identity sqlc.UserSSOIdentityCreateParams,
	) (sqlc.User, error)
}

var _ DBUseCase = (*db.UseCase)(nil)

type SessionService interface {
	Create(ctx context.Context, session domain.Session) (domain.Session, error)
}

var _ SessionService = (*sess.Service)(nil)
