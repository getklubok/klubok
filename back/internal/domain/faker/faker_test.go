package faker

import (
	"math"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

func TestID(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "id"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ID()

			assert.Len(t, got, 20)

			for _, r := range got {
				assert.True(
					t,
					strings.ContainsRune(idAlphabet, r),
					"symbol %q is outside the alphabet",
					r,
				)
			}
		})
	}
}

func TestNumber(t *testing.T) {
	tests := []struct {
		name  string
		min   int
		max   int
		fixed bool
		want  int
	}{
		{
			name:  "min equals max",
			min:   5,
			max:   5,
			fixed: true,
			want:  5,
		},
		{
			name: "min greater than max",
			min:  10,
			max:  3,
		},
		{
			name: "negative min greater than max",
			min:  -1,
			max:  -100,
		},
		{
			name: "negative range",
			min:  -100,
			max:  -50,
		},
		{
			name: "mixed range",
			min:  -50,
			max:  50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Number(tt.min, tt.max)

			lo, hi := tt.min, tt.max
			if lo > hi {
				lo, hi = hi, lo
			}

			assert.GreaterOrEqual(t, got, lo)
			assert.LessOrEqual(t, got, hi)

			if tt.fixed {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestDigitN(t *testing.T) {
	tests := []struct {
		name    string
		n       uint
		wantLen int
	}{
		{
			name:    "zero means one",
			n:       0,
			wantLen: 1,
		},
		{
			name:    "several digits",
			n:       4,
			wantLen: 4,
		},
		{
			name:    "long digits",
			n:       100,
			wantLen: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DigitN(tt.n)

			assert.Len(t, got, tt.wantLen)

			for _, r := range got {
				assert.True(
					t,
					strings.ContainsRune(digitAlphabet, r),
					"symbol %q is not a digit",
					r,
				)
			}
		})
	}
}

func TestWords(t *testing.T) {
	tests := []struct {
		name string
		word string
		list []string
	}{
		{
			name: "fruit",
			word: Fruit(),
			list: fruits,
		},
		{
			name: "vegetable",
			word: Vegetable(),
			list: vegetables,
		},
		{
			name: "animal",
			word: Animal(),
			list: animals,
		},
		{
			name: "bird",
			word: Bird(),
			list: birds,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotEmpty(t, tt.word)
			assert.Contains(t, tt.list, tt.word)
		})
	}
}

func TestWordListsNormalized(t *testing.T) {
	tests := []struct {
		name string
		list []string
	}{
		{name: "fruits", list: fruits},
		{name: "vegetables", list: vegetables},
		{name: "animals", list: animals},
		{name: "birds", list: birds},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seen := make(map[string]struct{}, len(tt.list))

			for _, w := range tt.list {
				assert.NotEmpty(t, w)
				assert.True(t, utf8.ValidString(w))

				for _, r := range w {
					ok := r == ' ' || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
					assert.True(t, ok, "word %q has a non-letter %q", w, r)
				}

				_, dup := seen[w]
				assert.False(t, dup, "duplicate word %q", w)
				seen[w] = struct{}{}
			}
		})
	}
}

func FuzzNumber(f *testing.F) {
	f.Add(0, 100)
	f.Add(-50, 50)
	f.Add(10, 3)
	f.Add(-1, -1)
	f.Add(math.MinInt, math.MaxInt)
	f.Add(math.MinInt+1, math.MaxInt-1)

	f.Fuzz(func(t *testing.T, minVal int, maxVal int) {
		got := Number(minVal, maxVal)

		lo, hi := minVal, maxVal
		if lo > hi {
			lo, hi = hi, lo
		}

		assert.GreaterOrEqual(t, got, lo)
		assert.LessOrEqual(t, got, hi)
	})
}

func FuzzDigitN(f *testing.F) {
	f.Add(uint(0))
	f.Add(uint(1))
	f.Add(uint(4))
	f.Add(uint(100))

	f.Fuzz(func(t *testing.T, n uint) {
		got := DigitN(n)

		wantLen := n
		if wantLen == 0 {
			wantLen = 1
		}

		assert.Len(t, got, int(wantLen))

		for _, r := range got {
			assert.True(t,
				strings.ContainsRune(digitAlphabet, r),
				"symbol %q is not a digit",
				r,
			)
		}
	})
}

func FuzzID(f *testing.F) {
	f.Add(0)

	f.Fuzz(func(t *testing.T, _ int) {
		got := ID()

		assert.Len(t, got, 20)

		for _, r := range got {
			assert.True(t,
				strings.ContainsRune(idAlphabet, r),
				"symbol %q is outside the alphabet",
				r,
			)
		}
	})
}
