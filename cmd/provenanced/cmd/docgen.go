package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"

	"github.com/cosmos/cosmos-sdk/version"
)

var docGenCmdStart = fmt.Sprintf("%s docgen", version.AppName)

const (
	FlagMarkdown = "markdown"
	FlagYaml     = "yaml"
	FlagRst      = "rst"
	FlagManpage  = "manpage"
)

func GetDocGenCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "docgen <target directory> (--markdown) (--yaml) (--rst) (--manpages) [flags]",
		Short: "Generates cli documentation for the Provenance Blockchain.",
		Long: `Generates cli documentation for the Provenance Blockchain.
Various documentation formats can be generated, including markdown, YAML, RST, and man pages. 
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
			rst, err := cmd.Flags().GetBool(FlagRst)
			if err != nil {
				return err
			}
			manpage, err := cmd.Flags().GetBool(FlagManpage)
			if err != nil {
				return err
			}

			if !markdown && !yaml && !rst && !manpage {
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
			if rst {
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
	cmd.Flags().Bool(FlagRst, false, "Generate documentation in the format of rst.")
	cmd.Flags().Bool(FlagManpage, false, "Generate documentation in the format of manpages.")

	return cmd
}

func exists(dir string) bool {
	_, err := os.Stat(dir)
	return err == nil
}

// GetTreeCmd gets the tree command that will output all the commands available.
func GetTreeCmd() *cobra.Command {
	aliasesFlag := "aliases"
	cmd := &cobra.Command{
		Use:    "tree [sub-command] [--" + aliasesFlag + "]",
		Short:  "Get a tree of the commands optionally limited to a specific sub-command",
		Hidden: true,
		Args:   cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 && args[0] != "provenanced" {
				args = append([]string{"provenanced"}, args...)
			}

			incAliases, _ := cmd.Flags().GetBool(aliasesFlag)
			cmds := getAllCommands(cmd.Root(), "", args, incAliases)
			cmd.Printf(strings.Join(cmds, "\n") + "\n")
			return nil
		},
		DisableFlagsInUseLine: true,
	}
	cmd.Flags().Bool(aliasesFlag, false, "Include command aliases in the output")

	return cmd
}

// hasNameOrAlias returns true if the provided command as a name or an alias equal to the provided arg.
func hasNameOrAlias(cmd *cobra.Command, arg string) bool {
	if cmd.Name() == arg {
		return true
	}
	for _, alias := range cmd.Aliases {
		if alias == arg {
			return true
		}
	}
	return false
}

// getNameAndAliases gets a list of the command's name and aliases (without any duplicates).
func getNameAndAliases(cmd *cobra.Command) []string {
	rv := make([]string, 1, 1+len(cmd.Aliases))
	rv[0] = cmd.Name()
	for _, alias := range cmd.Aliases {
		known := false
		for _, prev := range rv {
			if prev == alias {
				known = true
			}
		}
		if !known {
			rv = append(rv, alias)
		}
	}
	return rv
}

// getAllCommands gets all of the possible commands under the given command.
func getAllCommands(cmd *cobra.Command, prev string, limitTo []string, incAliases bool) []string {
	if len(limitTo) > 0 {
		if !hasNameOrAlias(cmd, limitTo[0]) {
			return nil
		}
		limitTo = limitTo[1:]
	}

	name := cmd.Name()
	if incAliases {
		names := getNameAndAliases(cmd)
		if len(names) > 1 {
			name = "[" + strings.Join(names, " ") + "]"
		}
	}
	cur := strings.TrimSpace(prev + " " + name)

	subCmds := cmd.Commands()
	if len(subCmds) == 0 {
		return []string{cur}
	}

	var rv []string
	for _, subCmd := range subCmds {
		rv = append(rv, getAllCommands(subCmd, cur, limitTo, incAliases)...)
	}
	return rv
}
