package types

import (
	wrapperspb "google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	"github.com/cosmos/gogoproto/proto"
)

// RegisterInterfaces registers concrete implementations for this module.
func RegisterInterfaces(registry types.InterfaceRegistry) {
	messages := make([]proto.Message, len(AllRequestMsgs))
	copy(messages, AllRequestMsgs)
	registry.RegisterImplementations((*sdk.Msg)(nil), messages...)
	registry.RegisterImplementations((*proto.Message)(nil), &wrapperspb.StringValue{})
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
