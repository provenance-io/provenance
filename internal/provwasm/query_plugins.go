// Package provwasm allows CosmWasm smart contracts to communicate with custom provenance modules.
package provwasm

import (
	"encoding/json"
	"fmt"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmvmtypes "github.com/CosmWasm/wasmvm/v2/types"

	abci "github.com/cometbft/cometbft/abci/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"

	provwasmtypes "github.com/provenance-io/provenance/x/wasm/types"
)

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
func QueryPlugins(queryRouter baseapp.GRPCQueryRouter, cdc codec.Codec) *wasmkeeper.QueryPlugins {
	protoCdc, ok := cdc.(*codec.ProtoCodec)
	if !ok {
		panic(fmt.Errorf("codec must be *codec.ProtoCodec type: actual: %T", cdc))
	}

	stargateCdc := codec.NewProtoCodec(provwasmtypes.NewWasmInterfaceRegistry(protoCdc.InterfaceRegistry()))

	return &wasmkeeper.QueryPlugins{
		Stargate: StargateQuerier(queryRouter, stargateCdc),
		Grpc:     GrpcQuerier(queryRouter),
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

		res, err := route(ctx, &abci.RequestQuery{
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

// GrpcQuerier dispatches whitelisted queries and returns protobuf encoded responses
func GrpcQuerier(queryRouter baseapp.GRPCQueryRouter) func(ctx sdk.Context, request *wasmvmtypes.GrpcQuery) (proto.Message, error) {
	return func(ctx sdk.Context, request *wasmvmtypes.GrpcQuery) (proto.Message, error) {
		_, err := GetWhitelistedQuery(request.Path)
		if err != nil {
			return nil, err
		}

		route := queryRouter.Route(request.Path)
		if route == nil {
			return nil, wasmvmtypes.UnsupportedRequest{Kind: fmt.Sprintf("No route to query '%s'", request.Path)}
		}

		res, err := route(ctx, &abci.RequestQuery{
			Data: request.Data,
			Path: request.Path,
		})
		if err != nil {
			return nil, err
		}

		return res, nil
	}
}

// ConvertProtoToJsonMarshal  unmarshals the given bytes into a proto message and then marshals it to json.
// This is done so that clients calling stargate queries do not need to define their own proto unmarshalers,
// being able to use response directly by json marshaling, which is supported in cosmwasm.
func ConvertProtoToJSONMarshal(protoResponseType proto.Message, bz []byte, cdc codec.Codec) ([]byte, error) {
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
