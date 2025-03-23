package commands

import (
	cryptoRand "crypto/rand"
	"encoding/base64"
	"fmt"
	"math/big"
	"time"

	"github.com/mkloubert/go-package-manager/types"
	"github.com/mkloubert/go-package-manager/utils"
	"github.com/spf13/cobra"
)

func Init_Password_Command(parentCmd *cobra.Command, app *types.AppContext) {
	var allBytes bool
	var base64Output bool
	var count uint16
	var charset string
	var copyToClipboard bool
	var length uint16
	var minLength uint16
	var noOutput bool
	var waitTime int

	var generatePasswordCmd = &cobra.Command{
		Use:     "password",
		Aliases: []string{"passwd", "passwds", "passwords", "pwd", "pwds"},
		Short:   "Generate password",
		Long:    `Generates one or more passwords.`,
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

			if minLength > 0 {
				if minLength > length {
					utils.CheckForError(fmt.Errorf("min-length %v cannot be greater then length %v", minLength, length))
				}
			}

			var i uint16 = 0
			for {
				if i == count {
					break
				}

				i++
				if i > 1 {
					app.WriteString(fmt.Sprintln())
					addClipboardContent(fmt.Sprintln())

					time.Sleep(time.Duration(waitTime) * time.Millisecond)
				}

				app.Debug(fmt.Sprintf("Generating passwords %v ...", i))

				var passwordLength uint16
				if minLength > 0 {
					randVal := utils.GenerateRandomUint16()

					passwordLength = utils.MaxUint16(randVal%length, minLength)
				} else {
					passwordLength = length
				}

				app.Debug(fmt.Sprintf("Password length %v ...", passwordLength))

				password := make([]byte, int(passwordLength))

				if allBytes {
					// use any byte
					app.Debug("Will use no charset ...")

					_, err := cryptoRand.Read(password)
					utils.CheckForError(err)
				} else {
					passwordCharset := charset
					if passwordCharset == "" {
						passwordCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()-_=+[]{}<>?/|"
					}

					app.Debug(fmt.Sprintf("Will use charset: %s", charset))

					for j := range password {
						index, err := cryptoRand.Int(cryptoRand.Reader, big.NewInt(int64(len(passwordCharset))))
						utils.CheckForError(err)

						password[j] = passwordCharset[index.Int64()]
					}
				}

				var passwordToOutput string
				if base64Output {
					app.Debug("Base64 output ...")

					passwordToOutput = base64.URLEncoding.EncodeToString(password)
				} else {
					passwordToOutput = string(password)
				}

				if !noOutput {
					app.WriteString(passwordToOutput)
				}

				addClipboardContent(passwordToOutput)
			}

			if copyToClipboard {
				app.Debug("Copy all to clipboard ...")

				err := app.Clipboard.WriteText(clipboardContent)
				utils.CheckForError(err)
			}
		},
	}

	generatePasswordCmd.Flags().BoolVarP(&allBytes, "all-bytes", "", false, "use any byte for password")
	generatePasswordCmd.Flags().BoolVarP(&base64Output, "base64", "", false, "output as Base64 string")
	generatePasswordCmd.Flags().StringVarP(&charset, "charset", "", "", "custom charset")
	generatePasswordCmd.Flags().BoolVarP(&copyToClipboard, "copy", "", false, "copy final content to clipboard")
	generatePasswordCmd.Flags().Uint16VarP(&count, "count", "", 1, "custom number password to generate at once")
	generatePasswordCmd.Flags().Uint16VarP(&length, "length", "", 20, "custom length of password")
	generatePasswordCmd.Flags().Uint16VarP(&minLength, "min-length", "", 0, "if defined the length of password will be flexible")
	generatePasswordCmd.Flags().BoolVarP(&noOutput, "no-output", "", false, "do not output to console")
	generatePasswordCmd.Flags().IntVarP(&waitTime, "wait-time", "", 0, "the time in millieconds to wait between two steps")

	parentCmd.AddCommand(
		generatePasswordCmd,
	)
}
