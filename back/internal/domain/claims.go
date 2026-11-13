package domain

import (
	"context"
	"errors"
)

type claimsCtxKey struct{}

type Claims struct {
	SessionID SessionID
	UserID    UserID
}

func (c Claims) Context(ctx context.Context) context.Context {
	return context.WithValue(ctx, claimsCtxKey{}, c)
}

func ClaimsFromContext(ctx context.Context) (Claims, error) {
	claims, ok := ctx.Value(claimsCtxKey{}).(Claims)
	if !ok {
		return Claims{}, errors.New("claims not found in context")
	}
	if claims.SessionID == "" {
		return Claims{}, errors.New("claims session id is empty")
	}
	if claims.UserID == "" {
		return Claims{}, errors.New("claims user id is empty")
	}

	return claims, nil
}
