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
	"fmt"
	"strings"

	"github.com/mkloubert/go-package-manager/types"
	"github.com/mkloubert/go-package-manager/utils"
	"github.com/spf13/cobra"
)

func init_remove_alias_command(parentCmd *cobra.Command, app *types.AppContext) {
	var removeAliasCmd = &cobra.Command{
		Use:     "alias [alias name]",
		Aliases: []string{"a"},
		Short:   "Remove package alias",
		Long:    `Removes one or more aliases.`,
		Run: func(cmd *cobra.Command, args []string) {
			for _, a := range args {
				alias := strings.TrimSpace(a)

				app.Debug(fmt.Sprintf("Removing package alias '%v' ...", alias))
				delete(app.AliasesFile.Aliases, alias)
			}

			err := app.UpdateAliasesFile()
			if err != nil {
				utils.CloseWithError(err)
			}
		},
	}

	parentCmd.AddCommand(
		removeAliasCmd,
	)
}

func Init_Remove_Command(parentCmd *cobra.Command, app *types.AppContext) {
	var removeCmd = &cobra.Command{
		Use:     "remove [resource]",
		Aliases: []string{"rm"},
		Short:   "Remove command",
		Long:    `Removes a resource.`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	init_remove_alias_command(removeCmd, app)

	parentCmd.AddCommand(
		removeCmd,
	)
}