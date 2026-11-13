package domain

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClaimsFromContext(t *testing.T) {
	tests := []struct {
		name    string
		ctxFunc func() context.Context
		want    Claims
		wantErr error
	}{
		{
			name: "should return Claims",
			ctxFunc: func() context.Context {
				return Claims{SessionID: "s1", UserID: "u1"}.Context(context.Background())
			},
			want: Claims{
				SessionID: "s1",
				UserID:    "u1",
			},
			wantErr: nil,
		},
		{
			name:    "should return error: claims not found in context",
			ctxFunc: context.Background,
			want:    Claims{},
			wantErr: errors.New("claims not found in context"),
		},
		{
			name: "should return error: claims session id is empty",
			ctxFunc: func() context.Context {
				return Claims{SessionID: "", UserID: "u1"}.Context(context.Background())
			},
			want:    Claims{},
			wantErr: errors.New("claims session id is empty"),
		},
		{
			name: "should return error: claims user id is empty",
			ctxFunc: func() context.Context {
				return Claims{SessionID: "s1", UserID: ""}.Context(context.Background())
			},
			want:    Claims{},
			wantErr: errors.New("claims user id is empty"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := ClaimsFromContext(tt.ctxFunc())
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantErr, gotErr)
		})
	}
}
