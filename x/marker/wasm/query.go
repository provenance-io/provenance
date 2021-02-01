// Package wasm supports smart contract integration with the provenance marker module.
package wasm

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/provenance-io/provenance/internal/provwasm"
	"github.com/provenance-io/provenance/x/marker/keeper"
	"github.com/provenance-io/provenance/x/marker/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MarkerQueryParams represent parameters used to query the marker module.
type MarkerQueryParams struct {
	// Get a marker by address.
	*GetMarkerByAddress `json:"get_marker_by_address,omitempty"`
	// Get a marker by denomination.
	*GetMarkerByDenom `json:"get_marker_by_denom,omitempty"`
}

// GetMarkerByAddress represent a query request to get a marker by address.
type GetMarkerByAddress struct {
	// The marker address
	Address string `json:"address,omitempty"`
}

// GetMarkerByDenom represent a query request to get a marker by denomination.
type GetMarkerByDenom struct {
	// The marker denomination
	Denom string `json:"denom,omitempty"`
}

// Querier returns a smart contract querier for the name module.
func Querier(keeper keeper.Keeper) provwasm.Querier {
	return func(ctx sdk.Context, query json.RawMessage, version string) ([]byte, error) {
		wrapper := struct {
			Params *MarkerQueryParams `json:"marker"`
		}{}
		if err := json.Unmarshal(query, &wrapper); err != nil {
			return nil, fmt.Errorf("wasm: invalid query: %w", err)
		}
		params := wrapper.Params
		if params == nil {
			return nil, fmt.Errorf("wasm: nil marker query params")
		}
		switch {
		case params.GetMarkerByAddress != nil:
			return params.GetMarkerByAddress.Run(ctx, keeper)
		case params.GetMarkerByDenom != nil:
			return params.GetMarkerByDenom.Run(ctx, keeper)
		default:
			return nil, fmt.Errorf("wasm: invalid marker query: %s", string(query))
		}
	}
}

// Run gets a marker by address or denomination.
func (params *GetMarkerByAddress) Run(ctx sdk.Context, keeper keeper.Keeper) ([]byte, error) {
	if strings.TrimSpace(params.Address) == "" {
		return nil, fmt.Errorf("wasm: marker address cannot be empty")
	}
	address, err := sdk.AccAddressFromBech32(params.Address)
	if err != nil {
		return nil, fmt.Errorf("wasm: address is invalid: %w", err)
	}
	marker, err := keeper.GetMarker(ctx, address)
	if err != nil {
		return nil, fmt.Errorf("wasm: no marker found for address '%s': %w", params.Address, err)
	}
	markerAccount, ok := marker.(*types.MarkerAccount)
	if !ok {
		return nil, fmt.Errorf("wasm: unable to type-cast marker account")
	}
	balance := keeper.GetEscrow(ctx, marker)
	bz, err := json.Marshal(createResponseType(markerAccount, balance))
	if err != nil {
		return nil, fmt.Errorf("wasm: marshal marker query response failed: %w", err)
	}
	return bz, nil
}

// Run gets a marker by address or denomination.
func (params *GetMarkerByDenom) Run(ctx sdk.Context, keeper keeper.Keeper) ([]byte, error) {
	if strings.TrimSpace(params.Denom) == "" {
		return nil, fmt.Errorf("wasm: marker denomination cannot be empty")
	}
	marker, err := keeper.GetMarkerByDenom(ctx, params.Denom)
	if err != nil {
		return nil, fmt.Errorf("wasm: no marker found for denomination '%s': %w", params.Denom, err)
	}
	markerAccount, ok := marker.(*types.MarkerAccount)
	if !ok {
		return nil, fmt.Errorf("wasm: unable to type-cast marker account")
	}
	balance := keeper.GetEscrow(ctx, marker)
	bz, err := json.Marshal(createResponseType(markerAccount, balance))
	if err != nil {
		return nil, fmt.Errorf("wasm: marshal marker query response failed: %w", err)
	}
	return bz, nil
}
