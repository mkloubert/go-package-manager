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
	"encoding/base64"
	"fmt"
	"os"
	"testing"

	tests "github.com/mkloubert/go-package-manager/tests"
)

func TestBase64CommandWithSTDIN(t *testing.T) {
	tests.WithApp(t, func(ctx1 *tests.WithAppActionContext) error {
		input := []byte("FOO")

		return ctx1.WithStdin(func(ctx *tests.WithAppActionContext) error {
			ctx.SetArgs("base64")

			expectedOutput := base64.StdEncoding.EncodeToString(input)

			if ctx.Execute() {
				realOutput := ctx.Output.String()

				ctx.ExpectValue(realOutput, expectedOutput, "")
			}

			return nil
		}, input)
	})
}

func TestBase64CommandWithFile(t *testing.T) {
	tests.WithApp(t, func(ctx *tests.WithAppActionContext) error {
		input := []byte("FOO")

		file, err := os.CreateTemp("", "gpm-testing-base64-command-*.txt")
		if err != nil {
			return err
		}

		defer func() {
			file.Close()
			os.Remove(file.Name())
		}()

		_, err = file.Write(input)
		if err != nil {
			return err
		}

		_, err = file.Seek(0, 0)
		if err != nil {
			return err
		}

		ctx.SetArgs("base64", file.Name())

		expectedOutput := base64.StdEncoding.EncodeToString(input)

		if ctx.Execute() {
			realOutput := ctx.Output.String()

			ctx.ExpectValue(realOutput, expectedOutput, "")
		}

		return nil
	})
}

func TestBase64CommandWithDataURI(t *testing.T) {
	tests.WithApp(t, func(ctx1 *tests.WithAppActionContext) error {
		input := []byte("FOO")

		return ctx1.WithStdin(func(ctx *tests.WithAppActionContext) error {
			ctx.SetArgs("base64", "--data-uri")

			expectedOutput := fmt.Sprintf(
				"data:text/plain;base64,%s",
				base64.StdEncoding.EncodeToString(input),
			)

			if ctx.Execute() {
				realOutput := ctx.Output.String()

				ctx.ExpectValue(realOutput, expectedOutput, "")
			}

			return nil
		}, input)
	})
}

func TestBase64CommandWithFileAndDataURI(t *testing.T) {
	tests.WithApp(t, func(ctx *tests.WithAppActionContext) error {
		input := []byte("FOO")

		file, err := os.CreateTemp("", "gpm-testing-base64-command-*.txt")
		if err != nil {
			return err
		}

		defer func() {
			file.Close()
			os.Remove(file.Name())
		}()

		_, err = file.Write(input)
		if err != nil {
			return err
		}

		_, err = file.Seek(0, 0)
		if err != nil {
			return err
		}

		ctx.SetArgs("base64", "--data-uri", file.Name())

		expectedOutput := fmt.Sprintf(
			"data:text/plain;base64,%s",
			base64.StdEncoding.EncodeToString(input),
		)

		if ctx.Execute() {
			realOutput := ctx.Output.String()

			ctx.ExpectValue(realOutput, expectedOutput, "")
		}

		return nil
	})
}
