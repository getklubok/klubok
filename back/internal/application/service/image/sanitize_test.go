package image

import (
	"strings"
	"testing"
	"unicode"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeOriginalName(t *testing.T) {
	tests := []struct {
		name string
		in   string
		ext  string
		want string
	}{
		{
			name: "Plain name",
			in:   "scheme.png",
			ext:  extPNG,
			want: "scheme.png",
		},
		{
			name: "Unix path",
			in:   "../dir/scheme.png",
			ext:  extPNG,
			want: "scheme.png",
		},
		{
			name: "Windows path",
			in:   `C:\photos\scheme.png`,
			ext:  extPNG,
			want: "scheme.png",
		},
		{
			name: "Mixed separators",
			in:   `..\dir/../scheme.png`,
			ext:  extPNG,
			want: "scheme.png",
		},
		{
			name: "Newlines and control characters",
			in:   "scheme\n\r\x00\x01.png",
			ext:  extPNG,
			want: "scheme.png",
		},
		{
			name: "Empty name",
			in:   "",
			ext:  extJPEG,
			want: "image.jpg",
		},
		{
			name: "Only separators",
			in:   `/\`,
			ext:  extJPEG,
			want: "image.jpg",
		},
		{
			name: "Only dots",
			in:   "..",
			ext:  extPNG,
			want: "image.png",
		},
		{
			name: "Only spaces",
			in:   "   ",
			ext:  extPNG,
			want: "image.png",
		},
		{
			name: "Unicode name",
			in:   "схема 1.png",
			ext:  extPNG,
			want: "схема 1.png",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, sanitizeOriginalName(tt.in, tt.ext))
		})
	}
}

func FuzzSanitizeOriginalName(f *testing.F) {
	f.Add("../evil.png")
	f.Add(`C:\dir\evil.png`)
	f.Add("evil\n.png")
	f.Add("evil\x00.png")
	f.Add("")
	f.Add("..")
	f.Add("схема.png")

	f.Fuzz(func(t *testing.T, name string) {
		got := sanitizeOriginalName(name, extPNG)

		if got == "" {
			t.Fatal("sanitized name is empty")
		}

		if strings.ContainsAny(got, `/\`) {
			t.Fatalf("sanitized name %q contains path separator", got)
		}

		if strings.ContainsAny(got, "\n\r") {
			t.Fatalf("sanitized name %q contains line break", got)
		}

		for _, r := range got {
			if unicode.IsControl(r) {
				t.Fatalf("sanitized name %q contains control character %U", got, r)
			}
		}
	})
}
