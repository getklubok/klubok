package handler

import (
	"context"
	"net/http"

	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/microcosm-cc/bluemonday"

	"klubok/internal/application/service/cookie"
	"klubok/internal/application/service/diagram"
	"klubok/internal/application/service/embed"
	"klubok/internal/application/service/permission"
	"klubok/internal/application/service/session"
	usr "klubok/internal/application/service/user"
	"klubok/internal/application/websocket"
	"klubok/internal/domain"
	"klubok/internal/provider/sso"
	"klubok/internal/provider/yandex"
)

// DBHealthChecker проверяет доступность БД для readiness-пробы.
type DBHealthChecker interface {
	Ping(ctx context.Context) error
}

var _ DBHealthChecker = (*pgxpool.Pool)(nil)

type UserService interface {
	Register(ctx context.Context, user domain.User, ip domain.IP) (domain.User, error)
	Activate(ctx context.Context, userID domain.UserID, code string) (cookieID domain.CookieID, err error)
	Login(ctx context.Context, email string) error
	LoginVerify(ctx context.Context, email string, code string) (cookieID domain.CookieID, err error)
	LoginSSO(ctx context.Context, issuer string, sub string, email string, preferredUsername string, ip domain.IP) (domain.CookieID, error)
	FindByID(ctx context.Context, id domain.UserID) (*domain.User, error)
	FindByName(ctx context.Context, name string) (*domain.User, error)
	CreateFake(ctx context.Context, ip domain.IP) (domain.User, domain.CookieID, error)
}

var _ UserService = (*usr.Service)(nil)

// SSOProvider вход через SSO по протоколу OIDC.
type SSOProvider interface {
	AuthCodeURL(state string, nonce string) string
	Authenticate(ctx context.Context, code string, nonce string) (sso.Claims, error)
}

var _ SSOProvider = (*sso.OIDCProvider)(nil)

type CookieService interface {
	Create(v domain.CookieID) http.Cookie
	Clear() http.Cookie
}

var _ CookieService = (*cookie.Service)(nil)

type SessionService interface {
	FindByCID(ctx context.Context, cid domain.CookieID) (domain.Session, error)
	DeleteByID(ctx context.Context, id domain.SessionID) error
}

var _ SessionService = (*session.Service)(nil)

type SessionUpdater interface {
	UpdateTime(sessionID domain.SessionID)
}

var _ SessionUpdater = (*session.Updater)(nil)

type SessionCache interface {
	Add(key domain.CookieID, value domain.Session) (evicted bool)
	Get(key domain.CookieID) (value domain.Session, ok bool)
}

var _ SessionCache = (*expirable.LRU[domain.CookieID, domain.Session])(nil)

type DiagramService interface {
	Create(ctx context.Context, title string, userID domain.UserID) (domain.Diagram, error)
	List(ctx context.Context, userID domain.UserID) ([]domain.Diagram, error)
	Delete(ctx context.Context, diagramID domain.DiagramID) error
	SetTitle(ctx context.Context, id domain.DiagramID, title string) error
	Info(ctx context.Context, diagramID domain.DiagramID) (domain.DiagramInfo, error)
	InfoByUser(ctx context.Context, diagramID domain.DiagramID, userID domain.UserID) (domain.DiagramInfo, error)
	Clone(ctx context.Context, diagramID domain.DiagramID, userID domain.UserID) (domain.Diagram, error)
	Import(ctx context.Context, userID domain.UserID, diagramInfo domain.DiagramInfo) (domain.Diagram, error)

	ShapeUpsert(ctx context.Context, userID domain.UserID, diagramID domain.DiagramID, shape domain.DiagramShape) error
	ShapeDelete(ctx context.Context, userID domain.UserID, diagramID domain.DiagramID, shapeID domain.ShapeID) error
	ShapeBatchUpsert(ctx context.Context, userID domain.UserID, diagramID domain.DiagramID, shapes []domain.DiagramShape) error

	ConnectionUpsert(ctx context.Context, userID domain.UserID, diagramID domain.DiagramID, connection domain.DiagramConnection) error
	ConnectionDelete(ctx context.Context, userID domain.UserID, diagramID domain.DiagramID, connection domain.DiagramConnection) error
	ConnectionBatchUpsert(ctx context.Context, userID domain.UserID, diagramID domain.DiagramID, connections []domain.DiagramConnection) error

	CommentCreate(ctx context.Context, userID domain.UserID, comment domain.DiagramComment) (domain.DiagramComment, error)
	CommentList(ctx context.Context, diagramID domain.DiagramID, diagramShapeID domain.ShapeID) ([]domain.DiagramComment, error)

	CollaboratorUpsert(ctx context.Context, collaborator domain.DiagramCollaborator) (domain.DiagramCollaborator, error)
	CollaboratorList(ctx context.Context, diagramID domain.DiagramID) ([]domain.DiagramCollaborator, error)
	CollaboratorDelete(ctx context.Context, diagramID domain.DiagramID, collaboratorID domain.CollaboratorID) error

	OnlineUsers(ctx context.Context, diagramID domain.DiagramID) []domain.WebSocketUser
}

var _ DiagramService = (*diagram.Service)(nil)

type PermissionService interface {
	HasDiagramPermission(
		ctx context.Context,
		diagramID domain.DiagramID,
		userID domain.UserID,
		permissions []domain.DiagramCollaboratorPermission,
	) (bool, error)
}

var _ PermissionService = (*permission.Service)(nil)

type WebSocketManager interface {
	GetDiagramStats(diagramID domain.DiagramID) domain.WebSocketStatDiagram
}

var _ WebSocketManager = (*websocket.Manager)(nil)

type CaptchaService interface {
	Check(ctx context.Context, token string, userIP domain.IP) bool
}

var _ CaptchaService = (*yandex.CaptchaService)(nil)

type TextSanitizer interface {
	Sanitize(s string) string
}

var _ TextSanitizer = (*bluemonday.Policy)(nil)

type EmbedService interface {
	Create(ctx context.Context, diagramID domain.DiagramID, userID domain.UserID) (domain.DiagramEmbedInfo, error)
	List(ctx context.Context, diagramID domain.DiagramID) ([]domain.DiagramEmbedInfo, error)
	GetByID(ctx context.Context, embedID domain.EmbedID) (domain.DiagramEmbedInfo, error)
	GetByToken(ctx context.Context, token string) (domain.DiagramEmbedInfo, error)
	SoftDelete(ctx context.Context, embedID domain.EmbedID) error
	IncrementViews(ctx context.Context, embedID domain.EmbedID) error
}

var _ EmbedService = (*embed.Service)(nil)
