package handler

import (
	"context"
	"errors"
	"strings"

	"go.uber.org/zap"

	"klubok/internal/application/zapctx"
	"klubok/internal/domain"
	"klubok/internal/restapi/oas"
)

type Collaborator struct {
	diagramService    DiagramService
	userService       UserService
	permissionService PermissionService
	log               *zap.Logger
}

func NewCollaborator(
	diagramService DiagramService,
	userService UserService,
	permissionService PermissionService,
	log *zap.Logger,
) *Collaborator {
	return &Collaborator{
		diagramService:    diagramService,
		userService:       userService,
		permissionService: permissionService,

		log: log.Named("diagram_handler"),
	}
}

// Upsert collaborator on diagram.
//
// POST /diagrams/{diagramId}/collaborators.
func (h *Collaborator) CollaboratorUpsert(
	ctx context.Context,
	req *oas.CollaboratorUpsertRequest,
	params oas.CollaboratorUpsertParams,
) (oas.CollaboratorUpsertRes, error) { //nolint:unparam // It's ok
	claims, err := domain.ClaimsFromContext(ctx)
	if err != nil {
		zapctx.Grasp(ctx, h.log).Error("Get claims from context",
			zap.Error(err),
		)
		return &oas.CollaboratorUpsertUnauthorized{}, nil
	}

	diagramID := domain.DiagramID(params.DiagramId)

	log := zapctx.Grasp(ctx, h.log).With(
		zap.String("diagram_id", diagramID.String()),
		zap.String("user_id", claims.UserID.String()),
	)

	if isEmail(req.UserNameOrEmail) {
		log.Warn("Email is not allowed as collaborator identifier",
			zap.String("user_name_or_email", req.UserNameOrEmail),
		)
		return &oas.CollaboratorUpsertNotFound{
			Message: oas.NewOptString(domain.ErrUserNotFound.Error()),
		}, nil
	}

	foundUser, err := h.userService.FindByName(ctx, req.UserNameOrEmail)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			log.Warn("User not found",
				zap.String("user_name_or_email", req.UserNameOrEmail),
			)
			return &oas.CollaboratorUpsertNotFound{
				Message: oas.NewOptString(domain.ErrUserNotFound.Error()),
			}, nil
		}

		log.Error("Find user by name",
			zap.Error(err),
		)
		return &oas.CollaboratorUpsertInternalServerError{}, nil
	}
	if foundUser.ID == claims.UserID {
		log.Warn("User can't add itself as collaborator")
		return &oas.CollaboratorUpsertForbidden{}, nil
	}

	if req.Permission == oas.DiagramCollaboratorPermissionEnumOWNER {
		log.Warn("User can't add owner as collaborator")
		return &oas.CollaboratorUpsertForbidden{}, nil
	}

	ok, err := h.permissionService.HasDiagramPermission(ctx, diagramID, claims.UserID, []domain.DiagramCollaboratorPermission{
		domain.DiagramCollaboratorPermissionOwner,
	})
	if err != nil {
		log.Error("Check diagram collaborator permission",
			zap.Error(err),
		)
		return &oas.CollaboratorUpsertInternalServerError{}, nil
	}
	if !ok {
		log.Warn("User is not diagram owner")
		return &oas.CollaboratorUpsertForbidden{}, nil
	}

	permission, err := diagramPermissionToDomain(req.Permission)
	if err != nil {
		log.Error("Convert permission",
			zap.Error(err),
		)
		return &oas.CollaboratorUpsertInternalServerError{}, nil
	}

	c, err := h.diagramService.CollaboratorUpsert(ctx, domain.DiagramCollaborator{
		DiagramID:  diagramID,
		UserID:     foundUser.ID,
		Permission: permission,
	})
	if err != nil {
		log.Error("Upsert diagram collaborator",
			zap.Error(err),
		)
		return &oas.CollaboratorUpsertInternalServerError{}, nil
	}

	c.UserName = foundUser.Username

	clb, err := diagramCollaboratorFromDomain(c)
	if err != nil {
		log.Error("Convert collaborator",
			zap.Error(err),
		)
		return &oas.CollaboratorUpsertInternalServerError{}, nil
	}

	return &oas.CollaboratorUpsertResponse{
		Collaborator: clb,
	}, nil
}

// List collaborators on diagram.
//
// GET /diagrams/{diagramId}/collaborators.
//
//nolint:unparam,dupl // It's ok
func (h *Collaborator) CollaboratorList(ctx context.Context, params oas.CollaboratorListParams) (oas.CollaboratorListRes, error) {
	claims, err := domain.ClaimsFromContext(ctx)
	if err != nil {
		zapctx.Grasp(ctx, h.log).Error("Get claims from context",
			zap.Error(err),
		)
		return &oas.CollaboratorListUnauthorized{}, nil
	}

	diagramID := domain.DiagramID(params.DiagramId)

	log := zapctx.Grasp(ctx, h.log).With(
		zap.String("diagram_id", diagramID.String()),
		zap.String("user_id", claims.UserID.String()),
	)

	ok, err := h.permissionService.HasDiagramPermission(ctx, diagramID, claims.UserID, []domain.DiagramCollaboratorPermission{
		domain.DiagramCollaboratorPermissionOwner,
	})
	if err != nil {
		log.Error("Check diagram collaborator permission",
			zap.Error(err),
		)
		return &oas.CollaboratorListInternalServerError{}, nil
	}
	if !ok {
		log.Warn("User is not diagram owner")
		return &oas.CollaboratorListForbidden{}, nil
	}

	cs, err := h.diagramService.CollaboratorList(ctx, diagramID)
	if err != nil {
		log.Error("List diagram collaborators",
			zap.Error(err),
		)
		return &oas.CollaboratorListInternalServerError{}, nil
	}

	return &oas.CollaboratorListResponse{
		Collaborators: diagramCollaboratorsFromDomain(cs),
	}, nil
}

// Delete collaborator on diagram.
//
// DELETE /diagrams/{diagramId}/collaborators/{collaboratorId}.
func (h *Collaborator) CollaboratorDelete(
	ctx context.Context,
	params oas.CollaboratorDeleteParams,
) (oas.CollaboratorDeleteRes, error) { //nolint:unparam // It's ok
	claims, err := domain.ClaimsFromContext(ctx)
	if err != nil {
		zapctx.Grasp(ctx, h.log).Error("Get claims from context",
			zap.Error(err),
		)
		return &oas.CollaboratorDeleteUnauthorized{}, nil
	}

	diagramID := domain.DiagramID(params.DiagramId)
	collaboratorID := domain.CollaboratorID(params.CollaboratorId)

	log := zapctx.Grasp(ctx, h.log).With(
		zap.String("diagram_id", diagramID.String()),
		zap.String("user_id", claims.UserID.String()),
	)

	if domain.CollaboratorID(claims.UserID) == collaboratorID {
		log.Warn("User can't delete yourself as collaborator")
		return &oas.CollaboratorDeleteForbidden{}, nil
	}

	ok, err := h.permissionService.HasDiagramPermission(ctx, diagramID, claims.UserID, []domain.DiagramCollaboratorPermission{
		domain.DiagramCollaboratorPermissionOwner,
	})
	if err != nil {
		log.Error("Check diagram collaborator permission",
			zap.Error(err),
		)
		return &oas.CollaboratorDeleteInternalServerError{}, nil
	}
	if !ok {
		log.Warn("User is not diagram owner")
		return &oas.CollaboratorDeleteForbidden{}, nil
	}

	if err := h.diagramService.CollaboratorDelete(ctx, diagramID, collaboratorID); err != nil {
		log.Error("Delete diagram collaborator",
			zap.Error(err),
		)
		return &oas.CollaboratorDeleteInternalServerError{}, nil
	}

	return &oas.CollaboratorDeleteNoContent{}, nil
}

// isEmail reports whether the string is recognizable as an email address.
// Nicknames registered by contract cannot contain "@", so any string with
// "@" in a valid position is rejected as an email before hitting the DB.
func isEmail(s string) bool {
	at := strings.IndexByte(s, '@')
	return at > 0 && at < len(s)-1
}
