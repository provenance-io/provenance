package keeper

import (
	"errors"
	"fmt"

	"github.com/gogo/protobuf/proto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/provenance-io/provenance/x/exchange"
)

// sumAssetsAndPrice gets the sum of assets, and the sum of prices of the provided orders.
func sumAssetsAndPrice(orders []*exchange.Order) (sdk.Coins, sdk.Coins) {
	var totalAssets, totalPrice sdk.Coins
	for _, order := range orders {
		totalAssets = totalAssets.Add(order.GetAssets()...)
		totalPrice = totalPrice.Add(order.GetPrice())
	}
	return totalAssets, totalPrice
}

// validateAcceptingOrdersAndCanUserSettle returns an error if the market isn't active or doesn't allow user settlement.
func validateAcceptingOrdersAndCanUserSettle(store sdk.KVStore, marketID uint32) error {
	if err := validateMarketIsAcceptingOrders(store, marketID); err != nil {
		return err
	}
	if !isUserSettlementAllowed(store, marketID) {
		return fmt.Errorf("market %d does not allow user settlement", marketID)
	}
	return nil
}

// FillBids settles one or more bid orders for a seller.
func (k Keeper) FillBids(ctx sdk.Context, msg *exchange.MsgFillBidsRequest) error {
	if err := msg.ValidateBasic(); err != nil {
		return err
	}

	marketID := msg.MarketId
	store := k.getStore(ctx)

	if err := validateAcceptingOrdersAndCanUserSettle(store, marketID); err != nil {
		return err
	}
	seller := sdk.MustAccAddressFromBech32(msg.Seller)
	if err := k.validateCanCreateAsk(ctx, marketID, seller); err != nil {
		return err
	}
	if err := validateCreateAskFees(store, marketID, msg.AskOrderCreationFee, msg.SellerSettlementFlatFee); err != nil {
		return err
	}

	orders, oerrs := k.getBidOrders(store, marketID, msg.BidOrderIds, msg.Seller)
	if oerrs != nil {
		return oerrs
	}

	var errs []error
	var totalSellerFee sdk.Coins
	assetOutputs := make([]banktypes.Output, 0, len(msg.BidOrderIds))
	priceInputs := make([]banktypes.Input, 0, len(msg.BidOrderIds))
	addrIndex := make(map[string]int)
	feeInputs := make([]banktypes.Input, 0, len(msg.BidOrderIds)+1)
	feeAddrIndex := make(map[string]int)
	for _, order := range orders {
		bidOrder := order.GetBidOrder()
		buyer := bidOrder.Buyer
		assets := bidOrder.Assets
		price := bidOrder.Price
		buyerSettlementFees := bidOrder.BuyerSettlementFees

		sellerRatioFee, rerr := calculateSellerSettlementRatioFee(store, marketID, price)
		if rerr != nil {
			errs = append(errs, fmt.Errorf("error calculating seller settlement ratio fee for order %d: %w",
				order.OrderId, rerr))
			continue
		}
		if err := k.releaseHoldOnOrder(ctx, order); err != nil {
			errs = append(errs, err)
			continue
		}

		if sellerRatioFee != nil {
			totalSellerFee = totalSellerFee.Add(*sellerRatioFee)
		}

		i, seen := addrIndex[buyer]
		if !seen {
			i = len(assetOutputs)
			addrIndex[buyer] = i
			assetOutputs = append(assetOutputs, banktypes.Output{Address: buyer})
			priceInputs = append(priceInputs, banktypes.Input{Address: buyer})
		}
		assetOutputs[i].Coins = assetOutputs[i].Coins.Add(assets)
		priceInputs[i].Coins = priceInputs[i].Coins.Add(price)

		if !buyerSettlementFees.IsZero() {
			j, known := feeAddrIndex[buyer]
			if !known {
				j = len(feeInputs)
				feeAddrIndex[buyer] = j
				feeInputs = append(feeInputs, banktypes.Input{Address: buyer})
			}
			feeInputs[j].Coins = feeInputs[j].Coins.Add(buyerSettlementFees...)
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	totalAssets, totalPrice := sumAssetsAndPrice(orders)

	if !exchange.CoinsEquals(totalAssets, msg.TotalAssets) {
		return fmt.Errorf("total assets %q does not equal sum of bid order assets %q", msg.TotalAssets, totalAssets)
	}

	if msg.SellerSettlementFlatFee != nil {
		totalSellerFee = totalSellerFee.Add(*msg.SellerSettlementFlatFee)
	}
	if !totalSellerFee.IsZero() {
		feeInputs = append(feeInputs, banktypes.Input{Address: msg.Seller, Coins: totalSellerFee})
	}

	assetInputs := []banktypes.Input{{Address: msg.Seller, Coins: msg.TotalAssets}}
	priceOutputs := []banktypes.Output{{Address: msg.Seller, Coins: totalPrice}}

	if err := k.bankKeeper.InputOutputCoins(ctx, assetInputs, assetOutputs); err != nil {
		return fmt.Errorf("error transferring assets from seller to buyers: %w", err)
	}

	if err := k.bankKeeper.InputOutputCoins(ctx, priceInputs, priceOutputs); err != nil {
		return fmt.Errorf("error transferring price from buyers to seller: %w", err)
	}

	if err := k.CollectFees(ctx, feeInputs, marketID); err != nil {
		return fmt.Errorf("error collecting settlement fees: %w", err)
	}

	// Collected last so that it's easier for a seller to fill bids without needing those funds first.
	// Collected separately so it's not combined with the seller settlement fees in the events.
	if msg.AskOrderCreationFee != nil {
		if err := k.CollectFee(ctx, seller, marketID, sdk.Coins{*msg.AskOrderCreationFee}); err != nil {
			return fmt.Errorf("error collecting create-ask fee %q: %w", msg.AskOrderCreationFee, err)
		}
	}

	events := make([]proto.Message, len(orders))
	for i, order := range orders {
		deleteAndDeIndexOrder(store, *order)
		events[i] = exchange.NewEventOrderFilled(order.OrderId)
	}

	return ctx.EventManager().EmitTypedEvents(events...)
}

// FillAsks settles one or more ask orders for a buyer.
func (k Keeper) FillAsks(ctx sdk.Context, msg *exchange.MsgFillAsksRequest) error {
	if err := msg.ValidateBasic(); err != nil {
		return err
	}

	marketID := msg.MarketId
	store := k.getStore(ctx)

	if err := validateAcceptingOrdersAndCanUserSettle(store, marketID); err != nil {
		return err
	}
	buyer, serr := sdk.AccAddressFromBech32(msg.Buyer)
	if serr != nil {
		return fmt.Errorf("invalid buyer %q: %w", msg.Buyer, serr)
	}
	if err := k.validateCanCreateBid(ctx, marketID, buyer); err != nil {
		return err
	}
	if err := validateCreateBidFees(store, marketID, msg.BidOrderCreationFee, msg.TotalPrice, msg.BuyerSettlementFees); err != nil {
		return err
	}

	orders, oerrs := k.getAskOrders(store, marketID, msg.AskOrderIds, msg.Buyer)
	if oerrs != nil {
		return oerrs
	}

	var errs []error
	assetInputs := make([]banktypes.Input, 0, len(msg.AskOrderIds))
	priceOutputs := make([]banktypes.Output, 0, len(msg.AskOrderIds))
	addrIndex := make(map[string]int)
	feeInputs := make([]banktypes.Input, 0, len(msg.AskOrderIds)+1)
	feeAddrIndex := make(map[string]int)
	for _, order := range orders {
		askOrder := order.GetAskOrder()
		seller := askOrder.Seller
		assets := askOrder.Assets
		price := askOrder.Price
		sellerSettlementFlatFee := askOrder.SellerSettlementFlatFee

		sellerRatioFee, rerr := calculateSellerSettlementRatioFee(store, marketID, price)
		if rerr != nil {
			errs = append(errs, fmt.Errorf("error calculating seller settlement ratio fee for order %d: %w",
				order.OrderId, rerr))
			continue
		}
		if err := k.releaseHoldOnOrder(ctx, order); err != nil {
			errs = append(errs, err)
			continue
		}

		var totalSellerFee sdk.Coins
		if sellerSettlementFlatFee != nil && !sellerSettlementFlatFee.IsZero() {
			totalSellerFee = totalSellerFee.Add(*sellerSettlementFlatFee)
		}
		if sellerRatioFee != nil && !sellerRatioFee.IsZero() {
			totalSellerFee = totalSellerFee.Add(*sellerRatioFee)
		}

		i, seen := addrIndex[seller]
		if !seen {
			i = len(assetInputs)
			addrIndex[seller] = i
			assetInputs = append(assetInputs, banktypes.Input{Address: seller})
			priceOutputs = append(priceOutputs, banktypes.Output{Address: seller})
		}
		assetInputs[i].Coins = assetInputs[i].Coins.Add(assets...)
		priceOutputs[i].Coins = priceOutputs[i].Coins.Add(price)

		if !totalSellerFee.IsZero() {
			j, known := feeAddrIndex[seller]
			if !known {
				j = len(feeInputs)
				feeAddrIndex[seller] = j
				feeInputs = append(feeInputs, banktypes.Input{Address: seller})
			}
			feeInputs[j].Coins = feeInputs[j].Coins.Add(totalSellerFee...)
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	totalAssets, totalPrice := sumAssetsAndPrice(orders)

	if !exchange.CoinsEquals(totalPrice, sdk.Coins{msg.TotalPrice}) {
		return fmt.Errorf("total price %q does not equal sum of ask order prices %q", msg.TotalPrice, totalPrice)
	}

	if !msg.BuyerSettlementFees.IsZero() {
		feeInputs = append(feeInputs, banktypes.Input{Address: msg.Buyer, Coins: msg.BuyerSettlementFees})
	}

	assetOutputs := []banktypes.Output{{Address: msg.Buyer, Coins: totalAssets}}
	priceInputs := []banktypes.Input{{Address: msg.Buyer, Coins: sdk.Coins{msg.TotalPrice}}}

	if err := k.bankKeeper.InputOutputCoins(ctx, assetInputs, assetOutputs); err != nil {
		return fmt.Errorf("error transferring assets from sellers to buyer: %w", err)
	}

	if err := k.bankKeeper.InputOutputCoins(ctx, priceInputs, priceOutputs); err != nil {
		return fmt.Errorf("error transferring price from buyer to sellers: %w", err)
	}

	if err := k.CollectFees(ctx, feeInputs, marketID); err != nil {
		return fmt.Errorf("error collecting settlement fees: %w", err)
	}

	// Collected last so that it's easier for a seller to fill asks without needing those funds first.
	// Collected separately so it's not combined with the buyer settlement fees in the events.
	if msg.BidOrderCreationFee != nil {
		if err := k.CollectFee(ctx, buyer, marketID, sdk.Coins{*msg.BidOrderCreationFee}); err != nil {
			return fmt.Errorf("error collecting create-ask fee %q: %w", msg.BidOrderCreationFee, err)
		}
	}

	events := make([]proto.Message, len(orders))
	for i, order := range orders {
		deleteAndDeIndexOrder(store, *order)
		events[i] = exchange.NewEventOrderFilled(order.OrderId)
	}

	return ctx.EventManager().EmitTypedEvents(events...)
}

// SettleOrders attempts to settle all the provided orders.
func (k Keeper) SettleOrders(ctx sdk.Context, marketID uint32, askOrderIDs, bidOrderIds []uint64, expectPartial bool) error {
	store := k.getStore(ctx)
	if err := validateMarketExists(store, marketID); err != nil {
		return err
	}

	askOrders, aoerr := k.getAskOrders(store, marketID, askOrderIDs, "")
	bidOrders, boerr := k.getBidOrders(store, marketID, bidOrderIds, "")
	if aoerr != nil || boerr != nil {
		return errors.Join(aoerr, boerr)
	}

	totalAssetsForSale, totalAskPrice := sumAssetsAndPrice(askOrders)
	totalAssetsToBuy, totalBidPrice := sumAssetsAndPrice(bidOrders)

	// TODO[1659]: Allow for multiple asset denoms in some cases.

	var errs []error
	if len(totalAssetsForSale) != 1 {
		errs = append(errs, fmt.Errorf("cannot settle with multiple ask order asset denoms %q", totalAssetsForSale))
	}
	if len(totalAskPrice) != 1 {
		errs = append(errs, fmt.Errorf("cannot settle with multiple ask order price denoms %q", totalAskPrice))
	}
	if len(totalAssetsToBuy) != 1 {
		errs = append(errs, fmt.Errorf("cannot settle with multiple bid order asset denoms %q", totalAssetsToBuy))
	}
	if len(totalBidPrice) != 1 {
		errs = append(errs, fmt.Errorf("cannot settle with multiple bid order price denoms %q", totalBidPrice))
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	if totalAssetsForSale[0].Denom != totalAssetsToBuy[0].Denom {
		errs = append(errs, fmt.Errorf("cannot settle different ask %q and bid %q asset denoms",
			totalAssetsForSale, totalAssetsToBuy))
	}
	if totalAskPrice[0].Denom != totalBidPrice[0].Denom {
		errs = append(errs, fmt.Errorf("cannot settle different ask %q and bid %q price denoms",
			totalAskPrice, totalBidPrice))
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	if !expectPartial && !exchange.CoinsEquals(totalAssetsForSale, totalAssetsToBuy) {
		return fmt.Errorf("total assets for sale %q does not equal total assets to buy %q",
			totalAssetsForSale, totalAssetsToBuy)
	}

	// TODO[1658]: Implement SettleOrders.
	panic("Not implemented")
}
