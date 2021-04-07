// Package provwasm allows CosmWasm smart contracts to communicate with custom provenance modules.
package provwasm

import (
	"encoding/json"
	"fmt"

	"github.com/CosmWasm/wasmd/x/wasm"
	cosmwasm "github.com/CosmWasm/wasmvm"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// This is just here to force the use of wasmvm dep so we can set the version in go.mod.
type Q cosmwasm.Querier

// Querier describes behavior for provenance smart contract query support.
type Querier func(ctx sdk.Context, query json.RawMessage, version string) ([]byte, error)

// QuerierRegistry maps routes to queriers.
type QuerierRegistry struct {
	queriers map[string]Querier
}

// NewQuerierRegistry creates a new registry for queriers.
func NewQuerierRegistry() *QuerierRegistry {
	return &QuerierRegistry{
		queriers: make(map[string]Querier),
	}
}

// RegisterQuerier adds a query handler for the given route.
func (qr *QuerierRegistry) RegisterQuerier(route string, querier Querier) {
	if _, exists := qr.queriers[route]; exists {
		panic(fmt.Sprintf("wasm: querier already registered for route: %s", route))
	}
	qr.queriers[route] = querier
}

// QueryPlugins provides provenance query support for smart contracts.
func QueryPlugins(registry *QuerierRegistry) *wasm.QueryPlugins {
	return &wasm.QueryPlugins{
		Custom: customPlugins(registry),
	}
}

// Custom provenance queriers for CosmWasm integration.
func customPlugins(registry *QuerierRegistry) wasm.CustomQuerier {
	return func(ctx sdk.Context, request json.RawMessage) ([]byte, error) {
		req := QueryRequest{}
		if err := json.Unmarshal(request, &req); err != nil {
			ctx.Logger().Error("failed to unmarshal query request", "err", err)
			return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
		}
		query, exists := registry.queriers[req.Route]
		if !exists {
			ctx.Logger().Error("querier not found", "route", req.Route)
			return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "querier not found for route: %s", req.Route)
		}
		bz, err := query(ctx, req.Params, req.Version)
		if err != nil {
			ctx.Logger().Error("failed to execute query", "err", err)
			return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
		}
		if !json.Valid(bz) {
			ctx.Logger().Error("invalid querier JSON", "route", req.Route)
			return nil, sdkerrors.Wrapf(sdkerrors.ErrJSONMarshal, "invalid querier JSON from route: %s", req.Route)
		}
		return bz, nil
	}
}
