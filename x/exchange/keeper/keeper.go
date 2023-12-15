package keeper

import (
	"fmt"
	"strings"

	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"

	// "github.com/cosmos/cosmos-sdk/x/quarantine" // TODO[1760]: quarantine

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/gogoproto/proto"

	"github.com/provenance-io/provenance/x/exchange"
)

var (
	// OneInt is an sdkmath.Int of 1.
	OneInt = sdkmath.NewInt(1)
	// TenKInt is an sdkmath.Int of 10,000.
	TenKInt = sdkmath.NewInt(10_000)
)

// Keeper provides the exchange module's state store interactions.
type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey

	accountKeeper exchange.AccountKeeper
	attrKeeper    exchange.AttributeKeeper
	bankKeeper    exchange.BankKeeper
	holdKeeper    exchange.HoldKeeper
	markerKeeper  exchange.MarkerKeeper

	authority        string
	feeCollectorName string
}

func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey, feeCollectorName string,
	accountKeeper exchange.AccountKeeper, attrKeeper exchange.AttributeKeeper,
	bankKeeper exchange.BankKeeper, holdKeeper exchange.HoldKeeper, markerKeeper exchange.MarkerKeeper,
) Keeper {
	rv := Keeper{
		cdc:              cdc,
		storeKey:         storeKey,
		accountKeeper:    accountKeeper,
		attrKeeper:       attrKeeper,
		bankKeeper:       bankKeeper,
		holdKeeper:       holdKeeper,
		markerKeeper:     markerKeeper,
		authority:        authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		feeCollectorName: feeCollectorName,
	}
	return rv
}

// logErrorf uses fmt.Sprintf to combine the msg and args, and logs the result as an error from this module.
// Note that this is different from the logging .Error(msg string, keyvals ...interface{}) syntax.
func (k Keeper) logErrorf(ctx sdk.Context, msg string, args ...interface{}) {
	ctx.Logger().Error(fmt.Sprintf(msg, args...), "module", "x/"+exchange.ModuleName)
}

// logInfof uses fmt.Sprintf to combine the msg and args, and logs the result as info from this module.
// Note that this is different from the logging .Info(msg string, keyvals ...interface{}) syntax.
func (k Keeper) logInfof(ctx sdk.Context, msg string, args ...interface{}) {
	ctx.Logger().Info(fmt.Sprintf(msg, args...), "module", "x/"+exchange.ModuleName)
}

// emitEvent emits the provided event and writes any error to the error log.
// See Also emitEvents.
func (k Keeper) emitEvent(ctx sdk.Context, event proto.Message) {
	err := ctx.EventManager().EmitTypedEvent(event)
	if err != nil {
		k.logErrorf(ctx, "error emitting event %#v: %v", event, err)
	}
}

// emitEvents emits the provided events and writes any error to the error log.
// See Also emitEvent.
func (k Keeper) emitEvents(ctx sdk.Context, events []proto.Message) {
	err := ctx.EventManager().EmitTypedEvents(events...)
	if err != nil {
		k.logErrorf(ctx, "error emitting events %#v: %v", events, err)
	}
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
	defer iter.Close()

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
func iterate(store storetypes.KVStore, pre []byte, cb func(key, value []byte) bool) {
	// Using an open iterator on a prefixed store here so that iter.Key() doesn't contain the prefix.
	pStore := prefix.NewStore(store, pre)
	iter := pStore.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		if cb(iter.Key(), iter.Value()) {
			break
		}
	}
}

// getStore gets the store for the exchange module.
func (k Keeper) getStore(ctx sdk.Context) storetypes.KVStore {
	return ctx.KVStore(k.storeKey)
}

// iterate iterates over all the entries in the store with the given prefix.
// The key provided to the callback will NOT have the provided prefix; it will be everything after it.
// The callback should return false to continue iteration, or true to stop.
func (k Keeper) iterate(ctx sdk.Context, pre []byte, cb func(key, value []byte) bool) {
	iterate(k.getStore(ctx), pre, cb)
}

// DoTransfer facilitates a transfer of things using the bank module.
func (k Keeper) DoTransfer(ctxIn sdk.Context, inputs []banktypes.Input, outputs []banktypes.Output) error {
	// We bypass the quarantine module here under the assumption that someone creating
	// an order counts as acceptance of the stuff to receive (that they defined when creating the order).
	ctx := ctxIn // quarantine.WithBypass(ctxIn) // TODO[1760]: quarantine
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
		return k.bankKeeper.SendCoins(ctx, fromAddr, toAddr, inputs[0].Coins)
	}
	// TODO[1760]: exchange: Put this back once we have InputOutputCoins again.
	// return k.bankKeeper.InputOutputCoins(ctx, inputs, outputs)
	return nil
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

		splitAmt, splitRem := exchange.QuoRemInt(coin.Amount.Mul(sdkmath.NewInt(split)), TenKInt)
		if !splitRem.IsZero() {
			splitAmt = splitAmt.Add(OneInt)
		}
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
	// TODO[1760]: exchange: Put this back once we have InputOutputCoins again.
	_ = outputs
	// if err := k.bankKeeper.InputOutputCoins(ctx, inputs, outputs); err != nil {
	// 	return fmt.Errorf("error collecting fees for market %d: %w", marketID, err)
	// }
	if !exchangeAmt.IsZero() {
		if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, marketAddr, k.feeCollectorName, exchangeAmt); err != nil {
			return fmt.Errorf("error collecting exchange fee %s (based off %s) from market %d: %w", exchangeAmt, feeAmt, marketID, err)
		}
	}

	return nil
}
