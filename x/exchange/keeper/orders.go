package keeper

import (
	"errors"
	"fmt"
	"strings"

	db "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/provenance-io/provenance/x/exchange"
)

// getLastOrderID gets the id of the last order created.
func getLastOrderID(store sdk.KVStore) uint64 {
	key := MakeKeyLastOrderID()
	value := store.Get(key)
	rv, _ := uint64FromBz(value)
	return rv
}

// setLastOrderID sets the id of the last order created.
func setLastOrderID(store sdk.KVStore, orderID uint64) {
	key := MakeKeyLastOrderID()
	value := uint64Bz(orderID)
	store.Set(key, value)
}

// nextOrderID finds the next available order id, updates the last order id
// store entry, and returns the unused id it found.
func nextOrderID(store sdk.KVStore) uint64 {
	orderID := getLastOrderID(store) + 1
	setLastOrderID(store, orderID)
	return orderID
}

// getOrderStoreKeyValue creates the store key and value representing the provided order.
func (k Keeper) getOrderStoreKeyValue(order exchange.Order) ([]byte, []byte, error) {
	// 200 chosen to hopefully be more than what's needed for 99% of orders.
	// The largest one I could make was 753 bytes for a bid order with all coins having 128
	// character denoms and 31 digits in the amounts, a 32 byte address, and max market id.
	// But the more realistic ones were 130 to 160, so 200 seems like a nice size to start with.
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

// parseOrderStoreKeyValue parses the provided key and value into an exchange.Order.
// The key's leading type byte is optional, only the last 8 bytes of it are used.
func (k Keeper) parseOrderStoreKeyValue(key, value []byte) (*exchange.Order, error) {
	if len(key) < 8 {
		return nil, fmt.Errorf("invalid order store key %v: length expected to be at least 8", key)
	}
	orderID, _ := uint64FromBz(key[len(key)-8:])
	order, err := k.parseOrderStoreValue(orderID, value)
	if err != nil {
		return nil, fmt.Errorf("failed to read order %d: %w", orderID, err)
	}
	return order, nil
}

// createConstantIndexEntries creates all the key/value index entries for an order that don't change.
// See also: createMarketExternalIDToOrderEntry.
func createConstantIndexEntries(order exchange.Order) []sdk.KVPair {
	marketID := order.GetMarketID()
	orderID := order.GetOrderID()
	orderTypeByte := order.GetOrderTypeByte()
	owner := order.GetOwner()
	addr := sdk.MustAccAddressFromBech32(owner)
	assets := order.GetAssets()

	return []sdk.KVPair{
		{
			Key:   MakeIndexKeyMarketToOrder(marketID, orderID),
			Value: []byte{orderTypeByte},
		},
		{
			Key:   MakeIndexKeyAddressToOrder(addr, orderID),
			Value: []byte{orderTypeByte},
		},
		{
			Key:   MakeIndexKeyAssetToOrder(assets.Denom, orderID),
			Value: []byte{orderTypeByte},
		},
	}
}

// createMarketExternalIDToOrderEntry creates the market external id to order store entry.
// See also createConstantIndexEntries
func createMarketExternalIDToOrderEntry(order exchange.OrderI) *sdk.KVPair {
	externalID := order.GetExternalID()
	if len(externalID) == 0 {
		return nil
	}
	return &sdk.KVPair{
		Key:   MakeIndexKeyMarketExternalIDToOrder(order.GetMarketID(), externalID),
		Value: uint64Bz(order.GetOrderID()),
	}
}

// getOrderFromStore looks up an order from the store. Returns nil, nil if the order does not exist.
func (k Keeper) getOrderFromStore(store sdk.KVStore, orderID uint64) (*exchange.Order, error) {
	key := MakeKeyOrder(orderID)
	value := store.Get(key)
	if len(value) == 0 {
		return nil, nil
	}
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
		return fmt.Errorf("failed to create order %d store key/value: %w", order.GetOrderID(), err)
	}

	externalIDEntry := createMarketExternalIDToOrderEntry(order)
	if externalIDEntry != nil && store.Has(externalIDEntry.Key) {
		orderIDBz := store.Get(externalIDEntry.Key)
		otherOrderID, ok := uint64FromBz(orderIDBz)
		if ok && otherOrderID != order.GetOrderID() {
			return fmt.Errorf("external id %q is already in use by order %d: cannot be used for order %d",
				order.GetExternalID(), otherOrderID, order.GetOrderID())
		}
	}

	isUpdate := store.Has(key)
	store.Set(key, value)

	if !isUpdate {
		indexEntries := createConstantIndexEntries(order)
		for _, entry := range indexEntries {
			store.Set(entry.Key, entry.Value)
		}
	}

	if externalIDEntry != nil {
		store.Set(externalIDEntry.Key, externalIDEntry.Value)
	}
	return nil
}

// deleteAndDeIndexOrder deletes an order from the store along with its indexes.
func deleteAndDeIndexOrder(store sdk.KVStore, order exchange.Order) {
	key := MakeKeyOrder(order.OrderId)
	store.Delete(key)
	indexEntries := createConstantIndexEntries(order)
	for _, entry := range indexEntries {
		store.Delete(entry.Key)
	}
	externalIDEntry := createMarketExternalIDToOrderEntry(order)
	if externalIDEntry != nil {
		store.Delete(externalIDEntry.Key)
	}
}

// iterateOrderIndex iterates over a <something>-to-order index with keys that have the provided prefixBz.
// The callback takes in the order id and order type byte and should return whether to stop iterating.
func (k Keeper) iterateOrderIndex(ctx sdk.Context, prefixBz []byte, cb func(orderID uint64, orderTypeByte byte) bool) {
	k.iterate(ctx, prefixBz, func(key, value []byte) bool {
		if len(value) == 0 {
			return false
		}
		orderID, ok := ParseIndexKeySuffixOrderID(key)
		return ok && cb(orderID, value[0])
	})
}

// getPageOfOrdersFromIndex gets a page of orders using a <something>-to-order index.
func (k Keeper) getPageOfOrdersFromIndex(
	ctx sdk.Context,
	prefixBz []byte,
	pageReq *query.PageRequest,
	orderType string,
	afterOrderID uint64,
) (*query.PageResponse, []*exchange.Order, error) {
	var orderTypeByte byte
	filterByType := false
	if len(orderType) > 0 {
		ot := strings.ToLower(orderType)
		// only look at the first 3 chars to handle stuff like "asks" or "bidOrders" too.
		if len(ot) > 3 {
			ot = ot[:3]
		}
		switch ot {
		case exchange.OrderTypeAsk:
			orderTypeByte = OrderKeyTypeAsk
		case exchange.OrderTypeBid:
			orderTypeByte = OrderKeyTypeBid
		default:
			return nil, nil, fmt.Errorf("unknown order type %q", orderType)
		}
		filterByType = true
	}

	rootStore := k.getStore(ctx)
	var orders []*exchange.Order
	accumulator := func(key []byte, value []byte, accumulate bool) (bool, error) {
		// If filtering by type, but the order type isn't known, or is something else, this entry doesn't count, move on.
		if filterByType && (len(value) == 0 || value[0] != orderTypeByte) {
			return false, nil
		}
		// If we can't get the order id from the key, just pretend like it doesn't exist.
		orderID, ok := ParseIndexKeySuffixOrderID(key)
		if !ok {
			return false, nil
		}
		if accumulate {
			// Only add it to the result if we can read it. This might result in fewer results than the limit,
			// but at least one bad entry won't block others by causing the whole thing to return an error.
			order, err := k.getOrderFromStore(rootStore, orderID)
			if err == nil && order != nil {
				orders = append(orders, order)
			}
		}
		return true, nil
	}
	prefixStore := prefix.NewStore(rootStore, prefixBz)
	pageResp, err := filteredPaginateAfterOrder(prefixStore, pageReq, afterOrderID, accumulator)

	return pageResp, orders, err
}

// filteredPaginateAfterOrder is similar to query.FilteredPaginate except
// allows limiting the iterator to only entries after a certain order id.
// afterOrderID is exclusive, i.e. if it's 2, this will go over all order ids that are 3 or greater.
//
// FilteredPaginate does pagination of all the results in the PrefixStore based on the
// provided PageRequest. onResult should be used to do actual unmarshaling and filter the results.
// If key is provided, the pagination uses the optimized querying.
// If offset is used, the pagination uses lazy filtering i.e., searches through all the records.
// The accumulate parameter represents if the response is valid based on the offset given.
// It will be false for the results (filtered) < offset  and true for `offset > accumulate <= end`.
// When accumulate is set to true the current result should be appended to the result set returned
// to the client.
func filteredPaginateAfterOrder(
	prefixStore sdk.KVStore,
	pageRequest *query.PageRequest,
	afterOrderID uint64,
	onResult func(key []byte, value []byte, accumulate bool) (bool, error),
) (*query.PageResponse, error) {
	// if the PageRequest is nil, use default PageRequest
	if pageRequest == nil {
		pageRequest = &query.PageRequest{}
	}

	offset := pageRequest.Offset
	key := pageRequest.Key
	limit := pageRequest.Limit
	countTotal := pageRequest.CountTotal
	reverse := pageRequest.Reverse

	if offset > 0 && key != nil {
		return nil, fmt.Errorf("invalid request, either offset or key is expected, got both")
	}

	if limit == 0 {
		limit = query.DefaultLimit

		// count total results when the limit is zero/not supplied
		countTotal = true
	}

	if len(key) != 0 {
		// This line is changed from the query.FilteredPaginate version.
		iterator := getOrderIterator(prefixStore, key, reverse, afterOrderID)
		defer iterator.Close()

		var (
			numHits uint64
			nextKey []byte
		)

		// This loop is modified from the query.FilteredPaginate version to set
		// NextKey to the next hit instead of the next entry. This matches the offset behavior.
		for ; iterator.Valid(); iterator.Next() {
			if iterator.Error() != nil {
				return nil, iterator.Error()
			}

			hit, err := onResult(iterator.Key(), iterator.Value(), numHits < limit)
			if err != nil {
				return nil, err
			}

			if hit {
				if numHits == limit {
					nextKey = iterator.Key()
					break
				}
				numHits++
			}
		}

		return &query.PageResponse{
			NextKey: nextKey,
		}, nil
	}

	// This line is changed from the query.FilteredPaginate version.
	iterator := getOrderIterator(prefixStore, nil, reverse, afterOrderID)
	defer iterator.Close()

	end := offset + limit

	var (
		numHits uint64
		nextKey []byte
	)

	for ; iterator.Valid(); iterator.Next() {
		if iterator.Error() != nil {
			return nil, iterator.Error()
		}

		accumulate := numHits >= offset && numHits < end
		hit, err := onResult(iterator.Key(), iterator.Value(), accumulate)
		if err != nil {
			return nil, err
		}

		if hit {
			numHits++
		}

		if numHits == end+1 {
			if nextKey == nil {
				nextKey = iterator.Key()
			}

			if !countTotal {
				break
			}
		}
	}

	res := &query.PageResponse{NextKey: nextKey}
	if countTotal {
		res.Total = numHits
	}

	return res, nil
}

// getOrderIterator is similar to query.getIterator but allows limiting it to only entries after a certain order id.
func getOrderIterator(prefixStore sdk.KVStore, start []byte, reverse bool, afterOrderID uint64) db.Iterator {
	if reverse {
		var end []byte
		if start != nil {
			itr := prefixStore.Iterator(start, nil)
			defer itr.Close()
			if itr.Valid() {
				itr.Next()
				end = itr.Key()
			}
		}
		// If an afterOrderID was given, use the key of it as the first key.
		// This orderIDKey is a change from the query.getIterator version.
		var orderIDKey []byte
		if afterOrderID != 0 {
			orderIDKey = uint64Bz(afterOrderID + 1)
		}
		return prefixStore.ReverseIterator(orderIDKey, end)
	}

	// If a start ("next key") was given, use that as the first key.
	// Otherwise, if an afterOrderID was given, use the key of the next order id as the first key.
	// This if block is a change from the query.getIterator version.
	if len(start) == 0 && afterOrderID != 0 {
		if afterOrderID != 18_446_744_073_709_551_615 {
			afterOrderID++
		}
		start = uint64Bz(afterOrderID)
	}
	return prefixStore.Iterator(start, nil)
}

// validateMarketIsAcceptingOrders makes sure the market exists and is accepting orders.
func validateMarketIsAcceptingOrders(store sdk.KVStore, marketID uint32) error {
	if err := validateMarketExists(store, marketID); err != nil {
		return err
	}
	if !isMarketActive(store, marketID) {
		return fmt.Errorf("market %d is not accepting orders", marketID)
	}
	return nil
}

// validateUserCanCreateAsk makes sure the user can create an ask order in the given market.
func (k Keeper) validateUserCanCreateAsk(ctx sdk.Context, marketID uint32, seller sdk.AccAddress) error {
	if !k.CanCreateAsk(ctx, marketID, seller) {
		return fmt.Errorf("account %s is not allowed to create ask orders in market %d", seller, marketID)
	}
	return nil
}

// validateUserCanCreateBid makes sure the user can create a bid order in the given market.
func (k Keeper) validateUserCanCreateBid(ctx sdk.Context, marketID uint32, buyer sdk.AccAddress) error {
	if !k.CanCreateBid(ctx, marketID, buyer) {
		return fmt.Errorf("account %s is not allowed to create bid orders in market %d", buyer, marketID)
	}
	return nil
}

// validateCreateAskFees makes sure the fees are okay for creating an ask order.
func validateCreateAskFees(store sdk.KVStore, marketID uint32, creationFee *sdk.Coin, settlementFlatFee *sdk.Coin) error {
	if err := validateCreateAskFlatFee(store, marketID, creationFee); err != nil {
		return err
	}
	return validateSellerSettlementFlatFee(store, marketID, settlementFlatFee)
}

// validateCreateBidFees makes sure the fees are okay for creating a bid order.
func validateCreateBidFees(store sdk.KVStore, marketID uint32, creationFee *sdk.Coin, price sdk.Coin, settlementFees sdk.Coins) error {
	if err := validateCreateBidFlatFee(store, marketID, creationFee); err != nil {
		return err
	}
	return validateBuyerSettlementFee(store, marketID, price, settlementFees)
}

// getAskOrders gets orders from the store, making sure they're ask orders in the given market
// and do not have the same seller as the provided buyer. If the buyer isn't yet known, just provide "" for it.
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

// placeHoldOnOrder places a hold on an order's funds in the owner's account.
func (k Keeper) placeHoldOnOrder(ctx sdk.Context, order exchange.OrderI) error {
	orderID := order.GetOrderID()
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
func (k Keeper) releaseHoldOnOrder(ctx sdk.Context, order exchange.OrderI) error {
	orderID := order.GetOrderID()
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

// GetOrder gets an order. Returns nil, nil if the order does not exist.
func (k Keeper) GetOrder(ctx sdk.Context, orderID uint64) (*exchange.Order, error) {
	return k.getOrderFromStore(k.getStore(ctx), orderID)
}

// GetOrderByExternalID gets an order by its market id and UUID.
func (k Keeper) GetOrderByExternalID(ctx sdk.Context, marketID uint32, externalID string) (*exchange.Order, error) {
	if marketID == 0 {
		return nil, errors.New("invalid market id: cannot be zero")
	}
	if len(externalID) == 0 || len(externalID) > exchange.MaxExternalIDLength {
		return nil, nil
	}

	store := k.getStore(ctx)
	key := MakeIndexKeyMarketExternalIDToOrder(marketID, externalID)
	orderIDBz := store.Get(key)
	orderID, ok := uint64FromBz(orderIDBz)
	if !ok {
		return nil, nil
	}

	return k.getOrderFromStore(store, orderID)
}

// CreateAskOrder creates an ask order, collects the creation fee, and places all needed holds.
func (k Keeper) CreateAskOrder(ctx sdk.Context, askOrder exchange.AskOrder, creationFee *sdk.Coin) (uint64, error) {
	if err := askOrder.Validate(); err != nil {
		return 0, err
	}

	store := k.getStore(ctx)
	marketID := askOrder.MarketId

	if err := validateMarketIsAcceptingOrders(store, marketID); err != nil {
		return 0, err
	}
	seller := sdk.MustAccAddressFromBech32(askOrder.Seller)
	if err := k.validateUserCanCreateAsk(ctx, marketID, seller); err != nil {
		return 0, err
	}
	if err := validateCreateAskFees(store, marketID, creationFee, askOrder.SellerSettlementFlatFee); err != nil {
		return 0, err
	}
	if err := validateAskPrice(store, marketID, askOrder.Price, askOrder.SellerSettlementFlatFee); err != nil {
		return 0, err
	}

	if creationFee != nil {
		err := k.CollectFee(ctx, marketID, seller, sdk.Coins{*creationFee})
		if err != nil {
			return 0, fmt.Errorf("error collecting ask order creation fee: %w", err)
		}
	}

	orderID := nextOrderID(store)
	order := exchange.NewOrder(orderID).WithAsk(&askOrder)
	if err := k.setOrderInStore(store, *order); err != nil {
		return 0, fmt.Errorf("error storing ask order: %w", err)
	}

	if err := k.placeHoldOnOrder(ctx, order); err != nil {
		return 0, err
	}

	k.emitEvent(ctx, exchange.NewEventOrderCreated(order))
	return orderID, nil
}

// CreateBidOrder creates a bid order, collects the creation fee, and places all needed holds.
func (k Keeper) CreateBidOrder(ctx sdk.Context, bidOrder exchange.BidOrder, creationFee *sdk.Coin) (uint64, error) {
	if err := bidOrder.Validate(); err != nil {
		return 0, err
	}

	store := k.getStore(ctx)
	marketID := bidOrder.MarketId

	if err := validateMarketIsAcceptingOrders(store, marketID); err != nil {
		return 0, err
	}
	buyer := sdk.MustAccAddressFromBech32(bidOrder.Buyer)
	if err := k.validateUserCanCreateBid(ctx, marketID, buyer); err != nil {
		return 0, err
	}
	if err := validateCreateBidFees(store, marketID, creationFee, bidOrder.Price, bidOrder.BuyerSettlementFees); err != nil {
		return 0, err
	}

	if creationFee != nil {
		err := k.CollectFee(ctx, marketID, buyer, sdk.Coins{*creationFee})
		if err != nil {
			return 0, fmt.Errorf("error collecting bid order creation fee: %w", err)
		}
	}

	orderID := nextOrderID(store)
	order := exchange.NewOrder(orderID).WithBid(&bidOrder)
	if err := k.setOrderInStore(store, *order); err != nil {
		return 0, fmt.Errorf("error storing bid order: %w", err)
	}

	if err := k.placeHoldOnOrder(ctx, order); err != nil {
		return 0, err
	}

	k.emitEvent(ctx, exchange.NewEventOrderCreated(order))
	return orderID, nil
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
	if signer != orderOwner && !k.CanCancelOrdersForMarket(ctx, order.GetMarketID(), signer) {
		return fmt.Errorf("account %s does not have permission to cancel order %d", signer, orderID)
	}

	orderOwnerAddr := sdk.MustAccAddressFromBech32(orderOwner)
	heldAmount := order.GetHoldAmount()
	err = k.holdKeeper.ReleaseHold(ctx, orderOwnerAddr, heldAmount)
	if err != nil {
		return fmt.Errorf("unable to release hold on order %d funds: %w", orderID, err)
	}

	deleteAndDeIndexOrder(k.getStore(ctx), *order)
	k.emitEvent(ctx, exchange.NewEventOrderCancelled(order, signer))

	return nil
}

// SetOrderExternalID updates an order's external id.
// The caller is responsible for making sure this update should be allowed (e.g. by calling CanSetIDs first).
func (k Keeper) SetOrderExternalID(ctx sdk.Context, marketID uint32, orderID uint64, newExternalID string) error {
	if err := exchange.ValidateExternalID(newExternalID); err != nil {
		return err
	}

	store := k.getStore(ctx)
	order, err := k.getOrderFromStore(store, orderID)
	if err != nil {
		return err
	}
	if order == nil {
		return fmt.Errorf("order %d not found", orderID)
	}

	orderMarketID := order.GetMarketID()
	if marketID != orderMarketID {
		return fmt.Errorf("order %d has market id %d, expected %d", orderID, orderMarketID, marketID)
	}

	orderExternalID := order.GetExternalID()
	if orderExternalID == newExternalID {
		return fmt.Errorf("order %d already has external id %q", orderID, orderExternalID)
	}

	switch {
	case order.IsAskOrder():
		order.GetAskOrder().ExternalId = newExternalID
	case order.IsBidOrder():
		order.GetBidOrder().ExternalId = newExternalID
	default:
		return fmt.Errorf("order %d has unexpected type %s", orderID, order.GetOrderType())
	}

	if len(orderExternalID) > 0 {
		oldIDIndex := MakeIndexKeyMarketExternalIDToOrder(orderMarketID, orderExternalID)
		store.Delete(oldIDIndex)
	}

	err = k.setOrderInStore(store, *order)
	if err != nil {
		return err
	}

	k.emitEvent(ctx, exchange.NewEventOrderExternalIDUpdated(order))
	return nil
}

// IterateOrders iterates over all orders. An error is returned if there was a problem
// reading an entry along the way. Such a problem does not interrupt iteration.
// The callback takes in the order and should return whether to stop iterating.
func (k Keeper) IterateOrders(ctx sdk.Context, cb func(order *exchange.Order) bool) error {
	var errs []error
	k.iterate(ctx, GetKeyPrefixOrder(), func(key, value []byte) bool {
		order, err := k.parseOrderStoreKeyValue(key, value)
		if err != nil {
			errs = append(errs, err)
			return false
		}
		return cb(order)
	})
	return errors.Join(errs...)
}

// IterateMarketOrders iterates over all orders for a market.
// The callback takes in the order id and order type byte and should return whether to stop iterating.
func (k Keeper) IterateMarketOrders(ctx sdk.Context, marketID uint32, cb func(orderID uint64, orderTypeByte byte) bool) {
	k.iterateOrderIndex(ctx, GetIndexKeyPrefixMarketToOrder(marketID), cb)
}

// IterateAddressOrders iterates over all orders for an address.
// The callback takes in the order id and order type byte and should return whether to stop iterating.
func (k Keeper) IterateAddressOrders(ctx sdk.Context, addr sdk.AccAddress, cb func(orderID uint64, orderTypeByte byte) bool) {
	k.iterateOrderIndex(ctx, GetIndexKeyPrefixAddressToOrder(addr), cb)
}

// IterateAssetOrders iterates over all orders for a given asset denom.
// The callback takes in the order id and order type byte and should return whether to stop iterating.
func (k Keeper) IterateAssetOrders(ctx sdk.Context, assetDenom string, cb func(orderID uint64, orderTypeByte byte) bool) {
	k.iterateOrderIndex(ctx, GetIndexKeyPrefixAssetToOrder(assetDenom), cb)
}

// CancelAllOrdersForMarket cancels all orders for a market, deleting them and releasing their holds.
func (k Keeper) CancelAllOrdersForMarket(ctx sdk.Context, marketID uint32, signer string) {
	var orderIDs []uint64
	k.IterateMarketOrders(ctx, marketID, func(orderID uint64, _ byte) bool {
		orderIDs = append(orderIDs, orderID)
		return false
	})

	var errs []error
	for _, orderID := range orderIDs {
		err := k.CancelOrder(ctx, orderID, signer)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		k.logErrorf(ctx, "%d error(s) encountered canceling all orders for market %d:\n%v",
			len(errs), marketID, errors.Join(errs...))
	}
}
