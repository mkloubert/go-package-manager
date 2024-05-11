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

	"github.com/mkloubert/go-package-manager/types"
	"github.com/spf13/cobra"
)

func init_list_aliases_command(parentCmd *cobra.Command, app *types.AppContext) {
	var listAliasesCmd = &cobra.Command{
		Use:     "aliases",
		Aliases: []string{"a", "alias"},
		Short:   "List package aliases",
		Long:    `Lists (all) aliases.`,
		Run: func(cmd *cobra.Command, args []string) {
			for alias, sources := range app.AliasesFile.Aliases {
				fmt.Printf("%v%v", alias, fmt.Sprintln())

				for _, s := range sources {
					fmt.Printf("\t%v%v", s, fmt.Sprintln())
				}
			}
		},
	}

	parentCmd.AddCommand(
		listAliasesCmd,
	)
}

func Init_List_Command(parentCmd *cobra.Command, app *types.AppContext) {
	var listCmd = &cobra.Command{
		Use:     "list [resource]",
		Aliases: []string{"l", "lst"},
		Short:   "List command",
		Long:    `Lists resource(s).`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	init_list_aliases_command(listCmd, app)

	parentCmd.AddCommand(
		listCmd,
	)
}
