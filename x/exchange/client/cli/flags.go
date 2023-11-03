package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

const (
	FlagAdmin         = "admin"
	FlagAsks          = "asks"
	FlagAssets        = "assets"
	FlagAuthority     = "authority"
	FlagBids          = "bids"
	FlagBuyer         = "buyer"
	FlagCreationFee   = "creation-fee"
	FlagExternalID    = "external-id"
	FlagMarket        = "market"
	FlagOrder         = "order"
	FlagPartial       = "partial"
	FlagPrice         = "price"
	FlagSeller        = "seller"
	FlagSettlementFee = "settlement-fee"
	FlagSigner        = "signer"
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

// readOrderIdsFlag reads a UintSlice flag and converts it into a []uint64.
func readOrderIdsFlag(flagSet *pflag.FlagSet, name string) ([]uint64, error) {
	ids, err := flagSet.GetUintSlice(name)
	if err != nil {
		return nil, err
	}
	rv := make([]uint64, len(ids))
	for i, id := range ids {
		rv[i] = uint64(id)
	}
	return rv, nil
}

// readAddrOrDefault gets the requested flag or, if it wasn't provided, gets the --from address.
func readAddrOrDefault(clientCtx client.Context, flagSet *pflag.FlagSet, name string) (string, error) {
	rv, err := flagSet.GetString(name)
	if err != nil || len(rv) > 0 {
		return rv, err
	}

	rv = clientCtx.FromAddress.String()
	if len(rv) > 0 {
		return rv, nil
	}

	return "", fmt.Errorf("no %s provided", name)
}

// AddFlagAdmin adds the optional --admin flag to a command.
// Also adds the --authority bool flag to the command.
func AddFlagAdmin(cmd *cobra.Command) {
	cmd.Flags().String(FlagAdmin, "", "The admin (defaults to --from account)")
	cmd.Flags().Bool(FlagAuthority, false, "Use the governance module account for the admin")
}

// ReadFlagAdminOrDefault reads the --admin flag if provided.
// If not, but the --authority flag was provided, the gov module account address is returned.
// If no -admin or --authority flag was provided, returns the --from address.
// Returns an error if none of those flags were provided or there was an error reading one.
func ReadFlagAdminOrDefault(clientCtx client.Context, flagSet *pflag.FlagSet) (string, error) {
	rv, err := flagSet.GetString(FlagAdmin)
	if err != nil {
		return "", err
	}

	useAuth, err := flagSet.GetBool(FlagAuthority)
	if err != nil {
		return "", err
	}

	if len(rv) > 0 {
		if useAuth {
			return "", fmt.Errorf("cannot provide both --%s <admin> and --%s", FlagAdmin, FlagAuthority)
		}
		return rv, nil
	}

	if useAuth {
		return authtypes.NewModuleAddress(govtypes.ModuleName).String(), nil
	}

	rv = clientCtx.FromAddress.String()
	if len(rv) > 0 {
		return rv, nil
	}

	return "", fmt.Errorf("no %s provided", FlagAdmin)
}

// AddFlagAsks adds the required --asks <uints> flag to a command.
func AddFlagAsks(cmd *cobra.Command) {
	cmd.Flags().UintSlice(FlagAsks, nil, "The ask order ids")
	markFlagRequired(cmd, FlagAsks)
}

// ReadFlagAsks reads the --asks flag.
func ReadFlagAsks(flagSet *pflag.FlagSet) ([]uint64, error) {
	return readOrderIdsFlag(flagSet, FlagAsks)
}

// AddFlagAssets adds the required --assets <string> flag to a command for providing an order's assets.
func AddFlagAssets(cmd *cobra.Command) {
	cmd.Flags().String(FlagAssets, "", "The assets for this order, e.g. 10nhash")
	markFlagRequired(cmd, FlagAssets)
}

// ReadFlagAssets reads the --assets flag as sdk.Coin.
func ReadFlagAssets(flagSet *pflag.FlagSet) (sdk.Coin, error) {
	return readReqCoinFlag(flagSet, FlagAssets)
}

// AddFlagTotalAssets adds the required --assets <string> flag to a command for providing total assets.
func AddFlagTotalAssets(cmd *cobra.Command) {
	cmd.Flags().String(FlagAssets, "", "The total assets you are filling, e.g. 10nhash")
	markFlagRequired(cmd, FlagAssets)
}

// ReadFlagTotalAssets reads the --assets flag as sdk.Coins.
func ReadFlagTotalAssets(flagSet *pflag.FlagSet) (sdk.Coins, error) {
	return readCoinsFlag(flagSet, FlagAssets)
}

// AddFlagBids adds the required --bids <uints> flag to a command.
func AddFlagBids(cmd *cobra.Command) {
	cmd.Flags().UintSlice(FlagBids, nil, "The bid order ids")
	markFlagRequired(cmd, FlagBids)
}

// ReadFlagBids reads the --bids flag.
func ReadFlagBids(flagSet *pflag.FlagSet) ([]uint64, error) {
	return readOrderIdsFlag(flagSet, FlagBids)
}

// AddFlagBuyer adds the optional --buyer flag to a command.
func AddFlagBuyer(cmd *cobra.Command) {
	cmd.Flags().String(FlagBuyer, "", "The buyer (defaults to --from account)")
}

// ReadFlagBuyerOrDefault reads the --buyer flag if provided, or returns the --from address.
// Returns an error if neither of those flags were provided, or there was an error reading one.
func ReadFlagBuyerOrDefault(clientCtx client.Context, flagSet *pflag.FlagSet) (string, error) {
	return readAddrOrDefault(clientCtx, flagSet, FlagBuyer)
}

// AddFlagCreationFee adds the optional --creation-fee <string> flag to a command.
func AddFlagCreationFee(cmd *cobra.Command) {
	cmd.Flags().String(FlagCreationFee, "", "The order creation fee, e.g. 10nhash")
}

// ReadFlagCreationFee reads the --creation-fee flag as sdk.Coin.
func ReadFlagCreationFee(flagSet *pflag.FlagSet) (*sdk.Coin, error) {
	return readCoinFlag(flagSet, FlagCreationFee)
}

// AddFlagExternalID adds the optional --external-id <string> flag to a command.
func AddFlagExternalID(cmd *cobra.Command) {
	cmd.Flags().String(FlagExternalID, "", "The external id for this order")
}

// ReadFlagExternalID reads the --external-id flag.
func ReadFlagExternalID(flagSet *pflag.FlagSet) (string, error) {
	return flagSet.GetString(FlagExternalID)
}

// AddFlagMarket adds the required --market <uint32> flag to a command.
func AddFlagMarket(cmd *cobra.Command) {
	cmd.Flags().Uint32(FlagMarket, 0, "The market id, e.g. 3")
	markFlagRequired(cmd, FlagMarket)
}

// ReadFlagMarket reads the --market flag.
func ReadFlagMarket(flagSet *pflag.FlagSet) (uint32, error) {
	return flagSet.GetUint32(FlagMarket)
}

// AddFlagOrder adds the optional --order <uint64> flag to a command.
func AddFlagOrder(cmd *cobra.Command) {
	cmd.Flags().Uint32(FlagOrder, 0, "The market id, e.g. 3")
}

// ReadFlagOrder reads the --order flag.
func ReadFlagOrder(flagSet *pflag.FlagSet) (uint64, error) {
	return flagSet.GetUint64(FlagOrder)
}

// AddFlagAllowPartial adds the optional --partial flag to a command to indicate partial fulfillment is allowed.
func AddFlagAllowPartial(cmd *cobra.Command) {
	cmd.Flags().Bool(FlagPartial, false, "Allow this order to be partially filled")
}

// AddFlagExpectPartial adds the optional --partial flag to a command to indicate partial fulfillment is expected.
func AddFlagExpectPartial(cmd *cobra.Command) {
	cmd.Flags().Bool(FlagPartial, false, "Expect partial settlement")
}

// ReadFlagPartial reads the --partial flag.
func ReadFlagPartial(flagSet *pflag.FlagSet) (bool, error) {
	return flagSet.GetBool(FlagPartial)
}

// AddFlagPrice adds the required --price <string> flag to a command for providing an order's price.
func AddFlagPrice(cmd *cobra.Command) {
	cmd.Flags().String(FlagPrice, "", "The price for this order, e.g. 10nhash")
	markFlagRequired(cmd, FlagPrice)
}

// AddFlagTotalPrice adds the required --price <string> flag to a command for providing a total price.
func AddFlagTotalPrice(cmd *cobra.Command) {
	cmd.Flags().String(FlagPrice, "", "The total price you are paying, e.g. 10nhash")
	markFlagRequired(cmd, FlagPrice)
}

// ReadFlagPrice reads the --price flag as sdk.Coin.
func ReadFlagPrice(flagSet *pflag.FlagSet) (sdk.Coin, error) {
	return readReqCoinFlag(flagSet, FlagPrice)
}

// AddFlagSeller adds the optional --seller flag to a command.
func AddFlagSeller(cmd *cobra.Command) {
	cmd.Flags().String(FlagSeller, "", "The seller (defaults to --from account)")
}

// ReadFlagSellerOrDefault reads the --seller flag if provided, or returns the --from address.
// Returns an error if neither of those flags were provided, or there was an error reading one.
func ReadFlagSellerOrDefault(clientCtx client.Context, flagSet *pflag.FlagSet) (string, error) {
	return readAddrOrDefault(clientCtx, flagSet, FlagSeller)
}

// AddFlagSettlementFee adds the optional --settlement-fee <string> flag to a command.
func AddFlagSettlementFee(cmd *cobra.Command) {
	cmd.Flags().String(FlagSettlementFee, "", "The settlement fee Coin string for this order, e.g. 10nhash")
}

// ReadFlagSettlementFeeCoins reads the --settlement-fee flag as sdk.Coins.
func ReadFlagSettlementFeeCoins(flagSet *pflag.FlagSet) (sdk.Coins, error) {
	return readCoinsFlag(flagSet, FlagSettlementFee)
}

// ReadFlagSettlementFeeCoin reads the --settlement-fee flag as sdk.Coin.
func ReadFlagSettlementFeeCoin(flagSet *pflag.FlagSet) (*sdk.Coin, error) {
	return readCoinFlag(flagSet, FlagSettlementFee)
}

// AddFlagSigner adds the optional --signer flag to a command.
func AddFlagSigner(cmd *cobra.Command) {
	cmd.Flags().String(FlagSigner, "", "The signer (defaults to --from account)")
}

// ReadFlagSignerOrDefault reads the --signer flag if provided, or returns the --from address.
// Returns an error if neither of those flags were provided or there was an error reading one.
func ReadFlagSignerOrDefault(clientCtx client.Context, flagSet *pflag.FlagSet) (string, error) {
	return readAddrOrDefault(clientCtx, flagSet, FlagSigner)
}
