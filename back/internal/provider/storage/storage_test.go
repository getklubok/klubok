package storage

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBlobKeys(t *testing.T) {
	hash := strings.Repeat("ab", 32)
	otherHash := strings.Repeat("cd", 32)

	tests := []struct {
		name string
		got  string
		want string
	}{
		{
			name: "Should build original key from hash and extension",
			got:  OriginalKey(hash, "png"),
			want: "aba/" + hash + ".orig.png",
		},
		{
			name: "Should build full key from hash and extension",
			got:  FullKey(hash, "jpg"),
			want: "aba/" + hash + ".full.jpg",
		},
		{
			name: "Should build a different key for a different hash",
			got:  OriginalKey(otherHash, "png"),
			want: "cdc/" + otherHash + ".orig.png",
		},
		{
			name: "Should build a different key for a different extension",
			got:  OriginalKey(hash, "jpg"),
			want: "aba/" + hash + ".orig.jpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.got)
		})
	}
}

func TestLocal(t *testing.T) {
	key := OriginalKey(strings.Repeat("ab", 32), "png")

	tests := []struct {
		name string
		run  func(t *testing.T, s *local)
	}{
		{
			name: "Should put and open object",
			run: func(t *testing.T, s *local) {
				t.Helper()

				ctx := t.Context()
				payload := []byte("image-bytes")

				require.NoError(t, s.Put(ctx, key, bytes.NewReader(payload), int64(len(payload)), "image/png"))

				r, size, err := s.Open(ctx, key)
				require.NoError(t, err)
				t.Cleanup(func() { require.NoError(t, r.Close()) })

				got, err := io.ReadAll(r)
				require.NoError(t, err)

				assert.Equal(t, payload, got)
				assert.Equal(t, int64(len(payload)), size)
			},
		},
		{
			name: "Should overwrite object by the same key",
			run: func(t *testing.T, s *local) {
				t.Helper()

				ctx := t.Context()
				first := []byte("first")
				second := []byte("second")

				require.NoError(t, s.Put(ctx, key, bytes.NewReader(first), int64(len(first)), "image/png"))
				require.NoError(t, s.Put(ctx, key, bytes.NewReader(second), int64(len(second)), "image/png"))

				r, size, err := s.Open(ctx, key)
				require.NoError(t, err)
				t.Cleanup(func() { require.NoError(t, r.Close()) })

				got, err := io.ReadAll(r)
				require.NoError(t, err)

				assert.Equal(t, second, got)
				assert.Equal(t, int64(len(second)), size)
			},
		},
		{
			name: "Should delete object",
			run: func(t *testing.T, s *local) {
				t.Helper()

				ctx := t.Context()
				payload := []byte("image-bytes")

				require.NoError(t, s.Put(ctx, key, bytes.NewReader(payload), int64(len(payload)), "image/png"))
				require.NoError(t, s.Delete(ctx, key))

				_, _, err := s.Open(ctx, key)
				require.ErrorIs(t, err, ErrNotFound)
			},
		},
		{
			name: "Should delete absent object repeatedly",
			run: func(t *testing.T, s *local) {
				t.Helper()

				require.NoError(t, s.Delete(t.Context(), key))
				require.NoError(t, s.Delete(t.Context(), key))
			},
		},
		{
			name: "Should return ErrNotFound for absent object",
			run: func(t *testing.T, s *local) {
				t.Helper()

				_, _, err := s.Open(t.Context(), key)
				require.ErrorIs(t, err, ErrNotFound)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := newLocal(t.TempDir())
			require.NoError(t, err)

			tt.run(t, s)
		})
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name           string
		cfgFunc        func(t *testing.T) Config
		wantErrContain string
		check          func(t *testing.T, cfg Config)
	}{
		{
			name: "Should create local storage",
			cfgFunc: func(t *testing.T) Config {
				t.Helper()

				return Config{Provider: "local", Local: Local{Path: filepath.Join(t.TempDir(), "nested", "storage")}}
			},
			check: func(t *testing.T, cfg Config) {
				t.Helper()

				info, err := os.Stat(cfg.Local.Path)
				require.NoError(t, err)
				assert.True(t, info.IsDir())
			},
		},
		{
			name: "Should fail to create local storage in a file path",
			cfgFunc: func(t *testing.T) Config {
				t.Helper()

				file := filepath.Join(t.TempDir(), "file")
				require.NoError(t, os.WriteFile(file, []byte("data"), 0o600))

				return Config{Provider: "local", Local: Local{Path: filepath.Join(file, "storage")}}
			},
			wantErrContain: "create storage directory",
		},
		{
			name: "Should fail to create local storage without path",
			cfgFunc: func(t *testing.T) Config {
				t.Helper()

				return Config{Provider: "local"}
			},
			wantErrContain: "storage path is empty",
		},
		{
			name: "Should fail on unknown provider",
			cfgFunc: func(t *testing.T) Config {
				t.Helper()

				return Config{Provider: "ftp"}
			},
			wantErrContain: `unknown storage provider "ftp"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.cfgFunc(t)

			_, err := New(t.Context(), cfg)
			if tt.wantErrContain != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrContain)

				return
			}

			require.NoError(t, err)
			if tt.check != nil {
				tt.check(t, cfg)
			}
		})
	}
}

type fakeAPIError struct {
	code string
}

func (e fakeAPIError) Error() string {
	return e.code
}

func (e fakeAPIError) ErrorCode() string {
	return e.code
}

type fakeStatusError struct {
	status int
}

func (e fakeStatusError) Error() string {
	return http.StatusText(e.status)
}

func (e fakeStatusError) HTTPStatusCode() int {
	return e.status
}

func TestMapGetObjectError(t *testing.T) {
	otherAPI := fakeAPIError{code: "InternalError"}
	otherStatus := fakeStatusError{status: http.StatusInternalServerError}
	other := errors.New("other error")

	tests := []struct {
		name string
		err  error
		want error
	}{
		{
			name: "Should map NoSuchKey API error",
			err:  fakeAPIError{code: "NoSuchKey"},
			want: ErrNotFound,
		},
		{
			name: "Should map NotFound API error",
			err:  fakeAPIError{code: "NotFound"},
			want: ErrNotFound,
		},
		{
			name: "Should map wrapped NoSuchKey API error",
			err:  fmt.Errorf("wrapped: %w", fakeAPIError{code: "NoSuchKey"}),
			want: ErrNotFound,
		},
		{
			name: "Should map 404 status error",
			err:  fakeStatusError{status: http.StatusNotFound},
			want: ErrNotFound,
		},
		{
			name: "Should keep other API error",
			err:  otherAPI,
			want: otherAPI,
		},
		{
			name: "Should keep other status error",
			err:  otherStatus,
			want: otherStatus,
		},
		{
			name: "Should keep unknown error",
			err:  other,
			want: other,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, mapGetObjectError(tt.err))
		})
	}
}

func TestS3ObjectKey(t *testing.T) {
	key := "abc/hash.orig.png"

	tests := []struct {
		name   string
		prefix string
		want   string
	}{
		{
			name:   "Should keep logical key without prefix",
			prefix: "",
			want:   key,
		},
		{
			name:   "Should add prefix only to physical key",
			prefix: "photos",
			want:   "photos/" + key,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &s3Storage{prefix: tt.prefix}

			assert.Equal(t, tt.want, s.objectKey(key))
		})
	}
}
