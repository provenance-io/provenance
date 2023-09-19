package keeper

import (
	"errors"
	"fmt"

	"github.com/gogo/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/provenance-io/provenance/x/exchange"
)

// getOrderStoreKeyValue creates the store key and value representing the provided order.
func (k Keeper) getOrderStoreKeyValue(order exchange.Order) ([]byte, []byte, error) {
	// 200 chosen to hopefully be more than what's needed for 99% of orders.
	// TODO[1658]: Marshal some ask and bid orders to get their actual sizes and make sure 200 is okay.
	key := MakeKeyOrder(order.OrderId)
	value := make([]byte, 1, 200)
	value[0] = order.GetOrderTypeByte()
	switch value[0] {
	case OrderKeyTypeAsk:
		ask := order.GetAskOrder()
		data, err := k.cdc.Marshal(ask)
		if err != nil {
			return nil, nil, fmt.Errorf("error marshaling ask order: %w", err)
		}
		value = append(value, data...)
		return key, value, nil
	case OrderKeyTypeBid:
		bid := order.GetBidOrder()
		data, err := k.cdc.Marshal(bid)
		if err != nil {
			return nil, nil, fmt.Errorf("error marshaling bid order: %w", err)
		}
		value = append(value, data...)
		return key, value, nil
	default:
		// GetOrderTypeByte panics if it's an unknown order type.
		// If we get here, it knew of a new order type that doesn't have a case in here.
		// So we panic here instead of returning an error because it's a programming error to address.
		panic(fmt.Sprintf("SetOrder missing case for order type byte %#x, order type %T", value[0], order.GetOrder()))
	}
}

// parseOrderStoreValue converts an order store value back into an order.
func (k Keeper) parseOrderStoreValue(orderID uint64, value []byte) (*exchange.Order, error) {
	if len(value) == 0 {
		return nil, nil
	}
	typeByte, data := value[0], value[1:]
	switch typeByte {
	case OrderKeyTypeAsk:
		var ask exchange.AskOrder
		err := k.cdc.Unmarshal(data, &ask)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal ask order: %w", err)
		}
		return exchange.NewOrder(orderID).WithAsk(&ask), nil
	case OrderKeyTypeBid:
		var bid exchange.BidOrder
		err := k.cdc.Unmarshal(data, &bid)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal bid order: %w", err)
		}
		return exchange.NewOrder(orderID).WithBid(&bid), nil
	default:
		// Returning an error here instead of panicking because, at this point, we don't really
		// have a trail of what happened to get this unknown entry here. And if we panic, it makes
		// it harder to identify and clean up bad entries once we figure that out.
		return nil, fmt.Errorf("unknown type byte %#x", typeByte)
	}
}

// createIndexEntries creates all the key/value index entries for an order.
func createIndexEntries(order exchange.Order) []sdk.KVPair {
	marketID := order.GetMarketID()
	orderID := order.GetOrderId()
	orderTypeByte := order.GetOrderTypeByte()
	owner := order.GetOwner()
	addr := sdk.MustAccAddressFromBech32(owner)
	assets := order.GetAssets()

	rv := []sdk.KVPair{
		{
			Key:   MakeIndexKeyMarketToOrder(marketID, orderID),
			Value: []byte{orderTypeByte},
		},
		{
			Key:   MakeIndexKeyAddressToOrder(addr, orderID),
			Value: []byte{orderTypeByte},
		},
	}

	for _, asset := range assets {
		rv = append(rv, sdk.KVPair{
			Key:   MakeIndexKeyAssetToOrder(asset.Denom, orderTypeByte, orderID),
			Value: nil,
		})
	}

	return rv
}

// getOrderFromStore looks up an order from the store. Returns nil, nil if the order does not exist.
func (k Keeper) getOrderFromStore(store sdk.KVStore, orderID uint64) (*exchange.Order, error) {
	key := MakeKeyOrder(orderID)
	value := store.Get(key)
	rv, err := k.parseOrderStoreValue(orderID, value)
	if err != nil {
		return nil, fmt.Errorf("failed to read order %d: %w", orderID, err)
	}
	return rv, nil
}

// setOrderInStore writes an order to the store (along with all its indexes).
func (k Keeper) setOrderInStore(store sdk.KVStore, order exchange.Order) error {
	key, value, err := k.getOrderStoreKeyValue(order)
	if err != nil {
		return fmt.Errorf("failed to create order %d store key/value: %w", order.GetOrderId(), err)
	}
	isUpdate := store.Has(key)
	store.Set(key, value)
	if !isUpdate {
		indexEntries := createIndexEntries(order)
		for _, entry := range indexEntries {
			store.Set(entry.Key, entry.Value)
		}
		// It is assumed that these index entries cannot change over the life of an order.
		// The only change that is allowed to an order is the assets (due to partial fulfillment).
		// Partial fulfillment is only allowed when there's a single asset type (denom), though.
		// That's why no attempt is made in here to update index entries when the order already exists.
	}
	return nil
}

// deleteAndDeIndexOrder deletes an order from the store along with its indexes.
func deleteAndDeIndexOrder(store sdk.KVStore, order exchange.Order) {
	key := MakeKeyOrder(order.OrderId)
	store.Delete(key)
	indexEntries := createIndexEntries(order)
	for _, entry := range indexEntries {
		store.Delete(entry.Key)
	}
}

// deleteOrder deletes an order (along with its indexes).
func (k Keeper) deleteOrder(ctx sdk.Context, order exchange.Order) {
	deleteAndDeIndexOrder(k.getStore(ctx), order)
}

// GetOrder gets an order. Returns nil, nil if the order does not exist.
func (k Keeper) GetOrder(ctx sdk.Context, orderID uint64) (*exchange.Order, error) {
	return k.getOrderFromStore(k.getStore(ctx), orderID)
}

// getNextOrderID gets the next available order id from the store.
func (k Keeper) getNextOrderID(ctx sdk.Context) uint64 {
	store := prefix.NewStore(k.getStore(ctx), GetKeyPrefixOrder())
	iter := store.ReverseIterator(nil, nil)
	defer iter.Close()
	if iter.Valid() {
		orderIDBz := iter.Key()
		orderID := uint64FromBz(orderIDBz)
		return orderID + 1
	}
	return 1
}

// IterateMarketOrders iterates over all orders for a market.
// The callback takes in the order id and order type byte and should return whether to stop iterating.
func (k Keeper) IterateMarketOrders(ctx sdk.Context, marketID uint32, cb func(orderID uint64, orderTypeByte byte) bool) {
	k.iterate(ctx, GetIndexKeyPrefixMarketToOrder(marketID), func(key, value []byte) bool {
		if len(value) == 0 {
			return false
		}
		_, orderID, err := ParseIndexKeyMarketToOrder(key)
		if err != nil {
			return false
		}
		return cb(orderID, value[0])
	})
}

// IterateAddressOrders iterates over all orders for an address.
// The callback takes in the order id and order type byte and should return whether to stop iterating.
func (k Keeper) IterateAddressOrders(ctx sdk.Context, addr sdk.AccAddress, cb func(orderID uint64, orderTypeByte byte) bool) {
	k.iterate(ctx, GetIndexKeyPrefixAddressToOrder(addr), func(key, value []byte) bool {
		if len(value) == 0 {
			return false
		}
		_, orderID, err := ParseIndexKeyAddressToOrder(key)
		if err != nil {
			return false
		}
		return cb(orderID, value[0])
	})
}

// IterateAssetOrders iterates over all orders for a given asset denom.
// The callback takes in the order id and order type byte and should return whether to stop iterating.
func (k Keeper) IterateAssetOrders(ctx sdk.Context, assetDenom string, cb func(orderID uint64, orderTypeByte byte) bool) {
	k.iterate(ctx, GetIndexKeyPrefixAssetToOrder(assetDenom), func(key, _ []byte) bool {
		_, orderTypeByte, orderID, err := ParseIndexKeyAssetToOrder(key)
		if err != nil {
			return false
		}
		return cb(orderID, orderTypeByte)
	})
}

// IterateAssetAskOrders iterates over all ask orders for a given asset denom.
// The callback takes in the order id and should return whether to stop iterating.
func (k Keeper) IterateAssetAskOrders(ctx sdk.Context, assetDenom string, cb func(orderID uint64) bool) {
	k.iterate(ctx, GetIndexKeyPrefixAssetToOrderAsks(assetDenom), func(key, _ []byte) bool {
		_, _, orderID, err := ParseIndexKeyAssetToOrder(key)
		if err != nil {
			return false
		}
		return cb(orderID)
	})
}

// IterateAssetBidOrders iterates over all bid orders for a given asset denom.
// The callback takes in the order id and should return whether to stop iterating.
func (k Keeper) IterateAssetBidOrders(ctx sdk.Context, assetDenom string, cb func(orderID uint64) bool) {
	k.iterate(ctx, GetIndexKeyPrefixAssetToOrderBids(assetDenom), func(key, _ []byte) bool {
		_, _, orderID, err := ParseIndexKeyAssetToOrder(key)
		if err != nil {
			return false
		}
		return cb(orderID)
	})
}

// placeHoldOnOrder places a hold on an order's funds in the owner's account.
func (k Keeper) placeHoldOnOrder(ctx sdk.Context, order *exchange.Order) error {
	orderID := order.OrderId
	orderType := order.GetOrderType()
	owner := order.GetOwner()
	ownerAddr, err := sdk.AccAddressFromBech32(owner)
	if err != nil {
		return fmt.Errorf("invalid %s order %d owner %q: %w", orderType, orderID, owner, err)
	}
	toHold := order.GetHoldAmount()
	err = k.holdKeeper.AddHold(ctx, ownerAddr, toHold, fmt.Sprintf("x/exchange: order %d", orderID))
	if err != nil {
		return fmt.Errorf("error placing hold for %s order %d: %w", orderType, orderID, err)
	}
	return nil
}

// releaseHoldOnOrder releases a hold that was placed on an order's funds in the owner's account.
func (k Keeper) releaseHoldOnOrder(ctx sdk.Context, order *exchange.Order) error {
	orderID := order.OrderId
	orderType := order.GetOrderType()
	owner := order.GetOwner()
	ownerAddr, err := sdk.AccAddressFromBech32(owner)
	if err != nil {
		return fmt.Errorf("invalid %s order %d owner %q: %w", orderType, orderID, owner, err)
	}
	held := order.GetHoldAmount()
	err = k.holdKeeper.ReleaseHold(ctx, ownerAddr, held)
	if err != nil {
		return fmt.Errorf("error releasing hold for %s order %d: %w", orderType, orderID, err)
	}
	return nil
}

// CreateAskOrder creates an ask order, collects the creation fee, and places all needed holds.
func (k Keeper) CreateAskOrder(ctx sdk.Context, askOrder exchange.AskOrder, creationFee *sdk.Coin) (uint64, error) {
	if err := askOrder.Validate(); err != nil {
		return 0, err
	}

	store := k.getStore(ctx)
	marketID := askOrder.MarketId

	if err := validateMarketExists(store, marketID); err != nil {
		return 0, err
	}
	if !isMarketActive(store, marketID) {
		return 0, fmt.Errorf("market %d is not accepting orders", marketID)
	}
	seller := sdk.MustAccAddressFromBech32(askOrder.Seller)
	if !k.CanCreateAsk(ctx, marketID, seller) {
		return 0, fmt.Errorf("account %s is not allowed to create ask orders in market %d", seller, marketID)
	}
	if err := validateCreateAskFlatFee(store, marketID, creationFee); err != nil {
		return 0, err
	}
	if err := validateSellerSettlementFlatFee(store, marketID, askOrder.SellerSettlementFlatFee); err != nil {
		return 0, err
	}
	if err := validateAskPrice(store, marketID, askOrder.Price, askOrder.SellerSettlementFlatFee); err != nil {
		return 0, err
	}

	if creationFee != nil {
		err := k.CollectFee(ctx, seller, marketID, sdk.Coins{*creationFee})
		if err != nil {
			return 0, fmt.Errorf("error collecting ask order creation fee: %w", err)
		}
	}

	orderID := k.getNextOrderID(ctx)
	order := exchange.NewOrder(orderID).WithAsk(&askOrder)
	if err := k.setOrderInStore(store, *order); err != nil {
		return 0, fmt.Errorf("error storing ask order: %w", err)
	}

	if err := k.placeHoldOnOrder(ctx, order); err != nil {
		return 0, err
	}

	return orderID, ctx.EventManager().EmitTypedEvent(exchange.NewEventOrderCreated(order))
}

// CreateBidOrder creates a bid order, collects the creation fee, and places all needed holds.
func (k Keeper) CreateBidOrder(ctx sdk.Context, bidOrder exchange.BidOrder, creationFee *sdk.Coin) (uint64, error) {
	if err := bidOrder.Validate(); err != nil {
		return 0, err
	}

	store := k.getStore(ctx)
	marketID := bidOrder.MarketId

	if err := validateMarketExists(store, marketID); err != nil {
		return 0, err
	}
	if !isMarketActive(store, marketID) {
		return 0, fmt.Errorf("market %d is not accepting orders", marketID)
	}
	buyer := sdk.MustAccAddressFromBech32(bidOrder.Buyer)
	if !k.CanCreateBid(ctx, marketID, buyer) {
		return 0, fmt.Errorf("account %s is not allowed to create bid orders in market %d", buyer, marketID)
	}
	if err := validateCreateBidFlatFee(store, marketID, creationFee); err != nil {
		return 0, err
	}
	if err := validateBuyerSettlementFee(store, marketID, bidOrder.Price, bidOrder.BuyerSettlementFees); err != nil {
		return 0, err
	}

	if creationFee != nil {
		err := k.CollectFee(ctx, buyer, marketID, sdk.Coins{*creationFee})
		if err != nil {
			return 0, fmt.Errorf("error collecting bid order creation fee: %w", err)
		}
	}

	orderID := k.getNextOrderID(ctx)
	order := exchange.NewOrder(orderID).WithBid(&bidOrder)
	if err := k.setOrderInStore(store, *order); err != nil {
		return 0, fmt.Errorf("error storing bid order: %w", err)
	}

	if err := k.placeHoldOnOrder(ctx, order); err != nil {
		return 0, err
	}

	return orderID, ctx.EventManager().EmitTypedEvent(exchange.NewEventOrderCreated(order))
}

// CancelOrder releases an order's held funds and deletes it.
func (k Keeper) CancelOrder(ctx sdk.Context, orderID uint64, signer string) error {
	order, err := k.GetOrder(ctx, orderID)
	if err != nil {
		return err
	}
	if order == nil {
		return fmt.Errorf("order %d does not exist", orderID)
	}

	orderOwner := order.GetOwner()
	if signer != orderOwner && !k.CanCancelMarketOrders(ctx, order.GetMarketID(), signer) {
		return fmt.Errorf("account %s does not have permission to cancel order %d", signer, orderID)
	}

	signerAddr := sdk.MustAccAddressFromBech32(signer)
	orderOwnerAddr := sdk.MustAccAddressFromBech32(orderOwner)
	heldAmount := order.GetHoldAmount()
	err = k.holdKeeper.ReleaseHold(ctx, orderOwnerAddr, heldAmount)
	if err != nil {
		return fmt.Errorf("unable to release hold on order %d funds: %w", orderID, err)
	}

	k.deleteOrder(ctx, *order)

	return ctx.EventManager().EmitTypedEvent(exchange.NewEventOrderCancelled(orderID, signerAddr))
}

// getBidOrders gets orders from the store, making sure they're bid orders in the given market
// and do not have the same buyer as the provided seller. If the seller isn't yet known, just provide "" for it.
func (k Keeper) getBidOrders(store sdk.KVStore, marketID uint32, orderIDs []uint64, seller string) ([]*exchange.Order, error) {
	var errs []error
	orders := make([]*exchange.Order, 0, len(orderIDs))

	for _, orderID := range orderIDs {
		order, oerr := k.getOrderFromStore(store, orderID)
		if oerr != nil {
			errs = append(errs, oerr)
			continue
		}
		if order == nil {
			errs = append(errs, fmt.Errorf("order %d not found", orderID))
			continue
		}
		if !order.IsBidOrder() {
			errs = append(errs, fmt.Errorf("order %d is type %s: expected bid", orderID, order.GetOrderType()))
			continue
		}

		bidOrder := order.GetBidOrder()
		orderMarketID := bidOrder.MarketId
		buyer := bidOrder.Buyer

		if orderMarketID != marketID {
			errs = append(errs, fmt.Errorf("order %d market id %d does not equal requested market id %d", orderID, orderMarketID, marketID))
			continue
		}
		if buyer == seller {
			errs = append(errs, fmt.Errorf("order %d has the same buyer %s as the requested seller", orderID, buyer))
			continue
		}

		orders = append(orders, order)
	}

	return orders, errors.Join(errs...)
}

// getAskOrders gets orders from the store, making sure they're ask orders in the given market
// and do not have the same sller as the provided buyer. If the buyer isn't yet known, just provide "" for it.
func (k Keeper) getAskOrders(store sdk.KVStore, marketID uint32, orderIDs []uint64, buyer string) ([]*exchange.Order, error) {
	var errs []error
	orders := make([]*exchange.Order, 0, len(orderIDs))

	for _, orderID := range orderIDs {
		order, oerr := k.getOrderFromStore(store, orderID)
		if oerr != nil {
			errs = append(errs, oerr)
			continue
		}
		if order == nil {
			errs = append(errs, fmt.Errorf("order %d not found", orderID))
			continue
		}
		if !order.IsAskOrder() {
			errs = append(errs, fmt.Errorf("order %d is type %s: expected ask", orderID, order.GetOrderType()))
			continue
		}

		askOrder := order.GetAskOrder()
		orderMarketID := askOrder.MarketId
		seller := askOrder.Seller

		if orderMarketID != marketID {
			errs = append(errs, fmt.Errorf("order %d market id %d does not equal requested market id %d", orderID, orderMarketID, marketID))
			continue
		}
		if seller == buyer {
			errs = append(errs, fmt.Errorf("order %d has the same seller %s as the requested buyer", orderID, seller))
			continue
		}

		orders = append(orders, order)
	}

	return orders, errors.Join(errs...)
}

// FillBids settles one or more bid orders for a seller.
func (k Keeper) FillBids(ctx sdk.Context, msg *exchange.MsgFillBidsRequest) error {
	if err := msg.ValidateBasic(); err != nil {
		return err
	}

	marketID := msg.MarketId
	store := k.getStore(ctx)

	if err := validateMarketExists(store, marketID); err != nil {
		return err
	}
	if !isMarketActive(store, marketID) {
		return fmt.Errorf("market %d is not accepting orders", marketID)
	}
	if !isUserSettlementAllowed(store, marketID) {
		return fmt.Errorf("market %d does not allow user settlement", marketID)
	}
	seller, serr := sdk.AccAddressFromBech32(msg.Seller)
	if serr != nil {
		return fmt.Errorf("invalid seller %q: %w", msg.Seller, serr)
	}
	if !k.CanCreateAsk(ctx, marketID, seller) {
		return fmt.Errorf("account %s is not allowed to create ask orders in market %d", seller, marketID)
	}
	if err := validateCreateAskFlatFee(store, marketID, msg.AskOrderCreationFee); err != nil {
		return err
	}
	if err := validateSellerSettlementFlatFee(store, marketID, msg.SellerSettlementFlatFee); err != nil {
		return err
	}

	orders, oerrs := k.getBidOrders(store, marketID, msg.BidOrderIds, msg.Seller)
	if oerrs != nil {
		return oerrs
	}

	var errs []error
	var totalAssets, totalPrice, totalSellerFee sdk.Coins
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
			errs = append(errs, fmt.Errorf("error calculating seller settlement ratio fee for order %d: %w", order.OrderId, rerr))
			continue
		}
		if err := k.releaseHoldOnOrder(ctx, order); err != nil {
			errs = append(errs, err)
			continue
		}

		totalAssets = totalAssets.Add(assets...)
		totalPrice = totalPrice.Add(price)
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
		assetOutputs[i].Coins = assetOutputs[i].Coins.Add(assets...)
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

	if !safeCoinsEquals(totalAssets, msg.TotalAssets) {
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

	if err := validateMarketExists(store, marketID); err != nil {
		return err
	}
	if !isMarketActive(store, marketID) {
		return fmt.Errorf("market %d is not accepting orders", marketID)
	}
	if !isUserSettlementAllowed(store, marketID) {
		return fmt.Errorf("market %d does not allow user settlement", marketID)
	}
	buyer, serr := sdk.AccAddressFromBech32(msg.Buyer)
	if serr != nil {
		return fmt.Errorf("invalid buyer %q: %w", msg.Buyer, serr)
	}
	if !k.CanCreateBid(ctx, marketID, buyer) {
		return fmt.Errorf("account %s is not allowed to create bid orders in market %d", buyer, marketID)
	}
	if err := validateCreateBidFlatFee(store, marketID, msg.BidOrderCreationFee); err != nil {
		return err
	}
	if err := validateBuyerSettlementFee(store, marketID, msg.TotalPrice, msg.BuyerSettlementFees); err != nil {
		return err
	}

	orders, oerrs := k.getAskOrders(store, marketID, msg.AskOrderIds, msg.Buyer)
	if oerrs != nil {
		return oerrs
	}

	var errs []error
	var totalAssets, totalPrice sdk.Coins
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
			errs = append(errs, fmt.Errorf("error calculating seller settlement ratio fee for order %d: %w", order.OrderId, rerr))
			continue
		}
		if err := k.releaseHoldOnOrder(ctx, order); err != nil {
			errs = append(errs, err)
			continue
		}

		totalAssets = totalAssets.Add(assets...)
		totalPrice = totalPrice.Add(price)
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

	if !safeCoinsEquals(totalPrice, sdk.Coins{msg.TotalPrice}) {
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

func (k Keeper) SettleOrders(ctx sdk.Context, marketID uint32, askOrderIDs, bidOrderIds []uint64, expectPartial bool) error {
	// TODO[1658]: Implement SettleOrders.
	panic("Not implemented")
}

// safeCoinsEquals returns true if the two provided coins are equal.
// Returns false instead of panicking like sdk.Coins.IsEqual.
func safeCoinsEquals(a, b sdk.Coins) (isEqual bool) {
	// The sdk.Coins.IsEqual function will panic if a and b have the same number of entries, but different denoms.
	// Really, that stuff is all pretty panic happy.
	// In here, we don't really care why it panics, but if it does, they're not equal.
	defer func() {
		if r := recover(); r != nil {
			isEqual = false
		}
	}()
	return a.IsEqual(b)
}
