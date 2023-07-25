package keeper

import (
	"errors"
	"fmt"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/escrow"
)

type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey

	bankKeeper escrow.BankKeeper
}

func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey, bankKeeper escrow.BankKeeper) Keeper {
	rv := Keeper{
		cdc:      cdc,
		storeKey: storeKey,
	}
	bankKeeper.AppendLockedCoinsGetter(rv.GetLockedCoins)
	return rv
}

// setEscrowCoinAmount updates the store with the provided escrow info.
// If the amount is zero, the escrow coin entry for addr+denom is deleted.
// Otherwise, the escrow coin entry for addr+denom is created/updated in the provided amount.
func (k Keeper) setEscrowCoinAmount(store sdk.KVStore, addr sdk.AccAddress, denom string, amount sdkmath.Int) error {
	if len(denom) == 0 {
		return fmt.Errorf("cannot store escrow with an empty denom for %s", addr)
	}
	if amount.IsNegative() {
		return fmt.Errorf("cannot store negative escrow amount %s%s for %s", amount, denom, addr)
	}

	key := CreateEscrowCoinKey(addr, denom)
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

// getEscrowCoinAmount gets (from the store) the amount marked as in escrow for the given address and denom.
func (k Keeper) getEscrowCoinAmount(store sdk.KVStore, addr sdk.AccAddress, denom string) (sdkmath.Int, error) {
	key := CreateEscrowCoinKey(addr, denom)
	amountBz := store.Get(key)
	return UnmarshalEscrowCoinValue(amountBz)
}

// ValidateNewEscrow checks the account's spendable balance to make sure it has at least as much as the funds provided.
func (k Keeper) ValidateNewEscrow(ctx sdk.Context, addr sdk.AccAddress, funds sdk.Coins) error {
	if funds.IsZero() {
		return nil
	}
	if funds.IsAnyNegative() {
		return fmt.Errorf("escrow amounts %q for %s cannot be negative", funds, addr)
	}

	// Not bypassing escrow's locked coins here because we're testing about new funds to be put into escrow.
	spendable := k.bankKeeper.SpendableCoins(ctx, addr)
	for _, toAdd := range funds {
		if toAdd.IsZero() {
			continue
		}
		has, available := spendable.Find(toAdd.Denom)
		if !has {
			return fmt.Errorf("account %s spendable balance 0%s is less than escrow amount %s", addr, toAdd.Denom, toAdd)
		}
		if available.Amount.LT(toAdd.Amount) {
			return fmt.Errorf("account %s spendable balance %s is less than escrow amount %s", addr, available, toAdd)
		}
	}

	return nil
}

// AddEscrow puts the provided funds in escrow for the provided account.
func (k Keeper) AddEscrow(ctx sdk.Context, addr sdk.AccAddress, funds sdk.Coins) error {
	if funds.IsZero() {
		return nil
	}

	if err := k.ValidateNewEscrow(ctx, addr, funds); err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	var errs []error
	for _, toAdd := range funds {
		if toAdd.IsZero() {
			continue
		}
		inEscrow, err := k.getEscrowCoinAmount(store, addr, toAdd.Denom)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to get current escrow amount for %s: %w", addr, err))
			continue
		}
		newEscrowAmt := inEscrow.Add(toAdd.Amount)
		err = k.setEscrowCoinAmount(store, addr, toAdd.Denom, newEscrowAmt)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to place %q in escrow for %s: %w", toAdd, addr, err))
		}
	}
	return errors.Join(errs...)
}

// RemoveEscrow takes the provided funds out of escrow for the provided account.
func (k Keeper) RemoveEscrow(ctx sdk.Context, addr sdk.AccAddress, funds sdk.Coins) error {
	if funds.IsZero() {
		return nil
	}
	if funds.IsAnyNegative() {
		return fmt.Errorf("cannot remove %q from escrow for %s: amounts cannot be negative", funds, addr)
	}

	store := ctx.KVStore(k.storeKey)
	var errs []error
	for _, toRemove := range funds {
		if toRemove.IsZero() {
			continue
		}
		inEscrow, err := k.getEscrowCoinAmount(store, addr, toRemove.Denom)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to get current escrow amount for %s: %w", addr, err))
			continue
		}
		newAmount := inEscrow.Sub(toRemove.Amount)
		if newAmount.IsNegative() {
			errs = append(errs, fmt.Errorf("cannot remove %q from escrow for %s: account only has %q in escrow", toRemove, addr, inEscrow))
			continue
		}
		err = k.setEscrowCoinAmount(store, addr, toRemove.Denom, newAmount)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to remove %q from escrow for %s: %w", toRemove, addr, err))
		}
	}
	return errors.Join(errs...)
}

// GetEscrowCoin gets the amount of a denom in escrow for a given account.
// Will return a zero Coin of the given denom if the store does not have an entry for it.
func (k Keeper) GetEscrowCoin(ctx sdk.Context, addr sdk.AccAddress, denom string) (sdk.Coin, error) {
	var err error
	rv := sdk.Coin{Denom: denom}
	rv.Amount, err = k.getEscrowCoinAmount(ctx.KVStore(k.storeKey), addr, denom)
	if err != nil {
		return rv, fmt.Errorf("could not get escrow coin amount for %s: %w", addr, err)
	}
	return rv, nil
}

// GetEscrowCoins gets all funds in escrow for a given account.
func (k Keeper) GetEscrowCoins(ctx sdk.Context, addr sdk.AccAddress) (sdk.Coins, error) {
	var rv sdk.Coins
	err := k.IterateEscrow(ctx, addr, func(coin sdk.Coin) bool {
		rv = rv.Add(coin)
		return false
	})

	return rv, err
}

// getEscrowCoinPrefixStore returns a kv store prefixed for escrow coin entries for the provided address.
func (k Keeper) getEscrowCoinPrefixStore(ctx sdk.Context, addr sdk.AccAddress) sdk.KVStore {
	pre := CreateEscrowCoinKeyAddrPrefix(addr)
	return prefix.NewStore(ctx.KVStore(k.storeKey), pre)
}

// IterateEscrow iterates over all funds in escrow for a given account.
// The process function should return whether to stop: false = keep iterating, true = stop.
// If an error is encountered while reading from the store, that entry is skipped and an error is
// returned for it when iteration is completed.
func (k Keeper) IterateEscrow(ctx sdk.Context, addr sdk.AccAddress, process func(sdk.Coin) bool) error {
	store := k.getEscrowCoinPrefixStore(ctx, addr)

	iter := store.Iterator(nil, nil)
	defer iter.Close()

	var errs []error
	for ; iter.Valid(); iter.Next() {
		key := iter.Key()
		value := iter.Value()

		amount, err := UnmarshalEscrowCoinValue(value)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		denom := string(key)

		if process(sdk.Coin{Denom: denom, Amount: amount}) {
			break
		}
	}

	return errors.Join(errs...)
}

// getAllEscrowCoinPrefixStore returns a kv store prefixed for all escrow coin entries.
func (k Keeper) getAllEscrowCoinPrefixStore(ctx sdk.Context) sdk.KVStore {
	return prefix.NewStore(ctx.KVStore(k.storeKey), KeyPrefixEscrowCoin)
}

// IterateAllEscrow iterates over all in escrow coin entries for all accounts.
// The process function should return whether to stop: false = keep iterating, true = stop.
// If an error is encountered while reading from the store, that entry is skipped and an error is
// returned for it when iteration is completed.
func (k Keeper) IterateAllEscrow(ctx sdk.Context, process func(sdk.AccAddress, sdk.Coin) bool) error {
	store := k.getAllEscrowCoinPrefixStore(ctx)

	iter := store.Iterator(nil, nil)
	defer iter.Close()

	var errs []error
	for ; iter.Valid(); iter.Next() {
		key := iter.Key()
		value := iter.Value()

		amount, err := UnmarshalEscrowCoinValue(value)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		addr, denom := ParseEscrowCoinKeyUnprefixed(key)

		if process(addr, sdk.Coin{Denom: denom, Amount: amount}) {
			break
		}
	}

	return errors.Join(errs...)
}

// GetAllAccountEscrows gets all the AccountEscrow entries currently in the state store.
func (k Keeper) GetAllAccountEscrows(ctx sdk.Context) (escrows []*escrow.AccountEscrow, err error) {
	defer func() {
		// A panic might happen with the Coins stuff. So handle that a bit better.
		if r := recover(); r != nil {
			if rE, isE := r.(error); isE {
				err = fmt.Errorf("recovered from panic: %w", rE)
			} else {
				err = fmt.Errorf("recovered from panic: %v", r)
			}
		}
		if err != nil {
			escrows = nil
		}
	}()
	var lastAddr sdk.AccAddress
	var lastEntry *escrow.AccountEscrow

	err = k.IterateAllEscrow(ctx, func(addr sdk.AccAddress, coin sdk.Coin) bool {
		if !addr.Equals(lastAddr) {
			lastAddr = addr
			lastEntry = &escrow.AccountEscrow{
				Address: addr.String(),
				Amount:  sdk.Coins{},
			}
			escrows = append(escrows, lastEntry)
		}
		lastEntry.Amount = lastEntry.Amount.Add(coin)
		return false
	})
	return escrows, err
}
