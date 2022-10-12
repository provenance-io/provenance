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

// AddExpiration adds a new expiration record.
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

// ExtendExpiration extends an existing expiration.
// Note, the extension duration time is added to the existing expiration time.
func (m msgServer) ExtendExpiration(
	goCtx context.Context,
	msg *types.MsgExtendExpirationRequest,
) (*types.MsgExtendExpirationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// retrieve expiration
	expiration, err := m.Keeper.GetExpiration(ctx, msg.ModuleAssetId)
	if err != nil {
		return nil, err
	}
	if expiration == nil {
		return nil, types.ErrExtendExpiration.Wrapf("%s [%s]", types.ErrNotFound.Error(), msg.ModuleAssetId)
	}

	duration, err := types.ParseDuration(msg.Duration)
	if err != nil {
		return nil, err
	}
	expiration.Time = expiration.Time.Add(*duration)

	// validate message
	if err := m.ValidateSetExpiration(ctx, *expiration, msg.Signers, msg.MsgTypeURL()); err != nil {
		return nil, err
	}

	// update expiration details from extension payload
	if err := m.Keeper.ExtendExpiration(ctx, *expiration); err != nil {
		return nil, err
	}

	return &types.MsgExtendExpirationResponse{}, nil
}

func (m msgServer) InvokeExpiration(
	goCtx context.Context,
	msg *types.MsgInvokeExpirationRequest,
) (*types.MsgInvokeExpirationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// validate invocation request
	exp, err := m.Keeper.ValidateInvokeExpiration(ctx, msg.ModuleAssetId, msg.Signers, msg.MsgTypeURL())
	if err != nil {
		return nil, err
	}

	// resolve the depositor to the owner or fallback to first signer if not found and after expiration
	refundTo, err := m.Keeper.ResolveDepositor(ctx, *exp, msg)
	if err != nil {
		return nil, err
	}

	// execute expiration logic
	if err := m.Keeper.InvokeExpiration(ctx, msg.ModuleAssetId, refundTo); err != nil {
		return nil, err
	}

	return &types.MsgInvokeExpirationResponse{}, nil
}
