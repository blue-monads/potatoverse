package xutils

import (
	"fmt"
	"strings"
)

const charset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func ReadableInt64(rawcode int64) string {
	if rawcode == 0 {
		return "0"
	}

	base := int64(len(charset))
	result := ""
	isNegative := rawcode < 0

	// Handle negative numbers by converting to positive
	if isNegative {
		rawcode = -rawcode
	}

	for rawcode > 0 {
		remainder := rawcode % base
		result = string(charset[remainder]) + result
		rawcode = rawcode / base
	}

	if isNegative {
		result = "-" + result
	}

	return result
}

func DecodeReadable(encoded string) (int64, error) {
	if encoded == "" {
		return 0, fmt.Errorf("encoded string is empty")
	}

	isNegative := encoded[0] == '-'
	if isNegative {
		encoded = encoded[1:]
	}

	base := int64(len(charset))
	var result int64 = 0

	for _, char := range encoded {
		value := int64(strings.IndexRune(charset, char))
		if value == -1 {
			return 0, fmt.Errorf("invalid character in encoded string: %c", char)
		}
		result = result*base + value
	}

	if isNegative {
		result = -result
	}

	return result, nil
}

func Main() {
	encoded := ReadableInt64(95672368)
	fmt.Println("Encoded:", encoded)

	decoded, err := DecodeReadable(encoded)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Decoded:", decoded)
	}

	// Test with negative number
	encodedNegative := ReadableInt64(-95672368)
	fmt.Println("Encoded negative:", encodedNegative)

	decodedNegative, err := DecodeReadable(encodedNegative)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Decoded negative:", decodedNegative)
	}

	// Test with zero
	encodedZero := ReadableInt64(0)
	fmt.Println("Encoded zero:", encodedZero)

	decodedZero, err := DecodeReadable(encodedZero)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Decoded zero:", decodedZero)
	}
}
