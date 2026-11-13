package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewString(t *testing.T) {
	tests := []struct {
		name    string
		length  int
		wantLen int
	}{
		{
			name:    "new string",
			length:  10,
			wantLen: 10,
		},
		{
			name:    "zero length",
			length:  0,
			wantLen: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewString(tt.length)
			assert.Len(t, got, tt.wantLen)
		})
	}
}

func TestNewDigitCode(t *testing.T) {
	tests := []struct {
		name    string
		length  int
		wantLen int
	}{
		{
			name:    "new digit code",
			length:  10,
			wantLen: 10,
		},
		{
			name:    "zero length",
			length:  0,
			wantLen: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewDigitCode(tt.length)
			assert.Len(t, got, tt.wantLen)
		})
	}
}
