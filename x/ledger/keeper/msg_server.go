package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/ledger/types"
)

type MsgServer struct {
	Keeper
}

func NewMsgServer(k Keeper) types.MsgServer {
	ms := MsgServer{
		Keeper: k,
	}
	return &ms
}

func (k *MsgServer) Append(goCtx context.Context, req *types.MsgAppendRequest) (*types.MsgAppendResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.HasLedger(ctx, req.Key) {
		return nil, types.NewErrCodeNotFound("ledger")
	}

	if err := k.RequireAuthority(ctx, req.Authority, req.Key.ToRegistryKey()); err != nil {
		return nil, err
	}

	if err := k.AppendEntries(ctx, req.Key, req.Entries); err != nil {
		return nil, err
	}

	return &types.MsgAppendResponse{}, nil
}

func (k *MsgServer) UpdateBalances(goCtx context.Context, req *types.MsgUpdateBalancesRequest) (*types.MsgUpdateBalancesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.HasLedger(ctx, req.Key) {
		return nil, types.NewErrCodeNotFound("ledger")
	}

	if err := k.RequireAuthority(ctx, req.Authority, req.Key.ToRegistryKey()); err != nil {
		return nil, err
	}

	if err := k.UpdateEntryBalances(ctx, req.Key, req.CorrelationId, req.BalanceAmounts, req.AppliedAmounts); err != nil {
		return nil, err
	}

	return &types.MsgUpdateBalancesResponse{}, nil
}

func (k *MsgServer) Create(goCtx context.Context, req *types.MsgCreateRequest) (*types.MsgCreateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if k.HasLedger(ctx, req.Ledger.Key) {
		return nil, types.NewErrCodeAlreadyExists("ledger")
	}

	if err := k.RequireAuthority(ctx, req.Authority, req.Ledger.Key.ToRegistryKey()); err != nil {
		return nil, err
	}
	if err := k.AddLedger(ctx, *req.Ledger); err != nil {
		return nil, err
	}

	return &types.MsgCreateResponse{}, nil
}

func (k *MsgServer) UpdateStatus(goCtx context.Context, req *types.MsgUpdateStatusRequest) (*types.MsgUpdateStatusResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.HasLedger(ctx, req.Key) {
		return nil, types.NewErrCodeNotFound("ledger")
	}

	if err := k.RequireAuthority(ctx, req.Authority, req.Key.ToRegistryKey()); err != nil {
		return nil, err
	}

	if err := k.UpdateLedgerStatus(ctx, req.Key, req.StatusTypeId); err != nil {
		return nil, err
	}

	return &types.MsgUpdateStatusResponse{}, nil
}

func (k *MsgServer) UpdateInterestRate(goCtx context.Context, req *types.MsgUpdateInterestRateRequest) (*types.MsgUpdateInterestRateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.HasLedger(ctx, req.Key) {
		return nil, types.NewErrCodeNotFound("ledger")
	}

	if err := k.RequireAuthority(ctx, req.Authority, req.Key.ToRegistryKey()); err != nil {
		return nil, err
	}

	if err := k.UpdateLedgerInterestRate(ctx, req.Key, req.InterestRate, req.InterestDayCountConvention, req.InterestAccrualMethod); err != nil {
		return nil, err
	}

	return &types.MsgUpdateInterestRateResponse{}, nil
}

func (k *MsgServer) UpdatePayment(goCtx context.Context, req *types.MsgUpdatePaymentRequest) (*types.MsgUpdatePaymentResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.HasLedger(ctx, req.Key) {
		return nil, types.NewErrCodeNotFound("ledger")
	}

	if err := k.RequireAuthority(ctx, req.Authority, req.Key.ToRegistryKey()); err != nil {
		return nil, err
	}

	if err := k.UpdateLedgerPayment(ctx, req.Key, req.NextPmtAmt, req.NextPmtDate, req.PaymentFrequency); err != nil {
		return nil, err
	}

	return &types.MsgUpdatePaymentResponse{}, nil
}

func (k *MsgServer) UpdateMaturityDate(goCtx context.Context, req *types.MsgUpdateMaturityDateRequest) (*types.MsgUpdateMaturityDateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.HasLedger(ctx, req.Key) {
		return nil, types.NewErrCodeNotFound("ledger")
	}

	if err := k.RequireAuthority(ctx, req.Authority, req.Key.ToRegistryKey()); err != nil {
		return nil, err
	}

	if err := k.UpdateLedgerMaturityDate(ctx, req.Key, req.MaturityDate); err != nil {
		return nil, err
	}

	return &types.MsgUpdateMaturityDateResponse{}, nil
}

func (k *MsgServer) TransferFundsWithSettlement(goCtx context.Context, req *types.MsgTransferFundsWithSettlementRequest) (*types.MsgTransferFundsWithSettlementResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.HasLedger(ctx, req.Transfers[0].Key) {
		return nil, types.NewErrCodeNotFound("ledger")
	}

	if err := k.RequireAuthority(ctx, req.Authority, req.Transfers[0].Key.ToRegistryKey()); err != nil {
		return nil, err
	}

	// Ignore the error here, as it is validated in the ValidateBasic method.
	authorityAddr, _ := sdk.AccAddressFromBech32(req.Authority)

	for _, ft := range req.Transfers {
		err := k.ProcessTransferFundsWithSettlement(ctx, authorityAddr, ft)
		if err != nil {
			return nil, err
		}
	}

	return &types.MsgTransferFundsWithSettlementResponse{}, nil
}

func (k *MsgServer) Destroy(goCtx context.Context, req *types.MsgDestroyRequest) (*types.MsgDestroyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.HasLedger(ctx, req.Key) {
		return nil, types.NewErrCodeNotFound("ledger")
	}

	if err := k.RequireAuthority(ctx, req.Authority, req.Key.ToRegistryKey()); err != nil {
		return nil, err
	}

	if err := k.DestroyLedger(ctx, req.Key); err != nil {
		return nil, err
	}

	return &types.MsgDestroyResponse{}, nil
}

// CreateLedgerClass handles the MsgCreateLedgerClassRequest message
func (k *MsgServer) CreateLedgerClass(goCtx context.Context, req *types.MsgCreateLedgerClassRequest) (*types.MsgCreateLedgerClassResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Note: No authorization is required for creation since the basic validation checks that the maintainer
	// matches the authority, and the create will fail if the class already exists.

	if err := k.AddLedgerClass(ctx, *req.LedgerClass); err != nil {
		return nil, err
	}

	return &types.MsgCreateLedgerClassResponse{}, nil
}

// AddLedgerClassStatusType handles the MsgAddLedgerClassStatusTypeRequest message
func (k *MsgServer) AddLedgerClassStatusType(goCtx context.Context, req *types.MsgAddLedgerClassStatusTypeRequest) (*types.MsgAddLedgerClassStatusTypeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	has, err := k.LedgerClasses.Has(ctx, req.LedgerClassId)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, types.NewErrCodeNotFound("ledger_class")
	}

	if !k.IsLedgerClassMaintainer(ctx, req.Authority, req.LedgerClassId) {
		return nil, types.NewErrCodeUnauthorized("ledger class maintainer")
	}

	err = k.AddClassStatusType(ctx, req.LedgerClassId, *req.StatusType)
	if err != nil {
		return nil, err
	}

	return &types.MsgAddLedgerClassStatusTypeResponse{}, nil
}

// AddLedgerClassEntryType handles the MsgAddLedgerClassEntryTypeRequest message
func (k *MsgServer) AddLedgerClassEntryType(goCtx context.Context, req *types.MsgAddLedgerClassEntryTypeRequest) (*types.MsgAddLedgerClassEntryTypeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	has, err := k.LedgerClasses.Has(ctx, req.LedgerClassId)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, types.NewErrCodeNotFound("ledger_class")
	}

	if !k.IsLedgerClassMaintainer(ctx, req.Authority, req.LedgerClassId) {
		return nil, types.NewErrCodeUnauthorized("ledger class maintainer")
	}

	if err := k.AddClassEntryType(ctx, req.LedgerClassId, *req.EntryType); err != nil {
		return nil, err
	}

	return &types.MsgAddLedgerClassEntryTypeResponse{}, nil
}

// AddLedgerClassBucketType handles the MsgAddLedgerClassBucketTypeRequest message
func (k *MsgServer) AddLedgerClassBucketType(goCtx context.Context, req *types.MsgAddLedgerClassBucketTypeRequest) (*types.MsgAddLedgerClassBucketTypeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	has, err := k.LedgerClasses.Has(ctx, req.LedgerClassId)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, types.NewErrCodeNotFound("ledger_class")
	}

	if !k.IsLedgerClassMaintainer(ctx, req.Authority, req.LedgerClassId) {
		return nil, types.NewErrCodeUnauthorized("ledger class maintainer")
	}

	if err := k.AddClassBucketType(ctx, req.LedgerClassId, *req.BucketType); err != nil {
		return nil, err
	}

	return &types.MsgAddLedgerClassBucketTypeResponse{}, nil
}

func (k *MsgServer) BulkImport(goCtx context.Context, req *types.MsgBulkImportRequest) (*types.MsgBulkImportResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO Add authority check

	if err := k.BulkImportLedgerData(ctx, *req.GenesisState); err != nil {
		return nil, err
	}

	return &types.MsgBulkImportResponse{}, nil
}
