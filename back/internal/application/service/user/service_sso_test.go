package user

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"klubok/internal/application/service/user/mocks"
	"klubok/internal/db/sqlc"
	"klubok/internal/domain"
)

const (
	testSSOIssuer = "https://issuer.example.com"
	testSSOSub    = "sub-123"
)

func newSSOService(q Querier, db DBUseCase) *Service {
	return &Service{
		querier:       q,
		dbUseCase:     db,
		ipMasker:      domain.IPMasker{},
		log:           zap.NewNop(),
		idFunc:        func() string { return "user-id" },
		fakeDigitFunc: func(_ uint) string { return "1234" },
	}
}

func testSSOUser() sqlc.User {
	return sqlc.User{
		ID:        "user-1",
		Email:     "bob@example.com",
		Username:  "bob1",
		CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}

func TestService_FindOrCreateBySSO(t *testing.T) {
	tests := []struct {
		name              string
		preferredUsername string
		email             string
		svc               func(t *testing.T) *Service
		want              domain.User
		wantErr           error
	}{
		{
			name:              "Existing identity returns user",
			preferredUsername: "bob1",
			email:             "bob@example.com",
			svc: func(t *testing.T) *Service {
				t.Helper()
				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					UserSSOIdentityFind(mock.Anything, testSSOIssuer, testSSOSub).
					Return(sqlc.UserSsoIdentity{UserID: "user-1", Issuer: testSSOIssuer, Sub: testSSOSub}, nil)
				q.EXPECT().
					UserFindByID(mock.Anything, "user-1").
					Return(testSSOUser(), nil)
				return newSSOService(q, mocks.NewDBUseCaseMock(t))
			},
			want: userToDomain(testSSOUser()),
		},
		{
			name:              "Known email attaches identity",
			preferredUsername: "bob1",
			email:             "bob@example.com",
			svc: func(t *testing.T) *Service {
				t.Helper()
				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					UserSSOIdentityFind(mock.Anything, testSSOIssuer, testSSOSub).
					Return(sqlc.UserSsoIdentity{}, sql.ErrNoRows)
				q.EXPECT().
					UserFindByEmail(mock.Anything, "bob@example.com").
					Return(testSSOUser(), nil)
				q.EXPECT().
					UserSSOIdentityCreate(mock.Anything, sqlc.UserSSOIdentityCreateParams{
						ID:     "user-id",
						UserID: "user-1",
						Issuer: testSSOIssuer,
						Sub:    testSSOSub,
					}).
					Return(sqlc.UserSsoIdentity{}, nil)
				return newSSOService(q, mocks.NewDBUseCaseMock(t))
			},
			want: userToDomain(testSSOUser()),
		},
		{
			name:              "Attach race re-reads identity",
			preferredUsername: "bob1",
			email:             "bob@example.com",
			svc: func(t *testing.T) *Service {
				t.Helper()
				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					UserSSOIdentityFind(mock.Anything, testSSOIssuer, testSSOSub).
					Return(sqlc.UserSsoIdentity{}, sql.ErrNoRows).
					Once()
				q.EXPECT().
					UserFindByEmail(mock.Anything, "bob@example.com").
					Return(testSSOUser(), nil)
				q.EXPECT().
					UserSSOIdentityCreate(mock.Anything, mock.Anything).
					Return(sqlc.UserSsoIdentity{}, &pgconn.PgError{Code: "23505"})
				q.EXPECT().
					UserSSOIdentityFind(mock.Anything, testSSOIssuer, testSSOSub).
					Return(sqlc.UserSsoIdentity{UserID: "user-1", Issuer: testSSOIssuer, Sub: testSSOSub}, nil).
					Once()
				q.EXPECT().
					UserFindByID(mock.Anything, "user-1").
					Return(testSSOUser(), nil)
				return newSSOService(q, mocks.NewDBUseCaseMock(t))
			},
			want: userToDomain(testSSOUser()),
		},
		{
			name:              "New user created with sanitized username",
			preferredUsername: "Dim.Nik",
			email:             "Bob.Smith+1@example.com",
			svc: func(t *testing.T) *Service {
				t.Helper()
				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					UserSSOIdentityFind(mock.Anything, testSSOIssuer, testSSOSub).
					Return(sqlc.UserSsoIdentity{}, sql.ErrNoRows)
				q.EXPECT().
					UserFindByEmail(mock.Anything, "Bob.Smith+1@example.com").
					Return(sqlc.User{}, sql.ErrNoRows)

				db := mocks.NewDBUseCaseMock(t)
				db.EXPECT().
					UserCreateWithSSO(mock.Anything, sqlc.UserCreateParams{
						ID:       "user-id",
						Email:    "Bob.Smith+1@example.com",
						Username: "bobsmith1",
					}, sqlc.PrivacyLogCreateParams{
						UserID:         "user-id",
						Ip:             "0.0.0.0",
						PrivacyVersion: privacyVersion,
					}, sqlc.UserSSOIdentityCreateParams{
						ID:     "user-id",
						UserID: "user-id",
						Issuer: testSSOIssuer,
						Sub:    testSSOSub,
					}).
					Return(sqlc.User{
						ID:        "user-id",
						Email:     "Bob.Smith+1@example.com",
						Username:  "bobsmith1",
						CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						UpdatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					}, nil)
				return newSSOService(q, db)
			},
			want: domain.User{
				ID:        domain.UserID("user-id"),
				Email:     "Bob.Smith+1@example.com",
				Username:  "bobsmith1",
				CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name:              "Valid preferred username is kept",
			preferredUsername: "Dimnik99",
			email:             "other@example.com",
			svc: func(t *testing.T) *Service {
				t.Helper()
				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					UserSSOIdentityFind(mock.Anything, testSSOIssuer, testSSOSub).
					Return(sqlc.UserSsoIdentity{}, sql.ErrNoRows)
				q.EXPECT().
					UserFindByEmail(mock.Anything, "other@example.com").
					Return(sqlc.User{}, sql.ErrNoRows)

				db := mocks.NewDBUseCaseMock(t)
				db.EXPECT().
					UserCreateWithSSO(mock.Anything, sqlc.UserCreateParams{
						ID:       "user-id",
						Email:    "other@example.com",
						Username: "dimnik99",
					}, mock.Anything, mock.Anything).
					Return(sqlc.User{
						ID:        "user-id",
						Email:     "other@example.com",
						Username:  "dimnik99",
						CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						UpdatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					}, nil)
				return newSSOService(q, db)
			},
			want: domain.User{
				ID:        domain.UserID("user-id"),
				Email:     "other@example.com",
				Username:  "dimnik99",
				CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name:              "Retry with suffix on username conflict",
			preferredUsername: "bob1",
			email:             "new@example.com",
			svc: func(t *testing.T) *Service {
				t.Helper()
				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					UserSSOIdentityFind(mock.Anything, testSSOIssuer, testSSOSub).
					Return(sqlc.UserSsoIdentity{}, sql.ErrNoRows)
				q.EXPECT().
					UserFindByEmail(mock.Anything, "new@example.com").
					Return(sqlc.User{}, sql.ErrNoRows)

				db := mocks.NewDBUseCaseMock(t)
				db.EXPECT().
					UserCreateWithSSO(mock.Anything, sqlc.UserCreateParams{
						ID:       "user-id",
						Email:    "new@example.com",
						Username: "bob1",
					}, mock.Anything, mock.Anything).
					Return(sqlc.User{}, domain.ErrUserAlreadyExists)
				db.EXPECT().
					UserCreateWithSSO(mock.Anything, sqlc.UserCreateParams{
						ID:       "user-id",
						Email:    "new@example.com",
						Username: "bob11234",
					}, mock.Anything, mock.Anything).
					Return(sqlc.User{
						ID:        "user-id",
						Email:     "new@example.com",
						Username:  "bob11234",
						CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						UpdatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					}, nil)
				return newSSOService(q, db)
			},
			want: domain.User{
				ID:        domain.UserID("user-id"),
				Email:     "new@example.com",
				Username:  "bob11234",
				CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name:              "Second conflict returns error",
			preferredUsername: "bob1",
			email:             "new@example.com",
			svc: func(t *testing.T) *Service {
				t.Helper()
				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					UserSSOIdentityFind(mock.Anything, testSSOIssuer, testSSOSub).
					Return(sqlc.UserSsoIdentity{}, sql.ErrNoRows)
				q.EXPECT().
					UserFindByEmail(mock.Anything, "new@example.com").
					Return(sqlc.User{}, sql.ErrNoRows)

				db := mocks.NewDBUseCaseMock(t)
				db.EXPECT().
					UserCreateWithSSO(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(sqlc.User{}, domain.ErrUserAlreadyExists)
				db.EXPECT().
					UserCreateWithSSO(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(sqlc.User{}, domain.ErrUserAlreadyExists)
				return newSSOService(q, db)
			},
			wantErr: errors.New("create sso user: user already exists"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.svc(t)
			got, err := svc.FindOrCreateBySSO(context.Background(), testSSOIssuer, testSSOSub, tt.email, tt.preferredUsername, "0.0.0.0")
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestService_LoginSSO(t *testing.T) {
	newSvc := func(t *testing.T, sessionSvc SessionService) *Service {
		t.Helper()

		q := mocks.NewQuerierMock(t)
		q.EXPECT().
			UserSSOIdentityFind(mock.Anything, testSSOIssuer, testSSOSub).
			Return(sqlc.UserSsoIdentity{UserID: "user-1", Issuer: testSSOIssuer, Sub: testSSOSub}, nil)
		q.EXPECT().
			UserFindByID(mock.Anything, "user-1").
			Return(testSSOUser(), nil)

		svc := newSSOService(q, mocks.NewDBUseCaseMock(t))
		svc.sessionService = sessionSvc
		svc.nowFunc = func() time.Time { return time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC) }
		svc.randLetterFunc = func(_ int) string { return "cookie-123" }

		return svc
	}

	t.Run("Success creates session", func(t *testing.T) {
		sessionSvc := mocks.NewSessionServiceMock(t)
		sessionSvc.EXPECT().
			Create(mock.Anything, domain.Session{
				ID:        "user-id",
				CookieID:  "cookie-123",
				UserID:    "user-1",
				CheckedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			}).
			Return(domain.Session{}, nil)

		got, err := newSvc(t, sessionSvc).LoginSSO(
			context.Background(), testSSOIssuer, testSSOSub, "bob@example.com", "bob1", "0.0.0.0",
		)
		assert.NoError(t, err)
		assert.Equal(t, domain.CookieID("cookie-123"), got)
	})

	t.Run("Session error", func(t *testing.T) {
		sessionSvc := mocks.NewSessionServiceMock(t)
		sessionSvc.EXPECT().
			Create(mock.Anything, mock.Anything).
			Return(domain.Session{}, errors.New("session error"))

		_, err := newSvc(t, sessionSvc).LoginSSO(
			context.Background(), testSSOIssuer, testSSOSub, "bob@example.com", "bob1", "0.0.0.0",
		)
		assert.Error(t, err)
	})
}

func TestSanitizeSSOUsername(t *testing.T) {
	tests := []struct {
		name      string
		preferred string
		email     string
		want      string
	}{
		{name: "Valid kept", preferred: "bob1", email: "a@b.c", want: "bob1"},
		{name: "Valid lowercased", preferred: "Dimnik99", email: "a@b.c", want: "dimnik99"},
		{name: "Invalid preferred falls back to email prefix", preferred: "Dim.Nik", email: "Bob.Smith+1@example.com", want: "bobsmith1"},
		{name: "Empty preferred uses email prefix", preferred: "", email: "Alice@Example.COM", want: "alice"},
		{name: "Preferred with space uses email prefix", preferred: "red fox", email: "a@b.c", want: "a"},
		{name: "Email without at-sign sanitized", preferred: "!!!", email: "plainaddress", want: "plainaddress"},
		{name: "Empty result falls back to user", preferred: "!!!", email: "!!!@example.com", want: "user"},
		{name: "Uppercase email prefix lowercased", preferred: "", email: "JOHN.DOE@Example.com", want: "johndoe"},
		{name: "Digits preserved", preferred: "x y", email: "user123@x.io", want: "user123"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, sanitizeSSOUsername(tt.preferred, tt.email))
		})
	}
}

func FuzzSanitizeSSOUsername(f *testing.F) {
	f.Add("bob1", "bob@example.com")
	f.Add("Dim.Nik", "Bob.Smith+1@example.com")
	f.Add("", "!!!@example.com")
	f.Add("!!!", "")

	f.Fuzz(func(t *testing.T, preferred string, email string) {
		got := sanitizeSSOUsername(preferred, email)

		if got == "" {
			t.Fatalf("username is empty for preferred %q email %q", preferred, email)
		}
		if !isSSOUsernameValid(got) {
			t.Fatalf("username %q does not match ^[a-z0-9]+$ for preferred %q email %q", got, preferred, email)
		}
	})
}
