package commands

import (
	"strings"

	"github.com/mkloubert/go-package-manager/types"
	"github.com/mkloubert/go-package-manager/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func init_generate_documentation_command(parentCmd *cobra.Command, app *types.AppContext) {
	var man bool
	var markdown bool
	var rest bool
	var yaml bool

	var documentationCmd = &cobra.Command{
		Use:     "documentation [resource]",
		Aliases: []string{"doc", "docs", "dox"},
		Short:   "Generate documentation",
		Long:    `Generate documentation into the current directory.`,
		PreRun: func(cmd *cobra.Command, args []string) {
			if !man && !markdown && !rest && !yaml {
				app.Debug("Setting 'markdown' as default format ...")

				// default is Markdown
				markdown = true
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			outDir := app.Cwd
			if len(args) > 0 {
				outDir = strings.TrimSpace(args[0])
			}

			outDir, err := app.EnsureFolder(outDir)
			utils.CheckForError(err)

			// collect generators by flags
			generators := make([]func(), 0)
			if man {
				// man pages
				generators = append(generators, func() {
					app.Debug("Generating man pages in", outDir, "...")

					header := doc.GenManHeader{}

					err := doc.GenManTree(cmd.Root(), &header, outDir)
					utils.CheckForError(err)
				})
			}
			if markdown {
				// Markdown files
				generators = append(generators, func() {
					app.Debug("Generating Markdown files in", outDir, "...")

					err := doc.GenMarkdownTree(cmd.Root(), outDir)
					utils.CheckForError(err)
				})
			}
			if rest {
				// ReST files
				generators = append(generators, func() {
					app.Debug("Generating ReST files in", outDir, "...")

					err := doc.GenReSTTree(cmd.Root(), outDir)
					utils.CheckForError(err)
				})
			}
			if yaml {
				// YAML files
				generators = append(generators, func() {
					app.Debug("Generating YAML files in", outDir, "...")

					err := doc.GenYamlTree(cmd.Root(), outDir)
					utils.CheckForError(err)
				})
			}

			// execute generators
			for _, generate := range generators {
				generate()
			}
		},
	}

	documentationCmd.Flags().BoolVarP(&man, "man", "", false, "generate man pages")
	documentationCmd.Flags().BoolVarP(&markdown, "markdown", "m", false, "generate Markdown files")
	documentationCmd.Flags().BoolVarP(&rest, "rest", "r", false, "generate ReST files")
	documentationCmd.Flags().BoolVarP(&yaml, "yaml", "y", false, "generate YAML files")

	parentCmd.AddCommand(
		documentationCmd,
	)
}
