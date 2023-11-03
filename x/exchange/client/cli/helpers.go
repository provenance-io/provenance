package cli

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	FlagMarketID      = "market-id"
	FlagAssets        = "assets"
	FlagPrice         = "price"
	FlagSettlementFee = "settlement-fee"
	FlagAllowPartial  = "allow-partial"
	FlagExternalID    = "external-id"
	FlagCreationFee   = "creation-fee"
)

var ExampleAddr = "pb1v4uxzmtsd3j47h6lta047h6lta047h6llzxa5d" // = sdk.AccAddress("example_____________")

// markFlagRequired this marks a flag as required and panics if there's a problem.
func markFlagRequired(cmd *cobra.Command, name string) {
	if err := cmd.MarkFlagRequired(name); err != nil {
		panic(fmt.Errorf("error marking %s flag required on %s: %w", name, cmd.Name(), err))
	}
}

// readCoinsFlag reads a string flag and converts it into sdk.Coins.
// If the flag wasn't provided, this returns nil, nil.
func readCoinsFlag(flagSet *pflag.FlagSet, name string) (sdk.Coins, error) {
	value, err := flagSet.GetString(name)
	if err != nil {
		return nil, err
	}
	if value == "" {
		return nil, nil
	}
	rv, err := sdk.ParseCoinsNormalized(value)
	if err != nil {
		return nil, fmt.Errorf("error parsing --%s value %q as coins: %w", name, value, err)
	}
	return rv, nil
}

// readCoinFlag reads a string flag and converts it into *sdk.Coin.
// If the flag wasn't provided, this returns nil, nil.
func readCoinFlag(flagSet *pflag.FlagSet, name string) (*sdk.Coin, error) {
	value, err := flagSet.GetString(name)
	if err != nil {
		return nil, err
	}
	if value == "" {
		return nil, nil
	}
	rv, err := sdk.ParseCoinNormalized(value)
	if err != nil {
		return nil, fmt.Errorf("error parsing --%s value %q as a coin: %w", name, value, err)
	}
	return &rv, nil
}

// readReqCoinFlag reads a string flag and converts it into a sdk.Coin and requires it to have a value.
func readReqCoinFlag(flagSet *pflag.FlagSet, name string) (sdk.Coin, error) {
	rv, err := readCoinFlag(flagSet, name)
	if err != nil {
		return sdk.Coin{}, err
	}
	if rv == nil {
		return sdk.Coin{}, fmt.Errorf("missing required --%s flag", name)
	}
	return *rv, nil
}

// AddFlagMarketID adds the --market-id <uint32> flag to a command.
// If required is true, it marks the flag as required for the command.
func AddFlagMarketID(cmd *cobra.Command, required bool) {
	cmd.Flags().Uint32(FlagMarketID, 0, "The market id, e.g. 3")
	if required {
		markFlagRequired(cmd, FlagMarketID)
	}
}

// ReadFlagMarketID reads the --market-id flag.
func ReadFlagMarketID(flagSet *pflag.FlagSet) (uint32, error) {
	return flagSet.GetUint32(FlagMarketID)
}

// AddFlagAssets adds the --assets <string> flag to a command and marks it required.
func AddFlagAssets(cmd *cobra.Command) {
	cmd.Flags().String(FlagAssets, "", "The assets for this order, e.g. 10nhash")
	markFlagRequired(cmd, FlagAssets)
}

// ReadFlagAssets reads the --assets flag as an sdk.Coin.
func ReadFlagAssets(flagSet *pflag.FlagSet) (sdk.Coin, error) {
	return readReqCoinFlag(flagSet, FlagAssets)
}

// AddFlagPrice adds the --price <string> flag to a command and marks it required.
func AddFlagPrice(cmd *cobra.Command) {
	cmd.Flags().String(FlagPrice, "", "The price for this order, e.g. 10nhash")
	markFlagRequired(cmd, FlagPrice)
}

// ReadFlagPrice reads the --price flag as an sdk.Coin.
func ReadFlagPrice(flagSet *pflag.FlagSet) (sdk.Coin, error) {
	return readReqCoinFlag(flagSet, FlagPrice)
}

// AddFlagSettlementFee adds the optional --settlement-fee <string> flag to a command.
func AddFlagSettlementFee(cmd *cobra.Command) {
	cmd.Flags().String(FlagSettlementFee, "", "The settlement fee Coin string for this order, e.g. 10nhash")
}

// ReadFlagSettlementFee reads the --price flag as an sdk.Coin.
func ReadFlagSettlementFee(flagSet *pflag.FlagSet) (sdk.Coins, error) {
	return readCoinsFlag(flagSet, FlagSettlementFee)
}

// AddFlagAllowPartial adds the optional --allow-partial flag to a command.
func AddFlagAllowPartial(cmd *cobra.Command) {
	cmd.Flags().Bool(FlagAllowPartial, false, "Allow this order to be partially filled")
}

// ReadFlagAllowPartial reads the --allow-partial flag.
func ReadFlagAllowPartial(flagSet *pflag.FlagSet) (bool, error) {
	return flagSet.GetBool(FlagAllowPartial)
}

// AddFlagExternalID adds the optional --external-id <string> flag to a command.
func AddFlagExternalID(cmd *cobra.Command) {
	cmd.Flags().String(FlagExternalID, "", "The external id for this order")
}

// ReadFlagExternalID reads the --external-id flag.
func ReadFlagExternalID(flagSet *pflag.FlagSet) (string, error) {
	return flagSet.GetString(FlagExternalID)
}

// AddFlagCreationFee adds the optional --creation-fee <string> flag to a command.
func AddFlagCreationFee(cmd *cobra.Command) {
	cmd.Flags().String(FlagCreationFee, "", "The order creation fee, e.g. 10nhash")
}

// ReadFlagCreationFee reads the --creation-fee flag as an sdk.Coin.
func ReadFlagCreationFee(flagSet *pflag.FlagSet) (*sdk.Coin, error) {
	return readCoinFlag(flagSet, FlagCreationFee)
}

// AddCreateOrderFlags adds all the flags used in the order creation commands.
func AddCreateOrderFlags(cmd *cobra.Command) {
	AddFlagMarketID(cmd, true)
	AddFlagAssets(cmd)
	AddFlagPrice(cmd)
	AddFlagSettlementFee(cmd)
	AddFlagAllowPartial(cmd)
	AddFlagExternalID(cmd)
	AddFlagCreationFee(cmd)
}

// CreateOrderFlags represents all the flags used in the order creation commands.
type CreateOrderFlags struct {
	MarketID      uint32
	Assets        sdk.Coin
	Price         sdk.Coin
	SettlementFee sdk.Coins
	AllowPartial  bool
	ExternalID    string
	CreationFee   *sdk.Coin
}

// ReadCreateOrderFlags reads all the order creation commands from the given FlagSet.
func ReadCreateOrderFlags(flagSet *pflag.FlagSet) (*CreateOrderFlags, error) {
	errs := make([]error, 7)
	rv := &CreateOrderFlags{}
	rv.MarketID, errs[0] = ReadFlagMarketID(flagSet)
	rv.Assets, errs[1] = ReadFlagAssets(flagSet)
	rv.Price, errs[2] = ReadFlagPrice(flagSet)
	rv.SettlementFee, errs[3] = ReadFlagSettlementFee(flagSet)
	rv.AllowPartial, errs[4] = ReadFlagAllowPartial(flagSet)
	rv.ExternalID, errs[5] = ReadFlagExternalID(flagSet)
	rv.CreationFee, errs[6] = ReadFlagCreationFee(flagSet)
	return rv, errors.Join(errs...)
}
