package handler

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"klubok/internal/domain"
	"klubok/internal/restapi/oas"
)

type EmbedPublic struct {
	diagramService DiagramService
	embedService   EmbedService

	log *zap.Logger
}

func NewEmbedPublic(
	diagramService DiagramService,
	embedService EmbedService,
	log *zap.Logger,
) *EmbedPublic {
	return &EmbedPublic{
		diagramService: diagramService,
		embedService:   embedService,

		log: log.Named("embed_public_handler"),
	}
}

//nolint:unparam // ogen handlers must match this signature
func (h *EmbedPublic) EmbedData(ctx context.Context, params oas.EmbedDataParams) (oas.EmbedDataRes, error) {
	embed, err := h.embedService.GetByToken(ctx, params.Token)
	if err != nil {
		if errors.Is(err, domain.ErrEmbedNotFound) {
			h.log.Warn("Embed not found by token")
			return &oas.EmbedDataNotFound{}, nil
		}
		h.log.Error("Get embed by token", zap.Error(err))
		return &oas.EmbedDataInternalServerError{}, nil
	}

	diagramInfo, err := h.diagramService.Info(ctx, embed.DiagramID)
	if err != nil {
		h.log.Error("Get diagram info for embed", zap.Error(err))
		return &oas.EmbedDataInternalServerError{}, nil
	}

	if err := h.embedService.IncrementViews(ctx, embed.ID); err != nil {
		h.log.Warn("Increment embed views", zap.Error(err))
	}

	di, err := diagramInfoFromDomain(diagramInfo)
	if err != nil {
		h.log.Error("Convert diagram info", zap.Error(err))
		return &oas.EmbedDataInternalServerError{}, nil
	}

	return &oas.EmbedDataResponse{
		DiagramInfo: di,
	}, nil
}
