package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/gogoproto/proto"
)

// ignoring RegisterLegacyAminoCodec registers all the necessary types and interfaces for the
// double check
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	messages := make([]proto.Message, len(allRequestMsgs))
	copy(messages, allRequestMsgs)
	registry.RegisterImplementations((*sdk.Msg)(nil), messages...)

	registry.RegisterImplementations(
		(*govtypesv1beta1.Content)(nil),
		&AddMsgFeeProposal{},
		&UpdateMsgFeeProposal{},
		&RemoveMsgFeeProposal{},
		&UpdateNhashPerUsdMilProposal{},
		&UpdateConversionFeeDenomProposal{},
	)
}

var (
	// moving to protoCodec since this is a new module and should not use the
	// amino codec..someone to double verify
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)
