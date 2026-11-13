package websocket

import (
	"context"
	"net/http"

	"github.com/hashicorp/golang-lru/v2/expirable"

	"klubok/internal/application/service/cookie"
	"klubok/internal/application/service/permission"
	"klubok/internal/application/service/session"
	"klubok/internal/application/service/user"
	"klubok/internal/domain"
)

type CookieManager interface {
	Get(r *http.Request) (domain.CookieID, error)
}

var _ CookieManager = (*cookie.Service)(nil)

type SessionService interface {
	FindByCID(ctx context.Context, cid domain.CookieID) (domain.Session, error)
}

var _ SessionService = (*session.Service)(nil)

type UserService interface {
	FindByID(ctx context.Context, id domain.UserID) (*domain.User, error)
}

var _ UserService = (*user.Service)(nil)

type SessionCache interface {
	Add(key domain.CookieID, value *domain.Session) (evicted bool)
	Get(key domain.CookieID) (value *domain.Session, ok bool)
}

var _ SessionCache = (*expirable.LRU[domain.CookieID, *domain.Session])(nil)

type UserCache interface {
	Add(key domain.UserID, value *domain.User) (evicted bool)
	Get(key domain.UserID) (value *domain.User, ok bool)
}

var _ UserCache = (*expirable.LRU[domain.UserID, *domain.User])(nil)

type PermissionService interface {
	HasDiagramPermission(
		ctx context.Context,
		diagramID domain.DiagramID,
		userID domain.UserID,
		permissions []domain.DiagramCollaboratorPermission,
	) (bool, error)
}

var _ PermissionService = (*permission.Service)(nil)
