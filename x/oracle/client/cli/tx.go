package cli

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/version"
	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"

	"github.com/provenance-io/provenance/internal/provcli"
	"github.com/provenance-io/provenance/x/oracle/types"
)

// NewTxCmd is the top-level command for oracle CLI transactions.
func NewTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Transaction commands for the oracle module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		GetCmdSendQuery(),
		GetCmdOracleUpdate(),
	)

	return txCmd
}

// GetCmdOracleUpdate is a command to update the address of the module's oracle
func GetCmdOracleUpdate() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update <address>",
		Short:   "Update the module's oracle address",
		Long:    "Submit an update oracle via governance proposal along with an initial deposit.",
		Args:    cobra.ExactArgs(1),
		Aliases: []string{"u"},
		Example: fmt.Sprintf(`%[1]s tx oracle update pb1skjwj5whet0lpe65qaq4rpq03hjxlwd9nf39lk --deposit 50000nhash`, version.AppName),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			flagSet := cmd.Flags()
			authority := provcli.GetAuthority(flagSet)

			msg := types.NewMsgUpdateOracle(authority, args[0])
			return provcli.GenerateOrBroadcastTxCLIAsGovProp(clientCtx, flagSet, msg)
		},
	}

	govcli.AddGovPropFlagsToCmd(cmd)
	provcli.AddAuthorityFlagToCmd(cmd)
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetCmdSendQuery is a command to send a query to another chain's oracle
func GetCmdSendQuery() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "send-query <channel-id> <json>",
		Short:   "Send a query to an oracle on a remote chain via IBC",
		Args:    cobra.ExactArgs(2),
		Aliases: []string{"sq"},
		Example: fmt.Sprintf(`%[1]s tx oracle send-query channel-1 '{"query_version":{}}'`, version.AppName),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			channelID := args[0]

			queryData := []byte(args[1])
			if !json.Valid(queryData) {
				return errors.New("query data must be json")
			}

			msg := types.NewMsgSendQueryOracle(
				clientCtx.GetFromAddress().String(),
				channelID,
				queryData,
			)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
