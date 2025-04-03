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

func (k *MsgServer) Append(goCtx context.Context, req *ledger.MsgAppendRequest) (*ledger.MsgAppendResponse, error) {
	resp := ledger.MsgAppendResponse{}
	return &resp, nil
}

func (k *MsgServer) Create(goCtx context.Context, req *ledger.MsgCreateRequest) (*ledger.MsgCreateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO VALIDATE

	err := k.CreateLedger(ctx, req.NftAddress, req.Denom)
	if err != nil {
		return nil, err
	}

	resp := ledger.MsgCreateResponse{}
	return &resp, nil
}
