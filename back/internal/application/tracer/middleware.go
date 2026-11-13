package tracer

import (
	"fmt"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
	"go.opentelemetry.io/otel/trace"

	"klubok/internal/domain"
)

func Middleware(tracerName string, ipMasker domain.IPMasker, excludedPaths ...string) func(http.Handler) http.Handler {
	tracer := otel.Tracer(tracerName)

	excluded := make(map[string]struct{}, len(excludedPaths))
	for _, path := range excludedPaths {
		excluded[path] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, ok := excluded[r.URL.Path]; ok {
				next.ServeHTTP(w, r)
				return
			}

			ctx, span := tracer.Start(r.Context(), fmt.Sprintf("%s %s", r.Method, r.URL.Path),
				trace.WithSpanKind(trace.SpanKindServer),
				trace.WithAttributes(
					semconv.HTTPRequestMethodKey.String(r.Method),
					semconv.URLPath(r.URL.Path),
					semconv.URLScheme(r.URL.Scheme),
					semconv.ServerAddress(r.Host),
					semconv.ClientAddress(ipMasker.Mask(r.RemoteAddr)),
					semconv.NetworkProtocolName("http"),
					semconv.NetworkProtocolVersion(r.Proto),
					semconv.UserAgentOriginal(r.UserAgent()),
				),
			)
			defer span.End()

			srw := &statusResponseWriter{ResponseWriter: w}
			next.ServeHTTP(srw, r.WithContext(ctx))

			span.SetAttributes(semconv.HTTPResponseStatusCode(srw.status))

			if srw.status >= http.StatusInternalServerError {
				span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", srw.status))
			}
		})
	}
}

type statusResponseWriter struct {
	http.ResponseWriter

	status int
}

func (w *statusResponseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}
