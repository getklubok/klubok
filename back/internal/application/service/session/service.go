package session

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
}

func NewService(querier Querier, log *zap.Logger) *Service {
	return &Service{
		querier: querier,

		log: log.Named("session_service"),
	}
}

func (s *Service) Create(ctx context.Context, session domain.Session) (domain.Session, error) {
	sess, err := s.querier.SessionCreate(ctx, sqlc.SessionCreateParams{
		ID:        session.ID.String(),
		Cid:       session.CookieID.String(),
		Uid:       session.UserID.String(),
		CheckedAt: session.CheckedAt,
	})
	if err != nil {
		return domain.Session{}, fmt.Errorf("create session: %w", err)
	}

	if err := s.querier.UserUpdateLastLoginAt(ctx, session.UserID.String()); err != nil {
		s.log.Error("Failed to update last login",
			zap.String("user_id", session.UserID.String()),
			zap.Error(err),
		)
	}

	return sessionToDomain(sess), nil
}

func (s *Service) UpdateCheckedAt(ctx context.Context, id domain.SessionID, checkedAt time.Time) error {
	_, err := s.querier.SessionUpdateCheckedAt(ctx, id.String(), checkedAt)
	if err != nil {
		return fmt.Errorf("update session: %w", err)
	}

	return nil
}

func (s *Service) DeleteByID(ctx context.Context, id domain.SessionID) error {
	if err := s.querier.SessionDeleteByID(ctx, id.String()); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}

	return nil
}

func (s *Service) CancelOld(ctx context.Context, olderThan time.Time) error {
	if err := s.querier.SessionCancelOld(ctx, olderThan); err != nil {
		return fmt.Errorf("cancel old sessions: %w", err)
	}

	return nil
}

func (s *Service) FindByID(ctx context.Context, id domain.SessionID) (domain.Session, error) {
	sess, err := s.querier.SessionFindByID(ctx, id.String())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Session{}, domain.ErrSessionNotFound
		}

		return domain.Session{}, fmt.Errorf("find session: %w", err)
	}

	return sessionToDomain(sess), nil
}

func (s *Service) FindByCID(ctx context.Context, cid domain.CookieID) (domain.Session, error) {
	sess, err := s.querier.SessionFindByCID(ctx, cid.String())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Session{}, domain.ErrSessionNotFound
		}

		return domain.Session{}, fmt.Errorf("find session: %w", err)
	}

	return sessionToDomain(sess), nil
}

func (s *Service) FindByUID(ctx context.Context, uid domain.UserID) ([]domain.Session, error) {
	sess, err := s.querier.SessionFindByUID(ctx, uid.String())
	if err != nil {
		return nil, fmt.Errorf("find session: %w", err)
	}

	return sessionsToDomains(sess), nil
}

func sessionToDomain(s sqlc.Session) domain.Session {
	return domain.Session{
		ID:        domain.SessionID(s.ID),
		CookieID:  domain.CookieID(s.Cid),
		UserID:    domain.UserID(s.Uid),
		CheckedAt: s.CheckedAt,
	}
}

func sessionsToDomains(ss []sqlc.Session) []domain.Session {
	ds := make([]domain.Session, 0, len(ss))
	for _, s := range ss {
		ds = append(ds, sessionToDomain(s))
	}

	return ds
}
