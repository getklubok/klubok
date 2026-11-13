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

func TestDiagram_DiagramCreate(t *testing.T) {
	tests := []struct {
		name    string
		req     *oas.DiagramCreateRequest
		ctx     func() context.Context
		handler func(t *testing.T) *Diagram
		want    oas.DiagramCreateRes
		wantErr error
	}{
		{
			name: "Success",
			req:  &oas.DiagramCreateRequest{Title: "Test Diagram"},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Diagram {
				t.Helper()

				sanitizer := mocks.NewTextSanitizerMock(t)
				sanitizer.EXPECT().
					Sanitize("Test Diagram").
					Return("Test Diagram")

				diagramSvc := mocks.NewDiagramServiceMock(t)
				diagramSvc.EXPECT().
					Create(mock.Anything, "Test Diagram", domain.UserID("user-1")).
					Return(domain.Diagram{
						ID:         "diag-1",
						Title:      "Test Diagram",
						Permission: domain.DiagramCollaboratorPermissionOwner,
						CreatedAt:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						UpdatedAt:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					}, nil)

				return &Diagram{
					diagramService: diagramSvc,
					sanitizer:      sanitizer,
					log:            zap.NewNop(),
				}
			},
			want: &oas.DiagramResponse{
				Diagram: oas.Diagram{
					ID:         oas.Ulid("diag-1"),
					Title:      oas.DiagramTitle("Test Diagram"),
					Permission: oas.DiagramCollaboratorPermissionEnumOWNER,
					CreatedAt:  oas.Datetime(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
					UpdatedAt:  oas.Datetime(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
				},
			},
		},
		{
			name: "Unauthorized - no claims",
			req:  &oas.DiagramCreateRequest{Title: "Test Diagram"},
			ctx:  context.Background,
			handler: func(t *testing.T) *Diagram {
				t.Helper()

				return &Diagram{log: zap.NewNop()}
			},
			want: &oas.DiagramCreateUnauthorized{},
		},
		{
			name: "Service error",
			req:  &oas.DiagramCreateRequest{Title: "Test Diagram"},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Diagram {
				t.Helper()

				sanitizer := mocks.NewTextSanitizerMock(t)
				sanitizer.EXPECT().
					Sanitize("Test Diagram").
					Return("Test Diagram")

				diagramSvc := mocks.NewDiagramServiceMock(t)
				diagramSvc.EXPECT().
					Create(mock.Anything, "Test Diagram", domain.UserID("user-1")).
					Return(domain.Diagram{}, errors.New("service error"))

				return &Diagram{
					diagramService: diagramSvc,
					sanitizer:      sanitizer,
					log:            zap.NewNop(),
				}
			},
			want: &oas.DiagramCreateInternalServerError{},
		},
		{
			name: "Title is sanitized",
			req:  &oas.DiagramCreateRequest{Title: "<script>alert('xss')</script>"},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Diagram {
				t.Helper()

				sanitizer := mocks.NewTextSanitizerMock(t)
				sanitizer.EXPECT().
					Sanitize("<script>alert('xss')</script>").
					Return("alert('xss')")

				diagramSvc := mocks.NewDiagramServiceMock(t)
				diagramSvc.EXPECT().
					Create(mock.Anything, "alert('xss')", domain.UserID("user-1")).
					Return(domain.Diagram{
						ID:         "diag-1",
						Title:      "alert('xss')",
						Permission: domain.DiagramCollaboratorPermissionOwner,
						CreatedAt:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						UpdatedAt:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					}, nil)

				return &Diagram{
					diagramService: diagramSvc,
					sanitizer:      sanitizer,
					log:            zap.NewNop(),
				}
			},
			want: &oas.DiagramResponse{
				Diagram: oas.Diagram{
					ID:         oas.Ulid("diag-1"),
					Title:      oas.DiagramTitle("alert('xss')"),
					Permission: oas.DiagramCollaboratorPermissionEnumOWNER,
					CreatedAt:  oas.Datetime(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
					UpdatedAt:  oas.Datetime(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := tt.handler(t)
			got, err := h.DiagramCreate(tt.ctx(), tt.req)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDiagram_DiagramList(t *testing.T) {
	tests := []struct {
		name    string
		ctx     func() context.Context
		handler func(t *testing.T) *Diagram
		want    oas.DiagramListRes
		wantErr error
	}{
		{
			name: "Success",
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Diagram {
				t.Helper()

				diagramSvc := mocks.NewDiagramServiceMock(t)
				diagramSvc.EXPECT().
					List(mock.Anything, domain.UserID("user-1")).
					Return([]domain.Diagram{
						{
							ID:               "diag-1",
							Title:            "Diagram 1",
							Permission:       domain.DiagramCollaboratorPermissionOwner,
							HasCollaborators: true,
							CreatedAt:        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
							UpdatedAt:        time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
						},
						{
							ID:         "diag-2",
							Title:      "Diagram 2",
							Permission: domain.DiagramCollaboratorPermissionView,
							CreatedAt:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
							UpdatedAt:  time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
						},
					}, nil)

				return &Diagram{
					diagramService: diagramSvc,
					log:            zap.NewNop(),
				}
			},
			want: &oas.DiagramListResponse{
				Diagrams: []oas.Diagram{
					{
						ID:               oas.Ulid("diag-1"),
						Title:            oas.DiagramTitle("Diagram 1"),
						Permission:       oas.DiagramCollaboratorPermissionEnumOWNER,
						HasCollaborators: true,
						CreatedAt:        oas.Datetime(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
						UpdatedAt:        oas.Datetime(time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)),
					},
					{
						ID:         oas.Ulid("diag-2"),
						Title:      oas.DiagramTitle("Diagram 2"),
						Permission: oas.DiagramCollaboratorPermissionEnumVIEW,
						CreatedAt:  oas.Datetime(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
						UpdatedAt:  oas.Datetime(time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)),
					},
				},
			},
		},
		{
			name: "Unauthorized",
			ctx:  context.Background,
			handler: func(t *testing.T) *Diagram {
				t.Helper()

				return &Diagram{log: zap.NewNop()}
			},
			want: &oas.DiagramListUnauthorized{},
		},
		{
			name: "Service error",
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Diagram {
				t.Helper()

				diagramSvc := mocks.NewDiagramServiceMock(t)
				diagramSvc.EXPECT().
					List(mock.Anything, domain.UserID("user-1")).
					Return(nil, errors.New("service error"))

				return &Diagram{
					diagramService: diagramSvc,
					log:            zap.NewNop(),
				}
			},
			want: &oas.DiagramListInternalServerError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := tt.handler(t)
			got, err := h.DiagramList(tt.ctx())
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDiagram_DiagramDelete(t *testing.T) {
	tests := []struct {
		name    string
		params  oas.DiagramDeleteParams
		ctx     func() context.Context
		handler func(t *testing.T) *Diagram
		want    oas.DiagramDeleteRes
		wantErr error
	}{
		{
			name:   "Success",
			params: oas.DiagramDeleteParams{DiagramId: "diag-1"},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Diagram {
				t.Helper()

				permSvc := mocks.NewPermissionServiceMock(t)
				permSvc.EXPECT().
					HasDiagramPermission(mock.Anything, domain.DiagramID("diag-1"), domain.UserID("user-1"), []domain.DiagramCollaboratorPermission{
						domain.DiagramCollaboratorPermissionOwner,
					}).
					Return(true, nil)

				diagramSvc := mocks.NewDiagramServiceMock(t)
				diagramSvc.EXPECT().
					Delete(mock.Anything, domain.DiagramID("diag-1")).
					Return(nil)

				return &Diagram{
					diagramService:    diagramSvc,
					permissionService: permSvc,
					log:               zap.NewNop(),
				}
			},
			want: &oas.DiagramDeleteNoContent{},
		},
		{
			name:   "Unauthorized",
			params: oas.DiagramDeleteParams{DiagramId: "diag-1"},
			ctx:    context.Background,
			handler: func(t *testing.T) *Diagram {
				t.Helper()

				return &Diagram{log: zap.NewNop()}
			},
			want: &oas.DiagramDeleteUnauthorized{},
		},
		{
			name:   "Forbidden - not owner",
			params: oas.DiagramDeleteParams{DiagramId: "diag-1"},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Diagram {
				t.Helper()

				permSvc := mocks.NewPermissionServiceMock(t)
				permSvc.EXPECT().
					HasDiagramPermission(mock.Anything, domain.DiagramID("diag-1"), domain.UserID("user-1"), []domain.DiagramCollaboratorPermission{
						domain.DiagramCollaboratorPermissionOwner,
					}).
					Return(false, nil)

				return &Diagram{
					permissionService: permSvc,
					log:               zap.NewNop(),
				}
			},
			want: &oas.DiagramDeleteForbidden{},
		},
		{
			name:   "Permission service error",
			params: oas.DiagramDeleteParams{DiagramId: "diag-1"},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Diagram {
				t.Helper()

				permSvc := mocks.NewPermissionServiceMock(t)
				permSvc.EXPECT().
					HasDiagramPermission(mock.Anything, domain.DiagramID("diag-1"), domain.UserID("user-1"), []domain.DiagramCollaboratorPermission{
						domain.DiagramCollaboratorPermissionOwner,
					}).
					Return(false, errors.New("permission error"))

				return &Diagram{
					permissionService: permSvc,
					log:               zap.NewNop(),
				}
			},
			want: &oas.DiagramDeleteInternalServerError{},
		},
		{
			name:   "Delete service error",
			params: oas.DiagramDeleteParams{DiagramId: "diag-1"},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Diagram {
				t.Helper()

				permSvc := mocks.NewPermissionServiceMock(t)
				permSvc.EXPECT().
					HasDiagramPermission(mock.Anything, domain.DiagramID("diag-1"), domain.UserID("user-1"), []domain.DiagramCollaboratorPermission{
						domain.DiagramCollaboratorPermissionOwner,
					}).
					Return(true, nil)

				diagramSvc := mocks.NewDiagramServiceMock(t)
				diagramSvc.EXPECT().
					Delete(mock.Anything, domain.DiagramID("diag-1")).
					Return(errors.New("delete error"))

				return &Diagram{
					diagramService:    diagramSvc,
					permissionService: permSvc,
					log:               zap.NewNop(),
				}
			},
			want: &oas.DiagramDeleteInternalServerError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := tt.handler(t)
			got, err := h.DiagramDelete(tt.ctx(), tt.params)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDiagram_DiagramSetTitle(t *testing.T) {
	tests := []struct {
		name    string
		req     *oas.DiagramSetTitleRequest
		params  oas.DiagramSetTitleParams
		ctx     func() context.Context
		handler func(t *testing.T) *Diagram
		want    oas.DiagramSetTitleRes
		wantErr error
	}{
		{
			name:   "Success",
			req:    &oas.DiagramSetTitleRequest{Title: "New Title"},
			params: oas.DiagramSetTitleParams{DiagramId: "diag-1"},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Diagram {
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
					Sanitize("New Title").
					Return("New Title")

				diagramSvc := mocks.NewDiagramServiceMock(t)
				diagramSvc.EXPECT().
					SetTitle(mock.Anything, domain.DiagramID("diag-1"), "New Title").
					Return(nil)

				return &Diagram{
					diagramService:    diagramSvc,
					permissionService: permSvc,
					sanitizer:         sanitizer,
					log:               zap.NewNop(),
				}
			},
			want: &oas.DiagramSetTitleNoContent{},
		},
		{
			name:   "Unauthorized",
			req:    &oas.DiagramSetTitleRequest{Title: "New Title"},
			params: oas.DiagramSetTitleParams{DiagramId: "diag-1"},
			ctx:    context.Background,
			handler: func(t *testing.T) *Diagram {
				t.Helper()

				return &Diagram{log: zap.NewNop()}
			},
			want: &oas.DiagramSetTitleUnauthorized{},
		},
		{
			name:   "Forbidden - no edit permission",
			req:    &oas.DiagramSetTitleRequest{Title: "New Title"},
			params: oas.DiagramSetTitleParams{DiagramId: "diag-1"},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Diagram {
				t.Helper()

				permSvc := mocks.NewPermissionServiceMock(t)
				permSvc.EXPECT().
					HasDiagramPermission(mock.Anything, domain.DiagramID("diag-1"), domain.UserID("user-1"), []domain.DiagramCollaboratorPermission{
						domain.DiagramCollaboratorPermissionOwner,
						domain.DiagramCollaboratorPermissionEdit,
					}).
					Return(false, nil)

				return &Diagram{
					permissionService: permSvc,
					log:               zap.NewNop(),
				}
			},
			want: &oas.DiagramSetTitleForbidden{},
		},
		{
			name:   "Title is sanitized",
			req:    &oas.DiagramSetTitleRequest{Title: "<script>alert('xss')</script>"},
			params: oas.DiagramSetTitleParams{DiagramId: "diag-1"},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Diagram {
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
					SetTitle(mock.Anything, domain.DiagramID("diag-1"), "alert('xss')").
					Return(nil)

				return &Diagram{
					diagramService:    diagramSvc,
					permissionService: permSvc,
					sanitizer:         sanitizer,
					log:               zap.NewNop(),
				}
			},
			want: &oas.DiagramSetTitleNoContent{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := tt.handler(t)
			got, err := h.DiagramSetTitle(tt.ctx(), tt.req, tt.params)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDiagram_DiagramInfo(t *testing.T) {
	tests := []struct {
		name    string
		params  oas.DiagramInfoParams
		ctx     func() context.Context
		handler func(t *testing.T) *Diagram
		want    oas.DiagramInfoRes
		wantErr error
	}{
		{
			name:   "Success without data",
			params: oas.DiagramInfoParams{DiagramId: "diag-1"},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Diagram {
				t.Helper()

				permSvc := mocks.NewPermissionServiceMock(t)
				permSvc.EXPECT().
					HasDiagramPermission(mock.Anything, domain.DiagramID("diag-1"), domain.UserID("user-1"), []domain.DiagramCollaboratorPermission{
						domain.DiagramCollaboratorPermissionOwner,
						domain.DiagramCollaboratorPermissionView,
						domain.DiagramCollaboratorPermissionEdit,
					}).
					Return(true, nil)

				diagramSvc := mocks.NewDiagramServiceMock(t)
				diagramSvc.EXPECT().
					InfoByUser(mock.Anything, domain.DiagramID("diag-1"), domain.UserID("user-1")).
					Return(domain.DiagramInfo{
						ID:         "diag-1",
						Title:      "Test Diagram",
						Permission: domain.DiagramCollaboratorPermissionOwner,
					}, nil)

				return &Diagram{
					diagramService:    diagramSvc,
					permissionService: permSvc,
					log:               zap.NewNop(),
				}
			},
			want: &oas.DiagramInfoResponse{
				DiagramInfo: oas.DiagramInfo{
					ID:         oas.Ulid("diag-1"),
					Title:      oas.DiagramTitle("Test Diagram"),
					Permission: oas.DiagramCollaboratorPermissionEnumOWNER,
				},
			},
		},
		{
			name:   "Success with data",
			params: oas.DiagramInfoParams{DiagramId: "diag-1"},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Diagram {
				t.Helper()

				permSvc := mocks.NewPermissionServiceMock(t)
				permSvc.EXPECT().
					HasDiagramPermission(mock.Anything, domain.DiagramID("diag-1"), domain.UserID("user-1"), []domain.DiagramCollaboratorPermission{
						domain.DiagramCollaboratorPermissionOwner,
						domain.DiagramCollaboratorPermissionView,
						domain.DiagramCollaboratorPermissionEdit,
					}).
					Return(true, nil)

				diagramSvc := mocks.NewDiagramServiceMock(t)
				diagramSvc.EXPECT().
					InfoByUser(mock.Anything, domain.DiagramID("diag-1"), domain.UserID("user-1")).
					Return(domain.DiagramInfo{
						ID:         "diag-1",
						Title:      "Test Diagram",
						Permission: domain.DiagramCollaboratorPermissionOwner,
						Shapes: []domain.DiagramShape{
							{
								ID:     "shape-1",
								Type:   domain.DiagramShapeTypeSquare,
								Left:   10,
								Top:    20,
								Width:  30,
								Height: 40,
								ScaleX: 1,
								ScaleY: 1,
								Fill:   "red",
								Stroke: "blue",
								Text:   "Hello",
								Order:  "a0",
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
								Fill:   "green",
								Stroke: "yellow",
								Text:   "World",
								Order:  "a1",
							},
							{
								ID:     "shape-3",
								Type:   domain.DiagramShapeTypeDiamond,
								Left:   90,
								Top:    90,
								ScaleX: 1,
								ScaleY: 1,
								Fill:   "blue",
								Stroke: "red",
								Text:   "!",
								Order:  "a2",
							},
							{
								ID:     "shape-4",
								Type:   domain.DiagramShapeTypeText,
								Left:   130,
								Top:    140,
								Width:  150,
								Height: 160,
								ScaleX: 1,
								ScaleY: 1,
								Fill:   "white",
								Stroke: "black",
								Text:   "Label",
								Order:  "a3",
							},
						},
						Connections: []domain.DiagramConnection{
							{
								FromShapeID: "shape-1",
								ToShapeID:   "shape-2",
							},
						},
						ShapeCommentStats: []domain.DiagramShapeCommentStat{
							{
								ShapeID:      "shape-1",
								CommentCount: 2,
							},
						},
					}, nil)

				return &Diagram{
					diagramService:    diagramSvc,
					permissionService: permSvc,
					log:               zap.NewNop(),
				}
			},
			want: &oas.DiagramInfoResponse{
				DiagramInfo: oas.DiagramInfo{
					ID:         oas.Ulid("diag-1"),
					Title:      oas.DiagramTitle("Test Diagram"),
					Permission: oas.DiagramCollaboratorPermissionEnumOWNER,
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
							Fc:  "red",
							Sc:  "blue",
							Txt: "Hello",
							O:   "a0",
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
							Fc:  "green",
							Sc:  "yellow",
							Txt: "World",
							O:   "a1",
						},
						{
							ID:  oas.Ulid("shape-3"),
							Tp:  oas.DiagramShapeTypeEnumD,
							L:   90,
							T:   90,
							W:   100, // Backward compatibility
							H:   100, // Backward compatibility
							Sx:  1,
							Sy:  1,
							Fc:  "blue",
							Sc:  "red",
							Txt: "!",
							O:   "a2",
						},
						{
							ID:  oas.Ulid("shape-4"),
							Tp:  oas.DiagramShapeTypeEnumT,
							L:   130,
							T:   140,
							W:   150,
							H:   160,
							Sx:  1,
							Sy:  1,
							Fc:  "white",
							Sc:  "black",
							Txt: "Label",
							O:   "a3",
						},
					},
					Connections: []oas.DiagramConnection{
						{
							Fid: oas.Ulid("shape-1"),
							Tid: oas.Ulid("shape-2"),
						},
					},
					ShapeCommentStats: []oas.DiagramShapeCommentStats{
						{
							ShapeId:      oas.Ulid("shape-1"),
							CommentCount: 2,
						},
					},
				},
			},
		},
		{
			name:   "Unauthorized",
			params: oas.DiagramInfoParams{DiagramId: "diag-1"},
			ctx:    context.Background,
			handler: func(t *testing.T) *Diagram {
				t.Helper()

				return &Diagram{log: zap.NewNop()}
			},
			want: &oas.DiagramInfoUnauthorized{},
		},
		{
			name:   "Forbidden",
			params: oas.DiagramInfoParams{DiagramId: "diag-1"},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Diagram {
				t.Helper()

				permSvc := mocks.NewPermissionServiceMock(t)
				permSvc.EXPECT().
					HasDiagramPermission(mock.Anything, domain.DiagramID("diag-1"), domain.UserID("user-1"), []domain.DiagramCollaboratorPermission{
						domain.DiagramCollaboratorPermissionOwner,
						domain.DiagramCollaboratorPermissionView,
						domain.DiagramCollaboratorPermissionEdit,
					}).
					Return(false, nil)

				return &Diagram{
					permissionService: permSvc,
					log:               zap.NewNop(),
				}
			},
			want: &oas.DiagramInfoForbidden{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := tt.handler(t)
			got, err := h.DiagramInfo(tt.ctx(), tt.params)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDiagram_DiagramStat(t *testing.T) {
	tests := []struct {
		name    string
		params  oas.DiagramStatParams
		ctx     func() context.Context
		handler func(t *testing.T) *Diagram
		want    oas.DiagramStatRes
		wantErr error
	}{
		{
			name:   "Success",
			params: oas.DiagramStatParams{DiagramId: "diag-1"},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Diagram {
				t.Helper()

				permSvc := mocks.NewPermissionServiceMock(t)
				permSvc.EXPECT().
					HasDiagramPermission(mock.Anything, domain.DiagramID("diag-1"), domain.UserID("user-1"), []domain.DiagramCollaboratorPermission{
						domain.DiagramCollaboratorPermissionOwner,
					}).
					Return(true, nil)

				wsManager := mocks.NewWebSocketManagerMock(t)
				wsManager.EXPECT().
					GetDiagramStats(domain.DiagramID("diag-1")).
					Return(domain.WebSocketStatDiagram{
						Connections: 5,
						CursorPositions: []domain.WebSocketCursorPosition{
							{
								UserID:   "user-1",
								Username: "John Doe",
								X:        10,
								Y:        20,
							},
							{
								UserID:   "user-2",
								Username: "Jane Doe",
								X:        30,
								Y:        40,
							},
						},
					})

				return &Diagram{
					permissionService: permSvc,
					wsManager:         wsManager,
					log:               zap.NewNop(),
				}
			},
			want: &oas.DiagramStatResponse{
				WebsocketConnections: 5,
				CursorPositions: []oas.CursorPosition{
					{
						UserId:   oas.Ulid("user-1"),
						UserName: oas.Username("John Doe"),
						X:        10,
						Y:        20,
					},
					{
						UserId:   oas.Ulid("user-2"),
						UserName: oas.Username("Jane Doe"),
						X:        30,
						Y:        40,
					},
				},
			},
		},
		{
			name:   "Unauthorized",
			params: oas.DiagramStatParams{DiagramId: "diag-1"},
			ctx:    context.Background,
			handler: func(t *testing.T) *Diagram {
				t.Helper()

				return &Diagram{log: zap.NewNop()}
			},
			want: &oas.DiagramStatUnauthorized{},
		},
		{
			name:   "Forbidden",
			params: oas.DiagramStatParams{DiagramId: "diag-1"},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Diagram {
				t.Helper()

				permSvc := mocks.NewPermissionServiceMock(t)
				permSvc.EXPECT().
					HasDiagramPermission(mock.Anything, domain.DiagramID("diag-1"), domain.UserID("user-1"), []domain.DiagramCollaboratorPermission{
						domain.DiagramCollaboratorPermissionOwner,
					}).
					Return(false, nil)

				return &Diagram{
					permissionService: permSvc,
					log:               zap.NewNop(),
				}
			},
			want: &oas.DiagramStatForbidden{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := tt.handler(t)
			got, err := h.DiagramStat(tt.ctx(), tt.params)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDiagram_DiagramOnlineUsers(t *testing.T) {
	tests := []struct {
		name    string
		params  oas.DiagramOnlineUsersParams
		ctx     func() context.Context
		handler func(t *testing.T) *Diagram
		want    oas.DiagramOnlineUsersRes
		wantErr error
	}{
		{
			name:   "Success",
			params: oas.DiagramOnlineUsersParams{DiagramId: "diag-1"},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Diagram {
				t.Helper()

				permSvc := mocks.NewPermissionServiceMock(t)
				permSvc.EXPECT().
					HasDiagramPermission(mock.Anything, domain.DiagramID("diag-1"), domain.UserID("user-1"), []domain.DiagramCollaboratorPermission{
						domain.DiagramCollaboratorPermissionOwner,
						domain.DiagramCollaboratorPermissionView,
						domain.DiagramCollaboratorPermissionEdit,
					}).
					Return(true, nil)

				diagramSvc := mocks.NewDiagramServiceMock(t)
				diagramSvc.EXPECT().
					OnlineUsers(mock.Anything, domain.DiagramID("diag-1")).
					Return([]domain.WebSocketUser{
						{
							ID:       "user-1",
							Username: "john",
						},
					})

				return &Diagram{
					diagramService:    diagramSvc,
					permissionService: permSvc,
					log:               zap.NewNop(),
				}
			},
			want: &oas.DiagramOnlineUsersResponse{
				OnlineUsers: []oas.OnlineUser{
					{
						ID:       oas.Ulid("user-1"),
						UserName: oas.Username("john"),
					},
				},
			},
		},
		{
			name:   "Unauthorized",
			params: oas.DiagramOnlineUsersParams{DiagramId: "diag-1"},
			ctx:    context.Background,
			handler: func(t *testing.T) *Diagram {
				t.Helper()

				return &Diagram{log: zap.NewNop()}
			},
			want: &oas.DiagramOnlineUsersUnauthorized{},
		},
		{
			name:   "Forbidden",
			params: oas.DiagramOnlineUsersParams{DiagramId: "diag-1"},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Diagram {
				t.Helper()

				permSvc := mocks.NewPermissionServiceMock(t)
				permSvc.EXPECT().
					HasDiagramPermission(mock.Anything, domain.DiagramID("diag-1"), domain.UserID("user-1"), []domain.DiagramCollaboratorPermission{
						domain.DiagramCollaboratorPermissionOwner,
						domain.DiagramCollaboratorPermissionView,
						domain.DiagramCollaboratorPermissionEdit,
					}).
					Return(false, nil)

				return &Diagram{
					permissionService: permSvc,
					log:               zap.NewNop(),
				}
			},
			want: &oas.DiagramOnlineUsersForbidden{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := tt.handler(t)
			got, err := h.DiagramOnlineUsers(tt.ctx(), tt.params)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDiagram_DiagramClone(t *testing.T) {
	tests := []struct {
		name    string
		params  oas.DiagramCloneParams
		ctx     func() context.Context
		handler func(t *testing.T) *Diagram
		want    oas.DiagramCloneRes
		wantErr error
	}{
		{
			name:   "Success",
			params: oas.DiagramCloneParams{DiagramId: "src-diagram-1"},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Diagram {
				t.Helper()

				permSvc := mocks.NewPermissionServiceMock(t)
				permSvc.EXPECT().
					HasDiagramPermission(mock.Anything, domain.DiagramID("src-diagram-1"), domain.UserID("user-1"), []domain.DiagramCollaboratorPermission{
						domain.DiagramCollaboratorPermissionOwner,
						domain.DiagramCollaboratorPermissionEdit,
					}).
					Return(true, nil)

				diagramSvc := mocks.NewDiagramServiceMock(t)
				diagramSvc.EXPECT().
					Clone(mock.Anything, domain.DiagramID("src-diagram-1"), domain.UserID("user-1")).
					Return(domain.Diagram{
						ID:               "new-diagram-1",
						Title:            "Original (copy)",
						Permission:       domain.DiagramCollaboratorPermissionOwner,
						HasCollaborators: true,
						CreatedAt:        time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
						UpdatedAt:        time.Date(2024, 2, 2, 0, 0, 0, 0, time.UTC),
					}, nil)

				return &Diagram{
					diagramService:    diagramSvc,
					permissionService: permSvc,
					log:               zap.NewNop(),
				}
			},
			want: &oas.DiagramResponse{
				Diagram: oas.Diagram{
					ID:               oas.Ulid("new-diagram-1"),
					Title:            oas.DiagramTitle("Original (copy)"),
					Permission:       oas.DiagramCollaboratorPermissionEnumOWNER,
					HasCollaborators: true,
					CreatedAt:        oas.Datetime(time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)),
					UpdatedAt:        oas.Datetime(time.Date(2024, 2, 2, 0, 0, 0, 0, time.UTC)),
				},
			},
		},
		{
			name:   "Unauthorized",
			params: oas.DiagramCloneParams{DiagramId: "src-diagram-1"},
			ctx:    context.Background,
			handler: func(t *testing.T) *Diagram {
				t.Helper()

				return &Diagram{log: zap.NewNop()}
			},
			want: &oas.DiagramCloneUnauthorized{},
		},
		{
			name:   "Forbidden - no edit permission",
			params: oas.DiagramCloneParams{DiagramId: "src-diagram-1"},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Diagram {
				t.Helper()

				permSvc := mocks.NewPermissionServiceMock(t)
				permSvc.EXPECT().
					HasDiagramPermission(mock.Anything, domain.DiagramID("src-diagram-1"), domain.UserID("user-1"), []domain.DiagramCollaboratorPermission{
						domain.DiagramCollaboratorPermissionOwner,
						domain.DiagramCollaboratorPermissionEdit,
					}).
					Return(false, nil)

				return &Diagram{
					permissionService: permSvc,
					log:               zap.NewNop(),
				}
			},
			want: &oas.DiagramCloneForbidden{},
		},
		{
			name:   "Permission service error",
			params: oas.DiagramCloneParams{DiagramId: "src-diagram-1"},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Diagram {
				t.Helper()

				permSvc := mocks.NewPermissionServiceMock(t)
				permSvc.EXPECT().
					HasDiagramPermission(mock.Anything, domain.DiagramID("src-diagram-1"), domain.UserID("user-1"), []domain.DiagramCollaboratorPermission{
						domain.DiagramCollaboratorPermissionOwner,
						domain.DiagramCollaboratorPermissionEdit,
					}).
					Return(false, errors.New("permission error"))

				return &Diagram{
					permissionService: permSvc,
					log:               zap.NewNop(),
				}
			},
			want: &oas.DiagramCloneInternalServerError{},
		},
		{
			name:   "Clone service error",
			params: oas.DiagramCloneParams{DiagramId: "src-diagram-1"},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Diagram {
				t.Helper()

				permSvc := mocks.NewPermissionServiceMock(t)
				permSvc.EXPECT().
					HasDiagramPermission(mock.Anything, domain.DiagramID("src-diagram-1"), domain.UserID("user-1"), []domain.DiagramCollaboratorPermission{
						domain.DiagramCollaboratorPermissionOwner,
						domain.DiagramCollaboratorPermissionEdit,
					}).
					Return(true, nil)

				diagramSvc := mocks.NewDiagramServiceMock(t)
				diagramSvc.EXPECT().
					Clone(mock.Anything, domain.DiagramID("src-diagram-1"), domain.UserID("user-1")).
					Return(domain.Diagram{}, errors.New("clone error"))

				return &Diagram{
					diagramService:    diagramSvc,
					permissionService: permSvc,
					log:               zap.NewNop(),
				}
			},
			want: &oas.DiagramCloneInternalServerError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := tt.handler(t)
			got, err := h.DiagramClone(tt.ctx(), tt.params)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDiagram_DiagramImport(t *testing.T) {
	tests := []struct {
		name    string
		req     *oas.DiagramImportRequest
		ctx     func() context.Context
		handler func(t *testing.T) *Diagram
		want    oas.DiagramImportRes
		wantErr error
	}{
		{
			name: "Success",
			req: &oas.DiagramImportRequest{
				Title: "Original",
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
						O:   "a0",
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
						O:   "a1",
					},
					{
						ID:  oas.Ulid("shape-3"),
						Tp:  oas.DiagramShapeTypeEnumD,
						L:   90,
						T:   90,
						W:   90,
						H:   90,
						Sx:  1,
						Sy:  1,
						Fc:  "#aaa",
						Sc:  "#111",
						Txt: "text",
						O:   "a2",
					},
				},
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
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Diagram {
				t.Helper()

				diagramSvc := mocks.NewDiagramServiceMock(t)
				diagramSvc.EXPECT().
					Import(mock.Anything, domain.UserID("user-1"), domain.DiagramInfo{
						Title:      "Original",
						Permission: domain.DiagramCollaboratorPermissionOwner,
						Shapes: []domain.DiagramShape{
							{
								ID:     "shape-1",
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
								ID:     "shape-2",
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
								ID:     "shape-3",
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
								FromShapeID: "shape-1",
								ToShapeID:   "shape-2",
							},
							{
								FromShapeID: "shape-2",
								ToShapeID:   "shape-3",
							},
						},
					}).
					Return(domain.Diagram{
						ID:         "new-diagram-1",
						Title:      "Original (imported 2026-05-08 12:34)",
						Permission: domain.DiagramCollaboratorPermissionOwner,
						CreatedAt:  time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
						UpdatedAt:  time.Date(2024, 2, 2, 0, 0, 0, 0, time.UTC),
					}, nil)

				sanitizer := mocks.NewTextSanitizerMock(t)
				sanitizer.EXPECT().Sanitize("Original").Return("Original")
				sanitizer.EXPECT().Sanitize("<script>alert('xss')</script>").Return("text")
				sanitizer.EXPECT().Sanitize("<b>bold</b> text").Return("text")
				sanitizer.EXPECT().Sanitize("text").Return("text")

				return &Diagram{
					diagramService: diagramSvc,
					sanitizer:      sanitizer,
					log:            zap.NewNop(),
				}
			},
			want: &oas.DiagramResponse{
				Diagram: oas.Diagram{
					ID:         oas.Ulid("new-diagram-1"),
					Title:      oas.DiagramTitle("Original (imported 2026-05-08 12:34)"),
					Permission: oas.DiagramCollaboratorPermissionEnumOWNER,
					CreatedAt:  oas.Datetime(time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)),
					UpdatedAt:  oas.Datetime(time.Date(2024, 2, 2, 0, 0, 0, 0, time.UTC)),
				},
			},
		},
		{
			name: "Unauthorized",
			req:  &oas.DiagramImportRequest{},
			ctx:  context.Background,
			handler: func(t *testing.T) *Diagram {
				t.Helper()

				return &Diagram{log: zap.NewNop()}
			},
			want: &oas.DiagramImportUnauthorized{},
		},
		{
			name: "Import service error",
			req:  &oas.DiagramImportRequest{},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Diagram {
				t.Helper()

				diagramSvc := mocks.NewDiagramServiceMock(t)
				diagramSvc.EXPECT().
					Import(mock.Anything, domain.UserID("user-1"), domain.DiagramInfo{Permission: domain.DiagramCollaboratorPermissionOwner}).
					Return(domain.Diagram{}, errors.New("import error"))

				sanitizer := mocks.NewTextSanitizerMock(t)
				sanitizer.EXPECT().Sanitize("").Return("")

				return &Diagram{
					diagramService: diagramSvc,
					sanitizer:      sanitizer,
					log:            zap.NewNop(),
				}
			},
			want: &oas.DiagramImportInternalServerError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := tt.handler(t)
			got, err := h.DiagramImport(tt.ctx(), tt.req)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
