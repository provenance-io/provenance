package types

import (
	"github.com/gogo/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

// RegisterInterfaces registers implementations for the tx messages
func RegisterInterfaces(registry types.InterfaceRegistry) {
	messages := make([]proto.Message, len(allRequestMsgs))
	for i, msg := range allRequestMsgs {
		messages[i] = msg
	}
	registry.RegisterImplementations((*sdk.Msg)(nil), messages...)

	registry.RegisterImplementations(
		(*govtypesv1beta1.Content)(nil),
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
	// ModuleCdc references the global x/account module codec. Note, the codec should
	// ONLY be used in certain instances of tests and for JSON encoding as Amino is
	// still used for that purpose.
	//
	// The actual codec used for serialization should be provided to x/account and
	// defined at the application level.
	ModuleCdc = codec.NewProtoCodec(types.NewInterfaceRegistry())
)
