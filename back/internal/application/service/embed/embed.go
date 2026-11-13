package embed

import (
	"time"

	"klubok/internal/db/sqlc"
	"klubok/internal/domain"
)

func embedToDomain(e sqlc.DiagramEmbed) domain.DiagramEmbedInfo {
	var lastViewedAt *time.Time
	if e.LastViewedAt.Valid {
		lastViewedAt = &e.LastViewedAt.Time
	}

	var embedCode string
	if e.IsActive {
		embedCode = e.Token
	}

	return domain.DiagramEmbedInfo{
		ID:           domain.EmbedID(e.ID),
		DiagramID:    domain.DiagramID(e.DiagramID),
		EmbedCode:    embedCode,
		IsActive:     e.IsActive,
		ViewsCount:   e.ViewsCount,
		LastViewedAt: lastViewedAt,
		CreatedAt:    e.CreatedAt,
		UpdatedAt:    e.UpdatedAt,
	}
}
