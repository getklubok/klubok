package image

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg" // регистрация JPEG-декодера
	_ "image/png"  // регистрация PNG-декодера
	"io"
	"net/http"
	"os"

	"klubok/internal/db/sqlc"
	"klubok/internal/domain"
	"klubok/internal/provider/storage"
)

const (
	copyBufferSize       = 64 * 1024
	detectHeaderSize     = 512
	previewDataURLPrefix = "data:image/jpeg;base64,"

	mimePNG  = "image/png"
	mimeJPEG = "image/jpeg"

	formatPNG  = "png"
	formatJPEG = "jpeg"

	extPNG  = "png"
	extJPEG = "jpg"
)

var (
	// ErrFileTooLarge возвращается, когда файл превышает IMAGE_MAX_FILE_SIZE.
	ErrFileTooLarge = errors.New("image file too large")
	// ErrInvalidImage возвращается для неподдерживаемого, повреждённого или слишком большого изображения.
	ErrInvalidImage = errors.New("invalid image")
)

// UploadResult описывает успешно загруженное изображение.
type UploadResult struct {
	ID           string
	OriginalName string
	Mime         string
	Width        int
	Height       int
	Pv           string
}

// Upload проверяет изображение, сохраняет оригинал и производные и создаёт строку images.
func (s *Service) Upload(
	ctx context.Context,
	diagramID domain.DiagramID,
	originalName string,
	body io.Reader,
) (UploadResult, error) {
	tmp, err := os.CreateTemp("", "klubok-image-*")
	if err != nil {
		return UploadResult{}, fmt.Errorf("create temp file: %w", err)
	}

	tmpName := tmp.Name()
	defer func() {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
	}()

	hash, size, err := writeUpload(tmp, body, s.cfg.MaxFileSizeBytes)
	if err != nil {
		return UploadResult{}, err
	}

	head, err := readHead(tmp)
	if err != nil {
		return UploadResult{}, err
	}

	format, ok := detectedFormat(http.DetectContentType(head))
	if !ok {
		return UploadResult{}, fmt.Errorf("%w: unsupported content type", ErrInvalidImage)
	}

	width, height, err := decodeSize(tmp, s.cfg.MaxPixels)
	if err != nil {
		return UploadResult{}, err
	}

	img, err := decodeImage(tmp)
	if err != nil {
		return UploadResult{}, err
	}

	preview, err := previewJPEG(img, s.cfg.PreviewMaxSidePx)
	if err != nil {
		return UploadResult{}, err
	}

	fullscreen, resized, err := fullscreenVariant(img, format, s.cfg.FullscreenMaxSidePx)
	if err != nil {
		return UploadResult{}, err
	}

	ext := extForFormat(format)
	mime := mimeForFormat(format)
	name := sanitizeOriginalName(originalName, ext)

	id, err := s.save(ctx, diagramID, uploadData{
		file:              tmp,
		size:              size,
		hash:              hash,
		ext:               ext,
		mime:              mime,
		originalName:      name,
		preview:           preview,
		fullscreen:        fullscreen,
		fullscreenResized: resized,
	})
	if err != nil {
		return UploadResult{}, err
	}

	return UploadResult{
		ID:           id,
		OriginalName: name,
		Mime:         mime,
		Width:        width,
		Height:       height,
		Pv:           previewDataURLPrefix + base64.StdEncoding.EncodeToString(preview),
	}, nil
}

type uploadData struct {
	file              *os.File
	size              int64
	hash              string
	ext               string
	mime              string
	originalName      string
	preview           []byte
	fullscreen        []byte
	fullscreenResized bool
}

func (s *Service) save(ctx context.Context, diagramID domain.DiagramID, in uploadData) (string, error) {
	id := s.idFunc()

	release := s.locker.lock(in.hash)
	defer release()

	if err := putFromFile(ctx, s.storage, storage.OriginalKey(in.hash, in.ext), in.file, in.size, in.mime); err != nil {
		return "", fmt.Errorf("put original image: %w", err)
	}

	if err := s.putFullscreen(ctx, in); err != nil {
		return "", err
	}

	if _, err := s.querier.ImageCreate(ctx, sqlc.ImageCreateParams{
		ID:           id,
		DiagramID:    diagramID.String(),
		BlobHash:     in.hash,
		Preview:      in.preview,
		OriginalName: in.originalName,
		Mime:         in.mime,
	}); err != nil {
		s.cleaner.enqueue([]Blob{{Hash: in.hash, Mime: in.mime}})

		return "", fmt.Errorf("create image: %w", err)
	}

	return id, nil
}

func (s *Service) putFullscreen(ctx context.Context, in uploadData) error {
	var err error
	if in.fullscreenResized {
		err = s.storage.Put(
			ctx,
			storage.FullKey(in.hash, in.ext),
			bytes.NewReader(in.fullscreen),
			int64(len(in.fullscreen)),
			in.mime,
		)
	} else {
		err = putFromFile(ctx, s.storage, storage.FullKey(in.hash, in.ext), in.file, in.size, in.mime)
	}

	if err != nil {
		s.cleaner.enqueue([]Blob{{Hash: in.hash, Mime: in.mime}})

		return fmt.Errorf("put fullscreen image: %w", err)
	}

	return nil
}

func writeUpload(dst io.Writer, body io.Reader, maxSize int64) (string, int64, error) {
	// nil ResponseWriter: сервис не управляет HTTP-ответом и только ограничивает чтение.
	limited := http.MaxBytesReader(nil, io.NopCloser(body), maxSize)

	hasher := sha256.New()

	size, err := io.CopyBuffer(io.MultiWriter(dst, hasher), limited, make([]byte, copyBufferSize))
	if err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			return "", 0, fmt.Errorf("%w: limit %d bytes", ErrFileTooLarge, maxSize)
		}

		return "", 0, fmt.Errorf("copy upload: %w", err)
	}

	return hex.EncodeToString(hasher.Sum(nil)), size, nil
}

func readHead(f *os.File) ([]byte, error) {
	head := make([]byte, detectHeaderSize)

	n, err := f.ReadAt(head, 0)
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("read upload head: %w", err)
	}

	return head[:n], nil
}

func decodeSize(f *os.File, maxPixels int) (width, height int, err error) {
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return 0, 0, fmt.Errorf("rewind upload file: %w", err)
	}

	cfg, _, err := image.DecodeConfig(f)
	if err != nil {
		return 0, 0, fmt.Errorf("%w: decode image config: %w", ErrInvalidImage, err)
	}

	if cfg.Width <= 0 || cfg.Height <= 0 {
		return 0, 0, fmt.Errorf("%w: invalid image dimensions %dx%d", ErrInvalidImage, cfg.Width, cfg.Height)
	}

	if int64(cfg.Width)*int64(cfg.Height) > int64(maxPixels) {
		return 0, 0, fmt.Errorf("%w: %d pixels exceed limit %d", ErrInvalidImage, int64(cfg.Width)*int64(cfg.Height), maxPixels)
	}

	return cfg.Width, cfg.Height, nil
}

func decodeImage(f *os.File) (image.Image, error) {
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("rewind upload file: %w", err)
	}

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("%w: decode image: %w", ErrInvalidImage, err)
	}

	return img, nil
}

func detectedFormat(contentType string) (string, bool) {
	switch contentType {
	case mimePNG:
		return formatPNG, true
	case mimeJPEG:
		return formatJPEG, true
	default:
		return "", false
	}
}

func mimeForFormat(format string) string {
	if format == formatJPEG {
		return mimeJPEG
	}

	return mimePNG
}

func extForFormat(format string) string {
	if format == formatJPEG {
		return extJPEG
	}

	return extPNG
}

func extForMime(mime string) (string, error) {
	switch mime {
	case mimePNG:
		return extPNG, nil
	case mimeJPEG:
		return extJPEG, nil
	default:
		return "", fmt.Errorf("unsupported image mime %q", mime)
	}
}

func putFromFile(ctx context.Context, store storage.Storage, key string, f *os.File, size int64, contentType string) error {
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("rewind upload file: %w", err)
	}

	if err := store.Put(ctx, key, f, size, contentType); err != nil {
		return fmt.Errorf("put storage object: %w", err)
	}

	return nil
}
