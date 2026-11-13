package handler

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
	"go.uber.org/zap"

	"klubok/internal/domain"
	"klubok/internal/restapi/oas"
)

const (
	defaultSessionCacheSize = 1000
	defaultSessionCacheTTL  = 10 * time.Minute
)

type Security struct {
	sessionService SessionService
	sessionUpdater SessionUpdater

	staticClaims *domain.Claims

	log *zap.Logger

	sessionCache SessionCache
}

func NewSecurity(
	sessionService SessionService,
	sessionUpdater SessionUpdater,
	staticClaims *domain.Claims,
	log *zap.Logger,
) *Security {
	return &Security{
		sessionService: sessionService,
		sessionUpdater: sessionUpdater,

		staticClaims: staticClaims,

		log: log.Named("security_handler"),

		sessionCache: expirable.NewLRU[domain.CookieID, domain.Session](defaultSessionCacheSize, nil, defaultSessionCacheTTL),
	}
}

func (s *Security) HandleSessionCookie(ctx context.Context, _ oas.OperationName, t oas.SessionCookie) (context.Context, error) {
	apiKey := t.GetAPIKey()

	if apiKey == "" && s.staticClaims != nil {
		return s.staticClaims.Context(ctx), nil
	}

	session, err := s.getSession(ctx, domain.CookieID(apiKey))
	if err != nil {
		if errors.Is(err, domain.ErrSessionNotFound) {
			s.log.Warn("Session not found")
			return ctx, errors.New("unauthorized")
		}

		s.log.Error("Find session", zap.Error(err))
		return ctx, errors.New("internal server error")
	}

	claims := domain.Claims{
		SessionID: session.ID,
		UserID:    session.UserID,
	}

	return claims.Context(ctx), nil
}

func (s *Security) getSession(ctx context.Context, cid domain.CookieID) (*domain.Session, error) {
	sess, ok := s.sessionCache.Get(cid)
	if !ok {
		var err error

		sess, err = s.sessionService.FindByCID(ctx, cid)
		if err != nil {
			return nil, fmt.Errorf("find session by cid: %w", err)
		}

		s.sessionCache.Add(cid, sess)
	}

	return &sess, nil
}
