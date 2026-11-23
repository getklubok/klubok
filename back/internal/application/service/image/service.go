package image

import (
	"context"

	"go.uber.org/zap"

	"klubok/internal/domain"
	"klubok/internal/provider/storage"
)

// Config задаёт лимиты обработки загружаемых изображений.
type Config struct {
	MaxFileSizeBytes    int64
	MaxPixels           int
	PreviewMaxSidePx    int
	FullscreenMaxSidePx int
}

// Service загружает изображения диаграмм и управляет жизненным циклом блобов.
type Service struct {
	querier Querier
	storage storage.Storage
	locker  *blobLocker
	cleaner *cleaner
	log     *zap.Logger
	cfg     Config

	idFunc func() string
}

// NewService создаёт сервис изображений.
func NewService(querier Querier, store storage.Storage, cfg Config, log *zap.Logger) *Service {
	s := &Service{
		querier: querier,
		storage: store,
		locker:  newBlobLocker(),
		log:     log.Named("image_service"),
		cfg:     cfg,

		idFunc: domain.NewID,
	}
	s.cleaner = newCleaner(store, querier, s.locker, s.log)

	return s
}

// Start запускает фоновый очиститель блобов.
func (s *Service) Start(ctx context.Context) {
	s.cleaner.start(ctx)
}

// Stop останавливает фоновый очиститель и ждёт его завершения.
func (s *Service) Stop() {
	s.cleaner.stop()
}

// EnqueueCleanup ставит в очередь удаление блобов, на которые больше нет ссылок.
func (s *Service) EnqueueCleanup(blobs []Blob) {
	s.cleaner.enqueue(blobs)
}
