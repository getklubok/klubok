package diagram

import (
	"klubok/internal/db/sqlc"
	"klubok/internal/domain"
)

func collaboratorToDomain(src sqlc.DiagramCollaborator) domain.DiagramCollaborator {
	return domain.DiagramCollaborator{
		ID:         domain.CollaboratorID(src.ID),
		DiagramID:  domain.DiagramID(src.DiagramID),
		UserID:     domain.UserID(src.UserID),
		Permission: src.Permission,
		CreatedAt:  src.CreatedAt,
		UpdatedAt:  src.UpdatedAt,
	}
}

func collaboratorListRowsToDomains(src []sqlc.DiagramCollaboratorListRow) []domain.DiagramCollaborator {
	dst := make([]domain.DiagramCollaborator, 0, len(src))
	for _, item := range src {
		dst = append(dst, collaboratorListRowToDomain(item))
	}

	return dst
}

func collaboratorListRowToDomain(src sqlc.DiagramCollaboratorListRow) domain.DiagramCollaborator {
	return domain.DiagramCollaborator{
		ID:         domain.CollaboratorID(src.ID),
		DiagramID:  domain.DiagramID(src.DiagramID),
		UserID:     domain.UserID(src.UserID),
		UserName:   src.UserName,
		Permission: src.Permission,
		CreatedAt:  src.CreatedAt,
		UpdatedAt:  src.UpdatedAt,
	}
}
