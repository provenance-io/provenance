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
	newMarker.MarkerType = c.MarkerType

	if err := newMarker.SetSupply(c.Amount); err != nil {
		return err
	}

	if err := newMarker.SetStatus(c.Status); err != nil {
		return err
	}

	if err := newMarker.Validate(); err != nil {
		return err
	}

	if err := k.AddMarkerAccount(ctx, newMarker); err != nil {
		return err
	}

	// active markers should have supply set.
	if newMarker.Status == types.StatusActive {
		if err := k.AdjustCirculation(ctx, newMarker, c.Amount); err != nil {
			return err
		}
	}

	logger := k.Logger(ctx)
	logger.Info("a new marker was added", "marker", c.Amount.Denom, "supply", c.Amount.String())

	return nil
}

// HandleSupplyIncreaseProposal handles a SupplyIncrease governance proposal request
func HandleSupplyIncreaseProposal(ctx sdk.Context, k Keeper, c *types.SupplyIncreaseProposal) error {
	logger := k.Logger(ctx)
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

	if m.GetStatus() == types.StatusProposed || m.GetStatus() == types.StatusFinalized {
		total := m.GetSupply().Add(c.Amount)
		if err = m.SetSupply(total); err != nil {
			return err
		}
		if err := m.Validate(); err != nil {
			return err
		}
		k.SetMarker(ctx, m)
		logger.Info("marker configured supply increased", "marker", c.Amount.Denom, "amount", c.Amount.Amount.String())
		return nil
	} else if m.GetStatus() != types.StatusActive {
		return fmt.Errorf("cannot mint coin for a marker that is not in Active status")
	}

	if err := k.IncreaseSupply(ctx, m, c.Amount); err != nil {
		return err
	}

	logger.Info("marker total supply increased", "marker", c.Amount.Denom, "amount", c.Amount.Amount.String())

	// If a target address for minted coins is given then send them there.
	if len(c.TargetAddress) > 0 {
		recipient, err := sdk.AccAddressFromBech32(c.TargetAddress)
		if err != nil {
			return err
		}

		if err := k.bankKeeper.InputOutputCoins(ctx, []banktypes.Input{banktypes.NewInput(addr, sdk.NewCoins(c.Amount))},
			[]banktypes.Output{banktypes.NewOutput(recipient, sdk.NewCoins(c.Amount))}); err != nil {
			return err
		}

		logger.Info("transferred escrowed coin from marker", "marker", c.Amount.Denom, "amount", c.Amount.String(), "recipient", c.TargetAddress)
	}

	return nil
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

	if err := k.DecreaseSupply(ctx, m, c.Amount); err != nil {
		return err
	}

	logger := k.Logger(ctx)
	logger.Info("marker total supply reduced", "marker", c.Amount.Denom, "amount", c.Amount.Amount.String())

	return nil
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
		if err := m.GrantAccess(types.NewAccessGrant(a.GetAddress(), a.Permissions)); err != nil {
			return err
		}
		logger := k.Logger(ctx)
		logger.Info("controlling access to marker assigned ", "marker", c.Denom, "access", a.String())
	}

	if err := m.Validate(); err != nil {
		return err
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

	if err := m.Validate(); err != nil {
		return err
	}

	k.SetMarker(ctx, m)

	logger := k.Logger(ctx)
	logger.Info("marker access revoked", "marker", c.Denom, "administrator", c.RemovedAddress)

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
	if c.NewStatus == types.StatusUndefined {
		return fmt.Errorf("error invalid marker status undefined")
	}
	if int(m.GetStatus()) > int(c.NewStatus) {
		return fmt.Errorf("invalid status transition %s precedes existing status of %s", c.NewStatus, m.GetStatus())
	}

	// activate (must be pending, finalized currently)
	if c.NewStatus == types.StatusActive {
		if err = k.AdjustCirculation(ctx, m, m.GetSupply()); err != nil {
			return fmt.Errorf("could not create marker supply: %w", err)
		}
	}

	// delete (must be cancelled currently)
	if c.NewStatus == types.StatusDestroyed {
		if m.GetStatus() != types.StatusCancelled {
			return fmt.Errorf("only cancelled markers can be deleted")
		}
		if err = k.AdjustCirculation(ctx, m, sdk.NewCoin(c.Denom, sdk.ZeroInt())); err != nil {
			return fmt.Errorf("could not dispose of marker supply: %w", err)
		}
	}

	if err := m.SetStatus(c.NewStatus); err != nil {
		return err
	}

	if err := m.Validate(); err != nil {
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

	recipient, err := sdk.AccAddressFromBech32(c.TargetAddress)
	if err != nil {
		return err
	}

	if err := k.bankKeeper.InputOutputCoins(ctx, []banktypes.Input{banktypes.NewInput(addr, c.Amount)},
		[]banktypes.Output{banktypes.NewOutput(recipient, c.Amount)}); err != nil {
		return err
	}

	logger := k.Logger(ctx)
	logger.Info("transferred escrowed coin from marker", "marker", c.Denom, "amount", c.Amount.String(), "recipient", c.TargetAddress)

	return nil
}

// HandleSetDenomMetadataProposal handles a Set Denom Metadata governance proposal request
func HandleSetDenomMetadataProposal(ctx sdk.Context, k Keeper, c *types.SetDenomMetadataProposal) error {
	addr, err := types.MarkerAddress(c.Metadata.Base)
	if err != nil {
		return err
	}
	m, err := k.GetMarker(ctx, addr)
	if err != nil {
		return err
	}
	if m == nil {
		return fmt.Errorf("%s marker does not exist", c.Metadata.Base)
	}
	if !m.HasGovernanceEnabled() {
		return fmt.Errorf("%s marker does not allow governance control", c.Metadata.Base)
	}

	k.bankKeeper.SetDenomMetaData(ctx, c.Metadata)

	k.Logger(ctx).Info("denom metadata set for marker", "marker", c.Metadata.Base, "denom metadata", c.Metadata.String())
	return nil
}
