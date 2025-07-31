package keeper

import (
	"context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ledger "github.com/provenance-io/provenance/x/ledger/types"
)

var _ ledger.QueryServer = LedgerQueryServer{}

type LedgerQueryServer struct {
	k  ViewKeeper
	sk FundTransferKeeper
}

func NewLedgerQueryServer(k ViewKeeper, sk FundTransferKeeper) LedgerQueryServer {
	return LedgerQueryServer{
		k:  k,
		sk: sk,
	}
}

// Ledger returns the ledger for a specific ledger
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

// LedgerEntries returns the entries for a specific ledger
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

// BalancesAsOf returns the balances for a specific NFT as of a given date
func (qs LedgerQueryServer) BalancesAsOf(ctx context.Context, req *ledger.QueryBalancesAsOfRequest) (*ledger.QueryBalancesAsOfResponse, error) {
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

	return &ledger.QueryBalancesAsOfResponse{
		Balances: balances,
	}, nil
}

// LedgerEntry returns a specific ledger entry for an NFT
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

// LedgerClassEntryTypes returns the entry types for a specific ledger class
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

// LedgerClassStatusTypes returns the status types for a specific ledger class
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

// LedgerClassBucketTypes returns the bucket types for a specific ledger class
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

// LedgerClass returns the ledger class for a specific ledger class
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

// Settlements returns the settlements for a specific ledger
func (qs LedgerQueryServer) Settlements(ctx context.Context, req *ledger.QuerySettlementsRequest) (*ledger.QuerySettlementsResponse, error) {
	if req == nil {
		return nil, ledger.NewLedgerCodedError(ledger.ErrCodeInvalidField, "request", "request is nil")
	}

	// convert the ledger key to a string
	keyStr, err := LedgerKeyToString(req.Key)
	if err != nil {
		return nil, err
	}

	settlements, err := qs.sk.GetAllSettlements(ctx, keyStr)
	if err != nil {
		return nil, err
	}

	return &ledger.QuerySettlementsResponse{
		Settlements: settlements,
	}, nil
}

// SettlementsByCorrelationId returns the settlement for a specific correlation id
func (qs LedgerQueryServer) SettlementsByCorrelationId(ctx context.Context, req *ledger.QuerySettlementsByCorrelationIdRequest) (*ledger.QuerySettlementsByCorrelationIdResponse, error) {
	if req == nil {
		return nil, ledger.NewLedgerCodedError(ledger.ErrCodeInvalidField, "request", "request is nil")
	}

	// convert the ledger key to a string
	keyStr, err := LedgerKeyToString(req.Key)
	if err != nil {
		return nil, err
	}

	settlement, err := qs.sk.GetSettlements(ctx, keyStr, req.CorrelationId)
	if err != nil {
		return nil, err
	}

	return &ledger.QuerySettlementsByCorrelationIdResponse{
		Settlement: settlement,
	}, nil
}
