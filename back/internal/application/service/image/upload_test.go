package image

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"image"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"klubok/internal/application/service/image/mocks"
	"klubok/internal/db/sqlc"
	"klubok/internal/domain"
	"klubok/internal/provider/storage"
)

const testMaxPixels = 10000

func TestService_Upload(t *testing.T) {
	tests := []struct {
		name              string
		width             int
		height            int
		encode            func(*testing.T, int, int) []byte
		fileName          string
		cfg               Config
		wantMime          string
		wantExt           string
		wantFormat        string
		wantName          string
		wantFullSame      bool
		wantPreviewWidth  int
		wantPreviewHeight int
	}{
		{
			name:       "PNG with jpg name",
			width:      80,
			height:     60,
			encode:     encodePNG,
			fileName:   "photo.jpg",
			cfg:        Config{MaxFileSizeBytes: 1 << 20, MaxPixels: testMaxPixels, PreviewMaxSidePx: 20, FullscreenMaxSidePx: 40},
			wantMime:   mimePNG,
			wantExt:    extPNG,
			wantFormat: formatPNG,
			wantName:   "photo.jpg",

			wantPreviewWidth:  20,
			wantPreviewHeight: 15,
		},
		{
			name:       "JPEG",
			width:      80,
			height:     60,
			encode:     encodeJPEG,
			fileName:   "photo.jpeg",
			cfg:        Config{MaxFileSizeBytes: 1 << 20, MaxPixels: testMaxPixels, PreviewMaxSidePx: 20, FullscreenMaxSidePx: 40},
			wantMime:   mimeJPEG,
			wantExt:    extJPEG,
			wantFormat: formatJPEG,
			wantName:   "photo.jpeg",

			wantPreviewWidth:  20,
			wantPreviewHeight: 15,
		},
		{
			name:       "Small JPEG is not upscaled",
			width:      10,
			height:     8,
			encode:     encodeJPEG,
			fileName:   "tiny.jpeg",
			cfg:        Config{MaxFileSizeBytes: 1 << 20, MaxPixels: testMaxPixels, PreviewMaxSidePx: 20, FullscreenMaxSidePx: 40},
			wantMime:   mimeJPEG,
			wantExt:    extJPEG,
			wantFormat: formatJPEG,
			wantName:   "tiny.jpeg",
			// fullscreen без уменьшения сохраняет байты оригинала
			wantFullSame: true,

			wantPreviewWidth:  10,
			wantPreviewHeight: 8,
		},
		{
			name:       "Tall PNG keeps proportions",
			width:      30,
			height:     80,
			encode:     encodePNG,
			fileName:   "scheme.png",
			cfg:        Config{MaxFileSizeBytes: 1 << 20, MaxPixels: testMaxPixels, PreviewMaxSidePx: 20, FullscreenMaxSidePx: 40},
			wantMime:   mimePNG,
			wantExt:    extPNG,
			wantFormat: formatPNG,
			wantName:   "scheme.png",

			wantPreviewWidth:  8,
			wantPreviewHeight: 20,
		},
		{
			name:       "Path components are removed from name",
			width:      40,
			height:     30,
			encode:     encodePNG,
			fileName:   `..\photos\scheme.png`,
			cfg:        Config{MaxFileSizeBytes: 1 << 20, MaxPixels: testMaxPixels, PreviewMaxSidePx: 20, FullscreenMaxSidePx: 40},
			wantMime:   mimePNG,
			wantExt:    extPNG,
			wantFormat: formatPNG,
			wantName:   "scheme.png",

			wantPreviewWidth:  20,
			wantPreviewHeight: 15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := tt.encode(t, tt.width, tt.height)
			hash := sha256Hex(file)

			store := newFakeStorage()

			var created sqlc.ImageCreateParams

			querier := mocks.NewQuerierMock(t)
			querier.EXPECT().
				ImageCreate(mock.Anything, mock.Anything).
				RunAndReturn(func(_ context.Context, arg sqlc.ImageCreateParams) (sqlc.Image, error) {
					created = arg

					return sqlc.Image{ID: arg.ID, DiagramID: arg.DiagramID, BlobHash: arg.BlobHash}, nil
				})

			svc := NewService(querier, store, tt.cfg, zap.NewNop())
			svc.idFunc = func() string { return "image-id" }

			got, err := svc.Upload(context.Background(), domain.DiagramID("diagram-id"), tt.fileName, bytes.NewReader(file))
			require.NoError(t, err)

			assert.Equal(t, "image-id", got.ID)
			assert.Equal(t, tt.wantName, got.OriginalName)
			assert.Equal(t, tt.wantMime, got.Mime)
			assert.Equal(t, tt.width, got.Width)
			assert.Equal(t, tt.height, got.Height)

			previewCfg, previewFormat := decodeDataURL(t, got.Pv)
			assert.Equal(t, formatJPEG, previewFormat)
			assert.Equal(t, tt.wantPreviewWidth, previewCfg.Width)
			assert.Equal(t, tt.wantPreviewHeight, previewCfg.Height)

			originalKey := storage.OriginalKey(hash, tt.wantExt)
			fullKey := storage.FullKey(hash, tt.wantExt)

			original, ok := store.object(originalKey)
			require.True(t, ok)
			assert.Equal(t, file, original)
			assert.Equal(t, tt.wantMime, store.contentType(originalKey))

			fullscreen, ok := store.object(fullKey)
			require.True(t, ok)
			assert.Equal(t, tt.wantMime, store.contentType(fullKey))

			if tt.wantFullSame {
				assert.Equal(t, file, fullscreen)
			} else {
				fullCfg, fullFormat, err := image.DecodeConfig(bytes.NewReader(fullscreen))
				require.NoError(t, err)

				wantWidth, wantHeight := fitSize(tt.width, tt.height, tt.cfg.FullscreenMaxSidePx)
				assert.Equal(t, tt.wantFormat, fullFormat)
				assert.Equal(t, wantWidth, fullCfg.Width)
				assert.Equal(t, wantHeight, fullCfg.Height)
			}

			assert.Equal(t, "image-id", created.ID)
			assert.Equal(t, "diagram-id", created.DiagramID)
			assert.Equal(t, hash, created.BlobHash)
			assert.Equal(t, tt.wantMime, created.Mime)
			assert.Equal(t, tt.wantName, created.OriginalName)
			assert.NotEmpty(t, created.Preview)
			assert.Empty(t, svc.cleaner.queue)
		})
	}
}

func TestService_Upload_Invalid(t *testing.T) {
	validPNG := encodePNG(t, 40, 30)
	validJPEG := encodeJPEG(t, 40, 30)

	tests := []struct {
		name    string
		file    []byte
		cfg     Config
		wantErr error
	}{
		{
			name:    "File too large",
			file:    validPNG,
			cfg:     Config{MaxFileSizeBytes: 100, MaxPixels: testMaxPixels, PreviewMaxSidePx: 10, FullscreenMaxSidePx: 20},
			wantErr: ErrFileTooLarge,
		},
		{
			name:    "Too many pixels",
			file:    validPNG,
			cfg:     Config{MaxFileSizeBytes: 1 << 20, MaxPixels: 100, PreviewMaxSidePx: 10, FullscreenMaxSidePx: 20},
			wantErr: ErrInvalidImage,
		},
		{
			name:    "GIF",
			file:    encodeGIF(t, 40, 30),
			cfg:     Config{MaxFileSizeBytes: 1 << 20, MaxPixels: testMaxPixels, PreviewMaxSidePx: 10, FullscreenMaxSidePx: 20},
			wantErr: ErrInvalidImage,
		},
		{
			name:    "Truncated PNG",
			file:    validPNG[:len(validPNG)/2],
			cfg:     Config{MaxFileSizeBytes: 1 << 20, MaxPixels: testMaxPixels, PreviewMaxSidePx: 10, FullscreenMaxSidePx: 20},
			wantErr: ErrInvalidImage,
		},
		{
			name:    "Truncated JPEG",
			file:    validJPEG[:len(validJPEG)/2],
			cfg:     Config{MaxFileSizeBytes: 1 << 20, MaxPixels: testMaxPixels, PreviewMaxSidePx: 10, FullscreenMaxSidePx: 20},
			wantErr: ErrInvalidImage,
		},
		{
			name:    "Empty file",
			file:    nil,
			cfg:     Config{MaxFileSizeBytes: 1 << 20, MaxPixels: testMaxPixels, PreviewMaxSidePx: 10, FullscreenMaxSidePx: 20},
			wantErr: ErrInvalidImage,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newFakeStorage()
			querier := mocks.NewQuerierMock(t)
			svc := NewService(querier, store, tt.cfg, zap.NewNop())

			_, err := svc.Upload(context.Background(), domain.DiagramID("diagram-id"), "image.png", bytes.NewReader(tt.file))
			require.ErrorIs(t, err, tt.wantErr)

			assert.Zero(t, store.putCallCount())
			assert.Empty(t, svc.cleaner.queue)
		})
	}
}

func TestService_Upload_PutErrors(t *testing.T) {
	errPut := errors.New("put failed")

	tests := []struct {
		name        string
		putErrOn    int
		wantEnqueue bool
	}{
		{
			name:     "Original put error",
			putErrOn: 1,
		},
		{
			name:        "Fullscreen put error",
			putErrOn:    2,
			wantEnqueue: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := encodePNG(t, 80, 60)
			store := newFakeStorage()
			store.putErrOn = tt.putErrOn
			store.putErr = errPut

			querier := mocks.NewQuerierMock(t)
			svc := NewService(querier, store, Config{
				MaxFileSizeBytes:    1 << 20,
				MaxPixels:           testMaxPixels,
				PreviewMaxSidePx:    20,
				FullscreenMaxSidePx: 40,
			}, zap.NewNop())

			_, err := svc.Upload(context.Background(), domain.DiagramID("diagram-id"), "image.png", bytes.NewReader(file))
			require.Error(t, err)

			if tt.wantEnqueue {
				assert.Len(t, svc.cleaner.queue, 1)
			} else {
				assert.Empty(t, svc.cleaner.queue)
			}
		})
	}
}

func TestService_Upload_InsertErrorEnqueuesCleanup(t *testing.T) {
	errInsert := errors.New("insert failed")

	file := encodePNG(t, 80, 60)
	hash := sha256Hex(file)

	store := newFakeStorage()
	querier := mocks.NewQuerierMock(t)
	querier.EXPECT().
		ImageCreate(mock.Anything, mock.Anything).
		Return(sqlc.Image{}, errInsert)

	svc := NewService(querier, store, Config{
		MaxFileSizeBytes:    1 << 20,
		MaxPixels:           testMaxPixels,
		PreviewMaxSidePx:    20,
		FullscreenMaxSidePx: 40,
	}, zap.NewNop())

	_, err := svc.Upload(context.Background(), domain.DiagramID("diagram-id"), "image.png", bytes.NewReader(file))
	require.ErrorIs(t, err, errInsert)

	require.Len(t, svc.cleaner.queue, 1)

	task := <-svc.cleaner.queue
	require.Len(t, task, 1)
	assert.Equal(t, Blob{Hash: hash, Mime: mimePNG}, task[0])
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)

	return hex.EncodeToString(sum[:])
}
