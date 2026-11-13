package domain

import (
	"errors"
	"time"
)

const EmbedTokenLength = 64

var ErrEmbedNotFound = errors.New("embed not found")

type DiagramEmbedInfo struct {
	ID           EmbedID
	DiagramID    DiagramID
	EmbedCode    string
	IsActive     bool
	ViewsCount   int32
	LastViewedAt *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
