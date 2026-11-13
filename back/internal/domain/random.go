package domain

import (
	"crypto/rand"
)

const (
	charsetString    = "acdefhkmnprstuvwxyzACDEFHKMNPRSTUVWXYZ2345679"
	charsetDigitCode = "0123456789"
)

// NewString возвращает новую случайную строку указанной длины.
func NewString(length int) string {
	b := make([]byte, length)

	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}

	for i := range b {
		b[i] = charsetString[b[i]%byte(len(charsetString))]
	}

	return string(b)
}

// NewDigitCode возвращает новый случайный цифровой код указанной длины.
func NewDigitCode(length int) string {
	b := make([]byte, length)

	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}

	for i := range b {
		b[i] = charsetDigitCode[b[i]%byte(len(charsetDigitCode))]
	}

	return string(b)
}
