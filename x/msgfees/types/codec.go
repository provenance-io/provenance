package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

// ignoring RegisterLegacyAminoCodec registers all the necessary types and interfaces for the
// double check

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*govtypes.Content)(nil),
		&AddMsgFeeProposal{},
		&UpdateMsgFeeProposal{},
		&RemoveMsgFeeProposal{},
		&UpdateNhashPerUsdMilProposal{},
	)
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgAssessCustomMsgFeeRequest{},
	)
}

var (
	// moving to protoCodec since this is a new module and should not use the
	// amino codec..someone to double verify
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)
