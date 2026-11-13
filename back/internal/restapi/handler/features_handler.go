package handler

import (
	"context"

	"klubok/internal/restapi/oas"
)

type Features struct {
	captchaEnabled   bool
	captchaSiteKey   string
	authEmailEnabled bool
	authSsoEnabled   bool
}

func NewFeatures(captchaEnabled bool, captchaSiteKey string, authEmailEnabled bool, authSsoEnabled bool) *Features {
	return &Features{
		captchaEnabled:   captchaEnabled,
		captchaSiteKey:   captchaSiteKey,
		authEmailEnabled: authEmailEnabled,
		authSsoEnabled:   authSsoEnabled,
	}
}

// Get feature flags.
//
// GET /features.
func (h *Features) Features(_ context.Context) (oas.FeaturesRes, error) { //nolint:unparam // ogen handlers must match this signature
	return &oas.FeaturesResponse{
		Captcha: oas.FeaturesCaptcha{
			Enabled: h.captchaEnabled,
			SiteKey: h.captchaSiteKey,
		},
		AuthEmail: oas.FeaturesAuthEmail{
			Enabled: h.authEmailEnabled,
		},
		AuthSso: oas.FeaturesAuthSso{
			Enabled: h.authSsoEnabled,
		},
	}, nil
}
