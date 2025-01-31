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
	"compress/gzip"
	"io"
	"testing"

	tests "github.com/mkloubert/go-package-manager/tests"
)

func TestUncompressCommand(t *testing.T) {
	tests.WithApp(t, func(ctx1 *tests.WithAppActionContext) error {
		valueToCompress := []byte("foo bar BUZZ")

		var temp1 bytes.Buffer
		defer temp1.Reset()

		writer := gzip.NewWriter(&temp1)

		_, err := writer.Write(valueToCompress)
		if err != nil {
			return err
		}

		err = writer.Close()
		if err != nil {
			return err
		}

		valueToDecompress := temp1.Bytes()

		reader, err := gzip.NewReader(&temp1)
		if err != nil {
			return err
		}

		var temp2 bytes.Buffer
		defer temp2.Reset()

		_, err = io.Copy(&temp2, reader)
		if err != nil {
			return err
		}

		expectedValue := temp2.Bytes()

		return ctx1.WithStdin(func(ctx *tests.WithAppActionContext) error {
			ctx.SetArgs("uncompress")

			if ctx.Execute() {
				actualValue := ctx.Output.Bytes()

				ctx.ExpectTrue(bytes.Equal(actualValue, expectedValue), "output is different to input")
			}

			return nil
		}, valueToDecompress)
	})
}
