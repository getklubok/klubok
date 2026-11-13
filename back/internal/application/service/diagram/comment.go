package diagram

import (
	"klubok/internal/db/sqlc"
	"klubok/internal/domain"
)

func commentListRowsToDomains(src []sqlc.DiagramCommentListRow) []domain.DiagramComment {
	dst := make([]domain.DiagramComment, 0, len(src))
	for _, item := range src {
		dst = append(dst, commentListRowToDomain(item))
	}

	return dst
}

func commentListRowToDomain(src sqlc.DiagramCommentListRow) domain.DiagramComment {
	return domain.DiagramComment{
		ID:             domain.CommentID(src.ID),
		DiagramID:      domain.DiagramID(src.DiagramID),
		DiagramShapeID: domain.ShapeID(src.DiagramShapeID),
		UserID:         domain.UserID(src.UserID),
		UserName:       src.UserName,
		Text:           src.Text,
		CreatedAt:      src.CreatedAt,
	}
}

func commentToDomain(src sqlc.DiagramComment) domain.DiagramComment {
	return domain.DiagramComment{
		ID:             domain.CommentID(src.ID),
		DiagramID:      domain.DiagramID(src.DiagramID),
		DiagramShapeID: domain.ShapeID(src.DiagramShapeID),
		UserID:         domain.UserID(src.UserID),
		Text:           src.Text,
		CreatedAt:      src.CreatedAt,
	}
}

func commentStatRowsToDomains(src []sqlc.DiagramCommentStatsRow) []domain.DiagramShapeCommentStat {
	dst := make([]domain.DiagramShapeCommentStat, 0, len(src))
	for _, item := range src {
		dst = append(dst, commentStatRow(item))
	}

	return dst
}

func commentStatRow(src sqlc.DiagramCommentStatsRow) domain.DiagramShapeCommentStat {
	return domain.DiagramShapeCommentStat{
		ShapeID:      domain.ShapeID(src.ShapeID),
		CommentCount: src.CommentCount,
	}
}
