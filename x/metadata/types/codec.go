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
	for i, msg := range AllRequestMsgs {
		messages[i] = msg
	}
	registry.RegisterImplementations((*sdk.Msg)(nil), messages...)

	registry.RegisterImplementations((*sdk.Msg)(nil),
		(*MsgWriteP8EContractSpecRequest)(nil),
		(*MsgP8EMemorializeContractRequest)(nil),
	)
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
