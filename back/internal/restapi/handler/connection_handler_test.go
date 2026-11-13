package handler

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"klubok/internal/domain"
	"klubok/internal/restapi/handler/mocks"
	"klubok/internal/restapi/oas"
)

func TestConnection_ConnectionUpsert(t *testing.T) {
	tests := []struct {
		name    string
		req     *oas.ConnectionUpsertRequest
		params  oas.ConnectionUpsertParams
		ctx     func() context.Context
		handler func(t *testing.T) *Connection
		want    oas.ConnectionUpsertRes
		wantErr error
	}{
		{
			name: "Success",
			req: &oas.ConnectionUpsertRequest{
				Connection: oas.DiagramConnection{
					Fid: oas.Ulid("shape-1"),
					Tid: oas.Ulid("shape-2"),
				},
			},
			params: oas.ConnectionUpsertParams{DiagramId: "diag-1"},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Connection {
				t.Helper()

				permSvc := mocks.NewPermissionServiceMock(t)
				permSvc.EXPECT().
					HasDiagramPermission(mock.Anything, domain.DiagramID("diag-1"), domain.UserID("user-1"), []domain.DiagramCollaboratorPermission{
						domain.DiagramCollaboratorPermissionOwner,
						domain.DiagramCollaboratorPermissionEdit,
					}).
					Return(true, nil)

				diagramSvc := mocks.NewDiagramServiceMock(t)
				diagramSvc.EXPECT().
					ConnectionUpsert(mock.Anything, domain.UserID("user-1"), domain.DiagramID("diag-1"), domain.DiagramConnection{
						FromShapeID: "shape-1",
						ToShapeID:   "shape-2",
					}).
					Return(nil)

				return &Connection{
					diagramService:    diagramSvc,
					permissionService: permSvc,
					log:               zap.NewNop(),
				}
			},
			want: &oas.ConnectionUpsertNoContent{},
		},
		{
			name: "Unauthorized",
			req: &oas.ConnectionUpsertRequest{
				Connection: oas.DiagramConnection{
					Fid: oas.Ulid("shape-1"),
					Tid: oas.Ulid("shape-2"),
				},
			},
			params: oas.ConnectionUpsertParams{DiagramId: "diag-1"},
			ctx:    context.Background,
			handler: func(t *testing.T) *Connection {
				t.Helper()

				return &Connection{log: zap.NewNop()}
			},
			want: &oas.ConnectionUpsertUnauthorized{},
		},
		{
			name: "Forbidden - no edit permission",
			req: &oas.ConnectionUpsertRequest{
				Connection: oas.DiagramConnection{
					Fid: oas.Ulid("shape-1"),
					Tid: oas.Ulid("shape-2"),
				},
			},
			params: oas.ConnectionUpsertParams{DiagramId: "diag-1"},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Connection {
				t.Helper()

				permSvc := mocks.NewPermissionServiceMock(t)
				permSvc.EXPECT().
					HasDiagramPermission(mock.Anything, domain.DiagramID("diag-1"), domain.UserID("user-1"), []domain.DiagramCollaboratorPermission{
						domain.DiagramCollaboratorPermissionOwner,
						domain.DiagramCollaboratorPermissionEdit,
					}).
					Return(false, nil)

				return &Connection{
					permissionService: permSvc,
					log:               zap.NewNop(),
				}
			},
			want: &oas.ConnectionUpsertForbidden{},
		},
		{
			name: "Service error",
			req: &oas.ConnectionUpsertRequest{
				Connection: oas.DiagramConnection{
					Fid: oas.Ulid("shape-1"),
					Tid: oas.Ulid("shape-2"),
				},
			},
			params: oas.ConnectionUpsertParams{DiagramId: "diag-1"},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Connection {
				t.Helper()

				permSvc := mocks.NewPermissionServiceMock(t)
				permSvc.EXPECT().
					HasDiagramPermission(mock.Anything, domain.DiagramID("diag-1"), domain.UserID("user-1"), []domain.DiagramCollaboratorPermission{
						domain.DiagramCollaboratorPermissionOwner,
						domain.DiagramCollaboratorPermissionEdit,
					}).
					Return(true, nil)

				diagramSvc := mocks.NewDiagramServiceMock(t)
				diagramSvc.EXPECT().
					ConnectionUpsert(mock.Anything, domain.UserID("user-1"), domain.DiagramID("diag-1"), mock.Anything).
					Return(errors.New("service error"))

				return &Connection{
					diagramService:    diagramSvc,
					permissionService: permSvc,
					log:               zap.NewNop(),
				}
			},
			want: &oas.ConnectionUpsertInternalServerError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := tt.handler(t)
			got, err := h.ConnectionUpsert(tt.ctx(), tt.req, tt.params)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestConnection_ConnectionDelete(t *testing.T) {
	tests := []struct {
		name    string
		params  oas.ConnectionDeleteParams
		ctx     func() context.Context
		handler func(t *testing.T) *Connection
		want    oas.ConnectionDeleteRes
		wantErr error
	}{
		{
			name: "Success",
			params: oas.ConnectionDeleteParams{
				DiagramId:   "diag-1",
				FromShapeId: oas.Ulid("shape-1"),
				ToShapeId:   oas.Ulid("shape-2"),
			},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Connection {
				t.Helper()

				permSvc := mocks.NewPermissionServiceMock(t)
				permSvc.EXPECT().
					HasDiagramPermission(mock.Anything, domain.DiagramID("diag-1"), domain.UserID("user-1"), []domain.DiagramCollaboratorPermission{
						domain.DiagramCollaboratorPermissionOwner,
						domain.DiagramCollaboratorPermissionEdit,
					}).
					Return(true, nil)

				diagramSvc := mocks.NewDiagramServiceMock(t)
				diagramSvc.EXPECT().
					ConnectionDelete(mock.Anything, domain.UserID("user-1"), domain.DiagramID("diag-1"), domain.DiagramConnection{
						FromShapeID: "shape-1",
						ToShapeID:   "shape-2",
					}).
					Return(nil)

				return &Connection{
					diagramService:    diagramSvc,
					permissionService: permSvc,
					log:               zap.NewNop(),
				}
			},
			want: &oas.ConnectionDeleteNoContent{},
		},
		{
			name: "Unauthorized",
			params: oas.ConnectionDeleteParams{
				DiagramId:   "diag-1",
				FromShapeId: oas.Ulid("shape-1"),
				ToShapeId:   oas.Ulid("shape-2"),
			},
			ctx: context.Background,
			handler: func(t *testing.T) *Connection {
				t.Helper()

				return &Connection{log: zap.NewNop()}
			},
			want: &oas.ConnectionDeleteUnauthorized{},
		},
		{
			name: "Forbidden - no edit permission",
			params: oas.ConnectionDeleteParams{
				DiagramId:   "diag-1",
				FromShapeId: oas.Ulid("shape-1"),
				ToShapeId:   oas.Ulid("shape-2"),
			},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Connection {
				t.Helper()

				permSvc := mocks.NewPermissionServiceMock(t)
				permSvc.EXPECT().
					HasDiagramPermission(mock.Anything, domain.DiagramID("diag-1"), domain.UserID("user-1"), []domain.DiagramCollaboratorPermission{
						domain.DiagramCollaboratorPermissionOwner,
						domain.DiagramCollaboratorPermissionEdit,
					}).
					Return(false, nil)

				return &Connection{
					permissionService: permSvc,
					log:               zap.NewNop(),
				}
			},
			want: &oas.ConnectionDeleteForbidden{},
		},
		{
			name: "Service error",
			params: oas.ConnectionDeleteParams{
				DiagramId:   "diag-1",
				FromShapeId: oas.Ulid("shape-1"),
				ToShapeId:   oas.Ulid("shape-2"),
			},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Connection {
				t.Helper()

				permSvc := mocks.NewPermissionServiceMock(t)
				permSvc.EXPECT().
					HasDiagramPermission(mock.Anything, domain.DiagramID("diag-1"), domain.UserID("user-1"), []domain.DiagramCollaboratorPermission{
						domain.DiagramCollaboratorPermissionOwner,
						domain.DiagramCollaboratorPermissionEdit,
					}).
					Return(true, nil)

				diagramSvc := mocks.NewDiagramServiceMock(t)
				diagramSvc.EXPECT().
					ConnectionDelete(mock.Anything, domain.UserID("user-1"), domain.DiagramID("diag-1"), mock.Anything).
					Return(errors.New("service error"))

				return &Connection{
					diagramService:    diagramSvc,
					permissionService: permSvc,
					log:               zap.NewNop(),
				}
			},
			want: &oas.ConnectionDeleteInternalServerError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := tt.handler(t)
			got, err := h.ConnectionDelete(tt.ctx(), tt.params)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestConnection_ConnectionBatchUpsert(t *testing.T) {
	tests := []struct {
		name    string
		req     *oas.ConnectionBatchUpsertRequest
		params  oas.ConnectionBatchUpsertParams
		ctx     func() context.Context
		handler func(t *testing.T) *Connection
		want    oas.ConnectionBatchUpsertRes
		wantErr error
	}{
		{
			name: "Success",
			req: &oas.ConnectionBatchUpsertRequest{
				Connections: []oas.DiagramConnection{
					{
						Fid: oas.Ulid("shape-1"),
						Tid: oas.Ulid("shape-2"),
					},
					{
						Fid: oas.Ulid("shape-2"),
						Tid: oas.Ulid("shape-3"),
					},
				},
			},
			params: oas.ConnectionBatchUpsertParams{DiagramId: "diag-1"},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Connection {
				t.Helper()

				permSvc := mocks.NewPermissionServiceMock(t)
				permSvc.EXPECT().
					HasDiagramPermission(mock.Anything, domain.DiagramID("diag-1"), domain.UserID("user-1"), []domain.DiagramCollaboratorPermission{
						domain.DiagramCollaboratorPermissionOwner,
						domain.DiagramCollaboratorPermissionEdit,
					}).
					Return(true, nil)

				diagramSvc := mocks.NewDiagramServiceMock(t)
				diagramSvc.EXPECT().
					ConnectionBatchUpsert(mock.Anything, domain.UserID("user-1"), domain.DiagramID("diag-1"), []domain.DiagramConnection{
						{
							FromShapeID: "shape-1",
							ToShapeID:   "shape-2",
						},
						{
							FromShapeID: "shape-2",
							ToShapeID:   "shape-3",
						},
					}).
					Return(nil)

				return &Connection{
					diagramService:    diagramSvc,
					permissionService: permSvc,
					log:               zap.NewNop(),
				}
			},
			want: &oas.ConnectionBatchUpsertNoContent{},
		},
		{
			name: "Unauthorized",
			req: &oas.ConnectionBatchUpsertRequest{
				Connections: []oas.DiagramConnection{
					{
						Fid: oas.Ulid("shape-1"),
						Tid: oas.Ulid("shape-2"),
					},
				},
			},
			params: oas.ConnectionBatchUpsertParams{DiagramId: "diag-1"},
			ctx:    context.Background,
			handler: func(t *testing.T) *Connection {
				t.Helper()

				return &Connection{log: zap.NewNop()}
			},
			want: &oas.ConnectionBatchUpsertUnauthorized{},
		},
		{
			name: "Forbidden - no edit permission",
			req: &oas.ConnectionBatchUpsertRequest{
				Connections: []oas.DiagramConnection{
					{
						Fid: oas.Ulid("shape-1"),
						Tid: oas.Ulid("shape-2"),
					},
				},
			},
			params: oas.ConnectionBatchUpsertParams{DiagramId: "diag-1"},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Connection {
				t.Helper()

				permSvc := mocks.NewPermissionServiceMock(t)
				permSvc.EXPECT().
					HasDiagramPermission(mock.Anything, domain.DiagramID("diag-1"), domain.UserID("user-1"), []domain.DiagramCollaboratorPermission{
						domain.DiagramCollaboratorPermissionOwner,
						domain.DiagramCollaboratorPermissionEdit,
					}).
					Return(false, nil)

				return &Connection{
					permissionService: permSvc,
					log:               zap.NewNop(),
				}
			},
			want: &oas.ConnectionBatchUpsertForbidden{},
		},
		{
			name: "Empty connections list",
			req: &oas.ConnectionBatchUpsertRequest{
				Connections: []oas.DiagramConnection{},
			},
			params: oas.ConnectionBatchUpsertParams{DiagramId: "diag-1"},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Connection {
				t.Helper()

				permSvc := mocks.NewPermissionServiceMock(t)
				permSvc.EXPECT().
					HasDiagramPermission(mock.Anything, domain.DiagramID("diag-1"), domain.UserID("user-1"), []domain.DiagramCollaboratorPermission{
						domain.DiagramCollaboratorPermissionOwner,
						domain.DiagramCollaboratorPermissionEdit,
					}).
					Return(true, nil)

				diagramSvc := mocks.NewDiagramServiceMock(t)
				diagramSvc.EXPECT().
					ConnectionBatchUpsert(mock.Anything, domain.UserID("user-1"), domain.DiagramID("diag-1"), []domain.DiagramConnection{}).
					Return(nil)

				return &Connection{
					diagramService:    diagramSvc,
					permissionService: permSvc,
					log:               zap.NewNop(),
				}
			},
			want: &oas.ConnectionBatchUpsertNoContent{},
		},
		{
			name: "Single connection",
			req: &oas.ConnectionBatchUpsertRequest{
				Connections: []oas.DiagramConnection{
					{
						Fid: oas.Ulid("shape-1"),
						Tid: oas.Ulid("shape-5"),
					},
				},
			},
			params: oas.ConnectionBatchUpsertParams{DiagramId: "diag-1"},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Connection {
				t.Helper()

				permSvc := mocks.NewPermissionServiceMock(t)
				permSvc.EXPECT().
					HasDiagramPermission(mock.Anything, domain.DiagramID("diag-1"), domain.UserID("user-1"), []domain.DiagramCollaboratorPermission{
						domain.DiagramCollaboratorPermissionOwner,
						domain.DiagramCollaboratorPermissionEdit,
					}).
					Return(true, nil)

				diagramSvc := mocks.NewDiagramServiceMock(t)
				diagramSvc.EXPECT().
					ConnectionBatchUpsert(mock.Anything, domain.UserID("user-1"), domain.DiagramID("diag-1"), []domain.DiagramConnection{
						{
							FromShapeID: "shape-1",
							ToShapeID:   "shape-5",
						},
					}).
					Return(nil)

				return &Connection{
					diagramService:    diagramSvc,
					permissionService: permSvc,
					log:               zap.NewNop(),
				}
			},
			want: &oas.ConnectionBatchUpsertNoContent{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := tt.handler(t)
			got, err := h.ConnectionBatchUpsert(tt.ctx(), tt.req, tt.params)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
