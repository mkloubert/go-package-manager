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

package commands

import (
	"strings"

	"github.com/mkloubert/go-package-manager/types"
	"github.com/mkloubert/go-package-manager/utils"
	"github.com/spf13/cobra"
)

func init_add_alias_command(parentCmd *cobra.Command, app *types.AppContext) {
	var addAliasCmd = &cobra.Command{
		Use:     "alias",
		Aliases: []string{"a"},
		Short:   "Add package alias",
		Long:    `Adds an alias for one or more packages.`,
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			alias := strings.TrimSpace(args[0])

			sources := app.AliasesFile.Aliases[alias]

			for _, s := range args[1:] {
				s = strings.TrimSpace(s)
				if s != "" {
					sources = append(sources, s)
				}
			}

			app.AliasesFile.Aliases[alias] = sources

			err := app.UpdateAliasesFile()
			if err != nil {
				utils.CloseWithError(err)
			}
		},
	}

	parentCmd.AddCommand(
		addAliasCmd,
	)
}

func Init_Add_Command(parentCmd *cobra.Command, app *types.AppContext) {
	var addCmd = &cobra.Command{
		Use:     "add [resource]",
		Aliases: []string{"+"},
		Short:   "Add command",
		Long:    `Adds a resource.`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	init_add_alias_command(addCmd, app)

	parentCmd.AddCommand(
		addCmd,
	)
}
