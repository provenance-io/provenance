package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
)

// RegisterInterfaces registers concrete implementations for this module.
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	messages := make([]proto.Message, len(AllRequestMsgs))
	copy(messages, AllRequestMsgs)
	registry.RegisterImplementations((*sdk.Msg)(nil), messages...)

	registry.RegisterImplementations(
		(*TriggerEventI)(nil),
		&TransactionEvent{},
		&BlockHeightEvent{},
		&BlockTimeEvent{},
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
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)
