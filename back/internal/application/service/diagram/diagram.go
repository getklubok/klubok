package diagram

import (
	"klubok/internal/db/sqlc"
	"klubok/internal/domain"
)

func diagramListRowsToDomains(ss []sqlc.DiagramListRow) []domain.Diagram {
	ds := make([]domain.Diagram, 0, len(ss))
	for _, s := range ss {
		ds = append(ds, diagramListRowToDomain(s))
	}

	return ds
}

func diagramListRowToDomain(s sqlc.DiagramListRow) domain.Diagram {
	return domain.Diagram{
		ID:               domain.DiagramID(s.ID),
		Title:            s.Title,
		Permission:       s.Permission,
		HasCollaborators: s.HasCollaborators,
		CreatedAt:        s.CreatedAt,
		UpdatedAt:        s.UpdatedAt,
	}
}

func diagramToDomain(s sqlc.DiagramGetByUserRow) domain.Diagram {
	return domain.Diagram{
		ID:         domain.DiagramID(s.ID),
		Title:      s.Title,
		Permission: s.Permission,
		CreatedAt:  s.CreatedAt,
		UpdatedAt:  s.UpdatedAt,
	}
}
