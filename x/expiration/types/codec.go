package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// ignoring RegisterLegacyAminoCodec since this is a new module

func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgAddExpirationRequest{},
		&MsgAddExpirationResponse{},
		&MsgExtendExpirationRequest{},
		&MsgExtendExpirationResponse{},
		&MsgDeleteExpirationRequest{},
		&MsgDeleteExpirationResponse{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	// ModuleCdc uses the new ProtoCodec and avoids the amino codec
	ModuleCdc = codec.NewProtoCodec(types.NewInterfaceRegistry())
)
