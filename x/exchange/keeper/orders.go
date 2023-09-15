package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

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

// getOrderFromStore gets an order from the store.
func (k Keeper) getOrderFromStore(store sdk.KVStore, orderID uint64) (*exchange.Order, error) {
	key := MakeKeyOrder(orderID)
	value := store.Get(key)
	rv, err := k.parseOrderStoreValue(orderID, value)
	if err != nil {
		return nil, fmt.Errorf("failed to read order %d: %w", orderID, err)
	}
	return rv, nil
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
		// The only change that is allowed to an order is the assets due to partial fulfillment.
		// But partial fulfillment is only allowed when there's a single asset type.
		// That's why no attempt is made in here to update index entries when the order already exists.
	}
	return nil
}

// deleteOrderFromStore deletes an order from the store (along with its indexes).
func deleteOrderFromStore(store sdk.KVStore, order exchange.Order) {
	key := MakeKeyOrder(order.OrderId)
	store.Delete(key)
	indexEntries := createIndexEntries(order)
	for _, entry := range indexEntries {
		store.Delete(entry.Key)
	}
}

// GetOrder gets an order. Returns nil, nil if the order does not exist.
func (k Keeper) GetOrder(ctx sdk.Context, orderID uint64) (*exchange.Order, error) {
	return k.getOrderFromStore(k.getStore(ctx), orderID)
}

// SetOrder stores the provided order in state.
func (k Keeper) SetOrder(ctx sdk.Context, order exchange.Order) error {
	return k.setOrderInStore(k.getStore(ctx), order)
}

// GetNextOrderID gets the next available order id from the store.
func (k Keeper) GetNextOrderID(ctx sdk.Context) uint64 {
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

// CreateAskOrder creates an ask order, collects the creation fee, and places all needed holds.
func (k Keeper) CreateAskOrder(ctx sdk.Context, msg *exchange.MsgCreateAskRequest) (uint64, error) {
	store := k.getStore(ctx)
	marketID := msg.AskOrder.MarketId
	seller := sdk.MustAccAddressFromBech32(msg.AskOrder.Seller)

	if err := validateMarketExists(store, marketID); err != nil {
		return 0, err
	}
	if !k.CanCreateAsk(ctx, marketID, seller) {
		return 0, fmt.Errorf("account %s is not allowed to create ask orders in market %d", seller, marketID)
	}
	if err := validateCreateAskFlatFee(store, marketID, msg.OrderCreationFee); err != nil {
		return 0, err
	}
	if err := validateSellerSettlementFlatFee(store, marketID, msg.AskOrder.SellerSettlementFlatFee); err != nil {
		return 0, err
	}
	if err := validateAskPrice(store, marketID, msg.AskOrder.Price, msg.AskOrder.SellerSettlementFlatFee); err != nil {
		return 0, err
	}

	if msg.OrderCreationFee != nil {
		err := k.CollectFee(ctx, seller, marketID, sdk.Coins{*msg.OrderCreationFee})
		if err != nil {
			return 0, fmt.Errorf("error collecting ask order creation fee: %w", err)
		}
	}

	orderID := k.GetNextOrderID(ctx)
	order := exchange.NewOrder(orderID).WithAsk(&msg.AskOrder)
	if err := k.setOrderInStore(store, *order); err != nil {
		return 0, fmt.Errorf("error storing ask order: %w", err)
	}

	toHold := order.GetHoldAmount()
	if err := k.holdKeeper.AddHold(ctx, seller, toHold, fmt.Sprintf("x/exchange: order %d", order.OrderId)); err != nil {
		return 0, fmt.Errorf("error placing hold for ask order: %w", err)
	}

	return orderID, nil
}

// CreateBidOrder creates a bid order, collects the creation fee, and places all needed holds.
func (k Keeper) CreateBidOrder(ctx sdk.Context, msg *exchange.MsgCreateBidRequest) (uint64, error) {
	store := k.getStore(ctx)
	marketID := msg.BidOrder.MarketId
	buyer := sdk.MustAccAddressFromBech32(msg.BidOrder.Buyer)

	if err := validateMarketExists(store, marketID); err != nil {
		return 0, err
	}
	if !k.CanCreateBid(ctx, marketID, buyer) {
		return 0, fmt.Errorf("account %s is not allowed to create bid orders in market %d", buyer, marketID)
	}
	if err := validateCreateBidFlatFee(store, marketID, msg.OrderCreationFee); err != nil {
		return 0, err
	}
	if err := validateBuyerSettlementFee(store, marketID, msg.BidOrder.Price, msg.BidOrder.BuyerSettlementFees); err != nil {
		return 0, err
	}

	if msg.OrderCreationFee != nil {
		err := k.CollectFee(ctx, buyer, marketID, sdk.Coins{*msg.OrderCreationFee})
		if err != nil {
			return 0, fmt.Errorf("error collecting bid order creation fee: %w", err)
		}
	}

	orderID := k.GetNextOrderID(ctx)
	order := exchange.NewOrder(orderID).WithBid(&msg.BidOrder)
	if err := k.setOrderInStore(store, *order); err != nil {
		return 0, fmt.Errorf("error storing bid order: %w", err)
	}

	toHold := order.GetHoldAmount()
	if err := k.holdKeeper.AddHold(ctx, buyer, toHold, fmt.Sprintf("x/exchange: order %d", order.OrderId)); err != nil {
		return 0, fmt.Errorf("error placing hold for bid order: %w", err)
	}

	return orderID, nil
}

// CancelOrder releases an order's held funds and deletes it.
func (k Keeper) CancelOrder(ctx sdk.Context, msg *exchange.MsgCancelOrderRequest) error {
	store := k.getStore(ctx)
	order, err := k.getOrderFromStore(store, msg.OrderId)
	if err != nil {
		return err
	}
	if order == nil {
		return fmt.Errorf("order %d does not exist", msg.OrderId)
	}

	orderOwner := order.GetOwner()
	if msg.Owner != orderOwner && !k.hasPermission(store, order.GetMarketID(), msg.Owner, exchange.Permission_cancel) {
		return fmt.Errorf("account %s cannot cancel order %d", msg.Owner, msg.OrderId)
	}

	orderOwnerAddr := sdk.MustAccAddressFromBech32(orderOwner)
	heldAmount := order.GetHoldAmount()
	err = k.holdKeeper.ReleaseHold(ctx, orderOwnerAddr, heldAmount)
	if err != nil {
		return fmt.Errorf("unable to release hold on order %d funds: %w", msg.OrderId, err)
	}

	deleteOrderFromStore(store, *order)

	return nil
}
