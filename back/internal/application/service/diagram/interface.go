package diagram

import (
	"context"

	"klubok/internal/application/websocket"
	"klubok/internal/db"
	"klubok/internal/db/sqlc"
	"klubok/internal/domain"
)

type Querier interface {
	DiagramGet(ctx context.Context, id string) (sqlc.Diagram, error)
	DiagramGetByUser(ctx context.Context, userID string, iD string) (sqlc.DiagramGetByUserRow, error)
	DiagramList(ctx context.Context, userID string) ([]sqlc.DiagramListRow, error)
	DiagramSetTitle(ctx context.Context, id string, title string) error
	DiagramDelete(ctx context.Context, id string) error

	DiagramShapeUpsert(ctx context.Context, arg sqlc.DiagramShapeUpsertParams) error
	DiagramShapeDelete(ctx context.Context, iD string, diagramID string) error
	DiagramShapeList(ctx context.Context, diagramID string) ([]sqlc.DiagramShape, error)

	DiagramConnectionUpsert(ctx context.Context, arg sqlc.DiagramConnectionUpsertParams) error
	DiagramConnectionDelete(ctx context.Context, arg sqlc.DiagramConnectionDeleteParams) error
	DiagramConnectionList(ctx context.Context, diagramID string) ([]sqlc.DiagramConnection, error)

	DiagramCommentCreate(ctx context.Context, arg sqlc.DiagramCommentCreateParams) (sqlc.DiagramComment, error)
	DiagramCommentList(ctx context.Context, diagramID string, diagramShapeID string) ([]sqlc.DiagramCommentListRow, error)
	DiagramCommentStats(ctx context.Context, diagramID string) ([]sqlc.DiagramCommentStatsRow, error)
	DiagramCommentFlatList(ctx context.Context, diagramID string) ([]sqlc.DiagramComment, error)

	DiagramCollaboratorUpsert(ctx context.Context, arg sqlc.DiagramCollaboratorUpsertParams) (sqlc.DiagramCollaborator, error)
	DiagramCollaboratorList(ctx context.Context, diagramID string) ([]sqlc.DiagramCollaboratorListRow, error)
	DiagramCollaboratorDelete(ctx context.Context, diagramID string, collaboratorID string) error
	DiagramCollaboratorFlatList(ctx context.Context, diagramID string) ([]sqlc.DiagramCollaborator, error)
}

var _ Querier = (*sqlc.Queries)(nil)

type DBUseCase interface {
	DiagramCreate(ctx context.Context, id string, title string, collaborator sqlc.DiagramCollaboratorUpsertParams) (sqlc.DiagramGetByUserRow, error)
	DiagramClone(ctx context.Context, params db.DiagramCloneParams) (sqlc.DiagramGetByUserRow, error)
	DiagramShapeBatchUpsert(ctx context.Context, diagramID string, shapes []domain.DiagramShape) error
	DiagramConnectionBatchUpsert(ctx context.Context, diagramID string, connections []domain.DiagramConnection) error
}

var _ DBUseCase = (*db.UseCase)(nil)

type WebSocketManager interface {
	BroadcastToDiagram(diagramID domain.DiagramID, msg websocket.Message, notForUserID *domain.UserID) error
	GetDiagramUsers(diagramID domain.DiagramID) []domain.WebSocketUser
}

var _ WebSocketManager = (*websocket.Manager)(nil)
