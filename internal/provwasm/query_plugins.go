// Package provwasm allows CosmWasm smart contracts to communicate with custom provenance modules.
package provwasm

import (
	"encoding/json"
	"fmt"


	"github.com/CosmWasm/wasmd/x/wasm"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// The maximum querier result size allowed, ~10MB.
const maxQueryResultSize = (10 << 20) - 1

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
func QueryPlugins(registry *QuerierRegistry, queryRouter baseapp.GRPCQueryRouter, codec codec.Codec) *wasm.QueryPlugins {
	return &wasm.QueryPlugins{
		Custom:   customPlugins(registry),
		Stargate: StargateQuerier(queryRouter, codec),
	}
}

// Custom provenance queriers for CosmWasm integration.
func customPlugins(registry *QuerierRegistry) wasm.CustomQuerier {
	return func(ctx sdk.Context, request json.RawMessage) ([]byte, error) {
		req := QueryRequest{}
		if err := json.Unmarshal(request, &req); err != nil {
			ctx.Logger().Error("failed to unmarshal query request", "err", err)
			return nil, sdkerrors.ErrJSONUnmarshal.Wrap(err.Error())
		}
		query, exists := registry.queriers[req.Route]
		if !exists {
			ctx.Logger().Error("querier not found", "route", req.Route)
			return nil, sdkerrors.ErrInvalidRequest.Wrapf("querier not found for route: %s", req.Route)
		}
		bz, err := query(ctx, req.Params, req.Version)
		if err != nil {
			ctx.Logger().Error("failed to execute query", "err", err)
			return nil, sdkerrors.ErrInvalidRequest.Wrap(err.Error())
		}
		if len(bz) > maxQueryResultSize {
			errm := "query result size limit exceeded"
			ctx.Logger().Error(errm, "maxQueryResultSize", maxQueryResultSize)
			return nil, sdkerrors.ErrInvalidRequest.Wrap(errm)
		}
		if !json.Valid(bz) {
			ctx.Logger().Error("invalid querier JSON", "route", req.Route)
			return nil, sdkerrors.ErrJSONMarshal.Wrapf("invalid querier JSON from route: %s", req.Route)
		}
		return bz, nil
	}
}

// StargateQuerier dispatches whitelisted stargate queries
func StargateQuerier(queryRouter baseapp.GRPCQueryRouter, cdc codec.Codec) func(ctx sdk.Context, request *wasmvmtypes.StargateQuery) ([]byte, error) {
	return func(ctx sdk.Context, request *wasmvmtypes.StargateQuery) ([]byte, error) {
		protoResponseType, err := GetWhitelistedQuery(request.Path)
		if err != nil {
			return nil, err
		}

		route := queryRouter.Route(request.Path)
		if route == nil {
			return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("No route to query '%s'", request.Path)}
		}

		res, err := route(ctx, abci.RequestQuery{
			Data: request.Data,
			Path: request.Path,
		})
		if err != nil {
			return nil, err
		}

		bz, err := ConvertProtoToJSONMarshal(protoResponseType, res.Value, cdc)
		if err != nil {
			return nil, err
		}

		return bz, nil
	}
}

// ConvertProtoToJsonMarshal  unmarshals the given bytes into a proto message and then marshals it to json.
// This is done so that clients calling stargate queries do not need to define their own proto unmarshalers,
// being able to use response directly by json marshaling, which is supported in cosmwasm.
func ConvertProtoToJSONMarshal(protoResponseType codec.ProtoMarshaler, bz []byte, cdc codec.Codec) ([]byte, error) {
	// unmarshal binary into stargate response data structure
	err := cdc.Unmarshal(bz, protoResponseType)
	if err != nil {
		return nil, wasmvmtypes.Unknown{}
	}

	bz, err = cdc.MarshalJSON(protoResponseType)
	if err != nil {
		return nil, wasmvmtypes.Unknown{}
	}

	protoResponseType.Reset()

	return bz, nil
}
