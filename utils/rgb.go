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

package utils

import (
	"crypto/md5"
	"encoding/hex"
	"image/color"
	"strconv"
)

func calculateBrightness(r, g, b uint8) float64 {
	return (0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)) / 255
}

// GenerateColorsFromString() - generates unique background
// and foregrund colors from a string
func GenerateColorsFromString(str string) (color.RGBA, color.RGBA) {
	hash := md5.Sum([]byte(str))
	hexColor := hex.EncodeToString(hash[:])

	r := hexToInt(hexColor[0:2])
	g := hexToInt(hexColor[2:4])
	b := hexToInt(hexColor[4:6])

	background := color.RGBA{R: r, G: g, B: b, A: 255}
	var foreground color.RGBA

	brightness := calculateBrightness(r, g, b)
	if brightness > 0.5 {
		// light background => black
		foreground = color.RGBA{R: 0, G: 0, B: 0, A: 255}
	} else {
		// dark background => white
		foreground = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	}

	return background, foreground
}

func hexToInt(hexStr string) uint8 {
	val, _ := strconv.ParseUint(hexStr, 16, 8)

	return uint8(val)
}
