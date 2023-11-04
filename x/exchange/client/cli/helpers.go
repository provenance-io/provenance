package cli

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/exchange"
)

// A msgMaker is a function that makes a Msg from a client.Context, FlagSet, and set of args.
type msgMaker func(clientCtx client.Context, flagSet *pflag.FlagSet, args []string) (sdk.Msg, error)

var (
	_ msgMaker = MakeMsgCreateAskRequest
	_ msgMaker = MakeMsgCreateBidRequest
	_ msgMaker = MakeMsgCancelOrderRequest
	_ msgMaker = MakeMsgFillBidsRequest
	_ msgMaker = MakeMsgFillAsksRequest
	_ msgMaker = MakeMsgMarketSettleRequest
)

// genericTxRunE returns a cobra.Command.RunE function that gets the client.Context, and FlagSet,
// then uses the provided maker to make the msg that it then generates or broadcasts.
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

// AddCreateAskFlags adds all the flags needed for the create-ask command.
func AddCreateAskFlags(cmd *cobra.Command) {
	AddFlagSeller(cmd)
	AddFlagMarket(cmd)
	AddFlagAssets(cmd)
	AddFlagPrice(cmd)
	AddFlagSettlementFee(cmd)
	AddFlagAllowPartial(cmd)
	AddFlagExternalID(cmd)
	AddFlagCreationFee(cmd)
}

// MakeMsgCreateAskRequest reads all the create-ask flags and creates the desired Msg.
func MakeMsgCreateAskRequest(clientCtx client.Context, flagSet *pflag.FlagSet, _ []string) (sdk.Msg, error) {
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

// AddCreateBidFlags adds all the flags needed for the create-bid command.
func AddCreateBidFlags(cmd *cobra.Command) {
	AddFlagBuyer(cmd)
	AddFlagMarket(cmd)
	AddFlagAssets(cmd)
	AddFlagPrice(cmd)
	AddFlagSettlementFee(cmd)
	AddFlagAllowPartial(cmd)
	AddFlagExternalID(cmd)
	AddFlagCreationFee(cmd)
}

// MakeMsgCreateBidRequest reads all the create-bid flags and creates the desired Msg.
func MakeMsgCreateBidRequest(clientCtx client.Context, flagSet *pflag.FlagSet, _ []string) (sdk.Msg, error) {
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

// AddCancelOrderFlags adds all the flags needed for the cancel-order command.
func AddCancelOrderFlags(cmd *cobra.Command) {
	AddFlagSigner(cmd)
	AddFlagOrder(cmd)
}

// MakeMsgCancelOrderRequest reads all the cancel-order flags and the provided args and creates the desired Msg.
func MakeMsgCancelOrderRequest(clientCtx client.Context, flagSet *pflag.FlagSet, args []string) (sdk.Msg, error) {
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

	return msg, errors.Join(errs...)
}

// AddFillBidsFlags adds all the flags needed for the fill-bids command.
func AddFillBidsFlags(cmd *cobra.Command) {
	AddFlagSeller(cmd)
	AddFlagMarket(cmd)
	AddFlagTotalAssets(cmd)
	AddFlagBids(cmd)
	AddFlagSettlementFee(cmd)
	AddFlagCreationFee(cmd)
}

// MakeMsgFillBidsRequest reads all the fill-bids flags and creates the desired Msg.
func MakeMsgFillBidsRequest(clientCtx client.Context, flagSet *pflag.FlagSet, _ []string) (sdk.Msg, error) {
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

// AddFillAsksFlags adds all the flags needed for the fill-asks command.
func AddFillAsksFlags(cmd *cobra.Command) {
	AddFlagBuyer(cmd)
	AddFlagMarket(cmd)
	AddFlagTotalPrice(cmd)
	AddFlagAsks(cmd)
	AddFlagSettlementFee(cmd)
	AddFlagCreationFee(cmd)
}

// MakeMsgFillAsksRequest reads all the fill-asks flags and creates the desired Msg.
func MakeMsgFillAsksRequest(clientCtx client.Context, flagSet *pflag.FlagSet, _ []string) (sdk.Msg, error) {
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

// AddMarketSettleFlags adds all the flags needed for the market-settle command.
func AddMarketSettleFlags(cmd *cobra.Command) {
	AddFlagAdmin(cmd)
	AddFlagMarket(cmd)
	AddFlagAsks(cmd)
	AddFlagBids(cmd)
	AddFlagExpectPartial(cmd)
}

// MakeMsgMarketSettleRequest reads all the fill-asks flags and creates the desired Msg.
func MakeMsgMarketSettleRequest(clientCtx client.Context, flagSet *pflag.FlagSet, _ []string) (sdk.Msg, error) {
	msg := &exchange.MsgMarketSettleRequest{}

	errs := make([]error, 5)
	msg.Admin, errs[0] = ReadFlagAdminOrDefault(clientCtx, flagSet)
	msg.MarketId, errs[1] = ReadFlagMarket(flagSet)
	msg.AskOrderIds, errs[2] = ReadFlagAsks(flagSet)
	msg.BidOrderIds, errs[3] = ReadFlagBids(flagSet)
	msg.ExpectPartial, errs[4] = ReadFlagPartial(flagSet)

	return msg, errors.Join(errs...)
}
