package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/ledger"
)

type MsgServer struct {
	Keeper
}

func NewMsgServer(k Keeper) ledger.MsgServer {
	ms := MsgServer{
		Keeper: k,
	}
	return &ms
}

func (k *MsgServer) Append(ctx context.Context, req *ledger.MsgAppendRequest) (*ledger.MsgAppendResponse, error) {
	resp := ledger.MsgAppendResponse{}
	return &resp, nil
}

func (k *MsgServer) Create(goCtx context.Context, req *ledger.MsgCreateRequest) (*ledger.MsgCreateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	ctx.Logger().Info("Doing some cool shit here...")

	resp := ledger.MsgCreateResponse{}
	return &resp, nil
}
