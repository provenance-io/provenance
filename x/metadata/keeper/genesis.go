package keeper

import (
	"github.com/provenance-io/provenance/x/metadata/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis creates the initial genesis state for the metadata module.
func (k Keeper) InitGenesis(ctx sdk.Context, data *types.GenesisState) {
	if err := data.Validate(); err != nil {
		panic(err)
	}
	if data.Scopes != nil {
		for _, s := range data.Scopes {
			k.SetScope(ctx, s)
		}
	}
	if data.Groups != nil {
		for _, r := range data.Groups {
			k.SetRecordGroup(ctx, r)
		}
	}
	if data.Records != nil {
		for _, r := range data.Records {
			k.SetRecord(ctx, r)
		}
	}
	if data.ScopeSpecifications != nil {
		for _, s := range data.ScopeSpecifications {
			k.SetScopeSpecification(ctx, s)
		}
	}
	if data.ContractSpecifications != nil {
		for _, s := range data.ContractSpecifications {
			k.SetContractSpecification(ctx, s)
		}
	}
	if data.RecordSpecifications != nil {
		for _, s := range data.RecordSpecifications {
			k.SetRecordSpecification(ctx, s)
		}
	}
}

// ExportGenesis exports the current keeper state of the metadata module.ExportGenesis
func (k Keeper) ExportGenesis(ctx sdk.Context) (data *types.GenesisState) {
	params := k.GetParams(ctx)
	scopes := make([]types.Scope, 0)
	groups := make([]types.RecordGroup, 0)
	records := make([]types.Record, 0)
	scopeSpecs := make([]types.ScopeSpecification, 0)
	contractSpecs := make([]types.ContractSpecification, 0)
	recordSpecs := make([]types.RecordSpecification, 0)

	appendToScopes := func(scope types.Scope) bool {
		scopes = append(scopes, scope)
		return false
	}

	appendToGroups := func(group types.RecordGroup) bool {
		groups = append(groups, group)
		return false
	}

	appendToRecords := func(record types.Record) bool {
		records = append(records, record)
		return false
	}

	appendToScopeSpecs := func(scopeSpec types.ScopeSpecification) bool {
		scopeSpecs = append(scopeSpecs, scopeSpec)
		return false
	}

	appendToContractSpecs := func(contractSpec types.ContractSpecification) bool {
		contractSpecs = append(contractSpecs, contractSpec)
		return false
	}

	appendToRecordSpecs := func(recordSpec types.RecordSpecification) bool {
		recordSpecs = append(recordSpecs, recordSpec)
		return false
	}

	if err := k.IterateScopes(ctx, appendToScopes); err != nil {
		panic(err)
	}
	if err := k.IterateGroups(ctx, types.MetadataAddress{}, appendToGroups); err != nil {
		panic(err)
	}
	if err := k.IterateRecords(ctx, types.MetadataAddress{}, appendToRecords); err != nil {
		panic(err)
	}
	if err := k.IterateScopeSpecs(ctx, appendToScopeSpecs); err != nil {
		panic(err)
	}
	if err := k.IterateContractSpecs(ctx, appendToContractSpecs); err != nil {
		panic(err)
	}
	if err := k.IterateRecordSpecs(ctx, appendToRecordSpecs); err != nil {
		panic(err)
	}

	return types.NewGenesisState(params, scopes, groups, records, scopeSpecs, contractSpecs, recordSpecs)
}
