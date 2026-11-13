package cookie

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"klubok/internal/domain"
)

func TestCookieService_Create(t *testing.T) {
	tests := []struct {
		name    string
		srvFunc func(t *testing.T) *Service
		v       domain.CookieID
		want    string
	}{
		{
			name: "Should create cookie",
			srvFunc: func(t *testing.T) *Service {
				t.Helper()

				return &Service{
					cookieConfig: domain.CookieConfig{
						Name:        "GIG_CID",
						ExpiresDays: 1,
						Domain:      "localhost",
						Secure:      true,
					},
					nowFunc: func() time.Time { return time.Date(2024, 12, 18, 14, 15, 0, 0, time.UTC) },
				}
			},
			v:    domain.CookieID("01JHHXTVAMTYVE7Q3H6VRP7206"),
			want: "GIG_CID=01JHHXTVAMTYVE7Q3H6VRP7206; Path=/; Domain=localhost; Expires=Thu, 19 Dec 2024 14:15:00 GMT; HttpOnly; Secure; SameSite=Strict",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.srvFunc(t).Create(tt.v)
			assert.Equal(t, tt.want, got.String())
		})
	}
}

func TestCookieService_Get(t *testing.T) {
	tests := []struct {
		name    string
		srvFunc func(t *testing.T) *Service
		r       *http.Request
		want    domain.CookieID
		wantErr error
	}{
		{
			name: "Should get cookie",
			srvFunc: func(t *testing.T) *Service {
				t.Helper()

				return &Service{
					cookieConfig: domain.CookieConfig{
						Name:        "GIG_CID",
						ExpiresDays: 1,
						Domain:      "localhost",
						Secure:      true,
					},
					nowFunc: func() time.Time { return time.Date(2024, 12, 18, 14, 15, 0, 0, time.UTC) },
				}
			},
			r: func() *http.Request {
				r := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
				r.Header.Set("Cookie", "GIG_CID=01JHHXTVAMTYVE7Q3H6VRP7206")
				return r
			}(),
			want: domain.CookieID("01JHHXTVAMTYVE7Q3H6VRP7206"),
		},
		{
			name: "Should return error - no cookie",
			srvFunc: func(t *testing.T) *Service {
				t.Helper()

				return &Service{
					cookieConfig: domain.CookieConfig{
						Name:        "GIG_CID",
						ExpiresDays: 1,
						Domain:      "localhost",
						Secure:      true,
					},
					nowFunc: func() time.Time { return time.Date(2024, 12, 18, 14, 15, 0, 0, time.UTC) },
				}
			},
			r:       httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil),
			wantErr: http.ErrNoCookie,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.srvFunc(t).Get(tt.r)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestCookieService_Clear(t *testing.T) {
	tests := []struct {
		name    string
		srvFunc func(t *testing.T) *Service
		want    string
	}{
		{
			name: "Should clear cookie",
			srvFunc: func(t *testing.T) *Service {
				t.Helper()

				return &Service{
					cookieConfig: domain.CookieConfig{
						Name:        "GIG_CID",
						ExpiresDays: 1,
						Domain:      "localhost",
						Secure:      true,
					},
				}
			},
			want: "GIG_CID=; Path=/; Domain=localhost; Expires=Thu, 01 Jan 1970 00:00:00 GMT; Max-Age=0; HttpOnly; Secure; SameSite=Strict",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.srvFunc(t).Clear()
			assert.Equal(t, tt.want, got.String())
		})
	}
}
