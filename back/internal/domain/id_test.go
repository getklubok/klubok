package domain

import (
	"strings"
	"testing"

	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/assert"
)

const (
	ulidLen        = 26
	ulidSuffixSize = 8
)

func FuzzIsID(f *testing.F) {
	f.Add("01J6GP155TZP3FMQ6F86KHE34Y")
	f.Add("01j6gp155tzp3fmq6f86khe34y")
	f.Add("0000000000000000000000000 ")
	f.Add("0000000000000000000000000!")
	f.Add("invalid")
	f.Add("")

	f.Fuzz(func(t *testing.T, id string) {
		if !IsID(id) {
			return
		}

		want := ulid.MustParse(id).String()
		if got := strings.ToUpper(id); got != want {
			t.Fatalf("IsID(%q) = true, but string is not canonical ULID: %q vs %q", id, got, want)
		}
	})
}

func TestNewID(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "new id",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewID()
			assert.True(t, IsID(got))
		})
	}
}

func TestIsID(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want bool
	}{
		{
			name: "valid id",
			id:   "01J6GP155TZP3FMQ6F86KHE34Y",
			want: true,
		},
		{
			name: "invalid id",
			id:   "invalid",
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsID(tt.id)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPartID(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want string
	}{
		{
			name: "valid id",
			id:   "01J6GP155TZP3FMQ6F86KHE34Y",
			want: "86KHE34Y",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PartID(tt.id)
			assert.Equal(t, tt.want, got)
		})
	}
}

func FuzzPartID(f *testing.F) {
	f.Add("01J6GP155TZP3FMQ6F86KHE34Y")
	f.Add("01j6gp155tzp3fmq6f86khe34y")
	f.Add("01J6GP155TZP3FMQ6F86KHE34z")
	f.Add("invalid")
	f.Add("")

	f.Fuzz(func(t *testing.T, id string) {
		if !IsID(id) {
			return
		}
		if len(id) != ulidLen {
			t.Fatalf("IsID(%q) = true, but length is %d, want %d", id, len(id), ulidLen)
		}

		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("PartID(%q) panicked: %v", id, r)
			}
		}()

		want := strings.ToUpper(id)[ulidLen-ulidSuffixSize:]
		if got := PartID(id); got != want {
			t.Fatalf("PartID(%q) = %q, want %q", id, got, want)
		}
	})
}
