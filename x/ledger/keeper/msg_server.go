package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
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

	if err := k.RequireAuthority(ctx, req.Authority, req.Key.ToRegistryKey()); err != nil {
		return nil, err
	}

	if err := k.AppendEntries(ctx, req.Key, req.Entries); err != nil {
		return nil, err
	}

	return &ledger.MsgAppendResponse{}, nil
}

func (k *MsgServer) UpdateBalances(goCtx context.Context, req *ledger.MsgUpdateBalancesRequest) (*ledger.MsgUpdateBalancesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := k.RequireAuthority(ctx, req.Authority, req.Key.ToRegistryKey()); err != nil {
		return nil, err
	}

	if err := k.UpdateEntryBalances(ctx, req.Key, req.CorrelationId, req.BalanceAmounts, req.AppliedAmounts); err != nil {
		return nil, err
	}

	return &ledger.MsgUpdateBalancesResponse{}, nil
}

func (k *MsgServer) Create(goCtx context.Context, req *ledger.MsgCreateRequest) (*ledger.MsgCreateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := k.RequireAuthority(ctx, req.Authority, req.Ledger.Key.ToRegistryKey()); err != nil {
		return nil, err
	}
	if err := k.AddLedger(ctx, *req.Ledger); err != nil {
		return nil, err
	}

	return &ledger.MsgCreateResponse{}, nil
}

func (k *MsgServer) UpdateStatus(goCtx context.Context, req *ledger.MsgUpdateStatusRequest) (*ledger.MsgUpdateStatusResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := k.RequireAuthority(ctx, req.Authority, req.Key.ToRegistryKey()); err != nil {
		return nil, err
	}

	if err := k.UpdateLedgerStatus(ctx, req.Key, req.StatusTypeId); err != nil {
		return nil, err
	}

	return &ledger.MsgUpdateStatusResponse{}, nil
}

func (k *MsgServer) UpdateInterestRate(goCtx context.Context, req *ledger.MsgUpdateInterestRateRequest) (*ledger.MsgUpdateInterestRateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := k.RequireAuthority(ctx, req.Authority, req.Key.ToRegistryKey()); err != nil {
		return nil, err
	}

	if err := k.UpdateLedgerInterestRate(ctx, req.Key, req.InterestRate, req.InterestDayCountConvention, req.InterestAccrualMethod); err != nil {
		return nil, err
	}

	return &ledger.MsgUpdateInterestRateResponse{}, nil
}

func (k *MsgServer) UpdatePayment(goCtx context.Context, req *ledger.MsgUpdatePaymentRequest) (*ledger.MsgUpdatePaymentResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := k.RequireAuthority(ctx, req.Authority, req.Key.ToRegistryKey()); err != nil {
		return nil, err
	}

	if err := k.UpdateLedgerPayment(ctx, req.Key, req.NextPmtAmt, req.NextPmtDate, req.PaymentFrequency); err != nil {
		return nil, err
	}

	return &ledger.MsgUpdatePaymentResponse{}, nil
}

func (k *MsgServer) UpdateMaturityDate(goCtx context.Context, req *ledger.MsgUpdateMaturityDateRequest) (*ledger.MsgUpdateMaturityDateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := k.RequireAuthority(ctx, req.Authority, req.Key.ToRegistryKey()); err != nil {
		return nil, err
	}

	if err := k.UpdateLedgerMaturityDate(ctx, req.Key, req.MaturityDate); err != nil {
		return nil, err
	}

	return &ledger.MsgUpdateMaturityDateResponse{}, nil
}

func (k *MsgServer) TransferFundsWithSettlement(goCtx context.Context, req *ledger.MsgTransferFundsWithSettlementRequest) (*ledger.MsgTransferFundsWithSettlementResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

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

	return &ledger.MsgTransferFundsWithSettlementResponse{}, nil
}

func (k *MsgServer) Destroy(goCtx context.Context, req *ledger.MsgDestroyRequest) (*ledger.MsgDestroyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := k.RequireAuthority(ctx, req.Authority, req.Key.ToRegistryKey()); err != nil {
		return nil, err
	}

	if err := k.DestroyLedger(ctx, req.Key); err != nil {
		return nil, err
	}

	return &ledger.MsgDestroyResponse{}, nil
}

// CreateLedgerClass handles the MsgCreateLedgerClassRequest message
func (k *MsgServer) CreateLedgerClass(goCtx context.Context, req *ledger.MsgCreateLedgerClassRequest) (*ledger.MsgCreateLedgerClassResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Note: No authorization is required for creation since the basic validation checks that the maintainer
	// matches the authority, and the create will fail if the class already exists.

	if err := k.AddLedgerClass(ctx, *req.LedgerClass); err != nil {
		return nil, err
	}

	return &ledger.MsgCreateLedgerClassResponse{}, nil
}

// AddLedgerClassStatusType handles the MsgAddLedgerClassStatusTypeRequest message
func (k *MsgServer) AddLedgerClassStatusType(goCtx context.Context, req *ledger.MsgAddLedgerClassStatusTypeRequest) (*ledger.MsgAddLedgerClassStatusTypeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.IsLedgerClassMaintainer(ctx, req.Authority, req.LedgerClassId) {
		return nil, ledger.NewLedgerCodedError(ledger.ErrCodeUnauthorized)
	}

	err := k.AddClassStatusType(ctx, req.LedgerClassId, *req.StatusType)
	if err != nil {
		return nil, err
	}

	return &ledger.MsgAddLedgerClassStatusTypeResponse{}, nil
}

// AddLedgerClassEntryType handles the MsgAddLedgerClassEntryTypeRequest message
func (k *MsgServer) AddLedgerClassEntryType(goCtx context.Context, req *ledger.MsgAddLedgerClassEntryTypeRequest) (*ledger.MsgAddLedgerClassEntryTypeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.IsLedgerClassMaintainer(ctx, req.Authority, req.LedgerClassId) {
		return nil, ledger.NewLedgerCodedError(ledger.ErrCodeUnauthorized)
	}

	if err := k.AddClassEntryType(ctx, req.LedgerClassId, *req.EntryType); err != nil {
		return nil, err
	}

	return &ledger.MsgAddLedgerClassEntryTypeResponse{}, nil
}

// AddLedgerClassBucketType handles the MsgAddLedgerClassBucketTypeRequest message
func (k *MsgServer) AddLedgerClassBucketType(goCtx context.Context, req *ledger.MsgAddLedgerClassBucketTypeRequest) (*ledger.MsgAddLedgerClassBucketTypeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.IsLedgerClassMaintainer(ctx, req.Authority, req.LedgerClassId) {
		return nil, ledger.NewLedgerCodedError(ledger.ErrCodeUnauthorized)
	}

	if err := k.AddClassBucketType(ctx, req.LedgerClassId, *req.BucketType); err != nil {
		return nil, err
	}

	return &ledger.MsgAddLedgerClassBucketTypeResponse{}, nil
}

func (k *MsgServer) BulkImport(goCtx context.Context, req *ledger.MsgBulkImportRequest) (*ledger.MsgBulkImportResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO Add authority check

	if err := k.BulkImportLedgerData(ctx, *req.GenesisState); err != nil {
		return nil, err
	}

	return &ledger.MsgBulkImportResponse{}, nil
}
