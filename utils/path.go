// MIT License
//
// Copyright (c) 2024 Marcel Joachim Kloubert (https://marcel.coffee)
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// original code from: https://github.com/egomobile/sanitize_filename

package utils

import (
	"regexp"
)

// SanitizeFilenameOptions stores options for `SanitizeFilename()` function
type SanitizeFilenameOptions struct {
	Replacement *string // character to replace unsafe characters with
}

// SanitizeFilename() - cleans up an input string to one which can be used in a filename
func SanitizeFilename(input string, options ...SanitizeFilenameOptions) string {
	var replacement string = ""
	for _, o := range options {
		if o.Replacement != nil {
			replacement = *o.Replacement
		}
	}

	// replace characters that are illegal for filenames
	illegalRe := regexp.MustCompile(`[\\/\?<>\:*|"]`)
	input = illegalRe.ReplaceAllString(input, replacement)

	// replace Unicode control characters
	controlRe := regexp.MustCompile(`[\x00-\x1f\x80-\x9f]`)
	input = controlRe.ReplaceAllString(input, replacement)

	// replace reserved filenames like '.' and '..'
	reservedRe := regexp.MustCompile(`^\.+$`)
	input = reservedRe.ReplaceAllString(input, replacement)

	// replace Windows reserved filenames
	windowsReservedRe := regexp.MustCompile(`(?i)^(con|prn|aux|nul|com[0-9]|lpt[0-9])(\..*)?$`)
	input = windowsReservedRe.ReplaceAllString(input, replacement)

	// remove trailing dots and spaces
	windowsTrailingRe := regexp.MustCompile(`[\. ]+$`)
	input = windowsTrailingRe.ReplaceAllString(input, replacement)

	// Ensure the result is no longer than 255 characters
	if len(input) > 255 {
		input = input[:255]
	}
	return input
}
