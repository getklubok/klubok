package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"klubok/internal/domain"
	"klubok/internal/restapi/handler/mocks"
	"klubok/internal/restapi/oas"
)

// contextWithIP builds a request context carrying the client IP via chi's
// ClientIPFromHeader middleware — the same path production middleware uses
// (middleware_manager.go). The context is captured inside the handler because
// a middleware builds a new request with WithContext.
func contextWithIP(ip domain.IP) context.Context {
	var req *http.Request
	handler := chimiddleware.ClientIPFromHeader("X-Real-IP")(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) { req = r },
	))
	httpReq := httptest.NewRequestWithContext(context.Background(), "GET", "/", nil)
	httpReq.Header.Set("X-Real-IP", ip.String())
	handler.ServeHTTP(httptest.NewRecorder(), httpReq)
	return req.Context()
}

func TestUser_UserRegister(t *testing.T) {
	tests := []struct {
		name            string
		req             *oas.UserRegisterRequest
		ctx             func() context.Context
		handler         func(t *testing.T) *User
		want            oas.UserRegisterRes
		wantErr         error
		assertNoCaptcha bool
	}{
		{
			ctx: func() context.Context { return contextWithIP("0.0.0.0") },
			req: &oas.UserRegisterRequest{
				Email:    oas.Email("john@example.com"),
				Username: oas.Username("john"),
			},
			handler: func(t *testing.T) *User {
				t.Helper()

				captchaSvc := mocks.NewCaptchaServiceMock(t)
				captchaSvc.EXPECT().
					Check(mock.Anything, mock.Anything, mock.Anything).
					Return(true)

				userSvc := mocks.NewUserServiceMock(t)
				userSvc.EXPECT().
					Register(mock.Anything, domain.User{
						Email:    "john@example.com",
						Username: "john",
					}, domain.IP("0.0.0.0")).
					Return(domain.User{
						ID:       "user-1",
						Email:    "john@example.com",
						Username: "john",
					}, nil)

				return &User{
					userService:    userSvc,
					captchaService: captchaSvc,
					cfg:            UserConfig{CaptchaEnabled: true, AuthEmailEnabled: true},
					log:            zap.NewNop(),
				}
			},
			want: &oas.UserRegisterResponse{
				ID:       oas.Ulid("user-1"),
				Email:    oas.Email("john@example.com"),
				Username: oas.Username("john"),
			},
		},
		{
			name: "Captcha disabled skips check and accepts empty token",
			req: &oas.UserRegisterRequest{
				Email:    oas.Email("john@example.com"),
				Username: oas.Username("john"),
			},
			ctx: func() context.Context { return contextWithIP("0.0.0.0") },
			handler: func(t *testing.T) *User {
				t.Helper()

				captchaSvc := mocks.NewCaptchaServiceMock(t)

				userSvc := mocks.NewUserServiceMock(t)
				userSvc.EXPECT().
					Register(mock.Anything, mock.Anything, mock.Anything).
					Return(domain.User{
						ID:       "user-1",
						Email:    "john@example.com",
						Username: "john",
					}, nil)

				return &User{
					userService:    userSvc,
					captchaService: captchaSvc,
					cfg:            UserConfig{CaptchaEnabled: false, AuthEmailEnabled: true},
					log:            zap.NewNop(),
				}
			},
			want: &oas.UserRegisterResponse{
				ID:       oas.Ulid("user-1"),
				Email:    oas.Email("john@example.com"),
				Username: oas.Username("john"),
			},
			assertNoCaptcha: true,
		},
		{
			name: "Denied username returns BadRequest",
			req: &oas.UserRegisterRequest{
				Email:    oas.Email("admin@example.com"),
				Username: oas.Username("admin"),
			},
			ctx: context.Background,
			handler: func(t *testing.T) *User {
				t.Helper()

				return &User{cfg: UserConfig{AuthEmailEnabled: true}, log: zap.NewNop()}
			},
			want: &oas.UserRegisterBadRequest{
				Message: oas.NewOptString(domain.ErrUsernameNotAllowed.Error()),
			},
		},
		{
			name: "User already exists",
			req: &oas.UserRegisterRequest{
				Email:    oas.Email("john@example.com"),
				Username: oas.Username("john"),
			},
			ctx: context.Background,
			handler: func(t *testing.T) *User {
				t.Helper()

				captchaSvc := mocks.NewCaptchaServiceMock(t)
				captchaSvc.EXPECT().
					Check(mock.Anything, mock.Anything, mock.Anything).
					Return(true)

				userSvc := mocks.NewUserServiceMock(t)
				userSvc.EXPECT().
					Register(mock.Anything, mock.Anything, mock.Anything).
					Return(domain.User{}, domain.ErrUserAlreadyExists)

				return &User{
					userService:    userSvc,
					captchaService: captchaSvc,
					cfg:            UserConfig{CaptchaEnabled: true, AuthEmailEnabled: true},
					log:            zap.NewNop(),
				}
			},
			want: &oas.UserRegisterConflict{
				Message: oas.NewOptString(domain.ErrUserAlreadyExists.Error()),
			},
		},
		{
			name: "Service error",
			req: &oas.UserRegisterRequest{
				Email:    oas.Email("john@example.com"),
				Username: oas.Username("john"),
			},
			ctx: context.Background,
			handler: func(t *testing.T) *User {
				t.Helper()

				captchaSvc := mocks.NewCaptchaServiceMock(t)
				captchaSvc.EXPECT().
					Check(mock.Anything, mock.Anything, mock.Anything).
					Return(true)

				userSvc := mocks.NewUserServiceMock(t)
				userSvc.EXPECT().
					Register(mock.Anything, mock.Anything, mock.Anything).
					Return(domain.User{}, errors.New("service error"))

				return &User{
					userService:    userSvc,
					captchaService: captchaSvc,
					cfg:            UserConfig{CaptchaEnabled: true, AuthEmailEnabled: true},
					log:            zap.NewNop(),
				}
			},
			want: &oas.UserRegisterInternalServerError{},
		},
		{
			name: "Allowed email domain passes with flag enabled",
			req: &oas.UserRegisterRequest{
				Email:    oas.Email("john@example.com"),
				Username: oas.Username("john"),
			},
			ctx: func() context.Context { return contextWithIP("0.0.0.0") },
			handler: func(t *testing.T) *User {
				t.Helper()

				userSvc := mocks.NewUserServiceMock(t)
				userSvc.EXPECT().
					Register(mock.Anything, mock.Anything, mock.Anything).
					Return(domain.User{
						ID:       "user-1",
						Email:    "john@example.com",
						Username: "john",
					}, nil)

				return &User{
					userService:    userSvc,
					captchaService: mocks.NewCaptchaServiceMock(t),
					cfg: UserConfig{
						AuthEmailEnabled:            true,
						AllowedEmailDomainsEnabled:  true,
						AllowedEmailDomainsSuffixes: []string{"example.com"},
					},
					log: zap.NewNop(),
				}
			},
			want: &oas.UserRegisterResponse{
				ID:       oas.Ulid("user-1"),
				Email:    oas.Email("john@example.com"),
				Username: oas.Username("john"),
			},
		},
		{
			name: "Not allowed email domain returns BadRequest with flag enabled",
			req: &oas.UserRegisterRequest{
				Email:    oas.Email("john@badexample.com"),
				Username: oas.Username("john"),
			},
			ctx: context.Background,
			handler: func(t *testing.T) *User {
				t.Helper()

				return &User{
					captchaService: mocks.NewCaptchaServiceMock(t),
					cfg: UserConfig{
						AuthEmailEnabled:            true,
						AllowedEmailDomainsEnabled:  true,
						AllowedEmailDomainsSuffixes: []string{"example.com"},
					},
					log: zap.NewNop(),
				}
			},
			want: &oas.UserRegisterBadRequest{
				Message: oas.NewOptString(domain.ErrEmailDomainNotAllowed.Error()),
			},
		},
		{
			name: "Email domain not checked with flag disabled",
			req: &oas.UserRegisterRequest{
				Email:    oas.Email("john@badexample.com"),
				Username: oas.Username("john"),
			},
			ctx: func() context.Context { return contextWithIP("0.0.0.0") },
			handler: func(t *testing.T) *User {
				t.Helper()

				userSvc := mocks.NewUserServiceMock(t)
				userSvc.EXPECT().
					Register(mock.Anything, mock.Anything, mock.Anything).
					Return(domain.User{
						ID:       "user-1",
						Email:    "john@badexample.com",
						Username: "john",
					}, nil)

				return &User{
					userService:    userSvc,
					captchaService: mocks.NewCaptchaServiceMock(t),
					cfg: UserConfig{
						AuthEmailEnabled:            true,
						AllowedEmailDomainsEnabled:  false,
						AllowedEmailDomainsSuffixes: []string{"example.com"},
					},
					log: zap.NewNop(),
				}
			},
			want: &oas.UserRegisterResponse{
				ID:       oas.Ulid("user-1"),
				Email:    oas.Email("john@badexample.com"),
				Username: oas.Username("john"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := tt.handler(t)
			got, err := h.UserRegister(tt.ctx(), tt.req)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
			if tt.assertNoCaptcha {
				mockCaptcha, ok := h.captchaService.(*mocks.CaptchaServiceMock)
				assert.True(t, ok)
				assert.Empty(t, mockCaptcha.Calls)
			}
		})
	}
}

func TestUser_UserActivate(t *testing.T) {
	tests := []struct {
		name    string
		req     *oas.UserActivateRequest
		ctx     func() context.Context
		handler func(t *testing.T) *User
		want    oas.UserActivateRes
		wantErr error
	}{
		{
			name: "Success",
			req: &oas.UserActivateRequest{
				UserId: oas.Ulid("user-1"),
				Code:   "123456",
			},
			ctx: context.Background,
			handler: func(t *testing.T) *User {
				t.Helper()

				userSvc := mocks.NewUserServiceMock(t)
				userSvc.EXPECT().
					Activate(mock.Anything, domain.UserID("user-1"), "123456").
					Return(domain.CookieID("cookie-123"), nil)

				cookieSvc := mocks.NewCookieServiceMock(t)
				cookieSvc.EXPECT().
					Create(domain.CookieID("cookie-123")).
					Return(http.Cookie{Name: "session", Value: "cookie-123"})

				return &User{
					cfg:           UserConfig{AuthEmailEnabled: true},
					userService:   userSvc,
					cookieService: cookieSvc,
					log:           zap.NewNop(),
				}
			},
			want: &oas.UserActivateNoContent{
				SetCookie: oas.NewOptString("session=cookie-123"),
			},
		},
		{
			name: "User not found",
			req: &oas.UserActivateRequest{
				UserId: oas.Ulid("user-1"),
				Code:   "123456",
			},
			ctx: context.Background,
			handler: func(t *testing.T) *User {
				t.Helper()

				userSvc := mocks.NewUserServiceMock(t)
				userSvc.EXPECT().
					Activate(mock.Anything, domain.UserID("user-1"), "123456").
					Return("", domain.ErrUserNotFound)

				return &User{
					cfg:         UserConfig{AuthEmailEnabled: true},
					userService: userSvc,
					log:         zap.NewNop(),
				}
			},
			want: &oas.UserActivateNotFound{
				Message: oas.NewOptString(domain.ErrUserNotFound.Error()),
			},
		},
		{
			name: "Service error",
			req: &oas.UserActivateRequest{
				UserId: oas.Ulid("user-1"),
				Code:   "123456",
			},
			ctx: context.Background,
			handler: func(t *testing.T) *User {
				t.Helper()

				userSvc := mocks.NewUserServiceMock(t)
				userSvc.EXPECT().
					Activate(mock.Anything, domain.UserID("user-1"), "123456").
					Return("", errors.New("service error"))

				return &User{
					cfg:         UserConfig{AuthEmailEnabled: true},
					userService: userSvc,
					log:         zap.NewNop(),
				}
			},
			want: &oas.UserActivateInternalServerError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := tt.handler(t)
			got, err := h.UserActivate(tt.ctx(), tt.req)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestUser_UserLogin(t *testing.T) {
	tests := []struct {
		name            string
		req             *oas.UserLoginRequest
		ctx             func() context.Context
		handler         func(t *testing.T) *User
		want            oas.UserLoginRes
		wantErr         error
		assertNoCaptcha bool
	}{
		{
			name: "Success",
			req: &oas.UserLoginRequest{
				Email: oas.Email("john@example.com"),
			},
			ctx: context.Background,
			handler: func(t *testing.T) *User {
				t.Helper()

				captchaSvc := mocks.NewCaptchaServiceMock(t)
				captchaSvc.EXPECT().
					Check(mock.Anything, mock.Anything, mock.Anything).
					Return(true)

				userSvc := mocks.NewUserServiceMock(t)
				userSvc.EXPECT().
					Login(mock.Anything, "john@example.com").
					Return(nil)

				return &User{
					userService:    userSvc,
					captchaService: captchaSvc,
					cfg:            UserConfig{CaptchaEnabled: true, AuthEmailEnabled: true},
					log:            zap.NewNop(),
				}
			},
			want: &oas.UserLoginNoContent{},
		},
		{
			name: "Captcha disabled skips check and accepts empty token",
			req: &oas.UserLoginRequest{
				Email: oas.Email("john@example.com"),
			},
			ctx: context.Background,
			handler: func(t *testing.T) *User {
				t.Helper()

				captchaSvc := mocks.NewCaptchaServiceMock(t)

				userSvc := mocks.NewUserServiceMock(t)
				userSvc.EXPECT().
					Login(mock.Anything, "john@example.com").
					Return(nil)

				return &User{
					userService:    userSvc,
					captchaService: captchaSvc,
					cfg:            UserConfig{CaptchaEnabled: false, AuthEmailEnabled: true},
					log:            zap.NewNop(),
				}
			},
			want:            &oas.UserLoginNoContent{},
			assertNoCaptcha: true,
		},
		{
			name: "User not found",
			req: &oas.UserLoginRequest{
				Email: oas.Email("notfound@example.com"),
			},
			ctx: context.Background,
			handler: func(t *testing.T) *User {
				t.Helper()

				captchaSvc := mocks.NewCaptchaServiceMock(t)
				captchaSvc.EXPECT().
					Check(mock.Anything, mock.Anything, mock.Anything).
					Return(true)

				userSvc := mocks.NewUserServiceMock(t)
				userSvc.EXPECT().
					Login(mock.Anything, "notfound@example.com").
					Return(domain.ErrUserNotFound)

				return &User{
					userService:    userSvc,
					captchaService: captchaSvc,
					cfg:            UserConfig{CaptchaEnabled: true, AuthEmailEnabled: true},
					log:            zap.NewNop(),
				}
			},
			want: &oas.UserLoginNotFound{
				Message: oas.NewOptString(domain.ErrUserNotFound.Error()),
			},
		},
		{
			name: "Failed precondition",
			req: &oas.UserLoginRequest{
				Email: oas.Email("john@example.com"),
			},
			ctx: context.Background,
			handler: func(t *testing.T) *User {
				t.Helper()

				captchaSvc := mocks.NewCaptchaServiceMock(t)
				captchaSvc.EXPECT().
					Check(mock.Anything, mock.Anything, mock.Anything).
					Return(true)

				userSvc := mocks.NewUserServiceMock(t)
				userSvc.EXPECT().
					Login(mock.Anything, "john@example.com").
					Return(domain.NewFailedPreconditionError("user not activated"))

				return &User{
					userService:    userSvc,
					captchaService: captchaSvc,
					cfg:            UserConfig{CaptchaEnabled: true, AuthEmailEnabled: true},
					log:            zap.NewNop(),
				}
			},
			want: &oas.UserLoginBadRequest{
				Message: oas.NewOptString("user not activated"),
			},
		},
		{
			name: "Service error",
			req: &oas.UserLoginRequest{
				Email: oas.Email("john@example.com"),
			},
			ctx: context.Background,
			handler: func(t *testing.T) *User {
				t.Helper()

				captchaSvc := mocks.NewCaptchaServiceMock(t)
				captchaSvc.EXPECT().
					Check(mock.Anything, mock.Anything, mock.Anything).
					Return(true)

				userSvc := mocks.NewUserServiceMock(t)
				userSvc.EXPECT().
					Login(mock.Anything, "john@example.com").
					Return(errors.New("service error"))

				return &User{
					userService:    userSvc,
					captchaService: captchaSvc,
					cfg:            UserConfig{CaptchaEnabled: true, AuthEmailEnabled: true},
					log:            zap.NewNop(),
				}
			},
			want: &oas.UserLoginInternalServerError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := tt.handler(t)
			got, err := h.UserLogin(tt.ctx(), tt.req)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
			if tt.assertNoCaptcha {
				mockCaptcha, ok := h.captchaService.(*mocks.CaptchaServiceMock)
				assert.True(t, ok)
				assert.Empty(t, mockCaptcha.Calls)
			}
		})
	}
}

func TestUser_UserLoginVerify(t *testing.T) {
	tests := []struct {
		name    string
		req     *oas.UserLoginVerifyRequest
		ctx     func() context.Context
		handler func(t *testing.T) *User
		want    oas.UserLoginVerifyRes
		wantErr error
	}{
		{
			name: "Success",
			req: &oas.UserLoginVerifyRequest{
				Email: oas.Email("john@example.com"),
				Code:  "123456",
			},
			ctx: context.Background,
			handler: func(t *testing.T) *User {
				t.Helper()

				userSvc := mocks.NewUserServiceMock(t)
				userSvc.EXPECT().
					LoginVerify(mock.Anything, "john@example.com", "123456").
					Return(domain.CookieID("cookie-123"), nil)

				cookieSvc := mocks.NewCookieServiceMock(t)
				cookieSvc.EXPECT().
					Create(domain.CookieID("cookie-123")).
					Return(http.Cookie{Name: "session", Value: "cookie-123"})

				return &User{
					cfg:           UserConfig{AuthEmailEnabled: true},
					userService:   userSvc,
					cookieService: cookieSvc,
					log:           zap.NewNop(),
				}
			},
			want: &oas.UserLoginVerifyNoContent{
				SetCookie: oas.NewOptString("session=cookie-123"),
			},
		},
		{
			name: "User not found",
			req: &oas.UserLoginVerifyRequest{
				Email: oas.Email("notfound@example.com"),
				Code:  "123456",
			},
			ctx: context.Background,
			handler: func(t *testing.T) *User {
				t.Helper()

				userSvc := mocks.NewUserServiceMock(t)
				userSvc.EXPECT().
					LoginVerify(mock.Anything, "notfound@example.com", "123456").
					Return("", domain.ErrUserNotFound)

				return &User{
					cfg:         UserConfig{AuthEmailEnabled: true},
					userService: userSvc,
					log:         zap.NewNop(),
				}
			},
			want: &oas.UserLoginVerifyNotFound{
				Message: oas.NewOptString(domain.ErrUserNotFound.Error()),
			},
		},
		{
			name: "Service error",
			req: &oas.UserLoginVerifyRequest{
				Email: oas.Email("john@example.com"),
				Code:  "123456",
			},
			ctx: context.Background,
			handler: func(t *testing.T) *User {
				t.Helper()

				userSvc := mocks.NewUserServiceMock(t)
				userSvc.EXPECT().
					LoginVerify(mock.Anything, "john@example.com", "123456").
					Return("", errors.New("service error"))

				return &User{
					cfg:         UserConfig{AuthEmailEnabled: true},
					userService: userSvc,
					log:         zap.NewNop(),
				}
			},
			want: &oas.UserLoginVerifyInternalServerError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := tt.handler(t)
			got, err := h.UserLoginVerify(tt.ctx(), tt.req)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestUser_UserFake(t *testing.T) {
	tests := []struct {
		name    string
		ctx     func() context.Context
		handler func(t *testing.T) *User
		want    oas.UserFakeRes
		wantErr error
	}{
		{
			name: "Disabled flag returns NotFound",
			ctx:  func() context.Context { return contextWithIP("0.0.0.0") },
			handler: func(t *testing.T) *User {
				t.Helper()
				return &User{
					cfg: UserConfig{FakeAutoAccountsEnabled: false},
					log: zap.NewNop(),
				}
			},
			want: &oas.UserFakeNotFound{
				Message: oas.NewOptString("fake auto accounts disabled"),
			},
		},
		{
			name: "Success creates fake user and sets cookie",
			ctx:  func() context.Context { return contextWithIP("0.0.0.0") },
			handler: func(t *testing.T) *User {
				t.Helper()

				userSvc := mocks.NewUserServiceMock(t)
				userSvc.EXPECT().
					CreateFake(mock.Anything, domain.IP("0.0.0.0")).
					Return(domain.User{
						ID:       "user-1",
						Email:    "fake@localhost.tld",
						Username: "apple1234",
						IsFake:   true,
					}, domain.CookieID("cookie-1"), nil)

				cookieSvc := mocks.NewCookieServiceMock(t)
				cookieSvc.EXPECT().
					Create(domain.CookieID("cookie-1")).
					Return(http.Cookie{Name: "session", Value: "cookie-1", MaxAge: 31536000})

				return &User{
					cfg:           UserConfig{FakeAutoAccountsEnabled: true},
					userService:   userSvc,
					cookieService: cookieSvc,
					log:           zap.NewNop(),
				}
			},
			want: &oas.UserMeResponseHeaders{
				SetCookie: oas.NewOptString("session=cookie-1; Max-Age=31536000"),
				Response: oas.UserMeResponse{
					ID:       oas.Ulid("user-1"),
					Email:    oas.Email("fake@localhost.tld"),
					Username: oas.Username("apple1234"),
					IsFake:   true,
				},
			},
		},
		{
			name: "Service error returns InternalServerError",
			ctx:  func() context.Context { return contextWithIP("0.0.0.0") },
			handler: func(t *testing.T) *User {
				t.Helper()

				userSvc := mocks.NewUserServiceMock(t)
				userSvc.EXPECT().
					CreateFake(mock.Anything, domain.IP("0.0.0.0")).
					Return(domain.User{}, domain.CookieID(""), errors.New("create error"))

				return &User{
					cfg:         UserConfig{FakeAutoAccountsEnabled: true},
					userService: userSvc,
					log:         zap.NewNop(),
				}
			},
			want: &oas.UserFakeInternalServerError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := tt.handler(t)
			got, err := h.UserFake(tt.ctx())
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestUser_UserMe(t *testing.T) {
	tests := []struct {
		name    string
		ctx     func() context.Context
		handler func(t *testing.T) *User
		want    oas.UserMeRes
		wantErr error
	}{
		{
			name: "Success",
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *User {
				t.Helper()

				userSvc := mocks.NewUserServiceMock(t)
				userSvc.EXPECT().
					FindByID(mock.Anything, domain.UserID("user-1")).
					Return(&domain.User{
						ID:       "user-1",
						Email:    "john@example.com",
						Username: "john",
					}, nil)

				updater := mocks.NewSessionUpdaterMock(t)
				updater.EXPECT().
					UpdateTime(domain.SessionID("sess-1"))

				return &User{
					userService:    userSvc,
					sessionUpdater: updater,
					log:            zap.NewNop(),
				}
			},
			want: &oas.UserMeResponse{
				ID:       oas.Ulid("user-1"),
				Email:    oas.Email("john@example.com"),
				Username: oas.Username("john"),
			},
		},
		{
			name: "Success with fake user and TTL extension",
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *User {
				t.Helper()

				userSvc := mocks.NewUserServiceMock(t)
				userSvc.EXPECT().
					FindByID(mock.Anything, domain.UserID("user-1")).
					Return(&domain.User{
						ID:       "user-1",
						Email:    "fake@localhost.tld",
						Username: "apple1234",
						IsFake:   true,
					}, nil)

				updater := mocks.NewSessionUpdaterMock(t)
				updater.EXPECT().
					UpdateTime(domain.SessionID("sess-1"))

				return &User{
					userService:    userSvc,
					sessionUpdater: updater,
					log:            zap.NewNop(),
				}
			},
			want: &oas.UserMeResponse{
				ID:       oas.Ulid("user-1"),
				Email:    oas.Email("fake@localhost.tld"),
				Username: oas.Username("apple1234"),
				IsFake:   true,
			},
		},
		{
			name: "Unauthorized - no claims",
			ctx:  context.Background,
			handler: func(t *testing.T) *User {
				t.Helper()

				return &User{log: zap.NewNop()}
			},
			want: &oas.UserMeUnauthorized{},
		},
		{
			name: "User not found",
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *User {
				t.Helper()

				userSvc := mocks.NewUserServiceMock(t)
				userSvc.EXPECT().
					FindByID(mock.Anything, domain.UserID("user-1")).
					Return(nil, domain.ErrUserNotFound)

				return &User{
					userService: userSvc,
					log:         zap.NewNop(),
				}
			},
			want: &oas.UserMeUnauthorized{},
		},
		{
			name: "User not activated",
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *User {
				t.Helper()

				userSvc := mocks.NewUserServiceMock(t)
				activateCode := "123456"
				userSvc.EXPECT().
					FindByID(mock.Anything, domain.UserID("user-1")).
					Return(&domain.User{
						ID:           "user-1",
						Email:        "john@example.com",
						Username:     "john",
						ActivateCode: &activateCode,
					}, nil)

				return &User{
					userService: userSvc,
					log:         zap.NewNop(),
				}
			},
			want: &oas.UserMeForbidden{},
		},
		{
			name: "Service error",
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *User {
				t.Helper()

				userSvc := mocks.NewUserServiceMock(t)
				userSvc.EXPECT().
					FindByID(mock.Anything, domain.UserID("user-1")).
					Return(nil, errors.New("service error"))

				return &User{
					userService: userSvc,
					log:         zap.NewNop(),
				}
			},
			want: &oas.UserMeInternalServerError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := tt.handler(t)
			got, err := h.UserMe(tt.ctx())
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestUser_UserLogout(t *testing.T) {
	tests := []struct {
		name    string
		ctx     func() context.Context
		handler func(t *testing.T) *User
		want    oas.UserLogoutRes
		wantErr error
	}{
		{
			name: "Success",
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *User {
				t.Helper()

				sessionSvc := mocks.NewSessionServiceMock(t)
				sessionSvc.EXPECT().
					DeleteByID(mock.Anything, domain.SessionID("sess-1")).
					Return(nil)

				cookieSvc := mocks.NewCookieServiceMock(t)
				cookieSvc.EXPECT().
					Clear().
					Return(http.Cookie{Name: "session", Value: "", MaxAge: -1})

				return &User{
					sessionService: sessionSvc,
					cookieService:  cookieSvc,
					log:            zap.NewNop(),
				}
			},
			want: &oas.UserLogoutNoContent{
				SetCookie: oas.NewOptString("session=; Max-Age=0"),
			},
		},
		{
			name: "Unauthorized - no claims",
			ctx:  context.Background,
			handler: func(t *testing.T) *User {
				t.Helper()

				return &User{log: zap.NewNop()}
			},
			want: &oas.UserLogoutUnauthorized{},
		},
		{
			name: "Session delete error - still success",
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *User {
				t.Helper()

				sessionSvc := mocks.NewSessionServiceMock(t)
				sessionSvc.EXPECT().
					DeleteByID(mock.Anything, domain.SessionID("sess-1")).
					Return(errors.New("session error"))

				cookieSvc := mocks.NewCookieServiceMock(t)
				cookieSvc.EXPECT().
					Clear().
					Return(http.Cookie{Name: "session", Value: "", MaxAge: -1})

				return &User{
					sessionService: sessionSvc,
					cookieService:  cookieSvc,
					log:            zap.NewNop(),
				}
			},
			want: &oas.UserLogoutNoContent{
				SetCookie: oas.NewOptString("session=; Max-Age=0"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := tt.handler(t)
			got, err := h.UserLogout(tt.ctx())
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestUser_AuthEmailDisabled(t *testing.T) {
	forbidden := oas.NewOptString(domain.ErrOperationNotAllowed.Error())

	t.Run("Register", func(t *testing.T) {
		h := &User{
			userService:    mocks.NewUserServiceMock(t),
			captchaService: mocks.NewCaptchaServiceMock(t),
			log:            zap.NewNop(),
		}
		got, err := h.UserRegister(contextWithIP("0.0.0.0"), &oas.UserRegisterRequest{
			Email:    oas.Email("john@example.com"),
			Username: oas.Username("john"),
		})
		assert.NoError(t, err)
		assert.Equal(t, &oas.UserRegisterForbidden{Message: forbidden}, got)
	})

	t.Run("Activate", func(t *testing.T) {
		h := &User{
			userService: mocks.NewUserServiceMock(t),
			log:         zap.NewNop(),
		}
		got, err := h.UserActivate(context.Background(), &oas.UserActivateRequest{
			UserId: oas.Ulid("user-1"),
			Code:   "123456",
		})
		assert.NoError(t, err)
		assert.Equal(t, &oas.UserActivateForbidden{Message: forbidden}, got)
	})

	t.Run("Login", func(t *testing.T) {
		h := &User{
			userService:    mocks.NewUserServiceMock(t),
			captchaService: mocks.NewCaptchaServiceMock(t),
			log:            zap.NewNop(),
		}
		got, err := h.UserLogin(context.Background(), &oas.UserLoginRequest{
			Email: oas.Email("john@example.com"),
		})
		assert.NoError(t, err)
		assert.Equal(t, &oas.UserLoginForbidden{Message: forbidden}, got)
	})

	t.Run("LoginVerify", func(t *testing.T) {
		h := &User{
			userService: mocks.NewUserServiceMock(t),
			log:         zap.NewNop(),
		}
		got, err := h.UserLoginVerify(context.Background(), &oas.UserLoginVerifyRequest{
			Email: oas.Email("john@example.com"),
			Code:  "123456",
		})
		assert.NoError(t, err)
		assert.Equal(t, &oas.UserLoginVerifyForbidden{Message: forbidden}, got)
	})
}
