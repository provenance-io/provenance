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
		PageFlagsUse,
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
		PageFlagsUse,
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
		PageFlagsUse,
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

	AddUseArgs(cmd, PageFlagsUse)
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

// SetupCmdQueryGetCommitment adds all the flags needed for MakeQueryGetCommitment.
func SetupCmdQueryGetCommitment(cmd *cobra.Command) {
	cmd.Flags().String(FlagAccount, "", "The account's address")
	cmd.Flags().Uint32(FlagMarket, 0, "The market id")

	MarkFlagsRequired(cmd, FlagAccount, FlagMarket)

	AddUseArgs(cmd,
		ReqFlagUse(FlagAccount, "account"),
		ReqFlagUse(FlagMarket, "market id"),
	)
	AddUseDetails(cmd)
	AddQueryExample(cmd, "--"+FlagAccount, ExampleAddr, "--"+FlagMarket, "3")

	cmd.Args = cobra.NoArgs
}

// MakeQueryGetCommitment reads all the SetupCmdQueryGetCommitment flags and creates the desired request.
// Satisfies the queryReqMaker type.
func MakeQueryGetCommitment(_ client.Context, flagSet *pflag.FlagSet, _ []string) (*exchange.QueryGetCommitmentRequest, error) {
	rv := &exchange.QueryGetCommitmentRequest{}

	errs := make([]error, 2)
	rv.Account, errs[0] = flagSet.GetString(FlagAccount)
	rv.MarketId, errs[1] = flagSet.GetUint32(FlagMarket)

	return rv, errors.Join(errs...)
}

// SetupCmdQueryGetAccountCommitments adds all the flags needed for MakeQueryGetAccountCommitments.
func SetupCmdQueryGetAccountCommitments(cmd *cobra.Command) {
	cmd.Flags().String(FlagAccount, "", "The account's address")

	AddUseArgs(cmd,
		fmt.Sprintf("{<account>|--%s <account>}", FlagAccount),
	)
	AddUseDetails(cmd,
		"An <account> is required as either an arg or flag, but not both.",
	)
	AddQueryExample(cmd, ExampleAddr)
	AddQueryExample(cmd, "--"+FlagAccount, ExampleAddr)

	cmd.Args = cobra.MaximumNArgs(1)
}

// MakeQueryGetAccountCommitments reads all the SetupCmdQueryGetAccountCommitments flags and creates the desired request.
// Satisfies the queryReqMaker type.
func MakeQueryGetAccountCommitments(_ client.Context, flagSet *pflag.FlagSet, args []string) (*exchange.QueryGetAccountCommitmentsRequest, error) {
	rv := &exchange.QueryGetAccountCommitmentsRequest{}

	var err error
	rv.Account, err = ReadStringFlagOrArg(flagSet, args, FlagAccount, "account")

	return rv, err
}

// SetupCmdQueryGetMarketCommitments adds all the flags needed for MakeQueryGetMarketCommitments.
func SetupCmdQueryGetMarketCommitments(cmd *cobra.Command) {
	flags.AddPaginationFlagsToCmd(cmd, "commitments")
	cmd.Flags().Uint32(FlagMarket, 0, "The market id")

	AddUseArgs(cmd,
		fmt.Sprintf("{<market id>|--%s <market id>}", FlagMarket),
		PageFlagsUse,
	)
	AddUseDetails(cmd, "A <market id> is required as either an arg or flag, but not both.")
	AddQueryExample(cmd, "3")
	AddQueryExample(cmd, "--"+FlagMarket, "1", "--limit", "10")

	cmd.Args = cobra.MaximumNArgs(1)
}

// MakeQueryGetMarketCommitments reads all the SetupCmdQueryGetMarketCommitments flags and creates the desired request.
// Satisfies the queryReqMaker type.
func MakeQueryGetMarketCommitments(_ client.Context, flagSet *pflag.FlagSet, args []string) (*exchange.QueryGetMarketCommitmentsRequest, error) {
	rv := &exchange.QueryGetMarketCommitmentsRequest{}

	errs := make([]error, 2)
	rv.MarketId, errs[0] = ReadFlagMarketOrArg(flagSet, args)
	rv.Pagination, errs[1] = client.ReadPageRequestWithPageKeyDecoded(flagSet)

	return rv, errors.Join(errs...)
}

// SetupCmdQueryGetAllCommitments adds all the flags needed for MakeQueryGetAllCommitments.
func SetupCmdQueryGetAllCommitments(cmd *cobra.Command) {
	flags.AddPaginationFlagsToCmd(cmd, "commitments")

	AddUseArgs(cmd, PageFlagsUse)
	AddUseDetails(cmd)
	AddQueryExample(cmd, "--"+flags.FlagLimit, "10")
	AddQueryExample(cmd, "--"+flags.FlagReverse)

	cmd.Args = cobra.NoArgs
}

// MakeQueryGetAllCommitments reads all the SetupCmdQueryGetAllCommitments flags and creates the desired request.
// Satisfies the queryReqMaker type.
func MakeQueryGetAllCommitments(_ client.Context, flagSet *pflag.FlagSet, _ []string) (*exchange.QueryGetAllCommitmentsRequest, error) {
	req := &exchange.QueryGetAllCommitmentsRequest{}

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

	AddUseArgs(cmd, PageFlagsUse)
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

// SetupCmdQueryCommitmentSettlementFeeCalc adds all the flags needed for MakeQueryCommitmentSettlementFeeCalc.
func SetupCmdQueryCommitmentSettlementFeeCalc(cmd *cobra.Command) {
	cmd.Flags().Bool(FlagDetails, false, "Include breakdown fields")
	cmd.Flags().String(flags.FlagFrom, "", "The from address")
	AddUseArgs(cmd, OptFlagUse(FlagDetails, ""))
	SetupCmdTxMarketCommitmentSettle(cmd)
}

// MakeQueryCommitmentSettlementFeeCalc reads all the SetupCmdQueryCommitmentSettlementFeeCalc flags and creates the desired request.
// Satisfies the queryReqMaker type.
func MakeQueryCommitmentSettlementFeeCalc(clientCtx client.Context, flagSet *pflag.FlagSet, args []string) (*exchange.QueryCommitmentSettlementFeeCalcRequest, error) {
	rv := &exchange.QueryCommitmentSettlementFeeCalcRequest{}

	errs := make([]error, 4)
	clientCtx.From, errs[0] = flagSet.GetString(flags.FlagFrom)
	if len(clientCtx.From) > 0 {
		if addr, err := sdk.AccAddressFromBech32(clientCtx.From); err == nil {
			clientCtx.FromAddress = addr
		} else {
			clientCtx.FromAddress, clientCtx.From, _, errs[1] = client.GetFromFields(clientCtx, clientCtx.Keyring, clientCtx.From)
		}
	}
	rv.Settlement, errs[2] = MakeMsgMarketCommitmentSettle(clientCtx, flagSet, args)
	rv.IncludeBreakdownFields, errs[3] = flagSet.GetBool(FlagDetails)

	return rv, errors.Join(errs...)
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

// SetupCmdQueryGetPayment adds all the flags needed for MakeQueryGetPayment.
func SetupCmdQueryGetPayment(cmd *cobra.Command) {
	cmd.Flags().String(FlagSource, "", "The payment's source account")
	cmd.Flags().String(FlagExternalID, "", "The payment's external id")

	AddUseArgs(cmd,
		fmt.Sprintf("{<source>|--%s <source>}", FlagSource),
		fmt.Sprintf("[<external id>|--%s <external id>]", FlagExternalID),
	)
	AddUseDetails(cmd,
		"A <source> is required as either the first arg or a flag, but not both.",
		"The <external id> can be provided as either the second arg or a flag, but not both.",
	)
	AddQueryExample(cmd, ExampleAddr, "myid")
	AddQueryExample(cmd, ExampleAddr, "--"+FlagExternalID, "myid")
	AddQueryExample(cmd, "--"+FlagSource, ExampleAddr, "--"+FlagExternalID, "myid")

	cmd.Args = cobra.MaximumNArgs(2)
}

// MakeQueryGetPayment reads all the SetupCmdQueryGetPayment flags and creates the desired request.
// Satisfies the queryReqMaker type.
func MakeQueryGetPayment(_ client.Context, flagSet *pflag.FlagSet, args []string) (*exchange.QueryGetPaymentRequest, error) {
	req := &exchange.QueryGetPaymentRequest{}

	errs := make([]error, 2)
	req.Source, errs[0] = ReadStringFlagOrArg(flagSet, args, FlagSource, "source")
	if len(args) > 0 {
		args = args[1:]
	}
	req.ExternalId, errs[1] = ReadOptStringFlagOrArg(flagSet, args, FlagExternalID, "external id")

	return req, errors.Join(errs...)
}

// SetupCmdQueryGetPaymentsWithSource adds all the flags needed for MakeQueryGetPaymentsWithSource.
func SetupCmdQueryGetPaymentsWithSource(cmd *cobra.Command) {
	flags.AddPaginationFlagsToCmd(cmd, "payments")
	cmd.Flags().String(FlagSource, "", "The source account of the payments")

	AddUseArgs(cmd,
		fmt.Sprintf("{<source>|--%s <source>}", FlagSource),
		PageFlagsUse,
	)
	AddUseDetails(cmd,
		"A <source> is required as either an arg or a flag, but not both.",
	)
	AddQueryExample(cmd, ExampleAddr)
	AddQueryExample(cmd, "--"+FlagSource, ExampleAddr)

	cmd.Args = cobra.MaximumNArgs(1)
}

// MakeQueryGetPaymentsWithSource reads all the SetupCmdQueryGetPaymentsWithSource flags and creates the desired request.
// Satisfies the queryReqMaker type.
func MakeQueryGetPaymentsWithSource(_ client.Context, flagSet *pflag.FlagSet, args []string) (*exchange.QueryGetPaymentsWithSourceRequest, error) {
	req := &exchange.QueryGetPaymentsWithSourceRequest{}

	errs := make([]error, 2)
	req.Source, errs[0] = ReadStringFlagOrArg(flagSet, args, FlagSource, "source")
	req.Pagination, errs[1] = client.ReadPageRequestWithPageKeyDecoded(flagSet)

	return req, errors.Join(errs...)
}

// SetupCmdQueryGetPaymentsWithTarget adds all the flags needed for MakeQueryGetPaymentsWithTarget.
func SetupCmdQueryGetPaymentsWithTarget(cmd *cobra.Command) {
	flags.AddPaginationFlagsToCmd(cmd, "payments")
	cmd.Flags().String(FlagTarget, "", "The target account of the payments")

	AddUseArgs(cmd,
		fmt.Sprintf("{<target>|--%s <target>}", FlagTarget),
		PageFlagsUse,
	)
	AddUseDetails(cmd,
		"A <target> is required as either an arg or a flag, but not both.",
	)
	AddQueryExample(cmd, ExampleAddr)
	AddQueryExample(cmd, "--"+FlagTarget, ExampleAddr)

	cmd.Args = cobra.MaximumNArgs(1)
}

// MakeQueryGetPaymentsWithTarget reads all the SetupCmdQueryGetPaymentsWithTarget flags and creates the desired request.
// Satisfies the queryReqMaker type.
func MakeQueryGetPaymentsWithTarget(_ client.Context, flagSet *pflag.FlagSet, args []string) (*exchange.QueryGetPaymentsWithTargetRequest, error) {
	req := &exchange.QueryGetPaymentsWithTargetRequest{}

	errs := make([]error, 2)
	req.Target, errs[0] = ReadStringFlagOrArg(flagSet, args, FlagTarget, "target")
	req.Pagination, errs[1] = client.ReadPageRequestWithPageKeyDecoded(flagSet)

	return req, errors.Join(errs...)
}

// SetupCmdQueryGetAllPayments adds all the flags needed for MakeQueryGetAllPayments.
func SetupCmdQueryGetAllPayments(cmd *cobra.Command) {
	flags.AddPaginationFlagsToCmd(cmd, "payments")

	AddUseArgs(cmd, PageFlagsUse)
	AddUseDetails(cmd)
	AddQueryExample(cmd, "--"+flags.FlagLimit, "10")
	AddQueryExample(cmd, "--"+flags.FlagReverse)

	cmd.Args = cobra.NoArgs
}

// MakeQueryGetAllPayments reads all the SetupCmdQueryGetAllPayments flags and creates the desired request.
// Satisfies the queryReqMaker type.
func MakeQueryGetAllPayments(_ client.Context, flagSet *pflag.FlagSet, _ []string) (*exchange.QueryGetAllPaymentsRequest, error) {
	req := &exchange.QueryGetAllPaymentsRequest{}

	var err error
	req.Pagination, err = client.ReadPageRequestWithPageKeyDecoded(flagSet)

	return req, err
}

// SetupCmdQueryPaymentFeeCalc adds all the flags needed for MakeQueryPaymentFeeCalc.
func SetupCmdQueryPaymentFeeCalc(cmd *cobra.Command) {
	cmd.Flags().String(FlagSource, "", "The source account")
	cmd.Flags().String(FlagSourceAmount, "", "The source funds, e.g. 10nhash")
	cmd.Flags().String(FlagTarget, "", "The target account")
	cmd.Flags().String(FlagTargetAmount, "", "The target funds, e.g. 10nhash")
	cmd.Flags().String(FlagExternalID, "", "The external id")
	cmd.Flags().String(FlagFile, "", "a json file of a Tx with a MsgCreatePaymentRequest or MsgAcceptPaymentRequest")

	AddUseArgs(cmd,
		OptFlagUse(FlagSource, "source"),
		OptFlagUse(FlagSourceAmount, "source amount"),
		UseFlagsBreak,
		OptFlagUse(FlagTarget, "target"),
		OptFlagUse(FlagTargetAmount, "target amount"),
		OptFlagUse(FlagExternalID, "external id"),
		UseFlagsBreak,
		OptFlagUse(FlagFile, "filename"),
	)
	AddUseDetails(cmd,
		MsgFileDesc(&exchange.MsgCreatePaymentRequest{}),
		fmt.Sprintf("Alternatively, the file can have a %s in it.", sdk.MsgTypeURL(&exchange.MsgAcceptPaymentRequest{})),
	)

	cmd.Args = cobra.NoArgs
}

// MakeQueryPaymentFeeCalc reads all the SetupCmdQueryPaymentFeeCalc flags and creates the desired request.
// Satisfies the queryReqMaker type.
func MakeQueryPaymentFeeCalc(clientCtx client.Context, flagSet *pflag.FlagSet, _ []string) (*exchange.QueryPaymentFeeCalcRequest, error) {
	req := &exchange.QueryPaymentFeeCalcRequest{}

	errs := make([]error, 6)
	req.Payment, errs[0] = ReadPaymentFromFileFlag(clientCtx, flagSet)
	req.Payment.Source, errs[1] = ReadFlagStringOrDefault(flagSet, FlagSource, req.Payment.Source)
	req.Payment.SourceAmount, errs[2] = ReadCoinsFlagOrDefault(flagSet, FlagSourceAmount, req.Payment.SourceAmount)
	req.Payment.Target, errs[3] = ReadFlagStringOrDefault(flagSet, FlagTarget, req.Payment.Target)
	req.Payment.TargetAmount, errs[4] = ReadCoinsFlagOrDefault(flagSet, FlagTargetAmount, req.Payment.TargetAmount)
	req.Payment.ExternalId, errs[5] = ReadFlagStringOrDefault(flagSet, FlagExternalID, req.Payment.ExternalId)

	return req, errors.Join(errs...)
}
