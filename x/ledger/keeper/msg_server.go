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

func (k *MsgServer) ProcessFundTransfers(goCtx context.Context, req *ledger.MsgProcessFundTransfersRequest) (*ledger.MsgProcessFundTransfersResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	for _, ft := range req.Transfers {
		err := k.ProcessFundTransfer(ctx, ft)
		if err != nil {
			return nil, err
		}
	}

	resp := ledger.MsgProcessFundTransfersResponse{}
	return &resp, nil
}
