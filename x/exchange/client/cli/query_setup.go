package cli

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/exchange"
)

// SetupCmdQueryOrderFeeCalc adds all the flags needed for MakeQueryOrderFeeCalc.
func SetupCmdQueryOrderFeeCalc(cmd *cobra.Command) {
	cmd.Flags().Bool(FlagAsk, false, "Run calculation on an ask order")
	cmd.Flags().Bool(FlagBid, false, "Run calculation on a bid order")
	cmd.Flags().Uint32(FlagMarket, 0, "The market id (required)")
	cmd.Flags().String(FlagSeller, "", "The seller (for an ask order)")
	cmd.Flags().String(FlagBuyer, "", "The buyer (for a bid order)")
	cmd.Flags().String(FlagAssets, "", "The order assets")
	cmd.Flags().String(FlagPrice, "", "The order price (required)")
	cmd.Flags().String(FlagSettlementFee, "", "The settlement fees")
	cmd.Flags().Bool(FlagPartial, false, "Allow the order to be partially filled")
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

// MakeQueryOrderFeeCalc reads all the SetupCmdQueryOrderFeeCalc flags and creates the desired request.
// Satisfies the queryReqMaker type.
func MakeQueryOrderFeeCalc(_ client.Context, flagSet *pflag.FlagSet, _ []string) (*exchange.QueryOrderFeeCalcRequest, error) {
	bidOrder := &exchange.BidOrder{}

	errs := make([]error, 10, 11)
	var isAsk, isBid bool
	isAsk, errs[0] = flagSet.GetBool(FlagAsk)
	isBid, errs[1] = flagSet.GetBool(FlagBid)
	bidOrder.MarketId, errs[2] = flagSet.GetUint32(FlagMarket)
	var seller string
	seller, errs[3] = flagSet.GetString(FlagSeller)
	bidOrder.Buyer, errs[4] = flagSet.GetString(FlagBuyer)
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

// SetupCmdQueryGetOrder adds all the flags needed for MakeQueryGetOrder.
func SetupCmdQueryGetOrder(cmd *cobra.Command) {
	cmd.Flags().Uint64(FlagOrder, 0, "The order id")

	AddUseArgs(cmd,
		fmt.Sprintf("{<order id>|--%s <order id>}", FlagOrder),
	)
	AddUseDetails(cmd, "An <order id> is required as either an arg or flag, but not both.")
	AddQueryExample(cmd, "8")
	AddQueryExample(cmd, "--"+FlagOrder, "8")

	cmd.Args = cobra.MaximumNArgs(1)
}

// MakeQueryGetOrder reads all the SetupCmdQueryGetOrder flags and creates the desired request.
// Satisfies the queryReqMaker type.
func MakeQueryGetOrder(_ client.Context, flagSet *pflag.FlagSet, args []string) (*exchange.QueryGetOrderRequest, error) {
	req := &exchange.QueryGetOrderRequest{}

	var err error
	req.OrderId, err = ReadFlagOrderOrArg(flagSet, args)

	return req, err
}

// SetupCmdQueryGetOrderByExternalID adds all the flags needed for MakeQueryGetOrderByExternalID.
func SetupCmdQueryGetOrderByExternalID(cmd *cobra.Command) {
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

// MakeQueryGetOrderByExternalID reads all the SetupCmdQueryGetOrderByExternalID flags and creates the desired request.
// Satisfies the queryReqMaker type.
func MakeQueryGetOrderByExternalID(_ client.Context, flagSet *pflag.FlagSet, _ []string) (*exchange.QueryGetOrderByExternalIDRequest, error) {
	req := &exchange.QueryGetOrderByExternalIDRequest{}

	errs := make([]error, 2)
	req.MarketId, errs[0] = flagSet.GetUint32(FlagMarket)
	req.ExternalId, errs[1] = flagSet.GetString(FlagExternalID)

	return req, errors.Join(errs...)
}

// SetupCmdQueryGetMarketOrders adds all the flags needed for MakeQueryGetMarketOrders.
func SetupCmdQueryGetMarketOrders(cmd *cobra.Command) {
	flags.AddPaginationFlagsToCmd(cmd, "orders")

	cmd.Flags().Uint32(FlagMarket, 0, "The market id (required)")
	AddFlagsAsksBidsBools(cmd)
	cmd.Flags().Uint64(FlagAfter, 0, "Limit results to only orders with ids larger than this")

	AddUseArgs(cmd,
		fmt.Sprintf("{<market id>|--%s <market id>}", FlagMarket),
		OptAsksBidsUse,
		OptFlagUse(FlagAfter, "after order id"),
		"[pagination flags]",
	)
	AddUseDetails(cmd,
		"A <market id> is required as either an arg or flag, but not both.",
		OptAsksBidsDesc,
	)
	AddQueryExample(cmd, "3", "--"+FlagAsks)
	AddQueryExample(cmd, "--"+FlagMarket, "1", "--"+FlagAfter, "15", "--"+flags.FlagLimit, "10")

	cmd.Args = cobra.MaximumNArgs(1)
}

// MakeQueryGetMarketOrders reads all the SetupCmdQueryGetMarketOrders flags and creates the desired request.
// Satisfies the queryReqMaker type.
func MakeQueryGetMarketOrders(_ client.Context, flagSet *pflag.FlagSet, args []string) (*exchange.QueryGetMarketOrdersRequest, error) {
	req := &exchange.QueryGetMarketOrdersRequest{}

	errs := make([]error, 4)
	req.MarketId, errs[0] = ReadFlagMarketOrArg(flagSet, args)
	req.OrderType, errs[1] = ReadFlagsAsksBidsOpt(flagSet)
	req.AfterOrderId, errs[2] = flagSet.GetUint64(FlagAfter)
	req.Pagination, errs[3] = client.ReadPageRequestWithPageKeyDecoded(flagSet)

	return req, errors.Join(errs...)
}

// SetupCmdQueryGetOwnerOrders adds all the flags needed for MakeQueryGetOwnerOrders.
func SetupCmdQueryGetOwnerOrders(cmd *cobra.Command) {
	flags.AddPaginationFlagsToCmd(cmd, "orders")

	cmd.Flags().String(FlagOwner, "", "The owner")
	AddFlagsAsksBidsBools(cmd)
	cmd.Flags().Uint64(FlagAfter, 0, "Limit results to only orders with ids larger than this")

	AddUseArgs(cmd,
		fmt.Sprintf("{<owner>|--%s <owner>}", FlagOwner),
		OptAsksBidsUse,
		OptFlagUse(FlagAfter, "after order id"),
		"[pagination flags]",
	)
	AddUseDetails(cmd,
		"An <owner> is required as either an arg or flag, but not both.",
		OptAsksBidsDesc,
	)
	AddQueryExample(cmd, ExampleAddr, "--"+FlagBids)
	AddQueryExample(cmd, "--"+FlagOwner, ExampleAddr, "--"+FlagAsks, "--"+FlagAfter, "15", "--"+flags.FlagLimit, "10")

	cmd.Args = cobra.MaximumNArgs(1)
}

// MakeQueryGetOwnerOrders reads all the SetupCmdQueryGetOwnerOrders flags and creates the desired request.
// Satisfies the queryReqMaker type.
func MakeQueryGetOwnerOrders(_ client.Context, flagSet *pflag.FlagSet, args []string) (*exchange.QueryGetOwnerOrdersRequest, error) {
	req := &exchange.QueryGetOwnerOrdersRequest{}

	errs := make([]error, 5)
	req.Owner, errs[0] = ReadStringFlagOrArg(flagSet, args, FlagOwner, "owner")
	req.OrderType, errs[1] = ReadFlagsAsksBidsOpt(flagSet)
	req.AfterOrderId, errs[2] = flagSet.GetUint64(FlagAfter)
	req.Pagination, errs[3] = client.ReadPageRequestWithPageKeyDecoded(flagSet)

	return req, errors.Join(errs...)
}

// SetupCmdQueryGetAssetOrders adds all the flags needed for MakeQueryGetAssetOrders.
func SetupCmdQueryGetAssetOrders(cmd *cobra.Command) {
	flags.AddPaginationFlagsToCmd(cmd, "orders")

	cmd.Flags().String(FlagDenom, "", "The asset denom")
	AddFlagsAsksBidsBools(cmd)
	cmd.Flags().Uint64(FlagAfter, 0, "Limit results to only orders with ids larger than this")

	AddUseArgs(cmd,
		fmt.Sprintf("{<asset>|--%s <asset>}", FlagDenom),
		OptAsksBidsUse,
		OptFlagUse(FlagAfter, "after order id"),
		"[pagination flags]",
	)
	AddUseDetails(cmd,
		"An <asset> is required as either an arg or flag, but not both.",
		OptAsksBidsDesc,
	)
	AddQueryExample(cmd, "nhash", "--"+FlagAsks)
	AddQueryExample(cmd, "--"+FlagDenom, "nhash", "--"+FlagAfter, "15", "--"+flags.FlagLimit, "10")

	cmd.Args = cobra.MaximumNArgs(1)
}

// MakeQueryGetAssetOrders reads all the SetupCmdQueryGetAssetOrders flags and creates the desired request.
// Satisfies the queryReqMaker type.
func MakeQueryGetAssetOrders(_ client.Context, flagSet *pflag.FlagSet, args []string) (*exchange.QueryGetAssetOrdersRequest, error) {
	req := &exchange.QueryGetAssetOrdersRequest{}

	errs := make([]error, 4)
	req.Asset, errs[0] = ReadStringFlagOrArg(flagSet, args, FlagDenom, "asset")
	req.OrderType, errs[1] = ReadFlagsAsksBidsOpt(flagSet)
	req.AfterOrderId, errs[2] = flagSet.GetUint64(FlagAfter)
	req.Pagination, errs[3] = client.ReadPageRequestWithPageKeyDecoded(flagSet)

	return req, errors.Join(errs...)
}

// SetupCmdQueryGetAllOrders adds all the flags needed for MakeQueryGetAllOrders.
func SetupCmdQueryGetAllOrders(cmd *cobra.Command) {
	flags.AddPaginationFlagsToCmd(cmd, "orders")

	AddUseArgs(cmd, "[pagination flags]")
	AddUseDetails(cmd)
	AddQueryExample(cmd, "--"+flags.FlagLimit, "10")
	AddQueryExample(cmd, "--"+flags.FlagReverse)

	cmd.Args = cobra.NoArgs
}

// MakeQueryGetAllOrders reads all the SetupCmdQueryGetAllOrders flags and creates the desired request.
// Satisfies the queryReqMaker type.
func MakeQueryGetAllOrders(_ client.Context, flagSet *pflag.FlagSet, _ []string) (*exchange.QueryGetAllOrdersRequest, error) {
	req := &exchange.QueryGetAllOrdersRequest{}

	var err error
	req.Pagination, err = client.ReadPageRequestWithPageKeyDecoded(flagSet)

	return req, err
}

// SetupCmdQueryGetMarket adds all the flags needed for MakeQueryGetMarket.
func SetupCmdQueryGetMarket(cmd *cobra.Command) {
	cmd.Flags().Uint32(FlagMarket, 0, "The market id")

	AddUseArgs(cmd,
		fmt.Sprintf("{<market id>|--%s <market id>}", FlagMarket),
	)
	AddUseDetails(cmd, "A <market id> is required as either an arg or flag, but not both.")
	AddQueryExample(cmd, "3")
	AddQueryExample(cmd, "--"+FlagMarket, "1")

	cmd.Args = cobra.MaximumNArgs(1)
}

// MakeQueryGetMarket reads all the SetupCmdQueryGetMarket flags and creates the desired request.
// Satisfies the queryReqMaker type.
func MakeQueryGetMarket(_ client.Context, flagSet *pflag.FlagSet, args []string) (*exchange.QueryGetMarketRequest, error) {
	req := &exchange.QueryGetMarketRequest{}

	var err error
	req.MarketId, err = ReadFlagMarketOrArg(flagSet, args)

	return req, err
}

// SetupCmdQueryGetAllMarkets adds all the flags needed for MakeQueryGetAllMarkets.
func SetupCmdQueryGetAllMarkets(cmd *cobra.Command) {
	flags.AddPaginationFlagsToCmd(cmd, "markets")

	AddUseArgs(cmd, "[pagination flags]")
	AddUseDetails(cmd)
	AddQueryExample(cmd, "--"+flags.FlagLimit, "10")
	AddQueryExample(cmd, "--"+flags.FlagReverse)

	cmd.Args = cobra.NoArgs
}

// MakeQueryGetAllMarkets reads all the SetupCmdQueryGetAllMarkets flags and creates the desired request.
// Satisfies the queryReqMaker type.
func MakeQueryGetAllMarkets(_ client.Context, flagSet *pflag.FlagSet, _ []string) (*exchange.QueryGetAllMarketsRequest, error) {
	req := &exchange.QueryGetAllMarketsRequest{}

	var err error
	req.Pagination, err = client.ReadPageRequestWithPageKeyDecoded(flagSet)

	return req, err
}

// SetupCmdQueryParams adds all the flags needed for MakeQueryParams.
func SetupCmdQueryParams(cmd *cobra.Command) {
	AddUseDetails(cmd)
	AddQueryExample(cmd)

	cmd.Args = cobra.NoArgs
}

// MakeQueryParams reads all the SetupCmdQueryParams flags and creates the desired request.
// Satisfies the queryReqMaker type.
func MakeQueryParams(_ client.Context, _ *pflag.FlagSet, _ []string) (*exchange.QueryParamsRequest, error) {
	return &exchange.QueryParamsRequest{}, nil
}

// SetupCmdQueryValidateCreateMarket adds all the flags needed for MakeQueryValidateCreateMarket.
func SetupCmdQueryValidateCreateMarket(cmd *cobra.Command) {
	SetupCmdTxGovCreateMarket(cmd)
}

// MakeQueryValidateCreateMarket reads all the SetupCmdQueryValidateCreateMarket flags and creates the desired request.
// Satisfies the queryReqMaker type.
func MakeQueryValidateCreateMarket(clientCtx client.Context, flags *pflag.FlagSet, args []string) (*exchange.QueryValidateCreateMarketRequest, error) {
	req := &exchange.QueryValidateCreateMarketRequest{}

	var err error
	req.CreateMarketRequest, err = MakeMsgGovCreateMarket(clientCtx, flags, args)

	return req, err
}

// SetupCmdQueryValidateMarket adds all the flags needed for MakeQueryValidateMarket.
func SetupCmdQueryValidateMarket(cmd *cobra.Command) {
	cmd.Flags().Uint32(FlagMarket, 0, "The market id")

	AddUseArgs(cmd,
		fmt.Sprintf("{<market id>|--%s <market id>}", FlagMarket),
	)
	AddUseDetails(cmd, "A <market id> is required as either an arg or flag, but not both.")
	AddQueryExample(cmd, "3")
	AddQueryExample(cmd, "--"+FlagMarket, "1")

	cmd.Args = cobra.MaximumNArgs(1)
}

// MakeQueryValidateMarket reads all the SetupCmdQueryValidateMarket flags and creates the desired request.
// Satisfies the queryReqMaker type.
func MakeQueryValidateMarket(_ client.Context, flagSet *pflag.FlagSet, args []string) (*exchange.QueryValidateMarketRequest, error) {
	req := &exchange.QueryValidateMarketRequest{}

	var err error
	req.MarketId, err = ReadFlagMarketOrArg(flagSet, args)

	return req, err
}

// SetupCmdQueryValidateManageFees adds all the flags needed for MakeQueryValidateManageFees.
func SetupCmdQueryValidateManageFees(cmd *cobra.Command) {
	SetupCmdTxGovManageFees(cmd)
}

// MakeQueryValidateManageFees reads all the SetupCmdQueryValidateManageFees flags and creates the desired request.
// Satisfies the queryReqMaker type.
func MakeQueryValidateManageFees(clientCtx client.Context, flags *pflag.FlagSet, args []string) (*exchange.QueryValidateManageFeesRequest, error) {
	req := &exchange.QueryValidateManageFeesRequest{}

	var err error
	req.ManageFeesRequest, err = MakeMsgGovManageFees(clientCtx, flags, args)

	return req, err
}
