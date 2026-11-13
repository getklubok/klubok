package noindex

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		nextStatus int
		nextBody   string
	}{
		{
			name:       "should set noindex header on GET",
			method:     http.MethodGet,
			nextStatus: http.StatusOK,
			nextBody:   "ok",
		},
		{
			name:       "should set noindex header on POST",
			method:     http.MethodPost,
			nextStatus: http.StatusCreated,
			nextBody:   `{"id": 1}`,
		},
		{
			name:       "should set noindex header on error status",
			method:     http.MethodGet,
			nextStatus: http.StatusNotFound,
			nextBody:   "not found",
		},
		{
			name:       "should set noindex header on HEAD",
			method:     http.MethodHead,
			nextStatus: http.StatusOK,
			nextBody:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var nextCalled bool

			handler := Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true

				assert.Equal(t, tt.method, r.Method)

				w.WriteHeader(tt.nextStatus)
				if tt.nextBody != "" {
					_, _ = w.Write([]byte(tt.nextBody))
				}
			}))

			req := httptest.NewRequestWithContext(t.Context(), tt.method, "/", nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			require.True(t, nextCalled, "next handler must be called")

			assert.Equal(t, "noindex", rec.Header().Get("X-Robots-Tag"))
			assert.Equal(t, tt.nextStatus, rec.Code)
			if tt.nextBody != "" {
				assert.Equal(t, tt.nextBody, rec.Body.String())
			}
		})
	}
}

func TestHandler_HeaderNotSetByNext(t *testing.T) {
	handler := Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "noindex", w.Header().Get("X-Robots-Tag"))
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
}
