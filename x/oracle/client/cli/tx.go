package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/provenance-io/provenance/x/oracle/types"
	"github.com/spf13/cobra"
)

// NewTxCmd is the top-level command for trigger CLI transactions.
func NewTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Aliases:                    []string{"t"},
		Short:                      "Transaction commands for the trigger module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand()

	return txCmd
}
