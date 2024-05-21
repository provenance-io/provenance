package keeper

import (
	"context"
	"errors"

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
//
//nolint:staticcheck // SA1019 Suppress warning for deprecated MsgGovUpdateParamsRequest usage
func (k MsgServer) GovUpdateParams(_ context.Context, _ *ibcratelimit.MsgGovUpdateParamsRequest) (*ibcratelimit.MsgGovUpdateParamsResponse, error) {
	return nil, errors.New("deprecated and unusable")
}

// UpdateParams is a governance proposal endpoint for updating the ibcratelimit module's params.
func (k MsgServer) UpdateParams(goCtx context.Context, msg *ibcratelimit.MsgUpdateParamsRequest) (*ibcratelimit.MsgUpdateParamsResponse, error) {
	if err := k.ValidateAuthority(msg.Authority); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	k.SetParams(ctx, msg.Params)
	k.emitEvent(ctx, ibcratelimit.NewEventParamsUpdated())

	return &ibcratelimit.MsgUpdateParamsResponse{}, nil
}
