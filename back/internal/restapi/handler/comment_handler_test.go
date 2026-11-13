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

func TestComment_CommentCreate(t *testing.T) {
	tests := []struct {
		name    string
		req     *oas.CommentCreateRequest
		params  oas.CommentCreateParams
		ctx     func() context.Context
		handler func(t *testing.T) *Comment
		want    oas.CommentCreateRes
		wantErr error
	}{
		{
			name: "Success",
			req: &oas.CommentCreateRequest{
				Text: "Test comment",
			},
			params: oas.CommentCreateParams{
				DiagramId:      "diag-1",
				DiagramShapeId: "shape-1",
			},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Comment {
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
					CommentCreate(mock.Anything, domain.UserID("user-1"), domain.DiagramComment{
						DiagramID:      "diag-1",
						DiagramShapeID: "shape-1",
						UserID:         "user-1",
						Text:           "Test comment",
					}).
					Return(domain.DiagramComment{
						ID:        "comment-1",
						UserName:  "user1",
						Text:      "Test comment",
						CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					}, nil)

				sanitizer := mocks.NewTextSanitizerMock(t)
				sanitizer.EXPECT().
					Sanitize("Test comment").
					Return("Test comment")

				return &Comment{
					diagramService:    diagramSvc,
					permissionService: permSvc,
					sanitizer:         sanitizer,
					log:               zap.NewNop(),
				}
			},
			want: &oas.CommentCreateResponse{
				Comment: oas.DiagramComment{
					ID:        oas.Ulid("comment-1"),
					UserName:  oas.Username("user1"),
					Text:      "Test comment",
					CreatedAt: oas.Datetime(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
				},
			},
		},
		{
			name: "Success with HTML sanitization",
			req: &oas.CommentCreateRequest{
				Text: "<script>alert('xss')</script>Hello",
			},
			params: oas.CommentCreateParams{
				DiagramId:      "diag-1",
				DiagramShapeId: "shape-1",
			},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Comment {
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
					CommentCreate(mock.Anything, domain.UserID("user-1"), domain.DiagramComment{
						DiagramID:      "diag-1",
						DiagramShapeID: "shape-1",
						UserID:         "user-1",
						Text:           "Hello",
					}).
					Return(domain.DiagramComment{
						ID:        "comment-1",
						UserName:  "user1",
						Text:      "Hello",
						CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					}, nil)

				sanitizer := mocks.NewTextSanitizerMock(t)
				sanitizer.EXPECT().
					Sanitize("<script>alert('xss')</script>Hello").
					Return("Hello")

				return &Comment{
					diagramService:    diagramSvc,
					permissionService: permSvc,
					sanitizer:         sanitizer,
					log:               zap.NewNop(),
				}
			},
			want: &oas.CommentCreateResponse{
				Comment: oas.DiagramComment{
					ID:        oas.Ulid("comment-1"),
					UserName:  oas.Username("user1"),
					Text:      "Hello",
					CreatedAt: oas.Datetime(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
				},
			},
		},
		{
			name: "Unauthorized",
			req: &oas.CommentCreateRequest{
				Text: "Test comment",
			},
			params: oas.CommentCreateParams{
				DiagramId:      "diag-1",
				DiagramShapeId: "shape-1",
			},
			ctx: context.Background,
			handler: func(t *testing.T) *Comment {
				t.Helper()

				return &Comment{
					log: zap.NewNop(),
				}
			},
			want: &oas.CommentCreateUnauthorized{},
		},
		{
			name: "Forbidden - no access",
			req: &oas.CommentCreateRequest{
				Text: "Test comment",
			},
			params: oas.CommentCreateParams{
				DiagramId:      "diag-1",
				DiagramShapeId: "shape-1",
			},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Comment {
				t.Helper()

				permSvc := mocks.NewPermissionServiceMock(t)
				permSvc.EXPECT().
					HasDiagramPermission(mock.Anything, domain.DiagramID("diag-1"), domain.UserID("user-1"), []domain.DiagramCollaboratorPermission{
						domain.DiagramCollaboratorPermissionOwner,
						domain.DiagramCollaboratorPermissionView,
						domain.DiagramCollaboratorPermissionEdit,
					}).
					Return(false, nil)

				return &Comment{
					permissionService: permSvc,
					log:               zap.NewNop(),
				}
			},
			want: &oas.CommentCreateForbidden{},
		},
		{
			name: "Service error",
			req: &oas.CommentCreateRequest{
				Text: "Test comment",
			},
			params: oas.CommentCreateParams{
				DiagramId:      "diag-1",
				DiagramShapeId: "shape-1",
			},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Comment {
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
					CommentCreate(mock.Anything, domain.UserID("user-1"), mock.Anything).
					Return(domain.DiagramComment{}, errors.New("service error"))

				sanitizer := mocks.NewTextSanitizerMock(t)
				sanitizer.EXPECT().
					Sanitize("Test comment").
					Return("Test comment")

				return &Comment{
					diagramService:    diagramSvc,
					permissionService: permSvc,
					sanitizer:         sanitizer,
					log:               zap.NewNop(),
				}
			},
			want: &oas.CommentCreateInternalServerError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := tt.handler(t)
			got, err := h.CommentCreate(tt.ctx(), tt.req, tt.params)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestComment_CommentList(t *testing.T) {
	tests := []struct {
		name    string
		params  oas.CommentListParams
		ctx     func() context.Context
		handler func(t *testing.T) *Comment
		want    oas.CommentListRes
		wantErr error
	}{
		{
			name: "Success",
			params: oas.CommentListParams{
				DiagramId:      "diag-1",
				DiagramShapeId: "shape-1",
			},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Comment {
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
					CommentList(mock.Anything, domain.DiagramID("diag-1"), domain.ShapeID("shape-1")).
					Return([]domain.DiagramComment{
						{
							ID:        "comment-1",
							UserName:  "user1",
							Text:      "Comment 1",
							CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						},
					}, nil)

				return &Comment{
					diagramService:    diagramSvc,
					permissionService: permSvc,
					log:               zap.NewNop(),
				}
			},
			want: &oas.CommentListResponse{
				Comments: []oas.DiagramComment{
					{
						ID:        oas.Ulid("comment-1"),
						UserName:  oas.Username("user1"),
						Text:      "Comment 1",
						CreatedAt: oas.Datetime(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
					},
				},
			},
		},
		{
			name: "Unauthorized",
			params: oas.CommentListParams{
				DiagramId:      "diag-1",
				DiagramShapeId: "shape-1",
			},
			ctx: context.Background,
			handler: func(t *testing.T) *Comment {
				t.Helper()

				return &Comment{
					log: zap.NewNop(),
				}
			},
			want: &oas.CommentListUnauthorized{},
		},
		{
			name: "Forbidden - no access",
			params: oas.CommentListParams{
				DiagramId:      "diag-1",
				DiagramShapeId: "shape-1",
			},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Comment {
				t.Helper()

				permSvc := mocks.NewPermissionServiceMock(t)
				permSvc.EXPECT().
					HasDiagramPermission(mock.Anything, domain.DiagramID("diag-1"), domain.UserID("user-1"), []domain.DiagramCollaboratorPermission{
						domain.DiagramCollaboratorPermissionOwner,
						domain.DiagramCollaboratorPermissionView,
						domain.DiagramCollaboratorPermissionEdit,
					}).
					Return(false, nil)

				return &Comment{
					permissionService: permSvc,
					log:               zap.NewNop(),
				}
			},
			want: &oas.CommentListForbidden{},
		},
		{
			name: "Service error",
			params: oas.CommentListParams{
				DiagramId:      "diag-1",
				DiagramShapeId: "shape-1",
			},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Comment {
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
					CommentList(mock.Anything, domain.DiagramID("diag-1"), domain.ShapeID("shape-1")).
					Return(nil, errors.New("service error"))

				return &Comment{
					diagramService:    diagramSvc,
					permissionService: permSvc,
					log:               zap.NewNop(),
				}
			},
			want: &oas.CommentListInternalServerError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := tt.handler(t)
			got, err := h.CommentList(tt.ctx(), tt.params)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
