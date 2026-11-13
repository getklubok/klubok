package faker

import (
	"crypto/rand"
	"encoding/binary"
)

const (
	idLength   = 20
	idAlphabet = "23456789abcdefgghjkmnpqrstuvwxyz"

	digitAlphabet = "0123456789"
)

// ID возвращает случайный идентификатор из 20 символов читаемого алфавита
// (без 0/o/1/i/l, с удвоенной g — как в gofakeit).
func ID() string {
	b := make([]byte, idLength)

	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}

	for i := range b {
		b[i] = idAlphabet[b[i]%byte(len(idAlphabet))]
	}

	return string(b)
}

// Number возвращает случайное число из отрезка [minVal, maxVal] включительно.
// При minVal > maxVal аргументы меняются местами, при minVal == maxVal возвращается minVal.
func Number(minVal int, maxVal int) int {
	if minVal == maxVal {
		return minVal
	}

	if minVal > maxVal {
		minVal, maxVal = maxVal, minVal
	}

	// Разность считается в uint64 с намеренным заворачиванием по модулю 2^64 —
	// иначе отрезок через границу int (например, [MinInt, MaxInt]) не выразить.
	size := uint64(maxVal) - uint64(minVal) //nolint:gosec // intentional two's-complement wraparound
	if size == ^uint64(0) {
		return int(randUint64()) //nolint:gosec // wraps to cover the full int range uniformly
	}

	return minVal + int(randUintN(size+1)) //nolint:gosec // wraps for negative spans; result stays within [minVal, maxVal]
}

// DigitN возвращает строку из n случайных цифр; n == 0 трактуется как 1.
func DigitN(n uint) string {
	if n == 0 {
		n = 1
	}

	b := make([]byte, n)

	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}

	for i := range b {
		b[i] = digitAlphabet[b[i]%byte(len(digitAlphabet))]
	}

	return string(b)
}

// Fruit возвращает случайное название фрукта.
func Fruit() string {
	return randomWord(fruits)
}

// Vegetable возвращает случайное название овоща.
func Vegetable() string {
	return randomWord(vegetables)
}

// Animal возвращает случайное название животного.
func Animal() string {
	return randomWord(animals)
}

// Bird возвращает случайное название птицы.
func Bird() string {
	return randomWord(birds)
}

// randomWord возвращает случайное слово из списка.
func randomWord(words []string) string {
	return words[randUintN(uint64(len(words)))]
}

// randUintN возвращает случайное число из [0, n) без смещения
// (rejection sampling вместо модульного отображения).
func randUintN(n uint64) uint64 {
	limit := ^uint64(0) - ^uint64(0)%n

	for {
		v := randUint64()
		if v < limit {
			return v % n
		}
	}
}

// randUint64 возвращает случайное число из [0, 2^64).
func randUint64() uint64 {
	var b [8]byte

	_, err := rand.Read(b[:])
	if err != nil {
		panic(err)
	}

	return binary.LittleEndian.Uint64(b[:])
}
