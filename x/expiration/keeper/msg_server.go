package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/expiration/types"
)

type msgServer struct {
	Keeper
}

func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

func (m msgServer) AddExpiration(
	goCtx context.Context,
	msg *types.MsgAddExpirationRequest,
) (*types.MsgAddExpirationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// validate message
	if err := m.ValidateSetExpiration(ctx, msg.Expiration, msg.Signers, msg.MsgTypeURL()); err != nil {
		return nil, err
	}

	// add expiration
	if err := m.Keeper.SetExpiration(ctx, msg.Expiration); err != nil {
		return nil, err
	}

	return &types.MsgAddExpirationResponse{}, nil
}

func (m msgServer) ExtendExpiration(
	goCtx context.Context,
	msg *types.MsgExtendExpirationRequest,
) (*types.MsgExtendExpirationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// validate message
	if err := m.ValidateSetExpiration(ctx, msg.Expiration, msg.Signers, msg.MsgTypeURL()); err != nil {
		return nil, err
	}

	// add expiration
	if err := m.Keeper.UpdateExpiration(ctx, msg.Expiration); err != nil {
		return nil, err
	}

	return &types.MsgExtendExpirationResponse{}, nil
}

func (m msgServer) DeleteExpiration(goCtx context.Context, msg *types.MsgDeleteExpirationRequest) (*types.MsgDeleteExpirationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// validate message
	if err := m.Keeper.ValidateDeleteExpiration(ctx, msg.ModuleAssetId, msg.Signers, msg.MsgTypeURL()); err != nil {
		return nil, err
	}

	// delete message
	if err := m.Keeper.DeleteExpiration(ctx, msg.ModuleAssetId); err != nil {
		return nil, err
	}

	return &types.MsgDeleteExpirationResponse{}, nil
}
