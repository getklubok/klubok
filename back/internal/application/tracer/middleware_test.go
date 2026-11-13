package tracer

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"klubok/internal/domain"
)

// setupTestProvider подменяет глобальный TracerProvider на SDK с
// in-memory экспортером и возвращает функцию восстановления.
func setupTestProvider(t *testing.T) *tracetest.SpanRecorder {
	t.Helper()
	recorder := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	previous := otel.GetTracerProvider()
	otel.SetTracerProvider(provider)
	t.Cleanup(func() { otel.SetTracerProvider(previous) })

	return recorder
}

func findAttr(attrs []attribute.KeyValue, key attribute.Key) (attribute.Value, bool) {
	for _, a := range attrs {
		if a.Key == key {
			return a.Value, true
		}
	}

	return attribute.Value{}, false
}

func TestMiddleware_ClientAddressMasking(t *testing.T) {
	tests := []struct {
		name   string
		masker domain.IPMasker
		remote string
		want   string
	}{
		{
			name:   "Masker disabled - full client address",
			masker: domain.IPMasker{Enabled: false, V4Octets: 3, V6Octets: 7},
			remote: "1.2.3.4:5678",
			want:   "1.2.3.4:5678",
		},
		{
			name:   "Masker enabled - truncated ipv4 client address",
			masker: domain.IPMasker{Enabled: true, V4Octets: 3, V6Octets: 7},
			remote: "1.2.3.4:5678",
			want:   "1.2.3.x",
		},
		{
			name:   "Masker enabled - truncated ipv6 client address",
			masker: domain.IPMasker{Enabled: true, V4Octets: 3, V6Octets: 7},
			remote: "[2001:db8:85a3:0:0:8a2e:370:7334]:443",
			want:   "2001:db8:85a3:0:0:8a2e:370:x",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := setupTestProvider(t)

			handler := Middleware("test/http", tt.masker)(http.NotFoundHandler())
			req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/any", nil)
			req.RemoteAddr = tt.remote
			handler.ServeHTTP(httptest.NewRecorder(), req)
			spans := recorder.Ended()
			assert.NotEmpty(t, spans)

			attrs := spans[0].Attributes()
			got, found := findAttr(attrs, "client.address")
			assert.True(t, found, "client.address attribute not found")
			assert.Equal(t, tt.want, got.AsString())
		})
	}
}
