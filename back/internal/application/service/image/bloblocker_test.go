package image

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBlobLocker_SameHashIsSerialized(t *testing.T) {
	locker := newBlobLocker()

	release := locker.lock("hash")

	acquired := make(chan struct{})

	go func() {
		unlock := locker.lock("hash")
		close(acquired)
		unlock()
	}()

	select {
	case <-acquired:
		t.Fatal("second lock acquired before first was released")
	case <-time.After(20 * time.Millisecond):
	}

	release()

	select {
	case <-acquired:
	case <-time.After(time.Second):
		t.Fatal("second lock was not acquired after release")
	}
}

func TestBlobLocker_RemovesReleasedEntry(t *testing.T) {
	locker := newBlobLocker()

	unlock := locker.lock("hash")

	locker.mu.Lock()
	_, ok := locker.entries["hash"]
	locker.mu.Unlock()

	assert.True(t, ok)

	unlock()

	locker.mu.Lock()
	_, ok = locker.entries["hash"]
	locker.mu.Unlock()

	assert.False(t, ok)
}
