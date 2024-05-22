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
	"os"
	"os/exec"
	"strings"

	"github.com/alecthomas/chroma/quick"
	"github.com/hashicorp/go-version"
	"github.com/spf13/cobra"

	"github.com/mkloubert/go-package-manager/types"
	"github.com/mkloubert/go-package-manager/utils"
)

func Init_Diff_Command(parentCmd *cobra.Command, app *types.AppContext) {
	var diffCmd = &cobra.Command{
		Use:     "diff [resource]",
		Aliases: []string{"df"},
		Short:   "Diff resources",
		Long:    `Compares two resources.`,
		Run: func(cmd *cobra.Command, args []string) {
			consoleFormatter := utils.GetBestChromaFormatterName()
			consoleStyle := utils.GetBestChromaStyleName()

			version1, err := version.NewVersion(strings.TrimSpace(args[0]))
			utils.CheckForError(err)

			tag1 := "v" + version1.String()
			var tag2 string

			if len(args) == 1 {
				tag2 = "HEAD"
			} else {
				version2, err := version.NewVersion(strings.TrimSpace(args[1]))
				utils.CheckForError(err)

				tag2 = "v" + version2.String()
			}

			p := exec.Command("git", "diff", tag1, tag2)
			p.Dir = app.Cwd

			diff, err := p.Output()
			utils.CheckForError(err)

			err = quick.Highlight(os.Stdout, string(diff), "diff", consoleFormatter, consoleStyle)
			if err != nil {
				fmt.Print(string(diff))
			}
		},
	}

	parentCmd.AddCommand(
		diffCmd,
	)
}
