package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/gogo/protobuf/proto"
)

// ignoring RegisterLegacyAminoCodec registers all the necessary types and interfaces for the
// double check
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgUpdateOracleRequest{},
		&MsgQueryOracleRequest{},
	)

	registry.RegisterImplementations((*proto.Message)(nil),
		&banktypes.QueryAllBalancesRequest{},
		&banktypes.QueryAllBalancesResponse{},
	)
}

var (
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)
