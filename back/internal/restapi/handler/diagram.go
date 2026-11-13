package handler

import (
	"fmt"

	"klubok/internal/domain"
	"klubok/internal/restapi/oas"
)

func diagramFromDomain(src domain.Diagram) (oas.Diagram, error) {
	permission, err := diagramPermissionFromDomain(src.Permission)
	if err != nil {
		return oas.Diagram{}, fmt.Errorf("convert permission: %w", err)
	}

	return oas.Diagram{
		ID:               oas.Ulid(src.ID),
		Title:            oas.DiagramTitle(src.Title),
		Permission:       permission,
		HasCollaborators: src.HasCollaborators,
		CreatedAt:        oas.Datetime(src.CreatedAt),
		UpdatedAt:        oas.Datetime(src.UpdatedAt),
	}, nil
}

func diagramsFromDomain(src []domain.Diagram) []oas.Diagram {
	dst := make([]oas.Diagram, 0, len(src))
	for _, item := range src {
		d, err := diagramFromDomain(item)
		if err != nil {
			continue
		}

		dst = append(dst, d)
	}

	return dst
}

func diagramInfoFromDomain(src domain.DiagramInfo) (oas.DiagramInfo, error) {
	permission, err := diagramPermissionFromDomain(src.Permission)
	if err != nil {
		return oas.DiagramInfo{}, fmt.Errorf("convert permission: %w", err)
	}

	return oas.DiagramInfo{
		ID:                oas.Ulid(src.ID),
		Title:             oas.DiagramTitle(src.Title),
		Permission:        permission,
		Shapes:            diagramShapesFromDomain(src.Shapes),
		Connections:       diagramConnectionsFromDomain(src.Connections),
		ShapeCommentStats: diagramShapeCommentStatsFromDomain(src.ShapeCommentStats),
	}, nil
}

func diagramCursorPositionsFromDomain(src []domain.WebSocketCursorPosition) []oas.CursorPosition {
	if src == nil {
		return nil
	}

	dst := make([]oas.CursorPosition, 0, len(src))
	for _, item := range src {
		dst = append(dst, diagramCursorPositionFromDomain(item))
	}

	return dst
}

func diagramCursorPositionFromDomain(src domain.WebSocketCursorPosition) oas.CursorPosition {
	return oas.CursorPosition{
		UserId:   oas.Ulid(src.UserID),
		UserName: oas.Username(src.Username),
		X:        src.X,
		Y:        src.Y,
	}
}

func diagramOnlineUserFromDomain(src domain.WebSocketUser) oas.OnlineUser {
	return oas.OnlineUser{
		ID:       oas.Ulid(src.ID),
		UserName: oas.Username(src.Username),
	}
}

func diagramOnlineUsersFromDomain(src []domain.WebSocketUser) []oas.OnlineUser {
	if src == nil {
		return nil
	}

	dst := make([]oas.OnlineUser, 0, len(src))
	for _, item := range src {
		dst = append(dst, diagramOnlineUserFromDomain(item))
	}

	return dst
}
