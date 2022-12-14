package cli

import (
	"fmt"
	"strings"

	"github.com/provenance-io/provenance/x/name/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	// FlagOwner is the flag for the owner address of a name record.
	FlagOwner = "owner"
	// FlagDescription is the flag for a description.
	FlagDescription = "description"
	// FlagTitle is the flag for a title.
	FlagTitle = "title"
	// FlagDeposit is the flag for a deposit.
	FlagDeposit string = "deposit"
)

func GetProposalCmd() *cobra.Command {
	return nil
}

// GetRootNameProposalCmd returns a command for registration with the gov module
func GetRootNameProposalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "root-name-proposal [name] (--owner [address]) (--restrict) [flags]",
		Short: "Submit a root name creation governance proposal",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit a root name creation governance proposal along with an initial deposit.
The proposal title and description must be provided through their respective flags.

IMPORTANT: The created root name will not restrict the creation of sub-names by default unless the
restrict flag is included.  When included the proposer will be the default owner that must approve
all child name creation unless an alterate owner is provided.

Example:
$ %s tx gov submit-legacy-proposal \
    root-name-proposal \
	<root name> \
	--restrict  \ 
	--owner <key_or_address> \
	--title "Proposal title" \
	--description "Description of proposal" 
	--from <key_or_address>
			`,
				version.AppName)),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			proposalTitle, err := cmd.Flags().GetString(FlagTitle)
			if err != nil {
				return fmt.Errorf("proposal title: %w", err)
			}
			proposalDescr, err := cmd.Flags().GetString(FlagDescription)
			if err != nil {
				return fmt.Errorf("proposal description: %w", err)
			}
			proposalOwner, err := cmd.Flags().GetString(FlagOwner)
			if err != nil {
				return fmt.Errorf("proposal root name owner: %w", err)
			}
			if len(proposalOwner) < 1 {
				proposalOwner = clientCtx.GetFromAddress().String()
			}
			_, err = sdk.AccAddressFromBech32(proposalOwner)
			if err != nil {
				return err
			}
			depositArg, err := cmd.Flags().GetString(FlagDeposit)
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

			msg, err := govtypesv1beta1.NewMsgSubmitProposal(&content, deposit, clientCtx.GetFromAddress())
			if err != nil {
				return err
			}
			if err = msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(FlagOwner, "", "The owner of the new name, optional (defaults to from address)")
	cmd.Flags().BoolP(flagRestricted, "r", true, "Restrict creation of child names to owner only, optional (default false)")
	// proposal flags
	cmd.Flags().String(FlagTitle, "", "Title of proposal")
	cmd.Flags().String(FlagDescription, "", "Description of proposal")
	cmd.Flags().String(FlagDeposit, "", "Deposit of proposal")
	return cmd
}

func GetModifyNameProposalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "modify-name-proposal [name] (--owner [address]) (--restrict) [flags]",
		Short: "Submit a modify name creation governance proposal",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit a modify name governance proposal along with an initial deposit.
The proposal title and description must be provided through their respective flags.

IMPORTANT: The creation of sub-names will be enabled by default unless the restrict flag is included.
When included the owner must approve all child name creation unless an alterate owner is provided.

Example:
$ %s tx gov submit-legacy-proposal \
	modify-name-proposal \
	<name> \
	--restrict  \ 
	--owner <key_or_address> \
	--title "Proposal title" \
	--description "Description of proposal" 
	--from <key_or_address>
			`,
				version.AppName)),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			proposalTitle, err := cmd.Flags().GetString(FlagTitle)
			if err != nil {
				return fmt.Errorf("proposal title: %w", err)
			}
			proposalDescr, err := cmd.Flags().GetString(FlagDescription)
			if err != nil {
				return fmt.Errorf("proposal description: %w", err)
			}
			proposalOwner, err := cmd.Flags().GetString(FlagOwner)
			if err != nil {
				return fmt.Errorf("proposal name owner: %w", err)
			}
			if len(proposalOwner) < 1 {
				proposalOwner = clientCtx.GetFromAddress().String()
			}
			_, err = sdk.AccAddressFromBech32(proposalOwner)
			if err != nil {
				return err
			}
			depositArg, err := cmd.Flags().GetString(FlagDeposit)
			if err != nil {
				return err
			}
			deposit, err := sdk.ParseCoinsNormalized(depositArg)
			if err != nil {
				return err
			}

			content := types.ModifyNameProposal{
				Title:       proposalTitle,
				Description: proposalDescr,
				Name:        strings.ToLower(args[0]),
				Owner:       proposalOwner,
				Restricted:  viper.GetBool(flagRestricted),
			}

			msg, err := govtypesv1beta1.NewMsgSubmitProposal(&content, deposit, clientCtx.GetFromAddress())
			if err != nil {
				return err
			}
			if err = msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(FlagOwner, "", "The new owner of the name, optional (defaults to from address)")
	cmd.Flags().BoolP(flagRestricted, "r", true, "Restrict creation of child names to owner only, optional (default false)")
	// proposal flags
	cmd.Flags().String(FlagTitle, "", "Title of proposal")
	cmd.Flags().String(FlagDescription, "", "Description of proposal")
	cmd.Flags().String(FlagDeposit, "", "Deposit of proposal")
	return cmd
}
