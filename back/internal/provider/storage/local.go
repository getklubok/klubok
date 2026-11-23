package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

const storageDirPerm = 0o750

// local хранит объекты на локальном диске.
type local struct {
	root string
}

// newLocal создаёт локальное хранилище и проверяет доступность каталога для записи.
func newLocal(root string) (*local, error) {
	if root == "" {
		return nil, errors.New("storage path is empty")
	}

	if err := os.MkdirAll(root, storageDirPerm); err != nil {
		return nil, fmt.Errorf("create storage directory: %w", err)
	}

	probe, err := os.CreateTemp(root, ".probe-*")
	if err != nil {
		return nil, fmt.Errorf("storage directory is not writable: %w", err)
	}

	name := probe.Name()
	closeErr := probe.Close()
	removeErr := os.Remove(name)

	if closeErr != nil {
		return nil, fmt.Errorf("close probe file: %w", closeErr)
	}

	if removeErr != nil {
		return nil, fmt.Errorf("remove probe file: %w", removeErr)
	}

	return &local{root: root}, nil
}

// Put записывает объект через временный файл и переименование.
func (l *local) Put(_ context.Context, key string, r io.Reader, size int64, _ string) error {
	dst := l.path(key)

	if err := os.MkdirAll(filepath.Dir(dst), storageDirPerm); err != nil {
		return fmt.Errorf("create shard directory: %w", err)
	}

	tmp, err := os.CreateTemp(filepath.Dir(dst), ".tmp-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}

	tmpName := tmp.Name()

	written, err := io.Copy(tmp, r)
	if err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpName)

		return fmt.Errorf("write temp file: %w", err)
	}

	if written != size {
		_ = tmp.Close()
		_ = os.Remove(tmpName)

		return fmt.Errorf("write temp file: size mismatch: got %d, want %d", written, size)
	}

	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)

		return fmt.Errorf("close temp file: %w", err)
	}

	if err := os.Rename(tmpName, dst); err != nil {
		_ = os.Remove(tmpName)

		return fmt.Errorf("rename temp file: %w", err)
	}

	return nil
}

// Open открывает объект и возвращает его размер.
func (l *local) Open(_ context.Context, key string) (io.ReadCloser, int64, error) {
	f, err := os.Open(l.path(key))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, 0, ErrNotFound
	}

	if err != nil {
		return nil, 0, fmt.Errorf("open storage object: %w", err)
	}

	info, err := f.Stat()
	if err != nil {
		_ = f.Close()

		return nil, 0, fmt.Errorf("stat storage object: %w", err)
	}

	return f, info.Size(), nil
}

// Delete удаляет объект; удаление отсутствующего объекта не является ошибкой.
func (l *local) Delete(_ context.Context, key string) error {
	err := os.Remove(l.path(key))
	if err == nil || errors.Is(err, fs.ErrNotExist) {
		return nil
	}

	return fmt.Errorf("delete storage object: %w", err)
}

func (l *local) path(key string) string {
	return filepath.Join(l.root, filepath.FromSlash(key))
}
