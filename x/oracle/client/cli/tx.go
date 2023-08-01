package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
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

	txCmd.AddCommand(
		CmdSendQueryAllBalances(),
	)

	return txCmd
}

var _ = strconv.Itoa(0)

func CmdSendQueryAllBalances() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send-query-all-balances [channel-id] [address]",
		Short: "Query the balances of an account on the remote chain via ICQ",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			msg := types.NewMsgSendQueryAllBalances(
				clientCtx.GetFromAddress().String(),
				args[0], // channel id
				args[1], // address
				pageReq,
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "send query all balances")

	return cmd
}
