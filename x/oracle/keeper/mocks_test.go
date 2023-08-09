package keeper

import (
	"context"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
)

// This file is available only to unit tests and exposes private things
// so that they can be used in unit tests.

type MockWasmServer struct {
}

func (k *Keeper) SetWasmQueryServer(server wasmtypes.QueryServer) {
	k.wasmQueryServer = server
}

func (m MockWasmServer) ContractInfo(context.Context, *wasmtypes.QueryContractInfoRequest) (*wasmtypes.QueryContractInfoResponse, error) {
	return nil, nil
}

func (m MockWasmServer) ContractHistory(context.Context, *wasmtypes.QueryContractHistoryRequest) (*wasmtypes.QueryContractHistoryResponse, error) {
	return nil, nil
}

func (m MockWasmServer) ContractsByCode(context.Context, *wasmtypes.QueryContractsByCodeRequest) (*wasmtypes.QueryContractsByCodeResponse, error) {
	return nil, nil
}

func (m MockWasmServer) AllContractState(context.Context, *wasmtypes.QueryAllContractStateRequest) (*wasmtypes.QueryAllContractStateResponse, error) {
	return nil, nil
}

func (m MockWasmServer) RawContractState(context.Context, *wasmtypes.QueryRawContractStateRequest) (*wasmtypes.QueryRawContractStateResponse, error) {
	return nil, nil
}

func (m MockWasmServer) SmartContractState(context.Context, *wasmtypes.QuerySmartContractStateRequest) (*wasmtypes.QuerySmartContractStateResponse, error) {
	return &wasmtypes.QuerySmartContractStateResponse{
		Data: []byte("{}"),
	}, nil
}

func (m MockWasmServer) Code(context.Context, *wasmtypes.QueryCodeRequest) (*wasmtypes.QueryCodeResponse, error) {
	return nil, nil
}

func (m MockWasmServer) Codes(context.Context, *wasmtypes.QueryCodesRequest) (*wasmtypes.QueryCodesResponse, error) {
	return nil, nil
}

func (m MockWasmServer) PinnedCodes(context.Context, *wasmtypes.QueryPinnedCodesRequest) (*wasmtypes.QueryPinnedCodesResponse, error) {
	return nil, nil
}

func (m MockWasmServer) Params(context.Context, *wasmtypes.QueryParamsRequest) (*wasmtypes.QueryParamsResponse, error) {
	return nil, nil
}

func (m MockWasmServer) ContractsByCreator(context.Context, *wasmtypes.QueryContractsByCreatorRequest) (*wasmtypes.QueryContractsByCreatorResponse, error) {
	return nil, nil
}
