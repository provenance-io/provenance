package keeper

import (
	"errors"
	"fmt"
	"strings"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/gogoproto/proto"

	"github.com/provenance-io/provenance/x/hold"
)

type Keeper struct {
	cdc           codec.BinaryCodec
	StoreService  store.KVStoreService
	accountKeeper hold.AccountKeeper
	bankKeeper    hold.BankKeeper
	authority     string
	Schema        collections.Schema
	Holds         collections.Map[collections.Pair[sdk.AccAddress, string], sdkmath.Int]
}

func NewKeeper(cdc codec.BinaryCodec, storeService store.KVStoreService, accountKeeper hold.AccountKeeper, bankKeeper hold.BankKeeper) Keeper {
	sb := collections.NewSchemaBuilder(storeService)

	// Create the pair key codec using our custom codecs
	pairKeyCodec := collections.PairKeyCodec(
		hold.AddressKeyCodec{},
		hold.DenomKeyCodec{},
	)

	k := Keeper{
		cdc:           cdc,
		accountKeeper: accountKeeper,
		bankKeeper:    bankKeeper,
		authority:     authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		StoreService:  storeService,
		Holds: collections.NewMap(
			sb,
			KeyPrefixHoldCoin,
			"holds",
			pairKeyCodec,
			hold.IntValueCodec{},
		),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema

	bankKeeper.AppendLockedCoinsGetter(k.GetLockedCoins)
	return k
}

// setHoldCoinAmount updates the store with the provided hold info.
// If the amount is zero, the hold coin entry for addr+denom is deleted.
// Otherwise, the hold coin entry for addr+denom is created/updated in the provided amount.
func (k Keeper) setHoldCoinAmount(ctx sdk.Context, addr sdk.AccAddress, denom string, amount sdkmath.Int) error {
	if len(denom) == 0 {
		return fmt.Errorf("cannot store hold with an empty denom for %s", addr)
	}
	if amount.IsNegative() {
		return fmt.Errorf("cannot store negative hold amount %s%s for %s", amount, denom, addr)
	}

	key := collections.Join(addr, denom)
	if amount.IsZero() {
		return k.Holds.Remove(ctx, key)
	}
	return k.Holds.Set(ctx, key, amount)
}

// getHoldCoinAmount gets (from the store) the amount marked as on hold for the given address and denom.
func (k Keeper) getHoldCoinAmount(ctx sdk.Context, addr sdk.AccAddress, denom string) (sdkmath.Int, error) {
	key := collections.Join(addr, denom)
	amount, err := k.Holds.Get(ctx, key)
	if errors.Is(err, collections.ErrNotFound) {
		return sdkmath.ZeroInt(), nil
	}
	return amount, err
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
	spendable := k.getSpendableForDenoms(ctx, addr, funds)
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

// getSpendableForDenoms gets the spendable balances of the denoms in the provided funds.
// Only the denoms in the provided funds are used (the amounts are ignored).
// This is preferable to the bank keeper's SpendableBalances query because that one will
// iterate over all denoms owned by the address even if we only need to know about one or two.
func (k Keeper) getSpendableForDenoms(ctx sdk.Context, addr sdk.AccAddress, funds sdk.Coins) sdk.Coins {
	allLocked := k.bankKeeper.LockedCoins(ctx, addr)

	rv := make(sdk.Coins, 0, len(funds))
	for _, coin := range funds {
		bal := k.bankKeeper.GetBalance(ctx, addr, coin.Denom)
		if !bal.IsPositive() {
			continue
		}

		locked := allLocked.AmountOf(coin.Denom)
		if locked.IsPositive() {
			if bal.Amount.LTE(locked) {
				continue
			}
			bal.Amount = bal.Amount.Sub(locked)
		}

		rv = append(rv, bal)
	}
	return rv
}

// AddHold puts the provided funds on hold for the provided account.
func (k Keeper) AddHold(ctx sdk.Context, addr sdk.AccAddress, funds sdk.Coins, reason string) error {
	if funds.IsZero() {
		return nil
	}

	if err := k.ValidateNewHold(ctx, addr, funds); err != nil {
		return err
	}

	var fundsAdded sdk.Coins
	var errs []error
	for _, toAdd := range funds {
		if toAdd.IsZero() {
			continue
		}
		onHold, err := k.getHoldCoinAmount(ctx, addr, toAdd.Denom)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to get current %s hold amount for %s: %w", toAdd.Denom, addr, err))
			continue
		}
		newHoldAmt := onHold.Add(toAdd.Amount)
		err = k.setHoldCoinAmount(ctx, addr, toAdd.Denom, newHoldAmt)
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

	var fundsReleased sdk.Coins
	var errs []error
	for _, toRelease := range funds {
		if toRelease.IsZero() {
			continue
		}

		onHold, err := k.getHoldCoinAmount(ctx, addr, toRelease.Denom)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to get current %s hold amount for %s: %w", toRelease.Denom, addr, err))
			continue
		}

		newAmount := onHold.Sub(toRelease.Amount)
		if newAmount.IsNegative() {
			errs = append(errs, fmt.Errorf("cannot release %s from hold for %s: account only has %s%s on hold", toRelease, addr, onHold, toRelease.Denom))
			continue
		}

		err = k.setHoldCoinAmount(ctx, addr, toRelease.Denom, newAmount)
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
	key := collections.Join(addr, denom)
	amount, err := k.Holds.Get(ctx, key)
	if err != nil {
		// Not found → zero coin, no error
		if errors.Is(err, collections.ErrNotFound) {
			return sdk.NewCoin(denom, sdkmath.ZeroInt()), nil
		}
		return sdk.NewCoin(denom, sdkmath.ZeroInt()),
			fmt.Errorf("could not get hold coin amount for %s: %w", addr, err)
	}
	return sdk.NewCoin(denom, amount), nil
}

// GetHoldCoins gets all funds on hold for a given account.
func (k Keeper) GetHoldCoins(ctx sdk.Context, addr sdk.AccAddress) (sdk.Coins, error) {
	var coins sdk.Coins
	rng := collections.NewPrefixedPairRange[sdk.AccAddress, string](addr)
	var errs []error
	iterator, err := k.Holds.Iterate(ctx, rng)
	if err != nil {
		return nil, err
	}
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		kv, err := iterator.KeyValue()
		if err != nil {
			denom := ""
			key, keyErr := iterator.Key()
			if keyErr == nil {
				denom = key.K2()
			}
			errs = append(errs, fmt.Errorf("failed to read amount of %s for account %s: %w", denom, addr, err))
			continue
		}
		coins = coins.Add(sdk.NewCoin(kv.Key.K2(), kv.Value))
	}

	return coins, errors.Join(errs...)
}

// IterateHolds iterates over all funds on hold for a given account.
// The process function should return whether to stop: false = keep iterating, true = stop.
// If an error is encountered while reading from the store, that entry is skipped and an error is
// returned for it when iteration is completed.
func (k Keeper) IterateHolds(ctx sdk.Context, addr sdk.AccAddress, process func(sdk.Coin) bool) error {
	rng := collections.NewPrefixedPairRange[sdk.AccAddress, string](addr)
	var errs []error
	iterator, err := k.Holds.Iterate(ctx, rng)
	if err != nil {
		return err
	}

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		kv, err := iterator.KeyValue()
		if err != nil {
			denom := ""
			key, keyErr := iterator.Key()
			if keyErr == nil {
				denom = key.K2()
			}
			errs = append(errs, fmt.Errorf("failed to read amount of %s for account %s: %w", denom, addr, err))
			continue
		}
		coin := sdk.NewCoin(kv.Key.K2(), kv.Value)
		if process(coin) {
			break
		}
	}

	return errors.Join(errs...)
}

// IterateAllHolds iterates over all hold coin entries for all accounts.
// The process function should return whether to stop: false = keep iterating, true = stop.
// If an error is encountered while reading from the store, that entry is skipped and an error is
// returned for it when iteration is completed.
func (k Keeper) IterateAllHolds(ctx sdk.Context, process func(sdk.AccAddress, sdk.Coin) bool) error {
	var errs []error
	iterator, err := k.Holds.Iterate(ctx, nil)
	if err != nil {
		return err
	}

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		kv, err := iterator.KeyValue()
		if err != nil {
			addr := sdk.AccAddress(nil)
			denom := ""
			key, keyErr := iterator.Key()
			if keyErr == nil {
				addr = key.K1()
				denom = key.K2()
			}
			errs = append(errs, fmt.Errorf("failed to read amount of %s for account %s: %w", denom, addr, err))
			continue
		}
		coin := sdk.NewCoin(kv.Key.K2(), kv.Value)
		if process(kv.Key.K1(), coin) {
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

// GetLogger gets the logger to use in the hold module keeper stuff.
func (k Keeper) GetLogger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "hold")
}

// emitTypedEvents will emit the provided event, logging an error about it if there's a problem.
func (k Keeper) emitTypedEvent(ctx sdk.Context, tev proto.Message) {
	if err := ctx.EventManager().EmitTypedEvent(tev); err != nil {
		k.GetLogger(ctx).Error("Could not emit typed event.", "event", tev, "error", err)
	}
}

// GetAuthority returns the module's authority address
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

// UnlockVestingAccounts unlocks each of the accounts with the given addrs.
// A failure to convert one of them does not prevent the conversion of the rest.
// Returns an error if anything went wrong (even if some stuff also went right).
func (k Keeper) UnlockVestingAccounts(ctx sdk.Context, addrs []string) error {
	var errs []error
	for _, addrStr := range addrs {
		addr, err := sdk.AccAddressFromBech32(addrStr)
		if err != nil {
			err = fmt.Errorf("invalid address %q: %w", addrStr, err)
			errs = append(errs, err)
			k.GetLogger(ctx).Error("Could not unlock vesting account with invalid address.", "address", addrStr, "error", err)
			continue
		}

		err = k.UnlockVestingAccount(ctx, addr)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) == 1 {
		return errs[0]
	}
	return errors.Join(errs...)
}

// UnlockVestingAccount converts a vesting account back to a base account
func (k Keeper) UnlockVestingAccount(ctx sdk.Context, addr sdk.AccAddress) (err error) {
	logger := k.GetLogger(ctx).With("address", addr.String())
	defer func() {
		if err != nil {
			logger.Error("Could not unlock vesting account.", "error", err)
		}
	}()

	account := k.accountKeeper.GetAccount(ctx, addr)
	if account == nil {
		return sdkerrors.ErrUnknownAddress.Wrapf("account %q does not exist", addr.String())
	}
	logger = logger.With("original_type", fmt.Sprintf("%T", account), "account_number", safeGetAcctNo(account))

	// Extract base account directly
	var baseVestAcct *vesting.BaseVestingAccount
	switch acct := account.(type) {
	case *vesting.ContinuousVestingAccount:
		baseVestAcct = acct.BaseVestingAccount
	case *vesting.DelayedVestingAccount:
		baseVestAcct = acct.BaseVestingAccount
	case *vesting.PeriodicVestingAccount:
		baseVestAcct = acct.BaseVestingAccount
	case *vesting.PermanentLockedAccount:
		baseVestAcct = acct.BaseVestingAccount
	default:
		return sdkerrors.ErrInvalidType.Wrapf("could not unlock account %s: unsupported account type %T", addr.String(), account)
	}
	if baseVestAcct == nil {
		return sdkerrors.ErrInvalidType.Wrapf("could not unlock account %s: base vesting account is nil", addr.String())
	}
	if baseVestAcct.BaseAccount == nil {
		return sdkerrors.ErrInvalidType.Wrapf("could not unlock account %s: base account is nil", addr.String())
	}

	k.accountKeeper.SetAccount(ctx, baseVestAcct.BaseAccount)
	logger.Info("Unlocked vesting account.")
	k.emitTypedEvent(ctx, hold.NewEventVestingAccountUnlocked(addr))
	return nil
}

// safeGetAcctNo returns acct.GetAccountNumber() ensuring that it doesn't panic.
func safeGetAcctNo(acct sdk.AccountI) (rv uint64) {
	defer func() {
		if r := recover(); r != nil {
			rv = 0
		}
	}()
	return acct.GetAccountNumber()
}
