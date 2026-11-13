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

func TestShape_ShapeUpsert(t *testing.T) {
	tests := []struct {
		name    string
		req     *oas.ShapeUpsertRequest
		params  oas.ShapeUpsertParams
		ctx     func() context.Context
		handler func(t *testing.T) *Shape
		want    oas.ShapeUpsertRes
		wantErr error
	}{
		{
			name: "Success",
			req: &oas.ShapeUpsertRequest{
				Shape: oas.DiagramShape{
					ID:  oas.Ulid("shape-1"),
					Tp:  oas.DiagramShapeTypeEnumS,
					L:   10,
					T:   20,
					W:   30,
					H:   40,
					Sx:  1,
					Sy:  1,
					Fc:  "#fff",
					Sc:  "#000",
					Txt: "Test",
				},
			},
			params: oas.ShapeUpsertParams{DiagramId: "diag-1"},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Shape {
				t.Helper()

				permSvc := mocks.NewPermissionServiceMock(t)
				permSvc.EXPECT().
					HasDiagramPermission(mock.Anything, domain.DiagramID("diag-1"), domain.UserID("user-1"), []domain.DiagramCollaboratorPermission{
						domain.DiagramCollaboratorPermissionOwner,
						domain.DiagramCollaboratorPermissionEdit,
					}).
					Return(true, nil)

				sanitizer := mocks.NewTextSanitizerMock(t)
				sanitizer.EXPECT().
					Sanitize("Test").
					Return("Test")

				diagramSvc := mocks.NewDiagramServiceMock(t)
				diagramSvc.EXPECT().
					ShapeUpsert(mock.Anything, domain.UserID("user-1"), domain.DiagramID("diag-1"), domain.DiagramShape{
						ID:     "shape-1",
						Type:   domain.DiagramShapeTypeSquare,
						Left:   10,
						Top:    20,
						Width:  30,
						Height: 40,
						ScaleX: 1,
						ScaleY: 1,
						Fill:   "#fff",
						Stroke: "#000",
						Text:   "Test",
					}).
					Return(nil)

				return &Shape{
					diagramService:    diagramSvc,
					permissionService: permSvc,
					sanitizer:         sanitizer,
					log:               zap.NewNop(),
				}
			},
			want: &oas.ShapeUpsertNoContent{},
		},
		{
			name: "Unauthorized",
			req: &oas.ShapeUpsertRequest{
				Shape: oas.DiagramShape{
					ID: oas.Ulid("shape-1"),
					Tp: oas.DiagramShapeTypeEnumS,
				},
			},
			params: oas.ShapeUpsertParams{DiagramId: "diag-1"},
			ctx:    context.Background,
			handler: func(t *testing.T) *Shape {
				t.Helper()

				return &Shape{log: zap.NewNop()}
			},
			want: &oas.ShapeUpsertUnauthorized{},
		},
		{
			name: "Forbidden - no edit permission",
			req: &oas.ShapeUpsertRequest{
				Shape: oas.DiagramShape{
					ID: oas.Ulid("shape-1"),
					Tp: oas.DiagramShapeTypeEnumS,
				},
			},
			params: oas.ShapeUpsertParams{DiagramId: "diag-1"},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Shape {
				t.Helper()

				permSvc := mocks.NewPermissionServiceMock(t)
				permSvc.EXPECT().
					HasDiagramPermission(mock.Anything, domain.DiagramID("diag-1"), domain.UserID("user-1"), []domain.DiagramCollaboratorPermission{
						domain.DiagramCollaboratorPermissionOwner,
						domain.DiagramCollaboratorPermissionEdit,
					}).
					Return(false, nil)

				return &Shape{
					permissionService: permSvc,
					log:               zap.NewNop(),
				}
			},
			want: &oas.ShapeUpsertForbidden{},
		},
		{
			name: "Permission service error",
			req: &oas.ShapeUpsertRequest{
				Shape: oas.DiagramShape{
					ID: oas.Ulid("shape-1"),
					Tp: oas.DiagramShapeTypeEnumS,
				},
			},
			params: oas.ShapeUpsertParams{DiagramId: "diag-1"},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Shape {
				t.Helper()

				permSvc := mocks.NewPermissionServiceMock(t)
				permSvc.EXPECT().
					HasDiagramPermission(mock.Anything, domain.DiagramID("diag-1"), domain.UserID("user-1"), []domain.DiagramCollaboratorPermission{
						domain.DiagramCollaboratorPermissionOwner,
						domain.DiagramCollaboratorPermissionEdit,
					}).
					Return(false, errors.New("permission error"))

				return &Shape{
					permissionService: permSvc,
					log:               zap.NewNop(),
				}
			},
			want: &oas.ShapeUpsertInternalServerError{},
		},
		{
			name: "Invalid shape type",
			req: &oas.ShapeUpsertRequest{
				Shape: oas.DiagramShape{
					ID: oas.Ulid("shape-1"),
					Tp: oas.DiagramShapeTypeEnum("INVALID"),
				},
			},
			params: oas.ShapeUpsertParams{DiagramId: "diag-1"},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Shape {
				t.Helper()

				permSvc := mocks.NewPermissionServiceMock(t)
				permSvc.EXPECT().
					HasDiagramPermission(mock.Anything, domain.DiagramID("diag-1"), domain.UserID("user-1"), []domain.DiagramCollaboratorPermission{
						domain.DiagramCollaboratorPermissionOwner,
						domain.DiagramCollaboratorPermissionEdit,
					}).
					Return(true, nil)

				sanitizer := mocks.NewTextSanitizerMock(t)
				sanitizer.EXPECT().
					Sanitize("").
					Return("")

				return &Shape{
					permissionService: permSvc,
					sanitizer:         sanitizer,
					log:               zap.NewNop(),
				}
			},
			want: &oas.ShapeUpsertInternalServerError{},
		},
		{
			name: "Service error",
			req: &oas.ShapeUpsertRequest{
				Shape: oas.DiagramShape{
					ID: oas.Ulid("shape-1"),
					Tp: oas.DiagramShapeTypeEnumS,
				},
			},
			params: oas.ShapeUpsertParams{DiagramId: "diag-1"},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Shape {
				t.Helper()

				permSvc := mocks.NewPermissionServiceMock(t)
				permSvc.EXPECT().
					HasDiagramPermission(mock.Anything, domain.DiagramID("diag-1"), domain.UserID("user-1"), []domain.DiagramCollaboratorPermission{
						domain.DiagramCollaboratorPermissionOwner,
						domain.DiagramCollaboratorPermissionEdit,
					}).
					Return(true, nil)

				sanitizer := mocks.NewTextSanitizerMock(t)
				sanitizer.EXPECT().
					Sanitize("").
					Return("")

				diagramSvc := mocks.NewDiagramServiceMock(t)
				diagramSvc.EXPECT().
					ShapeUpsert(mock.Anything, domain.UserID("user-1"), domain.DiagramID("diag-1"), mock.Anything).
					Return(errors.New("service error"))

				return &Shape{
					diagramService:    diagramSvc,
					permissionService: permSvc,
					sanitizer:         sanitizer,
					log:               zap.NewNop(),
				}
			},
			want: &oas.ShapeUpsertInternalServerError{},
		},
		{
			name: "Text is sanitized",
			req: &oas.ShapeUpsertRequest{
				Shape: oas.DiagramShape{
					ID:  oas.Ulid("shape-1"),
					Tp:  oas.DiagramShapeTypeEnumS,
					L:   10,
					T:   20,
					W:   30,
					H:   40,
					Sx:  1,
					Sy:  1,
					Fc:  "#fff",
					Sc:  "#000",
					Txt: "<script>alert('xss')</script>",
				},
			},
			params: oas.ShapeUpsertParams{DiagramId: "diag-1"},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Shape {
				t.Helper()

				permSvc := mocks.NewPermissionServiceMock(t)
				permSvc.EXPECT().
					HasDiagramPermission(mock.Anything, domain.DiagramID("diag-1"), domain.UserID("user-1"), []domain.DiagramCollaboratorPermission{
						domain.DiagramCollaboratorPermissionOwner,
						domain.DiagramCollaboratorPermissionEdit,
					}).
					Return(true, nil)

				sanitizer := mocks.NewTextSanitizerMock(t)
				sanitizer.EXPECT().
					Sanitize("<script>alert('xss')</script>").
					Return("alert('xss')")

				diagramSvc := mocks.NewDiagramServiceMock(t)
				diagramSvc.EXPECT().
					ShapeUpsert(mock.Anything, domain.UserID("user-1"), domain.DiagramID("diag-1"), domain.DiagramShape{
						ID:     "shape-1",
						Type:   domain.DiagramShapeTypeSquare,
						Left:   10,
						Top:    20,
						Width:  30,
						Height: 40,
						ScaleX: 1,
						ScaleY: 1,
						Fill:   "#fff",
						Stroke: "#000",
						Text:   "alert('xss')",
					}).
					Return(nil)

				return &Shape{
					diagramService:    diagramSvc,
					permissionService: permSvc,
					sanitizer:         sanitizer,
					log:               zap.NewNop(),
				}
			},
			want: &oas.ShapeUpsertNoContent{},
		},
		{
			name: "Success - text shape type",
			req: &oas.ShapeUpsertRequest{
				Shape: oas.DiagramShape{
					ID:  oas.Ulid("shape-1"),
					Tp:  oas.DiagramShapeTypeEnumT,
					L:   10,
					T:   20,
					W:   30,
					H:   40,
					Sx:  1,
					Sy:  1,
					Fc:  "#fff",
					Sc:  "#000",
					Txt: "Text",
				},
			},
			params: oas.ShapeUpsertParams{DiagramId: "diag-1"},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Shape {
				t.Helper()

				permSvc := mocks.NewPermissionServiceMock(t)
				permSvc.EXPECT().
					HasDiagramPermission(mock.Anything, domain.DiagramID("diag-1"), domain.UserID("user-1"), []domain.DiagramCollaboratorPermission{
						domain.DiagramCollaboratorPermissionOwner,
						domain.DiagramCollaboratorPermissionEdit,
					}).
					Return(true, nil)

				sanitizer := mocks.NewTextSanitizerMock(t)
				sanitizer.EXPECT().
					Sanitize("Text").
					Return("Text")

				diagramSvc := mocks.NewDiagramServiceMock(t)
				diagramSvc.EXPECT().
					ShapeUpsert(mock.Anything, domain.UserID("user-1"), domain.DiagramID("diag-1"), domain.DiagramShape{
						ID:     "shape-1",
						Type:   domain.DiagramShapeTypeText,
						Left:   10,
						Top:    20,
						Width:  30,
						Height: 40,
						ScaleX: 1,
						ScaleY: 1,
						Fill:   "#fff",
						Stroke: "#000",
						Text:   "Text",
					}).
					Return(nil)

				return &Shape{
					diagramService:    diagramSvc,
					permissionService: permSvc,
					sanitizer:         sanitizer,
					log:               zap.NewNop(),
				}
			},
			want: &oas.ShapeUpsertNoContent{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := tt.handler(t)
			got, err := h.ShapeUpsert(tt.ctx(), tt.req, tt.params)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestShape_ShapeDelete(t *testing.T) {
	tests := []struct {
		name    string
		params  oas.ShapeDeleteParams
		ctx     func() context.Context
		handler func(t *testing.T) *Shape
		want    oas.ShapeDeleteRes
		wantErr error
	}{
		{
			name: "Success",
			params: oas.ShapeDeleteParams{
				DiagramId:      "diag-1",
				DiagramShapeId: "shape-1",
			},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Shape {
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
					ShapeDelete(mock.Anything, domain.UserID("user-1"), domain.DiagramID("diag-1"), domain.ShapeID("shape-1")).
					Return(nil)

				return &Shape{
					diagramService:    diagramSvc,
					permissionService: permSvc,
					log:               zap.NewNop(),
				}
			},
			want: &oas.ShapeDeleteNoContent{},
		},
		{
			name: "Unauthorized",
			params: oas.ShapeDeleteParams{
				DiagramId:      "diag-1",
				DiagramShapeId: "shape-1",
			},
			ctx: context.Background,
			handler: func(t *testing.T) *Shape {
				t.Helper()

				return &Shape{log: zap.NewNop()}
			},
			want: &oas.ShapeDeleteUnauthorized{},
		},
		{
			name: "Forbidden - no edit permission",
			params: oas.ShapeDeleteParams{
				DiagramId:      "diag-1",
				DiagramShapeId: "shape-1",
			},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Shape {
				t.Helper()

				permSvc := mocks.NewPermissionServiceMock(t)
				permSvc.EXPECT().
					HasDiagramPermission(mock.Anything, domain.DiagramID("diag-1"), domain.UserID("user-1"), []domain.DiagramCollaboratorPermission{
						domain.DiagramCollaboratorPermissionOwner,
						domain.DiagramCollaboratorPermissionEdit,
					}).
					Return(false, nil)

				return &Shape{
					permissionService: permSvc,
					log:               zap.NewNop(),
				}
			},
			want: &oas.ShapeDeleteForbidden{},
		},
		{
			name: "Service error",
			params: oas.ShapeDeleteParams{
				DiagramId:      "diag-1",
				DiagramShapeId: "shape-1",
			},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Shape {
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
					ShapeDelete(mock.Anything, domain.UserID("user-1"), domain.DiagramID("diag-1"), domain.ShapeID("shape-1")).
					Return(errors.New("service error"))

				return &Shape{
					diagramService:    diagramSvc,
					permissionService: permSvc,
					log:               zap.NewNop(),
				}
			},
			want: &oas.ShapeDeleteInternalServerError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := tt.handler(t)
			got, err := h.ShapeDelete(tt.ctx(), tt.params)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestShape_ShapeBatchUpsert(t *testing.T) {
	tests := []struct {
		name    string
		req     *oas.ShapeBatchUpsertRequest
		params  oas.ShapeBatchUpsertParams
		ctx     func() context.Context
		handler func(t *testing.T) *Shape
		want    oas.ShapeBatchUpsertRes
		wantErr error
	}{
		{
			name: "Success",
			req: &oas.ShapeBatchUpsertRequest{
				Shapes: []oas.DiagramShape{
					{
						ID:  oas.Ulid("shape-1"),
						Tp:  oas.DiagramShapeTypeEnumS,
						L:   10,
						T:   20,
						W:   30,
						H:   40,
						Sx:  1,
						Sy:  1,
						Fc:  "#fff",
						Sc:  "#000",
						Txt: "<script>alert('xss')</script>",
					},
					{
						ID:  oas.Ulid("shape-2"),
						Tp:  oas.DiagramShapeTypeEnumE,
						L:   50,
						T:   60,
						W:   70,
						H:   80,
						Sx:  1,
						Sy:  1,
						Fc:  "#aaa",
						Sc:  "#111",
						Txt: "<b>bold</b> text",
					},
				},
			},
			params: oas.ShapeBatchUpsertParams{DiagramId: "diag-1"},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Shape {
				t.Helper()

				permSvc := mocks.NewPermissionServiceMock(t)
				permSvc.EXPECT().
					HasDiagramPermission(mock.Anything, domain.DiagramID("diag-1"), domain.UserID("user-1"), []domain.DiagramCollaboratorPermission{
						domain.DiagramCollaboratorPermissionOwner,
						domain.DiagramCollaboratorPermissionEdit,
					}).
					Return(true, nil)

				sanitizer := mocks.NewTextSanitizerMock(t)
				sanitizer.EXPECT().Sanitize("<script>alert('xss')</script>").Return("alert('xss')")
				sanitizer.EXPECT().Sanitize("<b>bold</b> text").Return("bold text")

				diagramSvc := mocks.NewDiagramServiceMock(t)
				diagramSvc.EXPECT().
					ShapeBatchUpsert(mock.Anything, domain.UserID("user-1"), domain.DiagramID("diag-1"), []domain.DiagramShape{
						{
							ID:     "shape-1",
							Type:   domain.DiagramShapeTypeSquare,
							Left:   10,
							Top:    20,
							Width:  30,
							Height: 40,
							ScaleX: 1,
							ScaleY: 1,
							Fill:   "#fff",
							Stroke: "#000",
							Text:   "alert('xss')",
						},
						{
							ID:     "shape-2",
							Type:   domain.DiagramShapeTypeEllipse,
							Left:   50,
							Top:    60,
							Width:  70,
							Height: 80,
							ScaleX: 1,
							ScaleY: 1,
							Fill:   "#aaa",
							Stroke: "#111",
							Text:   "bold text",
						},
					}).
					Return(nil)

				return &Shape{
					diagramService:    diagramSvc,
					permissionService: permSvc,
					sanitizer:         sanitizer,
					log:               zap.NewNop(),
				}
			},
			want: &oas.ShapeBatchUpsertNoContent{},
		},
		{
			name: "Unauthorized",
			req: &oas.ShapeBatchUpsertRequest{
				Shapes: []oas.DiagramShape{
					{
						ID: oas.Ulid("shape-1"),
						Tp: oas.DiagramShapeTypeEnumS,
					},
				},
			},
			params: oas.ShapeBatchUpsertParams{DiagramId: "diag-1"},
			ctx:    context.Background,
			handler: func(t *testing.T) *Shape {
				t.Helper()

				return &Shape{log: zap.NewNop()}
			},
			want: &oas.ShapeBatchUpsertUnauthorized{},
		},
		{
			name: "Forbidden - no edit permission",
			req: &oas.ShapeBatchUpsertRequest{
				Shapes: []oas.DiagramShape{
					{
						ID: oas.Ulid("shape-1"),
						Tp: oas.DiagramShapeTypeEnumS,
					},
				},
			},
			params: oas.ShapeBatchUpsertParams{DiagramId: "diag-1"},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Shape {
				t.Helper()

				permSvc := mocks.NewPermissionServiceMock(t)
				permSvc.EXPECT().
					HasDiagramPermission(mock.Anything, domain.DiagramID("diag-1"), domain.UserID("user-1"), []domain.DiagramCollaboratorPermission{
						domain.DiagramCollaboratorPermissionOwner,
						domain.DiagramCollaboratorPermissionEdit,
					}).
					Return(false, nil)

				return &Shape{
					permissionService: permSvc,
					log:               zap.NewNop(),
				}
			},
			want: &oas.ShapeBatchUpsertForbidden{},
		},
		{
			name: "Invalid shape type",
			req: &oas.ShapeBatchUpsertRequest{
				Shapes: []oas.DiagramShape{
					{
						ID: oas.Ulid("shape-1"),
						Tp: oas.DiagramShapeTypeEnum("INVALID"),
					},
				},
			},
			params: oas.ShapeBatchUpsertParams{DiagramId: "diag-1"},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Shape {
				t.Helper()

				permSvc := mocks.NewPermissionServiceMock(t)
				permSvc.EXPECT().
					HasDiagramPermission(mock.Anything, domain.DiagramID("diag-1"), domain.UserID("user-1"), []domain.DiagramCollaboratorPermission{
						domain.DiagramCollaboratorPermissionOwner,
						domain.DiagramCollaboratorPermissionEdit,
					}).
					Return(true, nil)

				sanitizer := mocks.NewTextSanitizerMock(t)
				sanitizer.EXPECT().Sanitize("").Return("")

				diagramSvc := mocks.NewDiagramServiceMock(t)
				diagramSvc.EXPECT().
					ShapeBatchUpsert(mock.Anything, domain.UserID("user-1"), domain.DiagramID("diag-1"), []domain.DiagramShape{}).
					Return(nil)

				return &Shape{
					diagramService:    diagramSvc,
					permissionService: permSvc,
					sanitizer:         sanitizer,
					log:               zap.NewNop(),
				}
			},
			want: &oas.ShapeBatchUpsertNoContent{},
		},
		{
			name: "Empty shapes list",
			req: &oas.ShapeBatchUpsertRequest{
				Shapes: []oas.DiagramShape{},
			},
			params: oas.ShapeBatchUpsertParams{DiagramId: "diag-1"},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Shape {
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
					ShapeBatchUpsert(mock.Anything, domain.UserID("user-1"), domain.DiagramID("diag-1"), []domain.DiagramShape{}).
					Return(nil)

				return &Shape{
					diagramService:    diagramSvc,
					permissionService: permSvc,
					log:               zap.NewNop(),
				}
			},
			want: &oas.ShapeBatchUpsertNoContent{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := tt.handler(t)
			got, err := h.ShapeBatchUpsert(tt.ctx(), tt.req, tt.params)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
