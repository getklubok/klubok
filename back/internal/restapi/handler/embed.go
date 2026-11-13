package handler

import (
	"klubok/internal/domain"
	"klubok/internal/restapi/oas"
)

func embedInfoFromDomain(src domain.DiagramEmbedInfo) oas.EmbedInfo {
	var lastViewedAt oas.OptDatetime
	if src.LastViewedAt != nil {
		lastViewedAt = oas.NewOptDatetime(oas.Datetime(*src.LastViewedAt))
	} else {
		lastViewedAt = oas.OptDatetime{
			Value: oas.Datetime{},
			Set:   false,
		}
	}

	return oas.EmbedInfo{
		ID:           oas.Ulid(src.ID),
		EmbedCode:    src.EmbedCode,
		IsActive:     src.IsActive,
		ViewsCount:   src.ViewsCount,
		LastViewedAt: lastViewedAt,
		CreatedAt:    oas.Datetime(src.CreatedAt),
		UpdatedAt:    oas.Datetime(src.UpdatedAt),
	}
}

func embedInfoListFromDomain(src []domain.DiagramEmbedInfo) []oas.EmbedInfo {
	dst := make([]oas.EmbedInfo, 0, len(src))
	for _, item := range src {
		dst = append(dst, embedInfoFromDomain(item))
	}
	return dst
}
