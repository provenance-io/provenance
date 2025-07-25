package types

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"

	v1beta1 "github.com/provenance-io/provenance/legacy_protos/cosmwasm/wasm/v1beta1"
)

var _ codectypes.InterfaceRegistry = &WasmInterfaceRegistry{}

// NewWasmInterfaceRegistry returns a new WasmInterfaceRegistry instance
func NewWasmInterfaceRegistry(registry codectypes.InterfaceRegistry) WasmInterfaceRegistry {
	registry.RegisterImplementations((*sdk.Msg)(nil), &v1beta1.MsgExecuteContract{})
	return WasmInterfaceRegistry{registry}
}

// WasmInterfaceRegistry represents an interface registry with a custom any resolver
type WasmInterfaceRegistry struct {
	codectypes.InterfaceRegistry
}

// Resolve implements codectypes.InterfaceRegistry
func (wir WasmInterfaceRegistry) Resolve(typeURL string) (proto.Message, error) {
	msg, err := wir.InterfaceRegistry.Resolve(typeURL)
	if err == nil {
		return msg, nil
	}
	if typeURL == "/cosmwasm.wasm.v1beta1.MsgExecuteContract" {
		return &v1beta1.MsgExecuteContract{}, nil
	}
	return new(WasmAny), nil
}
