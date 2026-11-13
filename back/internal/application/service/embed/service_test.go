package embed

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"klubok/internal/application/service/embed/mocks"
	"klubok/internal/db/sqlc"
	"klubok/internal/domain"
)

func TestService_Create(t *testing.T) {
	now := time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		diagramID domain.DiagramID
		userID    domain.UserID
		handler   func(t *testing.T) *Service
		want      domain.DiagramEmbedInfo
		wantErr   error
	}{
		{
			name:      "Success",
			diagramID: domain.DiagramID("diagram-1"),
			userID:    domain.UserID("user-1"),
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					EmbedCreate(mock.Anything, sqlc.EmbedCreateParams{
						DiagramID: "diagram-1",
						ID:        "embed-1",
						UserID:    "user-1",
						Token:     "test-token-64-chars-long-abcdefghijklmnopqrstuvwxyz012345678",
					}).
					Return(sqlc.DiagramEmbed{
						ID:         "embed-1",
						DiagramID:  "diagram-1",
						UserID:     "user-1",
						Token:      "test-token-64-chars-long-abcdefghijklmnopqrstuvwxyz012345678",
						IsActive:   true,
						ViewsCount: 0,
						CreatedAt:  now,
						UpdatedAt:  now,
					}, nil)

				return &Service{
					querier: q,

					randFunc: func(l int) string { return "test-token-64-chars-long-abcdefghijklmnopqrstuvwxyz012345678" },
					idFunc:   func() string { return "embed-1" },
					nowFunc:  func() time.Time { return now },
				}
			},
			want: domain.DiagramEmbedInfo{
				ID:         domain.EmbedID("embed-1"),
				DiagramID:  domain.DiagramID("diagram-1"),
				EmbedCode:  "test-token-64-chars-long-abcdefghijklmnopqrstuvwxyz012345678",
				IsActive:   true,
				ViewsCount: 0,
				CreatedAt:  now,
				UpdatedAt:  now,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.handler(t)
			got, err := svc.Create(context.Background(), tt.diagramID, tt.userID)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestService_List(t *testing.T) {
	now := time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		diagramID domain.DiagramID
		handler   func(t *testing.T) *Service
		want      []domain.DiagramEmbedInfo
		wantErr   error
	}{
		{
			name:      "Success - multiple embeds",
			diagramID: domain.DiagramID("diagram-1"),
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					EmbedListByDiagram(mock.Anything, "diagram-1").
					Return([]sqlc.DiagramEmbed{
						{
							ID:         "embed-1",
							DiagramID:  "diagram-1",
							UserID:     "user-1",
							Token:      "token-1",
							IsActive:   true,
							ViewsCount: 5,
							CreatedAt:  now,
							UpdatedAt:  now,
						},
						{
							ID:         "embed-2",
							DiagramID:  "diagram-1",
							UserID:     "user-1",
							Token:      "token-2",
							IsActive:   false,
							ViewsCount: 10,
							CreatedAt:  now,
							UpdatedAt:  now,
						},
					}, nil)

				return &Service{
					querier: q,
				}
			},
			want: []domain.DiagramEmbedInfo{
				{
					ID:         domain.EmbedID("embed-1"),
					DiagramID:  domain.DiagramID("diagram-1"),
					EmbedCode:  "token-1",
					IsActive:   true,
					ViewsCount: 5,
					CreatedAt:  now,
					UpdatedAt:  now,
				},
				{
					ID:         domain.EmbedID("embed-2"),
					DiagramID:  domain.DiagramID("diagram-1"),
					EmbedCode:  "",
					IsActive:   false,
					ViewsCount: 10,
					CreatedAt:  now,
					UpdatedAt:  now,
				},
			},
		},
		{
			name:      "Success - empty list",
			diagramID: domain.DiagramID("diagram-empty"),
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					EmbedListByDiagram(mock.Anything, "diagram-empty").
					Return([]sqlc.DiagramEmbed{}, nil)

				return &Service{
					querier: q,
				}
			},
			want: []domain.DiagramEmbedInfo{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.handler(t)
			got, err := svc.List(context.Background(), tt.diagramID)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestService_GetByID(t *testing.T) {
	now := time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		embedID domain.EmbedID
		handler func(t *testing.T) *Service
		want    domain.DiagramEmbedInfo
		wantErr error
	}{
		{
			name:    "Success",
			embedID: domain.EmbedID("embed-1"),
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					EmbedGetByID(mock.Anything, "embed-1").
					Return(sqlc.DiagramEmbed{
						ID:         "embed-1",
						DiagramID:  "diagram-1",
						UserID:     "user-1",
						Token:      "token-1",
						IsActive:   true,
						ViewsCount: 5,
						CreatedAt:  now,
						UpdatedAt:  now,
					}, nil)

				return &Service{
					querier: q,
				}
			},
			want: domain.DiagramEmbedInfo{
				ID:         domain.EmbedID("embed-1"),
				DiagramID:  domain.DiagramID("diagram-1"),
				EmbedCode:  "token-1",
				IsActive:   true,
				ViewsCount: 5,
				CreatedAt:  now,
				UpdatedAt:  now,
			},
		},
		{
			name:    "Not found returns ErrEmbedNotFound",
			embedID: domain.EmbedID("embed-nonexistent"),
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					EmbedGetByID(mock.Anything, "embed-nonexistent").
					Return(sqlc.DiagramEmbed{}, sql.ErrNoRows)

				return &Service{
					querier: q,
				}
			},
			wantErr: domain.ErrEmbedNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.handler(t)
			got, err := svc.GetByID(context.Background(), tt.embedID)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestService_GetByToken(t *testing.T) {
	now := time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		token   string
		handler func(t *testing.T) *Service
		want    domain.DiagramEmbedInfo
		wantErr error
	}{
		{
			name:  "Success - active embed",
			token: "active-token",
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					EmbedGetByToken(mock.Anything, "active-token").
					Return(sqlc.DiagramEmbed{
						ID:         "embed-1",
						DiagramID:  "diagram-1",
						UserID:     "user-1",
						Token:      "active-token",
						IsActive:   true,
						ViewsCount: 5,
						CreatedAt:  now,
						UpdatedAt:  now,
					}, nil)

				return &Service{
					querier: q,
				}
			},
			want: domain.DiagramEmbedInfo{
				ID:         domain.EmbedID("embed-1"),
				DiagramID:  domain.DiagramID("diagram-1"),
				EmbedCode:  "active-token",
				IsActive:   true,
				ViewsCount: 5,
				CreatedAt:  now,
				UpdatedAt:  now,
			},
		},
		{
			name:  "Inactive embed returns ErrEmbedNotFound",
			token: "inactive-token",
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					EmbedGetByToken(mock.Anything, "inactive-token").
					Return(sqlc.DiagramEmbed{
						ID:       "embed-1",
						Token:    "inactive-token",
						IsActive: false,
					}, nil)

				return &Service{
					querier: q,
				}
			},
			wantErr: domain.ErrEmbedNotFound,
		},
		{
			name:  "Not found returns ErrEmbedNotFound",
			token: "nonexistent-token",
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					EmbedGetByToken(mock.Anything, "nonexistent-token").
					Return(sqlc.DiagramEmbed{}, sql.ErrNoRows)

				return &Service{
					querier: q,
				}
			},
			wantErr: domain.ErrEmbedNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.handler(t)
			got, err := svc.GetByToken(context.Background(), tt.token)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestService_SoftDelete(t *testing.T) {
	tests := []struct {
		name    string
		embedID domain.EmbedID
		handler func(t *testing.T) *Service
		wantErr error
	}{
		{
			name:    "Success",
			embedID: domain.EmbedID("embed-1"),
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					EmbedSoftDelete(mock.Anything, "embed-1").
					Return(nil)

				return &Service{
					querier: q,
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.handler(t)
			err := svc.SoftDelete(context.Background(), tt.embedID)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestService_IncrementViews(t *testing.T) {
	tests := []struct {
		name    string
		embedID domain.EmbedID
		handler func(t *testing.T) *Service
		wantErr error
	}{
		{
			name:    "Success",
			embedID: domain.EmbedID("embed-1"),
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					EmbedIncrementViews(mock.Anything, "embed-1").
					Return(nil)

				return &Service{
					querier: q,
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.handler(t)
			err := svc.IncrementViews(context.Background(), tt.embedID)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}
