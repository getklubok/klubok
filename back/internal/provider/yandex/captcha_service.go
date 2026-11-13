package yandex

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.uber.org/zap"

	"klubok/internal/application/zapctx"
	"klubok/internal/domain"
)

const (
	defaultTimeout = 3 * time.Second

	smartCaptchaValidateURL = "https://smartcaptcha.cloud.yandex.ru/validate"
)

// CaptchaService реализует использование сервиса капчи.
type CaptchaService struct {
	serverKey string

	log *zap.Logger

	client *http.Client
}

// NewCaptchaService возвращает новый CaptchaService.
func NewCaptchaService(serverKey string, log *zap.Logger) *CaptchaService {
	return &CaptchaService{
		serverKey: serverKey,

		log: log.Named("captcha_service"),

		client: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

// Check проверяет капчу.
func (s *CaptchaService) Check(ctx context.Context, token string, userIP domain.IP) bool {
	data := url.Values{}
	data.Set("secret", s.serverKey)
	data.Set("token", token)
	data.Set("ip", userIP.String())

	reqBody := strings.NewReader(data.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, smartCaptchaValidateURL, reqBody)
	if err != nil {
		zapctx.Grasp(ctx, s.log).Error("Error creating request", zap.Error(err))

		return true
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.client.Do(req)
	if err != nil {
		zapctx.Grasp(ctx, s.log).Error("Error sending request", zap.Error(err))

		return true
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		zapctx.Grasp(ctx, s.log).Error("Error reading response body", zap.Error(err))

		return true
	}

	if resp.StatusCode != http.StatusOK {
		zapctx.Grasp(ctx, s.log).Warn("Allow access due to an error",
			zap.Int("code", resp.StatusCode),
			zap.String("body", string(body)),
		)

		return true
	}

	var captchaResponse CaptchaResponse
	if err := json.Unmarshal(body, &captchaResponse); err != nil {
		zapctx.Grasp(ctx, s.log).Error("Error unmarshaling JSON", zap.Error(err))

		return true
	}

	if captchaResponse.Status == "ok" {
		return true
	}

	zapctx.Grasp(ctx, s.log).Info("Robot detected",
		zap.String("status", captchaResponse.Status),
		zap.String("message", captchaResponse.Message),
		zap.String("host", captchaResponse.Host),
	)

	return false
}
