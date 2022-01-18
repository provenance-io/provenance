package cli

import (
	"fmt"
	"strings"

	"github.com/provenance-io/provenance/x/name/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// The flag for the owner address of a name record
const flagOwner = "owner"

// GetRootNameProposalCmd returns a command for registration with the gov module
func GetRootNameProposalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "root-name-proposal [name] --restrict --owner [address] --title [text] --description [text]",
		Short: "Submit a root name proposal",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			proposalTitle, err := cmd.Flags().GetString(cli.FlagTitle)
			if err != nil {
				return fmt.Errorf("proposal title: %s", err)
			}
			proposalDescr, err := cmd.Flags().GetString(cli.FlagDescription)
			if err != nil {
				return fmt.Errorf("proposal description: %s", err)
			}
			proposalOwner, err := cmd.Flags().GetString(flagOwner)
			if err != nil {
				return fmt.Errorf("proposal root name owner: %s", err)
			}
			if len(proposalOwner) < 1 {
				proposalOwner = clientCtx.GetFromAddress().String()
			}
			_, err = sdk.AccAddressFromBech32(proposalOwner)
			if err != nil {
				return err
			}
			depositArg, err := cmd.Flags().GetString(cli.FlagDeposit)
			if err != nil {
				return err
			}
			deposit, err := sdk.ParseCoinsNormalized(depositArg)
			if err != nil {
				return err
			}

			content := types.CreateRootNameProposal{
				Title:       proposalTitle,
				Description: proposalDescr,
				Name:        strings.ToLower(args[0]),
				Owner:       proposalOwner,
				Restricted:  viper.GetBool(flagRestricted),
			}

			msg, err := govtypes.NewMsgSubmitProposal(&content, deposit, clientCtx.GetFromAddress())
			if err != nil {
				return err
			}
			if err = msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(flagOwner, "", "The owner of the new name, optional (defaults to from address")
	cmd.Flags().BoolP(flagRestricted, "r", true, "Restrict creation of child names to owner only, optional (default false)")
	// proposal flags
	cmd.Flags().String(cli.FlagTitle, "", "Title of proposal")
	cmd.Flags().String(cli.FlagDescription, "", "Description of proposal")
	cmd.Flags().String(cli.FlagDeposit, "", "Deposit of proposal")
	return cmd
}
