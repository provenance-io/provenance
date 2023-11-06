package cli

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/provenance-io/provenance/x/exchange"
)

const (
	FlagAdmin         = "admin"
	FlagAmount        = "amount"
	FlagAskAdd        = "ask-add"
	FlagAskRemove     = "ask-remove"
	FlagAsks          = "asks"
	FlagAssets        = "assets"
	FlagAuthority     = "authority"
	FlagBidAdd        = "bid-add"
	FlagBidRemove     = "bid-remove"
	FlagBids          = "bids"
	FlagBuyer         = "buyer"
	FlagCreationFee   = "creation-fee"
	FlagDescription   = "description"
	FlagDisable       = "disable"
	FlagEnable        = "enable"
	FlagExternalID    = "external-id"
	FlagGrant         = "grant"
	FlagIcon          = "icon"
	FlagMarket        = "market"
	FlagName          = "name"
	FlagOrder         = "order"
	FlagPartial       = "partial"
	FlagPrice         = "price"
	FlagRevokeAll     = "revoke-all"
	FlagRevoke        = "revoke"
	FlagSeller        = "seller"
	FlagSettlementFee = "settlement-fee"
	FlagSigner        = "signer"
	FlagTo            = "to"
	FlagURL           = "url"
)

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

// readAccessGrants reads a StringSlice flag and converts it to a slice of AccessGrants.
func readAccessGrants(flagSet *pflag.FlagSet, name string) ([]exchange.AccessGrant, error) {
	vals, err := flagSet.GetStringSlice(name)
	if err != nil || len(vals) == 0 {
		return nil, err
	}

	return ParseAccessGrants(vals...)
}

// permSepRx is a regexp that matches characters that can be used to separate permissions.
var permSepRx = regexp.MustCompile(`[ +-.]`)

// ParseAccessGrant parses an AccessGrant from a string with the format "<address>:<perm 1>[+<perm 2>...]".
func ParseAccessGrant(val string) (*exchange.AccessGrant, error) {
	parts := strings.Split(val, ":")
	if len(parts) <= 1 {
		return nil, fmt.Errorf("could not parse %q as an AccessGrant: expected format <address>:<permissions>", val)
	}

	var permissions []exchange.Permission
	if strings.ToLower(strings.TrimSpace(parts[1])) == "all" {
		permissions = exchange.AllPermissions()
	} else {
		permVals := permSepRx.Split(parts[1], 0)
		var err error
		permissions, err = exchange.ParsePermissions(permVals...)
		if err != nil {
			return nil, fmt.Errorf("could not parse %q permissions from %q: %w", parts[0], parts[1], err)
		}
	}

	rv := &exchange.AccessGrant{
		Address:     parts[0],
		Permissions: permissions,
	}
	return rv, nil
}

// ParseAccessGrants parses an AccessGrant from each of the provided vals.
func ParseAccessGrants(vals ...string) ([]exchange.AccessGrant, error) {
	var errs []error
	rv := make([]exchange.AccessGrant, len(vals))
	for i, val := range vals {
		ag, err := ParseAccessGrant(val)
		if err != nil {
			errs = append(errs, err)
		}
		if ag != nil {
			rv[i] = *ag
		}
	}
	return rv, errors.Join(errs...)
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

// AddFlagAmount adds the required --amount <string> flag to a command.
func AddFlagAmount(cmd *cobra.Command) {
	cmd.Flags().String(FlagAmount, "", "The amount to withdraw")
	markFlagRequired(cmd, FlagAmount)
}

// ReadFlagAmount reads the --amount flag.
func ReadFlagAmount(flagSet *pflag.FlagSet) (sdk.Coins, error) {
	return readCoinsFlag(flagSet, FlagAmount)
}

// AddFlagAskAdd adds the optional --ask-add <strings> flag to a command.
func AddFlagAskAdd(cmd *cobra.Command) {
	cmd.Flags().StringSlice(FlagAskAdd, nil, "The create-ask required attributes to add")
}

// ReadFlagAskAdd reads the --ask-add flag.
func ReadFlagAskAdd(flagSet *pflag.FlagSet) ([]string, error) {
	return flagSet.GetStringSlice(FlagAskAdd)
}

// AddFlagAskRemove adds the optional --ask-remove <strings> flag to a command.
func AddFlagAskRemove(cmd *cobra.Command) {
	cmd.Flags().StringSlice(FlagAskRemove, nil, "The create-ask required attributes to remove")
}

// ReadFlagAskRemove reads the --ask-remove flag.
func ReadFlagAskRemove(flagSet *pflag.FlagSet) ([]string, error) {
	return flagSet.GetStringSlice(FlagAskRemove)
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

// AddFlagBidAdd adds the optional --bid-add <strings> flag to a command.
func AddFlagBidAdd(cmd *cobra.Command) {
	cmd.Flags().StringSlice(FlagBidAdd, nil, "The create-bid required attributes to add")
}

// ReadFlagBidAdd reads the --bid-add flag.
func ReadFlagBidAdd(flagSet *pflag.FlagSet) ([]string, error) {
	return flagSet.GetStringSlice(FlagBidAdd)
}

// AddFlagBidRemove adds the optional --bid-remove <strings> flag to a command.
func AddFlagBidRemove(cmd *cobra.Command) {
	cmd.Flags().StringSlice(FlagBidRemove, nil, "The create-bid required attributes to remove")
}

// ReadFlagBidRemove reads the --bid-remove flag.
func ReadFlagBidRemove(flagSet *pflag.FlagSet) ([]string, error) {
	return flagSet.GetStringSlice(FlagBidRemove)
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

// AddFlagDescription adds the optional --description <string> flag to a command.
func AddFlagDescription(cmd *cobra.Command) {
	cmd.Flags().String(FlagDescription, "", "A description of the market")
}

// ReadFlagDescription reads the --description flag.
func ReadFlagDescription(flagSet *pflag.FlagSet) (string, error) {
	return flagSet.GetString(FlagDescription)
}

// AddFlagDisable adds the optional --disable flag to a command.
func AddFlagDisable(cmd *cobra.Command, name string) {
	cmd.Flags().Bool(FlagDisable, false, fmt.Sprintf("Set the market's %s field to false", name))
}

// ReadFlagDisable reads the --disable flag.
func ReadFlagDisable(flagSet *pflag.FlagSet) (bool, error) {
	return flagSet.GetBool(FlagDisable)
}

// AddFlagEnable adds the optional --enable flag to a command.
func AddFlagEnable(cmd *cobra.Command, name string) {
	cmd.Flags().Bool(FlagEnable, false, fmt.Sprintf("Set the market's %s field to true", name))
}

// ReadFlagEnable reads the --enable flag.
func ReadFlagEnable(flagSet *pflag.FlagSet) (bool, error) {
	return flagSet.GetBool(FlagEnable)
}

// ReadFlagExternalID reads the --external-id flag.
func ReadFlagExternalID(flagSet *pflag.FlagSet) (string, error) {
	return flagSet.GetString(FlagExternalID)
}

// AddFlagGrant adds the optional --grant <access grants> flag to a command.
func AddFlagGrant(cmd *cobra.Command) {
	cmd.Flags().StringSlice(FlagGrant, nil, "AccessGrants to add to the market")
}

// ReadFlagGrant reads the --grant flag.
func ReadFlagGrant(flagSet *pflag.FlagSet) ([]exchange.AccessGrant, error) {
	return readAccessGrants(flagSet, FlagGrant)
}

// AddFlagIcon adds the optional --icon <string> flag to a command.
func AddFlagIcon(cmd *cobra.Command) {
	cmd.Flags().String(FlagIcon, "", "The market's icon URI")
}

// ReadFlagIcon reads the --icon flag.
func ReadFlagIcon(flagSet *pflag.FlagSet) (string, error) {
	return flagSet.GetString(FlagIcon)
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

// AddFlagName adds the optional --name <string> flag to a command.
func AddFlagName(cmd *cobra.Command) {
	cmd.Flags().String(FlagName, "", "A short name for the market")
}

// ReadFlagName reads the --name flag.
func ReadFlagName(flagSet *pflag.FlagSet) (string, error) {
	return flagSet.GetString(FlagName)
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

// AddFlagRevokeAll adds the optional --revoke-all <addresses> flag to a command.
func AddFlagRevokeAll(cmd *cobra.Command) {
	cmd.Flags().StringSlice(FlagRevokeAll, nil, "Addresses to revoke all permissions from")
}

// ReadFlagRevokeAll reads the --revoke-all flag.
func ReadFlagRevokeAll(flagSet *pflag.FlagSet) ([]string, error) {
	return flagSet.GetStringSlice(FlagRevokeAll)
}

// AddFlagRevoke adds the optional --revoke <access grants> flag to a command.
func AddFlagRevoke(cmd *cobra.Command) {
	cmd.Flags().StringSlice(FlagRevoke, nil, "AccessGrants to remove from the market")
}

// ReadFlagRevoke reads the --revoke flag.
func ReadFlagRevoke(flagSet *pflag.FlagSet) ([]exchange.AccessGrant, error) {
	return readAccessGrants(flagSet, FlagRevoke)
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

// AddFlagTo adds the required --to <string> flag to a command.
func AddFlagTo(cmd *cobra.Command) {
	cmd.Flags().String(FlagTo, "", "The address that will receive the funds")
	markFlagRequired(cmd, FlagTo)
}

// ReadFlagTo reads the --to flag.
func ReadFlagTo(flagSet *pflag.FlagSet) (string, error) {
	return flagSet.GetString(FlagTo)
}

// AddFlagURL adds the optional --url <string> flag to a command.
func AddFlagURL(cmd *cobra.Command) {
	cmd.Flags().String(FlagURL, "", "The market's website URL")
}

// ReadFlagURL reads the --url flag.
func ReadFlagURL(flagSet *pflag.FlagSet) (string, error) {
	return flagSet.GetString(FlagURL)
}
