package domain

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/stretchr/testify/assert"
)

func TestIPFromContext(t *testing.T) {
	// contextWithIP runs the request through chi's ClientIPFromHeader
	// middleware — the same code path used by production middleware
	// (middleware_manager.go), so the test exercises the real context key.
	// The context is captured INSIDE the handler: a middleware builds a new
	// request with WithContext and the caller's req pointer is not updated.
	contextWithIP := func(ip IP) func() context.Context {
		return func() context.Context {
			var req *http.Request
			handler := chimiddleware.ClientIPFromHeader("X-Real-IP")(http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) { req = r },
			))
			httpReq := httptest.NewRequestWithContext(context.Background(), "GET", "/", nil)
			if ip != "" {
				httpReq.Header.Set("X-Real-IP", ip.String())
			}
			handler.ServeHTTP(httptest.NewRecorder(), httpReq)
			return req.Context()
		}
	}

	tests := []struct {
		name    string
		ctxFunc func() context.Context
		want    IP
	}{
		{
			name:    "should return IP",
			ctxFunc: contextWithIP("1.1.1.1"),
			want:    IP("1.1.1.1"),
		},
		{
			name:    "should return empty IP if not found in context",
			ctxFunc: context.Background,
			want:    "",
		},
		{
			name:    "should return empty IP if header is empty",
			ctxFunc: contextWithIP(""),
			want:    "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IPFromContext(tt.ctxFunc())
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIPMasker_Mask(t *testing.T) {
	tests := []struct {
		name   string
		masker IPMasker
		addr   string
		want   string
	}{
		{
			name:   "disabled returns input unchanged",
			masker: IPMasker{Enabled: false, V4Octets: 3, V6Octets: 7},
			addr:   "1.2.3.4",
			want:   "1.2.3.4",
		},
		{
			name:   "disabled returns ipv6 input unchanged",
			masker: IPMasker{Enabled: false, V4Octets: 3, V6Octets: 7},
			addr:   "[2001:db8::1]:443",
			want:   "[2001:db8::1]:443",
		},
		{
			name:   "ipv4 plain, default 3 octets",
			masker: IPMasker{Enabled: true, V4Octets: 3, V6Octets: 7},
			addr:   "1.2.3.4",
			want:   "1.2.3.x",
		},
		{
			name:   "ipv4 with port, default 3 octets",
			masker: IPMasker{Enabled: true, V4Octets: 3, V6Octets: 7},
			addr:   "1.2.3.4:5678",
			want:   "1.2.3.x",
		},
		{
			name:   "ipv4 one octet",
			masker: IPMasker{Enabled: true, V4Octets: 1, V6Octets: 7},
			addr:   "1.2.3.4",
			want:   "1.x.x.x",
		},
		{
			masker: IPMasker{Enabled: true, V4Octets: 4, V6Octets: 7},
			addr:   "1.2.3.4",
			want:   "1.2.3.4",
		},
		{
			name:   "ipv4 full octets with port drops port",
			masker: IPMasker{Enabled: true, V4Octets: 4, V6Octets: 7},
			addr:   "1.2.3.4:5678",
			want:   "1.2.3.4",
		},
		{
			name:   "ipv6 bracketed with port, default 7 groups",
			masker: IPMasker{Enabled: true, V4Octets: 3, V6Octets: 7},
			addr:   "[2001:db8::1]:443",
			want:   "2001:db8:0:0:0:0:0:x",
		},
		{
			name:   "ipv6 bare, default 7 groups",
			masker: IPMasker{Enabled: true, V4Octets: 3, V6Octets: 7},
			addr:   "2001:db8::1",
			want:   "2001:db8:0:0:0:0:0:x",
		},
		{
			name:   "ipv6 one group",
			masker: IPMasker{Enabled: true, V4Octets: 3, V6Octets: 1},
			addr:   "2001:db8::1",
			want:   "2001:x:x:x:x:x:x:x",
		},
		{
			name:   "ipv6 three groups",
			masker: IPMasker{Enabled: true, V4Octets: 3, V6Octets: 3},
			addr:   "2001:db8:85a3::8a2e:370:7334",
			want:   "2001:db8:85a3:x:x:x:x:x",
		},
		{
			name:   "ipv6 seven groups keeps seventh",
			masker: IPMasker{Enabled: true, V4Octets: 3, V6Octets: 7},
			addr:   "2001:db8:85a3:0:0:8a2e:370:7334",
			want:   "2001:db8:85a3:0:0:8a2e:370:x",
		},
		{
			name:   "ipv6 all eight groups keeps all",
			masker: IPMasker{Enabled: true, V4Octets: 3, V6Octets: 8},
			addr:   "2001:db8:85a3:0:0:8a2e:370:7334",
			want:   "2001:db8:85a3:0:0:8a2e:370:7334",
		},
		{
			name:   "ipv6 loopback expands to eight groups",
			masker: IPMasker{Enabled: true, V4Octets: 3, V6Octets: 7},
			addr:   "::1",
			want:   "0:0:0:0:0:0:0:x",
		},
		{
			name:   "ipv6 unspecified expands to eight groups",
			masker: IPMasker{Enabled: true, V4Octets: 3, V6Octets: 7},
			addr:   "::",
			want:   "0:0:0:0:0:0:0:x",
		},
		{
			name:   "garbage string returned as is",
			masker: IPMasker{Enabled: true, V4Octets: 3, V6Octets: 7},
			addr:   "not-an-ip",
			want:   "not-an-ip",
		},
		{
			name:   "empty string returned as is",
			masker: IPMasker{Enabled: true, V4Octets: 3, V6Octets: 7},
			addr:   "",
			want:   "",
		},
		{
			name:   "out of range octets fall back to defaults",
			masker: IPMasker{Enabled: true, V4Octets: 0, V6Octets: 9},
			addr:   "1.2.3.4:80",
			want:   "1.2.3.x",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.masker.Mask(tt.addr)
			assert.Equal(t, tt.want, got)
		})
	}
}

func FuzzIPMasker_Mask(f *testing.F) {
	seeds := []string{
		"1.2.3.4", "1.2.3.4:5678", "[2001:db8::1]:443", "2001:db8::1",
		"::1", "::", "not-an-ip", "", "1.2.3.4:80",
		"255.255.255.255:65535", "[::ffff:1.2.3.4]:443", "0.0.0.0",
		"[2001:0db8:0000:0000:0000:8a2e:0370:7334]:8080", "1.2.3",
		"999.1.1.1", "[1:2:3:4:5:6:7:8]:9", ":8080", "1.2.3.4:",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	masker := IPMasker{Enabled: true, V4Octets: 3, V6Octets: 7}

	f.Fuzz(func(t *testing.T, addr string) {
		assert.NotPanics(t, func() {
			assertMaskedInvariants(t, masker, addr)

			// Выключенный маскер всегда возвращает вход.
			disabled := IPMasker{Enabled: false, V4Octets: 3, V6Octets: 7}
			assert.Equal(t, addr, disabled.Mask(addr))
		})
	})
}

// assertMaskedInvariants проверяет форму выхода: он либо равен входу
// (неразбираемая строка), либо является результатом усечения разобранного
// адреса.
func assertMaskedInvariants(t *testing.T, masker IPMasker, addr string) {
	t.Helper()

	out := masker.Mask(addr)

	switch {
	case out == addr:
		// вход возвращён как есть
	case strings.Count(out, ":") == 7 && !strings.Contains(out, "."):
		// IPv6-форма: ровно 8 непустых групп, хвост замаскирован как "x".
		groups := strings.Split(out, ":")
		assert.Len(t, groups, 8)
		for i, g := range groups {
			assert.NotEmpty(t, g)
			if i >= masker.V6Octets {
				assert.Equal(t, "x", g)
			}
		}
	case strings.Count(out, ".") == 3:
		// IPv4-форма: последние октеты заменены на "x".
		octets := strings.Split(out, ".")
		assert.Len(t, octets, 4)
		for i, o := range octets {
			if i >= masker.V4Octets {
				assert.Equal(t, "x", o)
			} else {
				assert.NotEqual(t, "x", o)
			}
		}
	default:
		t.Fatalf("unexpected output form for input %q: %q", addr, out)
	}
}
