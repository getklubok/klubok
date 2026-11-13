package noindex

import "net/http"

// Handler sets X-Robots-Tag: noindex on every response.
func Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Robots-Tag", "noindex")
		next.ServeHTTP(w, r)
	})
}
