package keeper

import (
	"context"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/smartaccounts/types"
)

func (k Keeper) ImportGenesis(ctx context.Context, data *types.GenesisState) error {
	// Set module parameters
	if err := k.SetParams(ctx, data.Params); err != nil {
		return fmt.Errorf("failed to set params: %w", err)
	}
	for _, record := range data.Accounts {
		addr, err := sdk.AccAddressFromBech32(record.Address)
		if err != nil {
			panic(err)
		}
		if _, err := k.SaveAccountDetails(ctx, &record, addr); err != nil {
			panic(err)
		}
	}
	return nil
}

func (k Keeper) ExportGenesis(ctx context.Context) (*types.GenesisState, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		panic(err)
	}
	var smartAccounts []types.ProvenanceAccount
	appendToSmartAccounts := func(sa types.ProvenanceAccount) bool {
		smartAccounts = append(smartAccounts, sa)
		return false
	}

	k.IterateSmartAccounts(ctx, appendToSmartAccounts)
	return NewGenesisState(*params, smartAccounts), nil
}

// DefaultGenesisState returns the initial module genesis state.
func DefaultGenesisState() *types.GenesisState {
	return NewGenesisState(types.DefaultParams(), []types.ProvenanceAccount{})
}

// NewGenesisState creates a new GenesisState object
func NewGenesisState(params types.Params, smartAccounts []types.ProvenanceAccount) *types.GenesisState {
	return &types.GenesisState{
		Params:   params,
		Accounts: smartAccounts,
	}
}
