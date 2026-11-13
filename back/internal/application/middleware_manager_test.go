package application

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"klubok/internal/application/zapctx"
)

const testRequestID = "test-request-id"

func TestSetupAPIMiddlewaresRequestContext(t *testing.T) {
	tests := []struct {
		name         string
		realIP       string
		wantClientIP string
	}{
		{
			name: "RequestID доступен хендлеру и zapctx",
		},
		{
			name:         "клиентский IP из X-Real-IP доступен хендлеру",
			realIP:       "203.0.113.7",
			wantClientIP: "203.0.113.7",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				gotRequestID string
				gotZapID     string
				gotClientIP  string
			)

			manager := NewMiddlewareManager(&Configuration{}, passthroughMiddleware, nil)
			handler := manager.SetupAPIMiddlewares(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotRequestID = chimiddleware.GetReqID(r.Context())
				gotZapID = zapRequestID(r.Context())
				gotClientIP = chimiddleware.GetClientIP(r.Context())

				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/health", nil)
			req.Header.Set(chimiddleware.RequestIDHeader, testRequestID)
			if tt.realIP != "" {
				req.Header.Set("X-Real-IP", tt.realIP)
			}

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, testRequestID, gotRequestID)
			assert.Equal(t, testRequestID, gotZapID)
			assert.Equal(t, tt.wantClientIP, gotClientIP)
		})
	}
}

func TestKeyByClientIP(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		headers    map[string]string
		want       string
	}{
		{
			name:       "X-Real-IP имеет приоритет",
			remoteAddr: "192.0.2.1:1234",
			headers:    map[string]string{"X-Real-IP": "203.0.113.7"},
			want:       "203.0.113.7",
		},
		{
			name:       "без X-Real-IP берётся RemoteAddr",
			remoteAddr: "192.0.2.1:1234",
			want:       "192.0.2.1",
		},
		{
			name:       "True-Client-IP и X-Forwarded-For не используются",
			remoteAddr: "192.0.2.1:1234",
			headers: map[string]string{
				"True-Client-IP":  "203.0.113.9",
				"X-Forwarded-For": "203.0.113.9",
			},
			want: "192.0.2.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				got string
				err error
			)

			handler := chimiddleware.ClientIPFromHeader("X-Real-IP")(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
				got, err = keyByClientIP(r)
			}))

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/health", nil)
			req.RemoteAddr = tt.remoteAddr
			for name, value := range tt.headers {
				req.Header.Set(name, value)
			}

			handler.ServeHTTP(httptest.NewRecorder(), req)

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRouterTimeout(t *testing.T) {
	tests := []struct {
		name         string
		writeTimeout time.Duration
		want         time.Duration
	}{
		{
			name:         "меньше WriteTimeout на запас",
			writeTimeout: 30 * time.Second,
			want:         29 * time.Second,
		},
		{
			name:         "WriteTimeout выключен",
			writeTimeout: 0,
			want:         defaultRouterTimeout,
		},
		{
			name:         "отрицательный WriteTimeout",
			writeTimeout: -time.Second,
			want:         defaultRouterTimeout,
		},
		{
			name:         "WriteTimeout равен запасу",
			writeTimeout: time.Second,
			want:         500 * time.Millisecond,
		},
		{
			name:         "WriteTimeout меньше запаса",
			writeTimeout: 500 * time.Millisecond,
			want:         250 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewMiddlewareManager(
				&Configuration{HTTP: HTTPServer{WriteTimeout: tt.writeTimeout}},
				passthroughMiddleware,
				nil,
			)

			assert.Equal(t, tt.want, manager.routerTimeout())
		})
	}
}

func TestSetupEmbedMiddlewares(t *testing.T) {
	var gotRequestID string

	manager := NewMiddlewareManager(&Configuration{}, markerMiddleware("X-Metrics", "on"), markerMiddleware("X-Tracer", "on"))
	handler := manager.SetupEmbedMiddlewares(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotRequestID = chimiddleware.GetReqID(r.Context())

		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/embed/token", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set(chimiddleware.RequestIDHeader, testRequestID)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "noindex", rec.Header().Get("X-Robots-Tag"))
	assert.Equal(t, "*", rec.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "on", rec.Header().Get("X-Metrics"))
	assert.Empty(t, rec.Header().Get("X-Tracer"), "трейсинг embed отключён из-за кардинальности токенов")
	assert.Equal(t, testRequestID, gotRequestID)
}

func TestSetupWebMiddlewaresRecoversPanic(t *testing.T) {
	manager := NewMiddlewareManager(&Configuration{}, passthroughMiddleware, nil)
	handler := manager.SetupWebMiddlewares(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil))

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestSetupAPIMiddlewaresObservabilityOrder(t *testing.T) {
	var order []string
	record := func(name string) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, name)
				next.ServeHTTP(w, r)
			})
		}
	}

	manager := NewMiddlewareManager(&Configuration{}, record("metrics"), record("tracer"))
	handler := manager.SetupAPIMiddlewares(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		order = append(order, "handler")
	}))

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/health", nil))

	assert.Equal(t, []string{"metrics", "tracer", "handler"}, order)
}

func TestRequestLoggerSkipsPaths(t *testing.T) {
	originalLogger := chimiddleware.DefaultLogger
	t.Cleanup(func() { chimiddleware.DefaultLogger = originalLogger })

	var logged int

	chimiddleware.DefaultLogger = func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logged++
			next.ServeHTTP(w, r)
		})
	}

	handler := requestLogger(healthLivePath, healthReadyPath)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	tests := []struct {
		name string
		path string
		want int
	}{
		{
			name: "liveness не логируется",
			path: healthLivePath,
		},
		{
			name: "readiness не логируется",
			path: healthReadyPath,
		},
		{
			name: "обычный запрос логируется",
			path: "/diagrams",
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logged = 0

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, tt.path, nil))

			assert.Equal(t, tt.want, logged)
		})
	}
}

func passthroughMiddleware(next http.Handler) http.Handler {
	return next
}

func markerMiddleware(header, value string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set(header, value)
			next.ServeHTTP(w, r)
		})
	}
}

func zapRequestID(ctx context.Context) string {
	for _, field := range zapctx.GetFields(ctx) {
		if field.Key == "request_id" {
			return field.String
		}
	}

	return ""
}
