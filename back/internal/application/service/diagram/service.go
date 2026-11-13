package diagram

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"

	"klubok/internal/application/websocket"
	"klubok/internal/db"
	"klubok/internal/db/sqlc"
	"klubok/internal/domain"
)

type Service struct {
	querier   Querier
	dbUseCase DBUseCase
	wsManager WebSocketManager

	log *zap.Logger

	idFunc  func() string
	nowFunc func() time.Time
}

func NewService(querier Querier, dbUseCase DBUseCase, wsManager WebSocketManager, log *zap.Logger) *Service {
	return &Service{
		querier:   querier,
		dbUseCase: dbUseCase,
		wsManager: wsManager,

		log: log.Named("diagram_service"),

		idFunc:  domain.NewID,
		nowFunc: time.Now,
	}
}

func (s *Service) Create(ctx context.Context, title string, userID domain.UserID) (domain.Diagram, error) {
	id := s.idFunc()

	d, err := s.dbUseCase.DiagramCreate(ctx, id, title, sqlc.DiagramCollaboratorUpsertParams{
		ID:         s.idFunc(),
		DiagramID:  id,
		UserID:     userID.String(),
		Permission: domain.DiagramCollaboratorPermissionOwner,
	})
	if err != nil {
		return domain.Diagram{}, fmt.Errorf("create diagram: %w", err)
	}

	return diagramToDomain(d), nil
}

func (s *Service) List(ctx context.Context, userID domain.UserID) ([]domain.Diagram, error) {
	ds, err := s.querier.DiagramList(ctx, userID.String())
	if err != nil {
		return nil, fmt.Errorf("find diagrams: %w", err)
	}

	return diagramListRowsToDomains(ds), nil
}

func (s *Service) Delete(ctx context.Context, diagramID domain.DiagramID) error {
	err := s.querier.DiagramDelete(ctx, diagramID.String())
	if err != nil {
		return fmt.Errorf("delete diagram: %w", err)
	}

	return nil
}

func (s *Service) SetTitle(ctx context.Context, id domain.DiagramID, title string) error {
	err := s.querier.DiagramSetTitle(ctx, id.String(), title)
	if err != nil {
		return fmt.Errorf("set title diagram: %w", err)
	}

	return nil
}

func (s *Service) Info(ctx context.Context, diagramID domain.DiagramID) (domain.DiagramInfo, error) {
	diagram, err := s.querier.DiagramGet(ctx, diagramID.String())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.DiagramInfo{}, domain.ErrDiagramNotFound
		}

		return domain.DiagramInfo{}, fmt.Errorf("get diagram: %w", err)
	}

	shConCom, err := s.getShConCom(ctx, diagramID)
	if err != nil {
		return domain.DiagramInfo{}, fmt.Errorf("get shapes, connections and comments: %w", err)
	}

	return domain.DiagramInfo{
		ID:                diagramID,
		Title:             diagram.Title,
		Permission:        domain.DiagramCollaboratorPermissionView,
		Shapes:            shConCom.shapes,
		Connections:       shConCom.connections,
		ShapeCommentStats: shConCom.shapeCommentStats,
	}, nil
}

func (s *Service) InfoByUser(ctx context.Context, diagramID domain.DiagramID, userID domain.UserID) (domain.DiagramInfo, error) {
	diagram, err := s.querier.DiagramGetByUser(ctx, userID.String(), diagramID.String())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.DiagramInfo{}, domain.ErrDiagramNotFound
		}

		return domain.DiagramInfo{}, fmt.Errorf("get diagram: %w", err)
	}

	shConCom, err := s.getShConCom(ctx, diagramID)
	if err != nil {
		return domain.DiagramInfo{}, fmt.Errorf("get shapes, connections and comments: %w", err)
	}

	return domain.DiagramInfo{
		ID:                diagramID,
		Title:             diagram.Title,
		Permission:        diagram.Permission,
		Shapes:            shConCom.shapes,
		Connections:       shConCom.connections,
		ShapeCommentStats: shConCom.shapeCommentStats,
	}, nil
}

func (s *Service) ShapeUpsert(ctx context.Context, userID domain.UserID, diagramID domain.DiagramID, shape domain.DiagramShape) error {
	if err := s.querier.DiagramShapeUpsert(ctx, sqlc.DiagramShapeUpsertParams{
		ID:        shape.ID.String(),
		DiagramID: diagramID.String(),
		Data:      shape,
	}); err != nil {
		return fmt.Errorf("upsert diagram shape: %w", err)
	}

	msg := websocket.Message{
		Type:      domain.WebSocketMessageTypeShapeUpserted,
		DiagramID: diagramID,
		Timestamp: s.nowFunc().Unix(),
		Data:      shape,
	}
	if err := s.wsManager.BroadcastToDiagram(diagramID, msg, &userID); err != nil {
		s.log.Warn("Broadcast to diagram",
			zap.String("diagramID", diagramID.String()),
			zap.String("userID", userID.String()),
			zap.Error(err),
		)
	}

	return nil
}

func (s *Service) ShapeBatchUpsert( //nolint:dupl // It`s ok
	ctx context.Context,
	userID domain.UserID,
	diagramID domain.DiagramID,
	shapes []domain.DiagramShape,
) error {
	if len(shapes) == 0 {
		return nil
	}

	if err := s.dbUseCase.DiagramShapeBatchUpsert(ctx, diagramID.String(), shapes); err != nil {
		return fmt.Errorf("batch upsert shapes: %w", err)
	}

	// Broadcast each shape change
	for _, shape := range shapes {
		msg := websocket.Message{
			Type:      domain.WebSocketMessageTypeShapeUpserted,
			DiagramID: diagramID,
			Timestamp: s.nowFunc().Unix(),
			Data:      shape,
		}
		if err := s.wsManager.BroadcastToDiagram(diagramID, msg, &userID); err != nil {
			s.log.Warn("Broadcast shape upsert",
				zap.String("diagramID", diagramID.String()),
				zap.String("userID", userID.String()),
				zap.Error(err),
			)
		}
	}

	return nil
}

func (s *Service) ShapeDelete(ctx context.Context, userID domain.UserID, diagramID domain.DiagramID, shapeID domain.ShapeID) error {
	if err := s.querier.DiagramShapeDelete(ctx, shapeID.String(), diagramID.String()); err != nil {
		return fmt.Errorf("delete diagram shape: %w", err)
	}

	msg := websocket.Message{
		Type:      domain.WebSocketMessageTypeShapeDeleted,
		DiagramID: diagramID,
		Timestamp: s.nowFunc().Unix(),
		Data:      shapeID,
	}
	if err := s.wsManager.BroadcastToDiagram(diagramID, msg, &userID); err != nil {
		s.log.Warn("Broadcast to diagram",
			zap.String("diagramID", diagramID.String()),
			zap.String("userID", userID.String()),
			zap.Error(err),
		)
	}

	return nil
}

//nolint:dupl // similar to ConnectionDelete
func (s *Service) ConnectionUpsert(ctx context.Context, userID domain.UserID, diagramID domain.DiagramID, connection domain.DiagramConnection) error {
	if err := s.querier.DiagramConnectionUpsert(ctx, sqlc.DiagramConnectionUpsertParams{
		DiagramID:   diagramID.String(),
		FromShapeID: connection.FromShapeID.String(),
		ToShapeID:   connection.ToShapeID.String(),
	}); err != nil {
		return fmt.Errorf("upsert diagram connection: %w", err)
	}

	msg := websocket.Message{
		Type:      domain.WebSocketMessageTypeConnectionUpserted,
		DiagramID: diagramID,
		Timestamp: s.nowFunc().Unix(),
		Data:      connection,
	}
	if err := s.wsManager.BroadcastToDiagram(diagramID, msg, &userID); err != nil {
		s.log.Warn("Broadcast to diagram",
			zap.String("diagramID", diagramID.String()),
			zap.String("userID", userID.String()),
			zap.Error(err),
		)
	}

	return nil
}

func (s *Service) ConnectionBatchUpsert( //nolint:dupl // It`s ok
	ctx context.Context,
	userID domain.UserID,
	diagramID domain.DiagramID,
	connections []domain.DiagramConnection,
) error {
	if len(connections) == 0 {
		return nil
	}

	if err := s.dbUseCase.DiagramConnectionBatchUpsert(ctx, diagramID.String(), connections); err != nil {
		return fmt.Errorf("batch upsert connections: %w", err)
	}

	// Broadcast each connection change
	for _, conn := range connections {
		msg := websocket.Message{
			Type:      domain.WebSocketMessageTypeConnectionUpserted,
			DiagramID: diagramID,
			Timestamp: s.nowFunc().Unix(),
			Data:      conn,
		}
		if err := s.wsManager.BroadcastToDiagram(diagramID, msg, &userID); err != nil {
			s.log.Warn("Broadcast connection upsert",
				zap.String("diagramID", diagramID.String()),
				zap.String("userID", userID.String()),
				zap.Error(err),
			)
		}
	}

	return nil
}

//nolint:dupl // similar to ConnectionUpsert
func (s *Service) ConnectionDelete(ctx context.Context, userID domain.UserID, diagramID domain.DiagramID, connection domain.DiagramConnection) error {
	if err := s.querier.DiagramConnectionDelete(ctx, sqlc.DiagramConnectionDeleteParams{
		DiagramID:   diagramID.String(),
		FromShapeID: connection.FromShapeID.String(),
		ToShapeID:   connection.ToShapeID.String(),
	}); err != nil {
		return fmt.Errorf("delete diagram connection: %w", err)
	}

	msg := websocket.Message{
		Type:      domain.WebSocketMessageTypeConnectionDeleted,
		DiagramID: diagramID,
		Timestamp: s.nowFunc().Unix(),
		Data:      connection,
	}
	if err := s.wsManager.BroadcastToDiagram(diagramID, msg, &userID); err != nil {
		s.log.Warn("Broadcast to diagram",
			zap.String("diagramID", diagramID.String()),
			zap.String("userID", userID.String()),
			zap.Error(err),
		)
	}

	return nil
}

func (s *Service) CommentCreate(ctx context.Context, userID domain.UserID, comment domain.DiagramComment) (domain.DiagramComment, error) {
	res, err := s.querier.DiagramCommentCreate(ctx, sqlc.DiagramCommentCreateParams{
		ID:             s.idFunc(),
		DiagramID:      comment.DiagramID.String(),
		DiagramShapeID: comment.DiagramShapeID.String(),
		UserID:         comment.UserID.String(),
		Text:           comment.Text,
	})
	if err != nil {
		return domain.DiagramComment{}, fmt.Errorf("create diagram comment: %w", err)
	}

	msg := websocket.Message{
		Type:      domain.WebSocketMessageTypeCommentAdded,
		DiagramID: comment.DiagramID,
		Timestamp: s.nowFunc().Unix(),
		Data:      comment.DiagramShapeID,
	}
	if err := s.wsManager.BroadcastToDiagram(comment.DiagramID, msg, &userID); err != nil {
		s.log.Warn("Broadcast to diagram",
			zap.String("diagramID", comment.DiagramID.String()),
			zap.String("userID", userID.String()),
			zap.Error(err),
		)
	}

	return commentToDomain(res), nil
}

func (s *Service) CommentList(ctx context.Context, diagramID domain.DiagramID, diagramShapeID domain.ShapeID) ([]domain.DiagramComment, error) {
	res, err := s.querier.DiagramCommentList(ctx, diagramID.String(), diagramShapeID.String())
	if err != nil {
		return nil, fmt.Errorf("get diagram comments: %w", err)
	}

	return commentListRowsToDomains(res), nil
}

func (s *Service) CollaboratorUpsert(ctx context.Context, collaborator domain.DiagramCollaborator) (domain.DiagramCollaborator, error) {
	res, err := s.querier.DiagramCollaboratorUpsert(ctx, sqlc.DiagramCollaboratorUpsertParams{
		ID:         s.idFunc(),
		DiagramID:  collaborator.DiagramID.String(),
		UserID:     collaborator.UserID.String(),
		Permission: collaborator.Permission,
	})
	if err != nil {
		return domain.DiagramCollaborator{}, fmt.Errorf("upsert diagram collaborator: %w", err)
	}

	return collaboratorToDomain(res), nil
}

func (s *Service) CollaboratorList(ctx context.Context, diagramID domain.DiagramID) ([]domain.DiagramCollaborator, error) {
	res, err := s.querier.DiagramCollaboratorList(ctx, diagramID.String())
	if err != nil {
		return nil, fmt.Errorf("get diagram collaborators: %w", err)
	}

	return collaboratorListRowsToDomains(res), nil
}

func (s *Service) CollaboratorDelete(ctx context.Context, diagramID domain.DiagramID, collaboratorID domain.CollaboratorID) error {
	err := s.querier.DiagramCollaboratorDelete(ctx, diagramID.String(), collaboratorID.String())
	if err != nil {
		return fmt.Errorf("delete diagram collaborator: %w", err)
	}

	return nil
}

func (s *Service) OnlineUsers(_ context.Context, diagramID domain.DiagramID) []domain.WebSocketUser {
	return s.wsManager.GetDiagramUsers(diagramID)
}

func (s *Service) Clone(ctx context.Context, diagramID domain.DiagramID, userID domain.UserID) (domain.Diagram, error) {
	src, err := s.InfoByUser(ctx, diagramID, userID)
	if err != nil {
		return domain.Diagram{}, fmt.Errorf("get source diagram info: %w", err)
	}

	dbComments, err := s.querier.DiagramCommentFlatList(ctx, diagramID.String())
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return domain.Diagram{}, fmt.Errorf("get diagram comments: %w", err)
	}

	dbCollaborators, err := s.querier.DiagramCollaboratorFlatList(ctx, diagramID.String())
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return domain.Diagram{}, fmt.Errorf("get diagram collaborators: %w", err)
	}

	shapeIDMap := make(map[domain.ShapeID]string, len(src.Shapes))
	newShapes := make([]db.DiagramShapeClone, 0, len(src.Shapes))
	for _, shape := range src.Shapes {
		newID := s.idFunc()
		shapeIDMap[shape.ID] = newID
		newShapes = append(newShapes, db.DiagramShapeClone{
			ID: newID,
			Data: domain.DiagramShape{
				ID:     domain.ShapeID(newID),
				Type:   shape.Type,
				Left:   shape.Left,
				Top:    shape.Top,
				Width:  shape.Width,
				Height: shape.Height,
				ScaleX: shape.ScaleX,
				ScaleY: shape.ScaleY,
				Fill:   shape.Fill,
				Stroke: shape.Stroke,
				Text:   shape.Text,
				Order:  shape.Order,
			},
		})
	}

	newConnections := make([]db.DiagramConnectionClone, 0, len(src.Connections))
	for _, conn := range src.Connections {
		newConnections = append(newConnections, db.DiagramConnectionClone{
			FromShapeID: shapeIDMap[conn.FromShapeID],
			ToShapeID:   shapeIDMap[conn.ToShapeID],
		})
	}

	newComments := make([]db.DiagramCommentClone, 0, len(dbComments))
	for _, c := range dbComments {
		newComments = append(newComments, db.DiagramCommentClone{
			ID:             s.idFunc(),
			DiagramShapeID: shapeIDMap[domain.ShapeID(c.DiagramShapeID)],
			UserID:         c.UserID,
			Text:           c.Text,
		})
	}

	newCollaborators := make([]db.DiagramCollaboratorClone, 0, len(dbCollaborators))
	for _, c := range dbCollaborators {
		if c.Permission == domain.DiagramCollaboratorPermissionOwner {
			continue
		}
		if c.UserID == userID.String() {
			continue
		}
		newCollaborators = append(newCollaborators, db.DiagramCollaboratorClone{
			ID:         s.idFunc(),
			UserID:     c.UserID,
			Permission: c.Permission,
		})
	}

	newDiagramID := s.idFunc()
	newTitle := src.Title + " (copy)"

	result, err := s.dbUseCase.DiagramClone(ctx, db.DiagramCloneParams{
		NewDiagramID:  newDiagramID,
		NewTitle:      newTitle,
		OwnerUserID:   userID.String(),
		Shapes:        newShapes,
		Connections:   newConnections,
		Comments:      newComments,
		Collaborators: newCollaborators,
	})
	if err != nil {
		return domain.Diagram{}, fmt.Errorf("clone diagram: %w", err)
	}

	return diagramToDomain(result), nil
}

func (s *Service) Import(ctx context.Context, userID domain.UserID, diagramInfo domain.DiagramInfo) (domain.Diagram, error) {
	shapeIDMap := make(map[domain.ShapeID]string, len(diagramInfo.Shapes))
	newShapes := make([]db.DiagramShapeClone, 0, len(diagramInfo.Shapes))
	for _, shape := range diagramInfo.Shapes {
		newID := s.idFunc()
		shapeIDMap[shape.ID] = newID
		newShapes = append(newShapes, db.DiagramShapeClone{
			ID: newID,
			Data: domain.DiagramShape{
				ID:     domain.ShapeID(newID),
				Type:   shape.Type,
				Left:   shape.Left,
				Top:    shape.Top,
				Width:  shape.Width,
				Height: shape.Height,
				ScaleX: shape.ScaleX,
				ScaleY: shape.ScaleY,
				Fill:   shape.Fill,
				Stroke: shape.Stroke,
				Text:   shape.Text,
				Order:  shape.Order,
			},
		})
	}

	newConnections := make([]db.DiagramConnectionClone, 0, len(diagramInfo.Connections))
	for _, conn := range diagramInfo.Connections {
		newConnections = append(newConnections, db.DiagramConnectionClone{
			FromShapeID: shapeIDMap[conn.FromShapeID],
			ToShapeID:   shapeIDMap[conn.ToShapeID],
		})
	}

	newDiagramID := s.idFunc()
	newTitle := fmt.Sprintf("%s (imported %s)",
		diagramInfo.Title,
		s.nowFunc().Format("2006-01-02 15:04"),
	)

	result, err := s.dbUseCase.DiagramClone(ctx, db.DiagramCloneParams{
		NewDiagramID: newDiagramID,
		NewTitle:     newTitle,
		OwnerUserID:  userID.String(),
		Shapes:       newShapes,
		Connections:  newConnections,
	})
	if err != nil {
		return domain.Diagram{}, fmt.Errorf("clone diagram: %w", err)
	}

	return diagramToDomain(result), nil
}

type shConCom struct {
	shapes            []domain.DiagramShape
	connections       []domain.DiagramConnection
	shapeCommentStats []domain.DiagramShapeCommentStat
}

func (s *Service) getShConCom(ctx context.Context, diagramID domain.DiagramID) (shConCom, error) {
	dbShapes, err := s.querier.DiagramShapeList(ctx, diagramID.String())
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return shConCom{}, fmt.Errorf("get diagram shapes: %w", err)
	}
	shapes := make([]domain.DiagramShape, 0, len(dbShapes))
	for _, dbShape := range dbShapes {
		shapes = append(shapes, dbShape.Data)
	}

	dbConnections, err := s.querier.DiagramConnectionList(ctx, diagramID.String())
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return shConCom{}, fmt.Errorf("get diagram connections: %w", err)
	}
	connections := make([]domain.DiagramConnection, 0, len(dbConnections))
	for _, dbConnection := range dbConnections {
		connections = append(connections, domain.DiagramConnection{
			FromShapeID: domain.ShapeID(dbConnection.FromShapeID),
			ToShapeID:   domain.ShapeID(dbConnection.ToShapeID),
		})
	}

	shapeCommentStats, err := s.querier.DiagramCommentStats(ctx, diagramID.String())
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return shConCom{}, fmt.Errorf("get diagram comment shapes: %w", err)
	}

	return shConCom{
		shapes:            shapes,
		connections:       connections,
		shapeCommentStats: commentStatRowsToDomains(shapeCommentStats),
	}, nil
}
