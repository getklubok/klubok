package image

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"klubok/internal/provider/storage"
)

const (
	cleanupQueueSize = 256
	cleanupAttempts  = 3
)

// Blob ссылается на блоб изображения, который нужно удалить после исчезновения ссылок.
type Blob struct {
	Hash string
	Mime string
}

// cleaner удаляет блобы без ссылок в фоновом воркере.
type cleaner struct {
	storage storage.Storage
	querier Querier
	locker  *blobLocker
	log     *zap.Logger

	queue       chan []Blob
	retryDelays []time.Duration
	sleep       func(context.Context, time.Duration)

	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func newCleaner(store storage.Storage, querier Querier, locker *blobLocker, log *zap.Logger) *cleaner {
	return &cleaner{
		storage: store,
		querier: querier,
		locker:  locker,
		log:     log.Named("image_cleaner"),

		queue:       make(chan []Blob, cleanupQueueSize),
		retryDelays: []time.Duration{time.Second, 5 * time.Second},
		sleep:       sleepContext,
	}
}

func (c *cleaner) start(ctx context.Context) {
	ctx, c.cancel = context.WithCancel(ctx)
	c.wg.Go(func() {
		c.run(ctx)
	})
}

func (c *cleaner) stop() {
	if c.cancel == nil {
		return
	}

	c.cancel()
	c.wg.Wait()
}

// enqueue ставит задачу без блокировки; при переполнении задачи теряются.
func (c *cleaner) enqueue(blobs []Blob) {
	if len(blobs) == 0 {
		return
	}

	select {
	case c.queue <- blobs:
	default:
		c.log.Warn("Image cleanup queue is full, blobs will leak", zap.Int("count", len(blobs)))
	}
}

func (c *cleaner) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case task := <-c.queue:
			for _, blob := range task {
				c.process(ctx, blob)
			}
		}
	}
}

func (c *cleaner) process(ctx context.Context, blob Blob) {
	for attempt := range cleanupAttempts {
		if ctx.Err() != nil {
			return
		}

		err := c.deleteBlob(ctx, blob)
		if err == nil {
			return
		}

		c.log.Warn("Failed to delete image blob",
			zap.String("hash", blob.Hash),
			zap.Int("attempt", attempt+1),
			zap.Error(err),
		)

		if attempt == cleanupAttempts-1 {
			c.log.Error("Gave up deleting image blob", zap.String("hash", blob.Hash))

			return
		}

		c.sleep(ctx, c.retryDelays[attempt])
	}
}

func (c *cleaner) deleteBlob(ctx context.Context, blob Blob) error {
	release := c.locker.lock(blob.Hash)
	defer release()

	exists, err := c.querier.ImageExistsByBlobHash(ctx, blob.Hash)
	if err != nil {
		return fmt.Errorf("check image blob references: %w", err)
	}
	if exists {
		return nil
	}

	ext, err := extForMime(blob.Mime)
	if err != nil {
		return err
	}

	if err := c.storage.Delete(ctx, storage.OriginalKey(blob.Hash, ext)); err != nil {
		return fmt.Errorf("delete original blob: %w", err)
	}

	if err := c.storage.Delete(ctx, storage.FullKey(blob.Hash, ext)); err != nil {
		return fmt.Errorf("delete fullscreen blob: %w", err)
	}

	return nil
}

func sleepContext(ctx context.Context, d time.Duration) {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
	case <-timer.C:
	}
}
