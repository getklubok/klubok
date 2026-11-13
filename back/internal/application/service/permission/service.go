package permission

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"

	"klubok/internal/domain"
)

type Service struct {
	querier Querier
}

func NewService(querier Querier) *Service {
	return &Service{
		querier: querier,
	}
}

func (s *Service) HasDiagramPermission(
	ctx context.Context,
	diagramID domain.DiagramID,
	userID domain.UserID,
	permissions []domain.DiagramCollaboratorPermission,
) (bool, error) {
	record, err := s.querier.DiagramCollaboratorExist(ctx, diagramID.String(), userID.String())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}

		return false, fmt.Errorf("check diagram collaborator: %w", err)
	}

	if slices.Contains(permissions, record.Permission) {
		return true, nil
	}

	return false, nil
}
