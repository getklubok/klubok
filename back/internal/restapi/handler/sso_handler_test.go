package handler

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"klubok/internal/domain"
	"klubok/internal/provider/sso"
	"klubok/internal/restapi/handler/mocks"
	"klubok/internal/restapi/oas"
)

func TestSSO_SSOLogin(t *testing.T) {
	t.Run("Disabled returns NotFound", func(t *testing.T) {
		h := NewSSO(
			SSOConfig{Enabled: false},
			mocks.NewSSOProviderMock(t),
			mocks.NewUserServiceMock(t),
			mocks.NewCookieServiceMock(t),
			zap.NewNop(),
		)

		got, err := h.SSOLogin(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, &oas.SSOLoginNotFound{
			Message: oas.NewOptString("sso login disabled"),
		}, got)
	})

	t.Run("Enabled redirects with state cookie", func(t *testing.T) {
		provider := mocks.NewSSOProviderMock(t)
		provider.EXPECT().
			AuthCodeURL(mock.Anything, mock.Anything).
			Run(func(state string, nonce string) {
				assert.NotEmpty(t, state)
				assert.Equal(t, state, nonce)
			}).
			Return("https://provider.example/authorize?state=xxx")

		h := NewSSO(
			SSOConfig{Enabled: true, Issuer: "https://provider.example", CookieSecure: true},
			provider,
			mocks.NewUserServiceMock(t),
			mocks.NewCookieServiceMock(t),
			zap.NewNop(),
		)

		got, err := h.SSOLogin(context.Background())
		assert.NoError(t, err)

		found, ok := got.(*oas.SSOLoginFound)
		assert.True(t, ok)
		assert.Equal(t, "https://provider.example/authorize?state=xxx", found.Location)
		assert.Contains(t, found.SetCookie, "sso_state=")
		assert.Contains(t, found.SetCookie, "HttpOnly")
		assert.Contains(t, found.SetCookie, "Max-Age=300")
	})
}

func TestSSO_SSOCallback(t *testing.T) {
	newEnabled := func(t *testing.T) (*SSO, *mocks.SSOProviderMock, *mocks.UserServiceMock, *mocks.CookieServiceMock) {
		t.Helper()

		provider := mocks.NewSSOProviderMock(t)
		userSvc := mocks.NewUserServiceMock(t)
		cookieSvc := mocks.NewCookieServiceMock(t)

		h := NewSSO(
			SSOConfig{Enabled: true, Issuer: "https://provider.example", CookieSecure: true, PostLoginRedirectURL: "http://localhost:5173"},
			provider,
			userSvc,
			cookieSvc,
			zap.NewNop(),
		)

		return h, provider, userSvc, cookieSvc
	}

	t.Run("Disabled returns NotFound", func(t *testing.T) {
		h := NewSSO(
			SSOConfig{Enabled: false},
			mocks.NewSSOProviderMock(t),
			mocks.NewUserServiceMock(t),
			mocks.NewCookieServiceMock(t),
			zap.NewNop(),
		)

		got, err := h.SSOCallback(context.Background(), oas.SSOCallbackParams{
			Code: "code", State: "state", SSOState: "state",
		})
		assert.NoError(t, err)
		assert.Equal(t, &oas.SSOCallbackNotFound{
			Message: oas.NewOptString("sso login disabled"),
		}, got)
	})

	t.Run("State mismatch returns BadRequest", func(t *testing.T) {
		h, _, _, _ := newEnabled(t)

		got, err := h.SSOCallback(context.Background(), oas.SSOCallbackParams{
			Code: "code", State: "aaa", SSOState: "bbb",
		})
		assert.NoError(t, err)
		assert.Equal(t, &oas.SSOCallbackBadRequest{
			Message: oas.NewOptString("invalid state"),
		}, got)
	})

	t.Run("Provider error returns BadRequest without session", func(t *testing.T) {
		h, provider, _, _ := newEnabled(t)
		provider.EXPECT().
			Authenticate(mock.Anything, "bad-code", "state").
			Return(sso.Claims{}, errors.New("exchange failed"))

		got, err := h.SSOCallback(context.Background(), oas.SSOCallbackParams{
			Code: "bad-code", State: "state", SSOState: "state",
		})
		assert.NoError(t, err)
		assert.Equal(t, &oas.SSOCallbackBadRequest{
			Message: oas.NewOptString("sso authentication failed"),
		}, got)
	})

	t.Run("Service error returns InternalServerError", func(t *testing.T) {
		h, provider, userSvc, _ := newEnabled(t)
		provider.EXPECT().
			Authenticate(mock.Anything, "code", "state").
			Return(sso.Claims{Sub: "sub-1", Email: "john@example.com", PreferredUsername: "john"}, nil)
		userSvc.EXPECT().
			LoginSSO(mock.Anything, "https://provider.example", "sub-1", "john@example.com", "john", mock.Anything).
			Return(domain.CookieID(""), errors.New("service error"))

		got, err := h.SSOCallback(contextWithIP("0.0.0.0"), oas.SSOCallbackParams{
			Code: "code", State: "state", SSOState: "state",
		})
		assert.NoError(t, err)
		assert.Equal(t, &oas.SSOCallbackInternalServerError{}, got)
	})

	t.Run("Success sets session and redirects", func(t *testing.T) {
		h, provider, userSvc, cookieSvc := newEnabled(t)
		provider.EXPECT().
			Authenticate(mock.Anything, "code", "state").
			Return(sso.Claims{Sub: "sub-1", Email: "john@example.com", PreferredUsername: "john"}, nil)
		userSvc.EXPECT().
			LoginSSO(mock.Anything, "https://provider.example", "sub-1", "john@example.com", "john", domain.IP("0.0.0.0")).
			Return(domain.CookieID("cookie-123"), nil)
		cookieSvc.EXPECT().
			Create(domain.CookieID("cookie-123")).
			Return(http.Cookie{Name: "session", Value: "cookie-123"})

		got, err := h.SSOCallback(contextWithIP("0.0.0.0"), oas.SSOCallbackParams{
			Code: "code", State: "state", SSOState: "state",
		})
		assert.NoError(t, err)

		found, ok := got.(*oas.SSOCallbackFound)
		assert.True(t, ok)
		assert.Equal(t, "http://localhost:5173", found.Location)
		assert.True(t, strings.HasPrefix(found.SetCookie, "session=cookie-123"))
	})
}
