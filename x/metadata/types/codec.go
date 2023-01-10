package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

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
	cryptocodec.RegisterCrypto(amino)
	amino.Seal()
}
