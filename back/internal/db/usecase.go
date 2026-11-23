package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
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
	Images        []ImageClone
}

// ImageClone описывает строку images, которую нужно скопировать при клонировании.
type ImageClone struct {
	ID           string
	BlobHash     string
	Preview      []byte
	OriginalName string
	Mime         string
}

// ImageBlob ссылается на блоб, который освободился после удаления строки images.
type ImageBlob struct {
	Hash string
	Mime string
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

	imageRows, imageIDs := cloneImageRows(params.NewDiagramID, params.Images, uc.idFunc, now)
	if len(imageRows) > 0 {
		if _, err := qtx.ImageBulkInsert(ctx, imageRows); err != nil {
			_ = tx.Rollback(ctx)
			return sqlc.DiagramGetByUserRow{}, fmt.Errorf("bulk insert images: %w", err)
		}
	}

	if len(params.Shapes) > 0 {
		shapeRows := make([]sqlc.DiagramShapeBulkInsertParams, 0, len(params.Shapes))
		for _, s := range params.Shapes {
			shapeRows = append(shapeRows, sqlc.DiagramShapeBulkInsertParams{
				ID:        s.ID,
				DiagramID: params.NewDiagramID,
				Data:      cloneShapeData(s.Data, imageIDs),
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

func (uc *UseCase) DiagramShapeGet(ctx context.Context, diagramID string, shapeID string) (sqlc.DiagramShape, error) {
	shape, err := uc.queries.DiagramShapeGet(ctx, shapeID, diagramID)
	if err != nil {
		return sqlc.DiagramShape{}, fmt.Errorf("get diagram shape: %w", err)
	}

	return shape, nil
}

// DiagramShapeDeleteWithImage удаляет шейп и связанную строку images в одной транзакции.
func (uc *UseCase) DiagramShapeDeleteWithImage(ctx context.Context, diagramID string, shapeID string) ([]ImageBlob, error) {
	tx, err := uc.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}

	qtx := uc.queries.WithTx(tx)

	shape, err := qtx.DiagramShapeGet(ctx, shapeID, diagramID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		_ = tx.Rollback(ctx)

		return nil, fmt.Errorf("get diagram shape: %w", err)
	}

	if err := qtx.DiagramShapeDelete(ctx, shapeID, diagramID); err != nil {
		_ = tx.Rollback(ctx)

		return nil, fmt.Errorf("delete diagram shape: %w", err)
	}

	var blobs []ImageBlob

	if imageID := shapeImageID(shape); imageID != "" {
		row, err := qtx.ImageDeleteForShape(ctx, imageID, diagramID)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			_ = tx.Rollback(ctx)

			return nil, fmt.Errorf("delete shape image: %w", err)
		}

		if err == nil {
			blobs = append(blobs, imageBlobFromShape(row))
		}
	}

	if err := tx.Commit(ctx); err != nil {
		_ = tx.Rollback(ctx)

		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return blobs, nil
}

// DiagramDeleteWithImages удаляет строки images и диаграмму в одной транзакции.
func (uc *UseCase) DiagramDeleteWithImages(ctx context.Context, diagramID string) ([]ImageBlob, error) {
	tx, err := uc.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}

	qtx := uc.queries.WithTx(tx)

	imageRows, err := qtx.ImageDeleteForDiagram(ctx, diagramID)
	if err != nil {
		_ = tx.Rollback(ctx)

		return nil, fmt.Errorf("delete diagram images: %w", err)
	}

	if err := qtx.DiagramDelete(ctx, diagramID); err != nil {
		_ = tx.Rollback(ctx)

		return nil, fmt.Errorf("delete diagram: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		_ = tx.Rollback(ctx)

		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return imageBlobsFromDiagram(imageRows), nil
}

func cloneImageRows(
	diagramID string,
	images []ImageClone,
	newID func() string,
	now time.Time,
) ([]sqlc.ImageBulkInsertParams, map[string]string) {
	rows := make([]sqlc.ImageBulkInsertParams, 0, len(images))
	imageIDs := make(map[string]string, len(images))

	for _, image := range images {
		id := newID()
		imageIDs[image.ID] = id

		rows = append(rows, sqlc.ImageBulkInsertParams{
			ID:           id,
			DiagramID:    diagramID,
			BlobHash:     image.BlobHash,
			Preview:      image.Preview,
			OriginalName: image.OriginalName,
			Mime:         image.Mime,
			CreatedAt:    now,
		})
	}

	return rows, imageIDs
}

func cloneShapeData(shape domain.DiagramShape, imageIDs map[string]string) domain.DiagramShape {
	data := shape
	data.Pv = ""

	if imageID, ok := imageIDs[data.Img]; ok {
		data.Img = imageID
	}

	return data
}

func shapeImageID(shape sqlc.DiagramShape) string {
	if shape.Data.Type != domain.DiagramShapeTypeImage {
		return ""
	}

	return shape.Data.Img
}

func imageBlobFromShape(row sqlc.ImageDeleteForShapeRow) ImageBlob {
	return ImageBlob{Hash: row.BlobHash, Mime: row.Mime}
}

func imageBlobsFromDiagram(rows []sqlc.ImageDeleteForDiagramRow) []ImageBlob {
	blobs := make([]ImageBlob, 0, len(rows))
	for _, row := range rows {
		blobs = append(blobs, ImageBlob{Hash: row.BlobHash, Mime: row.Mime})
	}

	return blobs
}
