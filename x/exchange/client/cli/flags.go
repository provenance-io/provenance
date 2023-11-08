package cli

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/exchange"
)

const (
	FlagAcceptingOrders    = "accepting-orders"
	FlagAccessGrants       = "access-grants"
	FlagAdmin              = "admin"
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
	FlagDescription        = "description"
	FlagDefault            = "default"
	FlagDenom              = "denom"
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
	FlagReqAttrAsk         = "req-attr-ask"
	FlagReqAttrBid         = "req-attr-Bid"
	FlagRevokeAll          = "revoke-all"
	FlagRevoke             = "revoke"
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

// AddFlagsAdmin adds the --admin and --authority flags to a command and makes them mutually exclusive.
// It also makes one of --admin, --authority, and --from required.
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

	rv = clientCtx.FromAddress.String()
	if len(rv) > 0 {
		return rv, nil
	}

	return "", fmt.Errorf("no %s provided", FlagAdmin)
}

// ReadFlagAuthority reads the --authority flag, or if not provided, returns the standard authority address.
func ReadFlagAuthority(flagSet *pflag.FlagSet) (string, error) {
	rv, err := flagSet.GetString(FlagAuthority)
	if len(rv) > 0 || err != nil {
		return rv, err
	}
	return AuthorityAddr.String(), nil
}

// AddFlagsEnableDisable adds the --enable and --disable flags and marks them mutually exclusive and one is required.
func AddFlagsEnableDisable(cmd *cobra.Command, name string) {
	cmd.Flags().Bool(FlagEnable, false, fmt.Sprintf("Set the market's %s field to true", name))
	cmd.Flags().Bool(FlagDisable, false, fmt.Sprintf("Set the market's %s field to false", name))
	cmd.MarkFlagsMutuallyExclusive(FlagEnable, FlagDisable)
	cmd.MarkFlagsOneRequired(FlagEnable, FlagDisable)
}

// ReadFlagsEnableDisable reads the --enable and --disable flags.
// If --enable is given, returns true, if --disable is given, returns false.
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

// ReadFlagsAsksBidsOpt reads the --asks and --bids flags, returning either "ask", "bid" or "".
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

// ReadOrderIDFlagOrArg gets a required order id from either the --order flag or the first provided arg.
func ReadOrderIDFlagOrArg(flagSet *pflag.FlagSet, args []string) (uint64, error) {
	orderID, err := flagSet.GetUint64(FlagOrder)
	if err != nil {
		return 0, err
	}

	if len(args) > 0 && len(args[0]) > 0 {
		if orderID != 0 {
			return 0, fmt.Errorf("cannot provide <order id> as both an arg (%s) and flag (--%s %d)", args[0], FlagOrder, orderID)
		}

		orderID, err = strconv.ParseUint(args[0], 10, 64)
		if err != nil {
			return 0, fmt.Errorf("could not convert <order id> arg %q to uint64: %w", args[0], err)
		}
		if orderID == 0 {
			return 0, errors.New("the <order id> cannot be zero")
		}
	}

	if orderID == 0 {
		return 0, errors.New("no <order id> provided")
	}

	return orderID, nil
}

// ReadMarketIDFlagOrArg gets a required market id from either the --market flag or the first provided arg.
func ReadMarketIDFlagOrArg(flagSet *pflag.FlagSet, args []string) (uint32, error) {
	marketID, err := flagSet.GetUint32(FlagMarket)
	if err != nil {
		return 0, err
	}

	if len(args) > 0 && len(args[0]) > 0 {
		if marketID != 0 {
			return 0, fmt.Errorf("cannot provide <market id> as both an arg (%s) and flag (--%s %d)", args[0], FlagMarket, marketID)
		}

		var marketID64 uint64
		marketID64, err = strconv.ParseUint(args[0], 10, 32)
		if err != nil {
			return 0, fmt.Errorf("could not convert <market id> arg %q to uint32: %w", args[0], err)
		}
		if marketID64 == 0 {
			return 0, errors.New("the <market id> cannot be zero")
		}
		marketID = uint32(marketID64)
	}

	if marketID == 0 {
		return 0, errors.New("no <market id> provided")
	}

	return marketID, nil
}

// ReadStringFlagOrArg gets a required string from either a flag or the first provided arg.
func ReadStringFlagOrArg(flagSet *pflag.FlagSet, args []string, flagName, varName string) (string, error) {
	rv, err := flagSet.GetString(flagName)
	if err != nil {
		return "", err
	}

	if len(args) > 0 && len(args[0]) > 0 {
		if len(rv) > 0 {
			return "", fmt.Errorf("cannot provide <%s> as both an arg (%s) and flag (--%s %s)", varName, args[0], flagName, rv)
		}

		return args[0], nil
	}

	if len(rv) == 0 {
		return "", fmt.Errorf("no <%s> provided", varName)
	}

	return rv, nil
}

// AddFlagsMsgCreateAsk adds all the flags needed for MakeMsgCreateAsk.
func AddFlagsMsgCreateAsk(cmd *cobra.Command) {
	cmd.Flags().String(FlagSeller, "", "The seller (defaults to --from account)")
	cmd.Flags().Uint32(FlagMarket, 0, "The market id (required)")
	cmd.Flags().String(FlagAssets, "", "The assets for this order, e.g. 10nhash (required)")
	cmd.Flags().String(FlagPrice, "", "The price for this order, e.g. 10nhash (required)")
	cmd.Flags().String(FlagSettlementFee, "", "The settlement fee Coin string for this order, e.g. 10nhash")
	cmd.Flags().Bool(FlagPartial, false, "Allow this order to be partially filled")
	cmd.Flags().String(FlagExternalID, "", "The external id for this order")
	cmd.Flags().String(FlagCreationFee, "", "The ask order creation fee, e.g. 10nhash")

	cmd.MarkFlagsOneRequired(flags.FlagFrom, FlagSeller)
	MarkFlagsRequired(cmd, FlagMarket, FlagAssets, FlagPrice)

	AddUseArgs(cmd,
		ReqSignerUse(FlagSeller),
		ReqFlagUse(FlagMarket, "market id"),
		ReqFlagUse(FlagAssets, "assets"),
		ReqFlagUse(FlagPrice, "price"),
		UseFlagsBreak,
		OptFlagUse(FlagSettlementFee, "seller settlement flat fee"),
		OptFlagUse(FlagPartial, ""),
		OptFlagUse(FlagExternalID, "external id"),
		OptFlagUse(FlagCreationFee, "creation fee"),
	)
	AddUseDetails(cmd, ReqSignerDesc(FlagSeller))

	cmd.Args = cobra.NoArgs
}

// MakeMsgCreateAsk reads all the AddFlagsMsgCreateAsk flags and creates the desired Msg.
// Satisfies the msgMaker type.
func MakeMsgCreateAsk(clientCtx client.Context, flagSet *pflag.FlagSet, _ []string) (sdk.Msg, error) {
	msg := &exchange.MsgCreateAskRequest{}

	errs := make([]error, 8)
	msg.AskOrder.Seller, errs[0] = ReadAddrFlagOrFrom(clientCtx, flagSet, FlagSeller)
	msg.AskOrder.MarketId, errs[1] = flagSet.GetUint32(FlagMarket)
	msg.AskOrder.Assets, errs[2] = ReadReqCoinFlag(flagSet, FlagAssets)
	msg.AskOrder.Price, errs[3] = ReadReqCoinFlag(flagSet, FlagPrice)
	msg.AskOrder.SellerSettlementFlatFee, errs[4] = ReadCoinFlag(flagSet, FlagSettlementFee)
	msg.AskOrder.AllowPartial, errs[5] = flagSet.GetBool(FlagPartial)
	msg.AskOrder.ExternalId, errs[6] = flagSet.GetString(FlagExternalID)
	msg.OrderCreationFee, errs[7] = ReadCoinFlag(flagSet, FlagCreationFee)

	return msg, errors.Join(errs...)
}

// AddFlagsMsgCreateBid adds all the flags needed for MakeMsgCreateBid.
func AddFlagsMsgCreateBid(cmd *cobra.Command) {
	cmd.Flags().String(FlagBuyer, "", "The buyer (defaults to --from account)")
	cmd.Flags().Uint32(FlagMarket, 0, "The market id (required)")
	cmd.Flags().String(FlagAssets, "", "The assets for this order, e.g. 10nhash (required)")
	cmd.Flags().String(FlagPrice, "", "The price for this order, e.g. 10nhash (required)")
	cmd.Flags().String(FlagSettlementFee, "", "The settlement fee Coin string for this order, e.g. 10nhash")
	cmd.Flags().Bool(FlagPartial, false, "Allow this order to be partially filled")
	cmd.Flags().String(FlagExternalID, "", "The external id for this order")
	cmd.Flags().String(FlagCreationFee, "", "The bid order creation fee, e.g. 10nhash")

	cmd.MarkFlagsOneRequired(flags.FlagFrom, FlagBuyer)
	MarkFlagsRequired(cmd, FlagMarket, FlagAssets, FlagPrice)

	AddUseArgs(cmd,
		ReqSignerUse(FlagBuyer),
		ReqFlagUse(FlagMarket, "market id"),
		ReqFlagUse(FlagAssets, "assets"),
		ReqFlagUse(FlagPrice, "price"),
		UseFlagsBreak,
		OptFlagUse(FlagSettlementFee, "seller settlement flat fee"),
		OptFlagUse(FlagPartial, ""),
		OptFlagUse(FlagExternalID, "external id"),
		OptFlagUse(FlagCreationFee, "creation fee"),
	)
	AddUseDetails(cmd, ReqSignerDesc(FlagBuyer))

	cmd.Args = cobra.NoArgs
}

// MakeMsgCreateBid reads all the AddFlagsMsgCreateBid flags and creates the desired Msg.
// Satisfies the msgMaker type.
func MakeMsgCreateBid(clientCtx client.Context, flagSet *pflag.FlagSet, _ []string) (sdk.Msg, error) {
	msg := &exchange.MsgCreateBidRequest{}

	errs := make([]error, 8)
	msg.BidOrder.Buyer, errs[0] = ReadAddrFlagOrFrom(clientCtx, flagSet, FlagBuyer)
	msg.BidOrder.MarketId, errs[1] = flagSet.GetUint32(FlagMarket)
	msg.BidOrder.Assets, errs[2] = ReadReqCoinFlag(flagSet, FlagAssets)
	msg.BidOrder.Price, errs[3] = ReadReqCoinFlag(flagSet, FlagPrice)
	msg.BidOrder.BuyerSettlementFees, errs[4] = ReadCoinsFlag(flagSet, FlagSettlementFee)
	msg.BidOrder.AllowPartial, errs[5] = flagSet.GetBool(FlagPartial)
	msg.BidOrder.ExternalId, errs[6] = flagSet.GetString(FlagExternalID)
	msg.OrderCreationFee, errs[7] = ReadCoinFlag(flagSet, FlagCreationFee)

	return msg, errors.Join(errs...)
}

// AddFlagsMsgCancelOrder adds all the flags needed for the MakeMsgCancelOrder.
func AddFlagsMsgCancelOrder(cmd *cobra.Command) {
	cmd.Flags().String(FlagSigner, "", "The signer (defaults to --from account)")
	cmd.Flags().Uint64(FlagOrder, 0, "The order id")

	cmd.MarkFlagsOneRequired(flags.FlagFrom, FlagSigner)

	AddUseArgs(cmd,
		fmt.Sprintf("{<order id>|--%s <order id>}", FlagOrder),
		ReqSignerUse(FlagSigner),
	)
	AddUseDetails(cmd,
		ReqSignerDesc(FlagSigner),
		"The <order id> must be provided either as the first argument or using the --order flag, but not both.",
	)

	cmd.Args = cobra.MaximumNArgs(1)
}

// MakeMsgCancelOrder reads all the AddFlagsMsgCancelOrder flags and the provided args and creates the desired Msg.
// Satisfies the msgMaker type.
func MakeMsgCancelOrder(clientCtx client.Context, flagSet *pflag.FlagSet, args []string) (sdk.Msg, error) {
	msg := &exchange.MsgCancelOrderRequest{}

	errs := make([]error, 2)
	msg.Signer, errs[0] = ReadAddrFlagOrFrom(clientCtx, flagSet, FlagSigner)
	msg.OrderId, errs[1] = ReadOrderIDFlagOrArg(flagSet, args)

	return msg, errors.Join(errs...)
}

// AddFlagsMsgFillBids adds all the flags needed for MakeMsgFillBids.
func AddFlagsMsgFillBids(cmd *cobra.Command) {
	cmd.Flags().String(FlagSeller, "", "The seller (defaults to --from account)")
	cmd.Flags().Uint32(FlagMarket, 0, "The market id (required)")
	cmd.Flags().String(FlagAssets, "", "The total assets you are filling, e.g. 10nhash (required)")
	cmd.Flags().UintSlice(FlagBids, nil, "The bid order ids (repeatable, required)")
	cmd.Flags().String(FlagSettlementFee, "", "The settlement fee Coin string for this order, e.g. 10nhash")
	cmd.Flags().String(FlagCreationFee, "", "The ask order creation fee, e.g. 10nhash")

	cmd.MarkFlagsOneRequired(flags.FlagFrom, FlagSeller)
	MarkFlagsRequired(cmd, FlagMarket, FlagAssets, FlagBids)

	AddUseArgs(cmd,
		ReqSignerUse(FlagSeller),
		ReqFlagUse(FlagMarket, "market id"),
		ReqFlagUse(FlagAssets, "total assets"),
		ReqFlagUse(FlagBids, "bid order ids"),
		UseFlagsBreak,
		OptFlagUse(FlagSettlementFee, "seller settlement flat fee"),
		OptFlagUse(FlagCreationFee, "ask order creation fee"),
	)
	AddUseDetails(cmd, ReqSignerDesc(FlagSeller), RepeatableDesc)

	cmd.Args = cobra.NoArgs
}

// MakeMsgFillBids reads all the AddFlagsMsgFillBids flags and creates the desired Msg.
// Satisfies the msgMaker type.
func MakeMsgFillBids(clientCtx client.Context, flagSet *pflag.FlagSet, _ []string) (sdk.Msg, error) {
	msg := &exchange.MsgFillBidsRequest{}

	errs := make([]error, 6)
	msg.Seller, errs[0] = ReadAddrFlagOrFrom(clientCtx, flagSet, FlagSeller)
	msg.MarketId, errs[1] = flagSet.GetUint32(FlagMarket)
	msg.TotalAssets, errs[2] = ReadCoinsFlag(flagSet, FlagAssets)
	msg.BidOrderIds, errs[3] = ReadOrderIdsFlag(flagSet, FlagBids)
	msg.SellerSettlementFlatFee, errs[4] = ReadCoinFlag(flagSet, FlagSettlementFee)
	msg.AskOrderCreationFee, errs[5] = ReadCoinFlag(flagSet, FlagCreationFee)

	return msg, errors.Join(errs...)
}

// AddFlagsMsgFillAsks adds all the flags needed for MakeMsgFillAsks.
func AddFlagsMsgFillAsks(cmd *cobra.Command) {
	cmd.Flags().String(FlagBuyer, "", "The buyer (defaults to --from account)")
	cmd.Flags().Uint32(FlagMarket, 0, "The market id (required)")
	cmd.Flags().String(FlagPrice, "", "The total price you are paying, e.g. 10nhash (required)")
	cmd.Flags().UintSlice(FlagAsks, nil, "The ask order ids (repeatable, required)")
	cmd.Flags().String(FlagSettlementFee, "", "The settlement fee Coin string for this order, e.g. 10nhash")
	cmd.Flags().String(FlagCreationFee, "", "The bid order creation fee, e.g. 10nhash")

	cmd.MarkFlagsOneRequired(flags.FlagFrom, FlagBuyer)
	MarkFlagsRequired(cmd, FlagMarket, FlagPrice, FlagAsks)

	AddUseArgs(cmd,
		ReqSignerUse(FlagBuyer),
		ReqFlagUse(FlagMarket, "market id"),
		ReqFlagUse(FlagPrice, "total price"),
		ReqFlagUse(FlagAsks, "ask order ids"),
		UseFlagsBreak,
		OptFlagUse(FlagSettlementFee, "buyer settlement fees"),
		OptFlagUse(FlagCreationFee, "bid order creation fee"),
	)
	AddUseDetails(cmd, ReqSignerDesc(FlagBuyer), RepeatableDesc)

	cmd.Args = cobra.NoArgs
}

// MakeMsgFillAsks reads all the AddFlagsMsgFillAsks flags and creates the desired Msg.
// Satisfies the msgMaker type.
func MakeMsgFillAsks(clientCtx client.Context, flagSet *pflag.FlagSet, _ []string) (sdk.Msg, error) {
	msg := &exchange.MsgFillAsksRequest{}

	errs := make([]error, 6)
	msg.Buyer, errs[0] = ReadAddrFlagOrFrom(clientCtx, flagSet, FlagBuyer)
	msg.MarketId, errs[1] = flagSet.GetUint32(FlagMarket)
	msg.TotalPrice, errs[2] = ReadReqCoinFlag(flagSet, FlagPrice)
	msg.AskOrderIds, errs[3] = ReadOrderIdsFlag(flagSet, FlagAsks)
	msg.BuyerSettlementFees, errs[4] = ReadCoinsFlag(flagSet, FlagSettlementFee)
	msg.BidOrderCreationFee, errs[5] = ReadCoinFlag(flagSet, FlagCreationFee)

	return msg, errors.Join(errs...)
}

// AddFlagsMsgMarketSettle adds all the flags needed for MakeMsgMarketSettle.
func AddFlagsMsgMarketSettle(cmd *cobra.Command) {
	AddFlagsAdmin(cmd)
	cmd.Flags().Uint32(FlagMarket, 0, "The market id (required)")
	cmd.Flags().UintSlice(FlagAsks, nil, "The ask order ids (repeatable, required)")
	cmd.Flags().UintSlice(FlagBids, nil, "The bid order ids (repeatable, required)")
	cmd.Flags().Bool(FlagPartial, false, "Expect partial settlement")

	MarkFlagsRequired(cmd, FlagMarket, FlagAsks, FlagBids)

	AddUseArgs(cmd,
		ReqAdminUse,
		ReqFlagUse(FlagMarket, "market id"),
		ReqFlagUse(FlagAsks, "ask order ids"),
		ReqFlagUse(FlagBids, "bid order ids"),
		OptFlagUse(FlagPartial, ""),
	)
	AddUseDetails(cmd, ReqAdminDesc, RepeatableDesc)

	cmd.Args = cobra.NoArgs
}

// MakeMsgMarketSettle reads all the AddFlagsMsgMarketSettle flags and creates the desired Msg.
// Satisfies the msgMaker type.
func MakeMsgMarketSettle(clientCtx client.Context, flagSet *pflag.FlagSet, _ []string) (sdk.Msg, error) {
	msg := &exchange.MsgMarketSettleRequest{}

	errs := make([]error, 5)
	msg.Admin, errs[0] = ReadFlagsAdminOrFrom(clientCtx, flagSet)
	msg.MarketId, errs[1] = flagSet.GetUint32(FlagMarket)
	msg.AskOrderIds, errs[2] = ReadOrderIdsFlag(flagSet, FlagAsks)
	msg.BidOrderIds, errs[3] = ReadOrderIdsFlag(flagSet, FlagBids)
	msg.ExpectPartial, errs[4] = flagSet.GetBool(FlagPartial)

	return msg, errors.Join(errs...)
}

// AddFlagsMsgMarketSetOrderExternalID adds all the flags needed for MakeMsgMarketSetOrderExternalID.
func AddFlagsMsgMarketSetOrderExternalID(cmd *cobra.Command) {
	AddFlagsAdmin(cmd)
	cmd.Flags().Uint32(FlagMarket, 0, "The market id (required)")
	cmd.Flags().Uint64(FlagOrder, 0, "The order id (required)")
	cmd.Flags().String(FlagExternalID, "", "The new external id for this order")

	MarkFlagsRequired(cmd, FlagMarket, FlagOrder)

	AddUseArgs(cmd,
		ReqAdminUse,
		ReqFlagUse(FlagMarket, "market id"),
		ReqFlagUse(FlagOrder, "order id"),
		OptFlagUse(FlagExternalID, "external id"),
	)
	AddUseDetails(cmd, ReqAdminDesc)

	cmd.Args = cobra.NoArgs
}

// MakeMsgMarketSetOrderExternalID reads all the AddFlagsMsgMarketSetOrderExternalID flags and creates the desired Msg.
// Satisfies the msgMaker type.
func MakeMsgMarketSetOrderExternalID(clientCtx client.Context, flagSet *pflag.FlagSet, _ []string) (sdk.Msg, error) {
	msg := &exchange.MsgMarketSetOrderExternalIDRequest{}

	errs := make([]error, 4)
	msg.Admin, errs[0] = ReadFlagsAdminOrFrom(clientCtx, flagSet)
	msg.MarketId, errs[1] = flagSet.GetUint32(FlagMarket)
	msg.OrderId, errs[2] = flagSet.GetUint64(FlagOrder)
	msg.ExternalId, errs[4] = flagSet.GetString(FlagExternalID)

	return msg, errors.Join(errs...)
}

// AddFlagsMsgMarketWithdraw adds all the flags needed for MakeMsgMarketWithdraw.
func AddFlagsMsgMarketWithdraw(cmd *cobra.Command) {
	AddFlagsAdmin(cmd)
	cmd.Flags().Uint32(FlagMarket, 0, "The market id (required)")
	cmd.Flags().String(FlagTo, "", "The address that will receive the funds (required)")
	cmd.Flags().String(FlagAmount, "", "The amount to withdraw (required)")

	MarkFlagsRequired(cmd, FlagMarket, FlagTo, FlagAmount)

	AddUseArgs(cmd,
		ReqAdminUse,
		ReqFlagUse(FlagMarket, "market id"),
		ReqFlagUse(FlagTo, "to address"),
		ReqFlagUse(FlagAmount, "amount"),
	)
	AddUseDetails(cmd, ReqAdminDesc)

	cmd.Args = cobra.NoArgs
}

// MakeMsgMarketWithdraw reads all the AddFlagsMsgMarketWithdraw flags and creates the desired Msg.
// Satisfies the msgMaker type.
func MakeMsgMarketWithdraw(clientCtx client.Context, flagSet *pflag.FlagSet, _ []string) (sdk.Msg, error) {
	msg := &exchange.MsgMarketWithdrawRequest{}

	errs := make([]error, 4)
	msg.Admin, errs[0] = ReadFlagsAdminOrFrom(clientCtx, flagSet)
	msg.MarketId, errs[1] = flagSet.GetUint32(FlagMarket)
	msg.ToAddress, errs[2] = flagSet.GetString(FlagTo)
	msg.Amount, errs[3] = ReadCoinsFlag(flagSet, FlagAmount)

	return msg, errors.Join(errs...)
}

// AddFlagsMarketDetails adds all the flags needed for ReadFlagsMarketDetails.
func AddFlagsMarketDetails(cmd *cobra.Command) {
	cmd.Flags().String(FlagName, "", fmt.Sprintf("A short name for the market (max %d chars)", exchange.MaxName))
	cmd.Flags().String(FlagDescription, "", fmt.Sprintf("A description of the market (max %d chars)", exchange.MaxDescription))
	cmd.Flags().String(FlagURL, "", fmt.Sprintf("The market's website URL (max %d chars)", exchange.MaxWebsiteURL))
	cmd.Flags().String(FlagIcon, "", fmt.Sprintf("The market's icon URI (max %d chars)", exchange.MaxIconURI))
}

// ReadFlagsMarketDetails reads all the AddFlagsMarketDetails flags and creates the desired MarketDetails.
func ReadFlagsMarketDetails(flagSet *pflag.FlagSet) (exchange.MarketDetails, error) {
	rv := exchange.MarketDetails{}

	errs := make([]error, 4)
	rv.Name, errs[0] = flagSet.GetString(FlagName)
	rv.Description, errs[1] = flagSet.GetString(FlagDescription)
	rv.WebsiteUrl, errs[2] = flagSet.GetString(FlagURL)
	rv.IconUri, errs[3] = flagSet.GetString(FlagIcon)

	return rv, errors.Join(errs...)
}

// AddFlagsMsgMarketUpdateDetails adds all the flags needed for MakeMsgMarketUpdateDetails.
func AddFlagsMsgMarketUpdateDetails(cmd *cobra.Command) {
	AddFlagsAdmin(cmd)
	cmd.Flags().Uint32(FlagMarket, 0, "The market id (required)")
	AddFlagsMarketDetails(cmd)

	MarkFlagsRequired(cmd, FlagMarket)

	AddUseArgs(cmd,
		ReqAdminUse,
		ReqFlagUse(FlagMarket, "market id"),
		UseFlagsBreak,
		OptFlagUse(FlagName, "name"),
		OptFlagUse(FlagDescription, "description"),
		OptFlagUse(FlagURL, "website url"),
		OptFlagUse(FlagIcon, "icon uri"),
	)
	AddUseDetails(cmd,
		ReqAdminDesc,
		`All fields of a market's details will be updated.
If you omit an optional flag, that field will be updated to an empty string.`,
	)

	cmd.Args = cobra.NoArgs
}

// MakeMsgMarketUpdateDetails reads all the AddFlagsMsgMarketUpdateDetails flags and creates the desired Msg.
// Satisfies the msgMaker type.
func MakeMsgMarketUpdateDetails(clientCtx client.Context, flagSet *pflag.FlagSet, _ []string) (sdk.Msg, error) {
	msg := &exchange.MsgMarketUpdateDetailsRequest{}

	errs := make([]error, 3)
	msg.Admin, errs[0] = ReadFlagsAdminOrFrom(clientCtx, flagSet)
	msg.MarketId, errs[1] = flagSet.GetUint32(FlagMarket)
	msg.MarketDetails, errs[2] = ReadFlagsMarketDetails(flagSet)

	return msg, errors.Join(errs...)
}

// AddFlagsMsgMarketUpdateEnabled adds all the flags needed for MakeMsgMarketUpdateEnabled.
func AddFlagsMsgMarketUpdateEnabled(cmd *cobra.Command) {
	AddFlagsAdmin(cmd)
	cmd.Flags().Uint32(FlagMarket, 0, "The market id (required)")
	AddFlagsEnableDisable(cmd, "accepting_orders")

	MarkFlagsRequired(cmd, FlagMarket)

	AddUseArgs(cmd,
		ReqAdminUse,
		ReqFlagUse(FlagMarket, "market id"),
		ReqEnableDisableUse,
	)
	AddUseDetails(cmd, ReqAdminDesc, ReqEnableDisableDesc)

	cmd.Args = cobra.NoArgs
}

// MakeMsgMarketUpdateEnabled reads all the AddFlagsMsgMarketUpdateEnabled flags and creates the desired Msg.
// Satisfies the msgMaker type.
func MakeMsgMarketUpdateEnabled(clientCtx client.Context, flagSet *pflag.FlagSet, _ []string) (sdk.Msg, error) {
	msg := &exchange.MsgMarketUpdateEnabledRequest{}

	errs := make([]error, 3)
	msg.Admin, errs[0] = ReadFlagsAdminOrFrom(clientCtx, flagSet)
	msg.MarketId, errs[1] = flagSet.GetUint32(FlagMarket)
	msg.AcceptingOrders, errs[2] = ReadFlagsEnableDisable(flagSet)

	return msg, errors.Join(errs...)
}

// AddFlagsMsgMarketUpdateUserSettle adds all the flags needed for MakeMsgMarketUpdateUserSettle.
func AddFlagsMsgMarketUpdateUserSettle(cmd *cobra.Command) {
	AddFlagsAdmin(cmd)
	cmd.Flags().Uint32(FlagMarket, 0, "The market id (required)")
	AddFlagsEnableDisable(cmd, "allow_user_settlement")

	MarkFlagsRequired(cmd, FlagMarket)

	AddUseArgs(cmd,
		ReqAdminUse,
		ReqFlagUse(FlagMarket, "market id"),
		ReqEnableDisableUse,
	)
	AddUseDetails(cmd, ReqAdminDesc, ReqEnableDisableDesc)

	cmd.Args = cobra.NoArgs
}

// MakeMsgMarketUpdateUserSettle reads all the AddFlagsMsgMarketUpdateUserSettle flags and creates the desired Msg.
// Satisfies the msgMaker type.
func MakeMsgMarketUpdateUserSettle(clientCtx client.Context, flagSet *pflag.FlagSet, _ []string) (sdk.Msg, error) {
	msg := &exchange.MsgMarketUpdateUserSettleRequest{}

	errs := make([]error, 3)
	msg.Admin, errs[0] = ReadFlagsAdminOrFrom(clientCtx, flagSet)
	msg.MarketId, errs[1] = flagSet.GetUint32(FlagMarket)
	msg.AllowUserSettlement, errs[2] = ReadFlagsEnableDisable(flagSet)

	return msg, errors.Join(errs...)
}

// AddFlagsMsgMarketManagePermissions adds all the flags needed for MakeMsgMarketManagePermissions.
func AddFlagsMsgMarketManagePermissions(cmd *cobra.Command) {
	AddFlagsAdmin(cmd)
	cmd.Flags().Uint32(FlagMarket, 0, "The market id (required)")
	cmd.Flags().StringSlice(FlagRevokeAll, nil, "Addresses to revoke all permissions from (repeatable)")
	cmd.Flags().StringSlice(FlagRevoke, nil, "<access grants> to remove from the market (repeatable)")
	cmd.Flags().StringSlice(FlagGrant, nil, "<access grants> to add to the market (repeatable)")

	cmd.MarkFlagsOneRequired(FlagRevokeAll, FlagRevoke, FlagGrant)
	MarkFlagsRequired(cmd, FlagMarket)

	AddUseArgs(cmd,
		ReqAdminUse,
		ReqFlagUse(FlagMarket, "market id"),
		UseFlagsBreak,
		OptFlagUse(FlagRevokeAll, "addresses"),
		OptFlagUse(FlagRevoke, "access grants"),
		OptFlagUse(FlagGrant, "access grants"),
	)
	AddUseDetails(cmd, ReqAdminDesc, RepeatableDesc, AccessGrantsDesc)

	cmd.Args = cobra.NoArgs
}

// MakeMsgMarketManagePermissions reads all the AddFlagsMsgMarketManagePermissions flags and creates the desired Msg.
// Satisfies the msgMaker type.
func MakeMsgMarketManagePermissions(clientCtx client.Context, flagSet *pflag.FlagSet, _ []string) (sdk.Msg, error) {
	msg := &exchange.MsgMarketManagePermissionsRequest{}

	errs := make([]error, 5)
	msg.Admin, errs[0] = ReadFlagsAdminOrFrom(clientCtx, flagSet)
	msg.MarketId, errs[1] = flagSet.GetUint32(FlagMarket)
	msg.RevokeAll, errs[2] = flagSet.GetStringSlice(FlagRevokeAll)
	msg.ToRevoke, errs[3] = ReadAccessGrantsFlag(flagSet, FlagRevoke)
	msg.ToGrant, errs[4] = ReadAccessGrantsFlag(flagSet, FlagGrant)

	return msg, errors.Join(errs...)
}

// AddFlagsMsgMarketManageReqAttrs adds all the flags needed for MakeMsgMarketManageReqAttrs.
func AddFlagsMsgMarketManageReqAttrs(cmd *cobra.Command) {
	AddFlagsAdmin(cmd)
	cmd.Flags().Uint32(FlagMarket, 0, "The market id (required)")
	cmd.Flags().StringSlice(FlagAskAdd, nil, "The create-ask required attributes to add (repeatable)")
	cmd.Flags().StringSlice(FlagAskRemove, nil, "The create-ask required attributes to remove (repeatable)")
	cmd.Flags().StringSlice(FlagBidAdd, nil, "The create-bid required attributes to add (repeatable)")
	cmd.Flags().StringSlice(FlagBidRemove, nil, "The create-bid required attributes to remove (repeatable)")

	cmd.MarkFlagsOneRequired(FlagAskAdd, FlagAskRemove, FlagBidAdd, FlagBidRemove)
	MarkFlagsRequired(cmd, FlagMarket)

	AddUseArgs(cmd,
		ReqAdminUse,
		ReqFlagUse(FlagMarket, "market id"),
		UseFlagsBreak,
		OptFlagUse(FlagAskAdd, "attrs"),
		OptFlagUse(FlagAskRemove, "attrs"),
		UseFlagsBreak,
		OptFlagUse(FlagBidAdd, "attrs"),
		OptFlagUse(FlagBidRemove, "attrs"),
	)
	AddUseDetails(cmd, ReqAdminDesc, RepeatableDesc)

	cmd.Args = cobra.NoArgs
}

// MakeMsgMarketManageReqAttrs reads all the AddFlagsMsgMarketManageReqAttrs flags and creates the desired Msg.
// Satisfies the msgMaker type.
func MakeMsgMarketManageReqAttrs(clientCtx client.Context, flagSet *pflag.FlagSet, _ []string) (sdk.Msg, error) {
	msg := &exchange.MsgMarketManageReqAttrsRequest{}

	errs := make([]error, 6)
	msg.Admin, errs[0] = ReadFlagsAdminOrFrom(clientCtx, flagSet)
	msg.MarketId, errs[1] = flagSet.GetUint32(FlagMarket)
	msg.CreateAskToAdd, errs[2] = flagSet.GetStringSlice(FlagAskAdd)
	msg.CreateAskToRemove, errs[3] = flagSet.GetStringSlice(FlagAskRemove)
	msg.CreateBidToAdd, errs[4] = flagSet.GetStringSlice(FlagBidAdd)
	msg.CreateBidToRemove, errs[5] = flagSet.GetStringSlice(FlagBidRemove)

	return msg, errors.Join(errs...)
}

// AddFlagsMsgGovCreateMarket adds all the flags needed for MakeMsgGovCreateMarket.
func AddFlagsMsgGovCreateMarket(cmd *cobra.Command) {
	cmd.Flags().String(FlagAuthority, "", "The authority address to use (defaults to the governance module account)")
	cmd.Flags().Uint32(FlagMarket, 0, "The market id")
	AddFlagsMarketDetails(cmd)
	cmd.Flags().StringSlice(FlagCreateAsk, nil, "The create-ask fee options, e.g. 10nhash (repeatable)")
	cmd.Flags().StringSlice(FlagCreateBid, nil, "The create-bid fee options, e.g. 10nhash (repeatable)")
	cmd.Flags().StringSlice(FlagSellerFlat, nil, "The seller settlement flat fee options, e.g. 10nhash (repeatable)")
	cmd.Flags().StringSlice(FlagSellerRatios, nil, "The seller settlement fee ratios, e.g. 100nhash:1nhash (repeatable)")
	cmd.Flags().StringSlice(FlagBuyerFlat, nil, "The buyer settlement flat fee options, e.g. 10nhash (repeatable)")
	cmd.Flags().StringSlice(FlagBuyerRatios, nil, "The buyer settlement fee ratios, e.g. 100nhash:1nhash (repeatable)")
	cmd.Flags().Bool(FlagAcceptingOrders, false, "The market should allow orders to be created")
	cmd.Flags().Bool(FlagAllowUserSettle, false, "The market should allow user-initiated settlement")
	cmd.Flags().StringSlice(FlagAccessGrants, nil, "The <access grants> that the market should have (repeatable)")
	cmd.Flags().StringSlice(FlagReqAttrAsk, nil, "Attributes required to create ask orders (repeatable)")
	cmd.Flags().StringSlice(FlagReqAttrBid, nil, "Attributes required to create bid orders (repeatable)")

	cmd.MarkFlagsOneRequired(
		FlagMarket, FlagCreateAsk, FlagCreateBid,
		FlagSellerFlat, FlagSellerRatios, FlagBuyerFlat, FlagBuyerRatios,
		FlagAcceptingOrders, FlagAllowUserSettle, FlagAccessGrants,
		FlagReqAttrAsk, FlagReqAttrBid,
	)

	AddUseArgs(cmd,
		OptFlagUse(FlagAuthority, "authority"),
		OptFlagUse(FlagMarket, "market id"),
		UseFlagsBreak,
		OptFlagUse(FlagName, "name"),
		OptFlagUse(FlagDescription, "description"),
		OptFlagUse(FlagURL, "website url"),
		OptFlagUse(FlagIcon, "icon uri"),
		UseFlagsBreak,
		OptFlagUse(FlagCreateAsk, "coins"),
		OptFlagUse(FlagCreateBid, "coins"),
		UseFlagsBreak,
		OptFlagUse(FlagSellerFlat, "coins"),
		OptFlagUse(FlagSellerRatios, "fee ratios"),
		UseFlagsBreak,
		OptFlagUse(FlagBuyerFlat, "coins"),
		OptFlagUse(FlagBuyerRatios, "fee ratios"),
		UseFlagsBreak,
		OptFlagUse(FlagAcceptingOrders, ""),
		OptFlagUse(FlagAllowUserSettle, ""),
		UseFlagsBreak,
		OptFlagUse(FlagAccessGrants, "access grants"),
		UseFlagsBreak,
		OptFlagUse(FlagReqAttrAsk, "attrs"),
		OptFlagUse(FlagReqAttrBid, "attrs"),
	)
	AddUseDetails(cmd, AuthorityDesc, RepeatableDesc, AccessGrantsDesc, FeeRatioDesc)

	cmd.Args = cobra.NoArgs
}

// MakeMsgGovCreateMarket reads all the AddFlagsMsgGovCreateMarket flags and creates the desired Msg.
// Satisfies the msgMaker type.
func MakeMsgGovCreateMarket(_ client.Context, flagSet *pflag.FlagSet, _ []string) (sdk.Msg, error) {
	msg := &exchange.MsgGovCreateMarketRequest{}

	errs := make([]error, 14)
	msg.Authority, errs[0] = ReadFlagAuthority(flagSet)
	msg.Market.MarketId, errs[1] = flagSet.GetUint32(FlagMarket)
	msg.Market.MarketDetails, errs[2] = ReadFlagsMarketDetails(flagSet)
	msg.Market.FeeCreateAskFlat, errs[3] = ReadFlatFeeFlag(flagSet, FlagCreateAsk)
	msg.Market.FeeCreateBidFlat, errs[4] = ReadFlatFeeFlag(flagSet, FlagCreateBid)
	msg.Market.FeeSellerSettlementFlat, errs[5] = ReadFlatFeeFlag(flagSet, FlagSellerFlat)
	msg.Market.FeeSellerSettlementRatios, errs[6] = ReadFeeRatiosFlag(flagSet, FlagSellerRatios)
	msg.Market.FeeBuyerSettlementFlat, errs[7] = ReadFlatFeeFlag(flagSet, FlagBuyerFlat)
	msg.Market.FeeBuyerSettlementRatios, errs[8] = ReadFeeRatiosFlag(flagSet, FlagBuyerRatios)
	msg.Market.AcceptingOrders, errs[9] = flagSet.GetBool(FlagAcceptingOrders)
	msg.Market.AllowUserSettlement, errs[10] = flagSet.GetBool(FlagAllowUserSettle)
	msg.Market.AccessGrants, errs[11] = ReadAccessGrantsFlag(flagSet, FlagAccessGrants)
	msg.Market.ReqAttrCreateAsk, errs[12] = flagSet.GetStringSlice(FlagReqAttrAsk)
	msg.Market.ReqAttrCreateBid, errs[13] = flagSet.GetStringSlice(FlagReqAttrBid)

	return msg, errors.Join(errs...)
}

// AddFlagsMsgGovManageFees adds all the flags needed for MakeMsgGovManageFees.
func AddFlagsMsgGovManageFees(cmd *cobra.Command) {
	cmd.Flags().String(FlagAuthority, "", "The authority address to use (defaults to the governance module account)")
	cmd.Flags().Uint32(FlagMarket, 0, "The market id (required)")
	cmd.Flags().StringSlice(FlagAskAdd, nil, "Create-ask flat fee options to add, e.g. 10nhash (repeatable)")
	cmd.Flags().StringSlice(FlagAskRemove, nil, "Create-ask flat fee options to remove, e.g. 10nhash (repeatable)")
	cmd.Flags().StringSlice(FlagBidAdd, nil, "Create-bid flat fee options to add, e.g. 10nhash (repeatable)")
	cmd.Flags().StringSlice(FlagBidRemove, nil, "Create-bid flat fee options to remove, e.g. 10nhash (repeatable)")
	cmd.Flags().StringSlice(FlagSellerFlatAdd, nil, "Seller settlement flat fee options to add, e.g. 10nhash (repeatable)")
	cmd.Flags().StringSlice(FlagSellerFlatRemove, nil, "Seller settlement flat fee options to remove, e.g. 10nhash (repeatable)")
	cmd.Flags().StringSlice(FlagSellerRatiosAdd, nil, "Seller settlement fee ratios to add, e.g. 100nhash:1nhash (repeatable)")
	cmd.Flags().StringSlice(FlagSellerRatiosRemove, nil, "Seller settlement fee ratios to remove, e.g. 100nhash:1nhash (repeatable)")
	cmd.Flags().StringSlice(FlagBuyerFlatAdd, nil, "Buyer settlement flat fee options to add, e.g. 10nhash (repeatable)")
	cmd.Flags().StringSlice(FlagBuyerFlatRemove, nil, "Buyer settlement flat fee options to remove, e.g. 10nhash (repeatable)")
	cmd.Flags().StringSlice(FlagBuyerRatiosAdd, nil, "Seller settlement fee ratios to add, e.g. 100nhash:1nhash (repeatable)")
	cmd.Flags().StringSlice(FlagBuyerRatiosRemove, nil, "Seller settlement fee ratios to remove, e.g. 100nhash:1nhash (repeatable)")

	MarkFlagsRequired(cmd, FlagMarket)
	cmd.MarkFlagsOneRequired(
		FlagAskAdd, FlagAskRemove, FlagBidAdd, FlagBidRemove,
		FlagSellerFlatAdd, FlagSellerFlatRemove, FlagSellerRatiosAdd, FlagSellerRatiosRemove,
		FlagBuyerFlatAdd, FlagBuyerFlatRemove, FlagBuyerRatiosAdd, FlagBuyerRatiosRemove,
	)

	AddUseArgs(cmd,
		ReqFlagUse(FlagMarket, "market id"),
		OptFlagUse(FlagAuthority, "authority"),
		UseFlagsBreak,
		OptFlagUse(FlagAskAdd, "coins"),
		OptFlagUse(FlagAskRemove, "coins"),
		UseFlagsBreak,
		OptFlagUse(FlagBidAdd, "coins"),
		OptFlagUse(FlagBidRemove, "coins"),
		UseFlagsBreak,
		OptFlagUse(FlagSellerFlatAdd, "coins"),
		OptFlagUse(FlagSellerFlatRemove, "coins"),
		UseFlagsBreak,
		OptFlagUse(FlagSellerRatiosAdd, "fee ratios"),
		OptFlagUse(FlagSellerRatiosRemove, "fee ratios"),
		UseFlagsBreak,
		OptFlagUse(FlagBuyerFlatAdd, "coins"),
		OptFlagUse(FlagBuyerFlatRemove, "coins"),
		UseFlagsBreak,
		OptFlagUse(FlagBuyerRatiosAdd, "fee ratios"),
		OptFlagUse(FlagBuyerRatiosRemove, "fee ratios"),
	)
	AddUseDetails(cmd, AuthorityDesc, RepeatableDesc, FeeRatioDesc)

	cmd.Args = cobra.NoArgs
}

// MakeMsgGovManageFees reads all the AddFlagsMsgGovManageFees flags and creates the desired Msg.
// Satisfies the msgMaker type.
func MakeMsgGovManageFees(_ client.Context, flagSet *pflag.FlagSet, _ []string) (sdk.Msg, error) {
	msg := &exchange.MsgGovManageFeesRequest{}

	errs := make([]error, 14)
	msg.Authority, errs[0] = ReadFlagAuthority(flagSet)
	msg.MarketId, errs[1] = flagSet.GetUint32(FlagMarket)
	msg.AddFeeCreateAskFlat, errs[2] = ReadFlatFeeFlag(flagSet, FlagAskAdd)
	msg.RemoveFeeCreateAskFlat, errs[3] = ReadFlatFeeFlag(flagSet, FlagAskRemove)
	msg.AddFeeCreateBidFlat, errs[4] = ReadFlatFeeFlag(flagSet, FlagBidAdd)
	msg.RemoveFeeCreateBidFlat, errs[5] = ReadFlatFeeFlag(flagSet, FlagBidRemove)
	msg.AddFeeSellerSettlementFlat, errs[6] = ReadFlatFeeFlag(flagSet, FlagSellerFlatAdd)
	msg.RemoveFeeSellerSettlementFlat, errs[7] = ReadFlatFeeFlag(flagSet, FlagSellerFlatRemove)
	msg.AddFeeSellerSettlementRatios, errs[8] = ReadFeeRatiosFlag(flagSet, FlagSellerRatiosAdd)
	msg.RemoveFeeSellerSettlementRatios, errs[9] = ReadFeeRatiosFlag(flagSet, FlagSellerRatiosRemove)
	msg.AddFeeBuyerSettlementFlat, errs[10] = ReadFlatFeeFlag(flagSet, FlagBuyerFlatAdd)
	msg.RemoveFeeBuyerSettlementFlat, errs[11] = ReadFlatFeeFlag(flagSet, FlagBuyerFlatRemove)
	msg.AddFeeBuyerSettlementRatios, errs[12] = ReadFeeRatiosFlag(flagSet, FlagBuyerRatiosAdd)
	msg.RemoveFeeBuyerSettlementRatios, errs[13] = ReadFeeRatiosFlag(flagSet, FlagBuyerRatiosRemove)

	return msg, errors.Join(errs...)
}

// AddFlagsMsgGovUpdateParams adds all the flags needed for MakeMsgGovUpdateParams.
func AddFlagsMsgGovUpdateParams(cmd *cobra.Command) {
	cmd.Flags().String(FlagAuthority, "", "The authority address to use (defaults to the governance module account)")
	cmd.Flags().Uint32(FlagDefault, 0, "The default split (required)")
	cmd.Flags().StringSlice(FlagSplit, nil, "The denom-splits (repeatable)")

	MarkFlagsRequired(cmd, FlagDefault)

	AddUseArgs(cmd,
		ReqFlagUse(FlagDefault, "amount"),
		OptFlagUse(FlagSplit, "splits"),
		OptFlagUse(FlagAuthority, "authority"),
	)
	AddUseDetails(cmd,
		AuthorityDesc,
		RepeatableDesc,
		`A <split> has the format "<denom>:<amount>".
An <amount> is in basis points and is limited to 0 to 10,000 (both inclusive).

Example <split>: nhash:500`,
	)

	cmd.Args = cobra.NoArgs
}

// MakeMsgGovUpdateParams reads all the AddFlagsMsgGovUpdateParams flags and creates the desired Msg.
// Satisfies the msgMaker type.
func MakeMsgGovUpdateParams(_ client.Context, flagSet *pflag.FlagSet, _ []string) (sdk.Msg, error) {
	msg := &exchange.MsgGovUpdateParamsRequest{}

	errs := make([]error, 3)
	msg.Authority, errs[0] = ReadFlagAuthority(flagSet)
	msg.Params.DefaultSplit, errs[1] = flagSet.GetUint32(FlagDefault)
	msg.Params.DenomSplits, errs[2] = ReadFlagSplitsFlag(flagSet, FlagSplit)

	return msg, errors.Join(errs...)
}

// AddFlagsQueryOrderFeeCalc adds all the flags needed for MakeQueryOrderFeeCalc.
func AddFlagsQueryOrderFeeCalc(cmd *cobra.Command) {
	cmd.Flags().Bool(FlagAsk, false, "Run calculation on an ask order")
	cmd.Flags().Bool(FlagBid, false, "Run calculation on a bid order")
	cmd.Flags().Uint32(FlagMarket, 0, "The market id (required)")
	cmd.Flags().String(FlagSeller, "", "The seller (for an ask order)")
	cmd.Flags().String(FlagBuyer, "", "The buyer (for a bid order)")
	cmd.Flags().String(FlagAssets, "", "The order assets")
	cmd.Flags().String(FlagPrice, "", "The order price (required)")
	cmd.Flags().String(FlagSettlementFee, "", "The settlement fees")
	cmd.Flags().String(FlagPartial, "", "Allow the order to be partially filled")
	cmd.Flags().String(FlagExternalID, "", "The external id")

	cmd.MarkFlagsMutuallyExclusive(FlagAsk, FlagBid)
	cmd.MarkFlagsOneRequired(FlagAsk, FlagBid)
	cmd.MarkFlagsMutuallyExclusive(FlagBuyer, FlagSeller)
	cmd.MarkFlagsMutuallyExclusive(FlagAsk, FlagBuyer)
	cmd.MarkFlagsMutuallyExclusive(FlagBid, FlagSeller)
	MarkFlagsRequired(cmd, FlagMarket, FlagPrice)

	AddUseArgs(cmd,
		ReqAskBidUse,
		ReqFlagUse(FlagMarket, "market id"),
		ReqFlagUse(FlagPrice, "price"),
	)
	AddUseDetails(cmd, ReqAskBidDesc)
	AddQueryExample(cmd, "--"+FlagAsk, "--"+FlagMarket, "3", "--"+FlagPrice, "10nhash")
	AddQueryExample(cmd, "--"+FlagBid, "--"+FlagMarket, "3", "--"+FlagPrice, "10nhash")

	cmd.Args = cobra.NoArgs
}

// MakeQueryOrderFeeCalc reads all the AddFlagsQueryOrderFeeCalc flags and creates the desired request.
// Satisfies the queryReqMaker type.
func MakeQueryOrderFeeCalc(flagSet *pflag.FlagSet, _ []string) (*exchange.QueryOrderFeeCalcRequest, error) {
	bidOrder := &exchange.BidOrder{}

	errs := make([]error, 10, 11)
	var isAsk, isBid bool
	isAsk, errs[0] = flagSet.GetBool(FlagAsk)
	isBid, errs[1] = flagSet.GetBool(FlagBid)
	bidOrder.MarketId, errs[2] = flagSet.GetUint32(FlagMarket)
	var seller string
	seller, errs[3] = flagSet.GetString(FlagBuyer)
	bidOrder.Buyer, errs[4] = flagSet.GetString(FlagSeller)
	var assets *sdk.Coin
	assets, errs[5] = ReadCoinFlag(flagSet, FlagAssets)
	if assets == nil {
		assets = &sdk.Coin{Denom: "filler", Amount: sdkmath.NewInt(0)}
	}
	bidOrder.Assets = *assets
	bidOrder.Price, errs[6] = ReadReqCoinFlag(flagSet, FlagPrice)
	bidOrder.BuyerSettlementFees, errs[7] = ReadCoinsFlag(flagSet, FlagSettlementFee)
	bidOrder.AllowPartial, errs[8] = flagSet.GetBool(FlagPartial)
	bidOrder.ExternalId, errs[9] = flagSet.GetString(FlagExternalID)

	req := &exchange.QueryOrderFeeCalcRequest{}

	if isAsk {
		req.AskOrder = &exchange.AskOrder{
			MarketId:     bidOrder.MarketId,
			Seller:       seller,
			Assets:       bidOrder.Assets,
			Price:        bidOrder.Price,
			AllowPartial: bidOrder.AllowPartial,
			ExternalId:   bidOrder.ExternalId,
		}
		if len(bidOrder.BuyerSettlementFees) > 0 {
			req.AskOrder.SellerSettlementFlatFee = &bidOrder.BuyerSettlementFees[0]
		}
		if len(bidOrder.BuyerSettlementFees) > 1 {
			errs = append(errs, errors.New("only one settlement fee coin is allowed for ask orders"))
		}
	}

	if isBid {
		req.BidOrder = bidOrder
	}

	return req, errors.Join(errs...)
}

// AddFlagsQueryGetOrder adds all the flags needed for MakeQueryGetOrder.
func AddFlagsQueryGetOrder(cmd *cobra.Command) {
	cmd.Flags().Uint64(FlagOrder, 0, "The order id")

	AddUseArgs(cmd,
		fmt.Sprintf("{<order id>|--%s <order id>}", FlagOrder),
	)
	AddUseDetails(cmd, "An <order id> is required as either an arg or flag, but not both.")
	AddQueryExample(cmd, "8")
	AddQueryExample(cmd, "--"+FlagOrder, "8")

	cmd.Args = cobra.MaximumNArgs(1)
}

// MakeQueryGetOrder reads all the AddFlagsQueryGetOrder flags and creates the desired request.
// Satisfies the queryReqMaker type.
func MakeQueryGetOrder(flagSet *pflag.FlagSet, args []string) (*exchange.QueryGetOrderRequest, error) {
	req := &exchange.QueryGetOrderRequest{}

	var err error
	req.OrderId, err = ReadOrderIDFlagOrArg(flagSet, args)

	return req, err
}

// AddFlagsQueryGetOrderByExternalID adds all the flags needed for MakeQueryGetOrderByExternalID.
func AddFlagsQueryGetOrderByExternalID(cmd *cobra.Command) {
	cmd.Flags().Uint32(FlagMarket, 0, "The market id (required)")
	cmd.Flags().String(FlagExternalID, "", "The external id (required)")

	MarkFlagsRequired(cmd, FlagMarket, FlagExternalID)

	AddUseArgs(cmd,
		ReqFlagUse(FlagMarket, "market id"),
		ReqFlagUse(FlagExternalID, "external id"),
	)
	AddUseDetails(cmd)
	AddQueryExample(cmd, "--"+FlagMarket, "3", "--"+FlagExternalID, "12BD2C9C-9641-4370-A503-802CD7079CAA")

	cmd.Args = cobra.NoArgs
}

// MakeQueryGetOrderByExternalID reads all the AddFlagsQueryGetOrderByExternalID flags and creates the desired request.
// Satisfies the queryReqMaker type.
func MakeQueryGetOrderByExternalID(flagSet *pflag.FlagSet, _ []string) (*exchange.QueryGetOrderByExternalIDRequest, error) {
	req := &exchange.QueryGetOrderByExternalIDRequest{}

	errs := make([]error, 2)
	req.MarketId, errs[0] = flagSet.GetUint32(FlagMarket)
	req.ExternalId, errs[1] = flagSet.GetString(FlagExternalID)

	return req, errors.Join(errs...)
}

// AddFlagsQueryGetMarketOrders adds all the flags needed for MakeQueryGetMarketOrders.
func AddFlagsQueryGetMarketOrders(cmd *cobra.Command) {
	flags.AddPaginationFlagsToCmd(cmd, "orders")
	cmd.Flags().Uint32(FlagMarket, 0, "The market id (required)")
	cmd.Flags().Bool(FlagAsks, false, "Limit results to only ask orders")
	cmd.Flags().Bool(FlagBids, false, "Limit results to only bid orders")
	cmd.Flags().Uint64(FlagOrder, 0, "Limit results to only orders with ids larger than this")

	cmd.MarkFlagsMutuallyExclusive(FlagAsks, FlagBids)

	AddUseArgs(cmd,
		fmt.Sprintf("{<market id>|--%s <market id>}", FlagMarket),
		OptAsksBidsUse,
		OptFlagUse(FlagOrder, "after order id"),
		"[pagination flags]",
	)
	AddUseDetails(cmd,
		"A <market id> is required as either an arg or flag, but not both.",
		OptAsksBidsDesc,
	)
	AddQueryExample(cmd, "3", "--"+FlagAsks)
	AddQueryExample(cmd, "--"+FlagMarket, "1", "--"+FlagOrder, "15", "--"+flags.FlagLimit, "10")

	cmd.Args = cobra.MaximumNArgs(1)
}

// MakeQueryGetMarketOrders reads all the AddFlagsQueryGetMarketOrders flags and creates the desired request.
// Satisfies the queryReqMaker type.
func MakeQueryGetMarketOrders(flagSet *pflag.FlagSet, args []string) (*exchange.QueryGetMarketOrdersRequest, error) {
	req := &exchange.QueryGetMarketOrdersRequest{}

	errs := make([]error, 4)
	req.MarketId, errs[0] = ReadMarketIDFlagOrArg(flagSet, args)
	req.OrderType, errs[1] = ReadFlagsAsksBidsOpt(flagSet)
	req.AfterOrderId, errs[2] = flagSet.GetUint64(FlagOrder)
	req.Pagination, errs[3] = client.ReadPageRequestWithPageKeyDecoded(flagSet)

	return req, errors.Join(errs...)
}

// AddFlagsQueryGetOwnerOrders adds all the flags needed for MakeQueryGetOwnerOrders.
func AddFlagsQueryGetOwnerOrders(cmd *cobra.Command) {
	flags.AddPaginationFlagsToCmd(cmd, "orders")
	cmd.Flags().String(FlagOwner, "", "The owner")
	cmd.Flags().Bool(FlagAsks, false, "Limit results to only ask orders")
	cmd.Flags().Bool(FlagBids, false, "Limit results to only bid orders")
	cmd.Flags().Uint64(FlagOrder, 0, "Limit results to only orders with ids larger than this")

	cmd.MarkFlagsMutuallyExclusive(FlagAsks, FlagBids)

	AddUseArgs(cmd,
		fmt.Sprintf("{<owner>|--%s <owner>}", FlagOwner),
		OptAsksBidsUse,
		OptFlagUse(FlagOrder, "after order id"),
		"[pagination flags]",
	)
	AddUseDetails(cmd,
		"An <owner> is required as either an arg or flag, but not both.",
		OptAsksBidsDesc,
	)
	AddQueryExample(cmd, ExampleAddr, "--"+FlagBids)
	AddQueryExample(cmd, "--"+FlagOwner, ExampleAddr, "--"+FlagAsks, "--"+FlagOrder, "15", "--"+flags.FlagLimit, "10")

	cmd.Args = cobra.MaximumNArgs(1)
}

// MakeQueryGetOwnerOrders reads all the AddFlagsQueryGetOwnerOrders flags and creates the desired request.
// Satisfies the queryReqMaker type.
func MakeQueryGetOwnerOrders(flagSet *pflag.FlagSet, args []string) (*exchange.QueryGetOwnerOrdersRequest, error) {
	req := &exchange.QueryGetOwnerOrdersRequest{}

	errs := make([]error, 5)
	req.Owner, errs[0] = ReadStringFlagOrArg(flagSet, args, FlagOwner, "owner")
	req.OrderType, errs[1] = ReadFlagsAsksBidsOpt(flagSet)
	req.AfterOrderId, errs[2] = flagSet.GetUint64(FlagOrder)
	req.Pagination, errs[3] = client.ReadPageRequestWithPageKeyDecoded(flagSet)

	return req, errors.Join(errs...)
}

// AddFlagsQueryGetAssetOrders adds all the flags needed for MakeQueryGetAssetOrders.
func AddFlagsQueryGetAssetOrders(cmd *cobra.Command) {
	flags.AddPaginationFlagsToCmd(cmd, "orders")
	cmd.Flags().String(FlagDenom, "", "The asset denom")
	cmd.Flags().Bool(FlagAsks, false, "Limit results to only ask orders")
	cmd.Flags().Bool(FlagBids, false, "Limit results to only bid orders")
	cmd.Flags().Uint64(FlagOrder, 0, "Limit results to only orders with ids larger than this")

	cmd.MarkFlagsMutuallyExclusive(FlagAsks, FlagBids)

	AddUseArgs(cmd,
		fmt.Sprintf("{<asset>|--%s <asset>}", FlagDenom),
		OptAsksBidsUse,
		OptFlagUse(FlagOrder, "after order id"),
		"[pagination flags]",
	)
	AddUseDetails(cmd,
		"An <asset> is required as either an arg or flag, but not both.",
		OptAsksBidsDesc,
	)
	AddQueryExample(cmd, "nhash", "--"+FlagAsks)
	AddQueryExample(cmd, "--"+FlagDenom, "nhash", "--"+FlagOrder, "15", "--"+flags.FlagLimit, "10")

	cmd.Args = cobra.MaximumNArgs(1)
}

// MakeQueryGetAssetOrders reads all the AddFlagsQueryGetAssetOrders flags and creates the desired request.
func MakeQueryGetAssetOrders(flagSet *pflag.FlagSet, args []string) (*exchange.QueryGetAssetOrdersRequest, error) {
	req := &exchange.QueryGetAssetOrdersRequest{}

	errs := make([]error, 4)
	req.Asset, errs[0] = ReadStringFlagOrArg(flagSet, args, FlagDenom, "asset")
	req.OrderType, errs[1] = ReadFlagsAsksBidsOpt(flagSet)
	req.AfterOrderId, errs[2] = flagSet.GetUint64(FlagOrder)
	req.Pagination, errs[3] = client.ReadPageRequestWithPageKeyDecoded(flagSet)

	return req, errors.Join(errs...)
}
