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
	"compress/gzip"
	"fmt"
	"io"

	"github.com/mkloubert/go-package-manager/types"
	"github.com/mkloubert/go-package-manager/utils"
	"github.com/spf13/cobra"
)

func Init_Uncompress_Command(parentCmd *cobra.Command, app *types.AppContext) {
	var uncompressCmd = &cobra.Command{
		Use:     "uncompress",
		Aliases: []string{"decompress"},
		Short:   "Uncompress data",
		Long:    `Uncompresses gzip compressed input data.`,
		Run: func(cmd *cobra.Command, args []string) {
			buffer := bytes.Buffer{}
			defer buffer.Reset()

			_, err := app.WriteAllInputsTo(&buffer, args...)
			utils.CheckForError(err)

			reader, err := gzip.NewReader(&buffer)
			utils.CheckForError(err)
			defer reader.Close()

			written, err := io.Copy(app.Out, reader)
			utils.CheckForError(err)

			if app.Verbose {
				fmt.Println()
			}
			app.Debug(fmt.Sprintf("Bytes written: %v", written))
		},
	}

	parentCmd.AddCommand(
		uncompressCmd,
	)
}
