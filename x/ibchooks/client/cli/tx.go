package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"

	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"

	"github.com/provenance-io/provenance/internal/provcli"
	"github.com/provenance-io/provenance/x/ibchooks/types"
)

// NewTxCmd is the top-level command for attribute CLI transactions.
func NewTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Aliases:                    []string{"ih"},
		Short:                      "Transaction commands for the ibchooks module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	txCmd.AddCommand(
		NewUpdateParamsCmd(),
	)
	return txCmd
}

// NewUpdateParamsCmd creates a command to update the ibchooks module's params via governance proposal.
func NewUpdateParamsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update-params <allowed-async-ack-contracts>",
		Short:   "Update the ibchooks module's params via governance proposal",
		Long:    "Submit an update params via governance proposal along with an initial deposit.",
		Args:    cobra.ExactArgs(1),
		Example: fmt.Sprintf(`%[1]s tx ibchooks update-params contract1,contract2 --deposit 50000nhash`, version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			flagSet := cmd.Flags()
			authority := provcli.GetAuthority(flagSet)
			allowedAsyncAckContracts := strings.Split(args[0], ",")

			msg := types.NewMsgUpdateParamsRequest(allowedAsyncAckContracts, authority)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return provcli.GenerateOrBroadcastTxCLIAsGovProp(clientCtx, flagSet, msg)
		},
	}

	govcli.AddGovPropFlagsToCmd(cmd)
	provcli.AddAuthorityFlagToCmd(cmd)
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
