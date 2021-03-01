package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterLegacyAminoCodec registers concrete types on the Amino codec
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgMemorializeContractRequest{}, "provenance/metadata/MemorializeContractRequest", nil)
	cdc.RegisterConcrete(&MsgChangeOwnershipRequest{}, "provenance/metadata/ChangeOwnershipRequest", nil)
	cdc.RegisterConcrete(&MsgAddScopeRequest{}, "provenance/metadata/AddScopeRequest", nil)
	cdc.RegisterConcrete(&MsgRemoveScopeRequest{}, "provenance/metadata/RemoveScopeRequest", nil)
	cdc.RegisterConcrete(&MsgAddRecordGroupRequest{}, "provenance/metadata/AddRecordGroupRequest", nil)
	cdc.RegisterConcrete(&MsgAddRecordRequest{}, "provenance/metadata/AddRecordRequest", nil)
	cdc.RegisterConcrete(&MsgAddScopeSpecificationRequest{}, "provenance/metadata/AddScopeSpecificationRequest", nil)
	cdc.RegisterConcrete(&MsgDeleteScopeSpecificationRequest{}, "provenance/metadata/DeleteScopeSpecificationRequest", nil)
	cdc.RegisterConcrete(&MsgAddContractSpecificationRequest{}, "provenance/metadata/AddContractSpecificationRequest", nil)
	cdc.RegisterConcrete(&MsgDeleteContractSpecificationRequest{}, "provenance/metadata/DeleteContractSpecificationRequest", nil)
}

// RegisterInterfaces registers implementations for the tx messages
func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgMemorializeContractRequest{},
		&MsgChangeOwnershipRequest{},
		&MsgAddScopeRequest{},
		&MsgRemoveScopeRequest{},
		&MsgAddRecordGroupRequest{},
		&MsgAddRecordRequest{},
		&MsgAddScopeSpecificationRequest{},
		&MsgDeleteScopeSpecificationRequest{},
		&MsgAddContractSpecificationRequest{},
		&MsgDeleteContractSpecificationRequest{},
	)
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	amino = codec.NewLegacyAmino()

	// ModuleCdc references the global x/metadata module codec. Note, the codec should
	// ONLY be used in certain instances of tests and for JSON encoding as Amino is
	// still used for that purpose.
	//
	// The actual codec used for serialization should be provided to x/metadata and
	// defined at the application level.
	ModuleCdc = codec.NewAminoCodec(amino)
)

func init() {
	RegisterLegacyAminoCodec(amino)
	cryptocodec.RegisterCrypto(amino)
	amino.Seal()
}
