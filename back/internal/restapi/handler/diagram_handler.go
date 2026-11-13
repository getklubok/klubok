package handler

import (
	"context"

	"go.uber.org/zap"

	"klubok/internal/application/zapctx"
	"klubok/internal/domain"
	"klubok/internal/restapi/oas"
)

type Diagram struct {
	diagramService    DiagramService
	permissionService PermissionService
	wsManager         WebSocketManager
	sanitizer         TextSanitizer
	log               *zap.Logger
}

func NewDiagram(
	diagramService DiagramService,
	permissionService PermissionService,
	wsManager WebSocketManager,
	sanitizer TextSanitizer,
	log *zap.Logger,
) *Diagram {
	return &Diagram{
		diagramService:    diagramService,
		permissionService: permissionService,
		wsManager:         wsManager,
		sanitizer:         sanitizer,

		log: log.Named("diagram_handler"),
	}
}

// Create diagram.
//
// POST /diagrams.
func (h *Diagram) DiagramCreate(ctx context.Context, req *oas.DiagramCreateRequest) (oas.DiagramCreateRes, error) { //nolint:unparam // It's ok
	claims, err := domain.ClaimsFromContext(ctx)
	if err != nil {
		zapctx.Grasp(ctx, h.log).Error("Get claims from context",
			zap.Error(err),
		)
		return &oas.DiagramCreateUnauthorized{}, nil
	}

	title := h.sanitizer.Sanitize(string(req.Title))
	if title == "" {
		zapctx.Grasp(ctx, h.log).Warn("Title is empty")
		return &oas.DiagramCreateBadRequest{}, nil
	}

	d, err := h.diagramService.Create(ctx, title, claims.UserID)
	if err != nil {
		zapctx.Grasp(ctx, h.log).Error("Create diagram",
			zap.Error(err),
		)
		return &oas.DiagramCreateInternalServerError{}, nil
	}

	dg, err := diagramFromDomain(d)
	if err != nil {
		zapctx.Grasp(ctx, h.log).Error("Convert diagram",
			zap.Error(err),
		)
		return &oas.DiagramCreateInternalServerError{}, nil
	}

	return &oas.DiagramResponse{
		Diagram: dg,
	}, nil
}

// List diagrams.
//
// GET /diagrams.
func (h *Diagram) DiagramList(ctx context.Context) (oas.DiagramListRes, error) { //nolint:unparam // It's ok
	claims, err := domain.ClaimsFromContext(ctx)
	if err != nil {
		zapctx.Grasp(ctx, h.log).Error("Get claims from context",
			zap.Error(err),
		)
		return &oas.DiagramListUnauthorized{}, nil
	}

	ds, err := h.diagramService.List(ctx, claims.UserID)
	if err != nil {
		zapctx.Grasp(ctx, h.log).Error("List diagrams",
			zap.Error(err),
		)
		return &oas.DiagramListInternalServerError{}, nil
	}

	return &oas.DiagramListResponse{
		Diagrams: diagramsFromDomain(ds),
	}, nil
}

// Delete diagram.
//
// DELETE /diagrams/{diagramId}.
func (h *Diagram) DiagramDelete(ctx context.Context, params oas.DiagramDeleteParams) (oas.DiagramDeleteRes, error) { //nolint:unparam // It's ok
	claims, err := domain.ClaimsFromContext(ctx)
	if err != nil {
		zapctx.Grasp(ctx, h.log).Error("Get claims from context",
			zap.Error(err),
		)
		return &oas.DiagramDeleteUnauthorized{}, nil
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
		return &oas.DiagramDeleteInternalServerError{}, nil
	}
	if !ok {
		log.Warn("User is not diagram owner")
		return &oas.DiagramDeleteForbidden{}, nil
	}

	if err := h.diagramService.Delete(ctx, diagramID); err != nil {
		log.Error("Delete diagram",
			zap.Error(err),
		)
		return &oas.DiagramDeleteInternalServerError{}, nil
	}

	return &oas.DiagramDeleteNoContent{}, nil
}

// Set title diagram.
//
// PATCH /diagrams/{diagramId}/title.
func (h *Diagram) DiagramSetTitle(
	ctx context.Context,
	req *oas.DiagramSetTitleRequest,
	params oas.DiagramSetTitleParams,
) (oas.DiagramSetTitleRes, error) { //nolint:unparam // It's ok
	claims, err := domain.ClaimsFromContext(ctx)
	if err != nil {
		zapctx.Grasp(ctx, h.log).Error("Get claims from context",
			zap.Error(err),
		)
		return &oas.DiagramSetTitleUnauthorized{}, nil
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
		return &oas.DiagramSetTitleInternalServerError{}, nil
	}
	if !ok {
		log.Warn("User is not diagram owner or has not access edit to diagram")
		return &oas.DiagramSetTitleForbidden{}, nil
	}

	title := h.sanitizer.Sanitize(string(req.Title))
	if title == "" {
		log.Warn("Title is empty")
		return &oas.DiagramSetTitleBadRequest{}, nil
	}

	if err := h.diagramService.SetTitle(ctx, diagramID, title); err != nil {
		log.Error("Set title diagram",
			zap.Error(err),
		)
		return &oas.DiagramSetTitleInternalServerError{}, nil
	}

	return &oas.DiagramSetTitleNoContent{}, nil
}

// Info diagram.
//
// GET /diagrams/{diagramId}.
func (h *Diagram) DiagramInfo(ctx context.Context, params oas.DiagramInfoParams) (oas.DiagramInfoRes, error) { //nolint:unparam // It's ok
	claims, err := domain.ClaimsFromContext(ctx)
	if err != nil {
		zapctx.Grasp(ctx, h.log).Error("Get claims from context",
			zap.Error(err),
		)
		return &oas.DiagramInfoUnauthorized{}, nil
	}

	diagramID := domain.DiagramID(params.DiagramId)

	log := zapctx.Grasp(ctx, h.log).With(
		zap.String("diagram_id", diagramID.String()),
		zap.String("user_id", claims.UserID.String()),
	)

	ok, err := h.permissionService.HasDiagramPermission(ctx, diagramID, claims.UserID, []domain.DiagramCollaboratorPermission{
		domain.DiagramCollaboratorPermissionOwner,
		domain.DiagramCollaboratorPermissionView,
		domain.DiagramCollaboratorPermissionEdit,
	})
	if err != nil {
		log.Error("Check diagram collaborator permission",
			zap.Error(err),
		)
		return &oas.DiagramInfoInternalServerError{}, nil
	}
	if !ok {
		log.Warn("User is not diagram access")
		return &oas.DiagramInfoForbidden{}, nil
	}

	d, err := h.diagramService.InfoByUser(ctx, diagramID, claims.UserID)
	if err != nil {
		log.Error("Info diagram",
			zap.Error(err),
		)
		return &oas.DiagramInfoInternalServerError{}, nil
	}

	di, err := diagramInfoFromDomain(d)
	if err != nil {
		log.Error("Convert diagram",
			zap.Error(err),
		)
		return &oas.DiagramInfoInternalServerError{}, nil
	}

	return &oas.DiagramInfoResponse{
		DiagramInfo: di,
	}, nil
}

// Stat diagram.
//
// GET /diagrams/{diagramId}/stat.
func (h *Diagram) DiagramStat(ctx context.Context, params oas.DiagramStatParams) (oas.DiagramStatRes, error) { //nolint:unparam // It's ok
	claims, err := domain.ClaimsFromContext(ctx)
	if err != nil {
		zapctx.Grasp(ctx, h.log).Error("Get claims from context",
			zap.Error(err),
		)
		return &oas.DiagramStatUnauthorized{}, nil
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
		return &oas.DiagramStatInternalServerError{}, nil
	}
	if !ok {
		log.Warn("User is not diagram access")
		return &oas.DiagramStatForbidden{}, nil
	}

	stat := h.wsManager.GetDiagramStats(diagramID)

	return &oas.DiagramStatResponse{
		WebsocketConnections: int32(stat.Connections), //nolint:gosec // It's ok
		CursorPositions:      diagramCursorPositionsFromDomain(stat.CursorPositions),
	}, nil
}

// Online users on diagram.
//
// GET /diagrams/{diagramId}/onlineUsers.
func (h *Diagram) DiagramOnlineUsers(ctx context.Context, params oas.DiagramOnlineUsersParams) (oas.DiagramOnlineUsersRes, error) { //nolint:unparam // It's ok
	claims, err := domain.ClaimsFromContext(ctx)
	if err != nil {
		zapctx.Grasp(ctx, h.log).Error("Get claims from context",
			zap.Error(err),
		)
		return &oas.DiagramOnlineUsersUnauthorized{}, nil
	}

	diagramID := domain.DiagramID(params.DiagramId)

	log := zapctx.Grasp(ctx, h.log).With(
		zap.String("diagram_id", diagramID.String()),
		zap.String("user_id", claims.UserID.String()),
	)

	ok, err := h.permissionService.HasDiagramPermission(ctx, diagramID, claims.UserID, []domain.DiagramCollaboratorPermission{
		domain.DiagramCollaboratorPermissionOwner,
		domain.DiagramCollaboratorPermissionView,
		domain.DiagramCollaboratorPermissionEdit,
	})
	if err != nil {
		log.Error("Check diagram collaborator permission",
			zap.Error(err),
		)
		return &oas.DiagramOnlineUsersInternalServerError{}, nil
	}
	if !ok {
		log.Warn("User is not diagram owner")
		return &oas.DiagramOnlineUsersForbidden{}, nil
	}

	return &oas.DiagramOnlineUsersResponse{
		OnlineUsers: diagramOnlineUsersFromDomain(
			h.diagramService.OnlineUsers(ctx, diagramID),
		),
	}, nil
}

// Clone a diagram with all its contents.
//
// POST /diagrams/{diagramId}/clone.
func (h *Diagram) DiagramClone(ctx context.Context, params oas.DiagramCloneParams) (oas.DiagramCloneRes, error) { //nolint:unparam // It's ok
	claims, err := domain.ClaimsFromContext(ctx)
	if err != nil {
		zapctx.Grasp(ctx, h.log).Error("Get claims from context",
			zap.Error(err),
		)
		return &oas.DiagramCloneUnauthorized{}, nil
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
		return &oas.DiagramCloneInternalServerError{}, nil
	}
	if !ok {
		log.Warn("User is not diagram owner or has not access edit to diagram")
		return &oas.DiagramCloneForbidden{}, nil
	}

	d, err := h.diagramService.Clone(ctx, diagramID, claims.UserID)
	if err != nil {
		log.Error("Clone diagram",
			zap.Error(err),
		)
		return &oas.DiagramCloneInternalServerError{}, nil
	}

	dg, err := diagramFromDomain(d)
	if err != nil {
		log.Error("Convert diagram",
			zap.Error(err),
		)
		return &oas.DiagramCloneInternalServerError{}, nil
	}

	return &oas.DiagramResponse{
		Diagram: dg,
	}, nil
}

// Import a diagram.
//
// POST /diagrams/import.
func (h *Diagram) DiagramImport(ctx context.Context, req *oas.DiagramImportRequest) (oas.DiagramImportRes, error) { //nolint:unparam // It's ok
	claims, err := domain.ClaimsFromContext(ctx)
	if err != nil {
		zapctx.Grasp(ctx, h.log).Error("Get claims from context",
			zap.Error(err),
		)
		return &oas.DiagramImportUnauthorized{}, nil
	}

	log := zapctx.Grasp(ctx, h.log).With(
		zap.String("user_id", claims.UserID.String()),
	)

	// Sanitize shapes text
	for i := range req.Shapes {
		req.Shapes[i].Txt = h.sanitizer.Sanitize(req.Shapes[i].Txt)
	}

	d, err := h.diagramService.Import(ctx, claims.UserID, domain.DiagramInfo{
		Title:       h.sanitizer.Sanitize(string(req.Title)),
		Permission:  domain.DiagramCollaboratorPermissionOwner,
		Shapes:      diagramShapesToDomain(req.Shapes),
		Connections: diagramConnectionsToDomain(req.Connections),
	})
	if err != nil {
		log.Error("Import diagram",
			zap.Error(err),
		)
		return &oas.DiagramImportInternalServerError{}, nil
	}

	dg, err := diagramFromDomain(d)
	if err != nil {
		log.Error("Convert diagram",
			zap.Error(err),
		)
		return &oas.DiagramImportInternalServerError{}, nil
	}

	return &oas.DiagramResponse{
		Diagram: dg,
	}, nil
}
