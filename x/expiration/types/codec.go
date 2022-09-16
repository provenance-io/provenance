package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"

	metadatatypes "github.com/provenance-io/provenance/x/metadata/types"
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

// whiteListedInterfaceRegistry whitelist of sdk.Msg types
// invoked when a module asset expiration has expired
func whiteListedInterfaceRegistry() types.InterfaceRegistry {
	ir := types.NewInterfaceRegistry()
	ir.RegisterImplementations((*sdk.Msg)(nil),
		&metadatatypes.MsgDeleteScopeRequest{},
		&metadatatypes.MsgDeleteRecordRequest{},
		// todo: what other messages do we need to whitelist?
	)
	return ir
}

var (
	// ModuleCdc references the global x/expiration module codec. Note, the codec should
	// ONLY be used in certain instances of tests and basic validation of `Expiration.Message`
	// during a tx broadcast.
	//
	// The actual codec used for serialization should be provided to x/expiration and
	// defined at the application level.
	ModuleCdc = codec.NewProtoCodec(whiteListedInterfaceRegistry())
)
