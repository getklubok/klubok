package embed

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"

	"klubok/internal/db/sqlc"
	"klubok/internal/domain"
)

type Service struct {
	querier Querier

	log *zap.Logger

	idFunc   func() string
	nowFunc  func() time.Time
	randFunc func(l int) string
}

func NewService(querier Querier, log *zap.Logger) *Service {
	return &Service{
		querier: querier,

		log: log.Named("embed_service"),

		idFunc:   domain.NewID,
		nowFunc:  time.Now,
		randFunc: domain.NewString,
	}
}

func (s *Service) Create(
	ctx context.Context,
	diagramID domain.DiagramID,
	userID domain.UserID,
) (domain.DiagramEmbedInfo, error) {
	token := s.randFunc(domain.EmbedTokenLength)

	embed, err := s.querier.EmbedCreate(ctx, sqlc.EmbedCreateParams{
		ID:        s.idFunc(),
		DiagramID: diagramID.String(),
		UserID:    userID.String(),
		Token:     token,
	})
	if err != nil {
		return domain.DiagramEmbedInfo{}, fmt.Errorf("create embed: %w", err)
	}

	return embedToDomain(embed), nil
}

func (s *Service) List(ctx context.Context, diagramID domain.DiagramID) ([]domain.DiagramEmbedInfo, error) {
	embeds, err := s.querier.EmbedListByDiagram(ctx, diagramID.String())
	if err != nil {
		return nil, fmt.Errorf("list embeds: %w", err)
	}

	result := make([]domain.DiagramEmbedInfo, 0, len(embeds))
	for _, e := range embeds {
		result = append(result, embedToDomain(e))
	}
	return result, nil
}

func (s *Service) GetByID(ctx context.Context, embedID domain.EmbedID) (domain.DiagramEmbedInfo, error) {
	embed, err := s.querier.EmbedGetByID(ctx, embedID.String())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.DiagramEmbedInfo{}, domain.ErrEmbedNotFound
		}
		return domain.DiagramEmbedInfo{}, fmt.Errorf("get embed by id: %w", err)
	}

	return embedToDomain(embed), nil
}

func (s *Service) GetByToken(ctx context.Context, token string) (domain.DiagramEmbedInfo, error) {
	embed, err := s.querier.EmbedGetByToken(ctx, token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.DiagramEmbedInfo{}, domain.ErrEmbedNotFound
		}
		return domain.DiagramEmbedInfo{}, fmt.Errorf("get embed by token: %w", err)
	}

	if !embed.IsActive {
		return domain.DiagramEmbedInfo{}, domain.ErrEmbedNotFound
	}

	return embedToDomain(embed), nil
}

func (s *Service) SoftDelete(ctx context.Context, embedID domain.EmbedID) error {
	if err := s.querier.EmbedSoftDelete(ctx, embedID.String()); err != nil {
		return fmt.Errorf("soft delete embed: %w", err)
	}
	return nil
}

func (s *Service) IncrementViews(ctx context.Context, embedID domain.EmbedID) error {
	if err := s.querier.EmbedIncrementViews(ctx, embedID.String()); err != nil {
		return fmt.Errorf("increment embed views: %w", err)
	}
	return nil
}
