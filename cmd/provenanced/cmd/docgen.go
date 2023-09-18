package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"

	"github.com/cosmos/cosmos-sdk/version"
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
		Use:   "docgen <target directory> (--markdown) (--yaml) (--rest) (--manpages) [flags]",
		Short: "Generates cli documentation for the Provenance Blockchain.",
		Long: `Generates cli documentation for the Provenance Blockchain.
Various documentation formats can be generated, including markdown, YAML, REST, and man pages. 
To ensure the command's success, you must specify at least one format.
A successful command will not only generate files in the selected formats but also create the target directory if it doesn't already exist.`,
		Example: fmt.Sprintf("%s '/tmp' --yaml --markdown", docGenCmdStart),
		Hidden:  true,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			markdown, err := cmd.Flags().GetBool(FlagMarkdown)
			if err != nil {
				return err
			}
			yaml, err := cmd.Flags().GetBool(FlagYaml)
			if err != nil {
				return err
			}
			rest, err := cmd.Flags().GetBool(FlagRest)
			if err != nil {
				return err
			}
			manpage, err := cmd.Flags().GetBool(FlagManpage)
			if err != nil {
				return err
			}

			if !markdown && !yaml && !rest && !manpage {
				return fmt.Errorf("at least one doc type must be specified")
			}

			dir := args[0]
			if !exists(dir) {
				err = os.Mkdir(dir, 0755)
				if err != nil {
					return err
				}
			}

			if markdown {
				err = doc.GenMarkdownTree(cmd.Root(), dir)
				if err != nil {
					return err
				}
			}
			if yaml {
				err = doc.GenYamlTree(cmd.Root(), dir)
				if err != nil {
					return err
				}
			}
			if rest {
				err = doc.GenReSTTree(cmd.Root(), dir)
				if err != nil {
					return err
				}
			}
			if manpage {
				err = doc.GenManTree(cmd.Root(), nil, dir)
				if err != nil {
					return err
				}
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
