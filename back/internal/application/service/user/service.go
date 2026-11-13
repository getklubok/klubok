package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"klubok/internal/application/service/mailer"
	"klubok/internal/db/sqlc"
	"klubok/internal/domain"
	"klubok/internal/domain/faker"
)

const (
	authCodeLength      = 6
	authCodeExpires     = 1 * time.Hour
	authCodeExpiresText = "1 час"

	activateCodeLength      = 40 // См. тип поля users.activate_code в файле migrations/00001_init.sql
	activateCodeExpires     = 7 * 24 * time.Hour
	activateCodeExpiresText = "7 дней"

	cookieIDLength = 30 // См. тип поля sessions.cid в файле migrations/00001_init.sql

	privacyVersion = "1.0.0"

	fakeNicknameSuffixLength uint = 4
)

type Service struct {
	querier        Querier
	dbUseCase      DBUseCase
	sessionService SessionService
	emailCh        chan<- mailer.EmailTask
	ipMasker       domain.IPMasker

	log *zap.Logger

	nowFunc        func() time.Time
	idFunc         func() string
	randDigitFunc  func(l int) string
	randLetterFunc func(l int) string

	fakeIDFunc    func() string
	fakeWordFunc  func() string
	fakeDigitFunc func(l uint) string
}

func NewService(
	querier Querier,
	dbUseCase DBUseCase,
	sessionService SessionService,
	emailCh chan<- mailer.EmailTask,
	ipMasker domain.IPMasker,
	log *zap.Logger,
) *Service {
	return &Service{
		querier:        querier,
		dbUseCase:      dbUseCase,
		sessionService: sessionService,
		emailCh:        emailCh,
		ipMasker:       ipMasker,

		log: log.Named("user_service"),

		nowFunc:        time.Now,
		idFunc:         domain.NewID,
		randDigitFunc:  domain.NewDigitCode,
		randLetterFunc: domain.NewString,

		fakeIDFunc:    faker.ID,
		fakeWordFunc:  randomFakeWord,
		fakeDigitFunc: faker.DigitN,
	}
}

func (s *Service) Register(ctx context.Context, user domain.User, ip domain.IP) (domain.User, error) {
	_, err := s.FindByEmail(ctx, user.Email)
	if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
		return domain.User{}, fmt.Errorf("find user by email: %w", err)
	}
	if err == nil {
		return domain.User{}, domain.ErrUserAlreadyExists
	}

	activateCode := s.randLetterFunc(activateCodeLength)

	user.ID = domain.UserID(s.idFunc())
	user.ActivateCode = &activateCode

	u, err := s.dbUseCase.UserCreate(ctx, sqlc.UserCreateParams{
		ID:           user.ID.String(),
		Email:        user.Email,
		Username:     user.Username,
		ActivateCode: user.ActivateCode,
	}, sqlc.PrivacyLogCreateParams{
		UserID:         user.ID.String(),
		Ip:             s.ipMasker.Mask(ip.String()),
		PrivacyVersion: privacyVersion,
	})
	if err != nil {
		return domain.User{}, fmt.Errorf("create user: %w", err)
	}

	s.enqueueEmail(mailer.EmailTask{
		Recipients: []string{user.Email},
		Subject:    "Успешная регистрация",
		Template:   mailer.TemplateRegisterSuccess,
		Data: map[string]any{
			"sysAppURL":    "auto",
			"userID":       user.ID,
			"activateCode": activateCode,
			"timeString":   activateCodeExpiresText,
		},
	})

	return userToDomain(u), nil
}

func (s *Service) Activate(ctx context.Context, userID domain.UserID, code string) (cookieID domain.CookieID, err error) {
	log := s.log.With(
		zap.String("user_id", userID.String()),
	)

	outdatedTime := s.nowFunc().Add(-activateCodeExpires)
	if err := s.querier.UserDeleteNotActivated(ctx, outdatedTime); err != nil {
		return "", fmt.Errorf("cancel old: %w", err)
	}

	user, err := s.FindByID(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("find by id: %w", err)
	}

	if user.ActivateCode == nil {
		log.Warn("Activate code not found")
		return "", domain.ErrUserNotFound
	}
	if *user.ActivateCode != code {
		log.Warn("Invalid activate code")
		return "", domain.ErrUserNotFound
	}

	cookieID = domain.CookieID(s.randLetterFunc(cookieIDLength))

	session := domain.Session{
		ID:        domain.SessionID(s.idFunc()),
		CookieID:  cookieID,
		UserID:    user.ID,
		CheckedAt: s.nowFunc(),
	}
	if _, err := s.sessionService.Create(ctx, session); err != nil {
		return "", fmt.Errorf("create session: %w", err)
	}

	if err := s.querier.UserCancelActivateCode(ctx, user.ID.String()); err != nil {
		log.Error("Failed to cancel activate code", zap.Error(err))
	}

	return cookieID, nil
}

func (s *Service) Login(ctx context.Context, email string) error {
	log := s.log.With(
		zap.String("email", email),
	)

	user, err := s.FindByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("find by email: %w", err)
	}

	if user.ActivateCode != nil {
		log.Warn("User not activated")
		return domain.NewFailedPreconditionError("user not activated") //nolint:wrapcheck // It's ok
	}

	code := s.randDigitFunc(authCodeLength)

	err = s.querier.AuthCodeUpsert(ctx, sqlc.AuthCodeUpsertParams{
		ID:        s.idFunc(),
		UserID:    user.ID.String(),
		UserEmail: user.Email,
		Code:      code,
	})
	if err != nil {
		return fmt.Errorf("upsert auth code: %w", err)
	}

	s.enqueueEmail(mailer.EmailTask{
		Recipients: []string{user.Email},
		Subject:    "Код авторизации",
		Template:   mailer.TemplateAuthCode,
		Data: map[string]any{
			"code":       code,
			"timeString": authCodeExpiresText,
		},
	})

	return nil
}

func (s *Service) LoginVerify(ctx context.Context, email string, code string) (cookieID domain.CookieID, err error) {
	outdatedTime := s.nowFunc().Add(-authCodeExpires)
	if err := s.querier.AuthCodeCancelOld(ctx, outdatedTime); err != nil {
		return "", fmt.Errorf("cancel old: %w", err)
	}

	authCode, err := s.querier.AuthCodeFind(ctx, email, code)
	if err != nil {
		// Не проверяем domain.ErrNotFound
		return "", fmt.Errorf("find auth code: %w", err)
	}

	cookieID = domain.CookieID(s.randLetterFunc(cookieIDLength))

	session := domain.Session{
		ID:        domain.SessionID(s.idFunc()),
		CookieID:  cookieID,
		UserID:    domain.UserID(authCode.UserID),
		CheckedAt: s.nowFunc(),
	}
	if _, err := s.sessionService.Create(ctx, session); err != nil {
		return "", fmt.Errorf("create session: %w", err)
	}

	if err := s.querier.AuthCodeDelete(ctx, authCode.ID); err != nil {
		s.log.Error("Failed to delete auth code", zap.Error(err))
	}

	return cookieID, nil
}

// LoginSSO находит или создаёт пользователя по данным SSO и открывает сессию.
func (s *Service) LoginSSO(
	ctx context.Context,
	issuer string,
	sub string,
	email string,
	preferredUsername string,
	ip domain.IP,
) (cookieID domain.CookieID, err error) {
	user, err := s.FindOrCreateBySSO(ctx, issuer, sub, email, preferredUsername, ip)
	if err != nil {
		return "", fmt.Errorf("find or create by sso: %w", err)
	}

	cookieID = domain.CookieID(s.randLetterFunc(cookieIDLength))

	session := domain.Session{
		ID:        domain.SessionID(s.idFunc()),
		CookieID:  cookieID,
		UserID:    user.ID,
		CheckedAt: s.nowFunc(),
	}
	if _, err := s.sessionService.Create(ctx, session); err != nil {
		return "", fmt.Errorf("create session: %w", err)
	}

	return cookieID, nil
}

func (s *Service) Update(ctx context.Context, user domain.User) (domain.User, error) {
	u, err := s.querier.UserUpdate(ctx, sqlc.UserUpdateParams{
		ID:       user.ID.String(),
		Email:    user.Email,
		Username: user.Username,
	})
	if err != nil {
		return domain.User{}, fmt.Errorf("update user: %w", err)
	}

	return userToDomain(u), nil
}

func (s *Service) FindByID(ctx context.Context, id domain.UserID) (*domain.User, error) {
	u, err := s.querier.UserFindByID(ctx, id.String())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}

		return nil, fmt.Errorf("find user by id: %w", err)
	}

	res := userToDomain(u)

	return &res, nil
}

func (s *Service) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	u, err := s.querier.UserFindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}

		return nil, fmt.Errorf("find user by email: %w", err)
	}

	res := userToDomain(u)

	return &res, nil
}

func (s *Service) FindByName(ctx context.Context, name string) (*domain.User, error) {
	u, err := s.querier.UserFindByName(ctx, name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}

		return nil, fmt.Errorf("find user by name: %w", err)
	}

	res := userToDomain(u)

	return &res, nil
}

func (s *Service) FindOrCreateBySSO(ctx context.Context, issuer string, sub string, email string, preferredUsername string, ip domain.IP) (domain.User, error) {
	identity, err := s.querier.UserSSOIdentityFind(ctx, issuer, sub)
	if err == nil {
		user, err := s.FindByID(ctx, domain.UserID(identity.UserID))
		if err != nil {
			return domain.User{}, fmt.Errorf("find user by sso identity: %w", err)
		}

		return *user, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return domain.User{}, fmt.Errorf("find sso identity: %w", err)
	}

	if user, err := s.FindByEmail(ctx, email); err == nil {
		if _, err := s.querier.UserSSOIdentityCreate(ctx, sqlc.UserSSOIdentityCreateParams{
			ID:     s.idFunc(),
			UserID: user.ID.String(),
			Issuer: issuer,
			Sub:    sub,
		}); err != nil {
			if domain.IsDBUniqueViolationError(err) {
				return s.findBySSOIdentity(ctx, issuer, sub)
			}

			return domain.User{}, fmt.Errorf("attach sso identity: %w", err)
		}

		return *user, nil
	} else if !errors.Is(err, domain.ErrUserNotFound) {
		return domain.User{}, fmt.Errorf("find user by email: %w", err)
	}

	username := sanitizeSSOUsername(preferredUsername, email)

	user, err := s.createSSOAttempt(ctx, email, username, issuer, sub, ip)
	if errors.Is(err, domain.ErrUserAlreadyExists) {
		user, err = s.createSSOAttempt(ctx, email, username+s.fakeDigitFunc(fakeNicknameSuffixLength), issuer, sub, ip)
	}
	if err != nil {
		return domain.User{}, fmt.Errorf("create sso user: %w", err)
	}

	return user, nil
}

// sanitizeSSOUsername приводит preferred_username к виду ^[a-z0-9]+$:
// уже подходящее имя только lower-ится, иначе берётся локальная часть email
// с вычищенными недопустимыми символами. Пустой результат заменяется "user".
func sanitizeSSOUsername(preferredUsername string, email string) string {
	if lower := strings.ToLower(preferredUsername); isSSOUsernameValid(lower) {
		return lower
	}

	local := email
	if i := strings.LastIndex(email, "@"); i >= 0 {
		local = email[:i]
	}

	var b strings.Builder
	for _, r := range strings.ToLower(local) {
		if isSSOUsernameChar(r) {
			b.WriteRune(r)
		}
	}

	if b.Len() == 0 {
		return "user"
	}

	return b.String()
}

func isSSOUsernameValid(s string) bool {
	if s == "" {
		return false
	}

	for _, r := range s {
		if !isSSOUsernameChar(r) {
			return false
		}
	}

	return true
}

func isSSOUsernameChar(r rune) bool {
	return r >= 'a' && r <= 'z' || r >= '0' && r <= '9'
}

func (s *Service) CreateFake(ctx context.Context, ip domain.IP) (domain.User, domain.CookieID, error) {
	email := s.generateFakeEmail()
	nickname := s.generateFakeNickname()

	user, cookieID, err := s.createFakeAttempt(ctx, email, nickname, ip)
	if errors.Is(err, domain.ErrUserAlreadyExists) {
		email = s.generateFakeEmail()
		nickname = s.generateFakeNickname()

		user, cookieID, err = s.createFakeAttempt(ctx, email, nickname, ip)
	}
	if err != nil {
		return domain.User{}, "", fmt.Errorf("create fake user: %w", err)
	}

	return user, cookieID, nil
}

func (s *Service) createFakeAttempt(ctx context.Context, email string, nickname string, ip domain.IP) (domain.User, domain.CookieID, error) {
	id := domain.UserID(s.idFunc())

	u, err := s.dbUseCase.UserCreate(ctx, sqlc.UserCreateParams{
		ID:       id.String(),
		Email:    email,
		Username: nickname,
		IsFake:   true,
	}, sqlc.PrivacyLogCreateParams{
		UserID:         id.String(),
		Ip:             s.ipMasker.Mask(ip.String()),
		PrivacyVersion: privacyVersion,
	})
	if err != nil {
		return domain.User{}, "", fmt.Errorf("create user: %w", err)
	}

	cookieID := domain.CookieID(s.randLetterFunc(cookieIDLength))

	session := domain.Session{
		ID:        domain.SessionID(s.idFunc()),
		CookieID:  cookieID,
		UserID:    id,
		CheckedAt: s.nowFunc(),
	}
	if _, err := s.sessionService.Create(ctx, session); err != nil {
		return domain.User{}, "", fmt.Errorf("create session: %w", err)
	}

	return userToDomain(u), cookieID, nil
}

func (s *Service) findBySSOIdentity(ctx context.Context, issuer string, sub string) (domain.User, error) {
	identity, err := s.querier.UserSSOIdentityFind(ctx, issuer, sub)
	if err != nil {
		return domain.User{}, fmt.Errorf("find sso identity: %w", err)
	}

	user, err := s.FindByID(ctx, domain.UserID(identity.UserID))
	if err != nil {
		return domain.User{}, fmt.Errorf("find user by sso identity: %w", err)
	}

	return *user, nil
}

func (s *Service) createSSOAttempt(ctx context.Context, email string, username string, issuer string, sub string, ip domain.IP) (domain.User, error) {
	id := domain.UserID(s.idFunc())

	u, err := s.dbUseCase.UserCreateWithSSO(ctx, sqlc.UserCreateParams{
		ID:       id.String(),
		Email:    email,
		Username: username,
	}, sqlc.PrivacyLogCreateParams{
		UserID:         id.String(),
		Ip:             s.ipMasker.Mask(ip.String()),
		PrivacyVersion: privacyVersion,
	}, sqlc.UserSSOIdentityCreateParams{
		ID:     s.idFunc(),
		UserID: id.String(),
		Issuer: issuer,
		Sub:    sub,
	})
	if err != nil {
		return domain.User{}, err //nolint:wrapcheck // Caller adds context, keeps errors.Is working
	}

	return userToDomain(u), nil
}

func (s *Service) generateFakeEmail() string {
	return s.fakeIDFunc() + "@localhost.tld"
}

func (s *Service) generateFakeNickname() string {
	word := strings.ToLower(s.fakeWordFunc())
	word = strings.ReplaceAll(word, " ", "")

	return word + s.fakeDigitFunc(fakeNicknameSuffixLength)
}

func randomFakeWord() string {
	generators := []func() string{
		faker.Fruit,
		faker.Vegetable,
		faker.Animal,
		faker.Bird,
	}

	return generators[faker.Number(0, len(generators)-1)]()
}

func (s *Service) enqueueEmail(task mailer.EmailTask) {
	select {
	case s.emailCh <- task:
	default:
		s.log.Warn("Email channel full, dropping email",
			zap.String("subject", task.Subject),
			zap.Strings("recipients", task.Recipients),
		)
	}
}

func userToDomain(u sqlc.User) domain.User {
	return domain.User{
		ID:           domain.UserID(u.ID),
		Email:        u.Email,
		Username:     u.Username,
		ActivateCode: u.ActivateCode,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
		IsFake:       u.IsFake,
	}
}
