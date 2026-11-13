package inject

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"klubok/internal/application/zapctx"
)

func TestSetRequestIDToContext(t *testing.T) {
	tests := []struct {
		name          string
		setupContext  func(r *http.Request) *http.Request
		wantRequestID string
		wantFields    int
	}{
		{
			name: "should set request_id to zapctx fields",
			setupContext: func(r *http.Request) *http.Request {
				ctx := context.WithValue(r.Context(), chimiddleware.RequestIDKey, "test-request-id-123")

				return r.WithContext(ctx)
			},
			wantRequestID: "test-request-id-123",
			wantFields:    1,
		},
		{
			name: "should not set fields when request_id is empty",
			setupContext: func(r *http.Request) *http.Request {
				return r
			},
			wantRequestID: "",
			wantFields:    0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotFields []zap.Field

			handler := SetRequestIDToContext(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
				gotFields = zapctx.GetFields(r.Context())
			}))

			req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
			req = tt.setupContext(req)

			handler.ServeHTTP(httptest.NewRecorder(), req)

			assert.Len(t, gotFields, tt.wantFields)

			if tt.wantRequestID != "" {
				assert.Len(t, gotFields, 1)
				assert.Equal(t, "request_id", gotFields[0].Key)
				assert.Equal(t, tt.wantRequestID, gotFields[0].String)
			}
		})
	}
}
