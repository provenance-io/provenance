package keeper

import (
	"errors"
	"fmt"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/provenance-io/provenance/x/exchange"
)

type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey

	accountKeeper exchange.AccountKeeper
	attrKeeper    exchange.AttributeKeeper
	bankKeeper    exchange.BankKeeper
	holdKeeper    exchange.HoldKeeper
	nameKeeper    exchange.NameKeeper

	// TODO[1658]: Finish the Keeper struct.
	authority        string
	feeCollectorName string
}

func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey, feeCollectorName string,
	accountKeeper exchange.AccountKeeper, attrKeeper exchange.AttributeKeeper,
	bankKeeper exchange.BankKeeper, holdKeeper exchange.HoldKeeper, nameKeeper exchange.NameKeeper,
) Keeper {
	// TODO[1658]: Finish NewKeeper.
	rv := Keeper{
		cdc:              cdc,
		storeKey:         storeKey,
		accountKeeper:    accountKeeper,
		attrKeeper:       attrKeeper,
		bankKeeper:       bankKeeper,
		holdKeeper:       holdKeeper,
		nameKeeper:       nameKeeper,
		authority:        authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		feeCollectorName: feeCollectorName,
	}
	return rv
}

// GetAuthority gets the address (as bech32) that has governance authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// GetFeeCollectorName gets the name of the fee collector.
func (k Keeper) GetFeeCollectorName() string {
	return k.feeCollectorName
}

// getAllKeys gets all the keys in the store with the given prefix.
func getAllKeys(store sdk.KVStore, pre []byte) [][]byte {
	// Using a prefix iterator so that iter.Key() is the whole key (including the prefix).
	iter := sdk.KVStorePrefixIterator(store, pre)
	defer iter.Close()

	var keys [][]byte
	for ; iter.Valid(); iter.Next() {
		keys = append(keys, iter.Key())
	}

	return keys
}

// deleteAll deletes all keys that have the given prefix.
func deleteAll(store sdk.KVStore, pre []byte) {
	keys := getAllKeys(store, pre)
	for _, key := range keys {
		store.Delete(key)
	}
}

// iterate iterates over all the entries in the store with the given prefix.
// The key provided to the callback will NOT have the provided prefix; it will be everything after it.
// The callback should return false to continue iteration, or true to stop.
func iterate(store sdk.KVStore, pre []byte, cb func(key, value []byte) bool) {
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
func (k Keeper) getStore(ctx sdk.Context) sdk.KVStore {
	return ctx.KVStore(k.storeKey)
}

// iterate iterates over all the entries in the store with the given prefix.
// The key provided to the callback will NOT have the provided prefix; it will be everything after it.
// The callback should return false to continue iteration, or true to stop.
func (k Keeper) iterate(ctx sdk.Context, pre []byte, cb func(key, value []byte) bool) {
	iterate(k.getStore(ctx), pre, cb)
}

// NormalizeReqAttrs normalizes/validates each of the provided require attributes.
func (k Keeper) NormalizeReqAttrs(ctx sdk.Context, reqAttrs []string) ([]string, error) {
	rv := make([]string, len(reqAttrs))
	errs := make([]error, len(reqAttrs))
	for i, attr := range reqAttrs {
		rv[i], errs[i] = k.nameKeeper.Normalize(ctx, attr)
	}
	return rv, errors.Join(errs...)
}

// CollectFee will transfer the fee amount to the market account,
// then the exchange's cut from the market to the fee collector.
func (k Keeper) CollectFee(ctx sdk.Context, payer sdk.AccAddress, marketID uint32, feeAmt sdk.Coins) error {
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

		splitAmt := coin.Amount.Mul(sdkmath.NewInt(split))
		roundUp := !splitAmt.ModRaw(10_000).IsZero()
		splitAmt = splitAmt.QuoRaw(10_000)
		if roundUp {
			splitAmt = splitAmt.Add(sdkmath.OneInt())
		}
		exchangeAmt = append(exchangeAmt, sdk.Coin{Denom: coin.Denom, Amount: splitAmt})
	}

	marketAddr := exchange.GetMarketAddress(marketID)
	if err := k.bankKeeper.SendCoins(ctx, payer, marketAddr, feeAmt); err != nil {
		return fmt.Errorf("error transferring %s from %s to market %d: %w", feeAmt, payer, marketID, err)
	}
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, marketAddr, k.feeCollectorName, exchangeAmt); err != nil {
		return fmt.Errorf("error collecting exchange fee %s (based off %s) from market %d: %w", exchangeAmt, feeAmt, marketID, err)
	}

	return nil
}
