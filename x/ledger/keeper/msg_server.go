package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/ledger/types"
	ledger "github.com/provenance-io/provenance/x/ledger/types"
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
	ctx := sdk.UnwrapSDKContext(goCtx)

	authorityAddr, err := sdk.AccAddressFromBech32(req.Authority)
	if err != nil {
		return nil, err
	}

	err = k.AppendEntries(ctx, authorityAddr, req.Key, req.Entries)
	if err != nil {
		return nil, err
	}

	resp := ledger.MsgAppendResponse{}
	return &resp, nil
}

func (k *MsgServer) UpdateBalances(goCtx context.Context, req *ledger.MsgUpdateBalancesRequest) (*ledger.MsgUpdateBalancesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	authorityAddr, err := sdk.AccAddressFromBech32(req.Authority)
	if err != nil {
		return nil, err
	}

	err = k.UpdateEntryBalances(ctx, authorityAddr, req.Key, req.CorrelationId, req.BalanceAmounts, req.AppliedAmounts)
	if err != nil {
		return nil, err
	}

	resp := ledger.MsgUpdateBalancesResponse{}
	return &resp, nil
}

func (k *MsgServer) Create(goCtx context.Context, req *ledger.MsgCreateRequest) (*ledger.MsgCreateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	authorityAddr, err := sdk.AccAddressFromBech32(req.Authority)
	if err != nil {
		return nil, err
	}

	err = k.AddLedger(ctx, authorityAddr, *req.Ledger)
	if err != nil {
		return nil, err
	}

	resp := ledger.MsgCreateResponse{}
	return &resp, nil
}

func (k *MsgServer) UpdateStatus(goCtx context.Context, req *ledger.MsgUpdateStatusRequest) (*ledger.MsgUpdateStatusResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	authorityAddr, err := sdk.AccAddressFromBech32(req.Authority)
	if err != nil {
		return nil, err
	}

	err = k.UpdateLedgerStatus(ctx, authorityAddr, req.Key, req.StatusTypeId)
	if err != nil {
		return nil, err
	}

	resp := ledger.MsgUpdateStatusResponse{}
	return &resp, nil
}

func (k *MsgServer) UpdateInterestRate(goCtx context.Context, req *ledger.MsgUpdateInterestRateRequest) (*ledger.MsgUpdateInterestRateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	authorityAddr, err := sdk.AccAddressFromBech32(req.Authority)
	if err != nil {
		return nil, err
	}

	err = k.UpdateLedgerInterestRate(ctx, authorityAddr, req.Key, req.InterestRate, req.InterestDayCountConvention, req.InterestAccrualMethod)
	if err != nil {
		return nil, err
	}

	resp := ledger.MsgUpdateInterestRateResponse{}
	return &resp, nil
}

func (k *MsgServer) UpdatePayment(goCtx context.Context, req *ledger.MsgUpdatePaymentRequest) (*ledger.MsgUpdatePaymentResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	authorityAddr, err := sdk.AccAddressFromBech32(req.Authority)
	if err != nil {
		return nil, err
	}

	err = k.UpdateLedgerPayment(ctx, authorityAddr, req.Key, req.NextPmtAmt, req.NextPmtDate, req.PaymentFrequency)
	if err != nil {
		return nil, err
	}

	resp := ledger.MsgUpdatePaymentResponse{}
	return &resp, nil
}

func (k *MsgServer) UpdateMaturityDate(goCtx context.Context, req *ledger.MsgUpdateMaturityDateRequest) (*ledger.MsgUpdateMaturityDateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	authorityAddr, err := sdk.AccAddressFromBech32(req.Authority)
	if err != nil {
		return nil, err
	}

	err = k.UpdateLedgerMaturityDate(ctx, authorityAddr, req.Key, req.MaturityDate)
	if err != nil {
		return nil, err
	}

	resp := ledger.MsgUpdateMaturityDateResponse{}
	return &resp, nil
}

func (k *MsgServer) TransferFundsWithSettlement(goCtx context.Context, req *ledger.MsgTransferFundsWithSettlementRequest) (*ledger.MsgTransferFundsWithSettlementResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	authorityAddr, err := sdk.AccAddressFromBech32(req.Authority)
	if err != nil {
		return nil, err
	}

	for _, ft := range req.Transfers {
		err := k.ProcessTransferFundsWithSettlement(ctx, authorityAddr, ft)
		if err != nil {
			return nil, err
		}
	}

	resp := ledger.MsgTransferFundsWithSettlementResponse{}
	return &resp, nil
}

func (k *MsgServer) Destroy(goCtx context.Context, req *ledger.MsgDestroyRequest) (*ledger.MsgDestroyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	authorityAddr, err := sdk.AccAddressFromBech32(req.Authority)
	if err != nil {
		return nil, err
	}

	err = k.DestroyLedger(ctx, authorityAddr, req.Key)
	if err != nil {
		return nil, err
	}

	resp := ledger.MsgDestroyResponse{}
	return &resp, nil
}

// CreateLedgerClass handles the MsgCreateLedgerClassRequest message
func (k *MsgServer) CreateLedgerClass(goCtx context.Context, req *ledger.MsgCreateLedgerClassRequest) (*ledger.MsgCreateLedgerClassResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if req == nil {
		return nil, types.NewLedgerCodedError(types.ErrCodeInvalidField, "request", "request is nil")
	}

	authority, err := sdk.AccAddressFromBech32(req.Authority)
	if err != nil {
		return nil, err
	}

	err = k.AddLedgerClass(sdk.UnwrapSDKContext(ctx), authority, *req.LedgerClass)
	if err != nil {
		return nil, err
	}

	return &ledger.MsgCreateLedgerClassResponse{}, nil
}

// AddLedgerClassStatusType handles the MsgAddLedgerClassStatusTypeRequest message
func (k *MsgServer) AddLedgerClassStatusType(goCtx context.Context, req *ledger.MsgAddLedgerClassStatusTypeRequest) (*ledger.MsgAddLedgerClassStatusTypeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if req == nil {
		return nil, types.NewLedgerCodedError(types.ErrCodeInvalidField, "request", "request is nil")
	}

	authority, err := sdk.AccAddressFromBech32(req.Authority)
	if err != nil {
		return nil, err
	}

	err = k.AddClassStatusType(sdk.UnwrapSDKContext(ctx), authority, req.LedgerClassId, *req.StatusType)
	if err != nil {
		return nil, err
	}

	return &ledger.MsgAddLedgerClassStatusTypeResponse{}, nil
}

// AddLedgerClassEntryType handles the MsgAddLedgerClassEntryTypeRequest message
func (k *MsgServer) AddLedgerClassEntryType(goCtx context.Context, req *ledger.MsgAddLedgerClassEntryTypeRequest) (*ledger.MsgAddLedgerClassEntryTypeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if req == nil {
		return nil, types.NewLedgerCodedError(types.ErrCodeInvalidField, "request", "request is nil")
	}

	authority, err := sdk.AccAddressFromBech32(req.Authority)
	if err != nil {
		return nil, err
	}

	err = k.AddClassEntryType(sdk.UnwrapSDKContext(ctx), authority, req.LedgerClassId, *req.EntryType)
	if err != nil {
		return nil, err
	}

	return &ledger.MsgAddLedgerClassEntryTypeResponse{}, nil
}

// AddLedgerClassBucketType handles the MsgAddLedgerClassBucketTypeRequest message
func (k *MsgServer) AddLedgerClassBucketType(goCtx context.Context, req *ledger.MsgAddLedgerClassBucketTypeRequest) (*ledger.MsgAddLedgerClassBucketTypeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	authorityAddr, err := sdk.AccAddressFromBech32(req.Authority)
	if err != nil {
		return nil, err
	}

	err = k.AddClassBucketType(ctx, authorityAddr, req.LedgerClassId, *req.BucketType)
	if err != nil {
		return nil, err
	}

	resp := ledger.MsgAddLedgerClassBucketTypeResponse{}
	return &resp, nil
}

func (k *MsgServer) BulkImport(goCtx context.Context, req *ledger.MsgBulkImportRequest) (*ledger.MsgBulkImportResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	authorityAddr, err := sdk.AccAddressFromBech32(req.Authority)
	if err != nil {
		return nil, err
	}

	err = k.BulkImportLedgerData(ctx, authorityAddr, *req.GenesisState)
	if err != nil {
		return nil, err
	}

	resp := ledger.MsgBulkImportResponse{}
	return &resp, nil
}
