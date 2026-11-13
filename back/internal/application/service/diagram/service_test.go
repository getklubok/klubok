package diagram

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"klubok/internal/application/service/diagram/mocks"
	"klubok/internal/db"
	"klubok/internal/db/sqlc"
	"klubok/internal/domain"
)

var errDB = errors.New("db error")

func TestService_Create(t *testing.T) {
	tests := []struct {
		name    string
		title   string
		userID  domain.UserID
		handler func(t *testing.T) *Service
		want    domain.Diagram
		wantErr error
	}{
		{
			name:   "Success",
			title:  "Test Diagram",
			userID: domain.UserID("user-1"),
			handler: func(t *testing.T) *Service {
				t.Helper()

				dbUC := mocks.NewDBUseCaseMock(t)
				dbUC.EXPECT().
					DiagramCreate(mock.Anything, "generated-id", "Test Diagram", sqlc.DiagramCollaboratorUpsertParams{
						ID:         "generated-id",
						DiagramID:  "generated-id",
						UserID:     "user-1",
						Permission: domain.DiagramCollaboratorPermissionOwner,
					}).
					Return(sqlc.DiagramGetByUserRow{
						ID:         "generated-id",
						Title:      "Test Diagram",
						CreatedAt:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						UpdatedAt:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						Permission: domain.DiagramCollaboratorPermissionOwner,
					}, nil)

				return &Service{
					dbUseCase: dbUC,
					log:       zap.NewNop(),
					idFunc:    func() string { return "generated-id" },
					nowFunc:   func() time.Time { return time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC) },
				}
			},
			want: domain.Diagram{
				ID:         domain.DiagramID("generated-id"),
				Title:      "Test Diagram",
				Permission: domain.DiagramCollaboratorPermissionOwner,
				CreatedAt:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name:   "DBUseCase error",
			title:  "Test Diagram",
			userID: domain.UserID("user-1"),
			handler: func(t *testing.T) *Service {
				t.Helper()

				dbUC := mocks.NewDBUseCaseMock(t)
				dbUC.EXPECT().
					DiagramCreate(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(sqlc.DiagramGetByUserRow{}, errDB)

				return &Service{
					dbUseCase: dbUC,
					log:       zap.NewNop(),
					idFunc:    func() string { return "generated-id" },
					nowFunc:   func() time.Time { return time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC) },
				}
			},
			wantErr: errors.New("create diagram: db error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.handler(t)
			got, err := svc.Create(context.Background(), tt.title, tt.userID)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestService_List(t *testing.T) {
	tests := []struct {
		name    string
		userID  domain.UserID
		handler func(t *testing.T) *Service
		want    []domain.Diagram
		wantErr error
	}{
		{
			name:   "Success - multiple diagrams",
			userID: "user-1",
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					DiagramList(mock.Anything, "user-1").
					Return([]sqlc.DiagramListRow{
						{
							ID:               "diagram-1",
							Title:            "Diagram 1",
							Permission:       domain.DiagramCollaboratorPermissionOwner,
							HasCollaborators: false,
							CreatedAt:        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
							UpdatedAt:        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						},
						{
							ID:               "diagram-2",
							Title:            "Diagram 2",
							Permission:       domain.DiagramCollaboratorPermissionEdit,
							HasCollaborators: true,
							CreatedAt:        time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
							UpdatedAt:        time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
						},
					}, nil)

				return &Service{
					querier: q,
					log:     zap.NewNop(),
				}
			},
			want: []domain.Diagram{
				{
					ID:               domain.DiagramID("diagram-1"),
					Title:            "Diagram 1",
					Permission:       domain.DiagramCollaboratorPermissionOwner,
					HasCollaborators: false,
					CreatedAt:        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					UpdatedAt:        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				},
				{
					ID:               domain.DiagramID("diagram-2"),
					Title:            "Diagram 2",
					Permission:       domain.DiagramCollaboratorPermissionEdit,
					HasCollaborators: true,
					CreatedAt:        time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
					UpdatedAt:        time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
				},
			},
		},
		{
			name:   "Success - empty list",
			userID: "user-1",
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					DiagramList(mock.Anything, "user-1").
					Return([]sqlc.DiagramListRow{}, nil)

				return &Service{
					querier: q,
					log:     zap.NewNop(),
				}
			},
			want: []domain.Diagram{},
		},
		{
			name:   "DB error",
			userID: "user-1",
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					DiagramList(mock.Anything, "user-1").
					Return(nil, errDB)

				return &Service{
					querier: q,
					log:     zap.NewNop(),
				}
			},
			wantErr: errors.New("find diagrams: db error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.handler(t)
			got, err := svc.List(context.Background(), tt.userID)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestService_Delete(t *testing.T) {
	tests := []struct {
		name      string
		diagramID domain.DiagramID
		handler   func(t *testing.T) *Service
		wantErr   error
	}{
		{
			name:      "Success",
			diagramID: "diagram-1",
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					DiagramDelete(mock.Anything, "diagram-1").
					Return(nil)

				return &Service{
					querier: q,
					log:     zap.NewNop(),
				}
			},
		},
		{
			name:      "DB error",
			diagramID: "diagram-1",
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					DiagramDelete(mock.Anything, "diagram-1").
					Return(errDB)

				return &Service{
					querier: q,
					log:     zap.NewNop(),
				}
			},
			wantErr: errors.New("delete diagram: db error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.handler(t)
			err := svc.Delete(context.Background(), tt.diagramID)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_SetTitle(t *testing.T) {
	tests := []struct {
		name    string
		id      domain.DiagramID
		title   string
		handler func(t *testing.T) *Service
		wantErr error
	}{
		{
			name:  "Success",
			id:    domain.DiagramID("diagram-1"),
			title: "New Title",
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					DiagramSetTitle(mock.Anything, "diagram-1", "New Title").
					Return(nil)

				return &Service{
					querier: q,
					log:     zap.NewNop(),
				}
			},
		},
		{
			name:  "DB error",
			id:    domain.DiagramID("diagram-1"),
			title: "New Title",
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					DiagramSetTitle(mock.Anything, "diagram-1", "New Title").
					Return(errDB)

				return &Service{
					querier: q,
					log:     zap.NewNop(),
				}
			},
			wantErr: errors.New("set title diagram: db error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.handler(t)
			err := svc.SetTitle(context.Background(), tt.id, tt.title)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_Info(t *testing.T) {
	tests := []struct {
		name      string
		diagramID domain.DiagramID
		handler   func(t *testing.T) *Service
		want      domain.DiagramInfo
		wantErr   error
	}{
		{
			name:      "Success",
			diagramID: "diagram-1",
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					DiagramGet(mock.Anything, "diagram-1").
					Return(sqlc.Diagram{
						ID:        "diagram-1",
						Title:     "Test Diagram",
						CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						UpdatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					}, nil)

				q.EXPECT().
					DiagramShapeList(mock.Anything, "diagram-1").
					Return([]sqlc.DiagramShape{
						{
							ID:        "shape-1",
							DiagramID: "diagram-1",
							Data: domain.DiagramShape{
								ID:   domain.ShapeID("shape-1"),
								Type: domain.DiagramShapeTypeSquare,
							},
						},
					}, nil)

				q.EXPECT().
					DiagramConnectionList(mock.Anything, "diagram-1").
					Return([]sqlc.DiagramConnection{
						{
							FromShapeID: "shape-1",
							ToShapeID:   "shape-2",
						},
					}, nil)

				q.EXPECT().
					DiagramCommentStats(mock.Anything, "diagram-1").
					Return([]sqlc.DiagramCommentStatsRow{
						{
							ShapeID:      "shape-1",
							CommentCount: 5,
						},
					}, nil)

				return &Service{
					querier: q,
					log:     zap.NewNop(),
				}
			},
			want: domain.DiagramInfo{
				ID:         domain.DiagramID("diagram-1"),
				Title:      "Test Diagram",
				Permission: domain.DiagramCollaboratorPermissionView,
				Shapes: []domain.DiagramShape{
					{
						ID:   domain.ShapeID("shape-1"),
						Type: domain.DiagramShapeTypeSquare,
					},
				},
				Connections: []domain.DiagramConnection{
					{
						FromShapeID: domain.ShapeID("shape-1"),
						ToShapeID:   domain.ShapeID("shape-2"),
					},
				},
				ShapeCommentStats: []domain.DiagramShapeCommentStat{
					{
						ShapeID:      domain.ShapeID("shape-1"),
						CommentCount: 5,
					},
				},
			},
		},
		{
			name:      "Diagram not found",
			diagramID: domain.DiagramID("diagram-1"),
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					DiagramGet(mock.Anything, "diagram-1").
					Return(sqlc.Diagram{}, sql.ErrNoRows)

				return &Service{
					querier: q,
					log:     zap.NewNop(),
				}
			},
			wantErr: domain.ErrDiagramNotFound,
		},
		{
			name:      "DiagramGet DB error",
			diagramID: domain.DiagramID("diagram-1"),
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					DiagramGet(mock.Anything, "diagram-1").
					Return(sqlc.Diagram{}, errDB)

				return &Service{
					querier: q,
					log:     zap.NewNop(),
				}
			},
			wantErr: errors.New("get diagram: db error"),
		},
		{
			name:      "DiagramShapeList error",
			diagramID: domain.DiagramID("diagram-1"),
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					DiagramGet(mock.Anything, "diagram-1").
					Return(sqlc.Diagram{
						ID:    "diagram-1",
						Title: "Test Diagram",
					}, nil)

				q.EXPECT().
					DiagramShapeList(mock.Anything, "diagram-1").
					Return(nil, errDB)

				return &Service{
					querier: q,
					log:     zap.NewNop(),
				}
			},
			wantErr: errors.New("get shapes, connections and comments: get diagram shapes: db error"),
		},
		{
			name:      "DiagramConnectionList error",
			diagramID: domain.DiagramID("diagram-1"),
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					DiagramGet(mock.Anything, "diagram-1").
					Return(sqlc.Diagram{
						ID:    "diagram-1",
						Title: "Test Diagram",
					}, nil)

				q.EXPECT().
					DiagramShapeList(mock.Anything, "diagram-1").
					Return([]sqlc.DiagramShape{}, nil)

				q.EXPECT().
					DiagramConnectionList(mock.Anything, "diagram-1").
					Return(nil, errDB)

				return &Service{
					querier: q,
					log:     zap.NewNop(),
				}
			},
			wantErr: errors.New("get shapes, connections and comments: get diagram connections: db error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.handler(t)
			got, err := svc.Info(context.Background(), tt.diagramID)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestService_InfoByUser(t *testing.T) {
	tests := []struct {
		name      string
		diagramID domain.DiagramID
		userID    domain.UserID
		handler   func(t *testing.T) *Service
		want      domain.DiagramInfo
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
					DiagramGetByUser(mock.Anything, "user-1", "diagram-1").
					Return(sqlc.DiagramGetByUserRow{
						ID:         "diagram-1",
						Title:      "Test Diagram",
						Permission: domain.DiagramCollaboratorPermissionOwner,
						CreatedAt:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						UpdatedAt:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					}, nil)

				q.EXPECT().
					DiagramShapeList(mock.Anything, "diagram-1").
					Return([]sqlc.DiagramShape{
						{
							ID:        "shape-1",
							DiagramID: "diagram-1",
							Data: domain.DiagramShape{
								ID:   domain.ShapeID("shape-1"),
								Type: domain.DiagramShapeTypeSquare,
							},
						},
					}, nil)

				q.EXPECT().
					DiagramConnectionList(mock.Anything, "diagram-1").
					Return([]sqlc.DiagramConnection{
						{
							FromShapeID: "shape-1",
							ToShapeID:   "shape-2",
						},
					}, nil)

				q.EXPECT().
					DiagramCommentStats(mock.Anything, "diagram-1").
					Return([]sqlc.DiagramCommentStatsRow{
						{
							ShapeID:      "shape-1",
							CommentCount: 5,
						},
					}, nil)

				return &Service{
					querier: q,
					log:     zap.NewNop(),
				}
			},
			want: domain.DiagramInfo{
				ID:         domain.DiagramID("diagram-1"),
				Title:      "Test Diagram",
				Permission: domain.DiagramCollaboratorPermissionOwner,
				Shapes: []domain.DiagramShape{
					{
						ID:   domain.ShapeID("shape-1"),
						Type: domain.DiagramShapeTypeSquare,
					},
				},
				Connections: []domain.DiagramConnection{
					{
						FromShapeID: domain.ShapeID("shape-1"),
						ToShapeID:   domain.ShapeID("shape-2"),
					},
				},
				ShapeCommentStats: []domain.DiagramShapeCommentStat{
					{
						ShapeID:      domain.ShapeID("shape-1"),
						CommentCount: 5,
					},
				},
			},
		},
		{
			name:      "Diagram not found",
			diagramID: domain.DiagramID("diagram-1"),
			userID:    domain.UserID("user-1"),
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					DiagramGetByUser(mock.Anything, "user-1", "diagram-1").
					Return(sqlc.DiagramGetByUserRow{}, sql.ErrNoRows)

				return &Service{
					querier: q,
					log:     zap.NewNop(),
				}
			},
			wantErr: domain.ErrDiagramNotFound,
		},
		{
			name:      "DiagramGet DB error",
			diagramID: domain.DiagramID("diagram-1"),
			userID:    domain.UserID("user-1"),
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					DiagramGetByUser(mock.Anything, "user-1", "diagram-1").
					Return(sqlc.DiagramGetByUserRow{}, errDB)

				return &Service{
					querier: q,
					log:     zap.NewNop(),
				}
			},
			wantErr: errors.New("get diagram: db error"),
		},
		{
			name:      "DiagramShapeList error",
			diagramID: domain.DiagramID("diagram-1"),
			userID:    domain.UserID("user-1"),
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					DiagramGetByUser(mock.Anything, "user-1", "diagram-1").
					Return(sqlc.DiagramGetByUserRow{
						ID:         "diagram-1",
						Title:      "Test Diagram",
						Permission: domain.DiagramCollaboratorPermissionOwner,
					}, nil)

				q.EXPECT().
					DiagramShapeList(mock.Anything, "diagram-1").
					Return(nil, errDB)

				return &Service{
					querier: q,
					log:     zap.NewNop(),
				}
			},
			wantErr: errors.New("get shapes, connections and comments: get diagram shapes: db error"),
		},
		{
			name:      "DiagramConnectionList error",
			diagramID: domain.DiagramID("diagram-1"),
			userID:    domain.UserID("user-1"),
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					DiagramGetByUser(mock.Anything, "user-1", "diagram-1").
					Return(sqlc.DiagramGetByUserRow{
						ID:         "diagram-1",
						Title:      "Test Diagram",
						Permission: domain.DiagramCollaboratorPermissionOwner,
					}, nil)

				q.EXPECT().
					DiagramShapeList(mock.Anything, "diagram-1").
					Return([]sqlc.DiagramShape{}, nil)

				q.EXPECT().
					DiagramConnectionList(mock.Anything, "diagram-1").
					Return(nil, errDB)

				return &Service{
					querier: q,
					log:     zap.NewNop(),
				}
			},
			wantErr: errors.New("get shapes, connections and comments: get diagram connections: db error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.handler(t)
			got, err := svc.InfoByUser(context.Background(), tt.diagramID, tt.userID)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestService_ShapeUpsert(t *testing.T) {
	tests := []struct {
		name      string
		userID    domain.UserID
		diagramID domain.DiagramID
		shape     domain.DiagramShape
		handler   func(t *testing.T) *Service
		wantErr   error
	}{
		{
			name:      "Success",
			userID:    domain.UserID("user-1"),
			diagramID: domain.DiagramID("diagram-1"),
			shape: domain.DiagramShape{
				ID:     domain.ShapeID("shape-1"),
				Type:   domain.DiagramShapeTypeSquare,
				Left:   10,
				Top:    20,
				Width:  30,
				Height: 40,
				ScaleX: 1,
				ScaleY: 1,
				Fill:   "red",
				Stroke: "blue",
				Text:   "text",
			},
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					DiagramShapeUpsert(mock.Anything, sqlc.DiagramShapeUpsertParams{
						ID:        "shape-1",
						DiagramID: "diagram-1",
						Data: domain.DiagramShape{
							ID:     domain.ShapeID("shape-1"),
							Type:   domain.DiagramShapeTypeSquare,
							Left:   10,
							Top:    20,
							Width:  30,
							Height: 40,
							ScaleX: 1,
							ScaleY: 1,
							Fill:   "red",
							Stroke: "blue",
							Text:   "text",
						},
					}).
					Return(nil)

				ws := mocks.NewWebSocketManagerMock(t)
				ws.EXPECT().
					BroadcastToDiagram(domain.DiagramID("diagram-1"), mock.Anything, mock.Anything).
					Return(nil)

				return &Service{
					querier:   q,
					wsManager: ws,
					log:       zap.NewNop(),
					nowFunc:   func() time.Time { return time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC) },
				}
			},
		},
		{
			name:      "DB error",
			userID:    domain.UserID("user-1"),
			diagramID: domain.DiagramID("diagram-1"),
			shape: domain.DiagramShape{
				ID:   domain.ShapeID("shape-1"),
				Type: domain.DiagramShapeTypeSquare,
			},
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					DiagramShapeUpsert(mock.Anything, mock.Anything).
					Return(errDB)

				return &Service{
					querier: q,
					log:     zap.NewNop(),
					nowFunc: func() time.Time { return time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC) },
				}
			},
			wantErr: errors.New("upsert diagram shape: db error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.handler(t)
			err := svc.ShapeUpsert(context.Background(), tt.userID, tt.diagramID, tt.shape)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_ShapeDelete(t *testing.T) {
	tests := []struct {
		name      string
		userID    domain.UserID
		diagramID domain.DiagramID
		shapeID   domain.ShapeID
		handler   func(t *testing.T) *Service
		wantErr   error
	}{
		{
			name:      "Success",
			userID:    domain.UserID("user-1"),
			diagramID: domain.DiagramID("diagram-1"),
			shapeID:   domain.ShapeID("shape-1"),
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					DiagramShapeDelete(mock.Anything, "shape-1", "diagram-1").
					Return(nil)

				ws := mocks.NewWebSocketManagerMock(t)
				ws.EXPECT().
					BroadcastToDiagram(domain.DiagramID("diagram-1"), mock.Anything, mock.Anything).
					Return(nil)

				return &Service{
					querier:   q,
					wsManager: ws,
					log:       zap.NewNop(),
					nowFunc:   func() time.Time { return time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC) },
				}
			},
		},
		{
			name:      "DB error",
			userID:    "user-1",
			diagramID: "diagram-1",
			shapeID:   domain.ShapeID("shape-1"),
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					DiagramShapeDelete(mock.Anything, mock.Anything, mock.Anything).
					Return(errDB)

				return &Service{
					querier: q,
					log:     zap.NewNop(),
					nowFunc: func() time.Time { return time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC) },
				}
			},
			wantErr: errors.New("delete diagram shape: db error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.handler(t)
			err := svc.ShapeDelete(context.Background(), tt.userID, tt.diagramID, tt.shapeID)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_ConnectionUpsert(t *testing.T) {
	tests := []struct {
		name       string
		userID     domain.UserID
		diagramID  domain.DiagramID
		connection domain.DiagramConnection
		handler    func(t *testing.T) *Service
		wantErr    error
	}{
		{
			name:      "Success",
			userID:    domain.UserID("user-1"),
			diagramID: domain.DiagramID("diagram-1"),
			connection: domain.DiagramConnection{
				FromShapeID: domain.ShapeID("shape-1"),
				ToShapeID:   domain.ShapeID("shape-2"),
			},
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					DiagramConnectionUpsert(mock.Anything, sqlc.DiagramConnectionUpsertParams{
						DiagramID:   "diagram-1",
						FromShapeID: "shape-1",
						ToShapeID:   "shape-2",
					}).
					Return(nil)

				ws := mocks.NewWebSocketManagerMock(t)
				ws.EXPECT().
					BroadcastToDiagram(domain.DiagramID("diagram-1"), mock.Anything, mock.Anything).
					Return(nil)

				return &Service{
					querier:   q,
					wsManager: ws,
					log:       zap.NewNop(),
					nowFunc:   func() time.Time { return time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC) },
				}
			},
		},
		{
			name:      "DB error",
			userID:    domain.UserID("user-1"),
			diagramID: domain.DiagramID("diagram-1"),
			connection: domain.DiagramConnection{
				FromShapeID: domain.ShapeID("shape-1"),
				ToShapeID:   domain.ShapeID("shape-2"),
			},
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					DiagramConnectionUpsert(mock.Anything, mock.Anything).
					Return(errDB)

				return &Service{
					querier: q,
					log:     zap.NewNop(),
					nowFunc: func() time.Time { return time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC) },
				}
			},
			wantErr: errors.New("upsert diagram connection: db error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.handler(t)
			err := svc.ConnectionUpsert(context.Background(), tt.userID, tt.diagramID, tt.connection)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_ConnectionDelete(t *testing.T) {
	tests := []struct {
		name       string
		userID     domain.UserID
		diagramID  domain.DiagramID
		connection domain.DiagramConnection
		handler    func(t *testing.T) *Service
		wantErr    error
	}{
		{
			name:      "Success",
			userID:    domain.UserID("user-1"),
			diagramID: domain.DiagramID("diagram-1"),
			connection: domain.DiagramConnection{
				FromShapeID: domain.ShapeID("shape-1"),
				ToShapeID:   domain.ShapeID("shape-2"),
			},
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					DiagramConnectionDelete(mock.Anything, sqlc.DiagramConnectionDeleteParams{
						DiagramID:   "diagram-1",
						FromShapeID: "shape-1",
						ToShapeID:   "shape-2",
					}).
					Return(nil)

				ws := mocks.NewWebSocketManagerMock(t)
				ws.EXPECT().
					BroadcastToDiagram(domain.DiagramID("diagram-1"), mock.Anything, mock.Anything).
					Return(nil)

				return &Service{
					querier:   q,
					wsManager: ws,
					log:       zap.NewNop(),
					nowFunc:   func() time.Time { return time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC) },
				}
			},
		},
		{
			name:      "DB error",
			userID:    domain.UserID("user-1"),
			diagramID: domain.DiagramID("diagram-1"),
			connection: domain.DiagramConnection{
				FromShapeID: domain.ShapeID("shape-1"),
				ToShapeID:   domain.ShapeID("shape-2"),
			},
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					DiagramConnectionDelete(mock.Anything, mock.Anything).
					Return(errDB)

				return &Service{
					querier: q,
					log:     zap.NewNop(),
					nowFunc: func() time.Time { return time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC) },
				}
			},
			wantErr: errors.New("delete diagram connection: db error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.handler(t)
			err := svc.ConnectionDelete(context.Background(), tt.userID, tt.diagramID, tt.connection)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_CommentCreate(t *testing.T) {
	tests := []struct {
		name    string
		userID  domain.UserID
		comment domain.DiagramComment
		handler func(t *testing.T) *Service
		want    domain.DiagramComment
		wantErr error
	}{
		{
			name:   "Success",
			userID: domain.UserID("user-1"),
			comment: domain.DiagramComment{
				DiagramID:      domain.DiagramID("diagram-1"),
				DiagramShapeID: domain.ShapeID("shape-1"),
				UserID:         domain.UserID("user-1"),
				Text:           "Test comment",
			},
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					DiagramCommentCreate(mock.Anything, mock.Anything).
					Return(sqlc.DiagramComment{
						ID:             "comment-1",
						DiagramID:      "diagram-1",
						DiagramShapeID: "shape-1",
						UserID:         "user-1",
						Text:           "Test comment",
						CreatedAt:      time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					}, nil)

				ws := mocks.NewWebSocketManagerMock(t)
				ws.EXPECT().
					BroadcastToDiagram(domain.DiagramID("diagram-1"), mock.Anything, mock.Anything).
					Return(nil)

				return &Service{
					querier:   q,
					wsManager: ws,
					log:       zap.NewNop(),
					idFunc:    func() string { return "generated-id" },
					nowFunc:   func() time.Time { return time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC) },
				}
			},
			want: domain.DiagramComment{
				ID:             domain.CommentID("comment-1"),
				DiagramID:      domain.DiagramID("diagram-1"),
				DiagramShapeID: domain.ShapeID("shape-1"),
				UserID:         domain.UserID("user-1"),
				Text:           "Test comment",
				CreatedAt:      time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name:   "DB error",
			userID: domain.UserID("user-1"),
			comment: domain.DiagramComment{
				DiagramID:      domain.DiagramID("diagram-1"),
				DiagramShapeID: domain.ShapeID("shape-1"),
				UserID:         domain.UserID("user-1"),
				Text:           "Test comment",
			},
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					DiagramCommentCreate(mock.Anything, mock.Anything).
					Return(sqlc.DiagramComment{}, errDB)

				return &Service{
					querier: q,
					log:     zap.NewNop(),
					idFunc:  func() string { return "generated-id" },
					nowFunc: func() time.Time { return time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC) },
				}
			},
			wantErr: errors.New("create diagram comment: db error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.handler(t)
			got, err := svc.CommentCreate(context.Background(), tt.userID, tt.comment)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestService_CommentList(t *testing.T) {
	tests := []struct {
		name           string
		diagramID      domain.DiagramID
		diagramShapeID domain.ShapeID
		handler        func(t *testing.T) *Service
		want           []domain.DiagramComment
		wantErr        error
	}{
		{
			name:           "Success",
			diagramID:      domain.DiagramID("diagram-1"),
			diagramShapeID: domain.ShapeID("shape-1"),
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					DiagramCommentList(mock.Anything, "diagram-1", "shape-1").
					Return([]sqlc.DiagramCommentListRow{
						{
							ID:             "comment-1",
							DiagramID:      "diagram-1",
							DiagramShapeID: "shape-1",
							UserID:         "user-1",
							Text:           "Test comment",
							UserName:       "testuser",
							CreatedAt:      time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						},
					}, nil)

				return &Service{
					querier: q,
					log:     zap.NewNop(),
				}
			},
			want: []domain.DiagramComment{
				{
					ID:             domain.CommentID("comment-1"),
					DiagramID:      domain.DiagramID("diagram-1"),
					DiagramShapeID: domain.ShapeID("shape-1"),
					UserID:         domain.UserID("user-1"),
					UserName:       "testuser",
					Text:           "Test comment",
					CreatedAt:      time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				},
			},
		},
		{
			name:           "DB error",
			diagramID:      domain.DiagramID("diagram-1"),
			diagramShapeID: domain.ShapeID("shape-1"),
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					DiagramCommentList(mock.Anything, "diagram-1", "shape-1").
					Return(nil, errDB)

				return &Service{
					querier: q,
					log:     zap.NewNop(),
				}
			},
			wantErr: errors.New("get diagram comments: db error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.handler(t)
			got, err := svc.CommentList(context.Background(), tt.diagramID, tt.diagramShapeID)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestService_CollaboratorUpsert(t *testing.T) {
	tests := []struct {
		name         string
		collaborator domain.DiagramCollaborator
		handler      func(t *testing.T) *Service
		want         domain.DiagramCollaborator
		wantErr      error
	}{
		{
			name: "Success",
			collaborator: domain.DiagramCollaborator{
				DiagramID:  domain.DiagramID("diagram-1"),
				UserID:     domain.UserID("user-1"),
				Permission: domain.DiagramCollaboratorPermissionEdit,
			},
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					DiagramCollaboratorUpsert(mock.Anything, sqlc.DiagramCollaboratorUpsertParams{
						ID:         "generated-id",
						DiagramID:  "diagram-1",
						UserID:     "user-1",
						Permission: domain.DiagramCollaboratorPermissionEdit,
					}).
					Return(sqlc.DiagramCollaborator{
						ID:         "collab-1",
						DiagramID:  "diagram-1",
						UserID:     "user-1",
						Permission: domain.DiagramCollaboratorPermissionEdit,
						CreatedAt:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						UpdatedAt:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					}, nil)

				return &Service{
					querier: q,
					log:     zap.NewNop(),
					idFunc:  func() string { return "generated-id" },
				}
			},
			want: domain.DiagramCollaborator{
				ID:         domain.CollaboratorID("collab-1"),
				DiagramID:  domain.DiagramID("diagram-1"),
				UserID:     domain.UserID("user-1"),
				Permission: domain.DiagramCollaboratorPermissionEdit,
				CreatedAt:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "DB error",
			collaborator: domain.DiagramCollaborator{
				DiagramID:  domain.DiagramID("diagram-1"),
				UserID:     domain.UserID("user-1"),
				Permission: domain.DiagramCollaboratorPermissionEdit,
			},
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					DiagramCollaboratorUpsert(mock.Anything, mock.Anything).
					Return(sqlc.DiagramCollaborator{}, errDB)

				return &Service{
					querier: q,
					log:     zap.NewNop(),
					idFunc:  func() string { return "generated-id" },
				}
			},
			wantErr: errors.New("upsert diagram collaborator: db error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.handler(t)
			got, err := svc.CollaboratorUpsert(context.Background(), tt.collaborator)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestService_CollaboratorList(t *testing.T) {
	tests := []struct {
		name      string
		diagramID domain.DiagramID
		handler   func(t *testing.T) *Service
		want      []domain.DiagramCollaborator
		wantErr   error
	}{
		{
			name:      "Success",
			diagramID: domain.DiagramID("diagram-1"),
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					DiagramCollaboratorList(mock.Anything, "diagram-1").
					Return([]sqlc.DiagramCollaboratorListRow{
						{
							ID:         "collab-1",
							DiagramID:  "diagram-1",
							UserID:     "user-1",
							UserName:   "user1",
							Permission: domain.DiagramCollaboratorPermissionEdit,
							CreatedAt:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
							UpdatedAt:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						},
					}, nil)

				return &Service{
					querier: q,
					log:     zap.NewNop(),
				}
			},
			want: []domain.DiagramCollaborator{
				{
					ID:         domain.CollaboratorID("collab-1"),
					DiagramID:  domain.DiagramID("diagram-1"),
					UserID:     domain.UserID("user-1"),
					UserName:   "user1",
					Permission: domain.DiagramCollaboratorPermissionEdit,
					CreatedAt:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					UpdatedAt:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				},
			},
		},
		{
			name:      "DB error",
			diagramID: domain.DiagramID("diagram-1"),
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					DiagramCollaboratorList(mock.Anything, "diagram-1").
					Return(nil, errDB)

				return &Service{
					querier: q,
					log:     zap.NewNop(),
				}
			},
			wantErr: errors.New("get diagram collaborators: db error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.handler(t)
			got, err := svc.CollaboratorList(context.Background(), tt.diagramID)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestService_CollaboratorDelete(t *testing.T) {
	tests := []struct {
		name           string
		diagramID      domain.DiagramID
		collaboratorID domain.CollaboratorID
		handler        func(t *testing.T) *Service
		wantErr        error
	}{
		{
			name:           "Success",
			diagramID:      domain.DiagramID("diagram-1"),
			collaboratorID: domain.CollaboratorID("collab-1"),
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					DiagramCollaboratorDelete(mock.Anything, "diagram-1", "collab-1").
					Return(nil)

				return &Service{
					querier: q,
					log:     zap.NewNop(),
				}
			},
		},
		{
			name:           "DB error",
			diagramID:      "diagram-1",
			collaboratorID: domain.CollaboratorID("collab-1"),
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					DiagramCollaboratorDelete(mock.Anything, "diagram-1", "collab-1").
					Return(errDB)

				return &Service{
					querier: q,
					log:     zap.NewNop(),
				}
			},
			wantErr: errors.New("delete diagram collaborator: db error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.handler(t)
			err := svc.CollaboratorDelete(context.Background(), tt.diagramID, tt.collaboratorID)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_OnlineUsers(t *testing.T) {
	tests := []struct {
		name      string
		diagramID domain.DiagramID
		handler   func(t *testing.T) *Service
		want      []domain.WebSocketUser
	}{
		{
			name:      "Success",
			diagramID: domain.DiagramID("diagram-1"),
			handler: func(t *testing.T) *Service {
				t.Helper()

				ws := mocks.NewWebSocketManagerMock(t)
				ws.EXPECT().
					GetDiagramUsers(domain.DiagramID("diagram-1")).
					Return([]domain.WebSocketUser{
						{
							ID:       "user-1",
							Username: "testuser",
						},
					})

				return &Service{
					wsManager: ws,
					log:       zap.NewNop(),
				}
			},
			want: []domain.WebSocketUser{
				{
					ID:       domain.UserID("user-1"),
					Username: "testuser",
				},
			},
		},
		{
			name:      "Empty",
			diagramID: "diagram-1",
			handler: func(t *testing.T) *Service {
				t.Helper()

				ws := mocks.NewWebSocketManagerMock(t)
				ws.EXPECT().
					GetDiagramUsers(domain.DiagramID("diagram-1")).
					Return(nil)

				return &Service{
					wsManager: ws,
					log:       zap.NewNop(),
				}
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.handler(t)
			got := svc.OnlineUsers(context.Background(), tt.diagramID)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestService_Clone(t *testing.T) {
	tests := []struct {
		name      string
		diagramID domain.DiagramID
		userID    domain.UserID
		handler   func(t *testing.T) *Service
		want      domain.Diagram
		wantErr   error
	}{
		{
			name:      "Success",
			diagramID: domain.DiagramID("src-diagram-1"),
			userID:    domain.UserID("user-1"),
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					DiagramGetByUser(mock.Anything, "user-1", "src-diagram-1").
					Return(sqlc.DiagramGetByUserRow{
						ID:         "src-diagram-1",
						Title:      "Original",
						Permission: domain.DiagramCollaboratorPermissionOwner,
						CreatedAt:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						UpdatedAt:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					}, nil)

				q.EXPECT().
					DiagramShapeList(mock.Anything, "src-diagram-1").
					Return([]sqlc.DiagramShape{
						{
							ID:        "shape-1",
							DiagramID: "src-diagram-1",
							Data: domain.DiagramShape{
								ID:   domain.ShapeID("shape-1"),
								Type: domain.DiagramShapeTypeSquare,
								Left: 100,
								Top:  100,
							},
						},
					}, nil)

				q.EXPECT().
					DiagramConnectionList(mock.Anything, "src-diagram-1").
					Return([]sqlc.DiagramConnection{
						{FromShapeID: "shape-1", ToShapeID: "shape-2"},
					}, nil)

				q.EXPECT().
					DiagramCommentStats(mock.Anything, "src-diagram-1").
					Return([]sqlc.DiagramCommentStatsRow{}, nil)

				q.EXPECT().
					DiagramCommentFlatList(mock.Anything, "src-diagram-1").
					Return([]sqlc.DiagramComment{}, nil)

				q.EXPECT().
					DiagramCollaboratorFlatList(mock.Anything, "src-diagram-1").
					Return([]sqlc.DiagramCollaborator{
						{
							ID:         "collab-1",
							DiagramID:  "src-diagram-1",
							UserID:     "user-2",
							Permission: domain.DiagramCollaboratorPermissionEdit,
						},
					}, nil)

				dbUC := mocks.NewDBUseCaseMock(t)
				dbUC.EXPECT().
					DiagramClone(mock.Anything, mock.MatchedBy(func(params db.DiagramCloneParams) bool {
						return params.NewTitle == "Original (copy)" &&
							params.OwnerUserID == "user-1" &&
							len(params.Shapes) == 1 &&
							len(params.Connections) == 1 &&
							len(params.Comments) == 0 &&
							len(params.Collaborators) == 1 &&
							params.Collaborators[0].UserID == "user-2" &&
							params.Collaborators[0].Permission == domain.DiagramCollaboratorPermissionEdit
					})).
					Return(sqlc.DiagramGetByUserRow{
						ID:         "new-diagram-1",
						Title:      "Original (copy)",
						Permission: domain.DiagramCollaboratorPermissionOwner,
						CreatedAt:  time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
						UpdatedAt:  time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
					}, nil)

				return &Service{
					querier:   q,
					dbUseCase: dbUC,
					log:       zap.NewNop(),
					idFunc:    func() string { return "generated-id" },
				}
			},
			want: domain.Diagram{
				ID:         domain.DiagramID("new-diagram-1"),
				Title:      "Original (copy)",
				Permission: domain.DiagramCollaboratorPermissionOwner,
				CreatedAt:  time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt:  time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name:      "Success - skips OWNER and current user from collaborators",
			diagramID: domain.DiagramID("src-diagram-1"),
			userID:    domain.UserID("user-1"),
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					DiagramGetByUser(mock.Anything, "user-1", "src-diagram-1").
					Return(sqlc.DiagramGetByUserRow{
						ID:         "src-diagram-1",
						Title:      "Original",
						Permission: domain.DiagramCollaboratorPermissionOwner,
					}, nil)
				q.EXPECT().
					DiagramShapeList(mock.Anything, "src-diagram-1").
					Return([]sqlc.DiagramShape{}, nil)
				q.EXPECT().
					DiagramConnectionList(mock.Anything, "src-diagram-1").
					Return([]sqlc.DiagramConnection{}, nil)
				q.EXPECT().
					DiagramCommentStats(mock.Anything, "src-diagram-1").
					Return([]sqlc.DiagramCommentStatsRow{}, nil)
				q.EXPECT().
					DiagramCommentFlatList(mock.Anything, "src-diagram-1").
					Return([]sqlc.DiagramComment{}, nil)
				q.EXPECT().
					DiagramCollaboratorFlatList(mock.Anything, "src-diagram-1").
					Return([]sqlc.DiagramCollaborator{
						{
							ID:         "collab-owner",
							DiagramID:  "src-diagram-1",
							UserID:     "user-99",
							Permission: domain.DiagramCollaboratorPermissionOwner,
						},
						{
							ID:         "collab-same-user",
							DiagramID:  "src-diagram-1",
							UserID:     "user-1",
							Permission: domain.DiagramCollaboratorPermissionEdit,
						},
						{
							ID:         "collab-viewer",
							DiagramID:  "src-diagram-1",
							UserID:     "user-3",
							Permission: domain.DiagramCollaboratorPermissionView,
						},
					}, nil)

				dbUC := mocks.NewDBUseCaseMock(t)
				dbUC.EXPECT().
					DiagramClone(mock.Anything, mock.MatchedBy(func(params db.DiagramCloneParams) bool {
						return len(params.Collaborators) == 1 &&
							params.Collaborators[0].UserID == "user-3" &&
							params.Collaborators[0].Permission == domain.DiagramCollaboratorPermissionView
					})).
					Return(sqlc.DiagramGetByUserRow{
						ID:         "new-diagram-1",
						Title:      "Original (copy)",
						Permission: domain.DiagramCollaboratorPermissionOwner,
						CreatedAt:  time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
						UpdatedAt:  time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
					}, nil)

				return &Service{
					querier:   q,
					dbUseCase: dbUC,
					log:       zap.NewNop(),
					idFunc:    func() string { return "generated-id" },
				}
			},
			want: domain.Diagram{
				ID:         domain.DiagramID("new-diagram-1"),
				Title:      "Original (copy)",
				Permission: domain.DiagramCollaboratorPermissionOwner,
				CreatedAt:  time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt:  time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name:      "Source diagram not found",
			diagramID: domain.DiagramID("nonexistent"),
			userID:    domain.UserID("user-1"),
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					DiagramGetByUser(mock.Anything, "user-1", "nonexistent").
					Return(sqlc.DiagramGetByUserRow{}, sql.ErrNoRows)

				return &Service{
					querier: q,
					log:     zap.NewNop(),
				}
			},
			wantErr: errors.New("get source diagram info: diagram not found"),
		},
		{
			name:      "Info service error",
			diagramID: domain.DiagramID("src-diagram-1"),
			userID:    domain.UserID("user-1"),
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					DiagramGetByUser(mock.Anything, "user-1", "src-diagram-1").
					Return(sqlc.DiagramGetByUserRow{}, errDB)

				return &Service{
					querier: q,
					log:     zap.NewNop(),
				}
			},
			wantErr: errors.New("get source diagram info: get diagram: db error"),
		},
		{
			name:      "Comment list error",
			diagramID: domain.DiagramID("src-diagram-1"),
			userID:    domain.UserID("user-1"),
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					DiagramGetByUser(mock.Anything, "user-1", "src-diagram-1").
					Return(sqlc.DiagramGetByUserRow{
						ID:         "src-diagram-1",
						Title:      "Original",
						Permission: domain.DiagramCollaboratorPermissionOwner,
					}, nil)
				q.EXPECT().
					DiagramShapeList(mock.Anything, "src-diagram-1").
					Return([]sqlc.DiagramShape{}, nil)
				q.EXPECT().
					DiagramConnectionList(mock.Anything, "src-diagram-1").
					Return([]sqlc.DiagramConnection{}, nil)
				q.EXPECT().
					DiagramCommentStats(mock.Anything, "src-diagram-1").
					Return([]sqlc.DiagramCommentStatsRow{}, nil)
				q.EXPECT().
					DiagramCommentFlatList(mock.Anything, "src-diagram-1").
					Return(nil, errDB)

				return &Service{
					querier: q,
					log:     zap.NewNop(),
				}
			},
			wantErr: errors.New("get diagram comments: db error"),
		},
		{
			name:      "Collaborator list error",
			diagramID: domain.DiagramID("src-diagram-1"),
			userID:    domain.UserID("user-1"),
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					DiagramGetByUser(mock.Anything, "user-1", "src-diagram-1").
					Return(sqlc.DiagramGetByUserRow{
						ID:         "src-diagram-1",
						Title:      "Original",
						Permission: domain.DiagramCollaboratorPermissionOwner,
					}, nil)
				q.EXPECT().
					DiagramShapeList(mock.Anything, "src-diagram-1").
					Return([]sqlc.DiagramShape{}, nil)
				q.EXPECT().
					DiagramConnectionList(mock.Anything, "src-diagram-1").
					Return([]sqlc.DiagramConnection{}, nil)
				q.EXPECT().
					DiagramCommentStats(mock.Anything, "src-diagram-1").
					Return([]sqlc.DiagramCommentStatsRow{}, nil)
				q.EXPECT().
					DiagramCommentFlatList(mock.Anything, "src-diagram-1").
					Return([]sqlc.DiagramComment{}, nil)
				q.EXPECT().
					DiagramCollaboratorFlatList(mock.Anything, "src-diagram-1").
					Return(nil, errDB)

				return &Service{
					querier: q,
					log:     zap.NewNop(),
				}
			},
			wantErr: errors.New("get diagram collaborators: db error"),
		},
		{
			name:      "DiagramClone error",
			diagramID: domain.DiagramID("src-diagram-1"),
			userID:    domain.UserID("user-1"),
			handler: func(t *testing.T) *Service {
				t.Helper()

				q := mocks.NewQuerierMock(t)
				q.EXPECT().
					DiagramGetByUser(mock.Anything, "user-1", "src-diagram-1").
					Return(sqlc.DiagramGetByUserRow{
						ID:         "src-diagram-1",
						Title:      "Original",
						Permission: domain.DiagramCollaboratorPermissionOwner,
					}, nil)
				q.EXPECT().
					DiagramShapeList(mock.Anything, "src-diagram-1").
					Return([]sqlc.DiagramShape{}, nil)
				q.EXPECT().
					DiagramConnectionList(mock.Anything, "src-diagram-1").
					Return([]sqlc.DiagramConnection{}, nil)
				q.EXPECT().
					DiagramCommentStats(mock.Anything, "src-diagram-1").
					Return([]sqlc.DiagramCommentStatsRow{}, nil)
				q.EXPECT().
					DiagramCommentFlatList(mock.Anything, "src-diagram-1").
					Return([]sqlc.DiagramComment{}, nil)
				q.EXPECT().
					DiagramCollaboratorFlatList(mock.Anything, "src-diagram-1").
					Return([]sqlc.DiagramCollaborator{}, nil)

				dbUC := mocks.NewDBUseCaseMock(t)
				dbUC.EXPECT().
					DiagramClone(mock.Anything, mock.Anything).
					Return(sqlc.DiagramGetByUserRow{}, errDB)

				return &Service{
					querier:   q,
					dbUseCase: dbUC,
					log:       zap.NewNop(),
					idFunc:    func() string { return "generated-id" },
				}
			},
			wantErr: errors.New("clone diagram: db error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.handler(t)
			got, err := svc.Clone(context.Background(), tt.diagramID, tt.userID)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestService_ShapeBatchUpsert(t *testing.T) {
	tests := []struct {
		name      string
		userID    domain.UserID
		diagramID domain.DiagramID
		shapes    []domain.DiagramShape
		handler   func(t *testing.T) *Service
		wantErr   error
	}{
		{
			name:      "Success",
			userID:    domain.UserID("user-1"),
			diagramID: domain.DiagramID("diagram-1"),
			shapes: []domain.DiagramShape{
				{
					ID:     domain.ShapeID("shape-1"),
					Type:   domain.DiagramShapeTypeSquare,
					Left:   10,
					Top:    20,
					Width:  30,
					Height: 40,
					ScaleX: 1,
					ScaleY: 1,
					Fill:   "red",
					Stroke: "blue",
					Text:   "text1",
				},
				{
					ID:     domain.ShapeID("shape-2"),
					Type:   domain.DiagramShapeTypeEllipse,
					Left:   50,
					Top:    60,
					Width:  70,
					Height: 80,
					ScaleX: 1,
					ScaleY: 1,
					Fill:   "green",
					Stroke: "black",
					Text:   "text2",
				},
			},
			handler: func(t *testing.T) *Service {
				t.Helper()

				dbUC := mocks.NewDBUseCaseMock(t)
				dbUC.EXPECT().
					DiagramShapeBatchUpsert(mock.Anything, "diagram-1", []domain.DiagramShape{
						{
							ID:     domain.ShapeID("shape-1"),
							Type:   domain.DiagramShapeTypeSquare,
							Left:   10,
							Top:    20,
							Width:  30,
							Height: 40,
							ScaleX: 1,
							ScaleY: 1,
							Fill:   "red",
							Stroke: "blue",
							Text:   "text1",
						},
						{
							ID:     domain.ShapeID("shape-2"),
							Type:   domain.DiagramShapeTypeEllipse,
							Left:   50,
							Top:    60,
							Width:  70,
							Height: 80,
							ScaleX: 1,
							ScaleY: 1,
							Fill:   "green",
							Stroke: "black",
							Text:   "text2",
						},
					}).
					Return(nil)

				ws := mocks.NewWebSocketManagerMock(t)
				ws.EXPECT().
					BroadcastToDiagram(domain.DiagramID("diagram-1"), mock.Anything, mock.Anything).
					Return(nil).
					Times(2)

				return &Service{
					dbUseCase: dbUC,
					wsManager: ws,
					log:       zap.NewNop(),
					nowFunc:   func() time.Time { return time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC) },
				}
			},
		},
		{
			name:      "Empty shapes list - returns early",
			userID:    domain.UserID("user-1"),
			diagramID: domain.DiagramID("diagram-1"),
			shapes:    []domain.DiagramShape{},
			handler: func(t *testing.T) *Service {
				t.Helper()
				return &Service{
					log: zap.NewNop(),
				}
			},
		},
		{
			name:      "Nil shapes list - returns early",
			userID:    domain.UserID("user-1"),
			diagramID: domain.DiagramID("diagram-1"),
			shapes:    nil,
			handler: func(t *testing.T) *Service {
				t.Helper()
				return &Service{
					log: zap.NewNop(),
				}
			},
		},
		{
			name:      "Single shape",
			userID:    "user-1",
			diagramID: domain.DiagramID("diagram-1"),
			shapes: []domain.DiagramShape{
				{
					ID:     domain.ShapeID("shape-1"),
					Type:   domain.DiagramShapeTypeDiamond,
					Left:   100,
					Top:    200,
					Width:  300,
					Height: 400,
					ScaleX: 1.5,
					ScaleY: 1.5,
					Fill:   "yellow",
					Stroke: "gray",
					Text:   "diamond",
				},
			},
			handler: func(t *testing.T) *Service {
				t.Helper()

				dbUC := mocks.NewDBUseCaseMock(t)
				dbUC.EXPECT().
					DiagramShapeBatchUpsert(mock.Anything, "diagram-1", []domain.DiagramShape{
						{
							ID:     domain.ShapeID("shape-1"),
							Type:   domain.DiagramShapeTypeDiamond,
							Left:   100,
							Top:    200,
							Width:  300,
							Height: 400,
							ScaleX: 1.5,
							ScaleY: 1.5,
							Fill:   "yellow",
							Stroke: "gray",
							Text:   "diamond",
						},
					}).
					Return(nil)

				ws := mocks.NewWebSocketManagerMock(t)
				ws.EXPECT().
					BroadcastToDiagram(domain.DiagramID("diagram-1"), mock.Anything, mock.Anything).
					Return(nil)

				return &Service{
					dbUseCase: dbUC,
					wsManager: ws,
					log:       zap.NewNop(),
					nowFunc:   func() time.Time { return time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC) },
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.handler(t)
			err := svc.ShapeBatchUpsert(context.Background(), tt.userID, tt.diagramID, tt.shapes)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_ConnectionBatchUpsert(t *testing.T) {
	tests := []struct {
		name        string
		userID      domain.UserID
		diagramID   domain.DiagramID
		connections []domain.DiagramConnection
		handler     func(t *testing.T) *Service
		wantErr     error
	}{
		{
			name:      "Success",
			userID:    domain.UserID("user-1"),
			diagramID: domain.DiagramID("diagram-1"),
			connections: []domain.DiagramConnection{
				{
					FromShapeID: domain.ShapeID("shape-1"),
					ToShapeID:   domain.ShapeID("shape-2"),
				},
				{
					FromShapeID: domain.ShapeID("shape-2"),
					ToShapeID:   domain.ShapeID("shape-3"),
				},
			},
			handler: func(t *testing.T) *Service {
				t.Helper()

				dbUC := mocks.NewDBUseCaseMock(t)
				dbUC.EXPECT().
					DiagramConnectionBatchUpsert(mock.Anything, "diagram-1", []domain.DiagramConnection{
						{
							FromShapeID: domain.ShapeID("shape-1"),
							ToShapeID:   domain.ShapeID("shape-2"),
						},
						{
							FromShapeID: domain.ShapeID("shape-2"),
							ToShapeID:   domain.ShapeID("shape-3"),
						},
					}).
					Return(nil)

				ws := mocks.NewWebSocketManagerMock(t)
				ws.EXPECT().
					BroadcastToDiagram(domain.DiagramID("diagram-1"), mock.Anything, mock.Anything).
					Return(nil).
					Times(2)

				return &Service{
					dbUseCase: dbUC,
					wsManager: ws,
					log:       zap.NewNop(),
					nowFunc:   func() time.Time { return time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC) },
				}
			},
		},
		{
			name:        "Empty connections list - returns early",
			userID:      domain.UserID("user-1"),
			diagramID:   domain.DiagramID("diagram-1"),
			connections: []domain.DiagramConnection{},
			handler: func(t *testing.T) *Service {
				t.Helper()
				return &Service{
					log: zap.NewNop(),
				}
			},
		},
		{
			name:        "Nil connections list - returns early",
			userID:      domain.UserID("user-1"),
			diagramID:   domain.DiagramID("diagram-1"),
			connections: nil,
			handler: func(t *testing.T) *Service {
				t.Helper()
				return &Service{
					log: zap.NewNop(),
				}
			},
		},
		{
			name:      "Single connection",
			userID:    "user-1",
			diagramID: domain.DiagramID("diagram-1"),
			connections: []domain.DiagramConnection{
				{
					FromShapeID: domain.ShapeID("shape-1"),
					ToShapeID:   domain.ShapeID("shape-3"),
				},
			},
			handler: func(t *testing.T) *Service {
				t.Helper()

				dbUC := mocks.NewDBUseCaseMock(t)
				dbUC.EXPECT().
					DiagramConnectionBatchUpsert(mock.Anything, "diagram-1", []domain.DiagramConnection{
						{
							FromShapeID: domain.ShapeID("shape-1"),
							ToShapeID:   domain.ShapeID("shape-3"),
						},
					}).
					Return(nil)

				ws := mocks.NewWebSocketManagerMock(t)
				ws.EXPECT().
					BroadcastToDiagram(domain.DiagramID("diagram-1"), mock.Anything, mock.Anything).
					Return(nil)

				return &Service{
					dbUseCase: dbUC,
					wsManager: ws,
					log:       zap.NewNop(),
					nowFunc:   func() time.Time { return time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC) },
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.handler(t)
			err := svc.ConnectionBatchUpsert(context.Background(), tt.userID, tt.diagramID, tt.connections)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_Import(t *testing.T) {
	tests := []struct {
		name        string
		userID      domain.UserID
		diagramInfo domain.DiagramInfo
		handler     func(t *testing.T) *Service
		want        domain.Diagram
		wantErr     error
	}{
		{
			name:   "Success",
			userID: domain.UserID("user-1"),
			diagramInfo: domain.DiagramInfo{
				Title:      "Original",
				Permission: domain.DiagramCollaboratorPermissionOwner,
				Shapes: []domain.DiagramShape{
					{
						ID:     domain.ShapeID("shape-1"),
						Type:   "SQUARE",
						Left:   10,
						Top:    20,
						Width:  30,
						Height: 40,
						ScaleX: 1,
						ScaleY: 1,
						Fill:   "#fff",
						Stroke: "#000",
						Text:   "text",
						Order:  "a0",
					},
					{
						ID:     domain.ShapeID("shape-2"),
						Type:   "ELLIPSE",
						Left:   50,
						Top:    60,
						Width:  70,
						Height: 80,
						ScaleX: 1,
						ScaleY: 1,
						Fill:   "#aaa",
						Stroke: "#111",
						Text:   "text",
						Order:  "a1",
					},
					{
						ID:     domain.ShapeID("shape-3"),
						Type:   "DIAMOND",
						Left:   90,
						Top:    90,
						Width:  90,
						Height: 90,
						ScaleX: 1,
						ScaleY: 1,
						Fill:   "#aaa",
						Stroke: "#111",
						Text:   "text",
						Order:  "a2",
					},
				},
				Connections: []domain.DiagramConnection{
					{
						FromShapeID: domain.ShapeID("shape-1"),
						ToShapeID:   domain.ShapeID("shape-2"),
					},
					{
						FromShapeID: domain.ShapeID("shape-2"),
						ToShapeID:   domain.ShapeID("shape-3"),
					},
				},
			},
			handler: func(t *testing.T) *Service {
				t.Helper()

				dbUC := mocks.NewDBUseCaseMock(t)
				dbUC.EXPECT().
					DiagramClone(mock.Anything, db.DiagramCloneParams{
						NewDiagramID: "generated-id",
						NewTitle:     "Original (imported 2024-02-01 00:00)",
						OwnerUserID:  "user-1",
						Shapes: []db.DiagramShapeClone{
							{
								ID: "generated-id",
								Data: domain.DiagramShape{
									ID:     domain.ShapeID("generated-id"),
									Type:   "SQUARE",
									Left:   10,
									Top:    20,
									Width:  30,
									Height: 40,
									ScaleX: 1,
									ScaleY: 1,
									Fill:   "#fff",
									Stroke: "#000",
									Text:   "text",
									Order:  "a0",
								},
							},
							{
								ID: "generated-id",
								Data: domain.DiagramShape{
									ID:     domain.ShapeID("generated-id"),
									Type:   "ELLIPSE",
									Left:   50,
									Top:    60,
									Width:  70,
									Height: 80,
									ScaleX: 1,
									ScaleY: 1,
									Fill:   "#aaa",
									Stroke: "#111",
									Text:   "text",
									Order:  "a1",
								},
							},
							{
								ID: "generated-id",
								Data: domain.DiagramShape{
									ID:     domain.ShapeID("generated-id"),
									Type:   "DIAMOND",
									Left:   90,
									Top:    90,
									Width:  90,
									Height: 90,
									ScaleX: 1,
									ScaleY: 1,
									Fill:   "#aaa",
									Stroke: "#111",
									Text:   "text",
									Order:  "a2",
								},
							},
						},
						Connections: []db.DiagramConnectionClone{
							{
								FromShapeID: "generated-id",
								ToShapeID:   "generated-id",
							},
							{
								FromShapeID: "generated-id",
								ToShapeID:   "generated-id",
							},
						},
					}).
					Return(sqlc.DiagramGetByUserRow{
						ID:         "generated-id",
						Title:      "Original (imported 2026-05-08 12:34)",
						Permission: domain.DiagramCollaboratorPermissionOwner,
						CreatedAt:  time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
						UpdatedAt:  time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
					}, nil)

				return &Service{
					dbUseCase: dbUC,
					log:       zap.NewNop(),
					idFunc:    func() string { return "generated-id" },
					nowFunc:   func() time.Time { return time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC) },
				}
			},
			want: domain.Diagram{
				ID:         domain.DiagramID("generated-id"),
				Title:      "Original (imported 2026-05-08 12:34)",
				Permission: domain.DiagramCollaboratorPermissionOwner,
				CreatedAt:  time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt:  time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.handler(t)
			got, err := svc.Import(context.Background(), tt.userID, tt.diagramInfo)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
