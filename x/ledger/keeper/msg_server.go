package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/ledger"
)

type MsgServer struct {
	BaseKeeper
}

func NewMsgServer(k BaseKeeper) ledger.MsgServer {
	ms := MsgServer{
		BaseKeeper: k,
	}
	return &ms
}

func (k *MsgServer) Append(goCtx context.Context, req *ledger.MsgAppendRequest) (*ledger.MsgAppendResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := k.AppendEntries(ctx, req.NftAddress, req.Entries)
	if err != nil {
		return nil, err
	}

	resp := ledger.MsgAppendResponse{}
	return &resp, nil
}

func (k *MsgServer) Create(goCtx context.Context, req *ledger.MsgCreateRequest) (*ledger.MsgCreateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := k.CreateLedger(ctx, *req.Ledger)
	if err != nil {
		return nil, err
	}

	resp := ledger.MsgCreateResponse{}
	return &resp, nil
}

func (k *MsgServer) ProcessFundTransfers(goCtx context.Context, req *ledger.MsgProcessFundTransfersRequest) (*ledger.MsgProcessFundTransfersResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	for _, ft := range req.Transfers {
		err := k.TransferFunds(ctx, ft)
		if err != nil {
			return nil, err
		}
	}

	resp := ledger.MsgProcessFundTransfersResponse{}
	return &resp, nil
}

func (k *MsgServer) ProcessFundTransfersWithSettlement(goCtx context.Context, req *ledger.MsgProcessFundTransfersWithSettlementRequest) (*ledger.MsgProcessFundTransfersResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	for _, ft := range req.Transfers {
		err := k.TransferFundsWithSettlement(ctx, ft)
		if err != nil {
			return nil, err
		}
	}

	resp := ledger.MsgProcessFundTransfersResponse{}
	return &resp, nil
}

func (k *MsgServer) Destroy(goCtx context.Context, req *ledger.MsgDestroyRequest) (*ledger.MsgDestroyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := k.DestroyLedger(ctx, req.NftAddress)
	if err != nil {
		return nil, err
	}

	resp := ledger.MsgDestroyResponse{}
	return &resp, nil
}
