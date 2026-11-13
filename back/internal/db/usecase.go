package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"klubok/internal/db/sqlc"
	"klubok/internal/domain"
)

type UseCase struct {
	db      *pgxpool.Pool
	queries *sqlc.Queries

	idFunc func() string
}

func NewUseCase(db *pgxpool.Pool, queries *sqlc.Queries) *UseCase {
	return &UseCase{
		db:      db,
		queries: queries,

		idFunc: domain.NewID,
	}
}

func (uc *UseCase) UserCreate(
	ctx context.Context,
	user sqlc.UserCreateParams,
	privacyLog sqlc.PrivacyLogCreateParams,
) (sqlc.User, error) {
	tx, err := uc.db.Begin(ctx)
	if err != nil {
		return sqlc.User{}, fmt.Errorf("begin transaction: %w", err)
	}

	qtx := uc.queries.WithTx(tx)

	u, err := qtx.UserCreate(ctx, user)
	if err != nil {
		_ = tx.Rollback(ctx)

		if domain.IsDBUniqueViolationError(err) {
			return sqlc.User{}, domain.ErrUserAlreadyExists
		}

		return sqlc.User{}, fmt.Errorf("create user: %w", err)
	}

	if err := qtx.PrivacyLogCreate(ctx, privacyLog); err != nil {
		_ = tx.Rollback(ctx)

		return sqlc.User{}, fmt.Errorf("create privacy log: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		_ = tx.Rollback(ctx)

		return sqlc.User{}, fmt.Errorf("commit transaction: %w", err)
	}

	return u, nil
}

func (uc *UseCase) UserCreateWithSSO(
	ctx context.Context,
	user sqlc.UserCreateParams,
	privacyLog sqlc.PrivacyLogCreateParams,
	identity sqlc.UserSSOIdentityCreateParams,
) (sqlc.User, error) {
	identity.UserID = user.ID

	tx, err := uc.db.Begin(ctx)
	if err != nil {
		return sqlc.User{}, fmt.Errorf("begin transaction: %w", err)
	}

	qtx := uc.queries.WithTx(tx)

	u, err := qtx.UserCreate(ctx, user)
	if err != nil {
		_ = tx.Rollback(ctx)

		if domain.IsDBUniqueViolationError(err) {
			return sqlc.User{}, domain.ErrUserAlreadyExists
		}

		return sqlc.User{}, fmt.Errorf("create user: %w", err)
	}

	if err := qtx.PrivacyLogCreate(ctx, privacyLog); err != nil {
		_ = tx.Rollback(ctx)

		return sqlc.User{}, fmt.Errorf("create privacy log: %w", err)
	}

	if _, err := qtx.UserSSOIdentityCreate(ctx, identity); err != nil {
		_ = tx.Rollback(ctx)

		if domain.IsDBUniqueViolationError(err) {
			return sqlc.User{}, domain.ErrUserAlreadyExists
		}

		return sqlc.User{}, fmt.Errorf("create sso identity: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		_ = tx.Rollback(ctx)

		return sqlc.User{}, fmt.Errorf("commit transaction: %w", err)
	}

	return u, nil
}

func (uc *UseCase) DiagramCreate(
	ctx context.Context,
	id string,
	title string,
	collaborator sqlc.DiagramCollaboratorUpsertParams,
) (sqlc.DiagramGetByUserRow, error) {
	tx, err := uc.db.Begin(ctx)
	if err != nil {
		return sqlc.DiagramGetByUserRow{}, fmt.Errorf("begin transaction: %w", err)
	}

	qtx := uc.queries.WithTx(tx)

	d, err := qtx.DiagramCreate(ctx, id, title)
	if err != nil {
		_ = tx.Rollback(ctx)

		return sqlc.DiagramGetByUserRow{}, fmt.Errorf("create diagram: %w", err)
	}

	if _, err := qtx.DiagramCollaboratorUpsert(ctx, collaborator); err != nil {
		_ = tx.Rollback(ctx)

		return sqlc.DiagramGetByUserRow{}, fmt.Errorf("upsert diagram collaborator: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		_ = tx.Rollback(ctx)

		return sqlc.DiagramGetByUserRow{}, fmt.Errorf("commit transaction: %w", err)
	}

	return sqlc.DiagramGetByUserRow{
		ID:         d.ID,
		Title:      d.Title,
		Permission: collaborator.Permission,
		CreatedAt:  d.CreatedAt,
		UpdatedAt:  d.UpdatedAt,
	}, nil
}

type DiagramCloneParams struct {
	NewDiagramID  string
	NewTitle      string
	OwnerUserID   string
	Shapes        []DiagramShapeClone
	Connections   []DiagramConnectionClone
	Comments      []DiagramCommentClone
	Collaborators []DiagramCollaboratorClone
}

type DiagramShapeClone struct {
	ID   string
	Data domain.DiagramShape
}

type DiagramConnectionClone struct {
	FromShapeID string
	ToShapeID   string
}

type DiagramCommentClone struct {
	ID             string
	DiagramShapeID string
	UserID         string
	Text           string
}

type DiagramCollaboratorClone struct {
	ID         string
	UserID     string
	Permission domain.DiagramCollaboratorPermission
}

//nolint:gocognit,gocyclo,cyclop // Bulk insert operation
func (uc *UseCase) DiagramClone(ctx context.Context, params DiagramCloneParams) (sqlc.DiagramGetByUserRow, error) {
	tx, err := uc.db.Begin(ctx)
	if err != nil {
		return sqlc.DiagramGetByUserRow{}, fmt.Errorf("begin transaction: %w", err)
	}

	qtx := uc.queries.WithTx(tx)

	d, err := qtx.DiagramCreate(ctx, params.NewDiagramID, params.NewTitle)
	if err != nil {
		_ = tx.Rollback(ctx)
		return sqlc.DiagramGetByUserRow{}, fmt.Errorf("create diagram: %w", err)
	}

	if _, err := qtx.DiagramCollaboratorUpsert(ctx, sqlc.DiagramCollaboratorUpsertParams{
		ID:         uc.idFunc(),
		DiagramID:  params.NewDiagramID,
		UserID:     params.OwnerUserID,
		Permission: domain.DiagramCollaboratorPermissionOwner,
	}); err != nil {
		_ = tx.Rollback(ctx)
		return sqlc.DiagramGetByUserRow{}, fmt.Errorf("upsert diagram collaborator: %w", err)
	}

	now := time.Now()

	if len(params.Shapes) > 0 {
		shapeRows := make([]sqlc.DiagramShapeBulkInsertParams, 0, len(params.Shapes))
		for _, s := range params.Shapes {
			shapeRows = append(shapeRows, sqlc.DiagramShapeBulkInsertParams{
				ID:        s.ID,
				DiagramID: params.NewDiagramID,
				Data:      s.Data,
				CreatedAt: now,
				UpdatedAt: now,
			})
		}
		if _, err := qtx.DiagramShapeBulkInsert(ctx, shapeRows); err != nil {
			_ = tx.Rollback(ctx)
			return sqlc.DiagramGetByUserRow{}, fmt.Errorf("bulk insert shapes: %w", err)
		}
	}

	if len(params.Connections) > 0 {
		connRows := make([]sqlc.DiagramConnectionBulkInsertParams, 0, len(params.Connections))
		for _, c := range params.Connections {
			connRows = append(connRows, sqlc.DiagramConnectionBulkInsertParams{
				DiagramID:   params.NewDiagramID,
				FromShapeID: c.FromShapeID,
				ToShapeID:   c.ToShapeID,
				CreatedAt:   now,
				UpdatedAt:   now,
			})
		}
		if _, err := qtx.DiagramConnectionBulkInsert(ctx, connRows); err != nil {
			_ = tx.Rollback(ctx)
			return sqlc.DiagramGetByUserRow{}, fmt.Errorf("bulk insert connections: %w", err)
		}
	}

	if len(params.Comments) > 0 {
		commentRows := make([]sqlc.DiagramCommentBulkInsertParams, 0, len(params.Comments))
		for _, c := range params.Comments {
			commentRows = append(commentRows, sqlc.DiagramCommentBulkInsertParams{
				ID:             c.ID,
				DiagramID:      params.NewDiagramID,
				DiagramShapeID: c.DiagramShapeID,
				UserID:         c.UserID,
				Text:           c.Text,
				CreatedAt:      now,
			})
		}
		if _, err := qtx.DiagramCommentBulkInsert(ctx, commentRows); err != nil {
			_ = tx.Rollback(ctx)
			return sqlc.DiagramGetByUserRow{}, fmt.Errorf("bulk insert comments: %w", err)
		}
	}

	if len(params.Collaborators) > 0 {
		collabRows := make([]sqlc.DiagramCollaboratorBulkInsertParams, 0, len(params.Collaborators))
		for _, c := range params.Collaborators {
			collabRows = append(collabRows, sqlc.DiagramCollaboratorBulkInsertParams{
				ID:         c.ID,
				DiagramID:  params.NewDiagramID,
				UserID:     c.UserID,
				Permission: c.Permission,
				CreatedAt:  now,
				UpdatedAt:  now,
			})
		}
		if _, err := qtx.DiagramCollaboratorBulkInsert(ctx, collabRows); err != nil {
			_ = tx.Rollback(ctx)
			return sqlc.DiagramGetByUserRow{}, fmt.Errorf("bulk insert collaborators: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		_ = tx.Rollback(ctx)
		return sqlc.DiagramGetByUserRow{}, fmt.Errorf("commit transaction: %w", err)
	}

	return sqlc.DiagramGetByUserRow{
		ID:         d.ID,
		Title:      d.Title,
		Permission: domain.DiagramCollaboratorPermissionOwner,
		CreatedAt:  d.CreatedAt,
		UpdatedAt:  d.UpdatedAt,
	}, nil
}

func (uc *UseCase) DiagramShapeBatchUpsert(ctx context.Context, diagramID string, shapes []domain.DiagramShape) error {
	if len(shapes) == 0 {
		return nil
	}

	tx, err := uc.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	qtx := uc.queries.WithTx(tx)

	for _, shape := range shapes {
		if err := qtx.DiagramShapeUpsert(ctx, sqlc.DiagramShapeUpsertParams{
			ID:        shape.ID.String(),
			DiagramID: diagramID,
			Data:      shape,
		}); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("upsert diagram shape %s: %w", shape.ID, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		_ = tx.Rollback(ctx)
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func (uc *UseCase) DiagramConnectionBatchUpsert(ctx context.Context, diagramID string, connections []domain.DiagramConnection) error {
	if len(connections) == 0 {
		return nil
	}

	tx, err := uc.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	qtx := uc.queries.WithTx(tx)

	for _, conn := range connections {
		if err := qtx.DiagramConnectionUpsert(ctx, sqlc.DiagramConnectionUpsertParams{
			DiagramID:   diagramID,
			FromShapeID: conn.FromShapeID.String(),
			ToShapeID:   conn.ToShapeID.String(),
		}); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("upsert diagram connection %s->%s: %w", conn.FromShapeID, conn.ToShapeID, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		_ = tx.Rollback(ctx)
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
