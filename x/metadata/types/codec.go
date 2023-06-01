package types

import (
	"github.com/gogo/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterInterfaces registers implementations for the tx messages
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	messages := make([]proto.Message, len(allRequestMsgs))
	for i, msg := range allRequestMsgs {
		messages[i] = msg
	}
	registry.RegisterImplementations((*sdk.Msg)(nil), messages...)

	registry.RegisterImplementations((*sdk.Msg)(nil),
		(*MsgWriteP8EContractSpecRequest)(nil),
		(*MsgP8EMemorializeContractRequest)(nil),
	)
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	// ModuleCdc references the global x/metadata module codec. Note, the codec should
	// ONLY be used in certain instances of tests and for JSON encoding as Amino is
	// still used for that purpose.
	//
	// The actual codec used for serialization should be provided to x/metadata and
	// defined at the application level.
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)
