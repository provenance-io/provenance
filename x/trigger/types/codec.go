package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ignoring RegisterLegacyAminoCodec registers all the necessary types and interfaces for the
// double check
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgCreateTriggerRequest{},
		&MsgDestroyTriggerRequest{},
	)

	registry.RegisterInterface(
		"provenance.trigger.v1.TransactionEvent",
		(*TriggerEventI)(nil),
		&TransactionEvent{},
	)

	registry.RegisterInterface(
		"provenance.trigger.v1.BlockHeightEvent",
		(*TriggerEventI)(nil),
		&BlockHeightEvent{},
	)

	registry.RegisterInterface(
		"provenance.trigger.v1.BlockTimeEvent",
		(*TriggerEventI)(nil),
		&BlockTimeEvent{},
	)
}

var (
	// moving to protoCodec since this is a new module and should not use the
	// amino codec..someone to double verify
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)
