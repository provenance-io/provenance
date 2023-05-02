package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/provenance-io/provenance/x/trigger/types"
	"github.com/spf13/cobra"
)

func GetQueryCmd() *cobra.Command {
	queryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the rewards module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	queryCmd.AddCommand()
	return queryCmd
}
