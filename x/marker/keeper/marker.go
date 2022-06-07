package keeper

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/provenance-io/provenance/x/marker/types"
)

// GetAllMarkerHolders returns an array of all account addresses holding the given denom (and the amount)
func (k Keeper) GetAllMarkerHolders(ctx sdk.Context, denom string) []types.Balance {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "get_all_marker_holders")

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
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "get_marker_by_denom")

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
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "add_marker_account")

	if err := marker.Validate(); err != nil {
		return err
	}
	markerAddress := types.MustGetMarkerAddress(marker.GetDenom())

	if !marker.GetAddress().Equals(markerAddress) {
		return fmt.Errorf("marker address does not match expected %s for denom %s", markerAddress, marker.GetDenom())
	}

	// Should not exist yet (or if exists must not be a marker and must have a zero sequence number)
	mac := k.authKeeper.GetAccount(ctx, markerAddress)
	if mac != nil {
		_, ok := mac.(types.MarkerAccountI)
		if ok {
			return fmt.Errorf("marker address already exists for %s", markerAddress)
		} else if mac.GetSequence() > 0 {
			// account exists, is not a marker, and has been signed for
			return fmt.Errorf("account at %s is not a marker account", markerAddress.String())
		}
	}

	// set base account number
	marker = k.NewMarker(ctx, marker)

	if err := marker.Validate(); err != nil {
		return err
	}
	k.SetMarker(ctx, marker)

	markerAddEvent := types.NewEventMarkerAdd(
		marker.GetSupply().Denom,
		marker.GetSupply().Amount.String(),
		marker.GetStatus().String(),
		marker.GetManager().String(),
		marker.GetMarkerType().String(),
	)
	if err := ctx.EventManager().EmitTypedEvent(markerAddEvent); err != nil {
		return err
	}

	return nil
}

// AddAccess adds the provided AccessGrant to the marker of the caller is allowed to make changes
func (k Keeper) AddAccess(
	ctx sdk.Context, caller sdk.AccAddress, denom string, grant types.AccessGrantI,
) error {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "add_access")

	// (if marker does not exist then fail)
	m, err := k.GetMarkerByDenom(ctx, denom)
	if err != nil {
		return fmt.Errorf("marker not found for %s: %s", denom, err)
	}
	switch m.GetStatus() {
	// marker is fixed/active, assert permission to make changes by checking for Grant Permission
	case types.StatusFinalized, types.StatusActive:
		if !(caller.Equals(m.GetManager()) && m.GetStatus() == types.StatusFinalized) &&
			!m.AddressHasAccess(caller, types.Access_Admin) &&
			!k.accountControlsAllSupply(ctx, caller, m) {
			return fmt.Errorf("%s is not authorized to make access list changes against finalized/active %s marker",
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
		if err := m.Validate(); err != nil {
			return err
		}
		k.SetMarker(ctx, m)
	// Undefined, Cancelled, Destroyed -- no modifications are supported in these states
	default:
		return fmt.Errorf("marker in %s state can not be modified", m.GetStatus())
	}

	markerAddAccessEvent := types.NewEventMarkerAddAccess(grant, denom, caller.String())
	if err := ctx.EventManager().EmitTypedEvent(markerAddAccessEvent); err != nil {
		return err
	}

	return nil
}

// RemoveAccess delete the AccessGrant for the specified user from the marker if the caller is allowed to make changes
func (k Keeper) RemoveAccess(ctx sdk.Context, caller sdk.AccAddress, denom string, remove sdk.AccAddress) error {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "remove_access")

	// (if marker does not exist then fail)
	m, err := k.GetMarkerByDenom(ctx, denom)
	if err != nil {
		return fmt.Errorf("marker not found for %s: %s", denom, err)
	}
	switch m.GetStatus() {
	// marker is fixed/active, assert permission to make changes by checking for Grant Permission
	case types.StatusFinalized, types.StatusActive:
		if !(caller.Equals(m.GetManager()) && m.GetStatus() == types.StatusFinalized) &&
			!m.AddressHasAccess(caller, types.Access_Admin) &&
			!k.accountControlsAllSupply(ctx, caller, m) {
			return fmt.Errorf("%s is not authorized to make access list changes against finalized/active %s marker",
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
			return fmt.Errorf("access revoke failed: %w", err)
		}
		if err := m.Validate(); err != nil {
			return err
		}
		k.SetMarker(ctx, m)
	// Undefined, Cancelled, Destroyed -- no modifications are supported in these states
	default:
		return fmt.Errorf("marker in %s state can not be modified", m.GetStatus())
	}

	markerDeleteAccessEvent := types.NewEventMarkerDeleteAccess(remove.String(), denom, caller.String())
	if err := ctx.EventManager().EmitTypedEvent(markerDeleteAccessEvent); err != nil {
		return err
	}

	return nil
}

// WithdrawCoins removes the specified coins from the MarkerAccount (both marker denominated coins and coins as assets
// are supported here)
func (k Keeper) WithdrawCoins(
	ctx sdk.Context, caller sdk.AccAddress, recipient sdk.AccAddress, denom string, coins sdk.Coins,
) error {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "withdraw_coins")

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
	if m.GetStatus() != types.StatusActive {
		return fmt.Errorf("cannot withdraw marker created coins from a marker that is not in Active status")
	}

	if recipient.Empty() {
		recipient = caller
	}

	if err := k.bankKeeper.InputOutputCoins(ctx, []banktypes.Input{banktypes.NewInput(m.GetAddress(), coins)},
		[]banktypes.Output{banktypes.NewOutput(recipient, coins)}); err != nil {
		return err
	}

	markerWithdrawEvent := types.NewEventMarkerWithdraw(coins.String(), denom, caller.String(), recipient.String())
	if err := ctx.EventManager().EmitTypedEvent(markerWithdrawEvent); err != nil {
		return err
	}

	return nil
}

// MintCoin increases the Supply of a coin by interacting with the supply keeper for the adjustment,
// updating the marker's record of expected total supply, and transferring the created coin to the MarkerAccount
// for holding pending further action.
func (k Keeper) MintCoin(ctx sdk.Context, caller sdk.AccAddress, coin sdk.Coin) error {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "mint_coin")

	// (if marker does not exist then fail)
	m, err := k.GetMarkerByDenom(ctx, coin.Denom)
	if err != nil {
		return fmt.Errorf("marker not found for %s: %s", coin.Denom, err)
	}
	if !m.AddressHasAccess(caller, types.Access_Mint) {
		return fmt.Errorf("%s does not have %s on %s markeraccount", caller, types.Access_Mint, m.GetDenom())
	}

	switch {
	// For proposed, finalized accounts we allow adjusting the total_supply of the marker but we do not
	// mint actual coin.
	case m.GetStatus() == types.StatusProposed || m.GetStatus() == types.StatusFinalized:
		total := m.GetSupply().Add(coin)
		if err = m.SetSupply(total); err != nil {
			return err
		}
		if err = m.Validate(); err != nil {
			return err
		}
		k.SetMarker(ctx, m)
	case m.GetStatus() != types.StatusActive:
		return fmt.Errorf("cannot mint coin for a marker that is not in Active status")
	default:
		// Increase the tracked supply value for the marker.
		err = k.IncreaseSupply(ctx, m, coin)
		if err != nil {
			return err
		}
	}

	markerMintEvent := types.NewEventMarkerMint(coin.Amount.String(), coin.Denom, caller.String())
	if err := ctx.EventManager().EmitTypedEvent(markerMintEvent); err != nil {
		return err
	}

	return nil
}

// BurnCoin removes supply from the marker by burning coins held within the marker acccount.
func (k Keeper) BurnCoin(ctx sdk.Context, caller sdk.AccAddress, coin sdk.Coin) error {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "burn_coin")

	// (if marker does not exist then fail)
	m, err := k.GetMarkerByDenom(ctx, coin.Denom)
	if err != nil {
		return fmt.Errorf("marker not found for %s: %s", coin.Denom, err)
	}
	if !m.AddressHasAccess(caller, types.Access_Burn) {
		return fmt.Errorf("%s does not have %s on %s markeraccount", caller, types.Access_Burn, m.GetDenom())
	}

	switch {
	// For proposed, finalized accounts we allow adjusting the total_supply of the marker but we do not
	// burn actual coin.
	case m.GetStatus() == types.StatusProposed || m.GetStatus() == types.StatusFinalized:
		total := m.GetSupply().Sub(coin)
		if err = m.SetSupply(total); err != nil {
			return err
		}
		if err = m.Validate(); err != nil {
			return err
		}
		k.SetMarker(ctx, m)
	case m.GetStatus() != types.StatusActive:
		return fmt.Errorf("cannot burn coin for a marker that is not in Active status")
	default:
		err = k.DecreaseSupply(ctx, m, coin)
		if err != nil {
			return err
		}
	}

	markerBurnEvent := types.NewEventMarkerBurn(coin.Amount.String(), coin.Denom, caller.String())
	if err := ctx.EventManager().EmitTypedEvent(markerBurnEvent); err != nil {
		return err
	}

	return nil
}

// Returns the current supply in network according to the bank module for the given marker
func (k Keeper) CurrentCirculation(ctx sdk.Context, marker types.MarkerAccountI) sdk.Int {
	return k.bankKeeper.GetSupply(ctx, marker.GetDenom()).Amount
}

// Retures the current escrow balance for the marker base account
func (k Keeper) CurrentEscrow(ctx sdk.Context, marker types.MarkerAccountI) sdk.Coins {
	return k.bankKeeper.GetAllBalances(ctx, marker.GetAddress())
}

// AdjustCirculation will mint/burn coin if required to ensure desired supply matches amount in circulation
func (k Keeper) AdjustCirculation(ctx sdk.Context, marker types.MarkerAccountI, desiredSupply sdk.Coin) error {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "adjust_circulation")

	currentSupply := k.bankKeeper.GetSupply(ctx, marker.GetDenom()).Amount
	if desiredSupply.Denom != marker.GetDenom() {
		return fmt.Errorf("invalid denom for desired supply")
	}

	if desiredSupply.Amount.GT(currentSupply) { // not enough coin in circulation, mint more.
		offset := sdk.NewCoin(marker.GetDenom(), desiredSupply.Amount.Sub(currentSupply))
		ctx.Logger().Info(
			fmt.Sprintf("Adjusting %s circulation: increasing supply by %s",
				marker.GetDenom(), offset))
		if err := k.bankKeeper.MintCoins(ctx, types.CoinPoolName, sdk.NewCoins(offset)); err != nil {
			return err
		}
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(
			ctx, types.CoinPoolName, marker.GetAddress(), sdk.NewCoins(offset),
		); err != nil {
			return err
		}
	} else if desiredSupply.Amount.LT(currentSupply) { // too much coin in circulation, attempt to burn from marker account.
		offset := sdk.NewCoin(marker.GetDenom(), currentSupply.Sub(desiredSupply.Amount))
		ctx.Logger().Info(
			fmt.Sprintf("Adjusting %s circulation: decreasing supply by %s",
				marker.GetDenom(), offset))
		if err := k.bankKeeper.SendCoinsFromAccountToModule(
			ctx, marker.GetAddress(), types.CoinPoolName, sdk.NewCoins(offset),
		); err != nil {
			return fmt.Errorf("could not send coin %v from marker account to module account: %w", offset, err)
		}
		// Perform controlled burn
		if err := k.bankKeeper.BurnCoins(ctx, types.CoinPoolName, sdk.NewCoins(offset)); err != nil {
			return fmt.Errorf("could not burn coin %v %w", offset, err)
		}
	}
	return nil
}

// IncreaseSupply will mint coins to the marker module coin pool account, then send these to the marker account
func (k Keeper) IncreaseSupply(ctx sdk.Context, marker types.MarkerAccountI, coin sdk.Coin) error {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "increase_supply")

	inCirculation := sdk.NewCoin(marker.GetDenom(), k.bankKeeper.GetSupply(ctx, marker.GetDenom()).Amount)
	total := inCirculation.Add(coin)
	maxAllowed := sdk.NewCoin(marker.GetDenom(), sdk.NewIntFromUint64(k.GetParams(ctx).MaxTotalSupply))
	if total.Amount.GT(maxAllowed.Amount) {
		return fmt.Errorf(
			"requested supply %d exceeds maximum allowed value %d", total.Amount, maxAllowed.Amount)
	}

	// If the marker has a fixed supply then adjust the supply to match the new total
	if marker.HasFixedSupply() {
		if err := marker.SetSupply(total); err != nil {
			return err
		}
		if err := marker.Validate(); err != nil {
			return err
		}
		k.SetMarker(ctx, marker)
	}

	return k.AdjustCirculation(ctx, marker, total)
}

// DecreaseSupply will move a given amount of coin from the marker to the markermodule coin pool account then burn it.
func (k Keeper) DecreaseSupply(ctx sdk.Context, marker types.MarkerAccountI, coin sdk.Coin) error {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "decrease_supply")

	inCirculation := sdk.NewCoin(marker.GetDenom(), k.bankKeeper.GetSupply(ctx, marker.GetDenom()).Amount)

	// Ensure the request will not send the total supply below zero
	if inCirculation.IsLT(coin) {
		return fmt.Errorf("cannot reduce marker total supply below zero %s, %v", coin.Denom, coin.Amount)
	}
	// ensure the current marker account is holding enough coin to cover burn request
	escrow := k.bankKeeper.GetBalance(ctx, marker.GetAddress(), marker.GetDenom())
	if !escrow.Amount.GTE(coin.Amount) {
		return fmt.Errorf("marker account contains insufficient funds to burn %s, %v", coin.Denom, coin.Amount)
	}
	// Update the supply (abort if this can not be done)
	inCirculation = inCirculation.Sub(coin)
	if marker.HasFixedSupply() {
		if err := marker.SetSupply(inCirculation); err != nil {
			return err
		}
		if err := marker.Validate(); err != nil {
			return err
		}
		// Finalize supply update in marker record
		k.SetMarker(ctx, marker)
	}

	// Adjust circulation to match configured supply.
	if err := k.AdjustCirculation(ctx, marker, inCirculation); err != nil {
		panic(err)
	}

	return nil
}

// FinalizeMarker sets the state of the marker to finalized, mints the associated supply, assigns the minted coin to
// the marker accounts, and if successful emits an EventMarkerFinalize event to transition the state to active
func (k Keeper) FinalizeMarker(ctx sdk.Context, caller sdk.Address, denom string) error {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "finalize")

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
	preexistingCoin := sdk.NewCoin(m.GetDenom(), k.bankKeeper.GetSupply(ctx, m.GetDenom()).Amount)

	// If the requested total is less than the existing total, the supply invariant would halt the chain if activated
	if supplyRequest.IsLT(preexistingCoin) {
		return fmt.Errorf("marker supply %v has been defined as less than pre-existing"+
			" supply %v, can not finalize marker", supplyRequest, preexistingCoin)
	}

	// transition to finalized state ... then to active once mint is complete
	if err = m.SetStatus(types.StatusFinalized); err != nil {
		return fmt.Errorf("could not transition marker account state to finalized: %w", err)
	}
	if err := m.Validate(); err != nil {
		return err
	}
	k.SetMarker(ctx, m)

	// record status as finalized.
	markerFinalizeEvent := types.NewEventMarkerFinalize(denom, caller.String())
	if err := ctx.EventManager().EmitTypedEvent(markerFinalizeEvent); err != nil {
		return err
	}

	return nil
}

// ActivateMarker transistions a marker into the active status, enforcing permissions, supply constraints, and minting
// any supply as required.
func (k Keeper) ActivateMarker(ctx sdk.Context, caller sdk.Address, denom string) error {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "activate")

	m, err := k.GetMarkerByDenom(ctx, denom)
	if err != nil {
		return fmt.Errorf("marker not found for %s: %s", denom, err)
	}
	// Only the manger can activate a marker
	if !m.GetManager().Equals(caller) {
		return fmt.Errorf("%s does not have permission to activate %s markeraccount", caller, m.GetDenom())
	}

	// must be in finalized state ... mint required supply amounts.
	if m.GetStatus() != types.StatusFinalized {
		return fmt.Errorf("can only activate markeraccounts in the Finalized status")
	}

	// Verify the send_enabled status of this coin denom matches the marker types
	switch m.GetMarkerType() {
	case types.MarkerType_Coin:
		k.ensureSendEnabledStatus(ctx, denom, true)
	case types.MarkerType_RestrictedCoin:
		k.ensureSendEnabledStatus(ctx, denom, false)
	default:
		return fmt.Errorf("marker of %s type can not be activated", m.GetMarkerType())
	}

	// Amount to mint is typically the defined supply however...
	supplyRequest := m.GetSupply()

	// Any pre-existing coin amounts for our denom need to be removed from our amount to mint
	preexistingCoin := sdk.NewCoin(m.GetDenom(), k.bankKeeper.GetSupply(ctx, m.GetDenom()).Amount)

	// If the requested total is less than the existing total, the supply invariant would halt the chain if activated
	if supplyRequest.IsLT(preexistingCoin) {
		return fmt.Errorf("marker supply %v has been defined as less than pre-existing"+
			" supply %v, can not finalize marker", supplyRequest, preexistingCoin)
	}

	// Ensure the supply amount requested is minted and placed in the marker's account
	if err = k.AdjustCirculation(ctx, m, supplyRequest); err != nil {
		return err
	}

	// With the coin supply minted and assigned to the marker we can transition to the Active state.
	// this will enable the Invariant supply enforcement constraint.
	if err = m.SetStatus(types.StatusActive); err != nil {
		return fmt.Errorf("could not set marker status to active: %w", err)
	}
	if err := m.Validate(); err != nil {
		return err
	}
	// record status as active
	k.SetMarker(ctx, m)

	markerActivateEvent := types.NewEventMarkerActivate(denom, caller.String())
	if err := ctx.EventManager().EmitTypedEvent(markerActivateEvent); err != nil {
		return err
	}

	return nil
}

// CancelMarker prepares transition to deleted state.
func (k Keeper) CancelMarker(ctx sdk.Context, caller sdk.AccAddress, denom string) error {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "cancel")

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
		totalSupply := k.bankKeeper.GetSupply(ctx, m.GetDenom()).Amount
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
	if err := m.Validate(); err != nil {
		return err
	}
	k.SetMarker(ctx, m)

	markerCancelEvent := types.NewEventMarkerCancel(denom, caller.String())
	if err := ctx.EventManager().EmitTypedEvent(markerCancelEvent); err != nil {
		return err
	}

	return nil
}

// DeleteMarker burns the entire coin supply, ensure no assets are pooled, and marks the current instance of the
// marker as destroyed.
func (k Keeper) DeleteMarker(ctx sdk.Context, caller sdk.AccAddress, denom string) error {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "delete")

	m, err := k.GetMarkerByDenom(ctx, denom)
	if err != nil {
		return fmt.Errorf("marker not found for %s: %s", denom, err)
	}

	// either the manager [set if a proposed marker was cancelled] or someone assigned `delete` can perform this action
	if !(m.GetManager().Equals(caller) || m.AddressHasAccess(caller, types.Access_Delete)) {
		return fmt.Errorf("%s does not have %s on %s markeraccount", caller, types.Access_Delete, m.GetDenom())
	}

	// status must currently be set to cancelled
	if m.GetStatus() != types.StatusCancelled {
		return fmt.Errorf("can only delete markeraccounts in the Cancelled status")
	}

	// require full supply of coin for marker to be contained within the marker account (no outstanding delegations)
	totalSupply := k.bankKeeper.GetSupply(ctx, denom).Amount
	escrow := k.bankKeeper.GetAllBalances(ctx, m.GetAddress())
	inCirculation := totalSupply.Sub(escrow.AmountOf(denom))
	if inCirculation.GT(sdk.ZeroInt()) {
		return fmt.Errorf("cannot delete marker with %d minted coin in circulation out of %d total."+
			" ensure marker account holds the entire supply of %s", inCirculation, totalSupply, denom)
	}

	err = k.DecreaseSupply(ctx, m, sdk.NewCoin(denom, totalSupply))
	if err != nil {
		return fmt.Errorf("could not decrease marker supply %s: %s", denom, err)
	}

	escrow = k.bankKeeper.GetAllBalances(ctx, m.GetAddress())
	if !escrow.IsZero() {
		return fmt.Errorf("can not destroy marker due to balances in escrow: %s", escrow)
	}

	// get the updated state of the marker afer supply burn...
	m, err = k.GetMarkerByDenom(ctx, denom)
	if err != nil {
		return fmt.Errorf("marker not found for %s: %s", denom, err)
	}
	if err = m.SetStatus(types.StatusDestroyed); err != nil {
		return fmt.Errorf("could not update marker status: %w", err)
	}
	if err := m.Validate(); err != nil {
		return err
	}
	k.SetMarker(ctx, m)

	markerDeleteEvent := types.NewEventMarkerDelete(denom, caller.String())
	if err := ctx.EventManager().EmitTypedEvent(markerDeleteEvent); err != nil {
		return err
	}

	return nil
}

// TransferCoin transfers restricted coins between to accounts when the administrator account holds the transfer
// access right and the marker type is restricted_coin
func (k Keeper) TransferCoin(ctx sdk.Context, from, to, admin sdk.AccAddress, amount sdk.Coin) error {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "transfer_coin")

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
	if !admin.Equals(from) {
		err = k.authzHandler(ctx, admin, from, amount)
		if err != nil {
			return err
		}
	}
	if k.bankKeeper.BlockedAddr(to) {
		return fmt.Errorf("%s is not allowed to receive funds", to)
	}

	// send the coins between accounts (does not check send_enabled on coin denom)
	if err = k.bankKeeper.SendCoins(ctx, from, to, sdk.NewCoins(amount)); err != nil {
		return err
	}

	markerTransferEvent := types.NewEventMarkerTransfer(
		amount.Amount.String(),
		amount.Denom,
		admin.String(),
		to.String(),
		from.String(),
	)
	if err := ctx.EventManager().EmitTypedEvent(markerTransferEvent); err != nil {
		return err
	}

	return nil
}

func (k Keeper) authzHandler(ctx sdk.Context, admin sdk.AccAddress, from sdk.AccAddress, amount sdk.Coin) error {
	markerAuth := types.MarkerTransferAuthorization{}
	authorization, expireTime := k.authzKeeper.GetCleanAuthorization(ctx, admin, from, markerAuth.MsgTypeURL())
	if authorization == nil {
		return fmt.Errorf("%s account has not been granted authority to withdraw from %s account", admin, from)
	}
	accept, err := authorization.Accept(ctx, &types.MsgTransferRequest{Amount: amount})
	if err != nil {
		return err
	}
	if accept.Accept {
		limitLeft, _ := authorization.(*types.MarkerTransferAuthorization).DecreaseTransferLimit(amount)
		if limitLeft.IsZero() {
			return k.authzKeeper.DeleteGrant(ctx, admin, from, markerAuth.MsgTypeURL())
		}
		return k.authzKeeper.SaveGrant(ctx, admin, from, &types.MarkerTransferAuthorization{TransferLimit: limitLeft}, expireTime)
	}
	return fmt.Errorf("authorization was not accepted for %s", admin)
}

// SetMarkerDenomMetadata updates the denom metadata records for the current marker.
func (k Keeper) SetMarkerDenomMetadata(ctx sdk.Context, metadata banktypes.Metadata, caller sdk.AccAddress) error {
	defer telemetry.MeasureSince(time.Now(), types.ModuleName, "set_marker_denom_metadata")

	if metadata.Base == "" {
		return fmt.Errorf("invalid metadata request, base denom must match existing marker")
	}
	marker, markerErr := k.GetMarkerByDenom(ctx, metadata.Base)
	if markerErr != nil {
		return fmt.Errorf("marker not found for %s: %w", metadata.Base, markerErr)
	}
	if !marker.GetManager().Equals(caller) && !marker.AddressHasAccess(caller, types.Access_Admin) {
		return fmt.Errorf("%s is not allowed to manage marker metadata", caller.String())
	}

	var existing *banktypes.Metadata
	if e, _ := k.bankKeeper.GetDenomMetaData(ctx, metadata.Base); len(e.Base) > 0 {
		existing = &e
	}

	if err := k.ValidateDenomMetadata(ctx, metadata, existing, marker.GetStatus()); err != nil {
		return err
	}

	// record the metadata with the bank
	k.bankKeeper.SetDenomMetaData(ctx, metadata)

	markerSetDenomMetaEvent := types.NewEventMarkerSetDenomMetadata(
		metadata,
		caller.String(),
	)
	if err := ctx.EventManager().EmitTypedEvent(markerSetDenomMetaEvent); err != nil {
		return err
	}

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

// ensureSendEnabledStatus checks to see if the configuration of SendEnabled for the current network matches
// the requested value, sets
func (k Keeper) ensureSendEnabledStatus(ctx sdk.Context, denom string, sendEnabled bool) {
	if k.bankKeeper.IsSendEnabledDenom(ctx, denom) != sendEnabled {
		switch k.bankKeeper.GetParams(ctx).DefaultSendEnabled {
		case sendEnabled:
			k.bankKeeper.DeleteSendEnabled(ctx, denom)
		default:
			k.bankKeeper.SetSendEnabled(ctx, denom, sendEnabled)
		}
	}
}
