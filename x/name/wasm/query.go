// Package wasm supports smart contract integration with the name module.
package wasm

import (
	"encoding/json"
	"fmt"

	"github.com/provenance-io/provenance/internal/provwasm"
	"github.com/provenance-io/provenance/x/name/keeper"
	"github.com/provenance-io/provenance/x/name/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NameQueryParams represents the request type for the name module sent by a smart contracts.
// Only one query field should be set.
type NameQueryParams struct {
	// Resolve the address bound to the given name.
	Resolve *ResolveQueryParams `json:"resolve,omitempty"`
	// Lookup all names an address is bound to.
	Lookup *LookupQueryParams `json:"lookup,omitempty"`
}

// ResolveQueryParams are the inputs for a resolve name query.
type ResolveQueryParams struct {
	// The name we want to resolve the address for.
	Name string `json:"name"`
}

// LookupQueryParams are the inpust for a lookup query.
type LookupQueryParams struct {
	// Find all names bound to this address.
	Address string `json:"address"`
}

// Querier returns a smart contract querier for the name module.
func Querier(keeper keeper.Keeper) provwasm.Querier {
	return func(ctx sdk.Context, query json.RawMessage, version string) ([]byte, error) {
		wrapper := struct {
			Params *NameQueryParams `json:"name"`
		}{}
		if err := json.Unmarshal(query, &wrapper); err != nil {
			return nil, fmt.Errorf("wasm: invalid query: %w", err)
		}
		params := wrapper.Params
		if params == nil {
			return nil, fmt.Errorf("wasm: nil name query params")
		}
		switch {
		case params.Resolve != nil:
			return params.Resolve.Run(ctx, keeper)
		case params.Lookup != nil:
			return params.Lookup.Run(ctx, keeper)
		default:
			return nil, fmt.Errorf("wasm: invalid name query: %s", string(query))
		}
	}
}

// Run resolves the address for a name.
func (params *ResolveQueryParams) Run(ctx sdk.Context, keeper keeper.Keeper) ([]byte, error) {
	record, err := keeper.GetRecordByName(ctx, params.Name)
	if err != nil {
		return nil, fmt.Errorf("wasm: resolve query failed: %w", err)
	}
	return createResponse(types.NameRecords{*record})
}

// Run looks up all names bound to a given address.
func (params *LookupQueryParams) Run(ctx sdk.Context, keeper keeper.Keeper) ([]byte, error) {
	acc, err := sdk.AccAddressFromBech32(params.Address)
	if err != nil {
		return nil, fmt.Errorf("wasm: invalid address: %w", err)
	}
	records, err := keeper.GetRecordsByAddress(ctx, acc)
	if err != nil {
		return nil, fmt.Errorf("wasm: lookup query failed: %w", err)
	}
	return createResponse(records)
}

// A helper function for converting name module record types into local query response types.
func createResponse(records types.NameRecords) ([]byte, error) {
	rep := &QueryResNames{}
	for _, r := range records {
		rep.Records = append(
			rep.Records,
			QueryResName{
				Name:       r.Name,
				Address:    r.Address,
				Restricted: r.Restricted,
			},
		)
	}
	bz, err := json.Marshal(rep)
	if err != nil {
		return nil, fmt.Errorf("wasm: marshal response failed: %w", err)
	}
	return bz, nil
}
