package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/provenance-io/provenance/x/marker/types"
)

// HandleAddMarkerProposal handles an Add Marker governance proposal request
func HandleAddMarkerProposal(ctx sdk.Context, k Keeper, c *types.AddMarkerProposal) error {
	addr, err := types.MarkerAddress(c.Amount.Denom)
	if err != nil {
		return err
	}
	existing, err := k.GetMarker(ctx, addr)
	if err != nil {
		return err
	}
	if existing != nil {
		return fmt.Errorf("%s marker already exists", c.Amount.Denom)
	}

	newMarker := types.NewEmptyMarkerAccount(c.Amount.Denom, c.Manager, c.AccessList)
	newMarker.AllowGovernanceControl = c.AllowGovernanceControl
	newMarker.SupplyFixed = c.SupplyFixed

	if err := newMarker.SetSupply(c.Amount); err != nil {
		return err
	}

	if err := newMarker.SetStatus(c.Status); err != nil {
		return err
	}

	if err := newMarker.Validate(); err != nil {
		return err
	}

	logger := k.Logger(ctx)
	logger.Info("a new marker was added", "marker", c.Amount.Denom, "supply", c.Amount.String())

	return k.AddMarkerAccount(ctx, newMarker)
}

// HandleSupplyIncreaseProposal handles a SupplyIncrease governance proposal request
func HandleSupplyIncreaseProposal(ctx sdk.Context, k Keeper, c *types.SupplyIncreaseProposal) error {
	addr, err := types.MarkerAddress(c.Amount.Denom)
	if err != nil {
		return err
	}
	m, err := k.GetMarker(ctx, addr)
	if err != nil {
		return err
	}
	if m == nil {
		return fmt.Errorf("%s marker does not exist", c.Amount.Denom)
	}
	if !m.HasGovernanceEnabled() {
		return fmt.Errorf("%s marker does not allow governance control", c.Amount.Denom)
	}

	logger := k.Logger(ctx)
	logger.Info("marker total supply increased", "marker", c.Amount.Denom, "amount", c.Amount.Amount.String())

	return k.IncreaseSupply(ctx, m, c.Amount)
}

// HandleSupplyDecreaseProposal handles a SupplyDecrease governance proposal request
func HandleSupplyDecreaseProposal(ctx sdk.Context, k Keeper, c *types.SupplyDecreaseProposal) error {
	addr, err := types.MarkerAddress(c.Amount.Denom)
	if err != nil {
		return err
	}
	m, err := k.GetMarker(ctx, addr)
	if err != nil {
		return err
	}
	if m == nil {
		return fmt.Errorf("%s marker does not exist", c.Amount.Denom)
	}

	if !m.HasGovernanceEnabled() {
		return fmt.Errorf("%s marker does not allow governance control", c.Amount.Denom)
	}

	logger := k.Logger(ctx)
	logger.Info("marker total supply reduced", "marker", c.Amount.Denom, "amount", c.Amount.Amount.String())

	return k.DecreaseSupply(ctx, m, c.Amount)
}

// HandleSetAdministratorProposal handles a SetAdministrator governance proposal request
func HandleSetAdministratorProposal(ctx sdk.Context, k Keeper, c *types.SetAdministratorProposal) error {
	addr, err := types.MarkerAddress(c.Denom)
	if err != nil {
		return err
	}
	m, err := k.GetMarker(ctx, addr)
	if err != nil {
		return err
	}
	if m == nil {
		return fmt.Errorf("%s marker does not exist", c.Denom)
	}
	if !m.HasGovernanceEnabled() {
		return fmt.Errorf("%s marker does not allow governance control", c.Denom)
	}
	for _, a := range c.Access {
		if err := m.GrantAccess(&a); err != nil {
			return err
		}
		logger := k.Logger(ctx)
		logger.Info("controlling access to marker assigned ", "marker", c.Denom, "access", a.String())
	}

	k.SetMarker(ctx, m)
	return nil
}

// HandleRemoveAdministratorProposal handles a RemoveAdministrator governance proposal request
func HandleRemoveAdministratorProposal(ctx sdk.Context, k Keeper, c *types.RemoveAdministratorProposal) error {
	addr, err := types.MarkerAddress(c.Denom)
	if err != nil {
		return err
	}
	m, err := k.GetMarker(ctx, addr)
	if err != nil {
		return err
	}
	if m == nil {
		return fmt.Errorf("%s marker does not exist", c.Denom)
	}
	if !m.HasGovernanceEnabled() {
		return fmt.Errorf("%s marker does not allow governance control", c.Denom)
	}
	for _, a := range c.RemovedAddress {
		addr, err := sdk.AccAddressFromBech32(a)
		if err != nil {
			return err
		}
		if err = m.RevokeAccess(addr); err != nil {
			return err
		}
	}

	logger := k.Logger(ctx)
	logger.Info("marker access revoked", "marker", c.Denom, "administrator", c.RemovedAddress)

	k.SetMarker(ctx, m)
	return nil
}

// HandleChangeStatusProposal handles a ChangeStatus governance proposal request
func HandleChangeStatusProposal(ctx sdk.Context, k Keeper, c *types.ChangeStatusProposal) error {
	addr, err := types.MarkerAddress(c.Denom)
	if err != nil {
		return err
	}
	m, err := k.GetMarker(ctx, addr)
	if err != nil {
		return err
	}
	if m == nil {
		return fmt.Errorf("%s marker does not exist", c.Denom)
	}
	if !m.HasGovernanceEnabled() {
		return fmt.Errorf("%s marker does not allow governance control", c.Denom)
	}

	if err := m.SetStatus(c.NewStatus); err != nil {
		return err
	}
	k.SetMarker(ctx, m)

	logger := k.Logger(ctx)
	logger.Info("changed marker status", "marker", c.Denom, "stats", c.NewStatus.String())

	return nil
}

// HandleWithdrawEscrowProposal handles a Withdraw escrowed coins governance proposal request
func HandleWithdrawEscrowProposal(ctx sdk.Context, k Keeper, c *types.WithdrawEscrowProposal) error {
	addr, err := types.MarkerAddress(c.Denom)
	if err != nil {
		return err
	}
	m, err := k.GetMarker(ctx, addr)
	if err != nil {
		return err
	}
	if m == nil {
		return fmt.Errorf("%s marker does not exist", c.Denom)
	}
	if !m.HasGovernanceEnabled() {
		return fmt.Errorf("%s marker does not allow governance control", c.Denom)
	}

	recipient, err := sdk.AccAddressFromBech32(c.TargetAdddress)
	if err != nil {
		return err
	}

	if err := k.bankKeeper.InputOutputCoins(ctx, []banktypes.Input{banktypes.NewInput(addr, c.Amount)},
		[]banktypes.Output{banktypes.NewOutput(recipient, c.Amount)}); err != nil {
		return err
	}

	logger := k.Logger(ctx)
	logger.Info("transferred escrowed coin from marker", "marker", c.Denom, "amount", c.Amount.String(), "recipient", c.TargetAdddress)

	return nil
}
