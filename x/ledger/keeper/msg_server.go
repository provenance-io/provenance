package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/ledger/types"
)

type MsgServer struct {
	Keeper
}

func NewMsgServer(k Keeper) types.MsgServer { return &MsgServer{Keeper: k} }

// CreateLedger creates a new NFT ledger.
func (k *MsgServer) CreateLedger(goCtx context.Context, req *types.MsgCreateLedgerRequest) (*types.MsgCreateLedgerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if k.HasLedger(ctx, req.Ledger.Key) {
		return nil, types.NewErrCodeAlreadyExists("ledger")
	}

	if err := k.RequireAuthorization(ctx, req.Signer, req.Ledger.Key.ToRegistryKey()); err != nil {
		return nil, err
	}
	if err := k.AddLedger(ctx, *req.Ledger); err != nil {
		return nil, err
	}

	return &types.MsgCreateLedgerResponse{}, nil
}

// UpdateStatus updates the Status of a ledger.
func (k *MsgServer) UpdateStatus(goCtx context.Context, req *types.MsgUpdateStatusRequest) (*types.MsgUpdateStatusResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.HasLedger(ctx, req.Key) {
		return nil, types.NewErrCodeNotFound("ledger")
	}

	if err := k.RequireAuthorization(ctx, req.Signer, req.Key.ToRegistryKey()); err != nil {
		return nil, err
	}

	if err := k.UpdateLedgerStatus(ctx, req.Key, req.StatusTypeId); err != nil {
		return nil, err
	}

	return &types.MsgUpdateStatusResponse{}, nil
}

// UpdateInterestRate updates the interest rate of a ledger.
func (k *MsgServer) UpdateInterestRate(goCtx context.Context, req *types.MsgUpdateInterestRateRequest) (*types.MsgUpdateInterestRateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.HasLedger(ctx, req.Key) {
		return nil, types.NewErrCodeNotFound("ledger")
	}

	if err := k.RequireAuthorization(ctx, req.Signer, req.Key.ToRegistryKey()); err != nil {
		return nil, err
	}

	if err := k.UpdateLedgerInterestRate(ctx, req.Key, req.InterestRate, req.InterestDayCountConvention, req.InterestAccrualMethod); err != nil {
		return nil, err
	}

	return &types.MsgUpdateInterestRateResponse{}, nil
}

// UpdatePayment updates the payment amount, next payment date, and payment frequency of a ledger.
func (k *MsgServer) UpdatePayment(goCtx context.Context, req *types.MsgUpdatePaymentRequest) (*types.MsgUpdatePaymentResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.HasLedger(ctx, req.Key) {
		return nil, types.NewErrCodeNotFound("ledger")
	}

	if err := k.RequireAuthorization(ctx, req.Signer, req.Key.ToRegistryKey()); err != nil {
		return nil, err
	}

	if err := k.UpdateLedgerPayment(ctx, req.Key, req.NextPmtAmt, req.NextPmtDate, req.PaymentFrequency); err != nil {
		return nil, err
	}

	return &types.MsgUpdatePaymentResponse{}, nil
}

// UpdateMaturityDate updates the maturity date of a ledger.
func (k *MsgServer) UpdateMaturityDate(goCtx context.Context, req *types.MsgUpdateMaturityDateRequest) (*types.MsgUpdateMaturityDateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.HasLedger(ctx, req.Key) {
		return nil, types.NewErrCodeNotFound("ledger")
	}

	if err := k.RequireAuthorization(ctx, req.Signer, req.Key.ToRegistryKey()); err != nil {
		return nil, err
	}

	if err := k.UpdateLedgerMaturityDate(ctx, req.Key, req.MaturityDate); err != nil {
		return nil, err
	}

	return &types.MsgUpdateMaturityDateResponse{}, nil
}

// Append adds an entry to a ledger.
func (k *MsgServer) Append(goCtx context.Context, req *types.MsgAppendRequest) (*types.MsgAppendResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.HasLedger(ctx, req.Key) {
		return nil, types.NewErrCodeNotFound("ledger")
	}

	if err := k.RequireAuthorization(ctx, req.Signer, req.Key.ToRegistryKey()); err != nil {
		return nil, err
	}

	if err := k.AppendEntries(ctx, req.Key, req.Entries); err != nil {
		return nil, err
	}

	return &types.MsgAppendResponse{}, nil
}

// UpdateBalances updates the balances for a ledger entry, allowing for retroactive adjustments to be applied.
func (k *MsgServer) UpdateBalances(goCtx context.Context, req *types.MsgUpdateBalancesRequest) (*types.MsgUpdateBalancesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.HasLedger(ctx, req.Key) {
		return nil, types.NewErrCodeNotFound("ledger")
	}

	if err := k.RequireAuthorization(ctx, req.Signer, req.Key.ToRegistryKey()); err != nil {
		return nil, err
	}

	if err := k.UpdateEntryBalances(ctx, req.Key, req.CorrelationId, req.BalanceAmounts, req.AppliedAmounts); err != nil {
		return nil, err
	}

	return &types.MsgUpdateBalancesResponse{}, nil
}

// TransferFundsWithSettlement processes multiple fund transfers with manual settlement instructions.
func (k *MsgServer) TransferFundsWithSettlement(goCtx context.Context, req *types.MsgTransferFundsWithSettlementRequest) (*types.MsgTransferFundsWithSettlementResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate each transfer targets an existing ledger and the signer has authority.
	for _, ft := range req.Transfers {
		if !k.HasLedger(ctx, ft.Key) {
			return nil, types.NewErrCodeNotFound("ledger")
		}
		if err := k.RequireAuthorization(ctx, req.Signer, ft.Key.ToRegistryKey()); err != nil {
			return nil, err
		}
	}

	// Ignore the error here, as it is validated in the ValidateBasic method.
	signerAddr, _ := sdk.AccAddressFromBech32(req.Signer)

	for _, ft := range req.Transfers {
		err := k.ProcessTransferFundsWithSettlement(ctx, signerAddr, ft)
		if err != nil {
			return nil, err
		}
	}

	return &types.MsgTransferFundsWithSettlementResponse{}, nil
}

// Destroy destroys a ledger by NFT address.
func (k *MsgServer) Destroy(goCtx context.Context, req *types.MsgDestroyRequest) (*types.MsgDestroyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.HasLedger(ctx, req.Key) {
		return nil, types.NewErrCodeNotFound("ledger")
	}

	if err := k.RequireAuthorization(ctx, req.Signer, req.Key.ToRegistryKey()); err != nil {
		return nil, err
	}

	if err := k.DestroyLedger(ctx, req.Key); err != nil {
		return nil, err
	}

	return &types.MsgDestroyResponse{}, nil
}

// CreateLedgerClass creates a new ledger class.
func (k *MsgServer) CreateLedgerClass(goCtx context.Context, req *types.MsgCreateLedgerClassRequest) (*types.MsgCreateLedgerClassResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Note: No authorization is required for creation since the basic validation checks that the maintainer
	// matches the signer, and the create will fail if the class already exists.

	if err := k.AddLedgerClass(ctx, *req.LedgerClass); err != nil {
		return nil, err
	}

	return &types.MsgCreateLedgerClassResponse{}, nil
}

// AddLedgerClassStatusType adds a status type to a ledger class.
func (k *MsgServer) AddLedgerClassStatusType(goCtx context.Context, req *types.MsgAddLedgerClassStatusTypeRequest) (*types.MsgAddLedgerClassStatusTypeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	has, err := k.LedgerClasses.Has(ctx, req.LedgerClassId)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, types.NewErrCodeNotFound("ledger_class")
	}

	if !k.IsLedgerClassMaintainer(ctx, req.Signer, req.LedgerClassId) {
		return nil, types.NewErrCodeUnauthorized("ledger class maintainer")
	}

	err = k.AddClassStatusType(ctx, req.LedgerClassId, *req.StatusType)
	if err != nil {
		return nil, err
	}

	return &types.MsgAddLedgerClassStatusTypeResponse{}, nil
}

// AddLedgerClassEntryType adds an entry type to a ledger class.
func (k *MsgServer) AddLedgerClassEntryType(goCtx context.Context, req *types.MsgAddLedgerClassEntryTypeRequest) (*types.MsgAddLedgerClassEntryTypeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	has, err := k.LedgerClasses.Has(ctx, req.LedgerClassId)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, types.NewErrCodeNotFound("ledger_class")
	}

	if !k.IsLedgerClassMaintainer(ctx, req.Signer, req.LedgerClassId) {
		return nil, types.NewErrCodeUnauthorized("ledger class maintainer")
	}

	if err := k.AddClassEntryType(ctx, req.LedgerClassId, *req.EntryType); err != nil {
		return nil, err
	}

	return &types.MsgAddLedgerClassEntryTypeResponse{}, nil
}

// AddLedgerClassBucketType adds a bucket type to a ledger class.
func (k *MsgServer) AddLedgerClassBucketType(goCtx context.Context, req *types.MsgAddLedgerClassBucketTypeRequest) (*types.MsgAddLedgerClassBucketTypeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	has, err := k.LedgerClasses.Has(ctx, req.LedgerClassId)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, types.NewErrCodeNotFound("ledger_class")
	}

	if !k.IsLedgerClassMaintainer(ctx, req.Signer, req.LedgerClassId) {
		return nil, types.NewErrCodeUnauthorized("ledger class maintainer")
	}

	if err := k.AddClassBucketType(ctx, req.LedgerClassId, *req.BucketType); err != nil {
		return nil, err
	}

	return &types.MsgAddLedgerClassBucketTypeResponse{}, nil
}

// BulkCreate creates ledgers and entries in bulk.
func (k *MsgServer) BulkCreate(goCtx context.Context, req *types.MsgBulkCreateRequest) (*types.MsgBulkCreateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Signer has to be able to add ledgers and entries for every key.
	for _, ledgerToEntries := range req.LedgerAndEntries {
		if err := k.RequireAuthorization(ctx, req.Signer, ledgerToEntries.LedgerKey.ToRegistryKey()); err != nil {
			return nil, err
		}
	}

	if err := k.Keeper.BulkCreate(ctx, req.LedgerAndEntries); err != nil {
		return nil, err
	}

	return &types.MsgBulkCreateResponse{}, nil
}
