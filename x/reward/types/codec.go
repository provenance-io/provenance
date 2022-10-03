package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RegisterLegacyAminoCodec registers all the necessary types and interfaces for the
// reward module.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgCreateRewardProgramRequest{}, "provenance/reward/MsgCreateRewardProgramRequest", nil)
	cdc.RegisterConcrete(&MsgEndRewardProgramRequest{}, "provenance/reward/MsgEndRewardProgramRequest", nil)
}

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
