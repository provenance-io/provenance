package cli

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/provenance-io/provenance/x/exchange"
)

var (
	AuthorityAddr = authtypes.NewModuleAddress(govtypes.ModuleName)

	ExampleAddr1 = "pb1g4uxzmtsd3j5zerywgc47h6lta047h6lwwxvlw" // = sdk.AccAddress("ExampleAddr1________")
	ExampleAddr2 = "pb1tazhsctdwpkx2styv3eryh6lta047h6l63dw8r" // = sdk.AccAddress("_ExampleAddr2_______")
	ExampleAddr3 = "pb195k527rpd4cxce2pv3j8yv6lta047h6l3kaj79" // = sdk.AccAddress("--ExampleAddr3______")
	ExampleAddr4 = "pb10el8u3tcv9khqmr9g9jxgu35ta047h6l9hc7xs" // = sdk.AccAddress("~~~ExampleAddr4_____")
	ExampleAddr5 = "pb1857n60290psk6urvv4qkgerjx4047h6l5vynnz" // = sdk.AccAddress("====ExampleAddr5____")
)

// A msgMaker is a function that makes a Msg from a client.Context, FlagSet, and set of args.
type msgMaker func(clientCtx client.Context, flagSet *pflag.FlagSet, args []string) (sdk.Msg, error)

var (
	_ msgMaker = MakeMsgCreateAsk
	_ msgMaker = MakeMsgCreateBid
	_ msgMaker = MakeMsgCancelOrder
	_ msgMaker = MakeMsgFillBids
	_ msgMaker = MakeMsgFillAsks
	_ msgMaker = MakeMsgMarketSettle
	_ msgMaker = MakeMsgMarketSetOrderExternalID
	_ msgMaker = MakeMsgMarketWithdraw
	_ msgMaker = MakeMsgMarketUpdateDetails
	_ msgMaker = MakeMsgMarketUpdateEnabled
	_ msgMaker = MakeMsgMarketUpdateUserSettle
	_ msgMaker = MakeMsgMarketManagePermissions
	_ msgMaker = MakeMsgMarketManageReqAttrs
	_ msgMaker = MakeMsgGovCreateMarket
)

// genericTxRunE returns a cobra.Command.RunE function that gets the client.Context, and FlagSet,
// then uses the provided maker to make the Msg that it then generates or broadcasts as a Tx.
func genericTxRunE(maker msgMaker) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		clientCtx, err := client.GetClientTxContext(cmd)
		if err != nil {
			return err
		}

		flagSet := cmd.Flags()
		msg, err := maker(clientCtx, flagSet, args)
		if err != nil {
			return err
		}

		return tx.GenerateOrBroadcastTxCLI(clientCtx, flagSet, msg)
	}
}

// govTxRunE returns a cobra.Command.RunE function that gets the client.Context, and FlagSet,
// then uses the provided maker to make the Msg. The Msg is then either generated or put in a
// governance proposal which is then broadcast as a Tx.
func govTxRunE(maker msgMaker) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		clientCtx, err := client.GetClientTxContext(cmd)
		if err != nil {
			return err
		}

		flagSet := cmd.Flags()
		msg, err := maker(clientCtx, flagSet, args)
		if err != nil {
			return err
		}

		return govcli.GenerateOrBroadcastTxCLIAsGovProp(clientCtx, flagSet, msg)
	}
}

// AddFlagsMsgCreateAsk adds all the flags needed for MakeMsgCreateAsk.
func AddFlagsMsgCreateAsk(cmd *cobra.Command) {
	AddFlagSeller(cmd)
	AddFlagMarket(cmd)
	AddFlagAssets(cmd)
	AddFlagPrice(cmd)
	AddFlagSettlementFee(cmd)
	AddFlagAllowPartial(cmd)
	AddFlagExternalID(cmd)
	AddFlagCreationFee(cmd)
}

// MakeMsgCreateAsk reads all the AddFlagsMsgCreateAsk flags and creates the desired Msg.
func MakeMsgCreateAsk(clientCtx client.Context, flagSet *pflag.FlagSet, _ []string) (sdk.Msg, error) {
	msg := &exchange.MsgCreateAskRequest{}

	errs := make([]error, 8)
	msg.AskOrder.Seller, errs[0] = ReadFlagSellerOrDefault(clientCtx, flagSet)
	msg.AskOrder.MarketId, errs[1] = ReadFlagMarket(flagSet)
	msg.AskOrder.Assets, errs[2] = ReadFlagAssets(flagSet)
	msg.AskOrder.Price, errs[3] = ReadFlagPrice(flagSet)
	msg.AskOrder.SellerSettlementFlatFee, errs[4] = ReadFlagSettlementFeeCoin(flagSet)
	msg.AskOrder.AllowPartial, errs[5] = ReadFlagPartial(flagSet)
	msg.AskOrder.ExternalId, errs[6] = ReadFlagExternalID(flagSet)
	msg.OrderCreationFee, errs[7] = ReadFlagCreationFee(flagSet)

	return msg, errors.Join(errs...)
}

// AddFlagsMsgCreateBid adds all the flags needed for MakeMsgCreateBid.
func AddFlagsMsgCreateBid(cmd *cobra.Command) {
	AddFlagBuyer(cmd)
	AddFlagMarket(cmd)
	AddFlagAssets(cmd)
	AddFlagPrice(cmd)
	AddFlagSettlementFee(cmd)
	AddFlagAllowPartial(cmd)
	AddFlagExternalID(cmd)
	AddFlagCreationFee(cmd)
}

// MakeMsgCreateBid reads all the AddFlagsMsgCreateBid flags and creates the desired Msg.
func MakeMsgCreateBid(clientCtx client.Context, flagSet *pflag.FlagSet, _ []string) (sdk.Msg, error) {
	msg := &exchange.MsgCreateBidRequest{}

	errs := make([]error, 8)
	msg.BidOrder.Buyer, errs[0] = ReadFlagBuyerOrDefault(clientCtx, flagSet)
	msg.BidOrder.MarketId, errs[1] = ReadFlagMarket(flagSet)
	msg.BidOrder.Assets, errs[2] = ReadFlagAssets(flagSet)
	msg.BidOrder.Price, errs[3] = ReadFlagPrice(flagSet)
	msg.BidOrder.BuyerSettlementFees, errs[4] = ReadFlagSettlementFeeCoins(flagSet)
	msg.BidOrder.AllowPartial, errs[5] = ReadFlagPartial(flagSet)
	msg.BidOrder.ExternalId, errs[6] = ReadFlagExternalID(flagSet)
	msg.OrderCreationFee, errs[7] = ReadFlagCreationFee(flagSet)

	return msg, errors.Join(errs...)
}

// AddFlagsMsgCancelOrder adds all the flags needed for the MakeMsgCancelOrder.
func AddFlagsMsgCancelOrder(cmd *cobra.Command) {
	AddFlagSigner(cmd)
	AddFlagOrder(cmd)
}

// MakeMsgCancelOrder reads all the AddFlagsMsgCancelOrder flags and the provided args and creates the desired Msg.
func MakeMsgCancelOrder(clientCtx client.Context, flagSet *pflag.FlagSet, args []string) (sdk.Msg, error) {
	msg := &exchange.MsgCancelOrderRequest{}

	errs := make([]error, 2)
	msg.Signer, errs[0] = ReadFlagSignerOrDefault(clientCtx, flagSet)
	msg.OrderId, errs[1] = ReadFlagOrder(flagSet)
	err := errors.Join(errs...)
	if err != nil {
		return nil, err
	}

	if len(args) > 0 && len(args[0]) > 0 {
		var orderID uint64
		orderID, err = strconv.ParseUint(args[0], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("could not convert <order id> arg %q to uint64: %w", args[0], err)
		}
		if msg.OrderId != 0 && orderID != 0 && msg.OrderId != orderID {
			return nil, fmt.Errorf("cannot provide an <order id> as both an arg (%d) and flag (--%s %d)",
				orderID, FlagOrder, msg.OrderId)
		}
		msg.OrderId = orderID
	}

	if msg.OrderId == 0 {
		return nil, errors.New("no <order id> provided")
	}

	return msg, nil
}

// AddFlagsMsgFillBids adds all the flags needed for MakeMsgFillBids.
func AddFlagsMsgFillBids(cmd *cobra.Command) {
	AddFlagSeller(cmd)
	AddFlagMarket(cmd)
	AddFlagTotalAssets(cmd)
	AddFlagBids(cmd)
	AddFlagSettlementFee(cmd)
	AddFlagCreationFee(cmd)
}

// MakeMsgFillBids reads all the AddFlagsMsgFillBids flags and creates the desired Msg.
func MakeMsgFillBids(clientCtx client.Context, flagSet *pflag.FlagSet, _ []string) (sdk.Msg, error) {
	msg := &exchange.MsgFillBidsRequest{}

	errs := make([]error, 6)
	msg.Seller, errs[0] = ReadFlagSellerOrDefault(clientCtx, flagSet)
	msg.MarketId, errs[1] = ReadFlagMarket(flagSet)
	msg.TotalAssets, errs[2] = ReadFlagTotalAssets(flagSet)
	msg.BidOrderIds, errs[3] = ReadFlagBids(flagSet)
	msg.SellerSettlementFlatFee, errs[4] = ReadFlagSettlementFeeCoin(flagSet)
	msg.AskOrderCreationFee, errs[5] = ReadFlagCreationFee(flagSet)

	return msg, errors.Join(errs...)
}

// AddFlagsMsgFillAsks adds all the flags needed for MakeMsgFillAsks.
func AddFlagsMsgFillAsks(cmd *cobra.Command) {
	AddFlagBuyer(cmd)
	AddFlagMarket(cmd)
	AddFlagTotalPrice(cmd)
	AddFlagAsks(cmd)
	AddFlagSettlementFee(cmd)
	AddFlagCreationFee(cmd)
}

// MakeMsgFillAsks reads all the AddFlagsMsgFillAsks flags and creates the desired Msg.
func MakeMsgFillAsks(clientCtx client.Context, flagSet *pflag.FlagSet, _ []string) (sdk.Msg, error) {
	msg := &exchange.MsgFillAsksRequest{}

	errs := make([]error, 6)
	msg.Buyer, errs[0] = ReadFlagBuyerOrDefault(clientCtx, flagSet)
	msg.MarketId, errs[1] = ReadFlagMarket(flagSet)
	msg.TotalPrice, errs[2] = ReadFlagPrice(flagSet)
	msg.AskOrderIds, errs[3] = ReadFlagAsks(flagSet)
	msg.BuyerSettlementFees, errs[4] = ReadFlagSettlementFeeCoins(flagSet)
	msg.BidOrderCreationFee, errs[5] = ReadFlagCreationFee(flagSet)

	return msg, errors.Join(errs...)
}

// AddFlagsMsgMarketSettle adds all the flags needed for MakeMsgMarketSettle.
func AddFlagsMsgMarketSettle(cmd *cobra.Command) {
	AddFlagAdmin(cmd)
	AddFlagMarket(cmd)
	AddFlagAsks(cmd)
	AddFlagBids(cmd)
	AddFlagExpectPartial(cmd)
}

// MakeMsgMarketSettle reads all the AddFlagsMsgMarketSettle flags and creates the desired Msg.
func MakeMsgMarketSettle(clientCtx client.Context, flagSet *pflag.FlagSet, _ []string) (sdk.Msg, error) {
	msg := &exchange.MsgMarketSettleRequest{}

	errs := make([]error, 5)
	msg.Admin, errs[0] = ReadFlagAdminOrDefault(clientCtx, flagSet)
	msg.MarketId, errs[1] = ReadFlagMarket(flagSet)
	msg.AskOrderIds, errs[2] = ReadFlagAsks(flagSet)
	msg.BidOrderIds, errs[3] = ReadFlagBids(flagSet)
	msg.ExpectPartial, errs[4] = ReadFlagPartial(flagSet)

	return msg, errors.Join(errs...)
}

// AddFlagsMsgMarketSetOrderExternalID adds all the flags needed for MakeMsgMarketSetOrderExternalID.
func AddFlagsMsgMarketSetOrderExternalID(cmd *cobra.Command) {
	AddFlagAdmin(cmd)
	AddFlagMarket(cmd)
	AddFlagOrder(cmd)
	AddFlagExternalID(cmd)
}

// MakeMsgMarketSetOrderExternalID reads all the AddFlagsMsgMarketSetOrderExternalID flags and creates the desired Msg.
func MakeMsgMarketSetOrderExternalID(clientCtx client.Context, flagSet *pflag.FlagSet, _ []string) (sdk.Msg, error) {
	msg := &exchange.MsgMarketSetOrderExternalIDRequest{}

	errs := make([]error, 4)
	msg.Admin, errs[0] = ReadFlagAdminOrDefault(clientCtx, flagSet)
	msg.MarketId, errs[1] = ReadFlagMarket(flagSet)
	msg.OrderId, errs[2] = ReadFlagOrder(flagSet)
	msg.ExternalId, errs[4] = ReadFlagExternalID(flagSet)

	return msg, errors.Join(errs...)
}

// AddFlagsMsgMarketWithdraw adds all the flags needed for MakeMsgMarketWithdraw.
func AddFlagsMsgMarketWithdraw(cmd *cobra.Command) {
	AddFlagAdmin(cmd)
	AddFlagMarket(cmd)
	AddFlagTo(cmd)
	AddFlagAmount(cmd)
}

// MakeMsgMarketWithdraw reads all the AddFlagsMsgMarketWithdraw flags and creates the desired Msg.
func MakeMsgMarketWithdraw(clientCtx client.Context, flagSet *pflag.FlagSet, _ []string) (sdk.Msg, error) {
	msg := &exchange.MsgMarketWithdrawRequest{}

	errs := make([]error, 4)
	msg.Admin, errs[0] = ReadFlagAdminOrDefault(clientCtx, flagSet)
	msg.MarketId, errs[1] = ReadFlagMarket(flagSet)
	msg.ToAddress, errs[2] = ReadFlagTo(flagSet)
	msg.Amount, errs[3] = ReadFlagAmount(flagSet)

	return msg, errors.Join(errs...)
}

// AddFlagsMarketDetails adds all the flags needed for ReadFlagsMarketDetails.
func AddFlagsMarketDetails(cmd *cobra.Command) {
	AddFlagName(cmd)
	AddFlagDescription(cmd)
	AddFlagURL(cmd)
	AddFlagIcon(cmd)
}

// ReadFlagsMarketDetails reads all the AddFlagsMarketDetails flags and creates the desired MarketDetails.
func ReadFlagsMarketDetails(flagSet *pflag.FlagSet) (exchange.MarketDetails, error) {
	rv := exchange.MarketDetails{}

	errs := make([]error, 4)
	rv.Name, errs[0] = ReadFlagName(flagSet)
	rv.Description, errs[1] = ReadFlagDescription(flagSet)
	rv.WebsiteUrl, errs[2] = ReadFlagURL(flagSet)
	rv.IconUri, errs[3] = ReadFlagIcon(flagSet)

	return rv, errors.Join(errs...)
}

// AddFlagsMsgMarketUpdateDetails adds all the flags needed for MakeMsgMarketUpdateDetails.
func AddFlagsMsgMarketUpdateDetails(cmd *cobra.Command) {
	AddFlagAdmin(cmd)
	AddFlagMarket(cmd)
	AddFlagsMarketDetails(cmd)
}

// MakeMsgMarketUpdateDetails reads all the AddFlagsMsgMarketUpdateDetails flags and creates the desired Msg.
func MakeMsgMarketUpdateDetails(clientCtx client.Context, flagSet *pflag.FlagSet, _ []string) (sdk.Msg, error) {
	msg := &exchange.MsgMarketUpdateDetailsRequest{}

	errs := make([]error, 3)
	msg.Admin, errs[0] = ReadFlagAdminOrDefault(clientCtx, flagSet)
	msg.MarketId, errs[1] = ReadFlagMarket(flagSet)
	msg.MarketDetails, errs[2] = ReadFlagsMarketDetails(flagSet)

	return msg, errors.Join(errs...)
}

// AddFlagsMsgMarketUpdateEnabled adds all the flags needed for MakeMsgMarketUpdateEnabled.
func AddFlagsMsgMarketUpdateEnabled(cmd *cobra.Command) {
	AddFlagAdmin(cmd)
	AddFlagMarket(cmd)
	AddFlagEnable(cmd, "accepting_orders")
	AddFlagDisable(cmd, "accepting_orders")
}

// MakeMsgMarketUpdateEnabled reads all the AddFlagsMsgMarketUpdateEnabled flags and creates the desired Msg.
func MakeMsgMarketUpdateEnabled(clientCtx client.Context, flagSet *pflag.FlagSet, _ []string) (sdk.Msg, error) {
	msg := &exchange.MsgMarketUpdateEnabledRequest{}

	errs := make([]error, 5)
	msg.Admin, errs[0] = ReadFlagAdminOrDefault(clientCtx, flagSet)
	msg.MarketId, errs[1] = ReadFlagMarket(flagSet)
	var disable bool
	msg.AcceptingOrders, errs[2] = ReadFlagEnable(flagSet)
	disable, errs[3] = ReadFlagDisable(flagSet)
	if errs[2] == nil && errs[3] == nil && msg.AcceptingOrders == disable {
		errs[4] = fmt.Errorf("exactly one of --%s or --%s must be provided", FlagEnable, FlagDisable)
	}

	return msg, errors.Join(errs...)
}

// AddFlagsMsgMarketUpdateUserSettle adds all the flags needed for MakeMsgMarketUpdateUserSettle.
func AddFlagsMsgMarketUpdateUserSettle(cmd *cobra.Command) {
	AddFlagAdmin(cmd)
	AddFlagMarket(cmd)
	AddFlagEnable(cmd, "allow_user_settlement")
	AddFlagDisable(cmd, "allow_user_settlement")
}

// MakeMsgMarketUpdateUserSettle reads all the AddFlagsMsgMarketUpdateUserSettle flags and creates the desired Msg.
func MakeMsgMarketUpdateUserSettle(clientCtx client.Context, flagSet *pflag.FlagSet, _ []string) (sdk.Msg, error) {
	msg := &exchange.MsgMarketUpdateUserSettleRequest{}

	errs := make([]error, 5)
	msg.Admin, errs[0] = ReadFlagAdminOrDefault(clientCtx, flagSet)
	msg.MarketId, errs[1] = ReadFlagMarket(flagSet)
	var disable bool
	msg.AllowUserSettlement, errs[2] = ReadFlagEnable(flagSet)
	disable, errs[3] = ReadFlagDisable(flagSet)
	if errs[2] == nil && errs[3] == nil && msg.AllowUserSettlement == disable {
		errs[4] = fmt.Errorf("exactly one of --%s or --%s must be provided", FlagEnable, FlagDisable)
	}

	return msg, errors.Join(errs...)
}

// SimplePerms returns a string containing all the Permission.SimpleString() values.
func SimplePerms() string {
	allPerms := exchange.AllPermissions()
	simple := make([]string, len(allPerms))
	for i, perm := range allPerms {
		simple[i] = perm.SimpleString()
	}
	return strings.Join(simple, "  ")
}

// AddFlagsMsgMarketManagePermissions adds all the flags needed for MakeMsgMarketManagePermissions.
func AddFlagsMsgMarketManagePermissions(cmd *cobra.Command) {
	AddFlagAdmin(cmd)
	AddFlagMarket(cmd)
	AddFlagRevokeAll(cmd)
	AddFlagRevoke(cmd)
	AddFlagGrant(cmd)
}

// MakeMsgMarketManagePermissions reads all the AddFlagsMsgMarketManagePermissions flags and creates the desired Msg.
func MakeMsgMarketManagePermissions(clientCtx client.Context, flagSet *pflag.FlagSet, _ []string) (sdk.Msg, error) {
	msg := &exchange.MsgMarketManagePermissionsRequest{}

	errs := make([]error, 5)
	msg.Admin, errs[0] = ReadFlagAdminOrDefault(clientCtx, flagSet)
	msg.MarketId, errs[1] = ReadFlagMarket(flagSet)
	msg.RevokeAll, errs[2] = ReadFlagRevokeAll(flagSet)
	msg.ToRevoke, errs[3] = ReadFlagRevoke(flagSet)
	msg.ToGrant, errs[4] = ReadFlagGrant(flagSet)

	return msg, errors.Join(errs...)
}

// AddFlagsMsgMarketManageReqAttrs adds all the flags needed for MakeMsgMarketManageReqAttrs.
func AddFlagsMsgMarketManageReqAttrs(cmd *cobra.Command) {
	AddFlagAdmin(cmd)
	AddFlagMarket(cmd)
	AddFlagAskAdd(cmd)
	AddFlagAskRemove(cmd)
	AddFlagBidAdd(cmd)
	AddFlagBidRemove(cmd)
}

// MakeMsgMarketManageReqAttrs reads all the AddFlagsMsgMarketManageReqAttrs flags and creates the desired Msg.
func MakeMsgMarketManageReqAttrs(clientCtx client.Context, flagSet *pflag.FlagSet, _ []string) (sdk.Msg, error) {
	msg := &exchange.MsgMarketManageReqAttrsRequest{}

	errs := make([]error, 6)
	msg.Admin, errs[0] = ReadFlagAdminOrDefault(clientCtx, flagSet)
	msg.MarketId, errs[1] = ReadFlagMarket(flagSet)
	msg.CreateAskToAdd, errs[2] = ReadFlagAskAdd(flagSet)
	msg.CreateAskToRemove, errs[3] = ReadFlagAskRemove(flagSet)
	msg.CreateBidToAdd, errs[4] = ReadFlagBidAdd(flagSet)
	msg.CreateBidToRemove, errs[5] = ReadFlagBidRemove(flagSet)

	return msg, errors.Join(errs...)
}

// AddFlagsMsgGovCreateMarket adds all the flags needed for MakeMsgGovCreateMarket.
func AddFlagsMsgGovCreateMarket(cmd *cobra.Command) {
	AddFlagAuthorityString(cmd)
	AddFlagMarket(cmd)
	AddFlagsMarketDetails(cmd)
	AddFlagCreateAsk(cmd)
	AddFlagCreateBid(cmd)
	AddFlagSellerFlat(cmd)
	AddFlagSellerRatios(cmd)
	AddFlagBuyerFlat(cmd)
	AddFlagBuyerRatios(cmd)
	AddFlagAcceptingOrders(cmd)
	AddFlagAllowUserSettle(cmd)
	AddFlagAccessGrants(cmd)
	AddFlagReqAttrAsk(cmd)
	AddFlagReqAttrBid(cmd)
}

// MakeMsgGovCreateMarket reads all the AddFlagsMsgGovCreateMarket flags and creates the desired Msg.
func MakeMsgGovCreateMarket(_ client.Context, flagSet *pflag.FlagSet, _ []string) (sdk.Msg, error) {
	msg := &exchange.MsgGovCreateMarketRequest{}

	errs := make([]error, 14)
	msg.Authority, errs[0] = ReadFlagAuthorityString(flagSet)
	msg.Market.MarketId, errs[1] = ReadFlagMarket(flagSet)
	msg.Market.MarketDetails, errs[2] = ReadFlagsMarketDetails(flagSet)
	msg.Market.FeeCreateAskFlat, errs[3] = ReadFlagCreateAsk(flagSet)
	msg.Market.FeeCreateBidFlat, errs[4] = ReadFlagCreateBid(flagSet)
	msg.Market.FeeSellerSettlementFlat, errs[5] = ReadFlagSellerFlat(flagSet)
	msg.Market.FeeSellerSettlementRatios, errs[6] = ReadFlagSellerRatios(flagSet)
	msg.Market.FeeBuyerSettlementFlat, errs[7] = ReadFlagBuyerFlat(flagSet)
	msg.Market.FeeBuyerSettlementRatios, errs[8] = ReadFlagBuyerRatios(flagSet)
	msg.Market.AcceptingOrders, errs[9] = ReadFlagAcceptingOrders(flagSet)
	msg.Market.AllowUserSettlement, errs[10] = ReadFlagAllowUserSettle(flagSet)
	msg.Market.AccessGrants, errs[11] = ReadFlagAccessGrants(flagSet)
	msg.Market.ReqAttrCreateAsk, errs[12] = ReadFlagReqAttrAsk(flagSet)
	msg.Market.ReqAttrCreateBid, errs[13] = ReadFlagReqAttrBid(flagSet)

	return msg, errors.Join(errs...)
}
