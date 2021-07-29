package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

// RegisterLegacyAminoCodec registers all the necessary types and interfaces for the
// account module.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	// // Register base marker account and access grants
	// cdc.RegisterInterface((*AccessGrantI)(nil), nil)
	// cdc.RegisterInterface((*MarkerAccountI)(nil), nil)
	// cdc.RegisterConcrete(&MarkerAccount{}, "provenance/marker/Account", nil)
	// cdc.RegisterConcrete(&AccessGrant{}, "provenance/marker/AcccessGrant", nil)

	// // Register messages used to update and manage the markers
	// cdc.RegisterConcrete(&MsgAddMarkerRequest{}, "provenance/marker/MsgAddMarkerRequest", nil)
	// cdc.RegisterConcrete(&MsgAddAccessRequest{}, "provenance/marker/MsgAddAccessRequest", nil)
	// cdc.RegisterConcrete(&MsgDeleteAccessRequest{}, "provenance/marker/MsgDeleteAccessRequest", nil)
	// cdc.RegisterConcrete(&MsgFinalizeRequest{}, "provenance/marker/MsgFinalizeRequest", nil)
	// cdc.RegisterConcrete(&MsgActivateRequest{}, "provenance/marker/MsgActivateRequest", nil)
	// cdc.RegisterConcrete(&MsgCancelRequest{}, "provenance/marker/MsgCancelRequest", nil)
	// cdc.RegisterConcrete(&MsgDeleteRequest{}, "provenance/marker/MsgDeleteRequest", nil)
	// cdc.RegisterConcrete(&MsgMintRequest{}, "provenance/marker/MsgMintRequest", nil)
	// cdc.RegisterConcrete(&MsgBurnRequest{}, "provenance/marker/MsgBurnRequest", nil)
	// cdc.RegisterConcrete(&MsgWithdrawRequest{}, "provenance/marker/MsgWithdrawRequest", nil)
	// cdc.RegisterConcrete(&MsgTransferRequest{}, "provenance/marker/MsgTransferRequest", nil)

	// // Governance proposal types for marker management.
	// cdc.RegisterConcrete(&SupplyIncreaseProposal{}, "provenance/marker/SupplyIncreaseProposal", nil)
	// cdc.RegisterConcrete(&SupplyDecreaseProposal{}, "provenance/marker/SupplyDecreaseProposal", nil)
	// cdc.RegisterConcrete(&SetAdministratorProposal{}, "provenance/marker/SetAdministratorProposal", nil)
	// cdc.RegisterConcrete(&RemoveAdministratorProposal{}, "provenance/marker/RemoveAdministratorProposal", nil)
	// cdc.RegisterConcrete(&ChangeStatusProposal{}, "provenance/marker/ChangeStatusProposal", nil)

	// legacytx.RegisterLegacyAminoCodec(cdc)
}

// RegisterInterfaces registers implementations for the tx messages
func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgAddMarkerRequest{},
		&MsgAddAccessRequest{},
		&MsgDeleteAccessRequest{},
		&MsgFinalizeRequest{},
		&MsgActivateRequest{},
		&MsgCancelRequest{},
		&MsgDeleteRequest{},
		&MsgMintRequest{},
		&MsgBurnRequest{},
		&MsgWithdrawRequest{},
		&MsgTransferRequest{},
		&MsgSetDenomMetadataRequest{},
	)

	registry.RegisterImplementations(
		(*govtypes.Content)(nil),
		&AddMarkerProposal{},
		&SupplyIncreaseProposal{},
		&SupplyDecreaseProposal{},
		&SetAdministratorProposal{},
		&RemoveAdministratorProposal{},
		&ChangeStatusProposal{},
		&WithdrawEscrowProposal{},
		&SetDenomMetadataProposal{},
	)

	registry.RegisterImplementations(
		(*authz.Authorization)(nil),
		&MarkerTransferAuthorization{},
	)

	registry.RegisterInterface(
		"provenance.marker.v1.MarkerAccount",
		(*MarkerAccountI)(nil),
		&MarkerAccount{},
	)

	registry.RegisterInterface(
		"provenance.marker.v1.MarkerAccount",
		(*authtypes.AccountI)(nil),
		&MarkerAccount{},
	)

	registry.RegisterInterface(
		"provenance.marker.v1.MarkerAccount",
		(*authtypes.GenesisAccount)(nil),
		&MarkerAccount{},
	)

	registry.RegisterInterface(
		"provenance.marker.v1.AccessGrant",
		(*AccessGrantI)(nil),
		&AccessGrant{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	amino = codec.NewLegacyAmino()

	// ModuleCdc references the global x/account module codec. Note, the codec should
	// ONLY be used in certain instances of tests and for JSON encoding as Amino is
	// still used for that purpose.
	//
	// The actual codec used for serialization should be provided to x/account and
	// defined at the application level.
	ModuleCdc = codec.NewAminoCodec(amino)
)

func init() {
	RegisterLegacyAminoCodec(amino)
	cryptocodec.RegisterCrypto(amino)
	amino.Seal()
}
