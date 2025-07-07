package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"

	"github.com/provenance-io/provenance/internal/provcli"
	"github.com/provenance-io/provenance/x/name/types"
)

const (
	// FlagOwner is the flag for the owner address of a name record.
	FlagOwner = "owner"

	// FlagGovProposal is the flag to specify that the command should be run as a gov proposal
	FlagGovProposal = "gov-proposal"

	// FlagUnrestricted is the flag for creating unrestricted names
	FlagUnrestricted = "unrestrict"
)

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
		GetGovRootNameCmd(),
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

			modifyMsg := &types.MsgModifyNameRequest{
				Record: types.NewNameRecord(strings.ToLower(args[0]), owner, !viper.GetBool(FlagUnrestricted)),
			}

			if isGov {
				modifyMsg.Authority = provcli.GetAuthority(cmd.Flags())
				return provcli.GenerateOrBroadcastTxCLIAsGovProp(clientCtx, cmd.Flags(), modifyMsg)
			}

			modifyMsg.Authority = clientCtx.GetFromAddress().String()
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), modifyMsg)
		},
	}

	cmd.Flags().BoolP(FlagUnrestricted, "u", false, "Allow child name creation by everyone")
	cmd.Flags().Bool(FlagGovProposal, false, "Run transaction as gov proposal")
	govcli.AddGovPropFlagsToCmd(cmd)
	provcli.AddAuthorityFlagToCmd(cmd)
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// GetGovRootNameCmd returns a command for registration with the gov module
func GetGovRootNameCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gov-root-name [name] (--owner [address]) (--unrestrict) [flags]",
		Short: "Submit a root name creation governance proposal",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit a root name creation governance proposal.

IMPORTANT: The created root name will restrict the creation of sub-names by default unless the
unrestrict flag is included. The proposer will be the default owner that must approve
all child name creation unless an alterate owner is provided.

Example:
$ %s tx name gov-root-name \
	<root name> \
	--unrestrict  \ 
	--owner <key_or_address> \
	--from <key_or_address>
			`,
				version.AppName)),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			flagSet := cmd.Flags()
			owner, err := owner(clientCtx, flagSet)
			if err != nil {
				return err
			}
			authority := provcli.GetAuthority(flagSet)
			name := strings.ToLower(args[0])
			restricted := !viper.GetBool(FlagUnrestricted)
			msg := types.NewMsgCreateRootNameRequest(authority, name, owner, restricted)

			return provcli.GenerateOrBroadcastTxCLIAsGovProp(clientCtx, flagSet, msg)
		},
	}

	cmd.Flags().String(FlagOwner, "", "The owner of the new name, optional (defaults to from address)")
	cmd.Flags().BoolP(FlagUnrestricted, "u", false, "Allow child name creation by everyone")
	govcli.AddGovPropFlagsToCmd(cmd)
	provcli.AddAuthorityFlagToCmd(cmd)
	return cmd
}

// owner returns the proposal owner
func owner(ctx client.Context, flags *pflag.FlagSet) (string, error) {
	proposalOwner, err := flags.GetString(FlagOwner)
	if err != nil {
		return "", fmt.Errorf("proposal root name owner: %w", err)
	}
	if len(proposalOwner) < 1 {
		proposalOwner = ctx.GetFromAddress().String()
	}
	_, err = sdk.AccAddressFromBech32(proposalOwner)
	if err != nil {
		return "", err
	}

	return proposalOwner, nil
}

// GetUpdateNameParamsCmd creates a command to update the name module's params via governance proposal.
func GetUpdateNameParamsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update-name-params <max-segment-length> <min-segment-length> <max-name-levels> <allow-unrestricted-names>",
		Short:   "Update the name module's params via governance proposal",
		Long:    "Submit an update name params via governance proposal along with an initial deposit.",
		Args:    cobra.ExactArgs(4),
		Example: fmt.Sprintf(`%[1]s tx name update-name-params 16 2 5 true --deposit 50000nhash`, version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			flagSet := cmd.Flags()
			authority := provcli.GetAuthority(flagSet)

			maxSegmentLength, err := strconv.ParseUint(args[0], 10, 32)
			if err != nil {
				return fmt.Errorf("invalid max segment length: %w", err)
			}

			minSegmentLength, err := strconv.ParseUint(args[1], 10, 32)
			if err != nil {
				return fmt.Errorf("invalid min segment length: %w", err)
			}

			maxNameLevels, err := strconv.ParseUint(args[2], 10, 32)
			if err != nil {
				return fmt.Errorf("invalid max name levels: %w", err)
			}

			allowUnrestrictedNames, err := strconv.ParseBool(args[3])
			if err != nil {
				return fmt.Errorf("invalid allow unrestricted names flag: %w", err)
			}

			msg := types.NewMsgUpdateParamsRequest(
				uint32(maxSegmentLength), //nolint:gosec // G115: ParseUint bitsize is 32, so we know this is okay.
				uint32(minSegmentLength), //nolint:gosec // G115: ParseUint bitsize is 32, so we know this is okay.
				uint32(maxNameLevels),    //nolint:gosec // G115: ParseUint bitsize is 32, so we know this is okay.
				allowUnrestrictedNames,
				authority,
			)
			return provcli.GenerateOrBroadcastTxCLIAsGovProp(clientCtx, flagSet, msg)
		},
	}

	govcli.AddGovPropFlagsToCmd(cmd)
	provcli.AddAuthorityFlagToCmd(cmd)
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
