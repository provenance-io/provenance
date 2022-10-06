package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterInterfaces registers the module 'Msg' types with the registry
func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgAddExpirationRequest{},
		&MsgExtendExpirationRequest{},
		&MsgInvokeExpirationRequest{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	// ModuleCdc uses the new proto codec.
	// The expiration module is a new module and does not require the legacy amino codec.
	ModuleCdc = codec.NewProtoCodec(types.NewInterfaceRegistry())
)
