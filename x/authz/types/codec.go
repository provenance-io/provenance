package types

import (
	types "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	"github.com/provenance-io/provenance/x/authz/exported"
	marker "github.com/provenance-io/provenance/x/marker/types"

)

// RegisterInterfaces registers the interfaces types with the interface registry
func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.MsgRequest)(nil),
		&MsgGrantAuthorizationRequest{},
		&MsgRevokeAuthorizationRequest{},
		&MsgExecAuthorizedRequest{},
	)

	registry.RegisterInterface(
		"provenance.authz.v1.Authorization",
		(*exported.Authorization)(nil),
		&marker.MarkerSendAuthorization{},
		&GenericAuthorization{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
