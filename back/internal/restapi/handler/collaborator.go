package handler

import (
	"fmt"

	"klubok/internal/domain"
	"klubok/internal/restapi/oas"
)

func diagramCollaboratorFromDomain(src domain.DiagramCollaborator) (oas.DiagramCollaborator, error) {
	permission, err := diagramPermissionFromDomain(src.Permission)
	if err != nil {
		return oas.DiagramCollaborator{}, fmt.Errorf("convert permission: %w", err)
	}

	return oas.DiagramCollaborator{
		ID:         oas.Ulid(src.ID),
		DiagramId:  oas.Ulid(src.DiagramID),
		UserName:   oas.Username(src.UserName),
		Permission: permission,
		CreatedAt:  oas.Datetime(src.CreatedAt),
		UpdatedAt:  oas.Datetime(src.UpdatedAt),
	}, nil
}

func diagramCollaboratorsFromDomain(src []domain.DiagramCollaborator) []oas.DiagramCollaborator {
	if src == nil {
		return nil
	}

	dst := make([]oas.DiagramCollaborator, 0, len(src))
	for _, item := range src {
		c, err := diagramCollaboratorFromDomain(item)
		if err != nil {
			continue
		}

		dst = append(dst, c)
	}

	return dst
}
