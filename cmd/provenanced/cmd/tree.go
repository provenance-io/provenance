package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

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
			if len(cmds) == 0 {
				cmd.SilenceUsage = true
				return fmt.Errorf("command not found: %q", args)
			}
			cmd.Printf("%s\n", strings.Join(cmds, "\n"))
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
