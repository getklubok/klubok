package user

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"klubok/internal/application/service/mailer"
	"klubok/internal/application/service/user/mocks"
	"klubok/internal/db/sqlc"
	"klubok/internal/domain"
)

var errDB = errors.New("db error")

func TestService_FindByID(t *testing.T) {
	tests := []struct {
		name    string
		id      domain.UserID
		svc     func(t *testing.T) *Service
		want    *domain.User
		wantErr error
	}{
		{
			name: "Success",
			id:   domain.UserID("user-1"),
			svc: func(t *testing.T) *Service {
				t.Helper()
				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					UserFindByID(mock.Anything, "user-1").
					Return(sqlc.User{
						ID:        "user-1",
						Email:     "test@example.com",
						Username:  "testuser",
						CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						UpdatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					}, nil)
				return &Service{
					querier: q,
					log:     zap.NewNop(),
				}
			},
			want: &domain.User{
				ID:        domain.UserID("user-1"),
				Email:     "test@example.com",
				Username:  "testuser",
				CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "Not found",
			id:   domain.UserID("user-1"),
			svc: func(t *testing.T) *Service {
				t.Helper()
				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					UserFindByID(mock.Anything, "user-1").
					Return(sqlc.User{}, sql.ErrNoRows)
				return &Service{
					querier: q,
					log:     zap.NewNop(),
				}
			},
			wantErr: domain.ErrUserNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.svc(t)
			got, err := svc.FindByID(context.Background(), tt.id)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestService_FindByEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		svc     func(t *testing.T) *Service
		want    *domain.User
		wantErr error
	}{
		{
			name:  "Success",
			email: "test@example.com",
			svc: func(t *testing.T) *Service {
				t.Helper()
				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					UserFindByEmail(mock.Anything, "test@example.com").
					Return(sqlc.User{
						ID:        "user-1",
						Email:     "test@example.com",
						Username:  "testuser",
						CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						UpdatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					}, nil)
				return &Service{
					querier: q,
					log:     zap.NewNop(),
				}
			},
			want: &domain.User{
				ID:        domain.UserID("user-1"),
				Email:     "test@example.com",
				Username:  "testuser",
				CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name:  "Not found",
			email: "test@example.com",
			svc: func(t *testing.T) *Service {
				t.Helper()
				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					UserFindByEmail(mock.Anything, "test@example.com").
					Return(sqlc.User{}, sql.ErrNoRows)
				return &Service{
					querier: q,
					log:     zap.NewNop(),
				}
			},
			wantErr: domain.ErrUserNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.svc(t)
			got, err := svc.FindByEmail(context.Background(), tt.email)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestService_FindByName(t *testing.T) {
	tests := []struct {
		name     string
		username string
		svc      func(t *testing.T) *Service
		want     *domain.User
		wantErr  error
	}{
		{
			name:     "Success",
			username: "testuser",
			svc: func(t *testing.T) *Service {
				t.Helper()
				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					UserFindByName(mock.Anything, "testuser").
					Return(sqlc.User{
						ID:        "user-1",
						Email:     "test@example.com",
						Username:  "testuser",
						CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						UpdatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					}, nil)
				return &Service{
					querier: q,
					log:     zap.NewNop(),
				}
			},
			want: &domain.User{
				ID:        domain.UserID("user-1"),
				Email:     "test@example.com",
				Username:  "testuser",
				CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name:     "Not found",
			username: "unknown",
			svc: func(t *testing.T) *Service {
				t.Helper()
				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					UserFindByName(mock.Anything, "unknown").
					Return(sqlc.User{}, sql.ErrNoRows)
				return &Service{
					querier: q,
					log:     zap.NewNop(),
				}
			},
			wantErr: domain.ErrUserNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.svc(t)
			got, err := svc.FindByName(context.Background(), tt.username)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestService_Update(t *testing.T) {
	tests := []struct {
		name    string
		user    domain.User
		svc     func(t *testing.T) *Service
		want    domain.User
		wantErr error
	}{
		{
			name: "Success",
			user: domain.User{
				ID:       domain.UserID("user-1"),
				Email:    "new@example.com",
				Username: "newuser",
			},
			svc: func(t *testing.T) *Service {
				t.Helper()
				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					UserUpdate(mock.Anything, sqlc.UserUpdateParams{
						ID:       "user-1",
						Email:    "new@example.com",
						Username: "newuser",
					}).
					Return(sqlc.User{
						ID:        "user-1",
						Email:     "new@example.com",
						Username:  "newuser",
						CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						UpdatedAt: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
					}, nil)
				return &Service{
					querier: q,
					log:     zap.NewNop(),
				}
			},
			want: domain.User{
				ID:        domain.UserID("user-1"),
				Email:     "new@example.com",
				Username:  "newuser",
				CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.svc(t)
			got, err := svc.Update(context.Background(), tt.user)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestService_Register(t *testing.T) {
	var emailCh chan mailer.EmailTask

	tests := []struct {
		name    string
		user    domain.User
		ip      domain.IP
		masker  domain.IPMasker
		svc     func(t *testing.T) *Service
		want    domain.User
		wantErr error
	}{
		{
			name: "Success",
			user: domain.User{
				Email:    "test@example.com",
				Username: "testuser",
			},
			ip: "127.0.0.1",
			svc: func(t *testing.T) *Service {
				t.Helper()
				emailCh = make(chan mailer.EmailTask, 1)
				q := mocks.NewQuerierMock(t)
				dbUC := mocks.NewDBUseCaseMock(t)

				q.EXPECT().
					UserFindByEmail(mock.Anything, "test@example.com").
					Return(sqlc.User{}, sql.ErrNoRows)

				dbUC.EXPECT().
					UserCreate(mock.Anything, mock.Anything, mock.Anything).
					Return(sqlc.User{
						ID:        "generated-id",
						Email:     "test@example.com",
						Username:  "testuser",
						CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						UpdatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					}, nil)

				return &Service{
					querier:        q,
					dbUseCase:      dbUC,
					emailCh:        emailCh,
					log:            zap.NewNop(),
					nowFunc:        func() time.Time { return time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC) },
					idFunc:         func() string { return "generated-id" },
					randLetterFunc: func(l int) string { return "generated-code" },
				}
			},
			want: domain.User{
				ID:        "generated-id",
				Email:     "test@example.com",
				Username:  "testuser",
				CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "User already exists",
			user: domain.User{
				Email:    "test@example.com",
				Username: "testuser",
			},
			ip: "127.0.0.1",
			svc: func(t *testing.T) *Service {
				t.Helper()
				q := mocks.NewQuerierMock(t)

				q.EXPECT().
					UserFindByEmail(mock.Anything, "test@example.com").
					Return(sqlc.User{
						ID:       "existing-id",
						Email:    "test@example.com",
						Username: "testuser",
					}, nil)

				return &Service{
					querier:        q,
					log:            zap.NewNop(),
					nowFunc:        func() time.Time { return time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC) },
					idFunc:         func() string { return "generated-id" },
					randLetterFunc: func(l int) string { return "generated-code" },
				}
			},
			wantErr: domain.ErrUserAlreadyExists,
		},
		{
			name: "DBUseCase error - already exists",
			user: domain.User{
				Email:    "test@example.com",
				Username: "testuser",
			},
			ip: "127.0.0.1",
			svc: func(t *testing.T) *Service {
				t.Helper()
				q := mocks.NewQuerierMock(t)
				dbUC := mocks.NewDBUseCaseMock(t)

				q.EXPECT().
					UserFindByEmail(mock.Anything, "test@example.com").
					Return(sqlc.User{}, sql.ErrNoRows)

				dbUC.EXPECT().
					UserCreate(mock.Anything, mock.Anything, mock.Anything).
					Return(sqlc.User{}, domain.ErrUserAlreadyExists)

				return &Service{
					querier:        q,
					dbUseCase:      dbUC,
					log:            zap.NewNop(),
					nowFunc:        func() time.Time { return time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC) },
					idFunc:         func() string { return "generated-id" },
					randLetterFunc: func(l int) string { return "generated-code" },
				}
			},
			wantErr: errors.New("create user: user already exists"),
		},
		{
			name: "DBUseCase error - other",
			user: domain.User{
				Email:    "test@example.com",
				Username: "testuser",
			},
			ip: "127.0.0.1",
			svc: func(t *testing.T) *Service {
				t.Helper()
				q := mocks.NewQuerierMock(t)
				dbUC := mocks.NewDBUseCaseMock(t)

				q.EXPECT().
					UserFindByEmail(mock.Anything, "test@example.com").
					Return(sqlc.User{}, sql.ErrNoRows)

				dbUC.EXPECT().
					UserCreate(mock.Anything, mock.Anything, mock.Anything).
					Return(sqlc.User{}, errDB)

				return &Service{
					querier:        q,
					dbUseCase:      dbUC,
					log:            zap.NewNop(),
					nowFunc:        func() time.Time { return time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC) },
					idFunc:         func() string { return "generated-id" },
					randLetterFunc: func(l int) string { return "generated-code" },
				}
			},
			wantErr: errors.New("create user: db error"),
		},
		{
			name: "Masking enabled - truncated IP in privacy log params",
			user: domain.User{
				Email:    "test@example.com",
				Username: "testuser",
			},
			ip:     "1.2.3.4",
			masker: domain.IPMasker{Enabled: true, V4Octets: 3, V6Octets: 7},
			svc: func(t *testing.T) *Service {
				t.Helper()
				emailCh = make(chan mailer.EmailTask, 1)
				q := mocks.NewQuerierMock(t)
				dbUC := mocks.NewDBUseCaseMock(t)

				q.EXPECT().
					UserFindByEmail(mock.Anything, "test@example.com").
					Return(sqlc.User{}, sql.ErrNoRows)

				dbUC.EXPECT().
					UserCreate(mock.Anything, mock.Anything, sqlc.PrivacyLogCreateParams{
						UserID:         "generated-id",
						Ip:             "1.2.3.x",
						PrivacyVersion: privacyVersion,
					}).
					Return(sqlc.User{
						ID:        "generated-id",
						Email:     "test@example.com",
						Username:  "testuser",
						CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						UpdatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					}, nil)

				return &Service{
					querier:        q,
					dbUseCase:      dbUC,
					emailCh:        emailCh,
					log:            zap.NewNop(),
					nowFunc:        func() time.Time { return time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC) },
					idFunc:         func() string { return "generated-id" },
					randLetterFunc: func(l int) string { return "generated-code" },
				}
			},
			want: domain.User{
				ID:        "generated-id",
				Email:     "test@example.com",
				Username:  "testuser",
				CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "Masking disabled - full IP in privacy log params",
			user: domain.User{
				Email:    "test@example.com",
				Username: "testuser",
			},
			ip:     "1.2.3.4",
			masker: domain.IPMasker{Enabled: false, V4Octets: 3, V6Octets: 7},
			svc: func(t *testing.T) *Service {
				t.Helper()
				emailCh = make(chan mailer.EmailTask, 1)
				q := mocks.NewQuerierMock(t)
				dbUC := mocks.NewDBUseCaseMock(t)

				q.EXPECT().
					UserFindByEmail(mock.Anything, "test@example.com").
					Return(sqlc.User{}, sql.ErrNoRows)

				dbUC.EXPECT().
					UserCreate(mock.Anything, mock.Anything, sqlc.PrivacyLogCreateParams{
						UserID:         "generated-id",
						Ip:             "1.2.3.4",
						PrivacyVersion: privacyVersion,
					}).
					Return(sqlc.User{
						ID:        "generated-id",
						Email:     "test@example.com",
						Username:  "testuser",
						CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						UpdatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					}, nil)

				return &Service{
					querier:        q,
					dbUseCase:      dbUC,
					emailCh:        emailCh,
					log:            zap.NewNop(),
					nowFunc:        func() time.Time { return time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC) },
					idFunc:         func() string { return "generated-id" },
					randLetterFunc: func(l int) string { return "generated-code" },
				}
			},
			want: domain.User{
				ID:        "generated-id",
				Email:     "test@example.com",
				Username:  "testuser",
				CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "Masking enabled - masked IPv6 with port in privacy log params",
			user: domain.User{
				Email:    "test@example.com",
				Username: "testuser",
			},
			ip:     "[2001:db8:85a3:0:0:8a2e:370:7334]:443",
			masker: domain.IPMasker{Enabled: true, V4Octets: 3, V6Octets: 7},
			svc: func(t *testing.T) *Service {
				t.Helper()
				emailCh = make(chan mailer.EmailTask, 1)
				q := mocks.NewQuerierMock(t)
				dbUC := mocks.NewDBUseCaseMock(t)

				q.EXPECT().
					UserFindByEmail(mock.Anything, "test@example.com").
					Return(sqlc.User{}, sql.ErrNoRows)

				dbUC.EXPECT().
					UserCreate(mock.Anything, mock.Anything, sqlc.PrivacyLogCreateParams{
						UserID:         "generated-id",
						Ip:             "2001:db8:85a3:0:0:8a2e:370:x",
						PrivacyVersion: privacyVersion,
					}).
					Return(sqlc.User{
						ID:        "generated-id",
						Email:     "test@example.com",
						Username:  "testuser",
						CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						UpdatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					}, nil)

				return &Service{
					querier:        q,
					dbUseCase:      dbUC,
					emailCh:        emailCh,
					log:            zap.NewNop(),
					nowFunc:        func() time.Time { return time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC) },
					idFunc:         func() string { return "generated-id" },
					randLetterFunc: func(l int) string { return "generated-code" },
				}
			},
			want: domain.User{
				ID:        "generated-id",
				Email:     "test@example.com",
				Username:  "testuser",
				CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			emailCh = nil
			svc := tt.svc(t)
			svc.ipMasker = tt.masker // zero value: disabled
			got, err := svc.Register(context.Background(), tt.user, tt.ip)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)

			if emailCh != nil {
				select {
				case task := <-emailCh:
					assert.Equal(t, []string{"test@example.com"}, task.Recipients)
					assert.Equal(t, "Успешная регистрация", task.Subject)
					assert.Equal(t, mailer.TemplateRegisterSuccess, task.Template)
					assert.Equal(t, "auto", task.Data["sysAppURL"])
					assert.Equal(t, domain.UserID("generated-id"), task.Data["userID"])
					assert.Equal(t, "generated-code", task.Data["activateCode"])
				case <-time.After(time.Second):
					t.Fatal("expected email task on channel")
				}
			}
		})
	}
}

func TestService_Login(t *testing.T) {
	var emailCh chan mailer.EmailTask

	tests := []struct {
		name    string
		email   string
		svc     func(t *testing.T) *Service
		wantErr error
	}{
		{
			name:  "Success",
			email: "test@example.com",
			svc: func(t *testing.T) *Service {
				t.Helper()
				emailCh = make(chan mailer.EmailTask, 1)
				q := mocks.NewQuerierMock(t)

				q.EXPECT().
					UserFindByEmail(mock.Anything, "test@example.com").
					Return(sqlc.User{
						ID:       "user-1",
						Email:    "test@example.com",
						Username: "testuser",
					}, nil)

				q.EXPECT().
					AuthCodeUpsert(mock.Anything, mock.Anything).
					Return(nil)

				return &Service{
					querier:        q,
					emailCh:        emailCh,
					log:            zap.NewNop(),
					nowFunc:        func() time.Time { return time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC) },
					idFunc:         func() string { return "generated-id" },
					randDigitFunc:  func(l int) string { return "123456" },
					randLetterFunc: func(l int) string { return "generated-code" },
				}
			},
		},
		{
			name:  "User not found",
			email: "test@example.com",
			svc: func(t *testing.T) *Service {
				t.Helper()
				q := mocks.NewQuerierMock(t)

				q.EXPECT().
					UserFindByEmail(mock.Anything, "test@example.com").
					Return(sqlc.User{}, sql.ErrNoRows)

				return &Service{
					querier:        q,
					log:            zap.NewNop(),
					nowFunc:        func() time.Time { return time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC) },
					idFunc:         func() string { return "generated-id" },
					randDigitFunc:  func(l int) string { return "123456" },
					randLetterFunc: func(l int) string { return "generated-code" },
				}
			},
			wantErr: errors.New("find by email: user not found"),
		},
		{
			name:  "User not activated",
			email: "test@example.com",
			svc: func(t *testing.T) *Service {
				t.Helper()
				q := mocks.NewQuerierMock(t)

				activateCode := "some-code"
				q.EXPECT().
					UserFindByEmail(mock.Anything, "test@example.com").
					Return(sqlc.User{
						ID:           "user-1",
						Email:        "test@example.com",
						Username:     "testuser",
						ActivateCode: &activateCode,
					}, nil)

				return &Service{
					querier:        q,
					log:            zap.NewNop(),
					nowFunc:        func() time.Time { return time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC) },
					idFunc:         func() string { return "generated-id" },
					randDigitFunc:  func(l int) string { return "123456" },
					randLetterFunc: func(l int) string { return "generated-code" },
				}
			},
			wantErr: domain.NewFailedPreconditionError("user not activated"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			emailCh = nil
			svc := tt.svc(t)
			err := svc.Login(context.Background(), tt.email)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}

			if emailCh != nil {
				select {
				case task := <-emailCh:
					assert.Equal(t, []string{"test@example.com"}, task.Recipients)
					assert.Equal(t, "Код авторизации", task.Subject)
					assert.Equal(t, mailer.TemplateAuthCode, task.Template)
					assert.Equal(t, "123456", task.Data["code"])
				case <-time.After(time.Second):
					t.Fatal("expected email task on channel")
				}
			}
		})
	}
}

func TestService_LoginVerify(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		code    string
		svc     func(t *testing.T) *Service
		want    domain.CookieID
		wantErr error
	}{
		{
			name:  "Success",
			email: "test@example.com",
			code:  "123456",
			svc: func(t *testing.T) *Service {
				t.Helper()
				q := mocks.NewQuerierMock(t)
				sess := mocks.NewSessionServiceMock(t)

				q.EXPECT().
					AuthCodeCancelOld(mock.Anything, mock.Anything).
					Return(nil)

				q.EXPECT().
					AuthCodeFind(mock.Anything, "test@example.com", "123456").
					Return(sqlc.AuthCode{
						ID:        "auth-1",
						UserID:    "user-1",
						UserEmail: "test@example.com",
						Code:      "123456",
					}, nil)

				sess.EXPECT().
					Create(mock.Anything, mock.Anything).
					Return(domain.Session{}, nil)

				q.EXPECT().
					AuthCodeDelete(mock.Anything, "auth-1").
					Return(nil)

				return &Service{
					querier:        q,
					sessionService: sess,
					log:            zap.NewNop(),
					nowFunc:        func() time.Time { return time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC) },
					idFunc:         func() string { return "generated-id" },
					randLetterFunc: func(l int) string { return "generated-cookie-id" },
				}
			},
			want: domain.CookieID("generated-cookie-id"),
		},
		{
			name:  "AuthCodeCancelOld error",
			email: "test@example.com",
			code:  "123456",
			svc: func(t *testing.T) *Service {
				t.Helper()
				q := mocks.NewQuerierMock(t)

				q.EXPECT().
					AuthCodeCancelOld(mock.Anything, mock.Anything).
					Return(errDB)

				return &Service{
					querier:        q,
					log:            zap.NewNop(),
					nowFunc:        func() time.Time { return time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC) },
					idFunc:         func() string { return "generated-id" },
					randLetterFunc: func(l int) string { return "generated-cookie-id" },
				}
			},
			wantErr: errors.New("cancel old: db error"),
		},
		{
			name:  "AuthCodeFind error",
			email: "test@example.com",
			code:  "123456",
			svc: func(t *testing.T) *Service {
				t.Helper()
				q := mocks.NewQuerierMock(t)

				q.EXPECT().
					AuthCodeCancelOld(mock.Anything, mock.Anything).
					Return(nil)

				q.EXPECT().
					AuthCodeFind(mock.Anything, "test@example.com", "123456").
					Return(sqlc.AuthCode{}, errDB)

				return &Service{
					querier:        q,
					log:            zap.NewNop(),
					nowFunc:        func() time.Time { return time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC) },
					idFunc:         func() string { return "generated-id" },
					randLetterFunc: func(l int) string { return "generated-cookie-id" },
				}
			},
			wantErr: errors.New("find auth code: db error"),
		},
		{
			name:  "Session create error",
			email: "test@example.com",
			code:  "123456",
			svc: func(t *testing.T) *Service {
				t.Helper()
				q := mocks.NewQuerierMock(t)
				sess := mocks.NewSessionServiceMock(t)

				q.EXPECT().
					AuthCodeCancelOld(mock.Anything, mock.Anything).
					Return(nil)

				q.EXPECT().
					AuthCodeFind(mock.Anything, "test@example.com", "123456").
					Return(sqlc.AuthCode{
						ID:        "auth-1",
						UserID:    "user-1",
						UserEmail: "test@example.com",
						Code:      "123456",
					}, nil)

				sess.EXPECT().
					Create(mock.Anything, mock.Anything).
					Return(domain.Session{}, errDB)

				return &Service{
					querier:        q,
					sessionService: sess,
					log:            zap.NewNop(),
					nowFunc:        func() time.Time { return time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC) },
					idFunc:         func() string { return "generated-id" },
					randLetterFunc: func(l int) string { return "generated-cookie-id" },
				}
			},
			wantErr: errors.New("create session: db error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.svc(t)
			got, err := svc.LoginVerify(context.Background(), tt.email, tt.code)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestService_Activate(t *testing.T) {
	tests := []struct {
		name    string
		userID  domain.UserID
		code    string
		svc     func(t *testing.T) *Service
		want    domain.CookieID
		wantErr error
	}{
		{
			name:   "Success",
			userID: domain.UserID("user-1"),
			code:   "activate-code",
			svc: func(t *testing.T) *Service {
				t.Helper()
				q := mocks.NewQuerierMock(t)
				sess := mocks.NewSessionServiceMock(t)

				q.EXPECT().
					UserDeleteNotActivated(mock.Anything, mock.Anything).
					Return(nil)

				activateCode := "activate-code"
				q.EXPECT().
					UserFindByID(mock.Anything, "user-1").
					Return(sqlc.User{
						ID:           "user-1",
						Email:        "test@example.com",
						Username:     "testuser",
						ActivateCode: &activateCode,
					}, nil)

				sess.EXPECT().
					Create(mock.Anything, mock.Anything).
					Return(domain.Session{}, nil)

				q.EXPECT().
					UserCancelActivateCode(mock.Anything, "user-1").
					Return(nil)

				return &Service{
					querier:        q,
					sessionService: sess,
					log:            zap.NewNop(),
					nowFunc:        func() time.Time { return time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC) },
					idFunc:         func() string { return "generated-id" },
					randLetterFunc: func(l int) string { return "generated-cookie-id" },
				}
			},
			want: domain.CookieID("generated-cookie-id"),
		},
		{
			name:   "UserDeleteNotActivated error",
			userID: domain.UserID("user-1"),
			code:   "activate-code",
			svc: func(t *testing.T) *Service {
				t.Helper()
				q := mocks.NewQuerierMock(t)

				q.EXPECT().
					UserDeleteNotActivated(mock.Anything, mock.Anything).
					Return(errDB)

				return &Service{
					querier:        q,
					log:            zap.NewNop(),
					nowFunc:        func() time.Time { return time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC) },
					idFunc:         func() string { return "generated-id" },
					randLetterFunc: func(l int) string { return "generated-cookie-id" },
				}
			},
			wantErr: errors.New("cancel old: db error"),
		},
		{
			name:   "User not found",
			userID: domain.UserID("user-1"),
			code:   "activate-code",
			svc: func(t *testing.T) *Service {
				t.Helper()
				q := mocks.NewQuerierMock(t)

				q.EXPECT().
					UserDeleteNotActivated(mock.Anything, mock.Anything).
					Return(nil)

				q.EXPECT().
					UserFindByID(mock.Anything, "user-1").
					Return(sqlc.User{}, sql.ErrNoRows)

				return &Service{
					querier:        q,
					log:            zap.NewNop(),
					nowFunc:        func() time.Time { return time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC) },
					idFunc:         func() string { return "generated-id" },
					randLetterFunc: func(l int) string { return "generated-cookie-id" },
				}
			},
			wantErr: errors.New("find by id: user not found"),
		},
		{
			name:   "Activate code is nil",
			userID: domain.UserID("user-1"),
			code:   "activate-code",
			svc: func(t *testing.T) *Service {
				t.Helper()
				q := mocks.NewQuerierMock(t)

				q.EXPECT().
					UserDeleteNotActivated(mock.Anything, mock.Anything).
					Return(nil)

				q.EXPECT().
					UserFindByID(mock.Anything, "user-1").
					Return(sqlc.User{
						ID:       "user-1",
						Email:    "test@example.com",
						Username: "testuser",
					}, nil)

				return &Service{
					querier:        q,
					log:            zap.NewNop(),
					nowFunc:        func() time.Time { return time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC) },
					idFunc:         func() string { return "generated-id" },
					randLetterFunc: func(l int) string { return "generated-cookie-id" },
				}
			},
			wantErr: domain.ErrUserNotFound,
		},
		{
			name:   "Invalid activate code",
			userID: domain.UserID("user-1"),
			code:   "wrong-code",
			svc: func(t *testing.T) *Service {
				t.Helper()
				q := mocks.NewQuerierMock(t)

				q.EXPECT().
					UserDeleteNotActivated(mock.Anything, mock.Anything).
					Return(nil)

				activateCode := "correct-code"
				q.EXPECT().
					UserFindByID(mock.Anything, "user-1").
					Return(sqlc.User{
						ID:           "user-1",
						Email:        "test@example.com",
						Username:     "testuser",
						ActivateCode: &activateCode,
					}, nil)

				return &Service{
					querier:        q,
					log:            zap.NewNop(),
					nowFunc:        func() time.Time { return time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC) },
					idFunc:         func() string { return "generated-id" },
					randLetterFunc: func(l int) string { return "generated-cookie-id" },
				}
			},
			wantErr: domain.ErrUserNotFound,
		},
		{
			name:   "Session create error",
			userID: domain.UserID("user-1"),
			code:   "activate-code",
			svc: func(t *testing.T) *Service {
				t.Helper()
				q := mocks.NewQuerierMock(t)
				sess := mocks.NewSessionServiceMock(t)

				q.EXPECT().
					UserDeleteNotActivated(mock.Anything, mock.Anything).
					Return(nil)

				activateCode := "activate-code"
				q.EXPECT().
					UserFindByID(mock.Anything, "user-1").
					Return(sqlc.User{
						ID:           "user-1",
						Email:        "test@example.com",
						Username:     "testuser",
						ActivateCode: &activateCode,
					}, nil)

				sess.EXPECT().
					Create(mock.Anything, mock.Anything).
					Return(domain.Session{}, errDB)

				return &Service{
					querier:        q,
					sessionService: sess,
					log:            zap.NewNop(),
					nowFunc:        func() time.Time { return time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC) },
					idFunc:         func() string { return "generated-id" },
					randLetterFunc: func(l int) string { return "generated-cookie-id" },
				}
			},
			wantErr: errors.New("create session: db error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.svc(t)
			got, err := svc.Activate(context.Background(), tt.userID, tt.code)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestService_GenerateFakeEmail(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want string
	}{
		{name: "Plain id", id: "abc123", want: "abc123@localhost.tld"},
		{name: "Id with spaces kept", id: "a b c", want: "a b c@localhost.tld"},
		{name: "Empty id", id: "", want: "@localhost.tld"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{fakeIDFunc: func() string { return tt.id }}
			assert.Equal(t, tt.want, s.generateFakeEmail())
		})
	}
}

func TestService_GenerateFakeNickname(t *testing.T) {
	tests := []struct {
		name   string
		word   string
		digits string
		want   string
	}{
		{name: "Lowercase word", word: "Apple", digits: "1234", want: "apple1234"},
		{name: "Spaces removed", word: "Red Fox", digits: "0000", want: "redfox0000"},
		{name: "Multiple spaces", word: "  Green   Grape ", digits: "7777", want: "greengrape7777"},
		{name: "Already lower", word: "banana", digits: "9999", want: "banana9999"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				fakeWordFunc:  func() string { return tt.word },
				fakeDigitFunc: func(_ uint) string { return tt.digits },
			}
			assert.Equal(t, tt.want, s.generateFakeNickname())
		})
	}
}

// newCreateFakeService строит Service для тестов CreateFake: фиксированные
// генераторы слов/цифр позволяют детерминированно предсказывать email и
// nickname каждой попытки.
func newCreateFakeService(t *testing.T, db DBUseCase, sess SessionService) *Service {
	t.Helper()
	return &Service{
		dbUseCase:      db,
		sessionService: sess,
		ipMasker:       domain.IPMasker{},
		log:            zap.NewNop(),
		idFunc:         func() string { return "user-id" },
		randLetterFunc: func(l int) string { return "cookie-id" },
		fakeIDFunc:     func() string { return "fake-id" },
		fakeWordFunc:   func() string { return "Apple" },
		fakeDigitFunc:  func(l uint) string { return "1234" },
		nowFunc:        func() time.Time { return time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC) },
	}
}

// newCreateFakeRetryService строит Service для тестов ретрая CreateFake:
// генераторы при каждом вызове возвращают новые значения, чтобы первая и
// повторная попытки различались по email/nickname.
func newCreateFakeRetryService(t *testing.T, db DBUseCase, sess SessionService) *Service {
	t.Helper()

	makeGen := func(first, second string) func() string {
		n := 0
		return func() string {
			n++
			if n == 1 {
				return first
			}
			return second
		}
	}

	digitN := 0

	return &Service{
		dbUseCase:      db,
		sessionService: sess,
		ipMasker:       domain.IPMasker{},
		log:            zap.NewNop(),
		idFunc:         func() string { return "user-id" },
		randLetterFunc: func(l int) string { return "cookie-id" },
		fakeIDFunc:     makeGen("fake-id", "fake-id-2"),
		fakeWordFunc:   makeGen("Apple", "Banana"),
		fakeDigitFunc: func(_ uint) string {
			digitN++
			if digitN == 1 {
				return "1234"
			}
			return "2345"
		},
		nowFunc: func() time.Time { return time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC) },
	}
}

func TestService_CreateFake(t *testing.T) {
	tests := []struct {
		name    string
		svc     func(t *testing.T) *Service
		want    domain.User
		wantErr error
	}{
		{
			name: "Success",
			svc: func(t *testing.T) *Service {
				t.Helper()
				db := mocks.NewDBUseCaseMock(t)
				sess := mocks.NewSessionServiceMock(t)

				db.EXPECT().
					UserCreate(mock.Anything, sqlc.UserCreateParams{
						ID:       "user-id",
						Email:    "fake-id@localhost.tld",
						Username: "apple1234",
						IsFake:   true,
					}, sqlc.PrivacyLogCreateParams{
						UserID:         "user-id",
						Ip:             "0.0.0.0",
						PrivacyVersion: privacyVersion,
					}).
					Return(sqlc.User{
						ID:        "user-id",
						Email:     "fake-id@localhost.tld",
						Username:  "apple1234",
						CreatedAt: time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC),
						UpdatedAt: time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC),
						IsFake:    true,
					}, nil)

				sess.EXPECT().
					Create(mock.Anything, domain.Session{
						ID:        domain.SessionID("user-id"),
						CookieID:  domain.CookieID("cookie-id"),
						UserID:    domain.UserID("user-id"),
						CheckedAt: time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC),
					}).
					Return(domain.Session{}, nil)

				return newCreateFakeService(t, db, sess)
			},
			want: domain.User{
				ID:        domain.UserID("user-id"),
				Email:     "fake-id@localhost.tld",
				Username:  "apple1234",
				CreatedAt: time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC),
				IsFake:    true,
			},
		},
		{
			name: "Retry once on ErrUserAlreadyExists",
			svc: func(t *testing.T) *Service {
				t.Helper()
				db := mocks.NewDBUseCaseMock(t)
				sess := mocks.NewSessionServiceMock(t)

				db.EXPECT().
					UserCreate(mock.Anything, sqlc.UserCreateParams{
						ID:       "user-id",
						Email:    "fake-id@localhost.tld",
						Username: "apple1234",
						IsFake:   true,
					}, mock.Anything).
					Return(sqlc.User{}, domain.ErrUserAlreadyExists)

				db.EXPECT().
					UserCreate(mock.Anything, sqlc.UserCreateParams{
						ID:       "user-id",
						Email:    "fake-id-2@localhost.tld",
						Username: "banana2345",
						IsFake:   true,
					}, mock.Anything).
					Return(sqlc.User{
						ID:        "user-id",
						Email:     "fake-id-2@localhost.tld",
						Username:  "banana2345",
						CreatedAt: time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC),
						UpdatedAt: time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC),
						IsFake:    true,
					}, nil)

				sess.EXPECT().
					Create(mock.Anything, mock.Anything).
					Return(domain.Session{}, nil)

				return newCreateFakeRetryService(t, db, sess)
			},
			want: domain.User{
				ID:        domain.UserID("user-id"),
				Email:     "fake-id-2@localhost.tld",
				Username:  "banana2345",
				CreatedAt: time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC),
				IsFake:    true,
			},
		},
		{
			name: "Second attempt fails, error returned",
			svc: func(t *testing.T) *Service {
				t.Helper()
				db := mocks.NewDBUseCaseMock(t)
				sess := mocks.NewSessionServiceMock(t)

				db.EXPECT().
					UserCreate(mock.Anything, mock.Anything, mock.Anything).
					Return(sqlc.User{}, domain.ErrUserAlreadyExists)
				db.EXPECT().
					UserCreate(mock.Anything, mock.Anything, mock.Anything).
					Return(sqlc.User{}, domain.ErrUserAlreadyExists)

				return newCreateFakeService(t, db, sess)
			},
			wantErr: errors.New("create fake user: create user: user already exists"),
		},
		{
			name: "Non-unique create error is not retried",
			svc: func(t *testing.T) *Service {
				t.Helper()
				db := mocks.NewDBUseCaseMock(t)
				sess := mocks.NewSessionServiceMock(t)

				db.EXPECT().
					UserCreate(mock.Anything, mock.Anything, mock.Anything).
					Return(sqlc.User{}, errDB)

				return newCreateFakeService(t, db, sess)
			},
			wantErr: errors.New("create fake user: create user: db error"),
		},
		{
			name: "Session create error",
			svc: func(t *testing.T) *Service {
				t.Helper()
				db := mocks.NewDBUseCaseMock(t)
				sess := mocks.NewSessionServiceMock(t)

				db.EXPECT().
					UserCreate(mock.Anything, mock.Anything, mock.Anything).
					Return(sqlc.User{
						ID:        "user-id",
						Email:     "fake-id@localhost.tld",
						Username:  "apple1234",
						CreatedAt: time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC),
						UpdatedAt: time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC),
						IsFake:    true,
					}, nil)

				sess.EXPECT().
					Create(mock.Anything, mock.Anything).
					Return(domain.Session{}, errDB)

				return newCreateFakeService(t, db, sess)
			},
			wantErr: errors.New("create fake user: create session: db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.svc(t)
			got, _, err := svc.CreateFake(context.Background(), "0.0.0.0")
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func FuzzGenerateFakeNickname(f *testing.F) {
	f.Add("Apple")
	f.Add("Red Fox")
	f.Add("  Green   Grape ")
	f.Add("")

	f.Fuzz(func(t *testing.T, word string) {
		s := &Service{
			fakeWordFunc:  func() string { return word },
			fakeDigitFunc: func(_ uint) string { return "1234" },
		}
		got := s.generateFakeNickname()

		if strings.Contains(got, " ") {
			t.Fatalf("nickname contains spaces: %q", got)
		}
		if got == "" {
			t.Fatalf("nickname is empty")
		}
		if !strings.HasSuffix(got, "1234") {
			t.Fatalf("nickname missing digit suffix: %q", got)
		}
		if got != strings.ToLower(got) {
			t.Fatalf("nickname not lowercased: %q", got)
		}
	})
}

func FuzzGenerateFakeEmail(f *testing.F) {
	f.Add("abc123")
	f.Add("user-id-42")
	f.Add("")

	f.Fuzz(func(t *testing.T, id string) {
		s := &Service{fakeIDFunc: func() string { return id }}
		got := s.generateFakeEmail()

		if !strings.HasSuffix(got, "@localhost.tld") {
			t.Fatalf("email missing @localhost.tld suffix: %q", got)
		}
	})
}
