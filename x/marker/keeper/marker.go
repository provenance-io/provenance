package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/provenance-io/provenance/x/marker/types"
)

// GetAllMarkerHolders returns an array of all account addresses holding the given denom (and the amount)
func (k Keeper) GetAllMarkerHolders(ctx sdk.Context, denom string) []types.Balance {
	var results []types.Balance
	k.bankKeeper.IterateAllBalances(ctx, func(addr sdk.AccAddress, coin sdk.Coin) (stop bool) {
		if coin.Denom == denom && !coin.Amount.IsZero() {
			results = append(results,
				types.Balance{
					Address: addr.String(),
					Coins:   sdk.NewCoins(coin),
				})
		}
		return false // do not stop iterating
	})
	return results
}

// GetMarkerByDenom looks up marker with the given denom
func (k Keeper) GetMarkerByDenom(ctx sdk.Context, denom string) (types.MarkerAccountI, error) {
	addr, err := types.MarkerAddress(denom)
	if err != nil {
		return nil, err
	}
	m, err := k.GetMarker(ctx, addr)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, fmt.Errorf("marker %s not found for address: %s", denom, addr)
	}
	return m, nil
}

// AddMarkerAccount persists marker to the account keeper store.
func (k Keeper) AddMarkerAccount(ctx sdk.Context, marker types.MarkerAccountI) error {
	if err := marker.Validate(); err != nil {
		return err
	}
	markerAddress := types.MustGetMarkerAddress(marker.GetDenom())

	if !marker.GetAddress().Equals(markerAddress) {
		return fmt.Errorf("marker address does not match expected %s for denom %s", markerAddress, marker.GetDenom())
	}

	// Should not exist yet
	existing, err := k.GetMarker(ctx, markerAddress)
	if err != nil {
		return err
	}
	if existing != nil {
		return fmt.Errorf("account already exists for %s", markerAddress)
	}

	// set base account number
	marker = k.NewMarker(ctx, marker)

	k.SetMarker(ctx, marker)

	return nil
}

// AddAccess adds the provided AccessGrant to the marker of the caller is allowed to make changes
func (k Keeper) AddAccess(
	ctx sdk.Context, caller sdk.AccAddress, denom string, grant types.AccessGrantI,
) error {
	// (if marker does not exist then fail)
	m, err := k.GetMarkerByDenom(ctx, denom)
	if err != nil {
		return fmt.Errorf("marker not found for %s: %s", denom, err)
	}
	switch m.GetStatus() {
	// marker is fixed/active, assert permission to make changes by checking for Grant Permission
	case types.StatusFinalized, types.StatusActive:
		if !m.AddressHasAccess(caller, types.Access_Admin) && !k.accountControlsAllSupply(ctx, caller, m) {
			return fmt.Errorf("%s is not authorized to make access list changes against active %s marker",
				caller, m.GetDenom())
		}
		fallthrough
	case types.StatusProposed:
		mgr := m.GetManager()
		// Check to see if fromAddr is the creator (and status is proposed against fallthrough case)
		if !mgr.Equals(caller) && m.GetStatus() == types.StatusProposed {
			return fmt.Errorf("updates to pending marker %s can only be made by %s", m.GetDenom(), mgr)
		}
		if err = m.GrantAccess(grant); err != nil {
			return fmt.Errorf("access grant failed: %w", err)
		}
		k.SetMarker(ctx, m)
	// Undefined, Cancelled, Destroyed -- no modifications are supported in these states
	default:
		return fmt.Errorf("marker in %s state can not be modified", m.GetStatus())
	}
	return nil
}

// RemoveAccess delete the AccessGrant for the specified user from the marker if the caller is allowed to make changes
func (k Keeper) RemoveAccess(ctx sdk.Context, caller sdk.AccAddress, denom string, remove sdk.AccAddress) error {
	// (if marker does not exist then fail)
	m, err := k.GetMarkerByDenom(ctx, denom)
	if err != nil {
		return fmt.Errorf("marker not found for %s: %s", denom, err)
	}
	switch m.GetStatus() {
	// marker is fixed/active, assert permission to make changes by checking for Grant Permission
	case types.StatusFinalized, types.StatusActive:
		if !m.AddressHasAccess(caller, types.Access_Admin) && !k.accountControlsAllSupply(ctx, caller, m) {
			return fmt.Errorf("%s is not authorized to make access list changes against active %s marker",
				caller, m.GetDenom())
		}
		fallthrough
	case types.StatusProposed:
		mgr := m.GetManager()
		// Check to see if fromAddr is the creator (and status is proposed against fallthrough case)
		if !mgr.Equals(caller) && m.GetStatus() == types.StatusProposed {
			return fmt.Errorf("updates to pending marker %s can only be made by %s", m.GetDenom(), mgr.String())
		}
		if err = m.RevokeAccess(remove); err != nil {
			return fmt.Errorf("access grant failed: %w", err)
		}
		k.SetMarker(ctx, m)
	// Undefined, Cancelled, Destroyed -- no modifications are supported in these states
	default:
		return fmt.Errorf("marker in %s state can not be modified", m.GetStatus())
	}

	return nil
}

// WithdrawCoins removes the specified coins from the MarkerAccount (both marker denominated coins and coins as assets
// are supported here)
func (k Keeper) WithdrawCoins(
	ctx sdk.Context, caller sdk.AccAddress, recipient sdk.AccAddress, denom string, coins sdk.Coins,
) error {
	// (if marker does not exist then fail)
	m, err := k.GetMarkerByDenom(ctx, denom)
	if err != nil {
		return fmt.Errorf("marker not found for %s: %s", denom, err)
	}
	if !m.AddressHasAccess(caller, types.Access_Withdraw) {
		return fmt.Errorf("%s does not have %s on %s markeraccount", caller, types.Access_Withdraw, m.GetDenom())
	}
	// check to see if marker is active (the coins created by a marker can only be withdrawn when it is active)
	// any other coins that may be present (collateralized assets?) can be transferred
	if m.GetStatus() != types.StatusActive && coins.AmountOf(denom).GT(sdk.ZeroInt()) {
		return fmt.Errorf("cannot withdraw marker created coins from a marker that is not in Active status")
	}

	if recipient.Empty() {
		recipient = caller
	}

	if err := k.bankKeeper.InputOutputCoins(ctx, []banktypes.Input{banktypes.NewInput(m.GetAddress(), coins)},
		[]banktypes.Output{banktypes.NewOutput(recipient, coins)}); err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeWithdraw,
			sdk.NewAttribute(types.EventAttributeDenomKey, denom),
			sdk.NewAttribute(types.EventAttributeAmountKey, coins.String()),
			sdk.NewAttribute(types.EventAttributeAdministratorKey, caller.String()),
			sdk.NewAttribute(types.EventAttributeModuleNameKey, types.ModuleName),
		),
	)
	return nil
}

// MintCoin increases the Supply of a coin by interacting with the supply keeper for the adjustment,
// updating the marker's record of expected total supply, and transferring the created coin to the MarkerAccount
// for holding pending further action.
func (k Keeper) MintCoin(ctx sdk.Context, caller sdk.AccAddress, coin sdk.Coin) error {
	// (if marker does not exist then fail)
	m, err := k.GetMarkerByDenom(ctx, coin.Denom)
	if err != nil {
		return fmt.Errorf("marker not found for %s: %s", coin.Denom, err)
	}
	if !m.AddressHasAccess(caller, types.Access_Mint) {
		return fmt.Errorf("%s does not have %s on %s markeraccount", caller, types.Access_Mint, m.GetDenom())
	}
	// For proposed, finalized accounts we allow adjusting the total_supply of the marker but we do not
	// mint actual coin.
	if m.GetStatus() == types.StatusProposed || m.GetStatus() == types.StatusFinalized {
		total := m.GetSupply().Add(coin)
		if err = m.SetSupply(total); err != nil {
			return err
		}
		k.SetMarker(ctx, m)
		return nil
	} else if m.GetStatus() != types.StatusActive {
		return fmt.Errorf("cannot mint coin for a marker that is not in Active status")
	}

	// Increase the tracked supply value for the marker.
	err = k.IncreaseSupply(ctx, m, coin)
	if err != nil {
		return err
	}
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeMint,
			sdk.NewAttribute(types.EventAttributeDenomKey, coin.Denom),
			sdk.NewAttribute(types.EventAttributeAmountKey, coin.Amount.String()),
			sdk.NewAttribute(types.EventAttributeAdministratorKey, caller.String()),
			sdk.NewAttribute(types.EventAttributeModuleNameKey, types.ModuleName),
		),
	)
	return nil
}

// BurnCoin removes supply from the marker by burning coins held within the marker acccount.
func (k Keeper) BurnCoin(ctx sdk.Context, caller sdk.AccAddress, coin sdk.Coin) error {
	// (if marker does not exist then fail)
	m, err := k.GetMarkerByDenom(ctx, coin.Denom)
	if err != nil {
		return fmt.Errorf("marker not found for %s: %s", coin.Denom, err)
	}
	if !m.AddressHasAccess(caller, types.Access_Burn) {
		return fmt.Errorf("%s does not have %s on %s markeraccount", caller, types.Access_Burn, m.GetDenom())
	}
	// For proposed, finalized accounts we allow adjusting the total_supply of the marker but we do not
	// burn actual coin.
	if m.GetStatus() == types.StatusProposed || m.GetStatus() == types.StatusFinalized {
		total := m.GetSupply().Sub(coin)
		if err = m.SetSupply(total); err != nil {
			return err
		}
		k.SetMarker(ctx, m)
		return nil
	} else if m.GetStatus() != types.StatusActive { // check to see if marker is active
		return fmt.Errorf("cannot mint coin for a marker that is not in Active status")
	}
	err = k.DecreaseSupply(ctx, m, coin)
	if err != nil {
		return err
	}
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeBurn,
			sdk.NewAttribute(types.EventAttributeDenomKey, coin.Denom),
			sdk.NewAttribute(types.EventAttributeAmountKey, coin.Amount.String()),
			sdk.NewAttribute(types.EventAttributeAdministratorKey, caller.String()),
			sdk.NewAttribute(types.EventAttributeModuleNameKey, types.ModuleName),
		),
	)
	return nil
}

// IncreaseSupply will mint coins to the marker module coin pool account, then send these to the marker account
func (k Keeper) IncreaseSupply(ctx sdk.Context, marker types.MarkerAccountI, coin sdk.Coin) error {
	// using the marker, update the supply to ensure successful (save pending), abort on fail
	total := marker.GetSupply().Add(coin)
	if err := marker.SetSupply(total); err != nil {
		return err
	}

	// Create the coins
	if err := k.bankKeeper.MintCoins(ctx, types.CoinPoolName, sdk.NewCoins(coin)); err != nil {
		return fmt.Errorf("could not increase coin supply in marker module: %v", err)
	}
	// Create successful, save the updated marker
	k.SetMarker(ctx, marker)

	// dispense the coins from the marker module to the marker account.
	return k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.CoinPoolName, marker.GetAddress(), sdk.NewCoins(coin))
}

// DecreaseSupply will move a given amount of coin from the marker to the markermodule coin pool account then burn it.
func (k Keeper) DecreaseSupply(ctx sdk.Context, marker types.MarkerAccountI, coin sdk.Coin) error {
	total := marker.GetSupply()
	// Ensure the request will not send the total supply below zero
	if total.IsLT(coin) {
		return fmt.Errorf("cannot reduce marker total supply below zero %s, %v", coin.Denom, coin.Amount)
	}
	// ensure the current marker account is holding enough coin to cover burn request
	escrow := k.bankKeeper.GetBalance(ctx, marker.GetAddress(), marker.GetDenom())
	if !escrow.Amount.GTE(coin.Amount) {
		return fmt.Errorf("marker account contains insufficient funds to burn %s, %v", coin.Denom, coin.Amount)
	}
	// Update the supply (abort if this can not be done)
	total = total.Sub(coin)
	if err := marker.SetSupply(total); err != nil {
		return err
	}

	// Finalize supply update in marker record
	k.SetMarker(ctx, marker)

	// Move coins to markermodule coin pool in preparation for burn
	if err := k.bankKeeper.SendCoinsFromAccountToModule(
		ctx, marker.GetAddress(), types.CoinPoolName, sdk.NewCoins(coin),
	); err != nil {
		panic(fmt.Errorf("could not send coin %v from marker account to module account: %w", coin, err))
	}
	// Perform controlled burn
	if err := k.bankKeeper.BurnCoins(ctx, types.CoinPoolName, sdk.NewCoins(coin)); err != nil {
		panic(fmt.Errorf("could not burn coin %v %w", coin, err))
	}

	return nil
}

// FinalizeMarker sets the state of the marker to finalized, mints the associated supply, assigns the minted coin to
// the marker accounts, and transitions the state to active if successful
func (k Keeper) FinalizeMarker(ctx sdk.Context, caller sdk.Address, denom string) error {
	// (if marker does not exist then fail)
	m, err := k.GetMarkerByDenom(ctx, denom)
	if err != nil {
		return fmt.Errorf("marker not found for %s: %s", denom, err)
	}
	// Only the manger can finalize a marker
	if !m.GetManager().Equals(caller) {
		return fmt.Errorf("%s does not have permission to finalize %s markeraccount", caller, m.GetDenom())
	}

	// status must currently be set to proposed
	if m.GetStatus() != types.StatusProposed {
		return fmt.Errorf("can only finalize markeraccounts in the Proposed status")
	}

	// verify marker configuration is sane
	if err = m.Validate(); err != nil {
		return fmt.Errorf("invalid marker, cannot be finalized: %w", err)
	}

	// Amount to mint is typically the defined supply however...
	supplyRequest := m.GetSupply()

	// Any pre-existing coin amounts for our denom need to be removed from our amount to mint
	preexistingCoin := sdk.NewCoin(m.GetDenom(), k.bankKeeper.GetSupply(ctx).GetTotal().AmountOf(m.GetDenom()))

	// If the requested total is less than the existing total, the supply invariant would halt the chain if activated
	if supplyRequest.IsLT(preexistingCoin) {
		return fmt.Errorf("marker supply %v has been defined as less than pre-existing"+
			" supply %v, can not finalize marker", supplyRequest, preexistingCoin)
	}

	// transition to finalized state ... then to active once mint is complete
	if err = m.SetStatus(types.StatusFinalized); err != nil {
		return fmt.Errorf("could not transition marker account state to finalized: %w", err)
	}

	// record status as finalized.
	k.SetMarker(ctx, m)
	return nil
}

// ActivateMarker transistions a marker into the active status, enforcing permissions, supply constraints, and minting
// any supply as required.
func (k Keeper) ActivateMarker(ctx sdk.Context, caller sdk.Address, denom string) error {
	m, err := k.GetMarkerByDenom(ctx, denom)
	if err != nil {
		return fmt.Errorf("marker not found for %s: %s", denom, err)
	}
	// Only the manger can activate a marker
	if !m.GetManager().Equals(caller) {
		return fmt.Errorf("%s does not have permission to finalize %s markeraccount", caller, m.GetDenom())
	}

	// must be in finalized state ... mint required supply amounts.
	if m.GetStatus() != types.StatusFinalized {
		return fmt.Errorf("can only activate markeraccounts in the Finalized status")
	}

	// Amount to mint is typically the defined supply however...
	supplyRequest := m.GetSupply()

	// Any pre-existing coin amounts for our denom need to be removed from our amount to mint
	preexistingCoin := sdk.NewCoin(m.GetDenom(), k.bankKeeper.GetSupply(ctx).GetTotal().AmountOf(m.GetDenom()))

	// If the requested total is less than the existing total, the supply invariant would halt the chain if activated
	if supplyRequest.IsLT(preexistingCoin) {
		return fmt.Errorf("marker supply %v has been defined as less than pre-existing"+
			" supply %v, can not finalize marker", supplyRequest, preexistingCoin)
	}

	// Amount we will mint is remainder after subtracting the existing supply from the system.
	mintAmount := sdk.NewCoins(supplyRequest.Sub(preexistingCoin))

	// coins are minted by the supply module and distributed to the markermodule
	if err = k.bankKeeper.MintCoins(ctx, types.CoinPoolName, mintAmount); err != nil {
		return fmt.Errorf("could not mint specified token supply for marker: %w", err)
	}
	// distribute minted coin to the markeraccount instance
	if err = k.bankKeeper.SendCoinsFromModuleToAccount(
		ctx, types.CoinPoolName, m.GetAddress(), mintAmount,
	); err != nil {
		return fmt.Errorf("could not distribute coin allocation from markermodule to marker instance: %w", err)
	}
	// reload our marker instance
	m, err = k.GetMarker(ctx, m.GetAddress())
	if err != nil {
		return fmt.Errorf("could not reload existing marker using address: %s, %w", m.GetAddress(), err)
	}

	// With the coin supply minted and assigned to the marker we can transition to the Active state.
	// this will enable the Invariant supply enforcement constraint.
	if err = m.SetStatus(types.StatusActive); err != nil {
		return fmt.Errorf("could not set marker status to active: %w", err)
	}

	// record status as active
	k.SetMarker(ctx, m)
	return nil
}

// CancelMarker prepares transition to deleted state.
func (k Keeper) CancelMarker(ctx sdk.Context, caller sdk.AccAddress, denom string) error {
	m, err := k.GetMarkerByDenom(ctx, denom)
	if err != nil {
		return fmt.Errorf("marker not found for %s: %s", denom, err)
	}

	switch m.GetStatus() {
	case types.StatusFinalized, types.StatusActive:
		// for active or finalized markers the caller must be assigned permission to perform this action.
		if !m.AddressHasAccess(caller, types.Access_Delete) {
			return fmt.Errorf("%s does not have %s on %s markeraccount", caller, types.Access_Delete, m.GetDenom())
		}
		// for finalized/active we need to ensure the full coin supply has been recalled as it will all be burned.
		totalSupply := k.bankKeeper.GetSupply(ctx).GetTotal().AmountOf(denom)
		escrow := k.bankKeeper.GetBalance(ctx, m.GetAddress(), m.GetDenom())
		inCirculation := totalSupply.Sub(escrow.Amount)
		if inCirculation.GT(sdk.ZeroInt()) {
			return fmt.Errorf("cannot cancel marker with %d minted coin in circulation out of %d total."+
				" ensure marker account holds the entire supply of %s", inCirculation, totalSupply, denom)
		}
	case types.StatusProposed:
		// for a proposed marker either the manager or someone assigned `delete` can perform this action
		if !(m.GetManager().Equals(caller) || m.AddressHasAccess(caller, types.Access_Delete)) {
			return fmt.Errorf("%s does not have %s on %s markeraccount", caller, types.Access_Delete, m.GetDenom())
		}
	case types.StatusCancelled:
		return nil // nothing to be done here.
	default:
		return fmt.Errorf("marker must be proposed, finalized, or active status to be cancelled")
	}
	if err = m.SetStatus(types.StatusCancelled); err != nil {
		return fmt.Errorf("could not update marker status: %w", err)
	}
	k.SetMarker(ctx, m)
	return nil
}

// DeleteMarker burns the entire coin supply, ensure no assets are pooled, and marks the current instance of the
// marker as destroyed.
func (k Keeper) DeleteMarker(ctx sdk.Context, caller sdk.AccAddress, denom string) error {
	m, err := k.GetMarkerByDenom(ctx, denom)
	if err != nil {
		return fmt.Errorf("marker not found for %s: %s", denom, err)
	}

	if !m.AddressHasAccess(caller, types.Access_Delete) {
		return fmt.Errorf("%s does not have %s on %s markeraccount", caller, types.Access_Delete, m.GetDenom())
	}

	// status must currently be set to cancelled
	if m.GetStatus() != types.StatusCancelled {
		return fmt.Errorf("can only delete markeraccounts in the Cancelled status")
	}

	// require full supply of coin for marker to be contained within the marker account (no outstanding delegations)
	totalSupply := k.bankKeeper.GetSupply(ctx).GetTotal().AmountOf(denom)
	escrow := k.bankKeeper.GetBalance(ctx, m.GetAddress(), m.GetDenom())
	inCirculation := totalSupply.Sub(escrow.Amount)
	if inCirculation.GT(sdk.ZeroInt()) {
		return fmt.Errorf("cannot delete marker with %d minted coin in circulation out of %d total."+
			" ensure marker account holds the entire supply of %s", inCirculation, totalSupply, denom)
	}

	err = k.DecreaseSupply(ctx, m, sdk.NewCoin(denom, totalSupply))
	if err != nil {
		return fmt.Errorf("could not decrease marker supply %s: %s", denom, err)
	}
	// TODO: check metadata module for scopes with this marker assigned, if found this delete call fails

	// get the updated state of the marker afer supply burn...
	m, err = k.GetMarkerByDenom(ctx, denom)
	if err != nil {
		return fmt.Errorf("marker not found for %s: %s", denom, err)
	}
	if err = m.SetStatus(types.StatusDestroyed); err != nil {
		return fmt.Errorf("could not update marker status: %w", err)
	}
	k.SetMarker(ctx, m)

	return nil
}

// TransferCoin transfers restricted coins between to accounts when the administrator account holds the transfer
// access right and the marker type is restricted_coin
func (k Keeper) TransferCoin(ctx sdk.Context, from, to, admin sdk.AccAddress, amount sdk.Coin) error {
	m, err := k.GetMarkerByDenom(ctx, amount.Denom)
	if err != nil {
		return fmt.Errorf("marker not found for %s: %s", amount.Denom, err)
	}
	if m.GetMarkerType() != types.MarkerType_RestrictedCoin {
		return fmt.Errorf("marker type is not restricted_coin, brokered transfer not supported")
	}
	if !m.AddressHasAccess(admin, types.Access_Transfer) {
		return fmt.Errorf("%s is not allowed to broker transfers", admin.String())
	}
	if k.bankKeeper.BlockedAddr(to) {
		return fmt.Errorf("%s is not allowed to receive funds", to)
	}
	// send the coins between accounts (does not check send_enabled on coin denom)
	if err = k.bankKeeper.SendCoins(ctx, from, to, sdk.NewCoins(amount)); err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeTransfer,
			sdk.NewAttribute(types.EventAttributeDenomKey, amount.Denom),
			sdk.NewAttribute(types.EventAttributeAmountKey, amount.String()),
			sdk.NewAttribute(types.EventAttributeAdministratorKey, admin.String()),
			sdk.NewAttribute(types.EventAttributeModuleNameKey, types.ModuleName),
		),
	)
	return nil
}

// accountControlsAllSupply return true if the caller account address possess 100% of the total supply of a marker.
// This check is used to determine if an account should be allowed to perform defacto admin operations on a marker.
func (k Keeper) accountControlsAllSupply(ctx sdk.Context, caller sdk.AccAddress, m types.MarkerAccountI) bool {
	balance := k.bankKeeper.GetBalance(ctx, caller, m.GetDenom())

	// if the given account is currently holding 100% of the supply of a marker then it should be able to invoke
	// the operations as an admin on the marker.
	return m.GetSupply().IsEqual(sdk.NewCoin(m.GetDenom(), balance.Amount))
}
