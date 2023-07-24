package keeper

import (
	"errors"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
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

// ValidateNewEscrow checks the account's spendable balance to make sure it has at least as much as the funds provided.
func (k Keeper) ValidateNewEscrow(ctx sdk.Context, addr sdk.AccAddress, funds sdk.Coins) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("cannot put %q in escrow for account %s: %w", funds, addr, err)
		}
	}()
	if funds.IsZero() {
		return nil
	}
	if funds.IsAnyNegative() {
		return errors.New("cannot escrow a negative amount")
	}

	// Not bypassing escrow's locked coins here because we're testing about new funds to be put in escrow.
	spendable := k.bankKeeper.SpendableCoins(ctx, addr)
	for _, fcoin := range funds {
		has, scoin := spendable.Find(fcoin.Denom)
		if !has {
			return fmt.Errorf("account has 0%s spendable", fcoin.Denom)
		}
		if scoin.Amount.LT(fcoin.Amount) {
			return fmt.Errorf("account has only %s spendable", scoin)
		}
	}

	return nil
}

// AddEscrow puts the provided funds in escrow for the provided account.
func (k Keeper) AddEscrow(ctx sdk.Context, addr sdk.AccAddress, funds sdk.Coins) error {
	if err := k.ValidateNewEscrow(ctx, addr, funds); err != nil {
		return err
	}

	// TODO[1607]: Implement AddEscrow.
	return nil
}

// RemoveEscrow takes the provided funds out of escrow for the provided account.
func (k Keeper) RemoveEscrow(ctx sdk.Context, addr sdk.AccAddress, funds sdk.Coins) error {
	// TODO[1607]: Implement RemoveEscrow.
	_, _, _ = ctx, addr, funds
	return nil
}

// GetEscrowCoin gets the amount of a denom in escrow for a given account.
func (k Keeper) GetEscrowCoin(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	// TODO[1607]: Implement GetEscrowCoin.
	_, _, _ = ctx, addr, denom
	return sdk.Coin{}
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

// IterateEscrow iterates over all funds in escrow for a given account.
// The process function should return whether to stop: false = keep iterating, true = stop.
func (k Keeper) IterateEscrow(ctx sdk.Context, addr sdk.AccAddress, process func(sdk.Coin) bool) error {
	// TODO[1607]: Implement IterateEscrow.
	_, _, _ = ctx, addr, process
	return nil
}

// IterateAllEscrow iterates over all in escrow coin entries for all accounts.
// The process function should return whether to stop: false = keep iterating, true = stop.
func (k Keeper) IterateAllEscrow(ctx sdk.Context, process func(sdk.AccAddress, sdk.Coin) bool) error {
	// TODO[1607]: Implement IterateAllEscrow.
	_, _ = ctx, process
	return nil
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
