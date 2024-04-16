package provcli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

const (
	FlagAuthority = "authority"
)

var (
	// DefaultAuthorityAddr is the default authority to provide governance proposal messages.
	// It is defined as a sdk.AccAddress to be independent of global bech32 HRP definition.
	DefaultAuthorityAddr = authtypes.NewModuleAddress(govtypes.ModuleName)
)

// AddAuthorityFlagToCmd adds the authority flag to a command.
func AddAuthorityFlagToCmd(cmd *cobra.Command) {
	// Note: Not setting a default here because the HRP might not yet be set correctly.
	cmd.Flags().String(FlagAuthority, "", "The authority to use. If not provided, a default is used")
}

// GetAuthority gets the authority string from the flagSet or returns the default.
func GetAuthority(flagSet *pflag.FlagSet) string {
	// Ignoring the error here since we really don't care,
	// and it's easier if this just returns a string.
	authority, _ := flagSet.GetString(FlagAuthority)
	if len(authority) > 0 {
		return authority
	}
	return DefaultAuthorityAddr.String()
}

// GenerateOrBroadcastTxCLIAsGovProp wraps the provided msgs in a governance proposal
// and calls GenerateOrBroadcastTxCLI for that proposal. At least one msg is required.
//
// This uses flags added by govcli.AddGovPropFlagsToCmd to fill in the rest of the proposal.
func GenerateOrBroadcastTxCLIAsGovProp(clientCtx client.Context, flagSet *pflag.FlagSet, msgs ...sdk.Msg) error {
	if len(msgs) == 0 {
		return fmt.Errorf("no messages to submit")
	}

	prop, err := govcli.ReadGovPropFlags(clientCtx, flagSet)
	if err != nil {
		return err
	}

	err = prop.SetMsgs(msgs)
	if err != nil {
		return fmt.Errorf("error wrapping msg(s) as Any: %w", err)
	}

	return tx.GenerateOrBroadcastTxCLI(clientCtx, flagSet, prop)
}
