package session

import (
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"klubok/internal/application/service/session/mocks"
	"klubok/internal/db/sqlc"
	"klubok/internal/domain"
)

func TestUpdater_UpdateTime(t *testing.T) {
	su := &Updater{
		cache:   make(map[domain.SessionID]time.Time),
		nowFunc: func() time.Time { return time.Date(2024, 12, 18, 14, 20, 0, 0, time.UTC) },
	}

	_, ok := su.cache[domain.SessionID("01JHHXTMRFMJ4DQJVTVCHZSZX3")]
	assert.False(t, ok)

	su.UpdateTime(domain.SessionID("01JHHXTMRFMJ4DQJVTVCHZSZX3"))

	actual, ok := su.cache[domain.SessionID("01JHHXTMRFMJ4DQJVTVCHZSZX3")]
	assert.True(t, ok)
	assert.Equal(t, time.Date(2024, 12, 18, 14, 20, 0, 0, time.UTC), actual)
}

func TestUpdater_Stop(t *testing.T) {
	su := &Updater{
		logger:   zap.NewNop(),
		stopChan: make(chan struct{}),
	}

	su.Stop()

	_, ok := <-su.stopChan
	assert.False(t, ok)
}

func TestUpdater_saveLoop(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		testSaveInterval := 3 * time.Second

		querier := mocks.NewQuerierMock(t)
		querier.EXPECT().
			SessionCancelOld(mock.Anything, time.Date(2024, 12, 18, 14, 20, 0, 0, time.UTC)).
			Return(nil)

		su := &Updater{
			service: &Service{
				querier: querier,
			},
			logger: zap.NewNop(),

			saveInterval: testSaveInterval,
			cache:        make(map[domain.SessionID]time.Time),
			stopChan:     make(chan struct{}),

			nowFunc: func() time.Time { return time.Date(2024, 12, 18, 14, 20, 0, 0, time.UTC) },
		}

		go su.saveLoop()
		synctest.Wait() // goroutine blocked on ticker.C

		time.Sleep(testSaveInterval) // virtual clock advances, ticker fires
		synctest.Wait()              // goroutine finished processing tick

		su.Stop()
	})
}

func TestUpdater_save(t *testing.T) {
	tests := []struct {
		name    string
		srvFunc func(t *testing.T) *Updater
	}{
		{
			name: "Successful save",
			srvFunc: func(t *testing.T) *Updater {
				t.Helper()

				querier := mocks.NewQuerierMock(t)
				querier.EXPECT().
					SessionUpdateCheckedAt(mock.Anything, "01JHHXTMRFMJ4DQJVTVCHZSZX3", time.Date(2024, 12, 18, 14, 15, 0, 0, time.UTC)).
					Return(sqlc.Session{
						ID:        "01JHHXTMRFMJ4DQJVTVCHZSZX3",
						Cid:       "cid",
						Uid:       "uid",
						CheckedAt: time.Date(2024, 12, 18, 14, 15, 0, 0, time.UTC),
					}, nil)

				return &Updater{
					service: &Service{
						querier: querier,
					},
					logger: zap.NewNop(),
					cache: map[domain.SessionID]time.Time{
						domain.SessionID("01JHHXTMRFMJ4DQJVTVCHZSZX3"): time.Date(2024, 12, 18, 14, 15, 0, 0, time.UTC),
					},
					stopChan: make(chan struct{}),
				}
			},
		},
		{
			name: "Empty cache",
			srvFunc: func(t *testing.T) *Updater {
				t.Helper()

				return &Updater{
					logger:   zap.NewNop(),
					cache:    map[domain.SessionID]time.Time{},
					stopChan: make(chan struct{}),
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := tt.srvFunc(t)
			defer srv.Stop() // на случай если тест упадет

			srv.save()

			srv.saveWG.Wait()

			assert.Empty(t, srv.cache)
		})
	}
}
