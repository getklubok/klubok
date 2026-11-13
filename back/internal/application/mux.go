package application

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// Пути раскладки образа приложения, см. build/Dockerfile. Передаются в newMux
// параметрами, чтобы в тестах подставить t.TempDir().
const (
	imageStaticDir   = "/app/static"
	imagePrivacyPath = "/app/privacy.md"
)

// CSP перенесены из deploy/nginx.conf; из connect-src убран явный домен
// развёртывания — 'self' покрывает http/https/ws/wss того же origin.
const (
	embedContentSecurityPolicy = "default-src 'self'; script-src 'self'; script-src-elem 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self' data:; font-src 'self'; object-src 'none'; media-src 'none'; frame-src 'self'; child-src 'none'; form-action 'none'; worker-src 'self'; manifest-src 'self'; frame-ancestors http: https: file:"

	defaultContentSecurityPolicy = "default-src 'self'; script-src 'self'; script-src-elem 'self' https://smartcaptcha.cloud.yandex.ru; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self' data:; font-src 'self'; object-src 'none'; media-src 'none'; frame-src 'self' https://smartcaptcha.cloud.yandex.ru; child-src 'none'; form-action 'none'; worker-src 'self'; manifest-src 'self'; frame-ancestors 'none'"
)

const (
	// indexCacheControl — index.html должен перечитываться, имена файлов в /assets хешированные.
	indexCacheControl  = "no-cache"
	staticCacheControl = "public, max-age=86400"

	// noCacheControl — anti-cache embed-страниц и файла политики.
	noCacheControl = "no-store, no-cache, must-revalidate, proxy-revalidate"
)

// muxMiddlewares — наборы мидлварей для групп маршрутов.
type muxMiddlewares struct {
	api   func(http.Handler) http.Handler
	embed func(http.Handler) http.Handler
	web   func(http.Handler) http.Handler
}

// newMux собирает HTTP-роутер приложения: API под /api/, SPA из staticDir, файл
// политики по /privacy.md. Статика и политика получают только web-набор —
// лимиты и метрики не должны распространяться на файлы.
func newMux(
	srv http.Handler,
	middlewares muxMiddlewares,
	staticDir string,
	privacyPath string,
) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/api/embed/", http.StripPrefix("/api", middlewares.embed(srv)))
	mux.Handle("/robots.txt", middlewares.web(robotsTxtHandler()))
	mux.Handle("/api/", http.StripPrefix("/api", middlewares.api(srv)))
	mux.Handle("/privacy.md", middlewares.web(privacyHandler(privacyPath)))
	mux.Handle("/", middlewares.web(staticHandler(staticDir)))

	return mux
}

func robotsTxtHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("User-agent: *\nDisallow: /embed/\n"))
	})
}

// staticHandler раздаёт собранный front из dir: существующие файлы, index.html
// для путей без расширения (SPA-fallback для / и /embed/) и 404 для остальных.
// Отсутствие каталога (dev-запуск) даёт 404 и не ломает старт.
func staticHandler(dir string) http.Handler {
	files := http.FileServer(http.Dir(dir))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := path.Clean(r.URL.Path)
		embed := isEmbedPage(p)

		if embed {
			writeEmbedPageHeaders(w)
		} else {
			writeDefaultPageHeaders(w)
		}

		info, err := os.Stat(filepath.Join(dir, p))
		if err == nil && !info.IsDir() {
			if p == "/index.html" {
				w.Header().Set("Cache-Control", indexCacheControl)
			} else {
				w.Header().Set("Cache-Control", staticCacheControl)
			}

			files.ServeHTTP(w, r)

			return
		}

		// Каталог или отсутствующий файл с расширением — 404.
		if path.Ext(p) != "" {
			http.NotFound(w, r)

			return
		}

		// SPA-fallback: неизвестный путь без расширения отдаёт index.html.
		if !embed {
			w.Header().Set("Cache-Control", indexCacheControl)
		}

		http.ServeFile(w, r, filepath.Join(dir, "index.html"))
	})
}

// privacyHandler читает файл политики с диска на каждый запрос — обновление
// внешнего volume видно без пересборки и перезапуска контейнера.
func privacyHandler(filePath string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
		writeDefaultPageHeaders(w)
		writeNoCacheHeaders(w)

		http.ServeFile(w, r, filePath)
	})
}

// isEmbedPage: страница embed-режима — /embed/ и вложенные пути.
func isEmbedPage(p string) bool {
	return p == "/embed" || strings.HasPrefix(p, "/embed/")
}

func writeEmbedPageHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Security-Policy", embedContentSecurityPolicy)
	writeNoCacheHeaders(w)
}

func writeDefaultPageHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Security-Policy", defaultContentSecurityPolicy)
	w.Header().Set("X-Frame-Options", "DENY")
}

func writeNoCacheHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", noCacheControl)
	w.Header().Set("Pragma", "no-cache")
}
