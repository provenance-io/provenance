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

func (k *MsgServer) AppendTx(goCtx context.Context, req *ledger.MsgAppendRequest) (*ledger.MsgAppendResponse, error) {
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

func (k *MsgServer) UpdateBalancesTx(goCtx context.Context, req *ledger.MsgUpdateBalancesRequest) (*ledger.MsgUpdateBalancesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	authorityAddr, err := sdk.AccAddressFromBech32(req.Authority)
	if err != nil {
		return nil, err
	}

	err = k.UpdateEntryBalances(ctx, authorityAddr, req.Key, req.CorrelationId, req.BucketBalances, req.AppliedAmounts)
	if err != nil {
		return nil, err
	}

	resp := ledger.MsgUpdateBalancesResponse{}
	return &resp, nil
}

func (k *MsgServer) CreateTx(goCtx context.Context, req *ledger.MsgCreateRequest) (*ledger.MsgCreateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	authorityAddr, err := sdk.AccAddressFromBech32(req.Authority)
	if err != nil {
		return nil, err
	}

	err = k.CreateLedger(ctx, authorityAddr, *req.Ledger)
	if err != nil {
		return nil, err
	}

	resp := ledger.MsgCreateResponse{}
	return &resp, nil
}

func (k *MsgServer) UpdateStatusTx(goCtx context.Context, req *ledger.MsgUpdateStatusRequest) (*ledger.MsgUpdateStatusResponse, error) {
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

func (k *MsgServer) UpdateInterestRateTx(goCtx context.Context, req *ledger.MsgUpdateInterestRateRequest) (*ledger.MsgUpdateInterestRateResponse, error) {
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

func (k *MsgServer) UpdatePaymentTx(goCtx context.Context, req *ledger.MsgUpdatePaymentRequest) (*ledger.MsgUpdatePaymentResponse, error) {
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

func (k *MsgServer) UpdateMaturityDateTx(goCtx context.Context, req *ledger.MsgUpdateMaturityDateRequest) (*ledger.MsgUpdateMaturityDateResponse, error) {
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

func (k *MsgServer) FundAssetTx(goCtx context.Context, req *ledger.MsgFundAssetRequest) (*ledger.MsgFundAssetResponse, error) {
	return nil, nil
}

func (k *MsgServer) TransferFundsTx(goCtx context.Context, req *ledger.MsgTransferFundsRequest) (*ledger.MsgTransferFundsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	for _, ft := range req.Transfers {
		err := k.TransferFunds(ctx, ft)
		if err != nil {
			return nil, err
		}
	}

	resp := ledger.MsgTransferFundsResponse{}
	return &resp, nil
}

func (k *MsgServer) TransferFundsWithSettlementTx(goCtx context.Context, req *ledger.MsgTransferFundsWithSettlementRequest) (*ledger.MsgTransferFundsWithSettlementResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	for _, ft := range req.Transfers {
		err := k.TransferFundsWithSettlement(ctx, ft)
		if err != nil {
			return nil, err
		}
	}

	resp := ledger.MsgTransferFundsWithSettlementResponse{}
	return &resp, nil
}

func (k *MsgServer) DestroyTx(goCtx context.Context, req *ledger.MsgDestroyRequest) (*ledger.MsgDestroyResponse, error) {
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
func (k *MsgServer) CreateLedgerClassTx(goCtx context.Context, req *ledger.MsgCreateLedgerClassRequest) (*ledger.MsgCreateLedgerClassResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if req == nil {
		return nil, NewLedgerCodedError(ErrCodeInvalidField, "request", "request is nil")
	}

	authority, err := sdk.AccAddressFromBech32(req.Authority)
	if err != nil {
		return nil, err
	}

	err = k.CreateLedgerClass(sdk.UnwrapSDKContext(ctx), authority, *req.LedgerClass)
	if err != nil {
		return nil, err
	}

	return &ledger.MsgCreateLedgerClassResponse{}, nil
}

// AddLedgerClassStatusType handles the MsgAddLedgerClassStatusTypeRequest message
func (k *MsgServer) AddLedgerClassStatusTypeTx(goCtx context.Context, req *ledger.MsgAddLedgerClassStatusTypeRequest) (*ledger.MsgAddLedgerClassStatusTypeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if req == nil {
		return nil, NewLedgerCodedError(ErrCodeInvalidField, "request", "request is nil")
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
func (k *MsgServer) AddLedgerClassEntryTypeTx(goCtx context.Context, req *ledger.MsgAddLedgerClassEntryTypeRequest) (*ledger.MsgAddLedgerClassEntryTypeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if req == nil {
		return nil, NewLedgerCodedError(ErrCodeInvalidField, "request", "request is nil")
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
func (k *MsgServer) AddLedgerClassBucketTypeTx(goCtx context.Context, req *ledger.MsgAddLedgerClassBucketTypeRequest) (*ledger.MsgAddLedgerClassBucketTypeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if req == nil {
		return nil, NewLedgerCodedError(ErrCodeInvalidField, "request", "request is nil")
	}

	authority, err := sdk.AccAddressFromBech32(req.Authority)
	if err != nil {
		return nil, err
	}

	err = k.AddClassBucketType(sdk.UnwrapSDKContext(ctx), authority, req.LedgerClassId, *req.BucketType)
	if err != nil {
		return nil, err
	}

	return &ledger.MsgAddLedgerClassBucketTypeResponse{}, nil
}
