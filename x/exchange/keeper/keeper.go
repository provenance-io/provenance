package keeper

import (
	"fmt"
	"strings"

	"cosmossdk.io/collections"
	collcodec "cosmossdk.io/collections/codec"
	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/gogoproto/proto"

	"github.com/provenance-io/provenance/x/exchange"
)

var (
	// TenKInt is an sdkmath.Int of 10,000.
	TenKInt = sdkmath.NewInt(10_000)
	// TwentyKInt is an sdkmath.Int of 20,000.
	TwentyKInt = sdkmath.NewInt(20_000)
)

// Keeper provides the exchange module's state store interactions.
type Keeper struct {
	cdc          codec.BinaryCodec
	storeService store.KVStoreService

	accountKeeper  exchange.AccountKeeper
	attrKeeper     exchange.AttributeKeeper
	bankKeeper     exchange.BankKeeper
	holdKeeper     exchange.HoldKeeper
	markerKeeper   exchange.MarkerKeeper
	metadataKeeper exchange.MetadataKeeper

	authority        string
	feeCollectorName string

	// The exchange module's collections. See the comment at the top of keys.go for the layout of each one.
	schema collections.Schema

	// params is the module's params.
	params collections.Item[exchange.Params]
	// lastMarketID is the last auto-selected market id.
	lastMarketID collections.Item[uint32]
	// knownMarketIDs is the set of ids of all the markets that exist.
	knownMarketIDs collections.KeySet[uint32]
	// lastOrderID is the id of the last order created.
	lastOrderID collections.Item[uint64]

	// marketCreateAskFlatFees maps (market id, denom) to the amount of the create-ask flat fee in that denom.
	marketCreateAskFlatFees collections.Map[collections.Pair[uint32, string], sdkmath.Int]
	// marketCreateBidFlatFees maps (market id, denom) to the amount of the create-bid flat fee in that denom.
	marketCreateBidFlatFees collections.Map[collections.Pair[uint32, string], sdkmath.Int]
	// marketSellerSettlementFlatFees maps (market id, denom) to the amount of the seller settlement flat fee in that denom.
	marketSellerSettlementFlatFees collections.Map[collections.Pair[uint32, string], sdkmath.Int]
	// marketSellerSettlementRatios maps (market id, price denom, fee denom) to a seller settlement fee ratio.
	marketSellerSettlementRatios collections.Map[collections.Triple[uint32, string, string], exchange.FeeRatio]
	// marketBuyerSettlementFlatFees maps (market id, denom) to the amount of the buyer settlement flat fee in that denom.
	marketBuyerSettlementFlatFees collections.Map[collections.Pair[uint32, string], sdkmath.Int]
	// marketBuyerSettlementRatios maps (market id, price denom, fee denom) to a buyer settlement fee ratio.
	marketBuyerSettlementRatios collections.Map[collections.Triple[uint32, string, string], exchange.FeeRatio]
	// marketNotAcceptingOrders is the set of ids of the markets that are not accepting orders.
	marketNotAcceptingOrders collections.KeySet[uint32]
	// marketUserSettle is the set of ids of the markets that allow user settlement.
	marketUserSettle collections.KeySet[uint32]
	// marketPermissions is the set of (market id, address, permission) permissions granted.
	marketPermissions collections.KeySet[collections.Triple[uint32, sdk.AccAddress, exchange.Permission]]
	// marketReqAttrsAsk maps a market id to the attributes required to create an ask order in it.
	marketReqAttrsAsk collections.Map[uint32, []string]
	// marketReqAttrsBid maps a market id to the attributes required to create a bid order in it.
	marketReqAttrsBid collections.Map[uint32, []string]
	// marketReqAttrsCommitment maps a market id to the attributes required to create a commitment in it.
	marketReqAttrsCommitment collections.Map[uint32, []string]
	// marketAcceptingCommitments is the set of ids of the markets that are accepting commitments.
	marketAcceptingCommitments collections.KeySet[uint32]
	// marketCreateCommitmentFlatFees maps (market id, denom) to the amount of the create-commitment flat fee in that denom.
	marketCreateCommitmentFlatFees collections.Map[collections.Pair[uint32, string], sdkmath.Int]
	// marketCommitmentSettlementBips maps a market id to the bips it is charged for commitment settlements.
	marketCommitmentSettlementBips collections.Map[uint32, uint32]
	// marketIntermediaryDenom maps a market id to the intermediary denom it uses in commitment fee calculations.
	marketIntermediaryDenom collections.Map[uint32, string]

	// orders maps an order id to the order. The order's OrderId is not stored in the value; it's the key.
	orders collections.Map[uint64, exchange.Order]
	// marketToOrderIndex maps (market id, order id) to the order's type byte.
	marketToOrderIndex collections.Map[collections.Pair[uint32, uint64], byte]
	// addressToOrderIndex maps (owner address, order id) to the order's type byte.
	addressToOrderIndex collections.Map[collections.Pair[sdk.AccAddress, uint64], byte]
	// assetToOrderIndex maps (asset denom, order id) to the order's type byte.
	assetToOrderIndex collections.Map[collections.Pair[string, uint64], byte]
	// marketExternalIDToOrderIndex maps (market id, external id) to an order id.
	marketExternalIDToOrderIndex collections.Map[collections.Pair[uint32, string], uint64]

	// commitments maps (market id, address) to the amount the address has committed to the market.
	commitments collections.Map[collections.Pair[uint32, sdk.AccAddress], sdk.Coins]

	// payments maps (source address, external id) to the payment.
	payments collections.Map[collections.Pair[sdk.AccAddress, string], exchange.Payment]
	// targetToPaymentIndex is the set of (target address, source address, external id) of payments with a target.
	targetToPaymentIndex collections.KeySet[collections.Triple[sdk.AccAddress, sdk.AccAddress, string]]
}

func NewKeeper(cdc codec.BinaryCodec, storeService store.KVStoreService, feeCollectorName string,
	accountKeeper exchange.AccountKeeper, attrKeeper exchange.AttributeKeeper,
	bankKeeper exchange.BankKeeper, holdKeeper exchange.HoldKeeper, markerKeeper exchange.MarkerKeeper,
	metadataKeeper exchange.MetadataKeeper,
) Keeper {
	sb := collections.NewSchemaBuilder(storeService)
	marketDenomKey := collections.PairKeyCodec(collections.Uint32Key, collections.StringKey)
	marketRatioKey := collections.TripleKeyCodec(collections.Uint32Key, collections.StringKey, collections.StringKey)

	rv := Keeper{
		cdc:              cdc,
		storeService:     storeService,
		accountKeeper:    accountKeeper,
		attrKeeper:       attrKeeper,
		bankKeeper:       bankKeeper,
		holdKeeper:       holdKeeper,
		markerKeeper:     markerKeeper,
		metadataKeeper:   metadataKeeper,
		authority:        authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		feeCollectorName: feeCollectorName,

		params: collections.NewItem(sb, ParamsPrefix, "params",
			codec.CollValue[exchange.Params](cdc)),
		lastMarketID: collections.NewItem(sb, LastMarketIDPrefix, "last_market_id",
			collections.Uint32Value),
		knownMarketIDs: collections.NewKeySet(sb, KnownMarketIDsPrefix, "known_market_ids",
			collections.Uint32Key),
		lastOrderID: collections.NewItem(sb, LastOrderIDPrefix, "last_order_id",
			collections.Uint64Value),

		marketCreateAskFlatFees: collections.NewMap(sb, MarketCreateAskFlatFeesPrefix, "market_create_ask_flat_fees",
			marketDenomKey, sdk.IntValue),
		marketCreateBidFlatFees: collections.NewMap(sb, MarketCreateBidFlatFeesPrefix, "market_create_bid_flat_fees",
			marketDenomKey, sdk.IntValue),
		marketSellerSettlementFlatFees: collections.NewMap(sb, MarketSellerSettlementFlatFeesPrefix, "market_seller_settlement_flat_fees",
			marketDenomKey, sdk.IntValue),
		marketSellerSettlementRatios: collections.NewMap(sb, MarketSellerSettlementRatiosPrefix, "market_seller_settlement_ratios",
			marketRatioKey, codec.CollValue[exchange.FeeRatio](cdc)),
		marketBuyerSettlementFlatFees: collections.NewMap(sb, MarketBuyerSettlementFlatFeesPrefix, "market_buyer_settlement_flat_fees",
			marketDenomKey, sdk.IntValue),
		marketBuyerSettlementRatios: collections.NewMap(sb, MarketBuyerSettlementRatiosPrefix, "market_buyer_settlement_ratios",
			marketRatioKey, codec.CollValue[exchange.FeeRatio](cdc)),
		marketNotAcceptingOrders: collections.NewKeySet(sb, MarketNotAcceptingOrdersPrefix, "market_not_accepting_orders",
			collections.Uint32Key),
		marketUserSettle: collections.NewKeySet(sb, MarketUserSettlePrefix, "market_user_settle",
			collections.Uint32Key),
		marketPermissions: collections.NewKeySet(sb, MarketPermissionsPrefix, "market_permissions",
			collections.TripleKeyCodec(collections.Uint32Key, sdk.AccAddressKey, collcodec.NewInt32Key[exchange.Permission]())),
		marketReqAttrsAsk: collections.NewMap(sb, MarketReqAttrsAskPrefix, "market_req_attrs_ask",
			collections.Uint32Key, exchange.StringSliceValue),
		marketReqAttrsBid: collections.NewMap(sb, MarketReqAttrsBidPrefix, "market_req_attrs_bid",
			collections.Uint32Key, exchange.StringSliceValue),
		marketReqAttrsCommitment: collections.NewMap(sb, MarketReqAttrsCommitmentPrefix, "market_req_attrs_commitment",
			collections.Uint32Key, exchange.StringSliceValue),
		marketAcceptingCommitments: collections.NewKeySet(sb, MarketAcceptingCommitmentsPrefix, "market_accepting_commitments",
			collections.Uint32Key),
		marketCreateCommitmentFlatFees: collections.NewMap(sb, MarketCreateCommitmentFlatFeesPrefix, "market_create_commitment_flat_fees",
			marketDenomKey, sdk.IntValue),
		marketCommitmentSettlementBips: collections.NewMap(sb, MarketCommitmentSettlementBipsPrefix, "market_commitment_settlement_bips",
			collections.Uint32Key, collections.Uint32Value),
		marketIntermediaryDenom: collections.NewMap(sb, MarketIntermediaryDenomPrefix, "market_intermediary_denom",
			collections.Uint32Key, collections.StringValue),

		orders: collections.NewMap(sb, OrdersPrefix, "orders",
			collections.Uint64Key, exchange.NewOrderValueCodec(cdc)),
		marketToOrderIndex: collections.NewMap(sb, MarketToOrderIndexPrefix, "market_to_order_index",
			collections.PairKeyCodec(collections.Uint32Key, collections.Uint64Key), exchange.ByteValue),
		addressToOrderIndex: collections.NewMap(sb, AddressToOrderIndexPrefix, "address_to_order_index",
			collections.PairKeyCodec(sdk.AccAddressKey, collections.Uint64Key), exchange.ByteValue),
		assetToOrderIndex: collections.NewMap(sb, AssetToOrderIndexPrefix, "asset_to_order_index",
			collections.PairKeyCodec(collections.StringKey, collections.Uint64Key), exchange.ByteValue),
		marketExternalIDToOrderIndex: collections.NewMap(sb, MarketExternalIDToOrderIndexPrefix, "market_external_id_to_order_index",
			collections.PairKeyCodec(collections.Uint32Key, collections.StringKey), collections.Uint64Value),

		commitments: collections.NewMap(sb, CommitmentsPrefix, "commitments",
			collections.PairKeyCodec(collections.Uint32Key, sdk.AccAddressKey), exchange.CoinsValue),

		payments: collections.NewMap(sb, PaymentsPrefix, "payments",
			collections.PairKeyCodec(sdk.AccAddressKey, collections.StringKey), codec.CollValue[exchange.Payment](cdc)),
		targetToPaymentIndex: collections.NewKeySet(sb, TargetToPaymentIndexPrefix, "target_to_payment_index",
			collections.TripleKeyCodec(sdk.AccAddressKey, sdk.AccAddressKey, collections.StringKey)),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(fmt.Errorf("failed to build the exchange collections schema: %w", err))
	}
	rv.schema = schema

	return rv
}

// getLogger gets a logger for the exchange module.
func (k Keeper) getLogger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+exchange.ModuleName)
}

// logEndpointError logs an error for an endpoint.
// This uses the standard key+val logging argument pattern instead of the fmt.Sprintf format.
func (k Keeper) logEndpointError(ctx sdk.Context, endpoint, msg string, keyVals ...interface{}) {
	k.getLogger(ctx).With("endpoint", endpoint).Error(msg, keyVals...)
}

// logErrorf uses fmt.Sprintf to combine the msg and args, and logs the result as an error from this module.
// Note that this is different from the logging .Error(msg string, keyvals ...interface{}) syntax.
func (k Keeper) logErrorf(ctx sdk.Context, msg string, args ...interface{}) {
	k.getLogger(ctx).Error(fmt.Sprintf(msg, args...))
}

// logInfof uses fmt.Sprintf to combine the msg and args, and logs the result as info from this module.
// Note that this is different from the logging .Info(msg string, keyvals ...interface{}) syntax.
func (k Keeper) logInfof(ctx sdk.Context, msg string, args ...interface{}) {
	k.getLogger(ctx).Info(fmt.Sprintf(msg, args...))
}

// emitEvent emits the provided event and writes any error to the error log.
// If you have multiple events to emit, consider using emitEvents.
func (k Keeper) emitEvent(ctx sdk.Context, event proto.Message) {
	err := ctx.EventManager().EmitTypedEvent(event)
	if err != nil {
		k.logErrorf(ctx, "error emitting event %#v: %v", event, err)
	}
}

// emitEvents emits the provided events and writes any error to the error log.
// If you only have one event to emit, consider using emitEvent.
// If your events slice is typed to a specific event type (or something other than exactly []proto.Message),
// use the non-keeper emitEvents(k, ctx, events) function instead.
func (k Keeper) emitEvents(ctx sdk.Context, events []proto.Message) {
	err := ctx.EventManager().EmitTypedEvents(events...)
	if err != nil {
		k.logErrorf(ctx, "error emitting events %#v: %v", events, err)
	}
}

// emitEvents emits the provided events and writes any error to the error log.
// If you only have one event to emit, consider using k.emitEvent.
// The difference between this and k.emitEvents is that this will accept a slice of
// specifically typed events instead of needing to be exactly a []proto.Message slice.
// E.g. events can be provided here as a []*exchange.EventPaymentRejected.
func emitEvents[S ~[]E, E proto.Message](k Keeper, ctx sdk.Context, events S) {
	e2 := make([]proto.Message, len(events))
	for i, event := range events {
		e2[i] = event
	}
	k.emitEvents(ctx, e2)
}

// GetAuthority gets the address (as bech32) that has governance authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// IsAuthority returns true if the provided address bech32 string is the authority address.
func (k Keeper) IsAuthority(addr string) bool {
	return strings.EqualFold(k.authority, addr)
}

// ValidateAuthority returns an error if the provided address is not the authority.
func (k Keeper) ValidateAuthority(addr string) error {
	if !k.IsAuthority(addr) {
		return govtypes.ErrInvalidSigner.Wrapf("expected %q got %q", k.GetAuthority(), addr)
	}
	return nil
}

// GetFeeCollectorName gets the name of the fee collector.
func (k Keeper) GetFeeCollectorName() string {
	return k.feeCollectorName
}

// getAllKeys gets all the keys in the store with the given prefix.
func getAllKeys(store storetypes.KVStore, pre []byte) [][]byte {
	// Using a prefix iterator so that iter.Key() is the whole key (including the prefix).
	iter := storetypes.KVStorePrefixIterator(store, pre)
	defer iter.Close() //nolint:errcheck // ignoring close error on iterator: not critical for this context.

	var keys [][]byte
	for ; iter.Valid(); iter.Next() {
		keys = append(keys, iter.Key())
	}

	return keys
}

// deleteAll deletes all keys that have the given prefix.
func deleteAll(store storetypes.KVStore, pre []byte) {
	keys := getAllKeys(store, pre)
	for _, key := range keys {
		store.Delete(key)
	}
}

// iterate iterates over all the entries in the store with the given prefix.
// The key provided to the callback will NOT have the provided prefix; it will be everything after it.
// The callback should return false to continue iteration, or true to stop.
func iterate(store storetypes.KVStore, keyPrefix []byte, cb func(keySuffix, value []byte) bool) {
	// Using an open iterator on a prefixed store here so that iter.Key() doesn't contain the prefix.
	var iter storetypes.Iterator
	if len(keyPrefix) > 0 {
		pStore := prefix.NewStore(store, keyPrefix)
		iter = pStore.Iterator(nil, nil)
	} else {
		iter = store.Iterator(nil, nil)
	}
	defer iter.Close() //nolint:errcheck // ignoring close error on iterator: not critical for this context.

	for ; iter.Valid(); iter.Next() {
		if cb(iter.Key(), iter.Value()) {
			break
		}
	}
}

// getStore gets the store for the exchange module.
func (k Keeper) getStore(ctx sdk.Context) storetypes.KVStore {
	return runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
}

// iterate iterates over all the entries in the store with the given prefix.
// The key provided to the callback will NOT have the provided prefix; it will be everything after it.
// The callback should return false to continue iteration, or true to stop.
func (k Keeper) iterate(ctx sdk.Context, keyPrefix []byte, cb func(keySuffix, value []byte) bool) {
	iterate(k.getStore(ctx), keyPrefix, cb)
}

// DoTransfer facilitates a transfer of things using the bank module.
func (k Keeper) DoTransfer(ctx sdk.Context, inputs []banktypes.Input, outputs []banktypes.Output) error {
	if len(inputs) == 1 && len(outputs) == 1 {
		// If there's only one of each, we use SendCoins for the nicer events.
		if !inputs[0].Coins.Equal(outputs[0].Coins) {
			return fmt.Errorf("input coins %q does not equal output coins %q",
				inputs[0].Coins, outputs[0].Coins)
		}
		fromAddr, err := sdk.AccAddressFromBech32(inputs[0].Address)
		if err != nil {
			return fmt.Errorf("invalid inputs[0] address %q: %w", inputs[0].Address, err)
		}
		toAddr, err := sdk.AccAddressFromBech32(outputs[0].Address)
		if err != nil {
			return fmt.Errorf("invalid outputs[0] address %q: %w", outputs[0].Address, err)
		}
		if k.bankKeeper.BlockedAddr(toAddr) {
			return fmt.Errorf("%s is not allowed to receive funds", toAddr)
		}
		return k.bankKeeper.SendCoins(ctx, fromAddr, toAddr, inputs[0].Coins)
	}

	for _, output := range outputs {
		toAddr, err := sdk.AccAddressFromBech32(output.Address)
		if err == nil && k.bankKeeper.BlockedAddr(toAddr) {
			return fmt.Errorf("%s is not allowed to receive funds", toAddr)
		}
	}

	return k.bankKeeper.InputOutputCoinsProv(ctx, inputs, outputs)
}

// CalculateExchangeSplit calculates the amount that the exchange will keep of the provided fee.
func (k Keeper) CalculateExchangeSplit(ctx sdk.Context, feeAmt sdk.Coins) sdk.Coins {
	if feeAmt.IsZero() {
		return nil
	}
	exchangeAmt := make(sdk.Coins, 0, len(feeAmt))
	for _, coin := range feeAmt {
		if coin.Amount.IsZero() {
			continue
		}

		split := int64(k.GetExchangeSplit(ctx, coin.Denom))
		if split == 0 {
			continue
		}

		splitAmt := exchange.QuoIntRoundUp(coin.Amount.Mul(sdkmath.NewInt(split)), TenKInt)
		exchangeAmt = append(exchangeAmt, sdk.NewCoin(coin.Denom, splitAmt))
	}
	if exchangeAmt.IsZero() {
		return nil
	}
	return exchangeAmt
}

// CollectFee will transfer the fee amount to the market account,
// then the exchange's cut from the market to the fee collector.
// If you have fees to collect from multiple payers, consider using CollectFees.
func (k Keeper) CollectFee(ctx sdk.Context, marketID uint32, payer sdk.AccAddress, fee sdk.Coins) error {
	if fee.IsZero() {
		return nil
	}
	exchangeSplit := k.CalculateExchangeSplit(ctx, fee)

	marketAddr := exchange.GetMarketAddress(marketID)
	if err := k.bankKeeper.SendCoins(ctx, payer, marketAddr, fee); err != nil {
		return fmt.Errorf("error transferring %s from %s to market %d: %w", fee, payer, marketID, err)
	}
	if !exchangeSplit.IsZero() {
		if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, marketAddr, k.feeCollectorName, exchangeSplit); err != nil {
			return fmt.Errorf("error collecting exchange fee %s (based off %s) from market %d: %w", exchangeSplit, fee, marketID, err)
		}
	}

	return nil
}

// CollectFees will transfer the inputs to the market account,
// then the exchange's cut from the market to the fee collector.
// If there is only one input, CollectFee is used.
func (k Keeper) CollectFees(ctx sdk.Context, marketID uint32, inputs []banktypes.Input) error {
	if len(inputs) == 0 {
		return nil
	}
	if len(inputs) == 1 {
		// If there's only one input, just use CollectFee for the nicer events.
		payer, err := sdk.AccAddressFromBech32(inputs[0].Address)
		if err != nil {
			return fmt.Errorf("invalid inputs[0] address address %q: %w", inputs[0].Address, err)
		}
		return k.CollectFee(ctx, marketID, payer, inputs[0].Coins)
	}

	var feeAmt sdk.Coins
	for _, input := range inputs {
		feeAmt = feeAmt.Add(input.Coins...)
	}
	if feeAmt.IsZero() {
		return nil
	}

	exchangeAmt := k.CalculateExchangeSplit(ctx, feeAmt)

	marketAddr := exchange.GetMarketAddress(marketID)
	outputs := []banktypes.Output{{Address: marketAddr.String(), Coins: feeAmt}}
	if err := k.bankKeeper.InputOutputCoinsProv(ctx, inputs, outputs); err != nil {
		return fmt.Errorf("error collecting fees for market %d: %w", marketID, err)
	}
	if !exchangeAmt.IsZero() {
		if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, marketAddr, k.feeCollectorName, exchangeAmt); err != nil {
			return fmt.Errorf("error collecting exchange fee %s (based off %s) from market %d: %w", exchangeAmt, feeAmt, marketID, err)
		}
	}

	return nil
}
