package image

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"

	"klubok/internal/domain"
	"klubok/internal/provider/storage"
)

// Variant — вариант выдачи изображения.
type Variant string

const (
	// VariantFull — полноэкранный вариант.
	VariantFull Variant = "full"
	// VariantDownload — оригинал для скачивания.
	VariantDownload Variant = "download"
)

// ErrNotFound возвращается, когда строки images нет или блоб отсутствует в хранилище.
var ErrNotFound = errors.New("image not found")

// File описывает открытый вариант изображения.
type File struct {
	Reader       io.ReadCloser
	Size         int64
	Mime         string
	OriginalName string
}

// ParseVariant разбирает вариант выдачи из параметра маршрута.
func ParseVariant(v string) (Variant, error) {
	switch Variant(v) {
	case VariantFull:
		return VariantFull, nil
	case VariantDownload:
		return VariantDownload, nil
	default:
		return "", fmt.Errorf("unknown image variant %q", v)
	}
}

// Open открывает строку images диаграммы и соответствующий ей блоб в хранилище.
func (s *Service) Open(
	ctx context.Context,
	diagramID domain.DiagramID,
	imageID domain.ImageID,
	variant Variant,
) (File, error) {
	row, err := s.querier.ImageGet(ctx, imageID.String(), diagramID.String())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return File{}, ErrNotFound
		}

		return File{}, fmt.Errorf("get image: %w", err)
	}

	ext, err := extForMime(row.Mime)
	if err != nil {
		return File{}, err
	}

	key := storage.OriginalKey(row.BlobHash, ext)
	if variant == VariantFull {
		key = storage.FullKey(row.BlobHash, ext)
	}

	reader, size, err := s.storage.Open(ctx, key)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return File{}, ErrNotFound
		}

		return File{}, fmt.Errorf("open storage object: %w", err)
	}

	return File{
		Reader:       reader,
		Size:         size,
		Mime:         row.Mime,
		OriginalName: row.OriginalName,
	}, nil
}
