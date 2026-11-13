package metrics

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestCleanPath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "No ULID",
			path: "/api/users",
			want: "/api/users",
		},
		{
			name: "Single ULID",
			path: "/api/users/01ARZ3NDEKTSV4RRFFQ69G5FAV",
			want: "/api/users/:id",
		},
		{
			name: "Multiple ULIDs",
			path: "/api/users/01ARZ3NDEKTSV4RRFFQ69G5FAV/posts/01BX5ZZKBKACTAV9WEVGEMMVRZ",
			want: "/api/users/:id/posts/:id",
		},
		{
			name: "ULID at start",
			path: "/01ARZ3NDEKTSV4RRFFQ69G5FAV/details",
			want: "/:id/details",
		},
		{
			name: "Only ULID",
			path: "/01ARZ3NDEKTSV4RRFFQ69G5FAV",
			want: "/:id",
		},
		{
			name: "Empty path",
			path: "",
			want: "",
		},
		{
			name: "Root path",
			path: "/",
			want: "/",
		},
		{
			name: "Short alphanumeric not replaced",
			path: "/api/items/ABC123",
			want: "/api/items/ABC123",
		},
		{
			name: "25 chars not replaced",
			path: "/api/items/01ARZ3NDEKTSV4RRFFQ69G5FA",
			want: "/api/items/01ARZ3NDEKTSV4RRFFQ69G5FA",
		},
		{
			name: "embed token replaced",
			path: "/embed/50dd8e4c4f0f8f8f9c1f5e8d7c6b5a4938271605f4e3d2c1b0a9988776655443",
			want: "/embed/:token",
		},
		{
			name: "nested embed path not replaced",
			path: "/embed/token/extra",
			want: "/embed/token/extra",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cleanPath(tt.path)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestWrapResponseWriter_Status(t *testing.T) {
	tests := []struct {
		name string
		code int
		want string
	}{
		{
			name: "OK",
			code: http.StatusOK,
			want: "200",
		},
		{
			name: "Not Found",
			code: http.StatusNotFound,
			want: "404",
		},
		{
			name: "Internal Server Error",
			code: http.StatusInternalServerError,
			want: "500",
		},
		{
			name: "Created",
			code: http.StatusCreated,
			want: "201",
		},
		{
			name: "Zero value (no WriteHeader called)",
			code: 0,
			want: "0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ww := &wrapResponseWriter{
				ResponseWriter: httptest.NewRecorder(),
				code:           tt.code,
			}
			assert.Equal(t, tt.want, ww.Status())
		})
	}
}

func TestMiddleware(t *testing.T) {
	tests := []struct {
		name          string
		method        string
		path          string
		handlerStatus int
	}{
		{
			name:          "GET 200",
			method:        http.MethodGet,
			path:          "/api/users",
			handlerStatus: http.StatusOK,
		},
		{
			name:          "POST 201",
			method:        http.MethodPost,
			path:          "/api/users",
			handlerStatus: http.StatusCreated,
		},
		{
			name:          "GET 404 with ULID path cleaned",
			method:        http.MethodGet,
			path:          "/api/users/01ARZ3NDEKTSV4RRFFQ69G5FAV",
			handlerStatus: http.StatusNotFound,
		},
		{
			name:          "DELETE 204",
			method:        http.MethodDelete,
			path:          "/api/items/01BX5ZZKBKACTAV9WEVGEMMVRZ",
			handlerStatus: http.StatusNoContent,
		},
		{
			name:          "GET 500",
			method:        http.MethodGet,
			path:          "/api/health",
			handlerStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New(zap.NewNop())

			handler := Middleware(m)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.handlerStatus)
			}))

			req := httptest.NewRequestWithContext(context.Background(), tt.method, tt.path, nil)
			rec := httptest.NewRecorder()

			assert.NotPanics(t, func() {
				handler.ServeHTTP(rec, req)
			})

			assert.Equal(t, tt.handlerStatus, rec.Code)
		})
	}
}

func TestMiddleware_NilMetrics(t *testing.T) {
	tests := []struct {
		name string
		m    *Metrics
	}{
		{
			name: "All nil fields does not panic",
			m:    &Metrics{log: zap.NewNop()},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := Middleware(tt.m)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()

			assert.NotPanics(t, func() {
				handler.ServeHTTP(rec, req)
			})
		})
	}
}
