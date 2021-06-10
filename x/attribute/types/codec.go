package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterLegacyAminoCodec registers all the necessary types and interfaces for the
// account module.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgAddAttributeRequest{}, "provenance/attribute/MsgAddAttributeRequest", nil)
	cdc.RegisterConcrete(&MsgUpdateAttributeRequest{}, "provenance/attribute/MsgUpdateAttributeRequest", nil)
	cdc.RegisterConcrete(&MsgDeleteAttributeRequest{}, "provenance/attribute/MsgDeleteAttributeRequest", nil)
	cdc.RegisterConcrete(&MsgDeleteDistinctAttributeRequest{}, "provenance/attribute/MsgDeleteDistinctAttributeRequest", nil)
}

func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgAddAttributeRequest{},
		&MsgUpdateAttributeRequest{},
		&MsgDeleteAttributeRequest{},
		&MsgDeleteDistinctAttributeRequest{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	amino = codec.NewLegacyAmino()

	// ModuleCdc references the global x/attribute module codec. Note, the codec should
	// ONLY be used in certain instances of tests and for JSON encoding as Amino is
	// still used for that purpose.
	//
	// The actual codec used for serialization should be provided to x/attribute and
	// defined at the application level.
	ModuleCdc = codec.NewAminoCodec(amino)
)

func init() {
	RegisterLegacyAminoCodec(amino)
	cryptocodec.RegisterCrypto(amino)
	amino.Seal()
}
