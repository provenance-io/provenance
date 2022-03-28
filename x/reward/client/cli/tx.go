package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"

	"github.com/provenance-io/provenance/x/reward/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

func NewTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Aliases:                    []string{"r"},
		Short:                      "Transaction commands for the reward module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		GetCmdRewardProgramProposal(),
	)

	return txCmd
}

func GetCmdRewardProgramProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proposal [add|update|remove] [title] [description] [deposit]",
		Args:  cobra.ExactArgs(4),
		Short: "Submit a reward program proposal along with an initial deposit",
		Long: strings.TrimSpace(`Submit a reward program proposal along with an initial deposit.
`),
		// 		Example: fmt.Sprintf(`$ %[1]s tx msgfees add "adding" "adding MsgWriterRecordRequest fee"  10nhash --msg-type=/provenance.metadata.v1.MsgWriteRecordRequest --additional-fee=612nhash
		// $ %[1]s tx msgfees update "updating" "updating MsgWriterRecordRequest fee"  10nhash --msg-type=/provenance.metadata.v1.MsgWriteRecordRequest --additional-fee=612000nhash
		// $ %[1]s tx msgfees remove "removing" "removing MsgWriterRecordRequest fee" 10nhash --msg-type=/provenance.metadata.v1.MsgWriteRecordRequest
		// `, version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// proposalType := args[0]
			// if proposalType != "remove" {
			// 	// TODO

			// }

			deposit, err := sdk.ParseCoinsNormalized(args[3])
			if err != nil {
				return err
			}

			callerAddr := clientCtx.GetFromAddress()
			msg, err := govtypes.NewMsgSubmitProposal(nil, deposit, callerAddr)
			if err != nil {
				return fmt.Errorf("invalid governance proposal. Error: %q", err)
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
