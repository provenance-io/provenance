package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/provenance-io/provenance/x/oracle/types"
)

type msgServer struct {
	*Keeper
}

// NewMsgServerImpl returns an implementation of the account MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper *Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// UpdateOracle changes the oracle's address to the provided one
func (s msgServer) UpdateOracle(goCtx context.Context, msg *types.MsgUpdateOracleRequest) (*types.MsgUpdateOracleResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg.Authority != s.Keeper.GetAuthority() {
		return nil, sdkerrors.ErrUnauthorized.Wrapf("expected authority %s got %s", s.Keeper.GetAuthority(), msg.GetAuthority())
	}

	s.Keeper.SetOracle(ctx, sdk.MustAccAddressFromBech32(msg.Address))

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		),
	)

	return &types.MsgUpdateOracleResponse{}, nil
}

// SendQueryOracle sends an icq to another chain's oracle
func (k msgServer) SendQueryOracle(goCtx context.Context, msg *types.MsgSendQueryOracleRequest) (*types.MsgSendQueryOracleResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	seq, err := k.QueryOracle(ctx, msg.Query, msg.Channel)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		),
	)

	return &types.MsgSendQueryOracleResponse{
		Sequence: seq,
	}, nil
}
