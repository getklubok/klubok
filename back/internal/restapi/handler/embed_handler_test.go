package handler

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"klubok/internal/domain"
	"klubok/internal/restapi/handler/mocks"
	"klubok/internal/restapi/oas"
)

func TestEmbed_EmbedCreate(t *testing.T) {
	tests := []struct {
		name    string
		params  oas.EmbedCreateParams
		ctx     func() context.Context
		handler func(t *testing.T) *Embed
		want    oas.EmbedCreateRes
	}{
		{
			name:   "Success",
			params: oas.EmbedCreateParams{DiagramId: "diag-1"},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Embed {
				t.Helper()

				permSvc := mocks.NewPermissionServiceMock(t)
				permSvc.EXPECT().
					HasDiagramPermission(mock.Anything, domain.DiagramID("diag-1"), domain.UserID("user-1"), []domain.DiagramCollaboratorPermission{
						domain.DiagramCollaboratorPermissionOwner,
					}).
					Return(true, nil)

				embedSvc := mocks.NewEmbedServiceMock(t)
				embedSvc.EXPECT().
					Create(mock.Anything, domain.DiagramID("diag-1"), domain.UserID("user-1")).
					Return(domain.DiagramEmbedInfo{
						ID:        "embed-1",
						DiagramID: "diag-1",
						EmbedCode: "<script src='...'></script>",
						IsActive:  true,
						CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						UpdatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					}, nil)

				return &Embed{
					embedService:      embedSvc,
					permissionService: permSvc,
					log:               zap.NewNop(),
				}
			},
			want: &oas.EmbedInfoResponse{
				Embed: oas.EmbedInfo{
					ID:        oas.Ulid("embed-1"),
					EmbedCode: "<script src='...'></script>",
					IsActive:  true,
					CreatedAt: oas.Datetime(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
					UpdatedAt: oas.Datetime(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
				},
			},
		},
		{
			name:   "Unauthorized",
			params: oas.EmbedCreateParams{DiagramId: "diag-1"},
			ctx:    context.Background,
			handler: func(t *testing.T) *Embed {
				t.Helper()
				return &Embed{log: zap.NewNop()}
			},
			want: &oas.EmbedCreateUnauthorized{},
		},
		{
			name:   "Forbidden - no edit permission",
			params: oas.EmbedCreateParams{DiagramId: "diag-1"},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Embed {
				t.Helper()

				permSvc := mocks.NewPermissionServiceMock(t)
				permSvc.EXPECT().
					HasDiagramPermission(mock.Anything, domain.DiagramID("diag-1"), domain.UserID("user-1"), []domain.DiagramCollaboratorPermission{
						domain.DiagramCollaboratorPermissionOwner,
					}).
					Return(false, nil)

				return &Embed{
					permissionService: permSvc,
					log:               zap.NewNop(),
				}
			},
			want: &oas.EmbedCreateForbidden{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := tt.handler(t)
			got, err := h.EmbedCreate(tt.ctx(), tt.params)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestEmbed_EmbedList(t *testing.T) {
	tests := []struct {
		name    string
		params  oas.EmbedListParams
		ctx     func() context.Context
		handler func(t *testing.T) *Embed
		want    oas.EmbedListRes
	}{
		{
			name:   "Success",
			params: oas.EmbedListParams{DiagramId: "diag-1"},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Embed {
				t.Helper()

				permSvc := mocks.NewPermissionServiceMock(t)
				permSvc.EXPECT().
					HasDiagramPermission(mock.Anything, domain.DiagramID("diag-1"), domain.UserID("user-1"), []domain.DiagramCollaboratorPermission{
						domain.DiagramCollaboratorPermissionOwner,
					}).
					Return(true, nil)

				embedSvc := mocks.NewEmbedServiceMock(t)
				embedSvc.EXPECT().
					List(mock.Anything, domain.DiagramID("diag-1")).
					Return([]domain.DiagramEmbedInfo{
						{
							ID:        "embed-1",
							DiagramID: "diag-1",
							IsActive:  true,
							CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
							UpdatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						},
						{
							ID:        "embed-2",
							DiagramID: "diag-1",
							IsActive:  false,
							CreatedAt: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
							UpdatedAt: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
						},
					}, nil)

				return &Embed{
					embedService:      embedSvc,
					permissionService: permSvc,
					log:               zap.NewNop(),
				}
			},
			want: &oas.EmbedListResponse{
				Embeds: []oas.EmbedInfo{
					{
						ID:        oas.Ulid("embed-1"),
						IsActive:  true,
						CreatedAt: oas.Datetime(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
						UpdatedAt: oas.Datetime(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
					},
					{
						ID:        oas.Ulid("embed-2"),
						IsActive:  false,
						CreatedAt: oas.Datetime(time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)),
						UpdatedAt: oas.Datetime(time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)),
					},
				},
			},
		},
		{
			name:   "Success - empty list",
			params: oas.EmbedListParams{DiagramId: "diag-1"},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Embed {
				t.Helper()

				permSvc := mocks.NewPermissionServiceMock(t)
				permSvc.EXPECT().
					HasDiagramPermission(mock.Anything, domain.DiagramID("diag-1"), domain.UserID("user-1"), []domain.DiagramCollaboratorPermission{
						domain.DiagramCollaboratorPermissionOwner,
					}).
					Return(true, nil)

				embedSvc := mocks.NewEmbedServiceMock(t)
				embedSvc.EXPECT().
					List(mock.Anything, domain.DiagramID("diag-1")).
					Return([]domain.DiagramEmbedInfo{}, nil)

				return &Embed{
					embedService:      embedSvc,
					permissionService: permSvc,
					log:               zap.NewNop(),
				}
			},
			want: &oas.EmbedListResponse{
				Embeds: []oas.EmbedInfo{},
			},
		},
		{
			name:   "Unauthorized",
			params: oas.EmbedListParams{DiagramId: "diag-1"},
			ctx:    context.Background,
			handler: func(t *testing.T) *Embed {
				t.Helper()
				return &Embed{log: zap.NewNop()}
			},
			want: &oas.EmbedListUnauthorized{},
		},
		{
			name:   "Forbidden - no access",
			params: oas.EmbedListParams{DiagramId: "diag-1"},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Embed {
				t.Helper()

				permSvc := mocks.NewPermissionServiceMock(t)
				permSvc.EXPECT().
					HasDiagramPermission(mock.Anything, domain.DiagramID("diag-1"), domain.UserID("user-1"), []domain.DiagramCollaboratorPermission{
						domain.DiagramCollaboratorPermissionOwner,
					}).
					Return(false, nil)

				return &Embed{
					permissionService: permSvc,
					log:               zap.NewNop(),
				}
			},
			want: &oas.EmbedListForbidden{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := tt.handler(t)
			got, err := h.EmbedList(tt.ctx(), tt.params)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestEmbed_EmbedGet(t *testing.T) {
	tests := []struct {
		name    string
		params  oas.EmbedGetParams
		ctx     func() context.Context
		handler func(t *testing.T) *Embed
		want    oas.EmbedGetRes
	}{
		{
			name:   "Success",
			params: oas.EmbedGetParams{DiagramId: "diag-1", EmbedId: "embed-1"},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Embed {
				t.Helper()

				permSvc := mocks.NewPermissionServiceMock(t)
				permSvc.EXPECT().
					HasDiagramPermission(mock.Anything, domain.DiagramID("diag-1"), domain.UserID("user-1"), []domain.DiagramCollaboratorPermission{
						domain.DiagramCollaboratorPermissionOwner,
					}).
					Return(true, nil)

				embedSvc := mocks.NewEmbedServiceMock(t)
				embedSvc.EXPECT().
					GetByID(mock.Anything, domain.EmbedID("embed-1")).
					Return(domain.DiagramEmbedInfo{
						ID:         "embed-1",
						DiagramID:  "diag-1",
						EmbedCode:  "<script src='...'></script>",
						IsActive:   true,
						ViewsCount: 10,
						CreatedAt:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						UpdatedAt:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					}, nil)

				return &Embed{
					embedService:      embedSvc,
					permissionService: permSvc,
					log:               zap.NewNop(),
				}
			},
			want: &oas.EmbedInfoResponse{
				Embed: oas.EmbedInfo{
					ID:         oas.Ulid("embed-1"),
					EmbedCode:  "<script src='...'></script>",
					IsActive:   true,
					ViewsCount: 10,
					CreatedAt:  oas.Datetime(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
					UpdatedAt:  oas.Datetime(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
				},
			},
		},
		{
			name:   "Success with last viewed at",
			params: oas.EmbedGetParams{DiagramId: "diag-1", EmbedId: "embed-1"},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Embed {
				t.Helper()

				permSvc := mocks.NewPermissionServiceMock(t)
				permSvc.EXPECT().
					HasDiagramPermission(mock.Anything, domain.DiagramID("diag-1"), domain.UserID("user-1"), []domain.DiagramCollaboratorPermission{
						domain.DiagramCollaboratorPermissionOwner,
					}).
					Return(true, nil)

				lastViewed := time.Date(2024, 3, 1, 12, 0, 0, 0, time.UTC)
				embedSvc := mocks.NewEmbedServiceMock(t)
				embedSvc.EXPECT().
					GetByID(mock.Anything, domain.EmbedID("embed-1")).
					Return(domain.DiagramEmbedInfo{
						ID:           "embed-1",
						DiagramID:    "diag-1",
						IsActive:     true,
						ViewsCount:   5,
						LastViewedAt: &lastViewed,
						CreatedAt:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
						UpdatedAt:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					}, nil)

				return &Embed{
					embedService:      embedSvc,
					permissionService: permSvc,
					log:               zap.NewNop(),
				}
			},
			want: &oas.EmbedInfoResponse{
				Embed: oas.EmbedInfo{
					ID:           oas.Ulid("embed-1"),
					IsActive:     true,
					ViewsCount:   5,
					LastViewedAt: oas.NewOptDatetime(oas.Datetime(time.Date(2024, 3, 1, 12, 0, 0, 0, time.UTC))),
					CreatedAt:    oas.Datetime(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
					UpdatedAt:    oas.Datetime(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
				},
			},
		},
		{
			name:   "Unauthorized",
			params: oas.EmbedGetParams{DiagramId: "diag-1", EmbedId: "embed-1"},
			ctx:    context.Background,
			handler: func(t *testing.T) *Embed {
				t.Helper()
				return &Embed{log: zap.NewNop()}
			},
			want: &oas.EmbedGetUnauthorized{},
		},
		{
			name:   "Forbidden - no access",
			params: oas.EmbedGetParams{DiagramId: "diag-1", EmbedId: "embed-1"},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Embed {
				t.Helper()

				permSvc := mocks.NewPermissionServiceMock(t)
				permSvc.EXPECT().
					HasDiagramPermission(mock.Anything, domain.DiagramID("diag-1"), domain.UserID("user-1"), []domain.DiagramCollaboratorPermission{
						domain.DiagramCollaboratorPermissionOwner,
					}).
					Return(false, nil)

				return &Embed{
					permissionService: permSvc,
					log:               zap.NewNop(),
				}
			},
			want: &oas.EmbedGetForbidden{},
		},
		{
			name:   "Not found",
			params: oas.EmbedGetParams{DiagramId: "diag-1", EmbedId: "embed-not-found"},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Embed {
				t.Helper()

				permSvc := mocks.NewPermissionServiceMock(t)
				permSvc.EXPECT().
					HasDiagramPermission(mock.Anything, domain.DiagramID("diag-1"), domain.UserID("user-1"), []domain.DiagramCollaboratorPermission{
						domain.DiagramCollaboratorPermissionOwner,
					}).
					Return(true, nil)

				embedSvc := mocks.NewEmbedServiceMock(t)
				embedSvc.EXPECT().
					GetByID(mock.Anything, domain.EmbedID("embed-not-found")).
					Return(domain.DiagramEmbedInfo{}, domain.ErrEmbedNotFound)

				return &Embed{
					embedService:      embedSvc,
					permissionService: permSvc,
					log:               zap.NewNop(),
				}
			},
			want: &oas.EmbedGetNotFound{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := tt.handler(t)
			got, err := h.EmbedGet(tt.ctx(), tt.params)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestEmbed_EmbedDelete(t *testing.T) {
	tests := []struct {
		name    string
		params  oas.EmbedDeleteParams
		ctx     func() context.Context
		handler func(t *testing.T) *Embed
		want    oas.EmbedDeleteRes
	}{
		{
			name:   "Success",
			params: oas.EmbedDeleteParams{DiagramId: "diag-1", EmbedId: "embed-1"},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Embed {
				t.Helper()

				permSvc := mocks.NewPermissionServiceMock(t)
				permSvc.EXPECT().
					HasDiagramPermission(mock.Anything, domain.DiagramID("diag-1"), domain.UserID("user-1"), []domain.DiagramCollaboratorPermission{
						domain.DiagramCollaboratorPermissionOwner,
					}).
					Return(true, nil)

				embedSvc := mocks.NewEmbedServiceMock(t)
				embedSvc.EXPECT().
					SoftDelete(mock.Anything, domain.EmbedID("embed-1")).
					Return(nil)

				return &Embed{
					embedService:      embedSvc,
					permissionService: permSvc,
					log:               zap.NewNop(),
				}
			},
			want: &oas.EmbedDeleteNoContent{},
		},
		{
			name:   "Unauthorized",
			params: oas.EmbedDeleteParams{DiagramId: "diag-1", EmbedId: "embed-1"},
			ctx:    context.Background,
			handler: func(t *testing.T) *Embed {
				t.Helper()
				return &Embed{log: zap.NewNop()}
			},
			want: &oas.EmbedDeleteUnauthorized{},
		},
		{
			name:   "Forbidden - not owner",
			params: oas.EmbedDeleteParams{DiagramId: "diag-1", EmbedId: "embed-1"},
			ctx: func() context.Context {
				return domain.Claims{UserID: "user-1", SessionID: "sess-1"}.Context(context.Background())
			},
			handler: func(t *testing.T) *Embed {
				t.Helper()

				permSvc := mocks.NewPermissionServiceMock(t)
				permSvc.EXPECT().
					HasDiagramPermission(mock.Anything, domain.DiagramID("diag-1"), domain.UserID("user-1"), []domain.DiagramCollaboratorPermission{
						domain.DiagramCollaboratorPermissionOwner,
					}).
					Return(false, nil)

				return &Embed{
					permissionService: permSvc,
					log:               zap.NewNop(),
				}
			},
			want: &oas.EmbedDeleteForbidden{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := tt.handler(t)
			got, err := h.EmbedDelete(tt.ctx(), tt.params)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
