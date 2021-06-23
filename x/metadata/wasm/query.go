// Package wasm supports smart contract integration with the metadata module.
package wasm

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/provenance-io/provenance/internal/provwasm"
	"github.com/provenance-io/provenance/x/metadata/keeper"
	"github.com/provenance-io/provenance/x/metadata/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MetadataQueryParams represents the query request type for the metadata module sent by smart contracts.
// Only one query field should be set.
type MetadataQueryParams struct {
	// Get a scope by ID.
	GetScope *GetScopeParams `json:"get_scope,omitempty"`
	// Get sessions by scope ID and name (optional).
	GetSessions *GetSessionsParams `json:"get_sessions,omitempty"`
	// Get records by scope ID and name (optional).
	GetRecords *GetRecordsParams `json:"get_records,omitempty"`
}

// GetScopeParams are the inputs for a scope query.
type GetScopeParams struct {
	// The bech32 address of the scope we want to get.
	ScopeID string `json:"scope_id"`
}

// GetSessionsParams are the inputs for a session query.
type GetSessionsParams struct {
	// The bech32 address of the scope we want to get sessions for.
	ScopeID string `json:"scope_id"`
}

// GetRecordsParams are the inputs for a records query.
type GetRecordsParams struct {
	// The bech32 address of the scope we want to get records for.
	ScopeID string `json:"scope_id"`
	// The optional record name.
	Name string `json:"name,omitempty"`
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
		case params.GetScope != nil:
			return params.GetScope.Run(ctx, keeper)
		case params.GetSessions != nil:
			return params.GetSessions.Run(ctx, keeper)
		case params.GetRecords != nil:
			return params.GetRecords.Run(ctx, keeper)
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

// Run gets sessions by scope ID and name (optional)
func (params *GetSessionsParams) Run(ctx sdk.Context, keeper keeper.Keeper) ([]byte, error) {
	scopeID, err := types.MetadataAddressFromBech32(params.ScopeID)
	if err != nil {
		return nil, fmt.Errorf("wasm: invalid scope ID: %w", err)
	}
	var sessions []types.Session
	err = keeper.IterateSessions(ctx, scopeID, func(s types.Session) bool {
		sessions = append(sessions, s)
		return false
	})
	if err != nil {
		return nil, fmt.Errorf("wasm: %w", err)
	}
	return createSessionsResponse(sessions)
}

// Run gets records by scope ID and name (optional)
func (params *GetRecordsParams) Run(ctx sdk.Context, keeper keeper.Keeper) ([]byte, error) {
	scopeID, err := types.MetadataAddressFromBech32(params.ScopeID)
	if err != nil {
		return nil, fmt.Errorf("wasm: invalid scope ID: %w", err)
	}
	records, err := keeper.GetRecords(ctx, scopeID, strings.TrimSpace(params.Name))
	if err != nil {
		return nil, fmt.Errorf("wasm: unable to get scope records: %w", err)
	}
	return createRecordsResponse(records)
}
