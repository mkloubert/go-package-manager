package utils

import "unicode"

// IsReadableText() - checks if data is text and mostly readable
func IsReadableText(data []byte) bool {
	textCount := 0
	binaryCount := 0

	for _, b := range data {
		if b >= 32 && b <= 126 { // printable ASCII
			textCount++
		} else if b == 9 || b == 10 || b == 13 { // tab, newline, carriage return
			textCount++
		} else if unicode.IsPrint(rune(b)) { // Unicode printable characters
			textCount++
		} else { // Non-readable (binary) character
			binaryCount++
		}

		// if binary characters exceed a threshold, treat as binary
		if binaryCount > len(data)/10 {
			return false
		}
	}

	// if most of the characters are readable, it's text
	return textCount > binaryCount
}
