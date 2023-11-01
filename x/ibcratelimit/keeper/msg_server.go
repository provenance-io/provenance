package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/ibcratelimit"
)

// MsgServer is an alias for a Keeper that implements the ibcratelimit.MsgServer interface.
type MsgServer struct {
	Keeper
}

func NewMsgServer(k Keeper) ibcratelimit.MsgServer {
	return MsgServer{
		Keeper: k,
	}
}

var _ ibcratelimit.MsgServer = MsgServer{}

// GovUpdateParams is a governance proposal endpoint for updating the ibcratelimit module's params.
func (k MsgServer) GovUpdateParams(goCtx context.Context, msg *ibcratelimit.MsgGovUpdateParamsRequest) (*ibcratelimit.MsgGovUpdateParamsResponse, error) {
	if err := k.ValidateAuthority(msg.Authority); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	k.SetParams(ctx, msg.Params)
	k.emitEvent(ctx, ibcratelimit.NewEventParamsUpdated())

	return &ibcratelimit.MsgGovUpdateParamsResponse{}, nil
}
