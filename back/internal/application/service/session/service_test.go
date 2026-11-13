package session

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"klubok/internal/application/service/session/mocks"
	"klubok/internal/db/sqlc"
	"klubok/internal/domain"
)

func TestService_Create(t *testing.T) {
	tests := []struct {
		name    string
		input   domain.Session
		handler func(t *testing.T) *Service
		want    domain.Session
		wantErr error
	}{
		{
			name: "Success",
			input: domain.Session{
				ID:        domain.SessionID("01JHHXTMRFMJ4DQJVTVCHZSZX3"),
				CookieID:  domain.CookieID("cid-123"),
				UserID:    domain.UserID("uid-456"),
				CheckedAt: time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC),
			},
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					SessionCreate(mock.Anything, sqlc.SessionCreateParams{
						ID:        "01JHHXTMRFMJ4DQJVTVCHZSZX3",
						Cid:       "cid-123",
						Uid:       "uid-456",
						CheckedAt: time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC),
					}).
					Return(sqlc.Session{
						ID:        "01JHHXTMRFMJ4DQJVTVCHZSZX3",
						Cid:       "cid-123",
						Uid:       "uid-456",
						CheckedAt: time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC),
					}, nil)
				q.EXPECT().
					UserUpdateLastLoginAt(mock.Anything, "uid-456").
					Return(nil)

				return &Service{
					querier: q,
					log:     zap.NewNop(),
				}
			},
			want: domain.Session{
				ID:        domain.SessionID("01JHHXTMRFMJ4DQJVTVCHZSZX3"),
				CookieID:  domain.CookieID("cid-123"),
				UserID:    domain.UserID("uid-456"),
				CheckedAt: time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.handler(t)
			got, err := svc.Create(context.Background(), tt.input)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestService_UpdateCheckedAt(t *testing.T) {
	tests := []struct {
		name      string
		id        domain.SessionID
		checkedAt time.Time
		handler   func(t *testing.T) *Service
		wantErr   error
	}{
		{
			name:      "Success",
			id:        domain.SessionID("01JHHXTMRFMJ4DQJVTVCHZSZX3"),
			checkedAt: time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC),
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					SessionUpdateCheckedAt(mock.Anything, "01JHHXTMRFMJ4DQJVTVCHZSZX3", time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC)).
					Return(sqlc.Session{}, nil)

				return &Service{
					querier: q,
					log:     zap.NewNop(),
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.handler(t)
			err := svc.UpdateCheckedAt(context.Background(), tt.id, tt.checkedAt)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_DeleteByID(t *testing.T) {
	tests := []struct {
		name    string
		id      domain.SessionID
		handler func(t *testing.T) *Service
		wantErr error
	}{
		{
			name: "Success",
			id:   domain.SessionID("01JHHXTMRFMJ4DQJVTVCHZSZX3"),
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					SessionDeleteByID(mock.Anything, "01JHHXTMRFMJ4DQJVTVCHZSZX3").
					Return(nil)

				return &Service{
					querier: q,
					log:     zap.NewNop(),
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.handler(t)
			err := svc.DeleteByID(context.Background(), tt.id)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_CancelOld(t *testing.T) {
	tests := []struct {
		name      string
		olderThan time.Time
		handler   func(t *testing.T) *Service
		wantErr   error
	}{
		{
			name:      "Success",
			olderThan: time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC),
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					SessionCancelOld(mock.Anything, time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC)).
					Return(nil)

				return &Service{
					querier: q,
					log:     zap.NewNop(),
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.handler(t)
			err := svc.CancelOld(context.Background(), tt.olderThan)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_FindByID(t *testing.T) {
	tests := []struct {
		name    string
		id      domain.SessionID
		handler func(t *testing.T) *Service
		want    domain.Session
		wantErr error
	}{
		{
			name: "Success",
			id:   domain.SessionID("01JHHXTMRFMJ4DQJVTVCHZSZX3"),
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					SessionFindByID(mock.Anything, "01JHHXTMRFMJ4DQJVTVCHZSZX3").
					Return(sqlc.Session{
						ID:        "01JHHXTMRFMJ4DQJVTVCHZSZX3",
						Cid:       "cid-123",
						Uid:       "uid-456",
						CheckedAt: time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC),
					}, nil)

				return &Service{
					querier: q,
					log:     zap.NewNop(),
				}
			},
			want: domain.Session{
				ID:        domain.SessionID("01JHHXTMRFMJ4DQJVTVCHZSZX3"),
				CookieID:  domain.CookieID("cid-123"),
				UserID:    domain.UserID("uid-456"),
				CheckedAt: time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "Not found",
			id:   domain.SessionID("01JHHXTMRFMJ4DQJVTVCHZSZX3"),
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					SessionFindByID(mock.Anything, mock.Anything).
					Return(sqlc.Session{}, domain.ErrSessionNotFound)

				return &Service{
					querier: q,
					log:     zap.NewNop(),
				}
			},
			wantErr: errors.New("find session: session not found"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.handler(t)
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

func TestService_FindByCID(t *testing.T) {
	tests := []struct {
		name    string
		cid     domain.CookieID
		handler func(t *testing.T) *Service
		want    domain.Session
		wantErr error
	}{
		{
			name: "Success",
			cid:  domain.CookieID("cid-123"),
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					SessionFindByCID(mock.Anything, "cid-123").
					Return(sqlc.Session{
						ID:        "01JHHXTMRFMJ4DQJVTVCHZSZX3",
						Cid:       "cid-123",
						Uid:       "uid-456",
						CheckedAt: time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC),
					}, nil)

				return &Service{
					querier: q,
					log:     zap.NewNop(),
				}
			},
			want: domain.Session{
				ID:        domain.SessionID("01JHHXTMRFMJ4DQJVTVCHZSZX3"),
				CookieID:  domain.CookieID("cid-123"),
				UserID:    domain.UserID("uid-456"),
				CheckedAt: time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "Not found",
			cid:  domain.CookieID("cid-123"),
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					SessionFindByCID(mock.Anything, mock.Anything).
					Return(sqlc.Session{}, domain.ErrSessionNotFound)

				return &Service{
					querier: q,
					log:     zap.NewNop(),
				}
			},
			wantErr: errors.New("find session: session not found"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.handler(t)
			got, err := svc.FindByCID(context.Background(), tt.cid)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestService_FindByUID(t *testing.T) {
	tests := []struct {
		name    string
		uid     domain.UserID
		handler func(t *testing.T) *Service
		want    []domain.Session
		wantErr error
	}{
		{
			name: "Success - multiple sessions",
			uid:  domain.UserID("uid-456"),
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					SessionFindByUID(mock.Anything, "uid-456").
					Return([]sqlc.Session{
						{
							ID:        "01JHHXTMRFMJ4DQJVTVCHZSZX3",
							Cid:       "cid-123",
							Uid:       "uid-456",
							CheckedAt: time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC),
						},
						{
							ID:        "01JHHXTMRFMJ4DQJVTVCHZSZX4",
							Cid:       "cid-789",
							Uid:       "uid-456",
							CheckedAt: time.Date(2024, 12, 19, 14, 0, 0, 0, time.UTC),
						},
					}, nil)

				return &Service{
					querier: q,
					log:     zap.NewNop(),
				}
			},
			want: []domain.Session{
				{
					ID:        domain.SessionID("01JHHXTMRFMJ4DQJVTVCHZSZX3"),
					CookieID:  domain.CookieID("cid-123"),
					UserID:    domain.UserID("uid-456"),
					CheckedAt: time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC),
				},
				{
					ID:        domain.SessionID("01JHHXTMRFMJ4DQJVTVCHZSZX4"),
					CookieID:  domain.CookieID("cid-789"),
					UserID:    domain.UserID("uid-456"),
					CheckedAt: time.Date(2024, 12, 19, 14, 0, 0, 0, time.UTC),
				},
			},
		},
		{
			name: "Success - empty result",
			uid:  domain.UserID("uid-456"),
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					SessionFindByUID(mock.Anything, "uid-456").
					Return([]sqlc.Session{}, nil)

				return &Service{
					querier: q,
					log:     zap.NewNop(),
				}
			},
			want: []domain.Session{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.handler(t)
			got, err := svc.FindByUID(context.Background(), tt.uid)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
