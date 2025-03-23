package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/mkloubert/go-package-manager/types"
	"github.com/mkloubert/go-package-manager/utils"
	"github.com/spf13/cobra"
)

type generatePowerpointMarkdownResponse struct {
	MarkdownCodeForPandoc string `json:"markdown_code_for_pandoc,omitempty"`
}

func init_generate_powerpoint_command(parentCmd *cobra.Command, app *types.AppContext) {
	var additionalContext string
	var customCwd string
	var customLanguage string
	var customTemplate string
	var focusOn string
	var minSlides int
	var maxSlides int

	var powerpointCmd = &cobra.Command{
		Use:     "powerpoint [output file] [resources]",
		Args:    cobra.MinimumNArgs(1),
		Aliases: []string{"pptx", "slideshow"},
		Short:   "Generate PowerPoint",
		Long:    `Generate PowerPoint presentation from sources like text files.`,
		Run: func(cmd *cobra.Command, args []string) {
			now := app.Now()

			outFile := app.GetFullPathOrDefault(args[0], "presentation.pptx")
			if !strings.HasSuffix(outFile, ".pptx") {
				outFile = outFile + ".pptx"
			}

			moreContext := strings.TrimSpace(additionalContext)

			cmdCwd := app.GetFullPathOrDefault(strings.TrimSpace(customCwd), app.Cwd)

			sourcesAndPatterns := make([]string, 0)
			if len(args) > 1 {
				sourcesAndPatterns = append(sourcesAndPatterns, args[1:]...)
			}

			app.Debug(fmt.Sprintf("Sources and patterns: %s", strings.Join(sourcesAndPatterns, ", ")))

			chat, err := app.CreateAIChat()
			utils.CheckForError(err)

			chat.UpdateTemperature(app.GetAITemperature(0.3))

			systemPrompt := strings.TrimSpace(app.SystemPrompt)
			if systemPrompt == "" {
				systemPrompt = `You are an assistant tasked with helping me create PowerPoint presentations from provided files.
I will share the content of files with you step by step.
During this process, you have to respond with 'OK' until I give you further instructions.`
			}

			language := strings.TrimSpace(customLanguage)
			if language == "" {
				language = "english"
			}

			app.Debug(fmt.Sprintf("Output language: %s", language))

			sources, err := app.FindSourceFiles(sourcesAndPatterns...)
			utils.CheckForError(err)

			if len(sources) == 0 {
				utils.CheckForError(fmt.Errorf("no sources found"))
			}

			app.Debug(fmt.Sprintf("Found %v sources", len(sources)))

			textData := map[string]string{}

			app.Debug("Checking sources if all is readable text and collect them ...")
			for _, s := range sources {
				app.Debug(fmt.Sprintf("Checking source '%v' ...", s))

				data, err := app.LoadDataFrom(s)
				utils.CheckForError(err)

				isReadable := utils.IsReadableText(data)
				if !isReadable {
					utils.CheckForError(fmt.Errorf("%s is binary data and cannot be handled", s))
				}

				text := strings.TrimSpace(string(data))
				if text != "" {
					textData[s] = text

					app.Debug(fmt.Sprintf("Added source '%s'", s))
				} else {
					app.Debug(fmt.Sprintf("Warning: '%s' has no data", s))
				}
			}

			if len(textData) == 0 {
				utils.CheckForError(fmt.Errorf("no data found that can be handled"))
			}

			fileNr := 0
			for src, text := range textData {
				fileNr = fileNr + 1

				app.Debug(fmt.Sprintf("Adding source (#%v) ('%s') with %v characters to chat history ...", fileNr, src, len(text)))

				chat.AddToHistory(
					"user",
					fmt.Sprintf("File number %v with path '%s':\n%s", fileNr, src, text),
				)
				chat.AddToHistory("assistant", "OK")
			}

			jsonSchema := map[string]interface{}{
				"type":     "object",
				"required": []string{"markdown_code"},
				"properties": map[string]interface{}{
					"markdown_code_for_pandoc": map[string]interface{}{
						"description": "The markdown code which can be used with Pandoc tool to create PowerPoint files.",
						"type":        "string",
					},
				},
			}

			jsonStr := ""

			slideCountInfo := ""
			if minSlides > -1 && maxSlides > -1 {
				slideCountInfo = fmt.Sprintf("Produce between %v and %v slides.", minSlides, maxSlides)
			} else if minSlides > -1 {
				slideCountInfo = fmt.Sprintf("Produce a minimum of %v slides.", minSlides)
			} else if maxSlides > -1 {
				slideCountInfo = fmt.Sprintf("Produce a maximum of %v slides.", maxSlides)
			}

			focusInfo := strings.TrimSpace(focusOn)
			if focusInfo != "" {
				focusInfo = fmt.Sprintf("Focus in particular on the following: %v", focusInfo)
			}

			app.Debug("Starting AI chat ...")
			chat.WithJsonSchema(
				fmt.Sprintf(`Now with all this information you will write Markdown code that can be handled by pandoc to create a .pptx file from it.
I need this PowerPoint to summerize all this information.

Here is an example:
<EXAMPLE-START>
  ---
  title: "Presentation Title"
  author: "Author Name"
  date: "1979-09-05"
  output: powerpoint_presentation
  ---

  # Short and descriptive title of slide 1
  - Bullet point 1
  - Bullet point 2

  # Short and descriptive title of slide 2
  Text content.
</EXAMPLE-END>

%s

%s

%s

Your final Pandoc compatible markdown in %s language (today is %s):`,
					moreContext,
					focusInfo,
					slideCountInfo,
					language,
					now.Format("January 02, 2006"),
				),
				"PandocMarkdownSchema",
				jsonSchema,
				func(chunk string) error {
					jsonStr = jsonStr + chunk

					return nil
				},
			)

			var response generatePowerpointMarkdownResponse
			err = json.Unmarshal([]byte(jsonStr), &response)
			utils.CheckForError(err)

			inFile, err := os.CreateTemp("", "gpm-md-to-pptx-*.md")
			utils.CheckForError(err)
			defer func() {
				app.Debug(fmt.Sprintf("Deleting file '%s' ...", inFile.Name()))

				os.Remove(inFile.Name())
			}()

			app.Debug(fmt.Sprintf("Output markdown to '%s' ...", inFile.Name()))
			bytesWritten, err := inFile.WriteString(response.MarkdownCodeForPandoc)
			utils.CheckForError(err)
			app.Debug(fmt.Sprintf("%v bytes written", bytesWritten))

			cmdTplCode := strings.TrimSpace(customTemplate)
			if cmdTplCode == "" {
				// now try from environment variable

				cmdTplCode = strings.TrimSpace(
					app.SettingsFile.GetString("generate.pptx.from.md.command", "", ""),
				)
			}
			if cmdTplCode == "" {
				// use default
				cmdTplCode = `pandoc -t pptx -o "{{.OutputFile}}" "{{.InputFile}}"`
			}

			app.Debug(
				fmt.Sprintf(
					"Using command template value '%s' to generate .pptx from Markdown ...",
					cmdTplCode,
				),
			)

			cmdTpl, err := template.New("command").Parse(cmdTplCode)
			utils.CheckForError(err)

			cmdArgs := map[string]string{
				"InputFile":  inFile.Name(),
				"OutputFile": outFile,
			}

			var finalCommand bytes.Buffer
			defer finalCommand.Reset()

			err = cmdTpl.Execute(&finalCommand, cmdArgs)
			utils.CheckForError(err)

			app.Debug(fmt.Sprintf("Executing '%s' ...", finalCommand.String()))
			p := utils.CreateShellCommand(finalCommand.String())
			p.Dir = cmdCwd
			p.Stdout = app.Out
			p.Stderr = app.ErrorOut
			p.Stdin = app.In

			err = p.Run()
			utils.CheckForError(err)
		},
	}

	powerpointCmd.Flags().StringVarP(&additionalContext, "context", "", "", "additional information for the AI")
	powerpointCmd.Flags().StringVarP(&customCwd, "cwd", "", "", "custom working directory for command that generates PowerPoint")
	powerpointCmd.Flags().StringVarP(&customLanguage, "language", "", "", "custom response language")
	powerpointCmd.Flags().StringVarP(&customTemplate, "template", "", "", "custom template for command that generates PowerPoint")
	powerpointCmd.Flags().StringVarP(&focusOn, "focus-on", "", "", "additional information about the focus")
	powerpointCmd.Flags().IntVarP(&maxSlides, "max-slides", "", -1, "tell AI number of maximum slides")
	powerpointCmd.Flags().IntVarP(&minSlides, "min-slides", "", -1, "tell AI number of minimum slides")

	parentCmd.AddCommand(
		powerpointCmd,
	)
}
