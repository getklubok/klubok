package handler

import (
	"klubok/internal/restapi/oas"
)

type Handler struct {
	oas.IndexHandler
	oas.HealthHandler
	oas.FeaturesHandler
	oas.DiagramHandler
	oas.ShapeHandler
	oas.ConnectionHandler
	oas.CollaboratorHandler
	oas.CommentHandler
	oas.UserHandler
	oas.SSOHandler
	oas.EmbedHandler
	oas.EmbedPublicHandler
	oas.ImageHandler
}

func NewHandler(
	indexHandler oas.IndexHandler,
	healthHandler oas.HealthHandler,
	featuresHandler oas.FeaturesHandler,
	diagramHandler oas.DiagramHandler,
	shapeHandler oas.ShapeHandler,
	connectionHandler oas.ConnectionHandler,
	collaboratorHandler oas.CollaboratorHandler,
	commentHandler oas.CommentHandler,
	userHandler oas.UserHandler,
	ssoHandler oas.SSOHandler,
	embedHandler oas.EmbedHandler,
	embedPublicHandler oas.EmbedPublicHandler,
	imageHandler oas.ImageHandler,
) *Handler {
	return &Handler{
		IndexHandler:        indexHandler,
		HealthHandler:       healthHandler,
		FeaturesHandler:     featuresHandler,
		DiagramHandler:      diagramHandler,
		ShapeHandler:        shapeHandler,
		ConnectionHandler:   connectionHandler,
		CollaboratorHandler: collaboratorHandler,
		CommentHandler:      commentHandler,
		UserHandler:         userHandler,
		SSOHandler:          ssoHandler,
		EmbedHandler:        embedHandler,
		EmbedPublicHandler:  embedPublicHandler,
		ImageHandler:        imageHandler,
	}
}
