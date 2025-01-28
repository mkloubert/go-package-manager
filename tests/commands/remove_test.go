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
	"fmt"
	"os"
	"testing"

	"github.com/goccy/go-yaml"
	tests "github.com/mkloubert/go-package-manager/tests"
)

func TestRemoveCommand(t *testing.T) {
	tests.WithApp(t, func(ctx *tests.WithAppActionContext) error {
		ctx.SetArgs("remove").
			ExecuteAndExpectHelp()

		return nil
	})
}

func TestRemoveAliasCommand(t *testing.T) {
	tests.WithApp(t, func(ctx *tests.WithAppActionContext) error {
		aliasName := "yaml"

		ctx.SetArgs("remove", "alias", aliasName)

		checkContent := func() error {
			yamlDataFromObject, err := yaml.Marshal(ctx.App.AliasesFile)
			if err != nil {
				return nil
			}

			yamlDataFromFile, err := os.ReadFile(ctx.App.AliasesFilePath)
			if err != nil {
				return nil
			}

			if !bytes.Equal(yamlDataFromFile, yamlDataFromObject) {
				return fmt.Errorf("'%v' file and current AliasesFile do not represent same content", ctx.App.AliasesFilePath)
			}
			return nil
		}

		if ctx.Execute() {
			err := checkContent()
			if err != nil {
				return err
			}

			v, ok := ctx.App.AliasesFile.Aliases[aliasName]

			ctx.ExpectTrue(!ok, fmt.Sprintf("%v does still exist", aliasName))
			ctx.ExpectTrue(len(v) == 0, fmt.Sprintf("number of items in %v is %v", aliasName, len(v)))
		}

		return nil
	})
}

func TestRemoveProjectsCommand(t *testing.T) {
	tests.WithApp(t, func(ctx *tests.WithAppActionContext) error {
		projectName := "yaml"

		ctx.SetArgs("remove", "project", projectName)

		checkContent := func() error {
			yamlDataFromObject, err := yaml.Marshal(ctx.App.ProjectsFile)
			if err != nil {
				return nil
			}

			yamlDataFromFile, err := os.ReadFile(ctx.App.ProjectsFilePath)
			if err != nil {
				return nil
			}

			if !bytes.Equal(yamlDataFromFile, yamlDataFromObject) {
				return fmt.Errorf("'%v' file and current ProjectsFile do not represent same content", ctx.App.ProjectsFilePath)
			}
			return nil
		}

		if ctx.Execute() {
			err := checkContent()
			if err != nil {
				return err
			}

			_, ok := ctx.App.ProjectsFile.Projects[projectName]

			ctx.ExpectTrue(!ok, fmt.Sprintf("%v does still exist", projectName))
		}

		return nil
	})
}
