package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/provenance-io/provenance/x/marker/types"
)

// SupplyIncreaseProposal handles a SupplyIncrease governance proposal request
func (k Keeper) SupplyIncreaseProposal(ctx sdk.Context, amount sdk.Coin, targetAddress string) error {
	logger := k.Logger(ctx)
	addr, err := types.MarkerAddress(amount.Denom)
	if err != nil {
		return err
	}
	m, err := k.GetMarker(ctx, addr)
	if err != nil {
		return err
	}
	if m == nil {
		return fmt.Errorf("%s marker does not exist", amount.Denom)
	}
	if !m.HasGovernanceEnabled() {
		return fmt.Errorf("%s marker does not allow governance control", amount.Denom)
	}

	if m.GetStatus() == types.StatusProposed || m.GetStatus() == types.StatusFinalized {
		total := m.GetSupply().Add(amount)
		if err = m.SetSupply(total); err != nil {
			return err
		}
		if err := m.Validate(); err != nil {
			return err
		}
		k.SetMarker(ctx, m)
		logger.Info("marker configured supply increased", "marker", amount.Denom, "amount", amount.Amount.String())
		return nil
	} else if m.GetStatus() != types.StatusActive {
		return fmt.Errorf("cannot mint coin for a marker that is not in Active status")
	}

	if err := k.IncreaseSupply(ctx, m, amount); err != nil {
		return err
	}

	logger.Info("marker total supply increased", "marker", amount.Denom, "amount", amount.Amount.String())

	// If a target address for minted coins is given then send them there.
	if len(targetAddress) > 0 {
		recipient, err := sdk.AccAddressFromBech32(targetAddress)
		if err != nil {
			return err
		}
		if err := k.bankKeeper.SendCoins(types.WithBypass(ctx), addr, recipient, sdk.NewCoins(amount)); err != nil {
			return err
		}
		logger.Info("transferred escrowed coin from marker", "marker", amount.Denom, "amount", amount.String(), "recipient", targetAddress)
	}

	return nil
}

// HandleSupplyDecreaseProposal handles a SupplyDecrease governance proposal request
func (k Keeper) HandleSupplyDecreaseProposal(ctx sdk.Context, amount sdk.Coin) error {
	addr, err := types.MarkerAddress(amount.Denom)
	if err != nil {
		return err
	}
	m, err := k.GetMarker(ctx, addr)
	if err != nil {
		return err
	}
	if m == nil {
		return fmt.Errorf("%s marker does not exist", amount.Denom)
	}

	if !m.HasGovernanceEnabled() {
		return fmt.Errorf("%s marker does not allow governance control", amount.Denom)
	}

	if err := k.DecreaseSupply(ctx, m, amount); err != nil {
		return err
	}

	logger := k.Logger(ctx)
	logger.Info("marker total supply reduced", "marker", amount.Denom, "amount", amount.Amount.String())

	return nil
}

// SetAdministratorProposal handles a SetAdministrator governance proposal request
func (k Keeper) SetAdministratorProposal(ctx sdk.Context, denom string, accessGrants []types.AccessGrant) error {
	addr, err := types.MarkerAddress(denom)
	if err != nil {
		return err
	}
	m, err := k.GetMarker(ctx, addr)
	if err != nil {
		return err
	}
	if m == nil {
		return fmt.Errorf("%s marker does not exist", denom)
	}
	if !m.HasGovernanceEnabled() {
		return fmt.Errorf("%s marker does not allow governance control", denom)
	}
	for _, a := range accessGrants {
		if err := m.GrantAccess(types.NewAccessGrant(a.GetAddress(), a.Permissions)); err != nil {
			return err
		}
		logger := k.Logger(ctx)
		logger.Info("controlling access to marker assigned ", "marker", denom, "access", a.String())
	}

	if err := m.Validate(); err != nil {
		return err
	}

	k.SetMarker(ctx, m)
	return nil
}

// RemoveAdministratorProposal handles a RemoveAdministrator governance proposal request
func (k Keeper) RemoveAdministratorProposal(ctx sdk.Context, denom string, removedAddress []string) error {
	addr, err := types.MarkerAddress(denom)
	if err != nil {
		return err
	}
	m, err := k.GetMarker(ctx, addr)
	if err != nil {
		return err
	}
	if m == nil {
		return fmt.Errorf("%s marker does not exist", denom)
	}
	if !m.HasGovernanceEnabled() {
		return fmt.Errorf("%s marker does not allow governance control", denom)
	}
	for _, a := range removedAddress {
		addr, err := sdk.AccAddressFromBech32(a)
		if err != nil {
			return err
		}
		if err = m.RevokeAccess(addr); err != nil {
			return err
		}
	}

	if err := m.Validate(); err != nil {
		return err
	}

	k.SetMarker(ctx, m)

	logger := k.Logger(ctx)
	logger.Info("marker access revoked", "marker", denom, "administrator", removedAddress)

	return nil
}

// ChangeStatusProposal handles a ChangeStatus governance proposal request
func (k Keeper) ChangeStatusProposal(ctx sdk.Context, denom string, status types.MarkerStatus) error {
	addr, err := types.MarkerAddress(denom)
	if err != nil {
		return err
	}
	m, err := k.GetMarker(ctx, addr)
	if err != nil {
		return err
	}
	if m == nil {
		return fmt.Errorf("%s marker does not exist", denom)
	}
	if !m.HasGovernanceEnabled() {
		return fmt.Errorf("%s marker does not allow governance control", denom)
	}
	if status == types.StatusUndefined {
		return fmt.Errorf("error invalid marker status undefined")
	}
	if int(m.GetStatus()) > int(status) {
		return fmt.Errorf("invalid status transition %s precedes existing status of %s", status, m.GetStatus())
	}

	// activate (must be pending, finalized currently)
	if status == types.StatusActive {
		if err = k.AdjustCirculation(ctx, m, m.GetSupply()); err != nil {
			return fmt.Errorf("could not create marker supply: %w", err)
		}
	}

	// delete (must be cancelled currently)
	if status == types.StatusDestroyed {
		if m.GetStatus() != types.StatusCancelled {
			return fmt.Errorf("only cancelled markers can be deleted")
		}
		if err = k.AdjustCirculation(ctx, m, sdk.NewInt64Coin(denom, 0)); err != nil {
			return fmt.Errorf("could not dispose of marker supply: %w", err)
		}
	}

	if err := m.SetStatus(status); err != nil {
		return err
	}

	if err := m.Validate(); err != nil {
		return err
	}

	k.SetMarker(ctx, m)

	logger := k.Logger(ctx)
	logger.Info("changed marker status", "marker", denom, "stats", status.String())

	return nil
}

// WithdrawEscrowProposal handles a Withdraw escrowed coins governance proposal request
func (k Keeper) WithdrawEscrowProposal(ctx sdk.Context, denom, targetAddress string, amount sdk.Coins) error {
	addr, err := types.MarkerAddress(denom)
	if err != nil {
		return err
	}
	m, err := k.GetMarker(ctx, addr)
	if err != nil {
		return err
	}
	if m == nil {
		return fmt.Errorf("%s marker does not exist", denom)
	}
	if !m.HasGovernanceEnabled() {
		return fmt.Errorf("%s marker does not allow governance control", denom)
	}

	recipient, err := sdk.AccAddressFromBech32(targetAddress)
	if err != nil {
		return err
	}
	if err := k.bankKeeper.SendCoins(types.WithBypass(ctx), addr, recipient, amount); err != nil {
		return err
	}
	logger := k.Logger(ctx)
	logger.Info("transferred escrowed coin from marker", "marker", denom, "amount", amount, "recipient", targetAddress)

	return nil
}

// SetDenomMetadataProposal handles a Set Denom Metadata governance proposal request
func (k Keeper) SetDenomMetadataProposal(ctx sdk.Context, metadata banktypes.Metadata) error {
	addr, err := types.MarkerAddress(metadata.Base)
	if err != nil {
		return err
	}
	m, err := k.GetMarker(ctx, addr)
	if err != nil {
		return err
	}
	if m == nil {
		return fmt.Errorf("%s marker does not exist", metadata.Base)
	}
	if !m.HasGovernanceEnabled() {
		return fmt.Errorf("%s marker does not allow governance control", metadata.Base)
	}

	k.bankKeeper.SetDenomMetaData(ctx, metadata)

	k.Logger(ctx).Info("denom metadata set for marker", "marker", metadata.Base, "denom metadata", metadata.String())
	return nil
}
