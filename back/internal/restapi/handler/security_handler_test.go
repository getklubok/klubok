package handler

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"klubok/internal/domain"
	"klubok/internal/restapi/handler/mocks"
	"klubok/internal/restapi/oas"
)

func TestHandleSessionCookie(t *testing.T) {
	session := domain.Session{
		ID:        "session01",
		CookieID:  "cookie123",
		UserID:    "user01",
		CheckedAt: time.Now(),
		IsCurrent: true,
	}

	tests := []struct {
		name          string
		srv           func() *Security
		sessionCookie oas.SessionCookie
		want          *domain.Claims
		err           error
	}{
		{
			name: "returns static claims when staticClaims is set",
			srv: func() *Security {
				return &Security{
					staticClaims: &domain.Claims{
						SessionID: "static-session",
						UserID:    "static-user",
					},
					log: zap.NewNop(),
				}
			},
			sessionCookie: oas.SessionCookie{},
			want: &domain.Claims{
				SessionID: "static-session",
				UserID:    "static-user",
			},
			err: nil,
		},
		{
			name: "returns session from cache",
			srv: func() *Security {
				sessionService := mocks.NewSessionServiceMock(t)
				sessionCache := mocks.NewSessionCacheMock(t)
				sessionCache.EXPECT().Get(domain.CookieID("cookie123")).Return(session, true)

				return &Security{
					sessionService: sessionService,
					sessionCache:   sessionCache,
					log:            zap.NewNop(),
				}
			},
			sessionCookie: func() oas.SessionCookie {
				sc := oas.SessionCookie{}
				sc.APIKey = "cookie123"
				return sc
			}(),
			want: &domain.Claims{
				SessionID: session.ID,
				UserID:    session.UserID,
			},
			err: nil,
		},
		{
			name: "returns unauthorized when session not found",
			srv: func() *Security {
				sessionService := mocks.NewSessionServiceMock(t)
				sessionCache := mocks.NewSessionCacheMock(t)
				sessionCache.EXPECT().Get(domain.CookieID("invalid")).Return(domain.Session{}, false)
				sessionService.EXPECT().FindByCID(mock.Anything, domain.CookieID("invalid")).Return(domain.Session{}, domain.ErrSessionNotFound)

				return &Security{
					sessionService: sessionService,
					sessionCache:   sessionCache,
					log:            zap.NewNop(),
				}
			},
			sessionCookie: func() oas.SessionCookie {
				sc := oas.SessionCookie{}
				sc.APIKey = "invalid"
				return sc
			}(),
			want: nil,
			err:  errors.New("unauthorized"),
		},
		{
			name: "returns internal error when session service fails",
			srv: func() *Security {
				sessionService := mocks.NewSessionServiceMock(t)
				sessionCache := mocks.NewSessionCacheMock(t)
				sessionCache.EXPECT().Get(domain.CookieID("error-cookie")).Return(domain.Session{}, false)
				sessionService.EXPECT().FindByCID(mock.Anything, domain.CookieID("error-cookie")).Return(domain.Session{}, errors.New("database error"))

				return &Security{
					sessionService: sessionService,
					sessionCache:   sessionCache,
					log:            zap.NewNop(),
				}
			},
			sessionCookie: func() oas.SessionCookie {
				sc := oas.SessionCookie{}
				sc.APIKey = "error-cookie"
				return sc
			}(),
			want: nil,
			err:  errors.New("internal server error"),
		},
		{
			name: "fetches session from service and caches it",
			srv: func() *Security {
				sessionService := mocks.NewSessionServiceMock(t)
				sessionCache := mocks.NewSessionCacheMock(t)
				sessionCache.EXPECT().Get(domain.CookieID("new-cookie")).Return(domain.Session{}, false)
				sessionService.EXPECT().FindByCID(mock.Anything, domain.CookieID("new-cookie")).Return(session, nil)
				sessionCache.EXPECT().Add(domain.CookieID("new-cookie"), session).Return(false)

				return &Security{
					sessionService: sessionService,
					sessionCache:   sessionCache,
					log:            zap.NewNop(),
				}
			},
			sessionCookie: func() oas.SessionCookie {
				sc := oas.SessionCookie{}
				sc.APIKey = "new-cookie"
				return sc
			}(),
			want: &domain.Claims{
				SessionID: session.ID,
				UserID:    session.UserID,
			},
			err: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := tt.srv()
			resultCtx, err := h.HandleSessionCookie(context.Background(), oas.OperationName("test"), tt.sessionCookie)
			assert.Equal(t, tt.err, err)

			if tt.want != nil {
				claims, err := domain.ClaimsFromContext(resultCtx)
				require.NoError(t, err)
				assert.Equal(t, *tt.want, claims)
			}
		})
	}
}
