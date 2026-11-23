// Package storage предоставляет единый контракт хранения блобов изображений.
package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
)

// ErrNotFound возвращается, когда объекта нет в хранилище.
var ErrNotFound = errors.New("storage object not found")

// Storage — контракт хранилища блобов.
type Storage interface {
	Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error
	Open(ctx context.Context, key string) (io.ReadCloser, int64, error)
	Delete(ctx context.Context, key string) error
}

var (
	_ Storage = (*local)(nil)
	_ Storage = (*s3Storage)(nil)
)

// Config параметры хранилища.
type Config struct {
	Provider string
	Local    Local
	S3       S3
}

// Local параметры локального хранилища.
type Local struct {
	Path string
}

// S3 параметры S3-совместимого хранилища.
type S3 struct {
	Endpoint       string
	Region         string
	Bucket         string
	AccessKey      string
	SecretKey      string
	ForcePathStyle bool
	Prefix         string
}

const (
	providerLocal = "local"
	providerS3    = "s3"
	hashShardLen  = 3
)

// New создаёт хранилище выбранного провайдера.
func New(ctx context.Context, cfg Config) (Storage, error) {
	switch cfg.Provider {
	case providerLocal:
		return newLocal(cfg.Local.Path)
	case providerS3:
		return newS3(ctx, cfg.S3)
	default:
		return nil, fmt.Errorf("unknown storage provider %q", cfg.Provider)
	}
}

// OriginalKey возвращает логический ключ оригинала по хешу и расширению.
func OriginalKey(hash, ext string) string {
	return blobKey(hash, "orig", ext)
}

// FullKey возвращает логический ключ полноэкранного варианта по хешу и расширению.
func FullKey(hash, ext string) string {
	return blobKey(hash, "full", ext)
}

func blobKey(hash, variant, ext string) string {
	if len(hash) < hashShardLen {
		return fmt.Sprintf("%s.%s.%s", hash, variant, ext)
	}

	return fmt.Sprintf("%s/%s.%s.%s", hash[:hashShardLen], hash, variant, ext)
}
