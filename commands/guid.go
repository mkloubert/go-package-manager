package commands

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/atotto/clipboard"
	"github.com/google/uuid"
	"github.com/mkloubert/go-package-manager/types"
	"github.com/mkloubert/go-package-manager/utils"
	"github.com/spf13/cobra"
)

func Init_GUID_Command(parentCmd *cobra.Command, app *types.AppContext) {
	var base64Output bool
	var count uint16
	var copyToClipboard bool
	var noOutput bool
	var waitTime int

	var guidCmd = &cobra.Command{
		Use:     "guid",
		Aliases: []string{"guids", "uuid", "uuids"},
		Short:   "Generate UUID",
		Long:    `Generates one or more UUIDs/GUIDs.`,
		Run: func(cmd *cobra.Command, args []string) {
			clipboardContent := ""

			var addClipboardContent func(text string)
			if copyToClipboard {
				addClipboardContent = func(text string) {
					clipboardContent += text
				}
			} else {
				addClipboardContent = func(text string) {
					// dummy
				}
			}

			var i uint16 = 0
			for {
				if i == count {
					break
				}

				i++
				if i > 1 {
					fmt.Println()
					addClipboardContent(fmt.Sprintln())

					time.Sleep(time.Duration(waitTime) * time.Millisecond)
				}

				app.Debug(fmt.Sprintf("Generating passwords %v ...", i))

				guid, err := uuid.NewRandom()
				utils.CheckForError(err)

				var passwordToOutput string
				if base64Output {
					app.Debug("Base64 output ...")

					passwordToOutput = base64.URLEncoding.EncodeToString(guid[:])
				} else {
					passwordToOutput = guid.String()
				}

				if !noOutput {
					fmt.Print(passwordToOutput)
				}

				addClipboardContent(passwordToOutput)
			}

			if copyToClipboard {
				app.Debug("Copy all to clipboard ...")

				err := clipboard.WriteAll(clipboardContent)
				utils.CheckForError(err)
			}
		},
	}

	guidCmd.Flags().BoolVarP(&base64Output, "base64", "", false, "output as Base64 string")
	guidCmd.Flags().BoolVarP(&copyToClipboard, "copy", "", false, "copy final content to clipboard")
	guidCmd.Flags().Uint16VarP(&count, "count", "", 1, "custom number of guids to generate at once")
	guidCmd.Flags().BoolVarP(&noOutput, "no-output", "", false, "do not output to console")
	guidCmd.Flags().IntVarP(&waitTime, "wait-time", "", 0, "the time in millieconds to wait between two steps")

	parentCmd.AddCommand(
		guidCmd,
	)
}
