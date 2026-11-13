package session

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"klubok/internal/domain"
)

const (
	saveInterval = 60 * time.Second // Периодичность сброса данных в БД

	oneSaveTimeoutSec  = 3
	defaultSaveTimeout = 10 * time.Second
)

// Updater периодически обновляет данные в БД.
type Updater struct {
	service     *Service
	expiresDays int
	logger      *zap.Logger

	saveInterval time.Duration
	wg           sync.WaitGroup                 // WaitGroup для всех горутин
	saveWG       sync.WaitGroup                 // отдельный WaitGroup только для save операций
	cache        map[domain.SessionID]time.Time // sessionID
	mu           sync.Mutex
	stopChan     chan struct{}

	nowFunc func() time.Time
}

// NewUpdater создает новый экземпляр Updater.
func NewUpdater(sessionService *Service, expiresDays int, logger *zap.Logger) *Updater {
	su := &Updater{
		service:     sessionService,
		expiresDays: expiresDays,
		logger:      logger.Named("session_updater"),

		saveInterval: saveInterval,
		cache:        make(map[domain.SessionID]time.Time),
		stopChan:     make(chan struct{}),

		nowFunc: time.Now,
	}

	su.wg.Go(func() {
		su.saveLoop()
	})
	su.logger.Info("Session updater started")

	return su
}

// UpdateTime обновляет время в кэше.
func (su *Updater) UpdateTime(sessionID domain.SessionID) {
	su.mu.Lock()
	defer su.mu.Unlock()

	su.cache[sessionID] = su.nowFunc()
}

// Stop останавливает фоновую горутину.
func (su *Updater) Stop() {
	close(su.stopChan)
	su.wg.Wait()     // ждем завершения saveLoop
	su.saveWG.Wait() // ждем завершения всех save операций
	su.logger.Info("Session updater stopped")
}

// saveLoop периодически сохраняет данные в БД.
func (su *Updater) saveLoop() {
	ticker := time.NewTicker(su.saveInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			outdatedTime := su.nowFunc().AddDate(0, 0, -1*su.expiresDays)
			if err := su.service.CancelOld(context.Background(), outdatedTime); err != nil {
				su.logger.Error("Failed to cancel old sessions", zap.Error(err))
			}

			su.save() // Сохраняем накопленные данные в БД
		case <-su.stopChan:
			su.save() // Сохраняем оставшиеся данные перед выходом
			return
		}
	}
}

// save сохраняет накопленные данные в БД.
func (su *Updater) save() {
	su.mu.Lock()
	defer su.mu.Unlock()

	if len(su.cache) == 0 {
		return
	}

	// Создаем копию данных для работы
	cacheCopy := su.cache
	su.cache = make(map[domain.SessionID]time.Time)

	su.saveWG.Add(1)
	go func(data map[domain.SessionID]time.Time) {
		defer su.saveWG.Done()

		timeout, err := time.ParseDuration(
			fmt.Sprintf("%ds", oneSaveTimeoutSec*len(data)),
		)
		if err != nil {
			timeout = defaultSaveTimeout
		}

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		for sessionID, checkedAt := range data {
			if err := su.service.UpdateCheckedAt(ctx, sessionID, checkedAt); err != nil {
				su.logger.Error("Failed to update session",
					zap.String("part_session_id", domain.PartID(sessionID.String())),
					zap.Error(err),
				)
			}
		}

		su.logger.Info("Sessions saved", zap.Int("count", len(data)))
	}(cacheCopy)
}
