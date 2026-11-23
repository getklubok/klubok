package image

import "sync"

type blobLockEntry struct {
	mu   sync.Mutex
	refs int
}

// blobLocker сериализует запись и очистку одного блоба внутри процесса.
type blobLocker struct {
	mu      sync.Mutex
	entries map[string]*blobLockEntry
}

func newBlobLocker() *blobLocker {
	return &blobLocker{entries: make(map[string]*blobLockEntry)}
}

// lock захватывает блокировку блоба и возвращает функцию освобождения.
func (l *blobLocker) lock(hash string) func() {
	l.mu.Lock()

	entry, ok := l.entries[hash]
	if !ok {
		entry = &blobLockEntry{}
		l.entries[hash] = entry
	}
	entry.refs++

	l.mu.Unlock()

	entry.mu.Lock()

	return func() {
		entry.mu.Unlock()

		l.mu.Lock()
		entry.refs--
		if entry.refs == 0 {
			delete(l.entries, hash)
		}
		l.mu.Unlock()
	}
}
