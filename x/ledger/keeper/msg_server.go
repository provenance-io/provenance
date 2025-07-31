package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/ledger/types"
	ledger "github.com/provenance-io/provenance/x/ledger/types"
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

// Append handles the MsgAppendRequest message
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

// UpdateBalances handles the MsgUpdateBalancesRequest message
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

// Create handles the MsgCreateRequest message
func (k *MsgServer) Create(goCtx context.Context, req *ledger.MsgCreateRequest) (*ledger.MsgCreateResponse, error) {
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

// UpdateStatus handles the MsgUpdateStatusRequest message
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

// UpdateInterestRate handles the MsgUpdateInterestRateRequest message
func (k *MsgServer) UpdateInterestRate(goCtx context.Context, req *ledger.MsgUpdateInterestRateRequest) (*ledger.MsgUpdateInterestRateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	authorityAddr, err := sdk.AccAddressFromBech32(req.Authority)
	if err != nil {
		return nil, err
	}

	err = k.UpdateLedgerInterestRate(ctx, authorityAddr, req.Key, req.InterestRate, req.InterestDayCount, req.InterestAccrual)
	if err != nil {
		return nil, err
	}

	resp := ledger.MsgUpdateInterestRateResponse{}
	return &resp, nil
}

// UpdatePayment handles the MsgUpdatePaymentRequest message
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

// UpdateMaturityDate handles the MsgUpdateMaturityDateRequest message
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

// TransferFundsWithSettlement handles the MsgTransferFundsWithSettlementRequest message
func (k *MsgServer) TransferFundsWithSettlement(goCtx context.Context, req *ledger.MsgTransferFundsWithSettlementRequest) (*ledger.MsgTransferFundsWithSettlementResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	authorityAddr, err := sdk.AccAddressFromBech32(req.Authority)
	if err != nil {
		return nil, err
	}

	for _, ft := range req.Transfers {
		err := k.TransferLedgerFundsWithSettlement(ctx, authorityAddr, ft)
		if err != nil {
			return nil, err
		}
	}

	resp := ledger.MsgTransferFundsWithSettlementResponse{}
	return &resp, nil
}

// Destroy handles the MsgDestroyRequest message
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

// CreateLedgerClass handles the MsgCreateClassRequest message
func (k *MsgServer) CreateClass(goCtx context.Context, req *ledger.MsgCreateClassRequest) (*ledger.MsgCreateClassResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if req == nil {
		return nil, types.NewLedgerCodedError(types.ErrCodeInvalidField, "request", "request is nil")
	}

	authority, err := sdk.AccAddressFromBech32(req.Authority)
	if err != nil {
		return nil, err
	}

	err = k.CreateLedgerClass(sdk.UnwrapSDKContext(ctx), authority, *req.LedgerClass)
	if err != nil {
		return nil, err
	}

	return &ledger.MsgCreateClassResponse{}, nil
}

// AddLedgerClassStatusType handles the MsgAddClassStatusTypeRequest message
func (k *MsgServer) AddClassStatusType(goCtx context.Context, req *ledger.MsgAddClassStatusTypeRequest) (*ledger.MsgAddClassStatusTypeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if req == nil {
		return nil, types.NewLedgerCodedError(types.ErrCodeInvalidField, "request", "request is nil")
	}

	authority, err := sdk.AccAddressFromBech32(req.Authority)
	if err != nil {
		return nil, err
	}

	err = k.AddLedgerClassStatusType(sdk.UnwrapSDKContext(ctx), authority, req.LedgerClassId, *req.StatusType)
	if err != nil {
		return nil, err
	}

	return &ledger.MsgAddClassStatusTypeResponse{}, nil
}

// AddLedgerClassEntryType handles the MsgAddClassEntryTypeRequest message
func (k *MsgServer) AddClassEntryType(goCtx context.Context, req *ledger.MsgAddClassEntryTypeRequest) (*ledger.MsgAddClassEntryTypeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if req == nil {
		return nil, types.NewLedgerCodedError(types.ErrCodeInvalidField, "request", "request is nil")
	}

	authority, err := sdk.AccAddressFromBech32(req.Authority)
	if err != nil {
		return nil, err
	}

	err = k.AddLedgerClassEntryType(sdk.UnwrapSDKContext(ctx), authority, req.LedgerClassId, *req.EntryType)
	if err != nil {
		return nil, err
	}

	return &ledger.MsgAddClassEntryTypeResponse{}, nil
}

// AddLedgerClassBucketType handles the MsgAddClassBucketTypeRequest message
func (k *MsgServer) AddClassBucketType(goCtx context.Context, req *ledger.MsgAddClassBucketTypeRequest) (*ledger.MsgAddClassBucketTypeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	authorityAddr, err := sdk.AccAddressFromBech32(req.Authority)
	if err != nil {
		return nil, err
	}

	err = k.AddLedgerClassBucketType(ctx, authorityAddr, req.LedgerClassId, *req.BucketType)
	if err != nil {
		return nil, err
	}

	resp := ledger.MsgAddClassBucketTypeResponse{}
	return &resp, nil
}

// BulkImport handles the MsgBulkImportRequest message
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
