package types

import (
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	"github.com/cosmos/gogoproto/proto"
)

// RegisterInterfaces registers concrete implementations for this module.
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	messages := make([]proto.Message, len(AllRequestMsgs))
	copy(messages, AllRequestMsgs)
	registry.RegisterImplementations((*sdk.Msg)(nil), messages...)

	registry.RegisterInterface(
		"provenance.trigger.v1.TriggerEventI",
		(*TriggerEventI)(nil),
		&TransactionEvent{},
		&BlockHeightEvent{},
		&BlockTimeEvent{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
