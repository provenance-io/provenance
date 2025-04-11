package types

import (
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/gogoproto/proto"
)

// TODO[fees]: Add a call to RegisterInterfaces from app.New.

// RegisterInterfaces registers concrete implementations for this module.
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	messages := make([]proto.Message, len(AllRequestMsgs))
	copy(messages, AllRequestMsgs)
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
