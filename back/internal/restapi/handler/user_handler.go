package handler

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"klubok/internal/application/zapctx"
	"klubok/internal/domain"
	"klubok/internal/restapi/oas"
)

// UserConfig конфигурация фич, влияющих на обработчик пользователей.
type UserConfig struct {
	FakeAutoAccountsEnabled     bool
	AuthEmailEnabled            bool
	CaptchaEnabled              bool
	AllowedEmailDomainsEnabled  bool
	AllowedEmailDomainsSuffixes []string
}

type User struct {
	cfg UserConfig

	userService    UserService
	cookieService  CookieService
	sessionService SessionService
	captchaService CaptchaService
	sessionUpdater SessionUpdater

	log *zap.Logger
}

func NewUser(
	cfg UserConfig,
	userService UserService,
	cookieService CookieService,
	sessionService SessionService,
	captchaService CaptchaService,
	sessionUpdater SessionUpdater,
	log *zap.Logger,
) *User {
	return &User{
		cfg: cfg,

		userService:    userService,
		cookieService:  cookieService,
		sessionService: sessionService,
		captchaService: captchaService,
		sessionUpdater: sessionUpdater,

		log: log.Named("user_handler"),
	}
}

// Register user.
//
// POST /users/register.
func (h *User) UserRegister(ctx context.Context, req *oas.UserRegisterRequest) (oas.UserRegisterRes, error) { //nolint:unparam // It's ok
	if h.cfg.FakeAutoAccountsEnabled {
		h.log.Info("Fake auto accounts enabled")
		return &oas.UserRegisterForbidden{
			Message: oas.NewOptString(domain.ErrOperationNotAllowed.Error()),
		}, nil
	}

	if !h.cfg.AuthEmailEnabled {
		h.log.Info("Email login disabled")
		return &oas.UserRegisterForbidden{
			Message: oas.NewOptString(domain.ErrOperationNotAllowed.Error()),
		}, nil
	}

	if h.cfg.AllowedEmailDomainsEnabled && !domain.EmailDomainAllowed(string(req.GetEmail()), h.cfg.AllowedEmailDomainsSuffixes) {
		return &oas.UserRegisterBadRequest{
			Message: oas.NewOptString(domain.ErrEmailDomainNotAllowed.Error()),
		}, nil
	}

	if domain.IsUsernameDenied(string(req.GetUsername())) {
		return &oas.UserRegisterBadRequest{
			Message: oas.NewOptString(domain.ErrUsernameNotAllowed.Error()),
		}, nil
	}

	ip := domain.IPFromContext(ctx)

	if h.cfg.CaptchaEnabled && !h.captchaService.Check(ctx, req.CaptchaToken, ip) {
		return &oas.UserRegisterBadRequest{
			Message: oas.NewOptString("invalid captcha"),
		}, nil
	}

	input := userRegisterParamsToDomain(req)

	log := zapctx.Grasp(ctx, h.log).With(
		zap.Any("input", input),
	)

	user, err := h.userService.Register(ctx, input, ip)
	if err != nil {
		if errors.Is(err, domain.ErrUserAlreadyExists) {
			log.Warn("User already exists")
			return &oas.UserRegisterConflict{
				Message: oas.NewOptString(domain.ErrUserAlreadyExists.Error()),
			}, nil
		}

		log.Error("Create user",
			zap.Error(err),
		)
		return &oas.UserRegisterInternalServerError{}, nil
	}

	return registerResponseFromDomain(user), nil
}

// Activate user.
//
// POST /users/activate.
func (h *User) UserActivate(ctx context.Context, req *oas.UserActivateRequest) (oas.UserActivateRes, error) { //nolint:unparam // It's ok
	if h.cfg.FakeAutoAccountsEnabled {
		h.log.Info("Fake auto accounts enabled")
		return &oas.UserActivateForbidden{
			Message: oas.NewOptString(domain.ErrOperationNotAllowed.Error()),
		}, nil
	}

	if !h.cfg.AuthEmailEnabled {
		h.log.Info("Email login disabled")
		return &oas.UserActivateForbidden{
			Message: oas.NewOptString(domain.ErrOperationNotAllowed.Error()),
		}, nil
	}

	userID := domain.UserID(req.UserId)
	code := req.Code

	log := zapctx.Grasp(ctx, h.log).With(
		zap.String("user_id", userID.String()),
	)

	cookieID, err := h.userService.Activate(ctx, userID, code)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			log.Warn("User not found")
			return &oas.UserActivateNotFound{
				Message: oas.NewOptString(domain.ErrUserNotFound.Error()),
			}, nil
		}

		log.Error("Activate user",
			zap.Error(err),
		)
		return &oas.UserActivateInternalServerError{}, nil
	}

	cookie := h.cookieService.Create(cookieID)

	return &oas.UserActivateNoContent{
		SetCookie: oas.NewOptString(cookie.String()),
	}, nil
}

// Login user.
//
// POST /users/login.
func (h *User) UserLogin(ctx context.Context, req *oas.UserLoginRequest) (oas.UserLoginRes, error) { //nolint:unparam // It's ok
	if h.cfg.FakeAutoAccountsEnabled {
		h.log.Info("Fake auto accounts enabled")
		return &oas.UserLoginForbidden{
			Message: oas.NewOptString(domain.ErrOperationNotAllowed.Error()),
		}, nil
	}

	if !h.cfg.AuthEmailEnabled {
		h.log.Info("Email login disabled")
		return &oas.UserLoginForbidden{
			Message: oas.NewOptString(domain.ErrOperationNotAllowed.Error()),
		}, nil
	}

	ip := domain.IPFromContext(ctx)
	if h.cfg.CaptchaEnabled && !h.captchaService.Check(ctx, req.CaptchaToken, ip) {
		return &oas.UserLoginBadRequest{
			Message: oas.NewOptString("invalid captcha"),
		}, nil
	}

	email := string(req.Email)

	log := zapctx.Grasp(ctx, h.log).With(
		zap.String("email", email),
	)

	if err := h.userService.Login(ctx, email); err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			log.Warn("User not found")
			return &oas.UserLoginNotFound{
				Message: oas.NewOptString(domain.ErrUserNotFound.Error()),
			}, nil
		}
		if domain.IsFailedPreconditionError(err) {
			log.Warn("Failed precondition",
				zap.Error(err),
			)
			return &oas.UserLoginBadRequest{
				Message: oas.NewOptString(err.Error()),
			}, nil
		}

		log.Error("Login user",
			zap.Error(err),
		)
		return &oas.UserLoginInternalServerError{}, nil
	}

	return &oas.UserLoginNoContent{}, nil
}

// Login verify user.
//
// POST /users/login/verify.
func (h *User) UserLoginVerify(ctx context.Context, req *oas.UserLoginVerifyRequest) (oas.UserLoginVerifyRes, error) { //nolint:unparam // It's ok
	if h.cfg.FakeAutoAccountsEnabled {
		h.log.Info("Fake auto accounts enabled")
		return &oas.UserLoginVerifyForbidden{
			Message: oas.NewOptString(domain.ErrOperationNotAllowed.Error()),
		}, nil
	}

	if !h.cfg.AuthEmailEnabled {
		h.log.Info("Email login disabled")
		return &oas.UserLoginVerifyForbidden{
			Message: oas.NewOptString(domain.ErrOperationNotAllowed.Error()),
		}, nil
	}

	email := string(req.Email)
	code := req.Code

	log := zapctx.Grasp(ctx, h.log).With(
		zap.String("email", email),
	)

	cookieID, err := h.userService.LoginVerify(ctx, email, code)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			log.Warn("User not found")
			return &oas.UserLoginVerifyNotFound{
				Message: oas.NewOptString(domain.ErrUserNotFound.Error()),
			}, nil
		}

		log.Error("Login verify user",
			zap.Error(err),
		)
		return &oas.UserLoginVerifyInternalServerError{}, nil
	}

	cookie := h.cookieService.Create(cookieID)

	return &oas.UserLoginVerifyNoContent{
		SetCookie: oas.NewOptString(cookie.String()),
	}, nil
}

// Create fake user.
//
// POST /users/fake.
func (h *User) UserFake(ctx context.Context) (oas.UserFakeRes, error) { //nolint:unparam // It's ok
	log := zapctx.Grasp(ctx, h.log)

	if !h.cfg.FakeAutoAccountsEnabled {
		log.Info("Fake auto accounts disabled")
		return &oas.UserFakeNotFound{
			Message: oas.NewOptString("fake auto accounts disabled"),
		}, nil
	}

	ip := domain.IPFromContext(ctx)

	user, cookieID, err := h.userService.CreateFake(ctx, ip)
	if err != nil {
		log.Error("Create fake user",
			zap.Error(err),
		)
		return &oas.UserFakeInternalServerError{}, nil
	}

	cookie := h.cookieService.Create(cookieID)

	return &oas.UserMeResponseHeaders{
		SetCookie: oas.NewOptString(cookie.String()),
		Response:  *meResponseFromDomain(user),
	}, nil
}

// Get current user.
//
// GET /users/me.
func (h *User) UserMe(ctx context.Context) (oas.UserMeRes, error) { //nolint:unparam // It's ok
	claims, err := domain.ClaimsFromContext(ctx)
	if err != nil {
		zapctx.Grasp(ctx, h.log).Error("Get claims from context",
			zap.Error(err),
		)
		return &oas.UserMeUnauthorized{}, nil
	}

	log := zapctx.Grasp(ctx, h.log).With(
		zap.String("user_id", claims.UserID.String()),
	)

	user, err := h.userService.FindByID(ctx, claims.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			log.Warn("User not found")
			return &oas.UserMeUnauthorized{}, nil
		}

		log.Error("Find user by id",
			zap.Error(err),
		)
		return &oas.UserMeInternalServerError{}, nil
	}

	if user.ActivateCode != nil {
		log.Warn("User not activated")
		return &oas.UserMeForbidden{}, nil
	}

	if h.sessionUpdater != nil {
		h.sessionUpdater.UpdateTime(claims.SessionID)
	}

	return meResponseFromDomain(*user), nil
}

// Logout current user.
//
// DELETE /users/logout.
func (h *User) UserLogout(ctx context.Context) (oas.UserLogoutRes, error) { //nolint:unparam // It's ok
	if h.cfg.FakeAutoAccountsEnabled {
		h.log.Info("Fake auto accounts enabled")
		return &oas.UserLogoutForbidden{
			Message: oas.NewOptString(domain.ErrOperationNotAllowed.Error()),
		}, nil
	}

	claims, err := domain.ClaimsFromContext(ctx)
	if err != nil {
		zapctx.Grasp(ctx, h.log).Error("Get claims from context",
			zap.Error(err),
		)
		return &oas.UserLogoutUnauthorized{}, nil
	}

	if err := h.sessionService.DeleteByID(ctx, claims.SessionID); err != nil {
		zapctx.Grasp(ctx, h.log).Error("Logout user",
			zap.String("user_id", claims.UserID.String()),
			zap.Error(err),
		)
	}

	cookie := h.cookieService.Clear()

	return &oas.UserLogoutNoContent{
		SetCookie: oas.NewOptString(cookie.String()),
	}, nil
}
