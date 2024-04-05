package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	// govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli" // TODO[1760]: gov-cli
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/provenance-io/provenance/x/ibcratelimit"
)

// NewTxCmd is the top-level command for oracle CLI transactions.
func NewTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        ibcratelimit.ModuleName,
		Aliases:                    []string{"rl"},
		Short:                      "Transaction commands for the ibcratelimit module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		GetCmdParamsUpdate(),
	)

	return txCmd
}

// GetCmdParamsUpdate is a command to update the params of the module's rate limiter.
func GetCmdParamsUpdate() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update-params <address>",
		Short:   "Update the module's params",
		Long:    "Submit an update params via governance proposal along with an initial deposit.",
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"u"},
		Example: fmt.Sprintf(`%[1]s tx ratelimitedibc update-params pb1skjwj5whet0lpe65qaq4rpq03hjxlwd9nf39lk --deposit 50000nhash`, version.AppName),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			authority := authtypes.NewModuleAddress(govtypes.ModuleName)

			msg := ibcratelimit.NewMsgGovUpdateParamsRequest(
				authority.String(),
				args[0],
			)

			// TODO[1760]: gov-cli: Replace this with GenerateOrBroadcastTxCLIAsGovProp once its back.
			_, _ = msg, clientCtx
			/*
				proposal, govErr := govcli.ReadGovPropFlags(clientCtx, cmd.Flags())
				if govErr != nil {
					return govErr
				}
				proposal.Messages, govErr = sdktx.SetMsgs([]sdk.Msg{msg})
				if govErr != nil {
					return govErr
				}

				return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), proposal)
			*/
			return fmt.Errorf("not yet updated")
		},
	}

	// govcli.AddGovPropFlagsToCmd(cmd) // TODO[1760]: gov-cli
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
