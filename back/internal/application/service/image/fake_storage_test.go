package image

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"

	"klubok/internal/provider/storage"
)

type fakeStorage struct {
	mu sync.Mutex

	objects      map[string][]byte
	contentTypes map[string]string

	putCount int
	putErrOn int
	putErr   error

	deleteErrs  map[string][]error
	deleteCalls []string
}

func newFakeStorage() *fakeStorage {
	return &fakeStorage{
		objects:      make(map[string][]byte),
		contentTypes: make(map[string]string),
		deleteErrs:   make(map[string][]error),
	}
}

func (s *fakeStorage) Put(_ context.Context, key string, r io.Reader, size int64, contentType string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.putCount++
	if s.putErrOn > 0 && s.putCount == s.putErrOn {
		return s.putErr
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("read object: %w", err)
	}

	if int64(len(data)) != size {
		return fmt.Errorf("size mismatch: got %d, want %d", len(data), size)
	}

	s.objects[key] = data
	s.contentTypes[key] = contentType

	return nil
}

func (s *fakeStorage) Open(_ context.Context, key string) (io.ReadCloser, int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, ok := s.objects[key]
	if !ok {
		return nil, 0, storage.ErrNotFound
	}

	return io.NopCloser(bytes.NewReader(data)), int64(len(data)), nil
}

func (s *fakeStorage) Delete(_ context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.deleteCalls = append(s.deleteCalls, key)

	if errs := s.deleteErrs[key]; len(errs) > 0 {
		s.deleteErrs[key] = errs[1:]

		return errs[0]
	}

	delete(s.objects, key)

	return nil
}

func (s *fakeStorage) object(key string) ([]byte, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, ok := s.objects[key]

	return data, ok
}

func (s *fakeStorage) contentType(key string) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.contentTypes[key]
}

func (s *fakeStorage) deleteCallCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return len(s.deleteCalls)
}

func (s *fakeStorage) putCallCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.putCount
}
