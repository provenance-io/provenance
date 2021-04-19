package types

import (
	types "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	"github.com/provenance-io/provenance/x/authz/exported"
)

// RegisterInterfaces registers the interfaces types with the interface registry
func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.MsgRequest)(nil),
		&MsgGrantAuthorizationRequest{},
		&MsgRevokeAuthorizationRequest{},
		&MsgExecAuthorizedRequest{},
	)

	registry.RegisterInterface(
		"cosmos.authz.v1beta1.Authorization",
		(*exported.Authorization)(nil),
		&GenericAuthorization{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
