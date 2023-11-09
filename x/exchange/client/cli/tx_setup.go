package cli

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"

	"github.com/provenance-io/provenance/x/exchange"
)

// SetupCmdTxCreateAsk adds all the flags needed for MakeMsgCreateAsk.
func SetupCmdTxCreateAsk(cmd *cobra.Command) {
	flags.AddTxFlagsToCmd(cmd)

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

// MakeMsgCreateAsk reads all the SetupCmdTxCreateAsk flags and creates the desired Msg.
// Satisfies the msgMaker type.
func MakeMsgCreateAsk(clientCtx client.Context, flagSet *pflag.FlagSet, _ []string) (*exchange.MsgCreateAskRequest, error) {
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

// SetupCmdTxCreateBid adds all the flags needed for MakeMsgCreateBid.
func SetupCmdTxCreateBid(cmd *cobra.Command) {
	flags.AddTxFlagsToCmd(cmd)

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

// MakeMsgCreateBid reads all the SetupCmdTxCreateBid flags and creates the desired Msg.
// Satisfies the msgMaker type.
func MakeMsgCreateBid(clientCtx client.Context, flagSet *pflag.FlagSet, _ []string) (*exchange.MsgCreateBidRequest, error) {
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

// SetupCmdTxCancelOrder adds all the flags needed for the MakeMsgCancelOrder.
func SetupCmdTxCancelOrder(cmd *cobra.Command) {
	flags.AddTxFlagsToCmd(cmd)

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

// MakeMsgCancelOrder reads all the SetupCmdTxCancelOrder flags and the provided args and creates the desired Msg.
// Satisfies the msgMaker type.
func MakeMsgCancelOrder(clientCtx client.Context, flagSet *pflag.FlagSet, args []string) (*exchange.MsgCancelOrderRequest, error) {
	msg := &exchange.MsgCancelOrderRequest{}

	errs := make([]error, 2)
	msg.Signer, errs[0] = ReadAddrFlagOrFrom(clientCtx, flagSet, FlagSigner)
	msg.OrderId, errs[1] = ReadFlagOrderOrArg(flagSet, args)

	return msg, errors.Join(errs...)
}

// SetupCmdTxFillBids adds all the flags needed for MakeMsgFillBids.
func SetupCmdTxFillBids(cmd *cobra.Command) {
	flags.AddTxFlagsToCmd(cmd)

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

// MakeMsgFillBids reads all the SetupCmdTxFillBids flags and creates the desired Msg.
// Satisfies the msgMaker type.
func MakeMsgFillBids(clientCtx client.Context, flagSet *pflag.FlagSet, _ []string) (*exchange.MsgFillBidsRequest, error) {
	msg := &exchange.MsgFillBidsRequest{}

	errs := make([]error, 6)
	msg.Seller, errs[0] = ReadAddrFlagOrFrom(clientCtx, flagSet, FlagSeller)
	msg.MarketId, errs[1] = flagSet.GetUint32(FlagMarket)
	msg.TotalAssets, errs[2] = ReadCoinsFlag(flagSet, FlagAssets)
	msg.BidOrderIds, errs[3] = ReadOrderIDsFlag(flagSet, FlagBids)
	msg.SellerSettlementFlatFee, errs[4] = ReadCoinFlag(flagSet, FlagSettlementFee)
	msg.AskOrderCreationFee, errs[5] = ReadCoinFlag(flagSet, FlagCreationFee)

	return msg, errors.Join(errs...)
}

// SetupCmdTxFillAsks adds all the flags needed for MakeMsgFillAsks.
func SetupCmdTxFillAsks(cmd *cobra.Command) {
	flags.AddTxFlagsToCmd(cmd)

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

// MakeMsgFillAsks reads all the SetupCmdTxFillAsks flags and creates the desired Msg.
// Satisfies the msgMaker type.
func MakeMsgFillAsks(clientCtx client.Context, flagSet *pflag.FlagSet, _ []string) (*exchange.MsgFillAsksRequest, error) {
	msg := &exchange.MsgFillAsksRequest{}

	errs := make([]error, 6)
	msg.Buyer, errs[0] = ReadAddrFlagOrFrom(clientCtx, flagSet, FlagBuyer)
	msg.MarketId, errs[1] = flagSet.GetUint32(FlagMarket)
	msg.TotalPrice, errs[2] = ReadReqCoinFlag(flagSet, FlagPrice)
	msg.AskOrderIds, errs[3] = ReadOrderIDsFlag(flagSet, FlagAsks)
	msg.BuyerSettlementFees, errs[4] = ReadCoinsFlag(flagSet, FlagSettlementFee)
	msg.BidOrderCreationFee, errs[5] = ReadCoinFlag(flagSet, FlagCreationFee)

	return msg, errors.Join(errs...)
}

// SetupCmdTxMarketSettle adds all the flags needed for MakeMsgMarketSettle.
func SetupCmdTxMarketSettle(cmd *cobra.Command) {
	flags.AddTxFlagsToCmd(cmd)

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

// MakeMsgMarketSettle reads all the SetupCmdTxMarketSettle flags and creates the desired Msg.
// Satisfies the msgMaker type.
func MakeMsgMarketSettle(clientCtx client.Context, flagSet *pflag.FlagSet, _ []string) (*exchange.MsgMarketSettleRequest, error) {
	msg := &exchange.MsgMarketSettleRequest{}

	errs := make([]error, 5)
	msg.Admin, errs[0] = ReadFlagsAdminOrFrom(clientCtx, flagSet)
	msg.MarketId, errs[1] = flagSet.GetUint32(FlagMarket)
	msg.AskOrderIds, errs[2] = ReadOrderIDsFlag(flagSet, FlagAsks)
	msg.BidOrderIds, errs[3] = ReadOrderIDsFlag(flagSet, FlagBids)
	msg.ExpectPartial, errs[4] = flagSet.GetBool(FlagPartial)

	return msg, errors.Join(errs...)
}

// SetupCmdTxMarketSetOrderExternalID adds all the flags needed for MakeMsgMarketSetOrderExternalID.
func SetupCmdTxMarketSetOrderExternalID(cmd *cobra.Command) {
	flags.AddTxFlagsToCmd(cmd)

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

// MakeMsgMarketSetOrderExternalID reads all the SetupCmdTxMarketSetOrderExternalID flags and creates the desired Msg.
// Satisfies the msgMaker type.
func MakeMsgMarketSetOrderExternalID(clientCtx client.Context, flagSet *pflag.FlagSet, _ []string) (*exchange.MsgMarketSetOrderExternalIDRequest, error) {
	msg := &exchange.MsgMarketSetOrderExternalIDRequest{}

	errs := make([]error, 4)
	msg.Admin, errs[0] = ReadFlagsAdminOrFrom(clientCtx, flagSet)
	msg.MarketId, errs[1] = flagSet.GetUint32(FlagMarket)
	msg.OrderId, errs[2] = flagSet.GetUint64(FlagOrder)
	msg.ExternalId, errs[4] = flagSet.GetString(FlagExternalID)

	return msg, errors.Join(errs...)
}

// SetupCmdTxMarketWithdraw adds all the flags needed for MakeMsgMarketWithdraw.
func SetupCmdTxMarketWithdraw(cmd *cobra.Command) {
	flags.AddTxFlagsToCmd(cmd)

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

// MakeMsgMarketWithdraw reads all the SetupCmdTxMarketWithdraw flags and creates the desired Msg.
// Satisfies the msgMaker type.
func MakeMsgMarketWithdraw(clientCtx client.Context, flagSet *pflag.FlagSet, _ []string) (*exchange.MsgMarketWithdrawRequest, error) {
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

// SetupCmdTxMarketUpdateDetails adds all the flags needed for MakeMsgMarketUpdateDetails.
func SetupCmdTxMarketUpdateDetails(cmd *cobra.Command) {
	flags.AddTxFlagsToCmd(cmd)

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

// MakeMsgMarketUpdateDetails reads all the SetupCmdTxMarketUpdateDetails flags and creates the desired Msg.
// Satisfies the msgMaker type.
func MakeMsgMarketUpdateDetails(clientCtx client.Context, flagSet *pflag.FlagSet, _ []string) (*exchange.MsgMarketUpdateDetailsRequest, error) {
	msg := &exchange.MsgMarketUpdateDetailsRequest{}

	errs := make([]error, 3)
	msg.Admin, errs[0] = ReadFlagsAdminOrFrom(clientCtx, flagSet)
	msg.MarketId, errs[1] = flagSet.GetUint32(FlagMarket)
	msg.MarketDetails, errs[2] = ReadFlagsMarketDetails(flagSet)

	return msg, errors.Join(errs...)
}

// SetupCmdTxMarketUpdateEnabled adds all the flags needed for MakeMsgMarketUpdateEnabled.
func SetupCmdTxMarketUpdateEnabled(cmd *cobra.Command) {
	flags.AddTxFlagsToCmd(cmd)

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

// MakeMsgMarketUpdateEnabled reads all the SetupCmdTxMarketUpdateEnabled flags and creates the desired Msg.
// Satisfies the msgMaker type.
func MakeMsgMarketUpdateEnabled(clientCtx client.Context, flagSet *pflag.FlagSet, _ []string) (*exchange.MsgMarketUpdateEnabledRequest, error) {
	msg := &exchange.MsgMarketUpdateEnabledRequest{}

	errs := make([]error, 3)
	msg.Admin, errs[0] = ReadFlagsAdminOrFrom(clientCtx, flagSet)
	msg.MarketId, errs[1] = flagSet.GetUint32(FlagMarket)
	msg.AcceptingOrders, errs[2] = ReadFlagsEnableDisable(flagSet)

	return msg, errors.Join(errs...)
}

// SetupCmdTxMarketUpdateUserSettle adds all the flags needed for MakeMsgMarketUpdateUserSettle.
func SetupCmdTxMarketUpdateUserSettle(cmd *cobra.Command) {
	flags.AddTxFlagsToCmd(cmd)

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

// MakeMsgMarketUpdateUserSettle reads all the SetupCmdTxMarketUpdateUserSettle flags and creates the desired Msg.
// Satisfies the msgMaker type.
func MakeMsgMarketUpdateUserSettle(clientCtx client.Context, flagSet *pflag.FlagSet, _ []string) (*exchange.MsgMarketUpdateUserSettleRequest, error) {
	msg := &exchange.MsgMarketUpdateUserSettleRequest{}

	errs := make([]error, 3)
	msg.Admin, errs[0] = ReadFlagsAdminOrFrom(clientCtx, flagSet)
	msg.MarketId, errs[1] = flagSet.GetUint32(FlagMarket)
	msg.AllowUserSettlement, errs[2] = ReadFlagsEnableDisable(flagSet)

	return msg, errors.Join(errs...)
}

// SetupCmdTxMarketManagePermissions adds all the flags needed for MakeMsgMarketManagePermissions.
func SetupCmdTxMarketManagePermissions(cmd *cobra.Command) {
	flags.AddTxFlagsToCmd(cmd)

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

// MakeMsgMarketManagePermissions reads all the SetupCmdTxMarketManagePermissions flags and creates the desired Msg.
// Satisfies the msgMaker type.
func MakeMsgMarketManagePermissions(clientCtx client.Context, flagSet *pflag.FlagSet, _ []string) (*exchange.MsgMarketManagePermissionsRequest, error) {
	msg := &exchange.MsgMarketManagePermissionsRequest{}

	errs := make([]error, 5)
	msg.Admin, errs[0] = ReadFlagsAdminOrFrom(clientCtx, flagSet)
	msg.MarketId, errs[1] = flagSet.GetUint32(FlagMarket)
	msg.RevokeAll, errs[2] = flagSet.GetStringSlice(FlagRevokeAll)
	msg.ToRevoke, errs[3] = ReadAccessGrantsFlag(flagSet, FlagRevoke)
	msg.ToGrant, errs[4] = ReadAccessGrantsFlag(flagSet, FlagGrant)

	return msg, errors.Join(errs...)
}

// SetupCmdTxMarketManageReqAttrs adds all the flags needed for MakeMsgMarketManageReqAttrs.
func SetupCmdTxMarketManageReqAttrs(cmd *cobra.Command) {
	flags.AddTxFlagsToCmd(cmd)

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

// MakeMsgMarketManageReqAttrs reads all the SetupCmdTxMarketManageReqAttrs flags and creates the desired Msg.
// Satisfies the msgMaker type.
func MakeMsgMarketManageReqAttrs(clientCtx client.Context, flagSet *pflag.FlagSet, _ []string) (*exchange.MsgMarketManageReqAttrsRequest, error) {
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

// SetupCmdTxGovCreateMarket adds all the flags needed for MakeMsgGovCreateMarket.
func SetupCmdTxGovCreateMarket(cmd *cobra.Command) {
	flags.AddTxFlagsToCmd(cmd)
	govcli.AddGovPropFlagsToCmd(cmd)

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

// MakeMsgGovCreateMarket reads all the SetupCmdTxGovCreateMarket flags and creates the desired Msg.
// Satisfies the msgMaker type.
func MakeMsgGovCreateMarket(_ client.Context, flagSet *pflag.FlagSet, _ []string) (*exchange.MsgGovCreateMarketRequest, error) {
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

// SetupCmdTxGovManageFees adds all the flags needed for MakeMsgGovManageFees.
func SetupCmdTxGovManageFees(cmd *cobra.Command) {
	flags.AddTxFlagsToCmd(cmd)
	govcli.AddGovPropFlagsToCmd(cmd)

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

// MakeMsgGovManageFees reads all the SetupCmdTxGovManageFees flags and creates the desired Msg.
// Satisfies the msgMaker type.
func MakeMsgGovManageFees(_ client.Context, flagSet *pflag.FlagSet, _ []string) (*exchange.MsgGovManageFeesRequest, error) {
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

// SetupCmdTxGovUpdateParams adds all the flags needed for MakeMsgGovUpdateParams.
func SetupCmdTxGovUpdateParams(cmd *cobra.Command) {
	flags.AddTxFlagsToCmd(cmd)
	govcli.AddGovPropFlagsToCmd(cmd)

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

// MakeMsgGovUpdateParams reads all the SetupCmdTxGovUpdateParams flags and creates the desired Msg.
// Satisfies the msgMaker type.
func MakeMsgGovUpdateParams(_ client.Context, flagSet *pflag.FlagSet, _ []string) (*exchange.MsgGovUpdateParamsRequest, error) {
	msg := &exchange.MsgGovUpdateParamsRequest{}

	errs := make([]error, 3)
	msg.Authority, errs[0] = ReadFlagAuthority(flagSet)
	msg.Params.DefaultSplit, errs[1] = flagSet.GetUint32(FlagDefault)
	msg.Params.DenomSplits, errs[2] = ReadSplitsFlag(flagSet, FlagSplit)

	return msg, errors.Join(errs...)
}
