package permission

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"klubok/internal/application/service/permission/mocks"
	"klubok/internal/db/sqlc"
	"klubok/internal/domain"
)

var errDB = errors.New("db error")

func TestService_HasDiagramPermission(t *testing.T) {
	tests := []struct {
		name        string
		diagramID   domain.DiagramID
		userID      domain.UserID
		permissions []domain.DiagramCollaboratorPermission
		handler     func(t *testing.T) *Service
		want        bool
		wantErr     error
	}{
		{
			name:        "Has owner permission",
			diagramID:   domain.DiagramID("diagram-1"),
			userID:      domain.UserID("user-1"),
			permissions: []domain.DiagramCollaboratorPermission{domain.DiagramCollaboratorPermissionOwner},
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					DiagramCollaboratorExist(mock.Anything, "diagram-1", "user-1").
					Return(sqlc.DiagramCollaborator{
						ID:         "collab-1",
						DiagramID:  "diagram-1",
						UserID:     "user-1",
						Permission: domain.DiagramCollaboratorPermissionOwner,
						CreatedAt:  time.Now(),
						UpdatedAt:  time.Now(),
					}, nil)

				return &Service{
					querier: q,
				}
			},
			want: true,
		},
		{
			name:        "Has edit permission",
			diagramID:   domain.DiagramID("diagram-1"),
			userID:      domain.UserID("user-1"),
			permissions: []domain.DiagramCollaboratorPermission{domain.DiagramCollaboratorPermissionEdit},
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					DiagramCollaboratorExist(mock.Anything, "diagram-1", "user-1").
					Return(sqlc.DiagramCollaborator{
						ID:         "collab-1",
						DiagramID:  "diagram-1",
						UserID:     "user-1",
						Permission: domain.DiagramCollaboratorPermissionEdit,
						CreatedAt:  time.Now(),
						UpdatedAt:  time.Now(),
					}, nil)

				return &Service{
					querier: q,
				}
			},
			want: true,
		},
		{
			name:        "Has view permission",
			diagramID:   domain.DiagramID("diagram-1"),
			userID:      domain.UserID("user-1"),
			permissions: []domain.DiagramCollaboratorPermission{domain.DiagramCollaboratorPermissionView},
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					DiagramCollaboratorExist(mock.Anything, "diagram-1", "user-1").
					Return(sqlc.DiagramCollaborator{
						ID:         "collab-1",
						DiagramID:  "diagram-1",
						UserID:     "user-1",
						Permission: domain.DiagramCollaboratorPermissionView,
						CreatedAt:  time.Now(),
						UpdatedAt:  time.Now(),
					}, nil)

				return &Service{
					querier: q,
				}
			},
			want: true,
		},
		{
			name:      "Has permission - multiple allowed",
			diagramID: "diagram-1",
			userID:    "user-1",
			permissions: []domain.DiagramCollaboratorPermission{
				domain.DiagramCollaboratorPermissionOwner,
				domain.DiagramCollaboratorPermissionEdit,
			},
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					DiagramCollaboratorExist(mock.Anything, "diagram-1", "user-1").
					Return(sqlc.DiagramCollaborator{
						ID:         "collab-1",
						DiagramID:  "diagram-1",
						UserID:     "user-1",
						Permission: domain.DiagramCollaboratorPermissionEdit,
						CreatedAt:  time.Now(),
						UpdatedAt:  time.Now(),
					}, nil)

				return &Service{
					querier: q,
				}
			},
			want: true,
		},
		{
			name:        "No matching permission",
			diagramID:   domain.DiagramID("diagram-1"),
			userID:      domain.UserID("user-1"),
			permissions: []domain.DiagramCollaboratorPermission{domain.DiagramCollaboratorPermissionOwner},
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					DiagramCollaboratorExist(mock.Anything, "diagram-1", "user-1").
					Return(sqlc.DiagramCollaborator{
						ID:         "collab-1",
						DiagramID:  "diagram-1",
						UserID:     "user-1",
						Permission: domain.DiagramCollaboratorPermissionView,
						CreatedAt:  time.Now(),
						UpdatedAt:  time.Now(),
					}, nil)

				return &Service{
					querier: q,
				}
			},
			want: false,
		},
		{
			name:        "No rows - not found",
			diagramID:   domain.DiagramID("diagram-1"),
			userID:      domain.UserID("user-1"),
			permissions: []domain.DiagramCollaboratorPermission{domain.DiagramCollaboratorPermissionOwner},
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					DiagramCollaboratorExist(mock.Anything, "diagram-1", "user-1").
					Return(sqlc.DiagramCollaborator{}, sql.ErrNoRows)

				return &Service{
					querier: q,
				}
			},
			want: false,
		},
		{
			name:        "DB error",
			diagramID:   domain.DiagramID("diagram-1"),
			userID:      domain.UserID("user-1"),
			permissions: []domain.DiagramCollaboratorPermission{domain.DiagramCollaboratorPermissionOwner},
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					DiagramCollaboratorExist(mock.Anything, "diagram-1", "user-1").
					Return(sqlc.DiagramCollaborator{}, errDB)

				return &Service{
					querier: q,
				}
			},
			wantErr: errors.New("check diagram collaborator: db error"),
		},
		{
			name:        "Empty permissions list",
			diagramID:   domain.DiagramID("diagram-1"),
			userID:      domain.UserID("user-1"),
			permissions: []domain.DiagramCollaboratorPermission{},
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					DiagramCollaboratorExist(mock.Anything, "diagram-1", "user-1").
					Return(sqlc.DiagramCollaborator{
						ID:         "collab-1",
						DiagramID:  "diagram-1",
						UserID:     "user-1",
						Permission: domain.DiagramCollaboratorPermissionOwner,
						CreatedAt:  time.Now(),
						UpdatedAt:  time.Now(),
					}, nil)

				return &Service{
					querier: q,
				}
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.handler(t)
			got, err := svc.HasDiagramPermission(context.Background(), tt.diagramID, tt.userID, tt.permissions)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
