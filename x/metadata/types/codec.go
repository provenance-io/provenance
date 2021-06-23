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
	cdc.RegisterConcrete(&MsgWriteScopeRequest{}, "provenance/metadata/WriteScopeRequest", nil)
	cdc.RegisterConcrete(&MsgDeleteScopeRequest{}, "provenance/metadata/DeleteScopeRequest", nil)
	cdc.RegisterConcrete(&MsgAddScopeDataAccessRequest{}, "provenance/metadata/AddScopeDataAccessRequest", nil)
	cdc.RegisterConcrete(&MsgDeleteScopeDataAccessRequest{}, "provenance/metadata/DeleteScopeDataAccessRequest", nil)
	cdc.RegisterConcrete(&MsgAddScopeOwnerRequest{}, "provenance/metadata/AddScopeOwnerRequest", nil)
	cdc.RegisterConcrete(&MsgDeleteScopeOwnerRequest{}, "provenance/metadata/DeleteScopeOwnerRequest", nil)

	cdc.RegisterConcrete(&MsgWriteSessionRequest{}, "provenance/metadata/WriteSessionRequest", nil)
	cdc.RegisterConcrete(&MsgWriteRecordRequest{}, "provenance/metadata/WriteRecordRequest", nil)
	cdc.RegisterConcrete(&MsgDeleteRecordRequest{}, "provenance/metadata/DeleteRecordRequest", nil)

	cdc.RegisterConcrete(&MsgWriteScopeSpecificationRequest{}, "provenance/metadata/WriteScopeSpecificationRequest", nil)
	cdc.RegisterConcrete(&MsgDeleteScopeSpecificationRequest{}, "provenance/metadata/DeleteScopeSpecificationRequest", nil)
	cdc.RegisterConcrete(&MsgWriteContractSpecificationRequest{}, "provenance/metadata/WriteContractSpecificationRequest", nil)
	cdc.RegisterConcrete(&MsgDeleteContractSpecificationRequest{}, "provenance/metadata/DeleteContractSpecificationRequest", nil)
	cdc.RegisterConcrete(&MsgAddContractSpecToScopeSpecRequest{}, "provenance/metadata/AddContractSpecToScopeSpecRequest", nil)
	cdc.RegisterConcrete(&MsgDeleteContractSpecFromScopeSpecRequest{}, "provenance/metadata/DeleteContractSpecFromScopeSpecRequest", nil)
	cdc.RegisterConcrete(&MsgWriteRecordSpecificationRequest{}, "provenance/metadata/WriteRecordSpecificationRequest", nil)
	cdc.RegisterConcrete(&MsgDeleteRecordSpecificationRequest{}, "provenance/metadata/DeleteRecordSpecificationRequest", nil)

	cdc.RegisterConcrete(&MsgWriteP8EContractSpecRequest{}, "provenance/metadata/WriteP8EContractSpecRequest", nil)
	cdc.RegisterConcrete(&MsgP8EMemorializeContractRequest{}, "provenance/metadata/P8EMemorializeContractRequest", nil)

	cdc.RegisterConcrete(&MsgBindOSLocatorRequest{}, "provenance/metadata/BindOSLocatorRequest", nil)
	cdc.RegisterConcrete(&MsgModifyOSLocatorRequest{}, "provenance/metadata/ModifyOSLocatorRequest", nil)
	cdc.RegisterConcrete(&MsgDeleteOSLocatorRequest{}, "provenance/metadata/DeleteOSLocatorRequest", nil)
}

// RegisterInterfaces registers implementations for the tx messages
func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgWriteScopeRequest{},
		&MsgDeleteScopeRequest{},
		&MsgAddScopeDataAccessRequest{},
		&MsgDeleteScopeDataAccessRequest{},
		&MsgAddScopeOwnerRequest{},
		&MsgDeleteScopeOwnerRequest{},
		&MsgWriteSessionRequest{},
		&MsgWriteRecordRequest{},
		&MsgDeleteRecordRequest{},

		&MsgWriteScopeSpecificationRequest{},
		&MsgDeleteScopeSpecificationRequest{},
		&MsgWriteContractSpecificationRequest{},
		&MsgDeleteContractSpecificationRequest{},
		&MsgAddContractSpecToScopeSpecRequest{},
		&MsgDeleteContractSpecFromScopeSpecRequest{},
		&MsgWriteRecordSpecificationRequest{},
		&MsgDeleteRecordSpecificationRequest{},

		&MsgWriteP8EContractSpecRequest{},
		&MsgP8EMemorializeContractRequest{},

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
