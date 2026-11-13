package domain

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func FuzzIsUsernameDenied(f *testing.F) {
	f.Add("admin")
	f.Add("Admin")
	f.Add("SUPPORT")
	f.Add("capital")
	f.Add("alice")
	f.Add("user123")
	f.Add("")

	f.Fuzz(func(t *testing.T, username string) {
		lower := strings.ToLower(username)
		upper := strings.ToUpper(username)

		if IsUsernameDenied(username) != IsUsernameDenied(lower) {
			t.Fatalf("IsUsernameDenied not case-insensitive: %q vs %q", username, lower)
		}
		if IsUsernameDenied(username) != IsUsernameDenied(upper) {
			t.Fatalf("IsUsernameDenied not case-insensitive (upper): %q vs %q", username, upper)
		}
		if IsUsernameDenied(username) && !IsUsernameDenied(username+"x") {
			t.Fatalf("IsUsernameDenied not prefix/contains-monotone: %q denied but %q allowed", username, username+"x")
		}
	})
}

func TestEmailDomainAllowed(t *testing.T) {
	tests := []struct {
		name            string
		email           string
		allowedSuffixes []string
		want            bool
	}{
		{
			name:            "Allowed exact domain",
			email:           "john@example.com",
			allowedSuffixes: []string{"example.com"},
			want:            true,
		},
		{
			name:            "Allowed subdomain",
			email:           "john@mail.example.com",
			allowedSuffixes: []string{"example.com"},
			want:            true,
		},
		{
			name:            "Similar but not allowed domain",
			email:           "john@notbadexample.com",
			allowedSuffixes: []string{"example.com"},
			want:            false,
		},
		{
			name:            "Case-insensitive email",
			email:           "JOHN@Example.COM",
			allowedSuffixes: []string{"example.com"},
			want:            true,
		},
		{
			name:            "Case-insensitive suffix",
			email:           "john@example.com",
			allowedSuffixes: []string{"Example.COM"},
			want:            true,
		},
		{
			name:            "Email without at-sign",
			email:           "john",
			allowedSuffixes: []string{"example.com"},
			want:            false,
		},
		{
			name:            "Empty allowed suffixes",
			email:           "john@example.com",
			allowedSuffixes: []string{},
			want:            false,
		},
		{
			name:            "Allowed with one of several suffixes",
			email:           "john@corp.example.com",
			allowedSuffixes: []string{"other.com", "example.com"},
			want:            true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EmailDomainAllowed(tt.email, tt.allowedSuffixes)
			assert.Equal(t, tt.want, got)
		})
	}
}

func FuzzEmailDomainAllowed(f *testing.F) {
	f.Add("john@example.com", "example.com")
	f.Add("john@mail.example.com", "example.com")
	f.Add("JOHN@Example.COM", "example.com")
	f.Add("john@example.com", "Example.COM")
	f.Add("john", "example.com")
	f.Add("john@example.com", "")

	f.Fuzz(func(t *testing.T, email, suffix string) {
		uppercase := strings.ToUpper(email)
		upperSuffix := strings.ToUpper(suffix)

		base := EmailDomainAllowed(email, []string{suffix})
		upper := EmailDomainAllowed(uppercase, []string{suffix})
		suffixUpper := EmailDomainAllowed(email, []string{upperSuffix})

		if base != upper {
			t.Fatalf("EmailDomainAllowed not case-insensitive on email: %q vs %q", email, uppercase)
		}
		if base != suffixUpper {
			t.Fatalf("EmailDomainAllowed not case-insensitive on suffix: %q vs %q", suffix, upperSuffix)
		}
	})
}
