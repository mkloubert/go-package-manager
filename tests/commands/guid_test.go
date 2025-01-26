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
	"fmt"
	"regexp"
	"testing"

	tests "github.com/mkloubert/go-package-manager/tests"
)

func TestGuidCommand(t *testing.T) {
	tests.WithApp(t, func(ctx *tests.WithAppActionContext) error {
		ctx.SetArgs("guid")

		if ctx.Execute() {
			output := ctx.Output.String()

			ctx.ExpectValue(len(output), 36, "").
				ExpectRegex(output, regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`), "")
		}

		return nil
	})
}

func TestGuidCommandWithClipboard(t *testing.T) {
	tests.WithApp(t, func(ctx *tests.WithAppActionContext) error {
		ctx.SetArgs("guid", "--copy")

		if ctx.Execute() {
			clipboard, err := ctx.App.Clipboard.ReadText()
			if err != nil {
				return err
			}

			ctx.ExpectValue(len(clipboard), 36, "").
				ExpectRegex(clipboard, regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`), "")
		}

		return nil
	})
}

func TestGuidCommandWithClipboardMultipleValues(t *testing.T) {
	tests.WithApp(t, func(ctx *tests.WithAppActionContext) error {
		for count := 1; count < 10; count++ {
			ctx.App.Clipboard.WriteText("")

			expectedLength := 36*count + (count-1)*len(fmt.Sprintln())

			ctx.SetArgs("guid", "--copy", fmt.Sprintf("--count=%v", count))

			if ctx.Execute() {
				clipboard, err := ctx.App.Clipboard.ReadText()
				if err != nil {
					return err
				}

				ctx.ExpectValue(len(clipboard), expectedLength, "")
			}
		}

		return nil
	})
}

func TestGuidCommandWithMultipleValues(t *testing.T) {
	tests.WithApp(t, func(ctx *tests.WithAppActionContext) error {
		for count := 1; count < 10; count++ {
			ctx.Output.Reset()

			expectedLength := 36*count + (count-1)*len(fmt.Sprintln())

			ctx.SetArgs("guid", fmt.Sprintf("--count=%v", count))

			if ctx.Execute() {
				output := ctx.Output.String()

				ctx.ExpectValue(len(output), expectedLength, "")
			}
		}

		return nil
	})
}

func TestGuidCommandWithoutOutput(t *testing.T) {
	tests.WithApp(t, func(ctx *tests.WithAppActionContext) error {
		for count := 1; count < 10; count++ {
			ctx.Output.Reset()

			ctx.SetArgs("guid", "--no-output", fmt.Sprintf("--count=%v", count))

			if ctx.Execute() {
				output := ctx.Output.String()

				ctx.ExpectValue(len(output), 0, "")
			}
		}

		return nil
	})
}
