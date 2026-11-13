package handler

import (
	"klubok/internal/domain"
	"klubok/internal/restapi/oas"
)

func diagramConnectionsFromDomain(src []domain.DiagramConnection) []oas.DiagramConnection {
	if src == nil {
		return nil
	}

	dst := make([]oas.DiagramConnection, 0, len(src))
	for _, item := range src {
		dst = append(dst, diagramConnectionFromDomain(item))
	}

	return dst
}

func diagramConnectionFromDomain(src domain.DiagramConnection) oas.DiagramConnection {
	return oas.DiagramConnection{
		Fid: oas.Ulid(src.FromShapeID),
		Tid: oas.Ulid(src.ToShapeID),
	}
}

func diagramConnectionsToDomain(src []oas.DiagramConnection) []domain.DiagramConnection {
	if src == nil {
		return nil
	}

	dst := make([]domain.DiagramConnection, 0, len(src))
	for _, item := range src {
		dst = append(dst, diagramConnectionToDomain(item))
	}

	return dst
}

func diagramConnectionToDomain(src oas.DiagramConnection) domain.DiagramConnection {
	return domain.DiagramConnection{
		FromShapeID: domain.ShapeID(src.Fid),
		ToShapeID:   domain.ShapeID(src.Tid),
	}
}
