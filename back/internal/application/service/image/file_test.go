package image

import (
	"context"
	"database/sql"
	"errors"
	"io"
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

const (
	testImageID   = domain.ImageID("image-id")
	testDiagramID = domain.DiagramID("diagram-id")
)

var errStorageBroken = errors.New("storage is broken")

func TestParseVariant(t *testing.T) {
	tests := []struct {
		name    string
		variant string
		want    Variant
		wantErr bool
	}{
		{name: "full", variant: "full", want: VariantFull},
		{name: "download", variant: "download", want: VariantDownload},
		{name: "preview is not served", variant: "preview", wantErr: true},
		{name: "empty", variant: "", wantErr: true},
		{name: "case sensitive", variant: "Full", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseVariant(tt.variant)
			if tt.wantErr {
				require.Error(t, err)
				assert.Empty(t, got)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestService_Open(t *testing.T) {
	const (
		hash = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
		blob = "blob content"
		name = "scheme.png"
	)

	tests := []struct {
		name       string
		variant    Variant
		mime       string
		objects    []string
		rowErr     error
		openErr    error
		wantKey    string
		wantAnyErr bool
		wantError  error
	}{
		{
			name:    "full variant reads fullscreen blob",
			variant: VariantFull,
			mime:    mimePNG,
			objects: []string{storage.OriginalKey(hash, extPNG), storage.FullKey(hash, extPNG)},
			wantKey: storage.FullKey(hash, extPNG),
		},
		{
			name:    "download variant reads original blob",
			variant: VariantDownload,
			mime:    mimeJPEG,
			objects: []string{storage.OriginalKey(hash, extJPEG), storage.FullKey(hash, extJPEG)},
			wantKey: storage.OriginalKey(hash, extJPEG),
		},
		{
			name:      "image row not found",
			variant:   VariantFull,
			mime:      mimePNG,
			rowErr:    sql.ErrNoRows,
			wantError: ErrNotFound,
		},
		{
			name:      "blob not found",
			variant:   VariantFull,
			mime:      mimePNG,
			wantError: ErrNotFound,
		},
		{
			name:       "unsupported mime",
			variant:    VariantFull,
			mime:       "image/gif",
			wantAnyErr: true,
		},
		{
			name:      "storage error",
			variant:   VariantFull,
			mime:      mimePNG,
			objects:   []string{storage.FullKey(hash, extPNG)},
			openErr:   errStorageBroken,
			wantError: errStorageBroken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newBrokenStorage(newFakeStorage(), tt.openErr)
			for _, key := range tt.objects {
				store.objects[key] = []byte(blob)
			}

			querier := mocks.NewQuerierMock(t)
			querier.EXPECT().
				ImageGet(mock.Anything, testImageID.String(), testDiagramID.String()).
				Return(sqlc.Image{
					ID:           testImageID.String(),
					DiagramID:    testDiagramID.String(),
					BlobHash:     hash,
					Mime:         tt.mime,
					OriginalName: name,
				}, tt.rowErr)

			svc := NewService(querier, store, Config{}, zap.NewNop())

			got, err := svc.Open(context.Background(), testDiagramID, testImageID, tt.variant)

			if tt.wantAnyErr {
				require.Error(t, err)
				assert.NotErrorIs(t, err, ErrNotFound)

				return
			}

			if tt.wantError != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantError)
				assert.Empty(t, got.Reader)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, int64(len(blob)), got.Size)
			assert.Equal(t, tt.mime, got.Mime)
			assert.Equal(t, name, got.OriginalName)
			assert.Equal(t, tt.wantKey, store.openedKey)

			data, err := io.ReadAll(got.Reader)
			require.NoError(t, err)
			assert.Equal(t, blob, string(data))

			require.NoError(t, got.Reader.Close())
		})
	}
}

// brokenStorage имитирует ошибку хранилища, отличную от ErrNotFound.
type brokenStorage struct {
	*fakeStorage

	err error

	openedKey string
}

func newBrokenStorage(inner *fakeStorage, err error) *brokenStorage {
	return &brokenStorage{fakeStorage: inner, err: err}
}

func (s *brokenStorage) Open(ctx context.Context, key string) (io.ReadCloser, int64, error) {
	s.openedKey = key

	if s.err != nil {
		return nil, 0, s.err
	}

	return s.fakeStorage.Open(ctx, key)
}
