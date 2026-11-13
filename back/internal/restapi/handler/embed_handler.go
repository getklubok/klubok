package handler

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"klubok/internal/application/zapctx"
	"klubok/internal/domain"
	"klubok/internal/restapi/oas"
)

type Embed struct {
	embedService      EmbedService
	permissionService PermissionService

	log *zap.Logger
}

func NewEmbed(
	embedService EmbedService,
	permissionService PermissionService,
	log *zap.Logger,
) *Embed {
	return &Embed{
		embedService:      embedService,
		permissionService: permissionService,

		log: log.Named("embed_handler"),
	}
}

//nolint:unparam // ogen handlers must match this signature // ogen handlers must match this signature
func (h *Embed) EmbedCreate(ctx context.Context, params oas.EmbedCreateParams) (oas.EmbedCreateRes, error) {
	claims, err := domain.ClaimsFromContext(ctx)
	if err != nil {
		zapctx.Grasp(ctx, h.log).Error("Get claims from context", zap.Error(err))
		return &oas.EmbedCreateUnauthorized{}, nil
	}

	diagramID := domain.DiagramID(params.DiagramId)

	log := zapctx.Grasp(ctx, h.log).With(
		zap.String("diagram_id", diagramID.String()),
		zap.String("user_id", claims.UserID.String()),
	)

	hasPermission, err := h.permissionService.HasDiagramPermission(ctx, diagramID, claims.UserID, []domain.DiagramCollaboratorPermission{
		domain.DiagramCollaboratorPermissionOwner,
	})
	if err != nil {
		log.Error("Check diagram collaborator permission", zap.Error(err))
		return &oas.EmbedCreateInternalServerError{}, nil
	}
	if !hasPermission {
		log.Warn("User has no edit permission")
		return &oas.EmbedCreateForbidden{}, nil
	}

	embed, err := h.embedService.Create(ctx, diagramID, claims.UserID)
	if err != nil {
		log.Error("Create embed", zap.Error(err))
		return &oas.EmbedCreateInternalServerError{}, nil
	}

	return &oas.EmbedInfoResponse{
		Embed: embedInfoFromDomain(embed),
	}, nil
}

//nolint:unparam,dupl // ogen handlers must match this signature
func (h *Embed) EmbedList(ctx context.Context, params oas.EmbedListParams) (oas.EmbedListRes, error) {
	claims, err := domain.ClaimsFromContext(ctx)
	if err != nil {
		zapctx.Grasp(ctx, h.log).Error("Get claims from context", zap.Error(err))
		return &oas.EmbedListUnauthorized{}, nil
	}

	diagramID := domain.DiagramID(params.DiagramId)

	log := zapctx.Grasp(ctx, h.log).With(
		zap.String("diagram_id", diagramID.String()),
		zap.String("user_id", claims.UserID.String()),
	)

	hasPermission, err := h.permissionService.HasDiagramPermission(ctx, diagramID, claims.UserID, []domain.DiagramCollaboratorPermission{
		domain.DiagramCollaboratorPermissionOwner,
	})
	if err != nil {
		log.Error("Check diagram collaborator permission", zap.Error(err))
		return &oas.EmbedListInternalServerError{}, nil
	}
	if !hasPermission {
		log.Warn("User has no access")
		return &oas.EmbedListForbidden{}, nil
	}

	embeds, err := h.embedService.List(ctx, diagramID)
	if err != nil {
		log.Error("List embeds", zap.Error(err))
		return &oas.EmbedListInternalServerError{}, nil
	}

	return &oas.EmbedListResponse{
		Embeds: embedInfoListFromDomain(embeds),
	}, nil
}

//nolint:unparam // ogen handlers must match this signature
func (h *Embed) EmbedGet(ctx context.Context, params oas.EmbedGetParams) (oas.EmbedGetRes, error) {
	claims, err := domain.ClaimsFromContext(ctx)
	if err != nil {
		zapctx.Grasp(ctx, h.log).Error("Get claims from context", zap.Error(err))
		return &oas.EmbedGetUnauthorized{}, nil
	}

	diagramID := domain.DiagramID(params.DiagramId)
	embedID := domain.EmbedID(params.EmbedId)

	log := zapctx.Grasp(ctx, h.log).With(
		zap.String("diagram_id", diagramID.String()),
		zap.String("embed_id", embedID.String()),
		zap.String("user_id", claims.UserID.String()),
	)

	hasPermission, err := h.permissionService.HasDiagramPermission(ctx, diagramID, claims.UserID, []domain.DiagramCollaboratorPermission{
		domain.DiagramCollaboratorPermissionOwner,
	})
	if err != nil {
		log.Error("Check diagram collaborator permission", zap.Error(err))
		return &oas.EmbedGetInternalServerError{}, nil
	}
	if !hasPermission {
		log.Warn("User has no access")
		return &oas.EmbedGetForbidden{}, nil
	}

	embed, err := h.embedService.GetByID(ctx, embedID)
	if err != nil {
		if errors.Is(err, domain.ErrEmbedNotFound) {
			log.Warn("Embed not found")
			return &oas.EmbedGetNotFound{}, nil
		}
		log.Error("Get embed", zap.Error(err))
		return &oas.EmbedGetInternalServerError{}, nil
	}

	return &oas.EmbedInfoResponse{
		Embed: embedInfoFromDomain(embed),
	}, nil
}

//nolint:unparam // ogen handlers must match this signature
func (h *Embed) EmbedDelete(ctx context.Context, params oas.EmbedDeleteParams) (oas.EmbedDeleteRes, error) {
	claims, err := domain.ClaimsFromContext(ctx)
	if err != nil {
		zapctx.Grasp(ctx, h.log).Error("Get claims from context", zap.Error(err))
		return &oas.EmbedDeleteUnauthorized{}, nil
	}

	diagramID := domain.DiagramID(params.DiagramId)
	embedID := domain.EmbedID(params.EmbedId)

	log := zapctx.Grasp(ctx, h.log).With(
		zap.String("diagram_id", diagramID.String()),
		zap.String("embed_id", embedID.String()),
		zap.String("user_id", claims.UserID.String()),
	)

	hasPermission, err := h.permissionService.HasDiagramPermission(ctx, diagramID, claims.UserID, []domain.DiagramCollaboratorPermission{
		domain.DiagramCollaboratorPermissionOwner,
	})
	if err != nil {
		log.Error("Check diagram collaborator permission", zap.Error(err))
		return &oas.EmbedDeleteInternalServerError{}, nil
	}
	if !hasPermission {
		log.Warn("User is not diagram owner")
		return &oas.EmbedDeleteForbidden{}, nil
	}

	if err := h.embedService.SoftDelete(ctx, embedID); err != nil {
		log.Error("Soft delete embed", zap.Error(err))
		return &oas.EmbedDeleteInternalServerError{}, nil
	}

	return &oas.EmbedDeleteNoContent{}, nil
}
