package handler

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"klubok/internal/domain"
	"klubok/internal/restapi/handler/mocks"
	"klubok/internal/restapi/oas"
)

func TestCollaborator_CollaboratorUpsert(t *testing.T) {
	tests := []struct {
		name    string
		req     *oas.CollaboratorUpsertRequest
		params  oas.CollaboratorUpsertParams
		ctx     func() context.Context
		handler func(t *testing.T) *Collaborator
		want    oas.CollaboratorUpsertRes
		wantErr error
	}{
		{
			name: "Success",
			req: &oas.CollaboratorUpsertRequest{
				UserNameOrEmail: "john",
				Permission:      oas.DiagramCollaboratorPermissionEnumEDIT,
			},
			params: oas.CollaboratorUpsertParams{DiagramId: "diag-1"},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Collaborator {
				t.Helper()

				userSvc := mocks.NewUserServiceMock(t)
				userSvc.EXPECT().
					FindByName(mock.Anything, "john").
					Return(&domain.User{
						ID:       "user-2",
						Email:    "john@example.com",
						Username: "john",
					}, nil)

				permSvc := mocks.NewPermissionServiceMock(t)
				permSvc.EXPECT().
					HasDiagramPermission(mock.Anything, domain.DiagramID("diag-1"), domain.UserID("user-1"), []domain.DiagramCollaboratorPermission{
						domain.DiagramCollaboratorPermissionOwner,
					}).
					Return(true, nil)

				diagramSvc := mocks.NewDiagramServiceMock(t)
				diagramSvc.EXPECT().
					CollaboratorUpsert(mock.Anything, domain.DiagramCollaborator{
						DiagramID:  "diag-1",
						UserID:     "user-2",
						Permission: domain.DiagramCollaboratorPermissionEdit,
					}).
					Return(domain.DiagramCollaborator{
						ID:         "collab-1",
						DiagramID:  "diag-1",
						UserID:     "user-2",
						UserName:   "john",
						Permission: domain.DiagramCollaboratorPermissionEdit,
						CreatedAt:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						UpdatedAt:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					}, nil)

				return &Collaborator{
					diagramService:    diagramSvc,
					userService:       userSvc,
					permissionService: permSvc,
					log:               zap.NewNop(),
				}
			},
			want: &oas.CollaboratorUpsertResponse{
				Collaborator: oas.DiagramCollaborator{
					ID:         oas.Ulid("collab-1"),
					DiagramId:  oas.Ulid("diag-1"),
					UserName:   oas.Username("john"),
					Permission: oas.DiagramCollaboratorPermissionEnumEDIT,
					CreatedAt:  oas.Datetime(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
					UpdatedAt:  oas.Datetime(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
				},
			},
		},
		{
			name: "Email rejected",
			req: &oas.CollaboratorUpsertRequest{
				UserNameOrEmail: "john@example.com",
				Permission:      oas.DiagramCollaboratorPermissionEnumEDIT,
			},
			params: oas.CollaboratorUpsertParams{DiagramId: "diag-1"},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Collaborator {
				t.Helper()

				// No userService expectations: email must be rejected without querying the DB,
				// and the existence of any account must not be revealed.
				return &Collaborator{
					userService: mocks.NewUserServiceMock(t),
					log:         zap.NewNop(),
				}
			},
			want: &oas.CollaboratorUpsertNotFound{
				Message: oas.NewOptString(domain.ErrUserNotFound.Error()),
			},
		},
		{
			name: "Unauthorized",
			req: &oas.CollaboratorUpsertRequest{
				UserNameOrEmail: "john",
				Permission:      oas.DiagramCollaboratorPermissionEnumEDIT,
			},
			params: oas.CollaboratorUpsertParams{DiagramId: "diag-1"},
			ctx:    context.Background,
			handler: func(t *testing.T) *Collaborator {
				t.Helper()

				return &Collaborator{log: zap.NewNop()}
			},
			want: &oas.CollaboratorUpsertUnauthorized{},
		},
		{
			name: "User not found",
			req: &oas.CollaboratorUpsertRequest{
				UserNameOrEmail: "unknown",
				Permission:      oas.DiagramCollaboratorPermissionEnumEDIT,
			},
			params: oas.CollaboratorUpsertParams{DiagramId: "diag-1"},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Collaborator {
				t.Helper()

				userSvc := mocks.NewUserServiceMock(t)
				userSvc.EXPECT().
					FindByName(mock.Anything, "unknown").
					Return(nil, domain.ErrUserNotFound)

				return &Collaborator{
					userService: userSvc,
					log:         zap.NewNop(),
				}
			},
			want: &oas.CollaboratorUpsertNotFound{
				Message: oas.NewOptString(domain.ErrUserNotFound.Error()),
			},
		},
		{
			name: "User can't add itself",
			req: &oas.CollaboratorUpsertRequest{
				UserNameOrEmail: "user1",
				Permission:      oas.DiagramCollaboratorPermissionEnumEDIT,
			},
			params: oas.CollaboratorUpsertParams{DiagramId: "diag-1"},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Collaborator {
				t.Helper()

				userSvc := mocks.NewUserServiceMock(t)
				userSvc.EXPECT().
					FindByName(mock.Anything, "user1").
					Return(&domain.User{
						ID:       "user-1",
						Email:    "user1@example.com",
						Username: "user1",
					}, nil)

				return &Collaborator{
					userService: userSvc,
					log:         zap.NewNop(),
				}
			},
			want: &oas.CollaboratorUpsertForbidden{},
		},
		{
			name: "Can't add owner permission",
			req: &oas.CollaboratorUpsertRequest{
				UserNameOrEmail: "john",
				Permission:      oas.DiagramCollaboratorPermissionEnumOWNER,
			},
			params: oas.CollaboratorUpsertParams{DiagramId: "diag-1"},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Collaborator {
				t.Helper()

				userSvc := mocks.NewUserServiceMock(t)
				userSvc.EXPECT().
					FindByName(mock.Anything, "john").
					Return(&domain.User{
						ID:       "user-2",
						Email:    "john@example.com",
						Username: "john",
					}, nil)

				return &Collaborator{
					userService: userSvc,
					log:         zap.NewNop(),
				}
			},
			want: &oas.CollaboratorUpsertForbidden{},
		},
		{
			name: "Forbidden - not owner",
			req: &oas.CollaboratorUpsertRequest{
				UserNameOrEmail: "john",
				Permission:      oas.DiagramCollaboratorPermissionEnumEDIT,
			},
			params: oas.CollaboratorUpsertParams{DiagramId: "diag-1"},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Collaborator {
				t.Helper()

				userSvc := mocks.NewUserServiceMock(t)
				userSvc.EXPECT().
					FindByName(mock.Anything, "john").
					Return(&domain.User{
						ID:       "user-2",
						Email:    "john@example.com",
						Username: "john",
					}, nil)

				permSvc := mocks.NewPermissionServiceMock(t)
				permSvc.EXPECT().
					HasDiagramPermission(mock.Anything, domain.DiagramID("diag-1"), domain.UserID("user-1"), []domain.DiagramCollaboratorPermission{
						domain.DiagramCollaboratorPermissionOwner,
					}).
					Return(false, nil)

				return &Collaborator{
					userService:       userSvc,
					permissionService: permSvc,
					log:               zap.NewNop(),
				}
			},
			want: &oas.CollaboratorUpsertForbidden{},
		},
		{
			name: "Service error",
			req: &oas.CollaboratorUpsertRequest{
				UserNameOrEmail: "john",
				Permission:      oas.DiagramCollaboratorPermissionEnumEDIT,
			},
			params: oas.CollaboratorUpsertParams{DiagramId: "diag-1"},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Collaborator {
				t.Helper()

				userSvc := mocks.NewUserServiceMock(t)
				userSvc.EXPECT().
					FindByName(mock.Anything, "john").
					Return(&domain.User{
						ID:       "user-2",
						Email:    "john@example.com",
						Username: "john",
					}, nil)

				permSvc := mocks.NewPermissionServiceMock(t)
				permSvc.EXPECT().
					HasDiagramPermission(mock.Anything, domain.DiagramID("diag-1"), domain.UserID("user-1"), []domain.DiagramCollaboratorPermission{
						domain.DiagramCollaboratorPermissionOwner,
					}).
					Return(true, nil)

				diagramSvc := mocks.NewDiagramServiceMock(t)
				diagramSvc.EXPECT().
					CollaboratorUpsert(mock.Anything, mock.Anything).
					Return(domain.DiagramCollaborator{}, errors.New("service error"))

				return &Collaborator{
					diagramService:    diagramSvc,
					userService:       userSvc,
					permissionService: permSvc,
					log:               zap.NewNop(),
				}
			},
			want: &oas.CollaboratorUpsertInternalServerError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := tt.handler(t)

			got, err := h.CollaboratorUpsert(tt.ctx(), tt.req, tt.params)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCollaborator_CollaboratorList(t *testing.T) {
	tests := []struct {
		name    string
		params  oas.CollaboratorListParams
		ctx     func() context.Context
		handler func(t *testing.T) *Collaborator
		want    oas.CollaboratorListRes
		wantErr error
	}{
		{
			name:   "Success",
			params: oas.CollaboratorListParams{DiagramId: "diag-1"},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Collaborator {
				t.Helper()

				permSvc := mocks.NewPermissionServiceMock(t)
				permSvc.EXPECT().
					HasDiagramPermission(mock.Anything, domain.DiagramID("diag-1"), domain.UserID("user-1"), []domain.DiagramCollaboratorPermission{
						domain.DiagramCollaboratorPermissionOwner,
					}).
					Return(true, nil)

				diagramSvc := mocks.NewDiagramServiceMock(t)
				diagramSvc.EXPECT().
					CollaboratorList(mock.Anything, domain.DiagramID("diag-1")).
					Return([]domain.DiagramCollaborator{
						{
							ID:         "collab-1",
							DiagramID:  "diag-1",
							UserID:     "user-2",
							UserName:   "john",
							Permission: domain.DiagramCollaboratorPermissionEdit,
						},
					}, nil)

				return &Collaborator{
					diagramService:    diagramSvc,
					permissionService: permSvc,
					log:               zap.NewNop(),
				}
			},
			want: &oas.CollaboratorListResponse{
				Collaborators: []oas.DiagramCollaborator{
					{
						ID:         oas.Ulid("collab-1"),
						DiagramId:  oas.Ulid("diag-1"),
						UserName:   oas.Username("john"),
						Permission: oas.DiagramCollaboratorPermissionEnumEDIT,
					},
				},
			},
		},
		{
			name:   "Unauthorized",
			params: oas.CollaboratorListParams{DiagramId: "diag-1"},
			ctx:    context.Background,
			handler: func(t *testing.T) *Collaborator {
				t.Helper()

				return &Collaborator{log: zap.NewNop()}
			},
			want: &oas.CollaboratorListUnauthorized{},
		},
		{
			name:   "Forbidden - not owner",
			params: oas.CollaboratorListParams{DiagramId: "diag-1"},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Collaborator {
				t.Helper()

				permSvc := mocks.NewPermissionServiceMock(t)
				permSvc.EXPECT().
					HasDiagramPermission(mock.Anything, domain.DiagramID("diag-1"), domain.UserID("user-1"), []domain.DiagramCollaboratorPermission{
						domain.DiagramCollaboratorPermissionOwner,
					}).
					Return(false, nil)

				return &Collaborator{
					permissionService: permSvc,
					log:               zap.NewNop(),
				}
			},
			want: &oas.CollaboratorListForbidden{},
		},
		{
			name:   "Service error",
			params: oas.CollaboratorListParams{DiagramId: "diag-1"},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Collaborator {
				t.Helper()

				permSvc := mocks.NewPermissionServiceMock(t)
				permSvc.EXPECT().
					HasDiagramPermission(mock.Anything, domain.DiagramID("diag-1"), domain.UserID("user-1"), []domain.DiagramCollaboratorPermission{
						domain.DiagramCollaboratorPermissionOwner,
					}).
					Return(true, nil)

				diagramSvc := mocks.NewDiagramServiceMock(t)
				diagramSvc.EXPECT().
					CollaboratorList(mock.Anything, domain.DiagramID("diag-1")).
					Return(nil, errors.New("service error"))

				return &Collaborator{
					diagramService:    diagramSvc,
					permissionService: permSvc,
					log:               zap.NewNop(),
				}
			},
			want: &oas.CollaboratorListInternalServerError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := tt.handler(t)

			got, err := h.CollaboratorList(tt.ctx(), tt.params)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCollaborator_CollaboratorDelete(t *testing.T) {
	tests := []struct {
		name    string
		params  oas.CollaboratorDeleteParams
		ctx     func() context.Context
		handler func(t *testing.T) *Collaborator
		want    oas.CollaboratorDeleteRes
		wantErr error
	}{
		{
			name: "Success",
			params: oas.CollaboratorDeleteParams{
				DiagramId:      "diag-1",
				CollaboratorId: oas.Ulid("user-2"),
			},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Collaborator {
				t.Helper()

				permSvc := mocks.NewPermissionServiceMock(t)
				permSvc.EXPECT().
					HasDiagramPermission(mock.Anything, domain.DiagramID("diag-1"), domain.UserID("user-1"), []domain.DiagramCollaboratorPermission{
						domain.DiagramCollaboratorPermissionOwner,
					}).
					Return(true, nil)

				diagramSvc := mocks.NewDiagramServiceMock(t)
				diagramSvc.EXPECT().
					CollaboratorDelete(mock.Anything, domain.DiagramID("diag-1"), domain.CollaboratorID("user-2")).
					Return(nil)

				return &Collaborator{
					diagramService:    diagramSvc,
					permissionService: permSvc,
					log:               zap.NewNop(),
				}
			},
			want: &oas.CollaboratorDeleteNoContent{},
		},
		{
			name: "Unauthorized",
			params: oas.CollaboratorDeleteParams{
				DiagramId:      "diag-1",
				CollaboratorId: oas.Ulid("user-2"),
			},
			ctx: context.Background,
			handler: func(t *testing.T) *Collaborator {
				t.Helper()

				return &Collaborator{log: zap.NewNop()}
			},
			want: &oas.CollaboratorDeleteUnauthorized{},
		},
		{
			name: "User can't delete yourself",
			params: oas.CollaboratorDeleteParams{
				DiagramId:      "diag-1",
				CollaboratorId: oas.Ulid("user-1"),
			},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Collaborator {
				t.Helper()

				return &Collaborator{log: zap.NewNop()}
			},
			want: &oas.CollaboratorDeleteForbidden{},
		},
		{
			name: "Forbidden - not owner",
			params: oas.CollaboratorDeleteParams{
				DiagramId:      "diag-1",
				CollaboratorId: oas.Ulid("user-2"),
			},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Collaborator {
				t.Helper()

				permSvc := mocks.NewPermissionServiceMock(t)
				permSvc.EXPECT().
					HasDiagramPermission(mock.Anything, domain.DiagramID("diag-1"), domain.UserID("user-1"), []domain.DiagramCollaboratorPermission{
						domain.DiagramCollaboratorPermissionOwner,
					}).
					Return(false, nil)

				return &Collaborator{
					permissionService: permSvc,
					log:               zap.NewNop(),
				}
			},
			want: &oas.CollaboratorDeleteForbidden{},
		},
		{
			name: "Service error",
			params: oas.CollaboratorDeleteParams{
				DiagramId:      "diag-1",
				CollaboratorId: oas.Ulid("user-2"),
			},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Collaborator {
				t.Helper()

				permSvc := mocks.NewPermissionServiceMock(t)
				permSvc.EXPECT().
					HasDiagramPermission(mock.Anything, domain.DiagramID("diag-1"), domain.UserID("user-1"), []domain.DiagramCollaboratorPermission{
						domain.DiagramCollaboratorPermissionOwner,
					}).
					Return(true, nil)

				diagramSvc := mocks.NewDiagramServiceMock(t)
				diagramSvc.EXPECT().
					CollaboratorDelete(mock.Anything, domain.DiagramID("diag-1"), domain.CollaboratorID("user-2")).
					Return(errors.New("service error"))

				return &Collaborator{
					diagramService:    diagramSvc,
					permissionService: permSvc,
					log:               zap.NewNop(),
				}
			},
			want: &oas.CollaboratorDeleteInternalServerError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := tt.handler(t)
			got, err := h.CollaboratorDelete(tt.ctx(), tt.params)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
