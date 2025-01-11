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
	"time"

	"github.com/mkloubert/go-package-manager/types"
	"github.com/spf13/cobra"
)

func Init_Now_Command(parentCmd *cobra.Command, app *types.AppContext) {
	var format string
	var local bool

	var nowCmd = &cobra.Command{
		Use:   "now",
		Short: "Output time",
		Long:  `Outputs current time.`,
		Run: func(cmd *cobra.Command, args []string) {
			now := time.Now()
			if !local {
				now = now.UTC()
			}

			outputFormat := format
			if outputFormat == "" {
				if local {
					outputFormat = "2006-01-02T15:04:05.000"
				} else {
					outputFormat = "2006-01-02T15:04:05.000Z"
				}
			}

			fmt.Print(now.Format(outputFormat))
		},
	}

	nowCmd.Flags().StringVarP(&format, "format", "", "", "custom output format")
	nowCmd.Flags().BoolVarP(&local, "local", "", false, "use local time")

	parentCmd.AddCommand(
		nowCmd,
	)
}
