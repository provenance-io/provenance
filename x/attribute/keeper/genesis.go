package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/attribute/types"
)

// InitGenesis creates the initial genesis state for the attribute module.
func (k Keeper) InitGenesis(ctx sdk.Context, data *types.GenesisState) {
	k.SetParams(ctx, data.Params)
	if err := data.ValidateBasic(); err != nil {
		panic(err)
	}
	for _, attr := range data.Attributes {
		if err := k.importAttribute(ctx, attr); err != nil {
			panic(err)
		}
	}

	if err := EnsureModuleAccountAndAccountDataNameRecord(ctx.WithLogger(log.NewNopLogger()), k.authKeeper, k.nameKeeper); err != nil {
		panic(err)
	}
}

// EnsureModuleAccountAndAccountDataNameRecord makes sure that the attribute module account exists and that
// the account data name record exists, is restricted, and owned by the attribute module.
// An error is returned if any of that cannot be made to be the way we want it.
func EnsureModuleAccountAndAccountDataNameRecord(ctx sdk.Context, accountK types.AccountKeeper, nameK types.NameKeeper) (err error) {
	// Note: The logging in here is primarily for its use during an upgrade.
	defer func() {
		if err != nil {
			ctx.Logger().Error(fmt.Sprintf("Error setting %q name record.", types.AccountDataName), "error", err)
		}
	}()
	ctx.Logger().Info(fmt.Sprintf("Setting %q name record.", types.AccountDataName))

	// Make sure that the module account exists and get its address. GetModuleAccount creates it if it doesn't exist.
	attrModAcc := accountK.GetModuleAccount(ctx, types.ModuleName)
	attrModAccAddr := attrModAcc.GetAddress()

	// If the name doesn't exist yet, create it and we're done here.
	if !nameK.NameExists(ctx, types.AccountDataName) {
		err = nameK.SetNameRecord(ctx, types.AccountDataName, attrModAccAddr, true)
		if err != nil {
			return err
		}
		ctx.Logger().Info(fmt.Sprintf("Successfully set %q name record.", types.AccountDataName))
		return nil
	}

	// If it already exists, but has a different address or isn't restricted, update it to what we want it to be.
	existing, err := nameK.GetRecordByName(ctx, types.AccountDataName)
	if err != nil {
		return err
	}
	updateNeeded := false
	if !existing.Restricted {
		updateNeeded = true
		ctx.Logger().Info(fmt.Sprintf("Existing %q name record is not restricted. It will be updated to be restricted.", types.AccountDataName))
	}
	attrModAccAddrStr := attrModAccAddr.String()
	if existing.Address != attrModAccAddrStr {
		updateNeeded = true
		ctx.Logger().Info(fmt.Sprintf("Existing %q name record has address %q. It will be updated to the attribute module account address %q",
			types.AccountDataName, existing.Address, attrModAccAddrStr))
	}
	if !updateNeeded {
		ctx.Logger().Info(fmt.Sprintf("The %q name record already exists as needed. Nothing to do.", types.AccountDataName))
		return nil
	}
	ctx.Logger().Info(fmt.Sprintf("Updating existing %q name record.", types.AccountDataName))
	err = nameK.UpdateNameRecord(ctx, types.AccountDataName, attrModAccAddr, true)
	if err != nil {
		return err
	}

	ctx.Logger().Info(fmt.Sprintf("Successfully updated %q name record.", types.AccountDataName))
	return nil
}

// ExportGenesis exports the current keeper state of the attribute module.
func (k Keeper) ExportGenesis(ctx sdk.Context) (data *types.GenesisState) {
	attrs := make([]types.Attribute, 0)
	params := k.GetParams(ctx)

	appendToRecords := func(attr types.Attribute) error {
		attrs = append(attrs, attr)
		return nil
	}

	if err := k.IterateRecords(ctx, types.AttributeKeyPrefix, appendToRecords); err != nil {
		panic(err)
	}

	return types.NewGenesisState(params, attrs)
}
