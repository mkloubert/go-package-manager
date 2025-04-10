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
	"bytes"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/mkloubert/go-package-manager/types"
	"github.com/mkloubert/go-package-manager/utils"
	"github.com/spf13/cobra"
)

func Init_Base64_Command(parentCmd *cobra.Command, app *types.AppContext) {
	var dataURI bool

	var base64Cmd = &cobra.Command{
		Use:     "base64 [files]",
		Aliases: []string{"b64"},
		Short:   "Encode Base64",
		Long:    `Encode data from STDIN to STDOUT as Base64 encoded data.`,
		Run: func(cmd *cobra.Command, args []string) {
			debugWrittenValue := func(written int64) {
				if app.Verbose {
					fmt.Println()
				}
				app.Debug(fmt.Sprintf("Bytes written: %v", written))
			}

			if dataURI {
				var buffer bytes.Buffer

				written, err := app.WriteAllInputsTo(&buffer, args...)
				utils.CheckForError(err)

				data := buffer.Bytes()
				defer buffer.Reset()

				encoder := base64.NewEncoder(base64.StdEncoding, app.Out)
				defer encoder.Close()

				mimeType := http.DetectContentType(data)
				if strings.Contains(mimeType, ";") {
					// remove things right to mime type like charset
					mimeType = strings.SplitN(mimeType, ";", 2)[0]
				}

				base64Data := base64.StdEncoding.EncodeToString(data)

				app.Write([]byte(
					fmt.Sprintf(
						"data:%s;base64,%s",
						strings.TrimSpace(
							strings.ToLower(mimeType),
						),
						base64Data,
					),
				))

				debugWrittenValue(written)
			} else {
				encoder := base64.NewEncoder(base64.StdEncoding, app.Out)
				defer encoder.Close()

				written, err := app.WriteAllInputsTo(encoder, args...)
				utils.CheckForError(err)

				debugWrittenValue(written)
			}
		},
	}

	base64Cmd.Flags().BoolVarP(&dataURI, "data-uri", "", false, "output as data URI")

	parentCmd.AddCommand(
		base64Cmd,
	)
}
