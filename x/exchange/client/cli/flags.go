package cli

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	"github.com/provenance-io/provenance/x/exchange"
)

const (
	FlagAcceptingOrders    = "accepting-orders"
	FlagAccessGrants       = "access-grants"
	FlagAdmin              = "admin"
	FlagAfter              = "after"
	FlagAllowUserSettle    = "allow-user-settle"
	FlagAmount             = "amount"
	FlagAsk                = "ask"
	FlagAskAdd             = "ask-add"
	FlagAskRemove          = "ask-remove"
	FlagAsks               = "asks"
	FlagAssets             = "assets"
	FlagAuthority          = "authority"
	FlagBid                = "bid"
	FlagBidAdd             = "bid-add"
	FlagBidRemove          = "bid-remove"
	FlagBids               = "bids"
	FlagBuyer              = "buyer"
	FlagBuyerFlat          = "buyer-flat"
	FlagBuyerFlatAdd       = "buyer-flat-add"
	FlagBuyerFlatRemove    = "buyer-flat-remove"
	FlagBuyerRatios        = "buyer-ratios"
	FlagBuyerRatiosAdd     = "buyer-ratios-add"
	FlagBuyerRatiosRemove  = "buyer-ratios-remove"
	FlagCreateAsk          = "create-ask"
	FlagCreateBid          = "create-bid"
	FlagCreationFee        = "creation-fee"
	FlagDefault            = "default"
	FlagDenom              = "denom"
	FlagDescription        = "description"
	FlagDisable            = "disable"
	FlagEnable             = "enable"
	FlagExternalID         = "external-id"
	FlagGrant              = "grant"
	FlagIcon               = "icon"
	FlagMarket             = "market"
	FlagName               = "name"
	FlagOrder              = "order"
	FlagOwner              = "owner"
	FlagPartial            = "partial"
	FlagPrice              = "price"
	FlagProposal           = "proposal"
	FlagReqAttrAsk         = "req-attr-ask"
	FlagReqAttrBid         = "req-attr-bid"
	FlagRevoke             = "revoke"
	FlagRevokeAll          = "revoke-all"
	FlagSeller             = "seller"
	FlagSellerFlat         = "seller-flat"
	FlagSellerFlatAdd      = "seller-flat-add"
	FlagSellerFlatRemove   = "seller-flat-remove"
	FlagSellerRatios       = "seller-ratios"
	FlagSellerRatiosAdd    = "seller-ratios-add"
	FlagSellerRatiosRemove = "seller-ratios-remove"
	FlagSettlementFee      = "settlement-fee"
	FlagSigner             = "signer"
	FlagSplit              = "split"
	FlagTo                 = "to"
	FlagURL                = "url"
)

// MarkFlagsRequired marks the provided flags as required and panics if there's a problem.
func MarkFlagsRequired(cmd *cobra.Command, names ...string) {
	for _, name := range names {
		if err := cmd.MarkFlagRequired(name); err != nil {
			panic(fmt.Errorf("error marking --%s flag required on %s: %w", name, cmd.Name(), err))
		}
	}
}

// AddFlagsAdmin adds the --admin and --authority flags to a command and makes them mutually exclusive.
// It also makes one of --admin, --authority, and --from required.
//
// Use ReadFlagsAdminOrFrom to read these flags.
func AddFlagsAdmin(cmd *cobra.Command) {
	cmd.Flags().String(FlagAdmin, "", "The admin (defaults to --from account)")
	cmd.Flags().Bool(FlagAuthority, false, "Use the governance module account for the admin")

	cmd.MarkFlagsMutuallyExclusive(FlagAdmin, FlagAuthority)
	cmd.MarkFlagsOneRequired(flags.FlagFrom, FlagAdmin, FlagAuthority)
}

// ReadFlagsAdminOrFrom reads the --admin flag if provided.
// If not, but the --authority flag was provided, the gov module account address is returned.
// If no -admin or --authority flag was provided, returns the --from address.
// Returns an error if none of those flags were provided or there was an error reading one.
//
// This assumes AddFlagsAdmin was used to define the flags, and that the context comes from client.GetClientTxContext.
func ReadFlagsAdminOrFrom(clientCtx client.Context, flagSet *pflag.FlagSet) (string, error) {
	rv, err := flagSet.GetString(FlagAdmin)
	if len(rv) > 0 || err != nil {
		return rv, err
	}

	useAuth, err := flagSet.GetBool(FlagAuthority)
	if err != nil {
		return "", err
	}
	if useAuth {
		return AuthorityAddr.String(), nil
	}

	rv = clientCtx.GetFromAddress().String()
	if len(rv) > 0 {
		return rv, nil
	}

	return "", errors.New("no <admin> provided")
}

// ReadFlagAuthority reads the --authority flag, or if not provided, returns the standard authority address.
// This assumes that the flag was defined with a default of "".
func ReadFlagAuthority(flagSet *pflag.FlagSet) (string, error) {
	return ReadFlagAuthorityOrDefault(flagSet, AuthorityAddr.String())
}

// ReadFlagAuthorityOrDefault reads the --authority flag, or if not provided, returns the default.
// If the provided default is "", the standard authority address is used as the default.
// This assumes that the flag was defined with a default of "".
func ReadFlagAuthorityOrDefault(flagSet *pflag.FlagSet, def string) (string, error) {
	rv, err := flagSet.GetString(FlagAuthority)
	if len(rv) == 0 || err != nil {
		if len(def) > 0 {
			return def, err
		}
		return AuthorityAddr.String(), err
	}
	return rv, nil
}

// ReadAddrFlagOrFrom gets the requested flag or, if it wasn't provided, gets the --from address.
// Returns an error if neither the flag nor --from were provided.
// This assumes that the flag was defined with a default of "".
func ReadAddrFlagOrFrom(clientCtx client.Context, flagSet *pflag.FlagSet, name string) (string, error) {
	rv, err := flagSet.GetString(name)
	if len(rv) > 0 || err != nil {
		return rv, err
	}

	rv = clientCtx.GetFromAddress().String()
	if len(rv) > 0 {
		return rv, nil
	}

	return "", fmt.Errorf("no <%s> provided", name)
}

// AddFlagsEnableDisable adds the --enable and --disable flags and marks them mutually exclusive and one is required.
//
// Use ReadFlagsEnableDisable to read these flags.
func AddFlagsEnableDisable(cmd *cobra.Command, name string) {
	cmd.Flags().Bool(FlagEnable, false, fmt.Sprintf("Set the market's %s field to true", name))
	cmd.Flags().Bool(FlagDisable, false, fmt.Sprintf("Set the market's %s field to false", name))
	cmd.MarkFlagsMutuallyExclusive(FlagEnable, FlagDisable)
	cmd.MarkFlagsOneRequired(FlagEnable, FlagDisable)
}

// ReadFlagsEnableDisable reads the --enable and --disable flags.
// If --enable is given, returns true, if --disable is given, returns false.
//
// This assumes that the flags were defined with AddFlagsEnableDisable.
func ReadFlagsEnableDisable(flagSet *pflag.FlagSet) (bool, error) {
	enable, err := flagSet.GetBool(FlagEnable)
	if enable || err != nil {
		return enable, err
	}
	disable, err := flagSet.GetBool(FlagDisable)
	if disable || err != nil {
		return false, err
	}
	return false, fmt.Errorf("exactly one of --%s or --%s must be provided", FlagEnable, FlagDisable)
}

// AddFlagsAsksBidsBools adds the --asks and --bids flags as bools for limiting search results.
// Marks them mutually exclusive (but not required).
//
// Use ReadFlagsAsksBidsOpt to read them.
func AddFlagsAsksBidsBools(cmd *cobra.Command) {
	cmd.Flags().Bool(FlagAsks, false, "Limit results to only ask orders")
	cmd.Flags().Bool(FlagBids, false, "Limit results to only bid orders")
	cmd.MarkFlagsMutuallyExclusive(FlagAsks, FlagBids)
}

// ReadFlagsAsksBidsOpt reads the --asks and --bids bool flags, returning either "ask", "bid" or "".
//
// This assumes that the flags were defined using AddFlagsAsksBidsBools.
func ReadFlagsAsksBidsOpt(flagSet *pflag.FlagSet) (string, error) {
	isAsk, err := flagSet.GetBool(FlagAsks)
	if err != nil {
		return "", err
	}
	if isAsk {
		return "ask", nil
	}

	isBid, err := flagSet.GetBool(FlagBids)
	if err != nil {
		return "", err
	}
	if isBid {
		return "bid", nil
	}

	return "", nil
}

// ReadFlagOrderOrArg gets a required order id from either the --order flag or the first provided arg.
// This assumes that the flag was defined with a default of 0.
func ReadFlagOrderOrArg(flagSet *pflag.FlagSet, args []string) (uint64, error) {
	orderID, err := flagSet.GetUint64(FlagOrder)
	if err != nil {
		return 0, err
	}

	if len(args) > 0 && len(args[0]) > 0 {
		if orderID != 0 {
			return 0, fmt.Errorf("cannot provide <order id> as both an arg (%q) and flag (--%s %d)", args[0], FlagOrder, orderID)
		}

		orderID, err = strconv.ParseUint(args[0], 10, 64)
		if err != nil {
			return 0, fmt.Errorf("could not convert <order id> arg: %w", err)
		}
	}

	if orderID == 0 {
		return 0, errors.New("no <order id> provided")
	}

	return orderID, nil
}

// ReadFlagMarketOrArg gets a required market id from either the --market flag or the first provided arg.
// This assumes that the flag was defined with a default of 0.
func ReadFlagMarketOrArg(flagSet *pflag.FlagSet, args []string) (uint32, error) {
	marketID, err := flagSet.GetUint32(FlagMarket)
	if err != nil {
		return 0, err
	}

	if len(args) > 0 && len(args[0]) > 0 {
		if marketID != 0 {
			return 0, fmt.Errorf("cannot provide <market id> as both an arg (%q) and flag (--%s %d)", args[0], FlagMarket, marketID)
		}

		var marketID64 uint64
		marketID64, err = strconv.ParseUint(args[0], 10, 32)
		if err != nil {
			return 0, fmt.Errorf("could not convert <market id> arg: %w", err)
		}
		marketID = uint32(marketID64)
	}

	if marketID == 0 {
		return 0, errors.New("no <market id> provided")
	}

	return marketID, nil
}

// ReadCoinsFlag reads a string flag and converts it into sdk.Coins.
// If the flag wasn't provided, this returns nil, nil.
//
// If the flag is a StringSlice, use ReadFlatFeeFlag.
func ReadCoinsFlag(flagSet *pflag.FlagSet, name string) (sdk.Coins, error) {
	value, err := flagSet.GetString(name)
	if len(value) == 0 || err != nil {
		return nil, err
	}
	rv, err := ParseCoins(value)
	if err != nil {
		return nil, fmt.Errorf("error parsing --%s as coins: %w", name, err)
	}
	return rv, nil
}

// ParseCoins parses a string into sdk.Coins.
func ParseCoins(coinsStr string) (sdk.Coins, error) {
	// The sdk.ParseCoinsNormalized func allows for decimals and just truncates if there are some.
	// But I want an error if there's a decimal portion.
	// Its errors also always have "invalid decimal coin expression", and I don't want "decimal" in these errors.
	// I also like having the offending coin string quoted since its safer and clarifies when the coinsStr is "".
	if len(coinsStr) == 0 {
		return nil, nil
	}
	var rv sdk.Coins
	for _, coinStr := range strings.Split(coinsStr, ",") {
		c, err := exchange.ParseCoin(coinStr)
		if err != nil {
			return nil, err
		}
		rv = rv.Add(c)
	}
	return rv, nil
}

// ReadCoinFlag reads a string flag and converts it into *sdk.Coin.
// If the flag wasn't provided, this returns nil, nil.
//
// Use ReadReqCoinFlag if the flag is required.
func ReadCoinFlag(flagSet *pflag.FlagSet, name string) (*sdk.Coin, error) {
	value, err := flagSet.GetString(name)
	if len(value) == 0 || err != nil {
		return nil, err
	}
	rv, err := exchange.ParseCoin(value)
	if err != nil {
		return nil, fmt.Errorf("error parsing --%s as a coin: %w", name, err)
	}
	return &rv, nil
}

// ReadReqCoinFlag reads a string flag and converts it into a sdk.Coin and requires it to have a value.
// Returns an error if not provided.
//
// Use ReadCoinFlag if the flag is optional.
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

// ReadOrderIDsFlag reads a UintSlice flag and converts it into a []uint64.
func ReadOrderIDsFlag(flagSet *pflag.FlagSet, name string) ([]uint64, error) {
	ids, err := flagSet.GetUintSlice(name)
	if len(ids) == 0 || err != nil {
		return nil, err
	}
	rv := make([]uint64, len(ids))
	for i, id := range ids {
		rv[i] = uint64(id)
	}
	return rv, nil
}

// ReadAccessGrantsFlag reads a StringSlice flag and converts it to a slice of AccessGrants.
// This assumes that the flag was defined with a default of nil or []string{}.
func ReadAccessGrantsFlag(flagSet *pflag.FlagSet, name string, def []exchange.AccessGrant) ([]exchange.AccessGrant, error) {
	vals, err := flagSet.GetStringSlice(name)
	if len(vals) == 0 || err != nil {
		return def, err
	}
	return ParseAccessGrants(vals)
}

// permSepRx is a regexp that matches characters that can be used to separate permissions.
var permSepRx = regexp.MustCompile(`[ +.]`)

// ParseAccessGrant parses an AccessGrant from a string with the format "<address>:<perm 1>[+<perm 2>...]".
func ParseAccessGrant(val string) (*exchange.AccessGrant, error) {
	parts := strings.Split(val, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("could not parse %q as an <access grant>: expected format <address>:<permissions>", val)
	}

	addr := strings.TrimSpace(parts[0])
	perms := strings.ToLower(strings.TrimSpace(parts[1]))
	if len(addr) == 0 || len(perms) == 0 {
		return nil, fmt.Errorf("invalid <access grant> %q: both an <address> and <permissions> are required", val)
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
	grants := make([]exchange.AccessGrant, 0, len(vals))
	for _, val := range vals {
		ag, err := ParseAccessGrant(val)
		if err != nil {
			errs = append(errs, err)
		}
		if ag != nil {
			grants = append(grants, *ag)
		}
	}
	return grants, errors.Join(errs...)
}

// ReadFlatFeeFlag reads a StringSlice flag and converts it into a slice of sdk.Coin.
// If the flag wasn't provided, the provided default is returned.
// This assumes that the flag was defined with a default of nil or []string{}.
//
// If the flag is a String, use ReadCoinsFlag.
func ReadFlatFeeFlag(flagSet *pflag.FlagSet, name string, def []sdk.Coin) ([]sdk.Coin, error) {
	vals, err := flagSet.GetStringSlice(name)
	if len(vals) == 0 || err != nil {
		return def, err
	}
	return ParseFlatFeeOptions(vals)
}

// ParseFlatFeeOptions parses each of the provided vals to sdk.Coin.
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

// ReadFeeRatiosFlag reads a StringSlice flag and converts it into a slice of exchange.FeeRatio.
// If the flag wasn't provided, the provided default is returned.
// This assumes that the flag was defined with a default of nil or []string{}.
func ReadFeeRatiosFlag(flagSet *pflag.FlagSet, name string, def []exchange.FeeRatio) ([]exchange.FeeRatio, error) {
	vals, err := flagSet.GetStringSlice(name)
	if len(vals) == 0 || err != nil {
		return def, err
	}
	return ParseFeeRatios(vals)
}

// ParseFeeRatios parses a FeeRatio from each of the provided vals.
func ParseFeeRatios(vals []string) ([]exchange.FeeRatio, error) {
	var errs []error
	ratios := make([]exchange.FeeRatio, 0, len(vals))
	for _, val := range vals {
		ratio, err := exchange.ParseFeeRatio(val)
		if err != nil {
			errs = append(errs, err)
		}
		if ratio != nil {
			ratios = append(ratios, *ratio)
		}
	}
	return ratios, errors.Join(errs...)
}

// ReadSplitsFlag reads a StringSlice flag and converts it into a slice of exchange.DenomSplit.
// This assumes that the flag was defined with a default of nil or []string{}.
func ReadSplitsFlag(flagSet *pflag.FlagSet, name string) ([]exchange.DenomSplit, error) {
	vals, err := flagSet.GetStringSlice(name)
	if len(vals) == 0 || err != nil {
		return nil, err
	}
	return ParseSplits(vals)
}

// ParseSplit parses a DenomSplit from a string with the format "<denom>:<amount>".
func ParseSplit(val string) (*exchange.DenomSplit, error) {
	parts := strings.Split(val, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid denom split %q: expected format <denom>:<amount>", val)
	}

	denom := strings.TrimSpace(parts[0])
	amountStr := strings.TrimSpace(parts[1])
	if len(denom) == 0 || len(amountStr) == 0 {
		return nil, fmt.Errorf("invalid denom split %q: both a <denom> and <amount> are required", val)
	}

	amount, err := strconv.ParseUint(amountStr, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("could not parse %q amount: %w", val, err)
	}

	return &exchange.DenomSplit{Denom: denom, Split: uint32(amount)}, nil
}

// ParseSplits parses a DenomSplit from each of the provided vals.
func ParseSplits(vals []string) ([]exchange.DenomSplit, error) {
	var errs []error
	splits := make([]exchange.DenomSplit, 0, len(vals))
	for _, val := range vals {
		split, err := ParseSplit(val)
		if err != nil {
			errs = append(errs, err)
		}
		if split != nil {
			splits = append(splits, *split)
		}
	}
	return splits, errors.Join(errs...)
}

// ReadStringFlagOrArg gets a required string from either a flag or the first provided arg.
// This assumes that the flag was defined with a default of "".
func ReadStringFlagOrArg(flagSet *pflag.FlagSet, args []string, flagName, varName string) (string, error) {
	rv, err := flagSet.GetString(flagName)
	if err != nil {
		return "", err
	}

	if len(args) > 0 && len(args[0]) > 0 {
		if len(rv) > 0 {
			return "", fmt.Errorf("cannot provide <%s> as both an arg (%q) and flag (--%s %q)", varName, args[0], flagName, rv)
		}

		return args[0], nil
	}

	if len(rv) == 0 {
		return "", fmt.Errorf("no <%s> provided", varName)
	}

	return rv, nil
}

// ReadProposalFlag gets the --proposal string value and attempts to read the file in as a Tx in json.
// It then attempts to extract any messages contained in any govv1.MsgSubmitProposal messages in that Tx.
// An error is returned if anything goes wrong.
// This assumes that the flag was defined with a default of "".
func ReadProposalFlag(clientCtx client.Context, flagSet *pflag.FlagSet) (string, []*codectypes.Any, error) {
	propFN, err := flagSet.GetString(FlagProposal)
	if len(propFN) == 0 || err != nil {
		return "", nil, err
	}

	propFileContents, err := os.ReadFile(propFN)
	if err != nil {
		return propFN, nil, err
	}

	var tx txtypes.Tx
	err = clientCtx.Codec.UnmarshalJSON(propFileContents, &tx)
	if err != nil {
		return propFN, nil, fmt.Errorf("failed to unmarshal --%s %q contents as Tx: %w", FlagProposal, propFN, err)
	}

	if tx.Body == nil {
		return propFN, nil, fmt.Errorf("the contents of %q does not have a \"body\"", propFN)
	}

	if len(tx.Body.Messages) == 0 {
		return propFN, nil, fmt.Errorf("the contents of %q does not have any body messages", propFN)
	}

	hadProp := false
	var rv []*codectypes.Any
	for _, msgAny := range tx.Body.Messages {
		prop, isProp := msgAny.GetCachedValue().(*govv1.MsgSubmitProposal)
		if isProp {
			hadProp = true
			rv = append(rv, prop.Messages...)
		}
	}

	if !hadProp {
		return propFN, nil, fmt.Errorf("no %T messages found in %q", &govv1.MsgSubmitProposal{}, propFN)
	}
	if len(rv) == 0 {
		return propFN, nil, fmt.Errorf("no messages found in any %T messages in %q", &govv1.MsgSubmitProposal{}, propFN)
	}

	return propFN, rv, nil
}

// getSingleMsgFromPropFlag reads the --proposal flag and extracts a Msg of a specific type from the file it points to.
// If --proposal wasn't provided, the emptyMsg is returned without error.
// An error is returned if anything goes wrong or the file doesn't have exactly one T.
// The emptyMsg is returned even if an error is returned.
//
// T is the specific type of Msg to look for.
func getSingleMsgFromPropFlag[T sdk.Msg](clientCtx client.Context, flagSet *pflag.FlagSet, emptyMsg T) (T, error) {
	fn, msgs, err := ReadProposalFlag(clientCtx, flagSet)
	if len(fn) == 0 || err != nil {
		return emptyMsg, err
	}

	rvs := make([]T, 0, 1)
	for _, msg := range msgs {
		rv, isRV := msg.GetCachedValue().(T)
		if isRV {
			rvs = append(rvs, rv)
		}
	}

	if len(rvs) == 0 {
		return emptyMsg, fmt.Errorf("no %T found in %q", emptyMsg, fn)
	}
	if len(rvs) != 1 {
		return emptyMsg, fmt.Errorf("%d %T found in %q", len(rvs), emptyMsg, fn)
	}

	return rvs[0], nil
}

// ReadMsgGovCreateMarketRequestFromProposalFlag reads the --proposal flag and extracts the MsgGovCreateMarketRequest from the file points to.
// An error is returned if anything goes wrong or the file doesn't have exactly one MsgGovCreateMarketRequest.
// A MsgGovCreateMarketRequest is returned even if an error is returned.
// This assumes that the flag was defined with a default of "".
func ReadMsgGovCreateMarketRequestFromProposalFlag(clientCtx client.Context, flagSet *pflag.FlagSet) (*exchange.MsgGovCreateMarketRequest, error) {
	return getSingleMsgFromPropFlag(clientCtx, flagSet, &exchange.MsgGovCreateMarketRequest{})
}

// ReadMsgGovManageFeesRequestFromProposalFlag reads the --proposal flag and extracts the MsgGovManageFeesRequest from the file points to.
// An error is returned if anything goes wrong or the file doesn't have exactly one MsgGovManageFeesRequest.
// A MsgGovManageFeesRequest is returned even if an error is returned.
// This assumes that the flag was defined with a default of "".
func ReadMsgGovManageFeesRequestFromProposalFlag(clientCtx client.Context, flagSet *pflag.FlagSet) (*exchange.MsgGovManageFeesRequest, error) {
	return getSingleMsgFromPropFlag(clientCtx, flagSet, &exchange.MsgGovManageFeesRequest{})
}

// ReadFlagUint32OrDefault gets a uit32 flag or returns the provided default.
// This assumes that the flag was defined with a default of 0.
func ReadFlagUint32OrDefault(flagSet *pflag.FlagSet, name string, def uint32) (uint32, error) {
	rv, err := flagSet.GetUint32(name)
	if rv == 0 || err != nil {
		return def, err
	}
	return rv, nil
}

// ReadFlagBoolOrDefault gets a bool flag or returns the provided default.
// This assumes that the flag was defined with a default of false (it actually just ignores that default).
func ReadFlagBoolOrDefault(flagSet *pflag.FlagSet, name string, def bool) (bool, error) {
	// A bool flag is a little different from the others.
	// If someone provides --<name>=false, I want to use that instead of the provided default.
	// The default in here should only be used if there's an error or the flag wasn't given.
	// This effectively ignores if the flag was defined with a default of true, which shouldn't be done anyway.
	rv, err := flagSet.GetBool(name)
	if err != nil {
		return def, err
	}
	flagGiven := false
	flagSet.Visit(func(flag *pflag.Flag) {
		if flag.Name == name {
			flagGiven = true
		}
	})
	if flagGiven {
		return rv, nil
	}
	return def, nil
}

// ReadFlagStringSliceOrDefault gets a string slice flag or returns the provided default.
// This assumes that the flag was defined with a default of nil or []string{}.
func ReadFlagStringSliceOrDefault(flagSet *pflag.FlagSet, name string, def []string) ([]string, error) {
	rv, err := flagSet.GetStringSlice(name)
	if len(rv) == 0 || err != nil {
		return def, err
	}
	return rv, nil
}

// ReadFlagStringOrDefault gets a string flag or returns the provided default.
// This assumes that the flag was defined with a default of "".
func ReadFlagStringOrDefault(flagSet *pflag.FlagSet, name string, def string) (string, error) {
	rv, err := flagSet.GetString(name)
	if len(rv) == 0 || err != nil {
		return def, err
	}
	return rv, nil
}
