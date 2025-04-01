package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"

	"github.com/provenance-io/provenance/internal/provcli"
	"github.com/provenance-io/provenance/x/flatfees/types"
)

// Flag names and values
const (
	FlagMinFee    = "additional-fee"
	FlagMsgType   = "msg-type"
	FlagRecipient = "recipient"
	FlagBips      = "bips"

	FlagSet   = "set"
	FlagUnset = "unset"
)

func NewTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Aliases:                    []string{"fees", "ff"},
		Short:                      "Transaction commands for the x/flatfees module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		NewCmdUpdateParams(),
		NewCmdUpdateMsgFees(),
	)

	return txCmd
}

var cmdStart = fmt.Sprintf("%s tx flatfees", version.AppName)

// ParseDefaultCost will parse the default cost coin string.
func ParseDefaultCost(arg string) (sdk.Coin, error) {
	rv, err := sdk.ParseCoinNormalized(arg)
	if err != nil {
		return rv, fmt.Errorf("invalid default cost %q: %w", arg, err)
	}
	return rv, nil
}

// ParseConversionFactor will parse a string with the format "<base>=<converted>" into a ConversionFactor.
// Both <base> and <convert> must be a single coin string.
func ParseConversionFactor(arg string) (types.ConversionFactor, error) {
	parts := strings.Split(arg, "=")
	if len(parts) != 2 {
		return types.ConversionFactor{}, fmt.Errorf("invalid conversion factor %q: expected exactly one equals sign", arg)
	}

	var err error
	rv := types.ConversionFactor{}

	rv.BaseAmount, err = sdk.ParseCoinNormalized(parts[0])
	if err != nil {
		return types.ConversionFactor{}, fmt.Errorf("invalid conversion factor %q: invalid base amount %q: %w", arg, parts[0], err)
	}

	rv.ConvertedAmount, err = sdk.ParseCoinNormalized(parts[1])
	if err != nil {
		return types.ConversionFactor{}, fmt.Errorf("invalid conversion factor %q: invalid converted amount %q: %w", arg, parts[1], err)
	}

	return rv, nil
}

// ReadFlagSet reads the --set flag values from the provided flagSet.
func ReadFlagSet(flagSet *pflag.FlagSet) ([]*types.MsgFee, error) {
	args, err := flagSet.GetStringArray(FlagSet)
	if err != nil {
		return nil, err
	}
	rv := make([]*types.MsgFee, len(args))
	for i, arg := range args {
		rv[i], err = ParseMsgFee(arg)
		if err != nil {
			return nil, err
		}
	}
	return rv, nil
}

// ParseMsgFee parses a string like "<msg-type-url>=<cost>" into a MsgFee.
func ParseMsgFee(arg string) (*types.MsgFee, error) {
	parts := strings.Split(arg, "=")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid set arg %q: expected exactly one equals sign", arg)
	}

	cost, err := sdk.ParseCoinsNormalized(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid set arg %q: invalid cost %q: %w", arg, parts[1], err)
	}

	return &types.MsgFee{MsgTypeUrl: parts[0], Cost: cost}, nil
}

// NewCmdUpdateParams creates the cmd that will submit a gov prop to update x/flatfees params.
func NewCmdUpdateParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params <default cost> <base>=<converted> <gov prop flags>",
		Short: "Submit a governance proposal to update the x/flatfees module params",
		Long: strings.TrimSpace(`Submit a governance proposal to update the x/flatfees module params.

The <default cost> is a standard Coin string.
The <base>=<converted> arg is the conversion factor, and each should be standard coin strings.
The denominations in the <default cost> and <base> should be the same.
`),
		Example: fmt.Sprintf("$ %[1]s params 100%[2]s 5%[2]s=7nhash'", cmdStart, types.DefaultFeeDefinitionDenom),
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			flagSet := cmd.Flags()
			msg := &types.MsgUpdateParamsRequest{
				Authority: provcli.GetAuthority(flagSet),
			}

			msg.Params.DefaultCost, err = ParseDefaultCost(args[0])
			if err != nil {
				return err
			}

			msg.Params.ConversionFactor, err = ParseConversionFactor(args[1])
			if err != nil {
				return err
			}

			return provcli.GenerateOrBroadcastTxCLIAsGovProp(clientCtx, flagSet, msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	govcli.AddGovPropFlagsToCmd(cmd)
	provcli.AddAuthorityFlagToCmd(cmd)
	return cmd
}

// NewCmdUpdateMsgFees creates the cmd that will submit a gov prop to update costs for specific Msgs.
func NewCmdUpdateMsgFees() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update [--set <msg-type-url>=<cost> [...]] [--unset <msg-type-url> [...]] <gov prop flags>",
		Aliases: []string{"costs"},
		Short:   "Submit a governance proposal to update the msg fees",
		Long: strings.TrimSpace(`Submit a governance proposal to update the msg fees.

To set or update the cost of a Msg, use the --set <msg-type-url>=<cost> option.
To unset a Msg (revert to default), use the --unset <msg-type-url> option.
Both --set and --unset can be provided multiple times, and multiple entries can be included after each.
`),
		Example: strings.TrimSpace(fmt.Sprintf(`$ %[1]s update --set '/cosmos.bank.v1beta1.MsgSend=5%[2]s'
$ %[1]s update --unset '/cosmos.bank.v1beta1.MsgSend'
`, cmdStart, types.DefaultFeeDefinitionDenom)),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			flagSet := cmd.Flags()
			msg := &types.MsgUpdateMsgFeesRequest{
				Authority: provcli.GetAuthority(flagSet),
			}

			msg.ToSet, err = ReadFlagSet(flagSet)
			if err != nil {
				return err
			}
			for _, msgFee := range msg.ToSet {
				if _, err = clientCtx.InterfaceRegistry.Resolve(msgFee.MsgTypeUrl); err != nil {
					return err
				}
			}

			msg.ToUnset, err = flagSet.GetStringSlice(FlagUnset)
			if err != nil {
				return err
			}
			for _, url := range msg.ToUnset {
				if _, err = clientCtx.InterfaceRegistry.Resolve(url); err != nil {
					return err
				}
			}

			return provcli.GenerateOrBroadcastTxCLIAsGovProp(clientCtx, flagSet, msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	govcli.AddGovPropFlagsToCmd(cmd)
	provcli.AddAuthorityFlagToCmd(cmd)
	cmd.Flags().StringArray(FlagSet, nil, "One or more MsgFees to set, arg format is <msg-type-url>=<cost>")
	cmd.Flags().StringSlice(FlagUnset, nil, "One or more msg type urls to unset")
	return cmd
}
