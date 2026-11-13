package domain

import (
	"errors"
	"time"
)

const (
	DefaultShapeSize = 100
)

var (
	ErrDiagramNotFound = errors.New("diagram not found")
)

type DiagramInfo struct {
	ID                DiagramID
	Title             string
	Permission        DiagramCollaboratorPermission
	Shapes            []DiagramShape
	Connections       []DiagramConnection
	ShapeCommentStats []DiagramShapeCommentStat
}

type Diagram struct {
	ID               DiagramID
	Title            string
	Permission       DiagramCollaboratorPermission
	HasCollaborators bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type DiagramShape struct {
	ID     ShapeID          `json:"id"`
	Type   DiagramShapeType `json:"tp"`  //nolint:tagliatelle // Optimization
	Left   float64          `json:"l"`   //nolint:tagliatelle // Optimization
	Top    float64          `json:"t"`   //nolint:tagliatelle // Optimization
	Width  float64          `json:"w"`   //nolint:tagliatelle // Optimization
	Height float64          `json:"h"`   //nolint:tagliatelle // Optimization
	ScaleX float64          `json:"sx"`  //nolint:tagliatelle // Optimization
	ScaleY float64          `json:"sy"`  //nolint:tagliatelle // Optimization
	Fill   string           `json:"fc"`  //nolint:tagliatelle // Optimization
	Stroke string           `json:"sc"`  //nolint:tagliatelle // Optimization
	Text   string           `json:"txt"` //nolint:tagliatelle // Optimization
	Order  string           `json:"o"`   //nolint:tagliatelle // Optimization
}

type DiagramShapeType string

func (v DiagramShapeType) String() string {
	return string(v)
}

const (
	DiagramShapeTypeSquare  DiagramShapeType = "SQUARE"
	DiagramShapeTypeEllipse DiagramShapeType = "ELLIPSE"
	DiagramShapeTypeDiamond DiagramShapeType = "DIAMOND"
	DiagramShapeTypeBox     DiagramShapeType = "BOX"
	DiagramShapeTypeText    DiagramShapeType = "TEXT"
)

type DiagramConnection struct {
	FromShapeID ShapeID `json:"fid"` //nolint:tagliatelle // Optimization
	ToShapeID   ShapeID `json:"tid"` //nolint:tagliatelle // Optimization
}

type DiagramShapeCommentStat struct {
	ShapeID      ShapeID
	CommentCount int32
}

type DiagramComment struct {
	ID             CommentID `json:"id"`
	DiagramID      DiagramID `json:"diagramId"`
	DiagramShapeID ShapeID   `json:"diagramShapeId"`
	UserID         UserID    `json:"userId"`
	UserName       string    `json:"userName"`
	Text           string    `json:"text"`
	CreatedAt      time.Time `json:"createdAt"`
}

type DiagramCollaborator struct {
	ID         CollaboratorID
	DiagramID  DiagramID
	UserID     UserID
	UserName   string
	Permission DiagramCollaboratorPermission
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type DiagramCollaboratorPermission string

func (v DiagramCollaboratorPermission) String() string {
	return string(v)
}

const (
	DiagramCollaboratorPermissionOwner DiagramCollaboratorPermission = "OWNER"
	DiagramCollaboratorPermissionView  DiagramCollaboratorPermission = "VIEW"
	DiagramCollaboratorPermissionEdit  DiagramCollaboratorPermission = "EDIT"
)
