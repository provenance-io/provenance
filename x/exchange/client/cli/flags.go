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
	"github.com/provenance-io/provenance/x/exchange"
)

const (
	FlagAcceptingOrders = "accepting-orders"
	FlagAccessGrants    = "access-grants"
	FlagAdmin           = "admin"
	FlagAllowUserSettle = "allow-user-settle"
	FlagAmount          = "amount"
	FlagAskAdd          = "ask-add"
	FlagAskRemove       = "ask-remove"
	FlagAsks            = "asks"
	FlagAssets          = "assets"
	FlagAuthority       = "authority"
	FlagBidAdd          = "bid-add"
	FlagBidRemove       = "bid-remove"
	FlagBids            = "bids"
	FlagBuyer           = "buyer"
	FlagBuyerFlat       = "buyer-flat"
	FlagBuyerRatios     = "buyer-ratios"
	FlagCreateAsk       = "create-ask"
	FlagCreateBid       = "create-bid"
	FlagCreationFee     = "creation-fee"
	FlagDescription     = "description"
	FlagDisable         = "disable"
	FlagEnable          = "enable"
	FlagExternalID      = "external-id"
	FlagGrant           = "grant"
	FlagIcon            = "icon"
	FlagMarket          = "market"
	FlagName            = "name"
	FlagOrder           = "order"
	FlagPartial         = "partial"
	FlagPrice           = "price"
	FlagReqAttrAsk      = "req-attr-ask"
	FlagReqAttrBid      = "req-attr-Bid"
	FlagRevokeAll       = "revoke-all"
	FlagRevoke          = "revoke"
	FlagSeller          = "seller"
	FlagSellerFlat      = "seller-flat"
	FlagSellerRatios    = "seller-ratios"
	FlagSettlementFee   = "settlement-fee"
	FlagSigner          = "signer"
	FlagTo              = "to"
	FlagURL             = "url"
)

// MarkFlagRequired this marks a flag as required and panics if there's a problem.
func MarkFlagRequired(cmd *cobra.Command, name string) {
	if err := cmd.MarkFlagRequired(name); err != nil {
		panic(fmt.Errorf("error marking --%s flag required on %s: %w", name, cmd.Name(), err))
	}
}

// ReadCoinsFlag reads a string flag and converts it into sdk.Coins.
// If the flag wasn't provided, this returns nil, nil.
func ReadCoinsFlag(flagSet *pflag.FlagSet, name string) (sdk.Coins, error) {
	value, err := flagSet.GetString(name)
	if len(value) == 0 || err != nil {
		return nil, err
	}
	rv, err := sdk.ParseCoinsNormalized(value)
	if err != nil {
		return nil, fmt.Errorf("error parsing --%s value %q as coins: %w", name, value, err)
	}
	return rv, nil
}

// ReadCoinFlag reads a string flag and converts it into *sdk.Coin.
// If the flag wasn't provided, this returns nil, nil.
func ReadCoinFlag(flagSet *pflag.FlagSet, name string) (*sdk.Coin, error) {
	value, err := flagSet.GetString(name)
	if len(value) == 0 || err != nil {
		return nil, err
	}
	rv, err := exchange.ParseCoin(value)
	if err != nil {
		return nil, fmt.Errorf("error parsing --%s value %q as a coin: %w", name, value, err)
	}
	return &rv, nil
}

// ReadReqCoinFlag reads a string flag and converts it into a sdk.Coin and requires it to have a value.
func ReadReqCoinFlag(flagSet *pflag.FlagSet, name string) (sdk.Coin, error) {
	rv, err := ReadCoinFlag(flagSet, name)
	if err != nil {
		return sdk.Coin{}, err
	}
	if rv == nil {
		return sdk.Coin{}, fmt.Errorf("missing required --%s flag", name)
	}
	return *rv, nil
}

// ReadOrderIdsFlag reads a UintSlice flag and converts it into a []uint64.
func ReadOrderIdsFlag(flagSet *pflag.FlagSet, name string) ([]uint64, error) {
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

// ReadAddrOrDefault gets the requested flag or, if it wasn't provided, gets the --from address.
func ReadAddrOrDefault(clientCtx client.Context, flagSet *pflag.FlagSet, name string) (string, error) {
	rv, err := flagSet.GetString(name)
	if len(rv) > 0 || err != nil {
		return rv, err
	}

	rv = clientCtx.FromAddress.String()
	if len(rv) > 0 {
		return rv, nil
	}

	return "", fmt.Errorf("no %s provided", name)
}

// ReadAccessGrants reads a StringSlice flag and converts it to a slice of AccessGrants.
func ReadAccessGrants(flagSet *pflag.FlagSet, name string) ([]exchange.AccessGrant, error) {
	vals, err := flagSet.GetStringSlice(name)
	if len(vals) == 0 || err != nil {
		return nil, err
	}
	return ParseAccessGrants(vals)
}

// ReadFlatFee reads a StringSlice flag and converts it into a slice of sdk.Coin.
func ReadFlatFee(flagSet *pflag.FlagSet, name string) ([]sdk.Coin, error) {
	vals, err := flagSet.GetStringSlice(name)
	if len(vals) == 0 || err != nil {
		return nil, err
	}
	return ParseFlatFeeOptions(vals)
}

// ReadFeeRatios reads a StringSlice flag and converts it into a slice of exchange.FeeRatio.
func ReadFeeRatios(flagSet *pflag.FlagSet, name string) ([]exchange.FeeRatio, error) {
	vals, err := flagSet.GetStringSlice(name)
	if len(vals) == 0 || err != nil {
		return nil, err
	}
	return ParseFeeRatios(vals)
}

// permSepRx is a regexp that matches characters that can be used to separate permissions.
var permSepRx = regexp.MustCompile(`[ +-.]`)

// ParseAccessGrant parses an AccessGrant from a string with the format "<address>:<perm 1>[+<perm 2>...]".
func ParseAccessGrant(val string) (*exchange.AccessGrant, error) {
	parts := strings.Split(val, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("could not parse %q as an <access grant>: expected format <address>:<permissions>", val)
	}

	addr := strings.TrimSpace(parts[0])
	perms := strings.ToLower(strings.TrimSpace(parts[1]))
	if len(addr) == 0 || len(perms) == 0 {
		return nil, fmt.Errorf("invalid <access grant>: both an <address> and <permissions> are required")
	}

	rv := &exchange.AccessGrant{Address: addr}

	if perms == "all" {
		rv.Permissions = exchange.AllPermissions()
		return rv, nil
	}

	permVals := permSepRx.Split(perms, -1)
	var err error
	rv.Permissions, err = exchange.ParsePermissions(permVals...)
	if err != nil {
		return nil, fmt.Errorf("could not parse permissions for %q from %q: %w", rv.Address, parts[1], err)
	}

	return rv, nil
}

// ParseAccessGrants parses an AccessGrant from each of the provided vals.
func ParseAccessGrants(vals []string) ([]exchange.AccessGrant, error) {
	var errs []error
	rv := make([]exchange.AccessGrant, 0, len(vals))
	for _, val := range vals {
		ag, err := ParseAccessGrant(val)
		if err != nil {
			errs = append(errs, err)
		}
		if ag != nil {
			rv = append(rv, *ag)
		}
	}
	return rv, errors.Join(errs...)
}

// ParseFlatFeeOptions parses an sdk.Coin from each of the provided vals.
func ParseFlatFeeOptions(vals []string) ([]sdk.Coin, error) {
	var errs []error
	rv := make([]sdk.Coin, 0, len(vals))
	for _, val := range vals {
		coin, err := exchange.ParseCoin(val)
		if err != nil {
			errs = append(errs, err)
		} else {
			rv = append(rv, coin)
		}
	}
	return rv, errors.Join(errs...)
}

// ParseFeeRatios parses a FeeRatio from each of the provided vals.
func ParseFeeRatios(vals []string) ([]exchange.FeeRatio, error) {
	var errs []error
	rv := make([]exchange.FeeRatio, 0, len(vals))
	for _, val := range vals {
		ratio, err := exchange.ParseFeeRatio(val)
		if err != nil {
			errs = append(errs, err)
		}
		if ratio != nil {
			rv = append(rv, *ratio)
		}
	}
	return rv, errors.Join(errs...)
}

// AddFlagAcceptingOrders adds the optional --accepting-orders flag to a command.
func AddFlagAcceptingOrders(cmd *cobra.Command) {
	cmd.Flags().Bool(FlagAcceptingOrders, false, "The market should allow orders to be created")
}

// ReadFlagAcceptingOrders reads the --accepting-orders flag.
func ReadFlagAcceptingOrders(flagSet *pflag.FlagSet) (bool, error) {
	return flagSet.GetBool(FlagAcceptingOrders)
}

// AddFlagAccessGrants adds the optional --access-grants <strings> flag to a command.
func AddFlagAccessGrants(cmd *cobra.Command) {
	cmd.Flags().StringSlice(FlagAccessGrants, nil, "The access grants that the market should have")
}

// ReadFlagAccessGrants reads the --access-grants flag.
func ReadFlagAccessGrants(flagSet *pflag.FlagSet) ([]exchange.AccessGrant, error) {
	return ReadAccessGrants(flagSet, FlagAccessGrants)
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
		return AuthorityAddr.String(), nil
	}

	rv = clientCtx.FromAddress.String()
	if len(rv) > 0 {
		return rv, nil
	}

	return "", fmt.Errorf("no %s provided", FlagAdmin)
}

// AddFlagAllowUserSettle adds the optional --allow-user-settle flag to a command.
func AddFlagAllowUserSettle(cmd *cobra.Command) {
	cmd.Flags().Bool(FlagAllowUserSettle, false, "The market should allow user-initiated settlement")
}

// ReadFlagAllowUserSettle reads the --allow-user-settle flag.
func ReadFlagAllowUserSettle(flagSet *pflag.FlagSet) (bool, error) {
	return flagSet.GetBool(FlagAllowUserSettle)
}

// AddFlagAmount adds the required --amount <string> flag to a command.
func AddFlagAmount(cmd *cobra.Command) {
	cmd.Flags().String(FlagAmount, "", "The amount to withdraw")
	MarkFlagRequired(cmd, FlagAmount)
}

// ReadFlagAmount reads the --amount flag.
func ReadFlagAmount(flagSet *pflag.FlagSet) (sdk.Coins, error) {
	return ReadCoinsFlag(flagSet, FlagAmount)
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
	MarkFlagRequired(cmd, FlagAsks)
}

// ReadFlagAsks reads the --asks flag.
func ReadFlagAsks(flagSet *pflag.FlagSet) ([]uint64, error) {
	return ReadOrderIdsFlag(flagSet, FlagAsks)
}

// AddFlagAssets adds the required --assets <string> flag to a command for providing an order's assets.
func AddFlagAssets(cmd *cobra.Command) {
	cmd.Flags().String(FlagAssets, "", "The assets for this order, e.g. 10nhash")
	MarkFlagRequired(cmd, FlagAssets)
}

// ReadFlagAssets reads the --assets flag as sdk.Coin.
func ReadFlagAssets(flagSet *pflag.FlagSet) (sdk.Coin, error) {
	return ReadReqCoinFlag(flagSet, FlagAssets)
}

// AddFlagTotalAssets adds the required --assets <string> flag to a command for providing total assets.
func AddFlagTotalAssets(cmd *cobra.Command) {
	cmd.Flags().String(FlagAssets, "", "The total assets you are filling, e.g. 10nhash")
	MarkFlagRequired(cmd, FlagAssets)
}

// ReadFlagTotalAssets reads the --assets flag as sdk.Coins.
func ReadFlagTotalAssets(flagSet *pflag.FlagSet) (sdk.Coins, error) {
	return ReadCoinsFlag(flagSet, FlagAssets)
}

// AddFlagAuthorityString adds the optional --authority <string> flag.
func AddFlagAuthorityString(cmd *cobra.Command) {
	cmd.Flags().String(FlagAuthority, "", "The authority address to use (defaults to the governance module account)")
}

// ReadFlagAuthorityString reads the --authority flag, or if not provided, returns the standard authority address.
func ReadFlagAuthorityString(flagSet *pflag.FlagSet) (string, error) {
	rv, err := flagSet.GetString(FlagAuthority)
	if len(rv) > 0 || err != nil {
		return rv, err
	}
	return AuthorityAddr.String(), nil
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
	MarkFlagRequired(cmd, FlagBids)
}

// ReadFlagBids reads the --bids flag.
func ReadFlagBids(flagSet *pflag.FlagSet) ([]uint64, error) {
	return ReadOrderIdsFlag(flagSet, FlagBids)
}

// AddFlagBuyer adds the optional --buyer flag to a command.
func AddFlagBuyer(cmd *cobra.Command) {
	cmd.Flags().String(FlagBuyer, "", "The buyer (defaults to --from account)")
}

// ReadFlagBuyerOrDefault reads the --buyer flag if provided, or returns the --from address.
// Returns an error if neither of those flags were provided, or there was an error reading one.
func ReadFlagBuyerOrDefault(clientCtx client.Context, flagSet *pflag.FlagSet) (string, error) {
	return ReadAddrOrDefault(clientCtx, flagSet, FlagBuyer)
}

// AddFlagBuyerFlat adds the optional --buyer-flat <strings> flag to a command.
func AddFlagBuyerFlat(cmd *cobra.Command) {
	cmd.Flags().StringSlice(FlagBuyerFlat, nil, "The buyer settlement flat fee options, e.g. 10nhash")
}

// ReadFlagBuyerFlat reads the --buyer-flat flag.
func ReadFlagBuyerFlat(flagSet *pflag.FlagSet) ([]sdk.Coin, error) {
	return ReadFlatFee(flagSet, FlagBuyerFlat)
}

// AddFlagBuyerRatios adds the optional --buyer-ratios <strings> flag to a command.
func AddFlagBuyerRatios(cmd *cobra.Command) {
	cmd.Flags().StringSlice(FlagBuyerRatios, nil, "The buyer settlement fee ratios, e.g. 100nhash:1nhash")
}

// ReadFlagBuyerRatios reads the --buyer-ratios flag.
func ReadFlagBuyerRatios(flagSet *pflag.FlagSet) ([]exchange.FeeRatio, error) {
	return ReadFeeRatios(flagSet, FlagBuyerRatios)
}

// AddFlagCreateAsk adds the optional --create-ask <strings> flag to a command.
func AddFlagCreateAsk(cmd *cobra.Command) {
	cmd.Flags().StringSlice(FlagCreateAsk, nil, "The create-ask fee options, e.g. 10nhash")
}

// ReadFlagCreateAsk reads the --create-ask flag.
func ReadFlagCreateAsk(flagSet *pflag.FlagSet) ([]sdk.Coin, error) {
	return ReadFlatFee(flagSet, FlagCreateAsk)
}

// AddFlagCreateBid adds the optional --create-bid <strings> flag to a command.
func AddFlagCreateBid(cmd *cobra.Command) {
	cmd.Flags().StringSlice(FlagCreateBid, nil, "The create-bid fee options, e.g. 10nhash")
}

// ReadFlagCreateBid reads the --create-bid flag.
func ReadFlagCreateBid(flagSet *pflag.FlagSet) ([]sdk.Coin, error) {
	return ReadFlatFee(flagSet, FlagCreateBid)
}

// AddFlagCreationFee adds the optional --creation-fee <string> flag to a command.
func AddFlagCreationFee(cmd *cobra.Command) {
	cmd.Flags().String(FlagCreationFee, "", "The order creation fee, e.g. 10nhash")
}

// ReadFlagCreationFee reads the --creation-fee flag as sdk.Coin.
func ReadFlagCreationFee(flagSet *pflag.FlagSet) (*sdk.Coin, error) {
	return ReadCoinFlag(flagSet, FlagCreationFee)
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
	return ReadAccessGrants(flagSet, FlagGrant)
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
	MarkFlagRequired(cmd, FlagMarket)
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
	MarkFlagRequired(cmd, FlagPrice)
}

// AddFlagTotalPrice adds the required --price <string> flag to a command for providing a total price.
func AddFlagTotalPrice(cmd *cobra.Command) {
	cmd.Flags().String(FlagPrice, "", "The total price you are paying, e.g. 10nhash")
	MarkFlagRequired(cmd, FlagPrice)
}

// ReadFlagPrice reads the --price flag as sdk.Coin.
func ReadFlagPrice(flagSet *pflag.FlagSet) (sdk.Coin, error) {
	return ReadReqCoinFlag(flagSet, FlagPrice)
}

// AddFlagReqAttrAsk adds the optional --req-attr-ask <strings> flag to a command.
func AddFlagReqAttrAsk(cmd *cobra.Command) {
	cmd.Flags().StringSlice(FlagReqAttrAsk, nil, "Attributes required to create ask orders")
}

// ReadFlagReqAttrAsk reads the --req-attr-ask flag.
func ReadFlagReqAttrAsk(flagSet *pflag.FlagSet) ([]string, error) {
	return flagSet.GetStringSlice(FlagReqAttrAsk)
}

// AddFlagReqAttrBid adds the optional --req-attr-bid <strings> flag to a command.
func AddFlagReqAttrBid(cmd *cobra.Command) {
	cmd.Flags().StringSlice(FlagReqAttrBid, nil, "Attributes required to create bid orders")
}

// ReadFlagReqAttrBid reads the --req-attr-bid flag.
func ReadFlagReqAttrBid(flagSet *pflag.FlagSet) ([]string, error) {
	return flagSet.GetStringSlice(FlagReqAttrBid)
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
	return ReadAccessGrants(flagSet, FlagRevoke)
}

// AddFlagSeller adds the optional --seller flag to a command.
func AddFlagSeller(cmd *cobra.Command) {
	cmd.Flags().String(FlagSeller, "", "The seller (defaults to --from account)")
}

// ReadFlagSellerOrDefault reads the --seller flag if provided, or returns the --from address.
// Returns an error if neither of those flags were provided, or there was an error reading one.
func ReadFlagSellerOrDefault(clientCtx client.Context, flagSet *pflag.FlagSet) (string, error) {
	return ReadAddrOrDefault(clientCtx, flagSet, FlagSeller)
}

// AddFlagSellerFlat adds the optional --seller-flat <strings> flag to a command.
func AddFlagSellerFlat(cmd *cobra.Command) {
	cmd.Flags().StringSlice(FlagSellerFlat, nil, "The seller settlement flat fee options, e.g. 10nhash")
}

// ReadFlagSellerFlat reads the --seller-flat flag.
func ReadFlagSellerFlat(flagSet *pflag.FlagSet) ([]sdk.Coin, error) {
	return ReadFlatFee(flagSet, FlagSellerFlat)
}

// AddFlagSellerRatios adds the optional --seller-ratios <strings> flag to a command.
func AddFlagSellerRatios(cmd *cobra.Command) {
	cmd.Flags().StringSlice(FlagSellerRatios, nil, "The seller settlement fee ratios, e.g. 100nhash:1nhash")
}

// ReadFlagSellerRatios reads the --seller-ratios flag.
func ReadFlagSellerRatios(flagSet *pflag.FlagSet) ([]exchange.FeeRatio, error) {
	return ReadFeeRatios(flagSet, FlagSellerRatios)
}

// AddFlagSettlementFee adds the optional --settlement-fee <string> flag to a command.
func AddFlagSettlementFee(cmd *cobra.Command) {
	cmd.Flags().String(FlagSettlementFee, "", "The settlement fee Coin string for this order, e.g. 10nhash")
}

// ReadFlagSettlementFeeCoins reads the --settlement-fee flag as sdk.Coins.
func ReadFlagSettlementFeeCoins(flagSet *pflag.FlagSet) (sdk.Coins, error) {
	return ReadCoinsFlag(flagSet, FlagSettlementFee)
}

// ReadFlagSettlementFeeCoin reads the --settlement-fee flag as sdk.Coin.
func ReadFlagSettlementFeeCoin(flagSet *pflag.FlagSet) (*sdk.Coin, error) {
	return ReadCoinFlag(flagSet, FlagSettlementFee)
}

// AddFlagSigner adds the optional --signer flag to a command.
func AddFlagSigner(cmd *cobra.Command) {
	cmd.Flags().String(FlagSigner, "", "The signer (defaults to --from account)")
}

// ReadFlagSignerOrDefault reads the --signer flag if provided, or returns the --from address.
// Returns an error if neither of those flags were provided or there was an error reading one.
func ReadFlagSignerOrDefault(clientCtx client.Context, flagSet *pflag.FlagSet) (string, error) {
	return ReadAddrOrDefault(clientCtx, flagSet, FlagSigner)
}

// AddFlagTo adds the required --to <string> flag to a command.
func AddFlagTo(cmd *cobra.Command) {
	cmd.Flags().String(FlagTo, "", "The address that will receive the funds")
	MarkFlagRequired(cmd, FlagTo)
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
