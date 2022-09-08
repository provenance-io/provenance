// Package wasm supports smart contract integration with the metadata module.
package wasm

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/internal/provwasm"
	"github.com/provenance-io/provenance/x/expiration/keeper"
	"github.com/provenance-io/provenance/x/expiration/types"
)

// ExpirationQueryParams represents the query request type for the expiration module sent by smart contracts.
// Only one query field should be set.
type ExpirationQueryParams struct {
	// Get a scope by ID.
	GetExpiration *GetExpirationParams `json:"get_expiration,omitempty"`
	// Get sessions by scope ID and name (optional).
	GetAllExpirations *GetAllExpirationsParams `json:"get_all_expirations,omitempty"`
	// Get records by scope ID and name (optional).
	GetAllExpirationsByOwner *GetAllExpirationsByOwnerParams `json:"get_all_expirations_by_owner,omitempty"`
}

// GetExpirationParams are the inputs for an expiration query.
type GetExpirationParams struct {
	// The bech32 address of the expiration we want to get.
	ModuleAssetID string `json:"module_asset_id"`
}

// GetAllExpirationsParams are the inputs for a session query.
type GetAllExpirationsParams struct {
}

// GetAllExpirationsByOwnerParams are the inputs for expirations by owner query.
type GetAllExpirationsByOwnerParams struct {
	// The bech32 address of the owner we want to get expirations for.
	Owner string `json:"owner"`
}

// Querier returns a smart contract querier for the metadata module.
func Querier(keeper keeper.Keeper) provwasm.Querier {
	return func(ctx sdk.Context, query json.RawMessage, version string) ([]byte, error) {
		wrapper := struct {
			Params *ExpirationQueryParams `json:"expiration"`
		}{}
		if err := json.Unmarshal(query, &wrapper); err != nil {
			return nil, fmt.Errorf("wasm: invalid %s query params: %w", types.ModuleName, err)
		}
		params := wrapper.Params
		if params == nil {
			return nil, fmt.Errorf("wasm: nil %s query params", types.ModuleName)
		}
		switch {
		case params.GetExpiration != nil:
			return params.GetExpiration.Run(ctx, keeper)
		case params.GetAllExpirations != nil:
			return params.GetAllExpirations.Run(ctx, keeper)
		case params.GetAllExpirationsByOwner != nil:
			return params.GetAllExpirationsByOwner.Run(ctx, keeper)
		default:
			return nil, fmt.Errorf("wasm: invalid %s query: %s", types.ModuleName, string(query))
		}
	}
}

// Run gets a scope by ID.
func (params *GetExpirationParams) Run(ctx sdk.Context, keeper keeper.Keeper) ([]byte, error) {
	expiration, err := keeper.GetExpiration(ctx, params.ModuleAssetID)
	if err != nil {
		return nil, fmt.Errorf("wasm: %s not found: %s", types.ModuleName, params.ModuleAssetID)
	}

	return createExpirationResponse(expiration)
}

// Run gets sessions by scope ID and name (optional)
func (params *GetAllExpirationsParams) Run(ctx sdk.Context, keeper keeper.Keeper) ([]byte, error) {
	var expirations []*types.Expiration

	// Callback func that adds expirations to expirations slice.
	expirationHandler := func(expiration types.Expiration) error {
		expirations = append(expirations, &expiration)
		return nil
	}

	if err := keeper.IterateExpirations(ctx, types.ModuleAssetKeyPrefix, expirationHandler); err != nil {
		return nil, fmt.Errorf("wasm: failed to retrieve %ss: %w", types.ModuleName, err)
	}

	return createExpirationsResponse(expirations)
}

// Run gets records by scope ID and name (optional)
func (params *GetAllExpirationsByOwnerParams) Run(ctx sdk.Context, keeper keeper.Keeper) ([]byte, error) {
	expirations, err := keeper.GetAllExpirationsByOwner(ctx, params.Owner)
	if err != nil {
		return nil, fmt.Errorf("wasm: failed to retrieve %ss by owner [%s]: %w", types.ModuleName, params.Owner, err)
	}

	return createExpirationsResponse(expirations)
}
