package keeper

import (
	"context"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	"github.com/provenance-io/provenance/x/oracle/types"
)

// This file is available only to unit tests and exposes private things
// so that they can be used in unit tests.

type MockWasmServer struct {
}

func (k Keeper) WithWasmQueryServer(server wasmtypes.QueryServer) Keeper {
	k.wasmQueryServer = server
	return k
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

type MockICS4Wrapper struct {
	counter uint64
}

func (k Keeper) WithMockICS4Wrapper(ics4wrapper types.ICS4Wrapper) Keeper {
	k.ics4Wrapper = ics4wrapper
	return k
}

func (k MockICS4Wrapper) SendPacket(
	ctx sdk.Context,
	chanCap *capabilitytypes.Capability,
	sourcePort string,
	sourceChannel string,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
	data []byte,
) (sequence uint64, err error) {
	k.counter += 1
	return k.counter, nil
}

type MockChannelKeeper struct {
	counter uint64
}

func (k Keeper) WithMockChannelKeeper(channelKeeper types.ChannelKeeper) Keeper {
	k.channelKeeper = channelKeeper
	return k
}

func (m *MockChannelKeeper) GetChannel(ctx sdk.Context, portID, channelID string) (channeltypes.Channel, bool) {
	return channeltypes.Channel{}, true
}

func (m *MockChannelKeeper) GetNextSequenceSend(ctx sdk.Context, portID, channelID string) (uint64, bool) {
	m.counter += 1
	return m.counter, true
}

type MockScopedKeeper struct {
}

func (k Keeper) WithMockScopedKeeper(scopedKeeper types.ScopedKeeper) Keeper {
	k.scopedKeeper = scopedKeeper
	return k
}

func (m MockScopedKeeper) GetCapability(ctx sdk.Context, name string) (*capabilitytypes.Capability, bool) {
	return &capabilitytypes.Capability{}, true
}

func (m MockScopedKeeper) AuthenticateCapability(ctx sdk.Context, cap *capabilitytypes.Capability, name string) bool {
	return true
}

func (m MockScopedKeeper) ClaimCapability(ctx sdk.Context, cap *capabilitytypes.Capability, name string) error {
	return nil
}
