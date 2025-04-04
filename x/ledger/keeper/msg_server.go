package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/ledger"
)

type MsgServer struct {
	LedgerKeeper
}

func NewMsgServer(k LedgerKeeper) ledger.MsgServer {
	ms := MsgServer{
		LedgerKeeper: k,
	}
	return &ms
}

func (k *MsgServer) Append(goCtx context.Context, req *ledger.MsgAppendRequest) (*ledger.MsgAppendResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := k.AppendEntry(ctx, req.NftAddress, *req.Entry)
	if err != nil {
		return nil, err
	}

	resp := ledger.MsgAppendResponse{}
	return &resp, nil
}

func (k *MsgServer) Create(goCtx context.Context, req *ledger.MsgCreateRequest) (*ledger.MsgCreateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	l := ledger.Ledger{
		// We omit the nftAddress intentionally as a minor optimization since it is also our data key.
		// NftAddress: nftAddress,

		NftAddress: req.NftAddress,
		Denom:      req.Denom,
	}

	err := k.CreateLedger(ctx, l)
	if err != nil {
		return nil, err
	}

	resp := ledger.MsgCreateResponse{}
	return &resp, nil
}
