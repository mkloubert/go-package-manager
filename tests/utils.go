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
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/mkloubert/go-package-manager/app"
	"github.com/mkloubert/go-package-manager/types"
	"github.com/spf13/cobra"

	_ "embed"
)

//go:embed aliases.yaml
var aliasesYAML string

//go:embed gpm.yaml
var gpmYAML string

//go:embed projects.yaml
var projectsYAML string

// WithAppAction is an action for `WithApp` function
type WithAppAction = func(ctx *WithAppActionContext) error

// WithAppActionContext is a context for `WithAppAction`
type WithAppActionContext struct {
	App         *types.AppContext // the underlying application context
	Output      *bytes.Buffer     // is the default output for commands
	RootCommand *cobra.Command    // the root command
	T           *testing.T        // the testing context
}

// WithAppOptions stores additional options for `WithApp` function
type WithAppOptions struct {
	PostRun func(err error, ctx *WithAppActionContext) error // an optional callback that is executed AFTER the action has been executed
	PreRun  func(ctx *WithAppActionContext) error            // an optional callback that is executed BEFORE the action is executed
}

// ctx.Execute() - executes the root command, logs an error on fail and a `false`
func (ctx *WithAppActionContext) Execute() bool {
	err := ctx.RootCommand.Execute()
	if err == nil {
		return true
	}

	ctx.T.Error(err)
	return false
}

// ctx.ExecuteAndExpectHelp() - executes command and expects help function being
// executed
func (ctx *WithAppActionContext) ExecuteAndExpectHelp() bool {
	defaultHelp := ctx.RootCommand.HelpFunc()
	defer ctx.RootCommand.SetHelpFunc(defaultHelp)

	helpExecuted := false
	ctx.RootCommand.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		helpExecuted = true

		defaultHelp(cmd, args)
	})

	if ctx.Execute() {
		ctx.ExpectTrue(helpExecuted, "help was not executed")
	}
	return helpExecuted
}

// ctx.ExpectRegex() - checks string against a regular expression and logs error on fail
func (ctx *WithAppActionContext) ExpectRegex(s string, r *regexp.Regexp, errorMessage string) *WithAppActionContext {
	if errorMessage == "" {
		errorMessage = fmt.Sprintf(
			`"%v" does not match regular expression '%v'`,
			s,
			r.String(),
		)
	}

	return ctx.ExpectTrue(r.Match([]byte(s)), errorMessage)
}

// ctx.ExpectTrue() - checks if condition is true and logs error on fail
func (ctx *WithAppActionContext) ExpectTrue(condition bool, errorMessage string) *WithAppActionContext {
	if !condition {
		if errorMessage == "" {
			ctx.T.Error("condition does not match")
		} else {
			ctx.T.Error(errorMessage)
		}
	}

	return ctx
}

// ctx.ExpectValue() - checks if an actual value is the same as an expected one and logs error on fail
func (ctx *WithAppActionContext) ExpectValue(actual interface{}, expected interface{}, errorMessage string) *WithAppActionContext {
	if errorMessage == "" {
		errorMessage = fmt.Sprintf("actual(%v) != expected(%v)", actual, expected)
	}

	return ctx.ExpectTrue(actual == expected, errorMessage)
}

// ctx.OpenTempFile() - opens a new temp file
func (ctx *WithAppActionContext) OpenTempFile() (*os.File, error) {
	return os.CreateTemp("", "gpm-testing-file-*.bin")
}

// ctx.SetArgs() - sets the arguments for root command
func (ctx *WithAppActionContext) SetArgs(args ...string) *WithAppActionContext {
	ctx.RootCommand.SetArgs(args)

	return ctx
}

// ctx.WithStdin() - prepares the context so it is using a temporary file as STDIN
func (ctx *WithAppActionContext) WithStdin(action WithAppAction, p []byte) error {
	stdin, err := ctx.OpenTempFile()
	if err != nil {
		return err
	}

	defer func() {
		stdin.Close()
		os.Remove(stdin.Name())
	}()

	if len(p) > 0 {
		stdin.Write(p)

		_, err := stdin.Seek(0, 0)
		if err != nil {
			return err
		}
	}

	ctx.App.In = stdin

	return action(ctx)
}

// WithApp() - runs a test action for an app session
func WithApp(t *testing.T, action WithAppAction, options ...WithAppOptions) {
	a, rc, err := app.New()
	if err != nil {
		t.Error(err)
		return
	}

	// #region ********** current working directory **********
	cwd, err := os.MkdirTemp("", "gpm-testing-cwd-*")
	if err != nil {
		t.Error(err)
		return
	}

	defer func() {
		os.RemoveAll(cwd)
	}()

	a.Cwd = cwd
	// #endregion ********** current working directory **********

	// #region ********** virtual projects.yaml file **********
	projectsFile, err := os.CreateTemp("", "gpm-test-projects-file-*.yaml")
	if err != nil {
		t.Error(err)
		return
	}

	defer func() {
		projectsFile.Close()
		os.Remove(projectsFile.Name())
	}()

	_, err = projectsFile.WriteString(projectsYAML)
	if err != nil {
		t.Error(err)
		return
	}

	a.ProjectsFilePath = projectsFile.Name()
	// #endregion

	// #region ********** virtual aliases.yaml file **********
	aliasesFile, err := os.CreateTemp("", "gpm-test-aliases-file-*.yaml")
	if err != nil {
		t.Error(err)
		return
	}

	defer func() {
		aliasesFile.Close()
		os.Remove(aliasesFile.Name())
	}()

	_, err = aliasesFile.WriteString(aliasesYAML)
	if err != nil {
		t.Error(err)
		return
	}

	a.AliasesFilePath = aliasesFile.Name()
	// #endregion

	// #region ********** virtual gpm.yaml file **********
	var gpmFile types.GpmFile

	err = yaml.Unmarshal([]byte(gpmYAML), &gpmFile)
	if err != nil {
		t.Error(err)
		return
	}

	gpmFilePath := filepath.Join(cwd, "gpm.yaml")
	err = os.WriteFile(gpmFilePath, []byte(gpmYAML), 0664)
	if err != nil {
		t.Error(err)
		return
	}
	// #endregion

	output := &bytes.Buffer{}
	defer output.Reset()

	// now start initializing the virtual app ...

	a.Out = output
	a.Clipboard = &types.MemoryClipboard{}

	mockTimeStart, err := time.Parse(time.RFC3339, "1979-09-05T23:09:00.000Z")
	if err != nil {
		panic(err)
	}

	var lastNow *time.Time
	a.Now = func() time.Time {
		realNow := time.Now()
		defer func() {
			lastNow = &realNow
		}()

		if lastNow != nil {
			diff := realNow.Sub(*lastNow)

			return mockTimeStart.Add(diff)
		}

		return mockTimeStart
	}

	ctx := &WithAppActionContext{
		App:         a,
		Output:      output,
		RootCommand: rc,
		T:           t,
	}

	// pre run
	for _, o := range options {
		if o.PreRun != nil {
			err := o.PreRun(ctx)
			if err != nil {
				t.Error(err)
				return
			}
		}
	}

	a.LoadAliasesFileIfExist()
	a.LoadProjectsFileIfExist()
	a.LoadGpmFileIfExist()

	err = action(ctx)

	// post run
	for _, o := range options {
		if o.PostRun != nil {
			err2 := o.PostRun(err, ctx)
			if err2 != nil {
				t.Error(err2)
				return
			}
		}
	}

	if err != nil {
		t.Error(err)
		return
	}
}
