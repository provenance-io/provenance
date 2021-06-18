// Package wasm supports smart contract integration with the metadata module.
package wasm

import (
	"encoding/json"
	"fmt"

	"github.com/provenance-io/provenance/internal/provwasm"
	"github.com/provenance-io/provenance/x/metadata/keeper"
	"github.com/provenance-io/provenance/x/metadata/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MetadataQueryParams represents the query request type for the metadata module sent by smart contracts.
// Only one query field should be set.
type MetadataQueryParams struct {
	// Get a scope by ID.
	Get *GetScopeParams `json:"get_scope,omitempty"`
}

// GetScopeParams are the inputs for a scope query.
type GetScopeParams struct {
	// The bech32 address of the scope we want to get.
	ScopeID string `json:"scope_id"`
}

// Querier returns a smart contract querier for the metadata module.
func Querier(keeper keeper.Keeper) provwasm.Querier {
	return func(ctx sdk.Context, query json.RawMessage, version string) ([]byte, error) {
		wrapper := struct {
			Params *MetadataQueryParams `json:"metadata"`
		}{}
		if err := json.Unmarshal(query, &wrapper); err != nil {
			return nil, fmt.Errorf("wasm: invalid metadata query params: %w", err)
		}
		params := wrapper.Params
		if params == nil {
			return nil, fmt.Errorf("wasm: nil metadata query params")
		}
		switch {
		case params.Get != nil:
			return params.Get.Run(ctx, keeper)
		default:
			return nil, fmt.Errorf("wasm: invalid metadata query: %s", string(query))
		}
	}
}

// Run gets a scope by ID.
func (params *GetScopeParams) Run(ctx sdk.Context, keeper keeper.Keeper) ([]byte, error) {
	scopeID, err := types.MetadataAddressFromBech32(params.ScopeID)
	if err != nil {
		return nil, fmt.Errorf("wasm: invalid scope ID: %w", err)
	}
	scope, found := keeper.GetScope(ctx, scopeID)
	if !found {
		return nil, fmt.Errorf("wasm: scope not found: %s", params.ScopeID)

	}
	return createScopeResponse(scope)
}
