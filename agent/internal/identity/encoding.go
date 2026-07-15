package identity

import (
	"encoding/base64"
	"fmt"
)

const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// base62Encode encodes a byte slice as a base62 string.
func base62Encode(data []byte) string {
	if len(data) == 0 {
		return ""
	}

	// Convert to a big-endian integer and encode in base62.
	var result []byte
	var num []byte
	num = append(num, data...)

	// Process all bytes as a big number.
	for !isZero(num) {
		var remainder byte
		num, remainder = divmod62(num)
		result = append(result, base62Chars[remainder])
	}

	// Reverse the result (LSB-first encoding).
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return string(result)
}

// isZero returns true if the big-endian byte slice represents zero.
func isZero(data []byte) bool {
	for _, b := range data {
		if b != 0 {
			return false
		}
	}
	return true
}

// divmod62 divides a big-endian byte slice by 62, returning the quotient
// and remainder.
func divmod62(data []byte) ([]byte, byte) {
	var remainder uint16
	result := make([]byte, len(data))

	for i, b := range data {
		remainder = remainder<<8 | uint16(b)
		result[i] = byte(remainder / 62)
		remainder %= 62
	}

	// Trim leading zeros from quotient.
	start := 0
	for start < len(result) && result[start] == 0 {
		start++
	}
	if start == len(result) {
		return []byte{0}, byte(remainder)
	}

	return result[start:], byte(remainder)
}

// base64Encode encodes a byte slice as a standard base64 string.
func base64Encode(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// base64DecodeKey decodes a base64-encoded key and checks its length.
func base64DecodeKey(encoded string, expectedLen int) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("base64 decode: %w", err)
	}
	if len(data) != expectedLen {
		return nil, fmt.Errorf("unexpected key length: got %d, want %d", len(data), expectedLen)
	}
	return data, nil
}
