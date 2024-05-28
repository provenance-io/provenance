package keeper

import (
	"context"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/ibchooks/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

func (m msgServer) EmitIBCAck(goCtx context.Context, msg *types.MsgEmitIBCAck) (*types.MsgEmitIBCAckResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.MsgEmitAckKey,
			sdk.NewAttribute(types.AttributeSender, msg.Sender),
			sdk.NewAttribute(types.AttributeChannel, msg.Channel),
			sdk.NewAttribute(types.AttributePacketSequence, strconv.FormatUint(msg.PacketSequence, 10)),
		),
	)

	ack, err := m.Keeper.EmitIBCAck(ctx, msg.Sender, msg.Channel, msg.PacketSequence)
	if err != nil {
		return nil, err
	}

	return &types.MsgEmitIBCAckResponse{ContractResult: string(ack), IbcAck: string(ack)}, nil
}

// UpdateParams is a governance proposal endpoint for updating the ibchooks module's params.
func (m msgServer) UpdateParams(goCtx context.Context, msg *types.MsgUpdateParamsRequest) (*types.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := m.ValidateAuthority(msg.Authority); err != nil {
		return nil, err
	}

	m.SetParams(ctx, msg.Params)
	if err := ctx.EventManager().EmitTypedEvent(&types.EventIBCHooksParamsUpdated{
		AllowedAsyncAckContracts: msg.Params.AllowedAsyncAckContracts,
	}); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}
