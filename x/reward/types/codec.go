package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ignoring RegisterLegacyAminoCodec registers all the necessary types and interfaces for the
// double check
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*RewardAction)(nil),
		&ActionTransfer{},
		&ActionDelegate{},
		&ActionVote{},
	)

	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgCreateRewardProgramRequest{},
		&MsgEndRewardProgramRequest{},
		&MsgClaimRewardsRequest{},
		&MsgClaimAllRewardsRequest{},
	)
}

var (
	// moving to protoCodec since this is a new module and should not use the
	// amino codec..someone to double verify
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)
