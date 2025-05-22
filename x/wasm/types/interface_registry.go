package types

import (
	wasmv1 "github.com/CosmWasm/wasmd/x/wasm/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	v1beta1 "github.com/provenance-io/provenance/x/wasm"
)

var _ codectypes.InterfaceRegistry = &WasmInterfaceRegistry{}

// NewWasmInterfaceRegistry returns a new WasmInterfaceRegistry instance
func NewWasmInterfaceRegistry(registry codectypes.InterfaceRegistry) WasmInterfaceRegistry {
	registry.RegisterImplementations((*sdk.Msg)(nil), &v1beta1.MsgExecuteContract{}, &wasmv1.MsgExecuteContract{})
	return WasmInterfaceRegistry{registry}
}

// WasmInterfaceRegistry represents an interface registry with a custom any resolver
type WasmInterfaceRegistry struct {
	codectypes.InterfaceRegistry
}

// Resolve implements codectypes.InterfaceRegistry
func (WasmInterfaceRegistry) Resolve(_ string) (proto.Message, error) {
	return new(WasmAny), nil
}
