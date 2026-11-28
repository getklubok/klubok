package handler

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"klubok/internal/application/service/diagram"
	"klubok/internal/application/zapctx"
	"klubok/internal/domain"
	"klubok/internal/restapi/oas"
)

type Shape struct {
	diagramService    DiagramService
	permissionService PermissionService
	sanitizer         TextSanitizer
	log               *zap.Logger
}

func NewShape(
	diagramService DiagramService,
	permissionService PermissionService,
	sanitizer TextSanitizer,
	log *zap.Logger,
) *Shape {
	return &Shape{
		diagramService:    diagramService,
		permissionService: permissionService,
		sanitizer:         sanitizer,

		log: log.Named("diagram_handler"),
	}
}

// Upsert shape on diagram.
//
// POST /diagrams/{diagramId}/shapes.
func (h *Shape) ShapeUpsert(
	ctx context.Context,
	req *oas.ShapeUpsertRequest,
	params oas.ShapeUpsertParams,
) (oas.ShapeUpsertRes, error) { //nolint:unparam // It's ok
	claims, err := domain.ClaimsFromContext(ctx)
	if err != nil {
		zapctx.Grasp(ctx, h.log).Error("Get claims from context",
			zap.Error(err),
		)
		return &oas.ShapeUpsertUnauthorized{}, nil
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
		return &oas.ShapeUpsertInternalServerError{}, nil
	}
	if !ok {
		log.Warn("User is not diagram owner or has not access edit to diagram")
		return &oas.ShapeUpsertForbidden{}, nil
	}

	req.Shape.Txt = h.sanitizer.Sanitize(req.Shape.Txt)

	shape, err := diagramShapeToDomain(req.Shape)
	if err != nil {
		log.Error("Convert shape to domain",
			zap.Error(err),
		)
		return &oas.ShapeUpsertInternalServerError{}, nil
	}

	if err := validateShape(shape); err != nil {
		log.Warn("Invalid shape",
			zap.Error(err),
		)
		return &oas.ShapeUpsertBadRequest{}, nil
	}

	err = h.diagramService.ShapeUpsert(ctx, claims.UserID, diagramID, shape)
	if err != nil {
		if errors.Is(err, diagram.ErrImageLimitExceeded) {
			log.Warn("Diagram image shapes limit exceeded")
			return &oas.ShapeUpsertBadRequest{}, nil
		}

		log.Error("Upsert shape diagram",
			zap.Error(err),
		)
		return &oas.ShapeUpsertInternalServerError{}, nil
	}

	return &oas.ShapeUpsertNoContent{}, nil
}

// Batch upsert shapes on diagram.
//
// POST /diagrams/{diagramId}/shapes/batch.
func (h *Shape) ShapeBatchUpsert(
	ctx context.Context,
	req *oas.ShapeBatchUpsertRequest,
	params oas.ShapeBatchUpsertParams,
) (oas.ShapeBatchUpsertRes, error) { //nolint:unparam // It's ok
	claims, err := domain.ClaimsFromContext(ctx)
	if err != nil {
		zapctx.Grasp(ctx, h.log).Error("Get claims from context",
			zap.Error(err),
		)
		return &oas.ShapeBatchUpsertUnauthorized{}, nil
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
		return &oas.ShapeBatchUpsertInternalServerError{}, nil
	}
	if !ok {
		log.Warn("User is not diagram owner or has not access edit to diagram")
		return &oas.ShapeBatchUpsertForbidden{}, nil
	}

	// Sanitize shapes text
	for i := range req.Shapes {
		req.Shapes[i].Txt = h.sanitizer.Sanitize(req.Shapes[i].Txt)
	}

	shapes := diagramShapesToDomain(req.Shapes)

	for _, shape := range shapes {
		if err := validateShape(shape); err != nil {
			log.Warn("Invalid shape",
				zap.Error(err),
			)
			return &oas.ShapeBatchUpsertBadRequest{}, nil
		}
	}

	err = h.diagramService.ShapeBatchUpsert(ctx, claims.UserID, diagramID, shapes)
	if err != nil {
		if errors.Is(err, diagram.ErrImageLimitExceeded) {
			log.Warn("Diagram image shapes limit exceeded")
			return &oas.ShapeBatchUpsertBadRequest{}, nil
		}

		log.Error("Batch upsert shapes diagram",
			zap.Error(err),
		)
		return &oas.ShapeBatchUpsertInternalServerError{}, nil
	}

	return &oas.ShapeBatchUpsertNoContent{}, nil
}

// Delete shape on diagram.
//
// DELETE /diagrams/{diagramId}/shapes/{diagramShapeId}.
func (h *Shape) ShapeDelete(ctx context.Context, params oas.ShapeDeleteParams) (oas.ShapeDeleteRes, error) { //nolint:unparam // It's ok
	claims, err := domain.ClaimsFromContext(ctx)
	if err != nil {
		zapctx.Grasp(ctx, h.log).Error("Get claims from context",
			zap.Error(err),
		)
		return &oas.ShapeDeleteUnauthorized{}, nil
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
		return &oas.ShapeDeleteInternalServerError{}, nil
	}
	if !ok {
		log.Warn("User is not diagram owner or has not access edit to diagram")
		return &oas.ShapeDeleteForbidden{}, nil
	}

	err = h.diagramService.ShapeDelete(ctx, claims.UserID, diagramID, domain.ShapeID(string(params.DiagramShapeId)))
	if err != nil {
		log.Error("Delete shape diagram",
			zap.Error(err),
		)
		return &oas.ShapeDeleteInternalServerError{}, nil
	}

	return &oas.ShapeDeleteNoContent{}, nil
}
