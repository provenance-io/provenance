package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/version"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/provenance-io/provenance/x/name/types"
)

// The flag for creating unrestricted names
const FlagUnrestricted = "unrestrict"

// The flag to specify that the command should be ran as a gov proposal
const FlagGovProposal = "gov-proposal"

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
		GetModifyNameCmd(),
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
					!viper.GetBool(FlagUnrestricted),
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
	cmd.Flags().BoolP(FlagUnrestricted, "u", false, "Allow child name creation by everyone")

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

func GetModifyNameCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "modify-name [name] [new_owner] (--unrestrict) [flags]",
		Short: "Submit a modify name tx",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit a modify name by name owner or via governance proposal along with an initial deposit.

IMPORTANT: The restricted creation of sub-names will be enabled by default unless the unrestricted flag is included.
The owner must approve all child name creation unless an alternate owner is provided.

Example:
$ %s tx name modify-name \
	<name> <new_owner> \
	--unrestrict  \ 
	--from <address>
			`,
				version.AppName)),
		Aliases: []string{"m", "mn"},
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			owner, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return err
			}
			isGov, err := cmd.Flags().GetBool(FlagGovProposal)
			if err != nil {
				return err
			}

			var authority sdk.AccAddress
			if !isGov {
				authority = clientCtx.GetFromAddress()
			} else {
				authority = authtypes.NewModuleAddress(govtypes.ModuleName)
			}

			modifyMsg := types.MsgModifyNameRequest{
				Authority: authority.String(),
				Record:    types.NewNameRecord(strings.ToLower(args[0]), owner, !viper.GetBool(FlagUnrestricted)),
			}

			var req sdk.Msg
			if isGov {
				var govErr error
				proposal, govErr := govcli.ReadGovPropFlags(clientCtx, cmd.Flags())
				if govErr != nil {
					return govErr
				}
				anys, govErr := sdktx.SetMsgs([]sdk.Msg{&modifyMsg})
				if govErr != nil {
					return govErr
				}
				proposal.Messages = anys
				req = proposal
			} else {
				req = &modifyMsg
			}
			if err = req.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), req)
		},
	}

	cmd.Flags().BoolP(FlagUnrestricted, "u", false, "Allow child name creation by everyone")
	cmd.Flags().Bool(FlagGovProposal, false, "Run transaction as gov proposal")
	govcli.AddGovPropFlagsToCmd(cmd)
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
