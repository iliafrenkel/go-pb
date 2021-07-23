// Package base62 provides simple implementation of Base62 encoding/ecoding.
// Borrowed from here: https://intersog.com/blog/how-to-write-a-custom-url-shortener-using-golang-and-redis/
// unfortunately the source didn't provide any licencing.
package base62

import (
	"errors"
	"math"
	"strings"
)

const (
	alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	length   = int64(len(alphabet))
)

// Encode returns a base62 encoded number.
func Encode(number int64) string {
	var encodedBuilder strings.Builder
	encodedBuilder.Grow(11)

	for ; number > 0; number = number / length {
		encodedBuilder.WriteByte(alphabet[(number % length)])
	}

	return encodedBuilder.String()
}

// Decode tries to decode a base62 encoded number.
func Decode(encoded string) (int64, error) {
	var number int64

	for i, symbol := range encoded {
		alphabeticPosition := strings.IndexRune(alphabet, symbol)

		if alphabeticPosition == -1 {
			return int64(alphabeticPosition), errors.New("invalid character: " + string(symbol))
		}
		number += int64(alphabeticPosition) * int64(math.Pow(float64(length), float64(i)))
	}

	return number, nil
}
