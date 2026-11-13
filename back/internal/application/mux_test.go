package application

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testIndexBody = "spa-index"
	testAssetBody = "app-js"
)

// testStaticDir создаёт каталог, повторяющий раскладку собранного front.
func testStaticDir(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(dir, "index.html"), []byte(testIndexBody), 0o600))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "assets"), 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "assets", "app.js"), []byte(testAssetBody), 0o600))

	return dir
}

// testServer подменяет ogen-сервер: отдаёт путь, который до него дошёл.
func testServer() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Server-Path", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	})
}

// testMiddleware помечает запросы, дошедшие до определённого набора мидлварей.
func testMiddleware(name string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Middleware", name)
			next.ServeHTTP(w, r)
		})
	}
}

func serve(t *testing.T, handler http.Handler, target string, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, target, nil)
	for name, value := range headers {
		req.Header.Set(name, value)
	}

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	return rec
}

func TestNewMuxRouting(t *testing.T) {
	staticDir := testStaticDir(t)
	privacyPath := filepath.Join(t.TempDir(), "privacy.md")
	require.NoError(t, os.WriteFile(privacyPath, []byte("policy-body"), 0o600))

	handler := newMux(testServer(), muxMiddlewares{
		api:   testMiddleware("api"),
		embed: testMiddleware("embed"),
		web:   testMiddleware("web"),
	}, staticDir, privacyPath)

	tests := []struct {
		name           string
		target         string
		wantStatus     int
		wantServerPath string
		wantBody       string
		wantMiddleware string
	}{
		{
			name:           "API: префикс /api срезается (liveness)",
			target:         "/api/health/live",
			wantStatus:     http.StatusOK,
			wantServerPath: "/health/live",
			wantMiddleware: "api",
		},
		{
			name:           "API: префикс /api срезается (readiness)",
			target:         "/api/health/ready",
			wantStatus:     http.StatusOK,
			wantServerPath: "/health/ready",
			wantMiddleware: "api",
		},
		{
			name:           "API: /api/ доходит до сервера как /",
			target:         "/api/",
			wantStatus:     http.StatusOK,
			wantServerPath: "/",
			wantMiddleware: "api",
		},
		{
			name:           "embed API: путь срезается и получает свой набор",
			target:         "/api/embed/token",
			wantStatus:     http.StatusOK,
			wantServerPath: "/embed/token",
			wantMiddleware: "embed",
		},
		{
			name:           "robots.txt получает web-набор",
			target:         "/robots.txt",
			wantStatus:     http.StatusOK,
			wantBody:       "Disallow: /embed/",
			wantMiddleware: "web",
		},
		{
			name:           "корень отдаёт SPA",
			target:         "/",
			wantStatus:     http.StatusOK,
			wantBody:       testIndexBody,
			wantMiddleware: "web",
		},
		{
			name:           "embed-страница отдаёт SPA",
			target:         "/embed/",
			wantStatus:     http.StatusOK,
			wantBody:       testIndexBody,
			wantMiddleware: "web",
		},
		{
			name:           "политика раздаётся из каталога образа, а не из статики",
			target:         "/privacy.md",
			wantStatus:     http.StatusOK,
			wantBody:       "policy-body",
			wantMiddleware: "web",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := serve(t, handler, tt.target, nil)

			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.Equal(t, tt.wantServerPath, rec.Header().Get("X-Server-Path"))
			assert.Equal(t, tt.wantMiddleware, rec.Header().Get("X-Middleware"))

			if tt.wantBody != "" {
				assert.Contains(t, rec.Body.String(), tt.wantBody)
			}
		})
	}
}

func TestStaticHandler(t *testing.T) {
	handler := staticHandler(testStaticDir(t))

	tests := []struct {
		name             string
		target           string
		wantStatus       int
		wantBody         string
		wantCacheControl string
		wantCSP          string
		wantFrameOptions string
		wantNoCache      bool
	}{
		{
			name:             "корень отдаёт index.html без кэша",
			target:           "/",
			wantStatus:       http.StatusOK,
			wantBody:         testIndexBody,
			wantCacheControl: indexCacheControl,
			wantCSP:          defaultContentSecurityPolicy,
			wantFrameOptions: "DENY",
		},
		{
			name:             "SPA-fallback для неизвестного пути без расширения",
			target:           "/some/route",
			wantStatus:       http.StatusOK,
			wantBody:         testIndexBody,
			wantCacheControl: indexCacheControl,
			wantCSP:          defaultContentSecurityPolicy,
			wantFrameOptions: "DENY",
		},
		{
			name:             "embed-страница отдаёт index.html без кэша и встраиваема",
			target:           "/embed/",
			wantStatus:       http.StatusOK,
			wantBody:         testIndexBody,
			wantCacheControl: noCacheControl,
			wantCSP:          embedContentSecurityPolicy,
			wantNoCache:      true,
		},
		{
			name:             "вложенный embed-путь отдаёт index.html",
			target:           "/embed/token/abc",
			wantStatus:       http.StatusOK,
			wantBody:         testIndexBody,
			wantCacheControl: noCacheControl,
			wantCSP:          embedContentSecurityPolicy,
			wantNoCache:      true,
		},
		{
			name:             "ассеты кэшируются",
			target:           "/assets/app.js",
			wantStatus:       http.StatusOK,
			wantBody:         testAssetBody,
			wantCacheControl: staticCacheControl,
			wantCSP:          defaultContentSecurityPolicy,
			wantFrameOptions: "DENY",
		},
		{
			name:             "отсутствующий файл с расширением — 404",
			target:           "/assets/missing.js",
			wantStatus:       http.StatusNotFound,
			wantCSP:          defaultContentSecurityPolicy,
			wantFrameOptions: "DENY",
		},
		{
			name:             "отсутствующий файл в корне — 404",
			target:           "/missing.svg",
			wantStatus:       http.StatusNotFound,
			wantCSP:          defaultContentSecurityPolicy,
			wantFrameOptions: "DENY",
		},
		{
			// Стандартная канонизация http.FileServer: /index.html → ./.
			name:             "index.html канонизируется в корень",
			target:           "/index.html",
			wantStatus:       http.StatusMovedPermanently,
			wantCacheControl: indexCacheControl,
			wantCSP:          defaultContentSecurityPolicy,
			wantFrameOptions: "DENY",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := serve(t, handler, tt.target, nil)

			assert.Equal(t, tt.wantStatus, rec.Code)

			if tt.wantBody != "" {
				assert.Contains(t, rec.Body.String(), tt.wantBody)
			}

			assert.Equal(t, tt.wantCacheControl, rec.Header().Get("Cache-Control"))
			assert.Equal(t, tt.wantNoCache, rec.Header().Get("Pragma") == "no-cache")
			assert.Equal(t, tt.wantCSP, rec.Header().Get("Content-Security-Policy"))
			assert.Equal(t, tt.wantFrameOptions, rec.Header().Get("X-Frame-Options"))
		})
	}
}

func TestStaticHandlerMissingDir(t *testing.T) {
	handler := staticHandler(filepath.Join(t.TempDir(), "static"))

	tests := []struct {
		name   string
		target string
	}{
		{name: "корень", target: "/"},
		{name: "embed-страница", target: "/embed/"},
		{name: "файл с расширением", target: "/assets/app.js"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := serve(t, handler, tt.target, nil)

			assert.Equal(t, http.StatusNotFound, rec.Code)
		})
	}
}

func TestPrivacyHandler(t *testing.T) {
	dir := t.TempDir()
	privacyPath := filepath.Join(dir, "privacy.md")
	require.NoError(t, os.WriteFile(privacyPath, []byte("policy v1"), 0o600))

	handler := privacyHandler(privacyPath)

	tests := []struct {
		name       string
		path       string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "файл раздаётся как markdown без кэша",
			path:       privacyPath,
			wantStatus: http.StatusOK,
			wantBody:   "policy v1",
		},
		{
			name:       "файл не подмонтирован — 404",
			path:       filepath.Join(dir, "missing.md"),
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := serve(t, privacyHandler(tt.path), "/privacy.md", nil)

			assert.Equal(t, tt.wantStatus, rec.Code)

			if tt.wantBody == "" {
				return
			}

			// Успешная раздача: markdown, anti-cache и запрет встраивания.
			assert.Equal(t, "text/markdown; charset=utf-8", rec.Header().Get("Content-Type"))
			assert.Equal(t, noCacheControl, rec.Header().Get("Cache-Control"))
			assert.Equal(t, "no-cache", rec.Header().Get("Pragma"))
			assert.Equal(t, defaultContentSecurityPolicy, rec.Header().Get("Content-Security-Policy"))
			assert.Equal(t, "DENY", rec.Header().Get("X-Frame-Options"))
			assert.Contains(t, rec.Body.String(), tt.wantBody)
		})
	}

	t.Run("обновлённый файл виден без перезапуска", func(t *testing.T) {
		require.NoError(t, os.WriteFile(privacyPath, []byte("policy v2"), 0o600))

		rec := serve(t, handler, "/privacy.md", nil)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "policy v2")
	})
}

// FuzzStaticHandler проверяет, что раздача статики не выпускает наружу файлы
// за пределами каталога статики и не отвечает серверными ошибками.
func FuzzStaticHandler(f *testing.F) {
	for _, seed := range []string{
		"/", "/index.html", "/assets/app.js", "/embed/", "/assets/missing.js",
		"/../../etc/passwd", "/a/../b", "/static/../secret", "/%2e%2e/secret",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, target string) {
		dir := testStaticDir(t)
		require.NoError(t, os.WriteFile(filepath.Join(filepath.Dir(dir), "secret"), []byte("SECRET"), 0o600))

		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
		req.URL.Path = target

		rec := httptest.NewRecorder()
		staticHandler(dir).ServeHTTP(rec, req)

		assert.NotContains(t, rec.Body.String(), "SECRET")
		assert.Less(t, rec.Code, http.StatusInternalServerError)
	})
}
