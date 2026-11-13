package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvironmentFromString(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want Environment
	}{
		{
			name: "should return Environment - LOCAL",
			s:    "LOCAL",
			want: EnvironmentLocal,
		},
		{
			name: "should return Environment - STAGE",
			s:    "STAGE",
			want: EnvironmentStage,
		},
		{
			name: "should return Environment - PROD",
			s:    "PROD",
			want: EnvironmentProd,
		},
		{
			name: "should return Environment - PROD",
			s:    "UNKNOWN",
			want: EnvironmentProd,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EnvironmentFromString(tt.s)
			assert.Equal(t, tt.want, got)
		})
	}
}

func FuzzEnvironmentFromString(f *testing.F) {
	f.Add("local")
	f.Add("LOCAL")
	f.Add("STAGE")
	f.Add("prod")
	f.Add("unknown")
	f.Add("")

	f.Fuzz(func(t *testing.T, s string) {
		got := EnvironmentFromString(s)

		switch got {
		case EnvironmentLocal, EnvironmentStage, EnvironmentProd:
		default:
			t.Fatalf("EnvironmentFromString(%q) returned unknown environment %q", s, got)
		}

		if again := EnvironmentFromString(got.String()); again != got {
			t.Fatalf("EnvironmentFromString is not idempotent: %q -> %q -> %q", s, got, again)
		}
	})
}
