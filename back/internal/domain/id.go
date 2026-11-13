package domain

import (
	"crypto/rand"
	"strings"

	"github.com/oklog/ulid/v2"
)

// NewID returns a new identifier.
func NewID() string {
	entropy := ulid.Monotonic(rand.Reader, 0)

	return ulid.MustNew(ulid.Now(), entropy).String()
}

func IsID(id string) bool {
	u, err := ulid.Parse(id)
	if err != nil {
		return false
	}

	return strings.ToUpper(id) == u.String()
}

func PartID(id string) string {
	return ulid.MustParse(id).String()[18:]
}
