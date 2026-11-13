package handler

import (
	"context"

	"go.uber.org/zap"

	"klubok/internal/application/zapctx"
	"klubok/internal/domain"
	"klubok/internal/restapi/oas"
)

type Connection struct {
	diagramService    DiagramService
	permissionService PermissionService
	log               *zap.Logger
}

func NewConnection(
	diagramService DiagramService,
	permissionService PermissionService,
	log *zap.Logger,
) *Connection {
	return &Connection{
		diagramService:    diagramService,
		permissionService: permissionService,

		log: log.Named("diagram_handler"),
	}
}

// Upsert connection on diagram.
//
// POST /diagrams/{diagramId}/connections.
func (h *Connection) ConnectionUpsert( //nolint:dupl // It's ok
	ctx context.Context,
	req *oas.ConnectionUpsertRequest,
	params oas.ConnectionUpsertParams,
) (oas.ConnectionUpsertRes, error) { //nolint:unparam // It's ok
	claims, err := domain.ClaimsFromContext(ctx)
	if err != nil {
		zapctx.Grasp(ctx, h.log).Error("Get claims from context",
			zap.Error(err),
		)
		return &oas.ConnectionUpsertUnauthorized{}, nil
	}

	diagramID := domain.DiagramID(params.DiagramId)

	log := zapctx.Grasp(ctx, h.log).With(
		zap.String("diagram_id", diagramID.String()),
		zap.String("user_id", claims.UserID.String()),
	)

	ok, err := h.permissionService.HasDiagramPermission(ctx, diagramID, claims.UserID, []domain.DiagramCollaboratorPermission{
		domain.DiagramCollaboratorPermissionOwner,
		domain.DiagramCollaboratorPermissionEdit,
	})
	if err != nil {
		log.Error("Check diagram collaborator permission",
			zap.Error(err),
		)
		return &oas.ConnectionUpsertInternalServerError{}, nil
	}
	if !ok {
		log.Warn("User is not diagram owner or has not access edit to diagram")
		return &oas.ConnectionUpsertForbidden{}, nil
	}

	err = h.diagramService.ConnectionUpsert(ctx, claims.UserID, diagramID, diagramConnectionToDomain(req.Connection))
	if err != nil {
		log.Error("Upsert connection diagram",
			zap.Error(err),
		)
		return &oas.ConnectionUpsertInternalServerError{}, nil
	}

	return &oas.ConnectionUpsertNoContent{}, nil
}

// Batch upsert connections on diagram.
//
// POST /diagrams/{diagramId}/connections/batch.
func (h *Connection) ConnectionBatchUpsert( //nolint:dupl // It's ok
	ctx context.Context,
	req *oas.ConnectionBatchUpsertRequest,
	params oas.ConnectionBatchUpsertParams,
) (oas.ConnectionBatchUpsertRes, error) { //nolint:unparam // It's ok
	claims, err := domain.ClaimsFromContext(ctx)
	if err != nil {
		zapctx.Grasp(ctx, h.log).Error("Get claims from context",
			zap.Error(err),
		)
		return &oas.ConnectionBatchUpsertUnauthorized{}, nil
	}

	diagramID := domain.DiagramID(params.DiagramId)

	log := zapctx.Grasp(ctx, h.log).With(
		zap.String("diagram_id", diagramID.String()),
		zap.String("user_id", claims.UserID.String()),
	)

	ok, err := h.permissionService.HasDiagramPermission(ctx, diagramID, claims.UserID, []domain.DiagramCollaboratorPermission{
		domain.DiagramCollaboratorPermissionOwner,
		domain.DiagramCollaboratorPermissionEdit,
	})
	if err != nil {
		log.Error("Check diagram collaborator permission",
			zap.Error(err),
		)
		return &oas.ConnectionBatchUpsertInternalServerError{}, nil
	}
	if !ok {
		log.Warn("User is not diagram owner or has not access edit to diagram")
		return &oas.ConnectionBatchUpsertForbidden{}, nil
	}

	err = h.diagramService.ConnectionBatchUpsert(ctx, claims.UserID, diagramID, diagramConnectionsToDomain(req.Connections))
	if err != nil {
		log.Error("Batch upsert connections diagram",
			zap.Error(err),
		)
		return &oas.ConnectionBatchUpsertInternalServerError{}, nil
	}

	return &oas.ConnectionBatchUpsertNoContent{}, nil
}

// Delete connection on diagram.
//
// DELETE /diagrams/{diagramId}/connections/{fromShapeId}/{toShapeId}.
func (h *Connection) ConnectionDelete(ctx context.Context, params oas.ConnectionDeleteParams) (oas.ConnectionDeleteRes, error) { //nolint:unparam // It's ok
	claims, err := domain.ClaimsFromContext(ctx)
	if err != nil {
		zapctx.Grasp(ctx, h.log).Error("Get claims from context",
			zap.Error(err),
		)
		return &oas.ConnectionDeleteUnauthorized{}, nil
	}

	diagramID := domain.DiagramID(params.DiagramId)

	log := zapctx.Grasp(ctx, h.log).With(
		zap.String("diagram_id", diagramID.String()),
		zap.String("user_id", claims.UserID.String()),
	)

	ok, err := h.permissionService.HasDiagramPermission(ctx, diagramID, claims.UserID, []domain.DiagramCollaboratorPermission{
		domain.DiagramCollaboratorPermissionOwner,
		domain.DiagramCollaboratorPermissionEdit,
	})
	if err != nil {
		log.Error("Check diagram collaborator permission",
			zap.Error(err),
		)
		return &oas.ConnectionDeleteInternalServerError{}, nil
	}
	if !ok {
		log.Warn("User is not diagram owner or has not access edit to diagram")
		return &oas.ConnectionDeleteForbidden{}, nil
	}

	err = h.diagramService.ConnectionDelete(ctx, claims.UserID, diagramID, domain.DiagramConnection{
		FromShapeID: domain.ShapeID(params.FromShapeId),
		ToShapeID:   domain.ShapeID(params.ToShapeId),
	})
	if err != nil {
		log.Error("Delete connection diagram",
			zap.Error(err),
		)
		return &oas.ConnectionDeleteInternalServerError{}, nil
	}

	return &oas.ConnectionDeleteNoContent{}, nil
}
