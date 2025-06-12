package cmd

import (
	"strings"

	"github.com/spf13/cobra"

	flatfeescli "github.com/provenance-io/provenance/x/flatfees/client/cli"
)

func GetCmdPioSimulateTx() *cobra.Command {
	// This is the command same as the one for the flatfees CalculateTxFees query.
	// However, we want this one named "simulate". So, if it's not, we put the current name in the aliases
	// remove "simulate" from the aliases, and then update the use to have the name "simulate".
	cmd := flatfeescli.NewCmdCalculateTxFees()

	// If it's already called "simulate", there's nothing more we need to do here.
	origName := cmd.Name()
	newName := "simulate"
	if origName == newName {
		return cmd
	}

	// Update the use line to have the new name with the same use string.
	cmd.Use = newName + strings.TrimPrefix(cmd.Use, origName)

	// If "simulate" is in the alias list, replace it with the original command name.
	for i, alias := range cmd.Aliases {
		if alias == newName {
			cmd.Aliases[i] = origName
			return cmd
		}
	}

	// If "simulate" wasn't in the alias list, make the original command name as the first alias.
	newAliases := make([]string, len(cmd.Aliases)+1)
	newAliases[0] = origName
	copy(newAliases[1:], cmd.Aliases)
	cmd.Aliases = newAliases

	return cmd
}
