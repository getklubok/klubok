package db

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"klubok/internal/db/sqlc"
	"klubok/internal/domain"
)

func TestShapeImageID(t *testing.T) {
	tests := []struct {
		name  string
		shape sqlc.DiagramShape
		want  string
	}{
		{
			name:  "Image shape with image",
			shape: sqlc.DiagramShape{Data: domain.DiagramShape{Type: domain.DiagramShapeTypeImage, Img: "image-1"}},
			want:  "image-1",
		},
		{
			name:  "Image shape without image",
			shape: sqlc.DiagramShape{Data: domain.DiagramShape{Type: domain.DiagramShapeTypeImage}},
			want:  "",
		},
		{
			name:  "Square shape ignores img",
			shape: sqlc.DiagramShape{Data: domain.DiagramShape{Type: domain.DiagramShapeTypeSquare, Img: "image-1"}},
			want:  "",
		},
		{
			name:  "Text shape",
			shape: sqlc.DiagramShape{Data: domain.DiagramShape{Type: domain.DiagramShapeTypeText}},
			want:  "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, shapeImageID(tt.shape))
		})
	}
}

func TestImageBlobFromShape(t *testing.T) {
	row := sqlc.ImageDeleteForShapeRow{BlobHash: "hash-1", Mime: "image/png"}

	assert.Equal(t, ImageBlob{Hash: "hash-1", Mime: "image/png"}, imageBlobFromShape(row))
}

func TestImageBlobsFromDiagram(t *testing.T) {
	tests := []struct {
		name string
		rows []sqlc.ImageDeleteForDiagramRow
		want []ImageBlob
	}{
		{
			name: "No images",
			rows: nil,
			want: []ImageBlob{},
		},
		{
			name: "Multiple images",
			rows: []sqlc.ImageDeleteForDiagramRow{
				{BlobHash: "hash-1", Mime: "image/png"},
				{BlobHash: "hash-2", Mime: "image/jpeg"},
			},
			want: []ImageBlob{
				{Hash: "hash-1", Mime: "image/png"},
				{Hash: "hash-2", Mime: "image/jpeg"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, imageBlobsFromDiagram(tt.rows))
		})
	}
}

func TestCloneImageRows(t *testing.T) {
	now := time.Date(2024, 12, 18, 14, 0, 0, 0, time.UTC)
	preview := []byte("preview")

	tests := []struct {
		name     string
		images   []ImageClone
		wantRows []sqlc.ImageBulkInsertParams
		wantIDs  map[string]string
	}{
		{
			name:     "No images",
			images:   nil,
			wantRows: []sqlc.ImageBulkInsertParams{},
			wantIDs:  map[string]string{},
		},
		{
			name: "Images get new ids and keep blob hash",
			images: []ImageClone{
				{
					ID:           "image-1",
					BlobHash:     "hash-1",
					Preview:      preview,
					OriginalName: "photo.png",
					Mime:         "image/png",
				},
				{
					ID:           "image-2",
					BlobHash:     "hash-1",
					Preview:      preview,
					OriginalName: "photo.jpg",
					Mime:         "image/jpeg",
				},
			},
			wantRows: []sqlc.ImageBulkInsertParams{
				{
					ID:           "new-image-1",
					DiagramID:    "new-diagram",
					BlobHash:     "hash-1",
					Preview:      preview,
					OriginalName: "photo.png",
					Mime:         "image/png",
					CreatedAt:    now,
				},
				{
					ID:           "new-image-2",
					DiagramID:    "new-diagram",
					BlobHash:     "hash-1",
					Preview:      preview,
					OriginalName: "photo.jpg",
					Mime:         "image/jpeg",
					CreatedAt:    now,
				},
			},
			wantIDs: map[string]string{
				"image-1": "new-image-1",
				"image-2": "new-image-2",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ids := []string{"new-image-1", "new-image-2", "new-image-3"}
			calls := 0

			newID := func() string {
				require.Less(t, calls, len(ids), "unexpected id call")

				id := ids[calls]
				calls++

				return id
			}

			rows, imageIDs := cloneImageRows("new-diagram", tt.images, newID, now)
			assert.Equal(t, tt.wantRows, rows)
			assert.Equal(t, tt.wantIDs, imageIDs)
		})
	}
}

func TestCloneShapeData(t *testing.T) {
	imageIDs := map[string]string{"image-1": "new-image-1"}

	tests := []struct {
		name  string
		shape domain.DiagramShape
		want  domain.DiagramShape
	}{
		{
			name:  "Image reference is remapped and preview is dropped",
			shape: domain.DiagramShape{ID: "shape-1", Type: domain.DiagramShapeTypeImage, Img: "image-1", Pv: "data:image/jpeg;base64,preview"},
			want:  domain.DiagramShape{ID: "shape-1", Type: domain.DiagramShapeTypeImage, Img: "new-image-1"},
		},
		{
			name:  "Unknown image reference is kept",
			shape: domain.DiagramShape{ID: "shape-2", Type: domain.DiagramShapeTypeImage, Img: "image-2"},
			want:  domain.DiagramShape{ID: "shape-2", Type: domain.DiagramShapeTypeImage, Img: "image-2"},
		},
		{
			name:  "Other shape types are copied without preview",
			shape: domain.DiagramShape{ID: "shape-3", Type: domain.DiagramShapeTypeSquare, Pv: "data:image/jpeg;base64,preview"},
			want:  domain.DiagramShape{ID: "shape-3", Type: domain.DiagramShapeTypeSquare},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, cloneShapeData(tt.shape, imageIDs))
		})
	}
}
