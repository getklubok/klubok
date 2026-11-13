package metrics

import (
	"net/http"
	"regexp"
	"strconv"
	"time"
)

var (
	ulidRegex       = regexp.MustCompile(`[0123456789ABCDEFGHJKMNPQRSTVWXYZ]{26}`)
	embedTokenRegex = regexp.MustCompile(`^/embed/[^/]+$`)
)

func Middleware(m *Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			ww := &wrapResponseWriter{ResponseWriter: w}
			m.HTTPRequestsCurrentInc()
			next.ServeHTTP(ww, r)
			m.HTTPRequestsCurrentDec()

			path := cleanPath(r.URL.Path)
			m.HTTPRequestsTotalInc(r.Method, path, ww.Status())
			m.HTTPRequestDurationSecondsObserve(r.Method, path, time.Since(start).Seconds())
		})
	}
}

func cleanPath(s string) string {
	if embedTokenRegex.MatchString(s) {
		return "/embed/:token"
	}

	return ulidRegex.ReplaceAllString(s, ":id")
}

type wrapResponseWriter struct {
	http.ResponseWriter

	code int
}

func (ww *wrapResponseWriter) WriteHeader(code int) {
	ww.code = code
	ww.ResponseWriter.WriteHeader(code)
}

func (ww *wrapResponseWriter) Status() string {
	return strconv.Itoa(ww.code)
}
