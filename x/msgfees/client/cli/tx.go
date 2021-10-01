package cli

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"

	"github.com/provenance-io/provenance/x/msgfees/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

func NewTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Aliases:                    []string{"mf", "mfees"},
		Short:                      "Transaction commands for the msgfees module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		GetCmdMsgBasedFeesProposal(),
	)

	return txCmd
}

func GetCmdMsgBasedFeesProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proposal [type] [proposal-file] [deposit]",
		Args:  cobra.ExactArgs(3),
		Short: "Submit a marker proposal along with an initial deposit",
		Long: strings.TrimSpace(`Submit a msg fees proposal along with an initial deposit.
Proposal title, description, deposit, and marker proposal params must be set in a provided JSON file.

`,
		),
		Example: "TODO",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			contents, err := ioutil.ReadFile(args[1])
			if err != nil {
				return err
			}

			var proposal govtypes.Content

			switch args[0] {
			case types.ProposalTypeAddMsgBasedFees:
				proposal = &types.AddMsgBasedFeesProposal{}
			case types.ProposalTypeUpdateMsgBasedFees:
				proposal = &types.UpdateMsgBasedFeesProposal{}
			case types.ProposalTypeRemoveMsgBasedFees:
				proposal = &types.RemoveMsgBasedFeesProposal{}
			default:
				return fmt.Errorf("unknown proposal type %s", args[0])
			}
			err = json.Unmarshal(contents, proposal)
			if err != nil {
				return err
			}

			deposit, err := sdk.ParseCoinsNormalized(args[2])
			if err != nil {
				return err
			}

			callerAddr := clientCtx.GetFromAddress()
			msg, err := govtypes.NewMsgSubmitProposal(proposal, deposit, callerAddr)
			if err != nil {
				return fmt.Errorf("invalid governance proposal. Error: %s", err)
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
