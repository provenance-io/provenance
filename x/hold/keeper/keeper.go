package keeper

import (
	"errors"
	"fmt"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/hold"
)

type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey

	bankKeeper hold.BankKeeper
}

func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey, bankKeeper hold.BankKeeper) Keeper {
	rv := Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		bankKeeper: bankKeeper,
	}
	bankKeeper.AppendLockedCoinsGetter(rv.GetLockedCoins)
	return rv
}

// setHoldCoinAmount updates the store with the provided hold info.
// If the amount is zero, the hold coin entry for addr+denom is deleted.
// Otherwise, the hold coin entry for addr+denom is created/updated in the provided amount.
func (k Keeper) setHoldCoinAmount(store sdk.KVStore, addr sdk.AccAddress, denom string, amount sdkmath.Int) error {
	if len(denom) == 0 {
		return fmt.Errorf("cannot store hold with an empty denom for %s", addr)
	}
	if amount.IsNegative() {
		return fmt.Errorf("cannot store negative hold amount %s%s for %s", amount, denom, addr)
	}

	key := CreateHoldCoinKey(addr, denom)
	if amount.IsZero() {
		store.Delete(key)
		return nil
	}

	amountBz, err := amount.Marshal()
	if err != nil {
		return err
	}
	store.Set(key, amountBz)
	return nil
}

// getHoldCoinAmount gets (from the store) the amount marked as on hold for the given address and denom.
func (k Keeper) getHoldCoinAmount(store sdk.KVStore, addr sdk.AccAddress, denom string) (sdkmath.Int, error) {
	key := CreateHoldCoinKey(addr, denom)
	amountBz := store.Get(key)
	return UnmarshalHoldCoinValue(amountBz)
}

// ValidateNewHold checks the account's spendable balance to make sure it has at least as much as the funds provided.
func (k Keeper) ValidateNewHold(ctx sdk.Context, addr sdk.AccAddress, funds sdk.Coins) error {
	if funds.IsZero() {
		return nil
	}
	if funds.IsAnyNegative() {
		return fmt.Errorf("hold amounts %q for %s cannot be negative", funds, addr)
	}

	// Not bypassing hold's locked coins here because we're testing about new funds to be put on hold.
	spendable := k.bankKeeper.SpendableCoins(ctx, addr)
	for _, toAdd := range funds {
		if toAdd.IsZero() {
			continue
		}
		has, available := spendable.Find(toAdd.Denom)
		if !has {
			return fmt.Errorf("account %s spendable balance 0%s is less than hold amount %s", addr, toAdd.Denom, toAdd)
		}
		if available.Amount.LT(toAdd.Amount) {
			return fmt.Errorf("account %s spendable balance %s is less than hold amount %s", addr, available, toAdd)
		}
	}

	return nil
}

// AddHold puts the provided funds on hold for the provided account.
func (k Keeper) AddHold(ctx sdk.Context, addr sdk.AccAddress, funds sdk.Coins, reason string) error {
	if funds.IsZero() {
		return nil
	}

	if err := k.ValidateNewHold(ctx, addr, funds); err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	var fundsAdded sdk.Coins
	var errs []error
	for _, toAdd := range funds {
		if toAdd.IsZero() {
			continue
		}
		onHold, err := k.getHoldCoinAmount(store, addr, toAdd.Denom)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to get current %s hold amount for %s: %w", toAdd.Denom, addr, err))
			continue
		}
		newHoldAmt := onHold.Add(toAdd.Amount)
		err = k.setHoldCoinAmount(store, addr, toAdd.Denom, newHoldAmt)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to place %s on hold for %s: %w", toAdd, addr, err))
		}
		fundsAdded = fundsAdded.Add(toAdd)
	}

	if !fundsAdded.IsZero() {
		err := ctx.EventManager().EmitTypedEvent(hold.NewEventHoldAdded(addr, fundsAdded, reason))
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// ReleaseHold releases the hold on the provided funds for the provided account.
func (k Keeper) ReleaseHold(ctx sdk.Context, addr sdk.AccAddress, funds sdk.Coins) error {
	if funds.IsZero() {
		return nil
	}
	if funds.IsAnyNegative() {
		return fmt.Errorf("cannot release %q from hold for %s: amounts cannot be negative", funds, addr)
	}

	store := ctx.KVStore(k.storeKey)
	var fundsReleased sdk.Coins
	var errs []error
	for _, toRelease := range funds {
		if toRelease.IsZero() {
			continue
		}

		onHold, err := k.getHoldCoinAmount(store, addr, toRelease.Denom)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to get current %s hold amount for %s: %w", toRelease.Denom, addr, err))
			continue
		}

		newAmount := onHold.Sub(toRelease.Amount)
		if newAmount.IsNegative() {
			errs = append(errs, fmt.Errorf("cannot release %s from hold for %s: account only has %s%s on hold", toRelease, addr, onHold, toRelease.Denom))
			continue
		}

		err = k.setHoldCoinAmount(store, addr, toRelease.Denom, newAmount)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to release %s from hold for %s: %w", toRelease, addr, err))
			continue
		}

		fundsReleased = fundsReleased.Add(toRelease)
	}

	if !fundsReleased.IsZero() {
		err := ctx.EventManager().EmitTypedEvent(hold.NewEventHoldReleased(addr, fundsReleased))
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// GetHoldCoin gets the amount of a denom on hold for a given account.
// Will return a zero Coin of the given denom if the store does not have an entry for it.
func (k Keeper) GetHoldCoin(ctx sdk.Context, addr sdk.AccAddress, denom string) (sdk.Coin, error) {
	var err error
	rv := sdk.Coin{Denom: denom}
	rv.Amount, err = k.getHoldCoinAmount(ctx.KVStore(k.storeKey), addr, denom)
	if err != nil {
		return rv, fmt.Errorf("could not get hold coin amount for %s: %w", addr, err)
	}
	return rv, nil
}

// GetHoldCoins gets all funds on hold for a given account.
func (k Keeper) GetHoldCoins(ctx sdk.Context, addr sdk.AccAddress) (sdk.Coins, error) {
	var rv sdk.Coins
	err := k.IterateHolds(ctx, addr, func(coin sdk.Coin) bool {
		rv = rv.Add(coin)
		return false
	})

	return rv, err
}

// getHoldCoinPrefixStore returns a kv store prefixed for hold coin entries for the provided address.
func (k Keeper) getHoldCoinPrefixStore(ctx sdk.Context, addr sdk.AccAddress) sdk.KVStore {
	pre := CreateHoldCoinKeyAddrPrefix(addr)
	return prefix.NewStore(ctx.KVStore(k.storeKey), pre)
}

// IterateHolds iterates over all funds on hold for a given account.
// The process function should return whether to stop: false = keep iterating, true = stop.
// If an error is encountered while reading from the store, that entry is skipped and an error is
// returned for it when iteration is completed.
func (k Keeper) IterateHolds(ctx sdk.Context, addr sdk.AccAddress, process func(sdk.Coin) bool) error {
	store := k.getHoldCoinPrefixStore(ctx, addr)

	iter := store.Iterator(nil, nil)
	defer iter.Close()

	var errs []error
	for ; iter.Valid(); iter.Next() {
		key := iter.Key()
		value := iter.Value()

		denom := string(key)
		amount, err := UnmarshalHoldCoinValue(value)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to read amount of %s for account %s: %w", denom, addr, err))
			continue
		}

		if process(sdk.Coin{Denom: denom, Amount: amount}) {
			break
		}
	}

	return errors.Join(errs...)
}

// getAllHoldCoinPrefixStore returns a kv store prefixed for all hold coin entries.
func (k Keeper) getAllHoldCoinPrefixStore(ctx sdk.Context) sdk.KVStore {
	return prefix.NewStore(ctx.KVStore(k.storeKey), KeyPrefixHoldCoin)
}

// IterateAllHolds iterates over all hold coin entries for all accounts.
// The process function should return whether to stop: false = keep iterating, true = stop.
// If an error is encountered while reading from the store, that entry is skipped and an error is
// returned for it when iteration is completed.
func (k Keeper) IterateAllHolds(ctx sdk.Context, process func(sdk.AccAddress, sdk.Coin) bool) error {
	store := k.getAllHoldCoinPrefixStore(ctx)

	iter := store.Iterator(nil, nil)
	defer iter.Close()

	var errs []error
	for ; iter.Valid(); iter.Next() {
		key := iter.Key()
		value := iter.Value()

		addr, denom := ParseHoldCoinKeyUnprefixed(key)
		amount, err := UnmarshalHoldCoinValue(value)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to read amount of %s for account %s: %w", denom, addr, err))
			continue
		}

		if process(addr, sdk.Coin{Denom: denom, Amount: amount}) {
			break
		}
	}

	return errors.Join(errs...)
}

// GetAllAccountHolds gets all the AccountHold entries currently in the state store.
func (k Keeper) GetAllAccountHolds(ctx sdk.Context) ([]*hold.AccountHold, error) {
	var holds []*hold.AccountHold
	var lastAddr sdk.AccAddress
	var lastEntry *hold.AccountHold

	err := k.IterateAllHolds(ctx, func(addr sdk.AccAddress, coin sdk.Coin) bool {
		if !addr.Equals(lastAddr) {
			lastAddr = addr
			lastEntry = &hold.AccountHold{
				Address: addr.String(),
				Amount:  sdk.Coins{},
			}
			holds = append(holds, lastEntry)
		}
		lastEntry.Amount = lastEntry.Amount.Add(coin)
		return false
	})
	return holds, err
}
