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
	if data.Sessions != nil {
		for _, r := range data.Sessions {
			k.SetSession(ctx, r)
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
	// TODO: data.RecordSpecifications
}

// ExportGenesis exports the current keeper state of the metadata module.ExportGenesis
func (k Keeper) ExportGenesis(ctx sdk.Context) (data *types.GenesisState) {
	params := k.GetParams(ctx)
	scopes := make([]types.Scope, 0)
	sessions := make([]types.Session, 0)
	records := make([]types.Record, 0)

	appendToScopes := func(scope types.Scope) bool {
		scopes = append(scopes, scope)
		return false
	}

	appendToSessions := func(session types.Session) bool {
		sessions = append(sessions, session)
		return false
	}

	appendToRecords := func(record types.Record) bool {
		records = append(records, record)
		return false
	}

	if err := k.IterateScopes(ctx, appendToScopes); err != nil {
		panic(err)
	}
	if err := k.IterateSessions(ctx, types.MetadataAddress{}, appendToSessions); err != nil {
		panic(err)
	}
	if err := k.IterateRecords(ctx, types.MetadataAddress{}, appendToRecords); err != nil {
		panic(err)
	}
	// TODO iterate over existing scope, group specifications and collect here for export

	return types.NewGenesisState(params, scopes, sessions, records, []types.ScopeSpecification{}, []types.ContractSpecification{})
}
