package yandex

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// MockTransport - мок транспорта для перехвата запросов.
type MockTransport struct {
	RoundTripFunc func(req *http.Request) (*http.Response, error)
}

// RoundTrip - реализация интерфейса http.RoundTripper.
func (m *MockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.RoundTripFunc(req)
}

func TestCaptchaService_Check(t *testing.T) {
	ctx := t.Context()

	tests := []struct {
		name    string
		srvFunc func(t *testing.T) *CaptchaService
		want    bool
	}{
		{
			name: "Should success check token",
			srvFunc: func(t *testing.T) *CaptchaService {
				t.Helper()

				mockTransport := &MockTransport{
					RoundTripFunc: func(req *http.Request) (*http.Response, error) {
						assert.Equal(t, "https://smartcaptcha.cloud.yandex.ru/validate", req.URL.String())
						assert.Equal(t, http.MethodPost, req.Method)
						assert.Equal(t, "application/x-www-form-urlencoded", req.Header.Get("Content-Type"))

						err := req.ParseForm()
						assert.NoError(t, err)

						assert.Equal(t, "key", req.Form.Get("secret"))
						assert.Equal(t, "token-string", req.Form.Get("token"))
						assert.Equal(t, "0.0.0.0", req.Form.Get("ip"))

						jsonBody := `{
							"status": "ok",
							"message": "success"
						}`

						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(strings.NewReader(jsonBody)),
						}, nil
					},
				}

				return &CaptchaService{
					serverKey: "key",
					log:       zap.NewNop(),
					client: &http.Client{
						Transport: mockTransport,
					},
				}
			},
			want: true,
		},
		{
			name: "Should fail check token",
			srvFunc: func(t *testing.T) *CaptchaService {
				t.Helper()

				mockTransport := &MockTransport{
					RoundTripFunc: func(req *http.Request) (*http.Response, error) {
						jsonBody := `{
							"status": "failed",
							"message": ""
						}`

						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(strings.NewReader(jsonBody)),
						}, nil
					},
				}

				return &CaptchaService{
					serverKey: "key",
					log:       zap.NewNop(),
					client: &http.Client{
						Transport: mockTransport,
					},
				}
			},
			want: false,
		},
		{
			name: "Status 500",
			srvFunc: func(t *testing.T) *CaptchaService {
				t.Helper()

				mockTransport := &MockTransport{
					RoundTripFunc: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: http.StatusInternalServerError,
						}, nil
					},
				}

				return &CaptchaService{
					serverKey: "key",
					log:       zap.NewNop(),
					client: &http.Client{
						Transport: mockTransport,
					},
				}
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := tt.srvFunc(t)
			got := srv.Check(ctx, "token-string", "0.0.0.0")
			assert.Equal(t, tt.want, got)
		})
	}
}
