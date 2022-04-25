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

// Flag names and values
const (
	FlagRewardProgramId     = "reward-program-id"
	FlagDistAddr            = "dist-address"
	FlagCoin                = "coin"
	FlagEpochId             = "epoch-id"
	FlagEpochOffset         = "epoch-offset"
	FlagNumEpochs           = "num-epochs"
	FlagEligibilityCriteria = "eligibility-criteria"
	FlagMinimum             = "minimum"
	FlagMaximum             = "maximum"
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
			rewardProgramId, err := cmd.Flags().GetUint64(FlagRewardProgramId)
			if err != nil {
				return err
			}
			distFromAddr, err := cmd.Flags().GetString(FlagDistAddr)
			if err != nil {
				return err
			}
			coinStr, err := cmd.Flags().GetString(FlagCoin)
			if err != nil {
				return err
			}
			coin, err := sdk.ParseCoinNormalized(coinStr)
			if err != nil {
				return err
			}
			epochId, err := cmd.Flags().GetString(FlagEpochId)
			if err != nil {
				return err
			}
			numEpochs, err := cmd.Flags().GetUint64(FlagNumEpochs)
			if err != nil {
				return err
			}
			epochOffset, err := cmd.Flags().GetUint64(FlagEpochOffset)
			if err != nil {
				return err
			}
			minimum, err := cmd.Flags().GetUint64(FlagMinimum)
			if err != nil {
				return err
			}
			maximum, err := cmd.Flags().GetUint64(FlagMaximum)
			if err != nil {
				return err
			}
			eligibilityCriteriaStr, err := cmd.Flags().GetString(FlagEligibilityCriteria)
			if err != nil {
				return err
			}

			var eligibilityCriteria types.EligibilityCriteria
			err = clientCtx.Codec.UnmarshalJSON([]byte(eligibilityCriteriaStr), &eligibilityCriteria)
			if err != nil {
				return fmt.Errorf("unable to parse eligibility criteria : %s", err)
			}

			var proposal govtypes.Content
			switch args[0] {
			case "add":
				proposal = types.NewAddRewardProgramProposal(
					args[1],
					args[2],
					rewardProgramId,
					distFromAddr,
					coin,
					epochId,
					epochOffset,
					numEpochs,
					eligibilityCriteria,
					minimum,
					maximum,
				)
			default:
				return fmt.Errorf("unknown proposal type : %s", args[0])
			}

			deposit, err := sdk.ParseCoinsNormalized(args[3])
			if err != nil {
				return err
			}

			callerAddr := clientCtx.GetFromAddress()
			msg, err := govtypes.NewMsgSubmitProposal(proposal, deposit, callerAddr)
			if err != nil {
				return fmt.Errorf("invalid governance proposal. Error: %q", err)
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	cmd.Flags().Uint64(FlagRewardProgramId, 0, "rewards program id")
	cmd.Flags().String(FlagDistAddr, "", "distribution from address")
	cmd.Flags().String(FlagCoin, "", "coins for reward program")
	cmd.Flags().String(FlagEpochId, "", "epoch identifier (day, week, month)")
	cmd.Flags().Uint64(FlagEpochOffset, 0, "epoch block offset used to calculate start of program")
	cmd.Flags().Uint64(FlagNumEpochs, 0, "number of epochs for the reward program")
	cmd.Flags().String(FlagEligibilityCriteria, "", "json of the eligibility criteria")
	cmd.Flags().Uint64(FlagMinimum, 0, "minimum amount of actions needed for reward program")
	cmd.Flags().Uint64(FlagMaximum, 1, "maximum amount of actions allowed for reward program")
	return cmd
}
