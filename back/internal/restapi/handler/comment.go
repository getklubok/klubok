package handler

import (
	"klubok/internal/domain"
	"klubok/internal/restapi/oas"
)

func diagramCommentFromDomain(src domain.DiagramComment) oas.DiagramComment {
	return oas.DiagramComment{
		ID:        oas.Ulid(src.ID),
		UserName:  oas.Username(src.UserName),
		Text:      src.Text,
		CreatedAt: oas.Datetime(src.CreatedAt),
	}
}

func diagramCommentsFromDomain(src []domain.DiagramComment) []oas.DiagramComment {
	if src == nil {
		return nil
	}

	dst := make([]oas.DiagramComment, 0, len(src))
	for _, item := range src {
		dst = append(dst, diagramCommentFromDomain(item))
	}

	return dst
}
