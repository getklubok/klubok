package inject

import (
	"net/http"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	"klubok/internal/application/zapctx"
)

func SetRequestIDToContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := chimiddleware.GetReqID(r.Context())
		if requestID != "" {
			r = r.WithContext(
				zapctx.WithFields(r.Context(), zap.String("request_id", requestID)),
			)
		}

		next.ServeHTTP(w, r)
	})
}
