package keeper

import (
	"context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ledger "github.com/provenance-io/provenance/x/ledger/types"
)

var _ ledger.QueryServer = LedgerQueryServer{}

type LedgerQueryServer struct {
	k Keeper
}

func NewLedgerQueryServer(k Keeper) LedgerQueryServer {
	return LedgerQueryServer{
		k: k,
	}
}

func (qs LedgerQueryServer) Ledger(goCtx context.Context, req *ledger.QueryLedgerRequest) (*ledger.QueryLedgerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	l, err := qs.k.GetLedger(ctx, req.Key)
	if err != nil {
		return nil, err
	}

	if l == nil {
		return nil, ledger.NewLedgerCodedError(ledger.ErrCodeNotFound, "ledger")
	}

	resp := ledger.QueryLedgerResponse{
		Ledger: l,
	}

	return &resp, nil
}

func (qs LedgerQueryServer) LedgerEntries(goCtx context.Context, req *ledger.QueryLedgerEntriesRequest) (*ledger.QueryLedgerEntriesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	entries, err := qs.k.ListLedgerEntries(ctx, req.Key)
	if err != nil {
		return nil, err
	}

	if len(entries) == 0 {
		return nil, ledger.NewLedgerCodedError(ledger.ErrCodeNotFound, "ledger")
	}

	resp := ledger.QueryLedgerEntriesResponse{}

	// Add entries to the response.
	for _, entry := range entries {
		resp.Entries = append(resp.Entries, entry)
	}

	return &resp, nil
}

// GetBalancesAsOf returns the balances for a specific NFT as of a given date
func (qs LedgerQueryServer) LedgerBalancesAsOf(ctx context.Context, req *ledger.QueryLedgerBalancesAsOfRequest) (*ledger.QueryLedgerBalancesAsOfResponse, error) {
	if req == nil {
		return nil, ledger.NewLedgerCodedError(ledger.ErrCodeInvalidField, "request", "request is nil")
	}

	if !qs.k.HasLedger(sdk.UnwrapSDKContext(ctx), req.Key) {
		return nil, ledger.NewLedgerCodedError(ledger.ErrCodeNotFound, "ledger")
	}

	// Parse the date string
	asOfDate, err := time.Parse("2006-01-02", req.AsOfDate)
	if err != nil {
		return nil, ledger.NewLedgerCodedError(ledger.ErrCodeInvalidField, "as-of-date", "invalid date format")
	}

	balances, err := qs.k.GetBalancesAsOf(ctx, req.Key, asOfDate)
	if err != nil {
		return nil, err
	}

	if balances == nil {
		return nil, ledger.NewLedgerCodedError(ledger.ErrCodeNotFound, "balances")
	}

	return &ledger.QueryLedgerBalancesAsOfResponse{
		BucketBalances: balances,
	}, nil
}

// GetLedgerEntry returns a specific ledger entry for an NFT
func (qs LedgerQueryServer) LedgerEntry(ctx context.Context, req *ledger.QueryLedgerEntryRequest) (*ledger.QueryLedgerEntryResponse, error) {
	if req == nil {
		return nil, ledger.NewLedgerCodedError(ledger.ErrCodeInvalidField, "request", "request is nil")
	}

	if !qs.k.HasLedger(sdk.UnwrapSDKContext(ctx), req.Key) {
		return nil, ledger.NewLedgerCodedError(ledger.ErrCodeNotFound, "ledger")
	}

	entry, err := qs.k.GetLedgerEntry(ctx, req.Key, req.CorrelationId)
	if err != nil {
		return nil, err
	}

	if entry == nil {
		return nil, ledger.NewLedgerCodedError(ledger.ErrCodeNotFound, "ledger entry")
	}

	return &ledger.QueryLedgerEntryResponse{
		Entry: entry,
	}, nil
}

func (qs LedgerQueryServer) LedgerClassEntryTypes(ctx context.Context, req *ledger.QueryLedgerClassEntryTypesRequest) (*ledger.QueryLedgerClassEntryTypesResponse, error) {
	if req == nil {
		return nil, ledger.NewLedgerCodedError(ledger.ErrCodeInvalidField, "request", "request is nil")
	}

	types, err := qs.k.GetLedgerClassEntryTypes(ctx, req.LedgerClassId)
	if err != nil {
		return nil, err
	}

	return &ledger.QueryLedgerClassEntryTypesResponse{
		EntryTypes: types,
	}, nil
}

func (qs LedgerQueryServer) LedgerClassStatusTypes(ctx context.Context, req *ledger.QueryLedgerClassStatusTypesRequest) (*ledger.QueryLedgerClassStatusTypesResponse, error) {
	if req == nil {
		return nil, ledger.NewLedgerCodedError(ledger.ErrCodeInvalidField, "request", "request is nil")
	}

	types, err := qs.k.GetLedgerClassStatusTypes(ctx, req.LedgerClassId)
	if err != nil {
		return nil, err
	}

	return &ledger.QueryLedgerClassStatusTypesResponse{
		StatusTypes: types,
	}, nil
}

func (qs LedgerQueryServer) LedgerClassBucketTypes(ctx context.Context, req *ledger.QueryLedgerClassBucketTypesRequest) (*ledger.QueryLedgerClassBucketTypesResponse, error) {
	if req == nil {
		return nil, ledger.NewLedgerCodedError(ledger.ErrCodeInvalidField, "request", "request is nil")
	}

	types, err := qs.k.GetLedgerClassBucketTypes(ctx, req.LedgerClassId)
	if err != nil {
		return nil, err
	}

	return &ledger.QueryLedgerClassBucketTypesResponse{
		BucketTypes: types,
	}, nil
}

func (qs LedgerQueryServer) LedgerClass(ctx context.Context, req *ledger.QueryLedgerClassRequest) (*ledger.QueryLedgerClassResponse, error) {
	if req == nil {
		return nil, ledger.NewLedgerCodedError(ledger.ErrCodeInvalidField, "request", "request is nil")
	}

	ledgerClass, err := qs.k.GetLedgerClass(ctx, req.LedgerClassId)
	if err != nil {
		return nil, err
	}

	return &ledger.QueryLedgerClassResponse{
		LedgerClass: ledgerClass,
	}, nil
}

func (qs LedgerQueryServer) LedgerSettlements(ctx context.Context, req *ledger.QueryLedgerSettlementsRequest) (*ledger.QueryLedgerSettlementsResponse, error) {
	if req == nil {
		return nil, ledger.NewLedgerCodedError(ledger.ErrCodeInvalidField, "request", "request is nil")
	}

	// convert the ledger key to a string
	keyStr := req.Key.String()

	settlements, err := qs.k.GetAllSettlements(ctx, &keyStr)
	if err != nil {
		return nil, err
	}

	return &ledger.QueryLedgerSettlementsResponse{
		Settlements: settlements,
	}, nil
}

func (qs LedgerQueryServer) LedgerSettlementsByCorrelationId(ctx context.Context, req *ledger.QueryLedgerSettlementsByCorrelationIdRequest) (*ledger.QueryLedgerSettlementsByCorrelationIdResponse, error) {
	if req == nil {
		return nil, ledger.NewLedgerCodedError(ledger.ErrCodeInvalidField, "request", "request is nil")
	}

	// convert the ledger key to a string
	keyStr := req.Key.String()

	settlements, err := qs.k.GetSettlements(ctx, &keyStr, req.CorrelationId)
	if err != nil {
		return nil, err
	}

	return &ledger.QueryLedgerSettlementsByCorrelationIdResponse{
		Settlements: settlements,
	}, nil
}
