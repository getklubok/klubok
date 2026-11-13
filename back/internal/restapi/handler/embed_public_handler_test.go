package handler

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"klubok/internal/domain"
	"klubok/internal/restapi/handler/mocks"
	"klubok/internal/restapi/oas"
)

func TestEmbedPublic_EmbedData(t *testing.T) {
	tests := []struct {
		name    string
		params  oas.EmbedDataParams
		ctx     func() context.Context
		handler func(t *testing.T) *EmbedPublic
		want    oas.EmbedDataRes
	}{
		{
			name:   "Success",
			params: oas.EmbedDataParams{Token: "embed-token-1"},
			ctx:    context.Background,
			handler: func(t *testing.T) *EmbedPublic {
				t.Helper()

				embedSvc := mocks.NewEmbedServiceMock(t)
				embedSvc.EXPECT().
					GetByToken(mock.Anything, "embed-token-1").
					Return(domain.DiagramEmbedInfo{
						ID:        "embed-1",
						DiagramID: "diag-1",
					}, nil)

				diagramSvc := mocks.NewDiagramServiceMock(t)
				diagramSvc.EXPECT().
					Info(mock.Anything, domain.DiagramID("diag-1")).
					Return(domain.DiagramInfo{
						ID:         "diag-1",
						Title:      "Test Diagram",
						Permission: domain.DiagramCollaboratorPermissionView,
					}, nil)

				embedSvc.EXPECT().
					IncrementViews(mock.Anything, domain.EmbedID("embed-1")).
					Return(nil)

				return &EmbedPublic{
					embedService:   embedSvc,
					diagramService: diagramSvc,
					log:            zap.NewNop(),
				}
			},
			want: &oas.EmbedDataResponse{
				DiagramInfo: oas.DiagramInfo{
					ID:         oas.Ulid("diag-1"),
					Title:      oas.DiagramTitle("Test Diagram"),
					Permission: oas.DiagramCollaboratorPermissionEnumVIEW,
				},
			},
		},
		{
			name:   "Success - increment views fails but response returned",
			params: oas.EmbedDataParams{Token: "embed-token-3"},
			ctx:    context.Background,
			handler: func(t *testing.T) *EmbedPublic {
				t.Helper()

				embedSvc := mocks.NewEmbedServiceMock(t)
				embedSvc.EXPECT().
					GetByToken(mock.Anything, "embed-token-3").
					Return(domain.DiagramEmbedInfo{
						ID:        "embed-3",
						DiagramID: "diag-3",
					}, nil)

				diagramSvc := mocks.NewDiagramServiceMock(t)
				diagramSvc.EXPECT().
					Info(mock.Anything, domain.DiagramID("diag-3")).
					Return(domain.DiagramInfo{
						ID:         "diag-3",
						Title:      "Diagram",
						Permission: domain.DiagramCollaboratorPermissionView,
					}, nil)

				embedSvc.EXPECT().
					IncrementViews(mock.Anything, domain.EmbedID("embed-3")).
					Return(assert.AnError)

				return &EmbedPublic{
					embedService:   embedSvc,
					diagramService: diagramSvc,
					log:            zap.NewNop(),
				}
			},
			want: &oas.EmbedDataResponse{
				DiagramInfo: oas.DiagramInfo{
					ID:         oas.Ulid("diag-3"),
					Title:      oas.DiagramTitle("Diagram"),
					Permission: oas.DiagramCollaboratorPermissionEnumVIEW,
				},
			},
		},
		{
			name:   "Embed not found",
			params: oas.EmbedDataParams{Token: "invalid-token"},
			ctx:    context.Background,
			handler: func(t *testing.T) *EmbedPublic {
				t.Helper()

				embedSvc := mocks.NewEmbedServiceMock(t)
				embedSvc.EXPECT().
					GetByToken(mock.Anything, "invalid-token").
					Return(domain.DiagramEmbedInfo{}, domain.ErrEmbedNotFound)

				return &EmbedPublic{
					embedService: embedSvc,
					log:          zap.NewNop(),
				}
			},
			want: &oas.EmbedDataNotFound{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := tt.handler(t)
			got, err := h.EmbedData(tt.ctx(), tt.params)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
