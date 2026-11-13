package sso

import (
	"context"
	"errors"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// Claims — проверенные данные пользователя из id_token.
type Claims struct {
	Sub               string
	Email             string
	PreferredUsername string
}

// Config параметры OIDC-провайдера.
type Config struct {
	IssuerURL    string
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// OIDCProvider — вход через SSO по протоколу OIDC (generic, discovery по issuer URL).
type OIDCProvider struct {
	issuer   string
	oauthCfg oauth2.Config
	verifier *oidc.IDTokenVerifier
}

// NewOIDCProvider выполняет OIDC discovery по issuer URL и возвращает провайдер.
// Ошибка discovery означает недоступность SSO-входа.
func NewOIDCProvider(ctx context.Context, cfg Config) (*OIDCProvider, error) {
	provider, err := oidc.NewProvider(ctx, cfg.IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("oidc discovery: %w", err)
	}

	return &OIDCProvider{
		issuer: cfg.IssuerURL,
		oauthCfg: oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  cfg.RedirectURL,
			Endpoint:     provider.Endpoint(),
			Scopes:       []string{oidc.ScopeOpenID, oidc.ScopeEmail, oidc.ScopeProfile},
		},
		verifier: provider.Verifier(&oidc.Config{ClientID: cfg.ClientID}),
	}, nil
}

// AuthCodeURL формирует authorization URL провайдера с state/nonce.
func (p *OIDCProvider) AuthCodeURL(state string, nonce string) string {
	return p.oauthCfg.AuthCodeURL(state, oidc.Nonce(nonce))
}

// Authenticate обменивает code на токены, проверяет id_token
// (issuer, audience, expiry — библиотекой, nonce — вручную) и извлекает claims.
// Email в claims обязателен: без него пользователя не создать и не найти.
func (p *OIDCProvider) Authenticate(ctx context.Context, code string, nonce string) (Claims, error) {
	token, err := p.oauthCfg.Exchange(ctx, code)
	if err != nil {
		return Claims{}, fmt.Errorf("exchange code: %w", err)
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok || rawIDToken == "" {
		return Claims{}, errors.New("id_token missing in token response")
	}

	idToken, err := p.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return Claims{}, fmt.Errorf("verify id_token: %w", err)
	}

	if idToken.Nonce != nonce {
		return Claims{}, errors.New("invalid nonce")
	}

	var claims struct {
		Sub   string `json:"sub"`
		Email string `json:"email"`
		//nolint:tagliatelle // Имя claim зафиксировано спецификацией OIDC
		PreferredUsername string `json:"preferred_username"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return Claims{}, fmt.Errorf("parse claims: %w", err)
	}
	if claims.Sub == "" || claims.Email == "" {
		return Claims{}, errors.New("sub or email missing in claims")
	}

	return Claims{
		Sub:               claims.Sub,
		Email:             claims.Email,
		PreferredUsername: claims.PreferredUsername,
	}, nil
}
