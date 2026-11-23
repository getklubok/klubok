package image

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"klubok/internal/application/service/image/mocks"
	"klubok/internal/provider/storage"
)

func TestCleaner_Process(t *testing.T) {
	const hash = "hash"

	originalKey := storage.OriginalKey(hash, extPNG)
	fullKey := storage.FullKey(hash, extPNG)

	errDelete := errors.New("delete failed")
	errDB := errors.New("db failed")

	tests := []struct {
		name               string
		exists             bool
		existsErr          error
		deleteErrs         []error
		wantDeleteCalls    int
		wantOriginalExists bool
		wantFullExists     bool
	}{
		{
			name:               "Referenced blob is kept",
			exists:             true,
			wantDeleteCalls:    0,
			wantOriginalExists: true,
			wantFullExists:     true,
		},
		{
			name:            "Unreferenced blob is deleted",
			wantDeleteCalls: 2,
		},
		{
			name:            "Transient delete error is retried",
			deleteErrs:      []error{errDelete},
			wantDeleteCalls: 3,
		},
		{
			name:               "Permanent delete error gives up",
			deleteErrs:         []error{errDelete, errDelete, errDelete},
			wantDeleteCalls:    3,
			wantOriginalExists: true,
			wantFullExists:     true,
		},
		{
			name:               "Reference check error gives up",
			existsErr:          errDB,
			wantDeleteCalls:    0,
			wantOriginalExists: true,
			wantFullExists:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newFakeStorage()
			putTestObject(t, store, originalKey)
			putTestObject(t, store, fullKey)

			if len(tt.deleteErrs) > 0 {
				store.deleteErrs[originalKey] = tt.deleteErrs
			}

			querier := mocks.NewQuerierMock(t)
			querier.EXPECT().
				ImageExistsByBlobHash(mock.Anything, hash).
				Return(tt.exists, tt.existsErr)

			cleaner := newCleaner(store, querier, newBlobLocker(), zap.NewNop())
			cleaner.sleep = func(context.Context, time.Duration) {}

			cleaner.process(context.Background(), Blob{Hash: hash, Mime: mimePNG})

			assert.Equal(t, tt.wantDeleteCalls, store.deleteCallCount())
			assert.Equal(t, tt.wantOriginalExists, objectExists(t, store, originalKey))
			assert.Equal(t, tt.wantFullExists, objectExists(t, store, fullKey))
		})
	}
}

func TestCleaner_EnqueueOverflow(t *testing.T) {
	cleaner := newCleaner(newFakeStorage(), mocks.NewQuerierMock(t), newBlobLocker(), zap.NewNop())
	cleaner.queue = make(chan []Blob, 1)

	cleaner.enqueue([]Blob{{Hash: "first", Mime: mimePNG}})
	cleaner.enqueue([]Blob{{Hash: "second", Mime: mimePNG}})

	assert.Len(t, cleaner.queue, 1)
	assert.Equal(t, "first", (<-cleaner.queue)[0].Hash)
}

func TestCleaner_RepeatedDelete(t *testing.T) {
	const hash = "hash"

	store := newFakeStorage()
	putTestObject(t, store, storage.OriginalKey(hash, extPNG))
	putTestObject(t, store, storage.FullKey(hash, extPNG))

	querier := mocks.NewQuerierMock(t)
	querier.EXPECT().
		ImageExistsByBlobHash(mock.Anything, hash).
		Return(false, nil)

	cleaner := newCleaner(store, querier, newBlobLocker(), zap.NewNop())
	cleaner.sleep = func(context.Context, time.Duration) {}

	cleaner.process(context.Background(), Blob{Hash: hash, Mime: mimePNG})
	cleaner.process(context.Background(), Blob{Hash: hash, Mime: mimePNG})

	assert.Equal(t, 4, store.deleteCallCount())
}

func TestCleaner_StartStop(t *testing.T) {
	cleaner := newCleaner(newFakeStorage(), mocks.NewQuerierMock(t), newBlobLocker(), zap.NewNop())

	cleaner.start(context.Background())
	cleaner.stop()
	cleaner.stop()
}

func putTestObject(t *testing.T, store *fakeStorage, key string) {
	t.Helper()

	data := []byte("image")

	require.NoError(t, store.Put(context.Background(), key, bytes.NewReader(data), int64(len(data)), mimePNG))
}

func objectExists(t *testing.T, store *fakeStorage, key string) bool {
	t.Helper()

	_, ok := store.object(key)

	return ok
}
