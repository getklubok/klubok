package handler

import (
	"context"

	"go.uber.org/zap"

	"klubok/internal/application/zapctx"
	"klubok/internal/domain"
	"klubok/internal/restapi/oas"
)

type Comment struct {
	diagramService    DiagramService
	permissionService PermissionService
	sanitizer         TextSanitizer
	log               *zap.Logger
}

func NewComment(
	diagramService DiagramService,
	permissionService PermissionService,
	sanitizer TextSanitizer,
	log *zap.Logger,
) *Comment {
	return &Comment{
		diagramService:    diagramService,
		permissionService: permissionService,
		sanitizer:         sanitizer,

		log: log.Named("diagram_handler"),
	}
}

// Create comment on diagram.
//
// POST /diagrams/{diagramId}/comments/{diagramShapeId}.
func (h *Comment) CommentCreate(
	ctx context.Context,
	req *oas.CommentCreateRequest,
	params oas.CommentCreateParams,
) (oas.CommentCreateRes, error) { //nolint:unparam // It's ok
	claims, err := domain.ClaimsFromContext(ctx)
	if err != nil {
		zapctx.Grasp(ctx, h.log).Error("Get claims from context",
			zap.Error(err),
		)
		return &oas.CommentCreateUnauthorized{}, nil
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
		return &oas.CommentCreateInternalServerError{}, nil
	}
	if !ok {
		log.Warn("User is not diagram access")
		return &oas.CommentCreateForbidden{}, nil
	}

	c, err := h.diagramService.CommentCreate(ctx, claims.UserID, domain.DiagramComment{
		DiagramID:      diagramID,
		DiagramShapeID: domain.ShapeID(params.DiagramShapeId),
		UserID:         claims.UserID,
		Text:           h.sanitizer.Sanitize(req.Text),
	})
	if err != nil {
		log.Error("Create diagram comment",
			zap.Error(err),
		)
		return &oas.CommentCreateInternalServerError{}, nil
	}

	return &oas.CommentCreateResponse{
		Comment: diagramCommentFromDomain(c),
	}, nil
}

// List comments on diagram.
//
// GET /diagrams/{diagramId}/comments/{diagramShapeId}.
func (h *Comment) CommentList(ctx context.Context, params oas.CommentListParams) (oas.CommentListRes, error) { //nolint:unparam // It's ok
	claims, err := domain.ClaimsFromContext(ctx)
	if err != nil {
		zapctx.Grasp(ctx, h.log).Error("Get claims from context",
			zap.Error(err),
		)
		return &oas.CommentListUnauthorized{}, nil
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
		return &oas.CommentListInternalServerError{}, nil
	}
	if !ok {
		log.Warn("User is not diagram access")
		return &oas.CommentListForbidden{}, nil
	}

	cs, err := h.diagramService.CommentList(ctx, diagramID, domain.ShapeID(string(params.DiagramShapeId)))
	if err != nil {
		log.Error("List diagram comments",
			zap.Error(err),
		)
		return &oas.CommentListInternalServerError{}, nil
	}

	return &oas.CommentListResponse{
		Comments: diagramCommentsFromDomain(cs),
	}, nil
}
