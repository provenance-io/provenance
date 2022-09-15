package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

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
		return nil, sdkerrors.Wrap(err, "failed to set expiration")
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

	// update expiration details from extension payload
	if err := m.Keeper.ExtendExpiration(ctx, msg.Expiration); err != nil {
		return nil, sdkerrors.Wrap(err, "failed to extend validation")
	}

	return &types.MsgExtendExpirationResponse{}, nil
}

//func (m msgServer) DeleteExpiration(goCtx context.Context, msg *types.MsgDeleteExpirationRequest) (*types.MsgDeleteExpirationResponse, error) {
//	ctx := sdk.UnwrapSDKContext(goCtx)
//
//	// validate message
//	if err := m.Keeper.ValidateDeleteExpiration(ctx, msg.ModuleAssetId, msg.Signers, msg.MsgTypeURL()); err != nil {
//		return nil, err
//	}
//
//	// delete message
//	if err := m.Keeper.DeleteExpiration(ctx, msg.ModuleAssetId); err != nil {
//		return nil, err
//	}
//
//	return &types.MsgDeleteExpirationResponse{}, nil
//}

func (m msgServer) InvokeExpiration(goCtx context.Context, msg *types.MsgInvokeExpirationRequest) (*types.MsgInvokeExpirationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// validate invocation request
	if err := m.Keeper.ValidateInvokeExpiration(ctx, msg.ModuleAssetId, msg.Signers, msg.MsgTypeURL()); err != nil {
		return nil, err
	}

	// execute expiration logic
	if err := m.Keeper.InvokeExpiration(ctx, msg.ModuleAssetId); err != nil {
		return nil, sdkerrors.Wrapf(types.ErrInvoke, fmt.Sprintf(": %v", err))
	}

	return &types.MsgInvokeExpirationResponse{}, nil
}
