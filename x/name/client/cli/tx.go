package cli

import (
	"fmt"
	"strings"

	"github.com/provenance-io/provenance/x/name/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// The flag for creating unrestricted names
const flagUnrestricted = "unrestrict"

// NewTxCmd is the top-level command for name CLI transactions.
func NewTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Transaction commands for the name module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	txCmd.AddCommand(
		GetBindNameCmd(),
		GetDeleteNameCmd(),
		GetModifyNameProposalCmd(),
	)
	return txCmd
}

// GetBindNameCmd is the CLI command for binding a name to an address.
func GetBindNameCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "bind [name] [address] [root]",
		Short:   "Bind a name to an address under the given root name in the provenance blockchain",
		Example: fmt.Sprintf(`$ %s tx name bind sample pb1skjwj5whet0lpe65qaq4rpq03hjxlwd9nf39lk root.example`, version.AppName),
		Args:    cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			address, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return err
			}
			msg := types.NewMsgBindNameRequest(
				types.NewNameRecord(
					strings.ToLower(args[0]),
					address,
					!viper.GetBool(flagUnrestricted),
				),
				types.NewNameRecord(
					strings.ToLower(args[2]),
					clientCtx.FromAddress,
					false,
				),
			)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	cmd.Flags().BoolP(flagUnrestricted, "u", false, "Allow child name creation by everyone")

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetDeleteNameCmd is the CLI command for deleting a bound name.
func GetDeleteNameCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete [name]",
		Short:   "Delete a bound name from the provenance blockchain",
		Example: fmt.Sprintf(`$ %s tx name delete sample`, version.AppName),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			msg := types.NewMsgDeleteNameRequest(
				types.NewNameRecord(
					strings.TrimSpace(strings.ToLower(args[0])),
					clientCtx.FromAddress,
					false,
				),
			)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func GetModifyNameProposalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "modify-name-proposal [name] [new_owner] (--unrestrict) [flags]",
		Short: "Submit a modify name creation governance proposal",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit a modify name governance proposal along with an initial deposit.

IMPORTANT: The restricted creation of sub-names will be enabled by default unless the unrestricted flag is included.
The owner must approve all child name creation unless an alterate owner is provided.

Example:
$ %s tx name modify-name-proposal \
	<name> <new_owner> \
	--unrestrict  \ 
	--from <address>
			`,
				version.AppName)),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			owner, err := sdk.AccAddressFromBech32(args[1])
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
			metadata, err := cmd.Flags().GetString(FlagMetadata)
			if err != nil {
				return fmt.Errorf("name metadata: %w", err)
			}

			content := types.MsgModifyNameRequest{
				Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
				Record:    types.NewNameRecord(strings.ToLower(args[0]), owner, !viper.GetBool(flagUnrestricted)),
			}
			if err = content.ValidateBasic(); err != nil {
				return err
			}

			msg, err := govtypesv1.NewMsgSubmitProposal([]sdk.Msg{&content}, deposit, clientCtx.GetFromAddress().String(), metadata)
			if err != nil {
				return err
			}
			if err = msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().BoolP(flagUnrestricted, "u", false, "Allow child name creation by everyone")
	// proposal flags
	cmd.Flags().String(FlagMetadata, "", "Metadata of proposal")
	cmd.Flags().String(FlagDeposit, "", "Deposit of proposal")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
