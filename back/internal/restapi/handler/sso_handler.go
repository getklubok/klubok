package handler

import (
	"context"
	"net/http"

	"go.uber.org/zap"

	"klubok/internal/application/zapctx"
	"klubok/internal/domain"
	"klubok/internal/restapi/oas"
)

// SSOStateCookieName — имя cookie для защиты SSO-потока от CSRF/replay.
const SSOStateCookieName = "sso_state"

const (
	ssoStateLength = 32
	ssoStateMaxAge = 300 // 5 мин
)

// SSOConfig конфигурация SSO-обработчика.
type SSOConfig struct {
	Enabled              bool
	Issuer               string
	CookieSecure         bool
	PostLoginRedirectURL string
}

type SSO struct {
	cfg SSOConfig

	provider      SSOProvider
	userService   UserService
	cookieService CookieService

	log *zap.Logger
}

func NewSSO(
	cfg SSOConfig,
	provider SSOProvider,
	userService UserService,
	cookieService CookieService,
	log *zap.Logger,
) *SSO {
	return &SSO{
		cfg: cfg,

		provider:      provider,
		userService:   userService,
		cookieService: cookieService,

		log: log.Named("sso_handler"),
	}
}

// Start SSO login.
//
// GET /auth/sso/login.
func (h *SSO) SSOLogin(_ context.Context) (oas.SSOLoginRes, error) { //nolint:unparam // It's ok
	if !h.cfg.Enabled {
		h.log.Info("SSO login disabled")
		return &oas.SSOLoginNotFound{
			Message: oas.NewOptString("sso login disabled"),
		}, nil
	}

	// ponytail: одно случайное значение служит и state, и nonce — пару хранить не нужно
	state := domain.NewString(ssoStateLength)

	stateCookie := (&http.Cookie{
		Name:     SSOStateCookieName,
		Value:    state,
		Path:     "/",
		MaxAge:   ssoStateMaxAge,
		Secure:   h.cfg.CookieSecure,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}).String()

	return &oas.SSOLoginFound{
		Location:  h.provider.AuthCodeURL(state, state),
		SetCookie: stateCookie,
	}, nil
}

// SSO callback from provider.
//
// GET /auth/sso/callback.
func (h *SSO) SSOCallback(ctx context.Context, params oas.SSOCallbackParams) (oas.SSOCallbackRes, error) { //nolint:unparam // It's ok
	if !h.cfg.Enabled {
		h.log.Info("SSO login disabled")
		return &oas.SSOCallbackNotFound{
			Message: oas.NewOptString("sso login disabled"),
		}, nil
	}

	log := zapctx.Grasp(ctx, h.log)

	if params.State != params.SSOState {
		log.Warn("SSO state mismatch")
		return &oas.SSOCallbackBadRequest{
			Message: oas.NewOptString("invalid state"),
		}, nil
	}

	// Детали ошибки провайдера не утекают в ответ.
	claims, err := h.provider.Authenticate(ctx, params.Code, params.SSOState)
	if err != nil {
		log.Warn("SSO authenticate failed",
			zap.Error(err),
		)
		return &oas.SSOCallbackBadRequest{
			Message: oas.NewOptString("sso authentication failed"),
		}, nil
	}

	ip := domain.IPFromContext(ctx)

	cookieID, err := h.userService.LoginSSO(ctx, h.cfg.Issuer, claims.Sub, claims.Email, claims.PreferredUsername, ip)
	if err != nil {
		log.Error("SSO login",
			zap.Error(err),
		)
		return &oas.SSOCallbackInternalServerError{}, nil
	}

	cookie := h.cookieService.Create(cookieID)

	return &oas.SSOCallbackFound{
		Location:  h.cfg.PostLoginRedirectURL,
		SetCookie: cookie.String(),
	}, nil
}
