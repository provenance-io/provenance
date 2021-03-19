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
	cdc.RegisterConcrete(&MsgDeleteScopeRequest{}, "provenance/metadata/DeleteScopeRequest", nil)
	cdc.RegisterConcrete(&MsgAddSessionRequest{}, "provenance/metadata/AddSessionRequest", nil)
	cdc.RegisterConcrete(&MsgAddRecordRequest{}, "provenance/metadata/AddRecordRequest", nil)
	cdc.RegisterConcrete(&MsgAddScopeSpecificationRequest{}, "provenance/metadata/AddScopeSpecificationRequest", nil)
	cdc.RegisterConcrete(&MsgDeleteScopeSpecificationRequest{}, "provenance/metadata/DeleteScopeSpecificationRequest", nil)
	cdc.RegisterConcrete(&MsgAddContractSpecificationRequest{}, "provenance/metadata/AddContractSpecificationRequest", nil)
	cdc.RegisterConcrete(&MsgDeleteContractSpecificationRequest{}, "provenance/metadata/DeleteContractSpecificationRequest", nil)
	cdc.RegisterConcrete(&MsgAddRecordSpecificationRequest{}, "provenance/metadata/AddRecordSpecificationRequest", nil)
	cdc.RegisterConcrete(&MsgDeleteRecordSpecificationRequest{}, "provenance/metadata/DeleteRecordSpecificationRequest", nil)
	cdc.RegisterConcrete(&MsgAddP8EContractSpecRequest{}, "provenance/metadata/AddP8EContractSpecRequest", nil)
	cdc.RegisterConcrete(&MsgBindOSLocatorRequest{}, "provenance/metadata/MsgBindOSLocatorRequest", nil)
	cdc.RegisterConcrete(&MsgModifyOSLocatorRequest{}, "provenance/metadata/MsgModifyOSLocatorRequest", nil)
	cdc.RegisterConcrete(&MsgDeleteOSLocatorRequest{}, "provenance/metadata/MsgDeleteOSLocatorRequest", nil)
}

// RegisterInterfaces registers implementations for the tx messages
func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgMemorializeContractRequest{},
		&MsgChangeOwnershipRequest{},
		&MsgAddScopeRequest{},
		&MsgDeleteScopeRequest{},
		&MsgAddSessionRequest{},
		&MsgAddRecordRequest{},
		&MsgAddScopeSpecificationRequest{},
		&MsgDeleteScopeSpecificationRequest{},
		&MsgAddContractSpecificationRequest{},
		&MsgDeleteContractSpecificationRequest{},
		&MsgAddRecordSpecificationRequest{},
		&MsgDeleteRecordSpecificationRequest{},
		&MsgAddP8EContractSpecRequest{},
		&MsgBindOSLocatorRequest{},
		&MsgModifyOSLocatorRequest{},
		&MsgDeleteOSLocatorRequest{},
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
