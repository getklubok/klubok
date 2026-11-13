package handler

import (
	"fmt"

	"klubok/internal/domain"
	"klubok/internal/restapi/oas"
)

func diagramShapesToDomain(src []oas.DiagramShape) []domain.DiagramShape {
	if src == nil {
		return nil
	}

	dst := make([]domain.DiagramShape, 0, len(src))
	for _, item := range src {
		s, err := diagramShapeToDomain(item)
		if err != nil {
			continue
		}

		dst = append(dst, s)
	}

	return dst
}

func diagramShapeToDomain(src oas.DiagramShape) (domain.DiagramShape, error) {
	typ, err := diagramShapeTypeToDomain(src.Tp)
	if err != nil {
		return domain.DiagramShape{}, fmt.Errorf("convert shape type: %w", err)
	}

	return domain.DiagramShape{
		ID:     domain.ShapeID(src.ID),
		Type:   typ,
		Left:   src.L,
		Top:    src.T,
		Width:  src.W,
		Height: src.H,
		ScaleX: src.Sx,
		ScaleY: src.Sy,
		Fill:   src.Fc,
		Stroke: src.Sc,
		Text:   src.Txt,
		Order:  src.O,
	}, nil
}

func diagramShapesFromDomain(src []domain.DiagramShape) []oas.DiagramShape {
	if src == nil {
		return nil
	}

	dst := make([]oas.DiagramShape, 0, len(src))
	for _, item := range src {
		s, err := diagramShapeFromDomain(item)
		if err != nil {
			continue
		}

		dst = append(dst, s)
	}

	return dst
}

func diagramShapeFromDomain(src domain.DiagramShape) (oas.DiagramShape, error) {
	typ, err := diagramShapeTypeFromDomain(src.Type)
	if err != nil {
		return oas.DiagramShape{}, fmt.Errorf("convert shape type: %w", err)
	}

	// Backward compatibility
	if src.Width == 0 {
		src.Width = domain.DefaultShapeSize
	}

	// Backward compatibility
	if src.Height == 0 {
		src.Height = domain.DefaultShapeSize
	}

	return oas.DiagramShape{
		ID:  oas.Ulid(src.ID),
		Tp:  typ,
		L:   src.Left,
		T:   src.Top,
		W:   src.Width,
		H:   src.Height,
		Sx:  src.ScaleX,
		Sy:  src.ScaleY,
		Fc:  src.Fill,
		Sc:  src.Stroke,
		Txt: src.Text,
		O:   src.Order,
	}, nil
}

func diagramShapeCommentStatsFromDomain(src []domain.DiagramShapeCommentStat) []oas.DiagramShapeCommentStats {
	if src == nil {
		return nil
	}

	dst := make([]oas.DiagramShapeCommentStats, 0, len(src))
	for _, item := range src {
		dst = append(dst, diagramShapeCommentStatFromDomain(item))
	}

	return dst
}

func diagramShapeCommentStatFromDomain(src domain.DiagramShapeCommentStat) oas.DiagramShapeCommentStats {
	return oas.DiagramShapeCommentStats{
		ShapeId:      oas.Ulid(src.ShapeID),
		CommentCount: src.CommentCount,
	}
}

func diagramShapeTypeToDomain(src oas.DiagramShapeTypeEnum) (domain.DiagramShapeType, error) {
	switch src {
	case oas.DiagramShapeTypeEnumS:
		return domain.DiagramShapeTypeSquare, nil
	case oas.DiagramShapeTypeEnumE:
		return domain.DiagramShapeTypeEllipse, nil
	case oas.DiagramShapeTypeEnumD:
		return domain.DiagramShapeTypeDiamond, nil
	case oas.DiagramShapeTypeEnumB:
		return domain.DiagramShapeTypeBox, nil
	case oas.DiagramShapeTypeEnumT:
		return domain.DiagramShapeTypeText, nil
	}

	return domain.DiagramShapeType(""), fmt.Errorf("unknown shape type: %s", src)
}

func diagramShapeTypeFromDomain(src domain.DiagramShapeType) (oas.DiagramShapeTypeEnum, error) {
	switch src {
	case domain.DiagramShapeTypeSquare:
		return oas.DiagramShapeTypeEnumS, nil
	case domain.DiagramShapeTypeEllipse:
		return oas.DiagramShapeTypeEnumE, nil
	case domain.DiagramShapeTypeDiamond:
		return oas.DiagramShapeTypeEnumD, nil
	case domain.DiagramShapeTypeBox:
		return oas.DiagramShapeTypeEnumB, nil
	case domain.DiagramShapeTypeText:
		return oas.DiagramShapeTypeEnumT, nil
	}

	// Backward compatibility
	if string(src) == "CIRCLE" {
		return oas.DiagramShapeTypeEnumE, nil
	}

	return oas.DiagramShapeTypeEnum(""), fmt.Errorf("unknown shape type: %s", src)
}
