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

package tests

import (
	"bytes"
	"testing"

	tests "github.com/mkloubert/go-package-manager/tests"
)

func TestCatCommand(t *testing.T) {
	tests.WithApp(t, func(ctx1 *tests.WithAppActionContext) error {
		expectedValue := []byte("foo bar BUZZ")

		return ctx1.WithStdin(func(ctx *tests.WithAppActionContext) error {
			ctx.SetArgs("cat")

			if ctx.Execute() {
				actualValue := ctx.Output.Bytes()

				ctx.ExpectTrue(bytes.Equal(actualValue, expectedValue), "output is different to input")
			}

			return nil
		}, expectedValue)
	})
}
