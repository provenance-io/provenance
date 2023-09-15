package cmd

import (
	"fmt"

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
			markdown, err := cmd.Flags().GetBool(FlagMarkdown)
			if err != nil {
				return err
			}
			if markdown {
				doc.GenMarkdownTree(cmd.Root(), args[0])
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
