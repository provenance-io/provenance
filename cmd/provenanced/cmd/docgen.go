package cmd

import (
	"fmt"
	"os"

	"github.com/cosmos/cosmos-sdk/version"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var docGenCmdStart = fmt.Sprintf("%s docgen", version.AppName)

const (
	FlagMarkdown = "markdown"
	FlagYaml     = "yaml"
	FlagRest     = "rest"
	FlagManpage  = "manpage"
)

func GetDocGenCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "docgen",
		Short:   "Generates cli documentation for the Provenance Blockchain.",
		Long:    "",
		Example: fmt.Sprintf("%s '/tmp' --yaml --markdown", docGenCmdStart),
		Hidden:  true,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := args[0]
			if !exists(dir) {
				err := os.Mkdir(dir, 0755)
				if err != nil {
					return err
				}
			}

			// Check if at least one flag is enabled

			markdown, err := cmd.Flags().GetBool(FlagMarkdown)
			if err != nil {
				return err
			}
			if markdown {
				doc.GenMarkdownTree(cmd.Root(), dir)
			}

			yaml, err := cmd.Flags().GetBool(FlagYaml)
			if err != nil {
				return err
			}
			if yaml {
				doc.GenYamlTree(cmd.Root(), dir)
			}

			rest, err := cmd.Flags().GetBool(FlagRest)
			if err != nil {
				return err
			}
			if rest {
				doc.GenReSTTree(cmd.Root(), dir)
			}

			manpage, err := cmd.Flags().GetBool(FlagManpage)
			if err != nil {
				return err
			}
			if manpage {
				doc.GenManTree(cmd.Root(), nil, dir)
			}

			return nil
		},
	}

	cmd.Flags().Bool(FlagMarkdown, false, "Generate documentation in the format of markdown pages.")
	cmd.Flags().Bool(FlagYaml, false, "Generate documentation in the format of yaml.")
	cmd.Flags().Bool(FlagRest, false, "Generate documentation in the format of rest.")
	cmd.Flags().Bool(FlagManpage, false, "Generate documentation in the format of manpages.")

	return cmd
}

func exists(dir string) bool {
	_, err := os.Stat(dir)
	return err == nil
}
