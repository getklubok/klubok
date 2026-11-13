package handler

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"klubok/internal/restapi/oas"
)

func TestFeatures_Features(t *testing.T) {
	tests := []struct {
		name             string
		captchaEnabled   bool
		captchaSiteKey   string
		authEmailEnabled bool
		authSsoEnabled   bool
		want             oas.FeaturesRes
	}{
		{
			name:             "Only email enabled",
			captchaEnabled:   false,
			captchaSiteKey:   "",
			authEmailEnabled: true,
			authSsoEnabled:   false,
			want: &oas.FeaturesResponse{
				Captcha:   oas.FeaturesCaptcha{Enabled: false, SiteKey: ""},
				AuthEmail: oas.FeaturesAuthEmail{Enabled: true},
				AuthSso:   oas.FeaturesAuthSso{Enabled: false},
			},
		},
		{
			name:             "Only SSO enabled",
			captchaEnabled:   true,
			captchaSiteKey:   "site-key",
			authEmailEnabled: false,
			authSsoEnabled:   true,
			want: &oas.FeaturesResponse{
				Captcha:   oas.FeaturesCaptcha{Enabled: true, SiteKey: "site-key"},
				AuthEmail: oas.FeaturesAuthEmail{Enabled: false},
				AuthSso:   oas.FeaturesAuthSso{Enabled: true},
			},
		},
		{
			name:             "Both login methods enabled",
			captchaEnabled:   true,
			captchaSiteKey:   "site-key",
			authEmailEnabled: true,
			authSsoEnabled:   true,
			want: &oas.FeaturesResponse{
				Captcha:   oas.FeaturesCaptcha{Enabled: true, SiteKey: "site-key"},
				AuthEmail: oas.FeaturesAuthEmail{Enabled: true},
				AuthSso:   oas.FeaturesAuthSso{Enabled: true},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewFeatures(tt.captchaEnabled, tt.captchaSiteKey, tt.authEmailEnabled, tt.authSsoEnabled)
			got, err := h.Features(context.Background())
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFeatures_Features_StableUntilRestart(t *testing.T) {
	h := NewFeatures(true, "site-key", true, false)

	t.Setenv("FEATURE_CAPTCHA_ENABLED", "false")
	t.Setenv("FEATURE_AUTH_EMAIL_ENABLED", "false")
	t.Setenv("FEATURE_AUTH_SSO_ENABLED", "true")

	got, err := h.Features(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, &oas.FeaturesResponse{
		Captcha:   oas.FeaturesCaptcha{Enabled: true, SiteKey: "site-key"},
		AuthEmail: oas.FeaturesAuthEmail{Enabled: true},
		AuthSso:   oas.FeaturesAuthSso{Enabled: false},
	}, got)
}
