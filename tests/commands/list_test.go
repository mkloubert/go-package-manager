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
	"os"
	"testing"

	tests "github.com/mkloubert/go-package-manager/tests"
)

func TestListCommand(t *testing.T) {
	tests.WithApp(t, func(ctx *tests.WithAppActionContext) error {
		ctx.SetArgs("list").
			ExecuteAndExpectHelp()

		return nil
	})
}

func TestListAliasesCommand(t *testing.T) {
	tests.WithApp(t, func(ctx *tests.WithAppActionContext) error {
		ctx.SetArgs("list", "aliases")

		expectedOutput := `yaml
	https://github.com/go-yaml/yaml
`

		if ctx.Execute() {
			ctx.ExpectValue(ctx.Output.String(), expectedOutput, "")
		}

		return nil
	})
}

func TestListBinariesCommand(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gpm-testing-list-binaries-*")
	if err != nil {
		t.Error(err)
		return
	}

	tests.WithApp(t, func(ctx *tests.WithAppActionContext) error {
		ctx.SetArgs("list", "binaries")

		expectedOutput := fmt.Sprintf(`%v`, tempDir)

		if ctx.Execute() {
			ctx.ExpectValue(ctx.Output.String(), expectedOutput, "")
		}

		return nil
	}, tests.WithAppOptions{
		PreRun: func(ctx *tests.WithAppActionContext) error {
			ctx.T.Setenv("GPM_BIN_PATH", tempDir)

			return nil
		},
		PostRun: func(err error, ctx *tests.WithAppActionContext) error {
			os.RemoveAll(tempDir)

			return nil
		},
	})
}

func TestListProjectsCommand(t *testing.T) {
	tests.WithApp(t, func(ctx *tests.WithAppActionContext) error {
		ctx.SetArgs("list", "projects")

		expectedOutput := `gpm
	https://github.com/mkloubert/go-package-manager
`

		if ctx.Execute() {
			ctx.ExpectValue(ctx.Output.String(), expectedOutput, "")
		}

		return nil
	})
}
