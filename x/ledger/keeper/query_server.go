package keeper

import (
	"context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/ledger/types"
)

var _ types.QueryServer = LedgerQueryServer{}

type LedgerQueryServer struct {
	Keeper
}

func NewLedgerQueryServer(k Keeper) LedgerQueryServer {
	return LedgerQueryServer{
		Keeper: k,
	}
}

// LedgerClass returns the ledger class for a given ledger class id.
func (qs LedgerQueryServer) LedgerClass(ctx context.Context, req *types.QueryLedgerClassRequest) (*types.QueryLedgerClassResponse, error) {
	ledgerClass, err := qs.GetLedgerClass(ctx, req.LedgerClassId)
	if err != nil {
		return nil, err
	}

	return &types.QueryLedgerClassResponse{
		LedgerClass: ledgerClass,
	}, nil
}

// LedgerClasses returns a paginated list of all ledger classes.
func (qs LedgerQueryServer) LedgerClasses(ctx context.Context, req *types.QueryLedgerClassesRequest) (*types.QueryLedgerClassesResponse, error) {
	ledgerClasses, pageRes, err := qs.GetAllLedgerClasses(ctx, req.Pagination)
	if err != nil {
		return nil, err
	}

	return &types.QueryLedgerClassesResponse{
		LedgerClasses: ledgerClasses,
		Pagination:    pageRes,
	}, nil
}

// LedgerClassEntryTypes returns the entry types for a given ledger class id.
func (qs LedgerQueryServer) LedgerClassEntryTypes(ctx context.Context, req *types.QueryLedgerClassEntryTypesRequest) (*types.QueryLedgerClassEntryTypesResponse, error) {
	entryTypes, err := qs.GetLedgerClassEntryTypes(ctx, req.LedgerClassId)
	if err != nil {
		return nil, err
	}

	return &types.QueryLedgerClassEntryTypesResponse{
		EntryTypes: entryTypes,
	}, nil
}

// LedgerClassStatusTypes returns the status types for a given ledger class id.
func (qs LedgerQueryServer) LedgerClassStatusTypes(ctx context.Context, req *types.QueryLedgerClassStatusTypesRequest) (*types.QueryLedgerClassStatusTypesResponse, error) {
	statusTypes, err := qs.GetLedgerClassStatusTypes(ctx, req.LedgerClassId)
	if err != nil {
		return nil, err
	}

	return &types.QueryLedgerClassStatusTypesResponse{
		StatusTypes: statusTypes,
	}, nil
}

// LedgerClassBucketTypes returns the bucket types for a given ledger class id.
func (qs LedgerQueryServer) LedgerClassBucketTypes(ctx context.Context, req *types.QueryLedgerClassBucketTypesRequest) (*types.QueryLedgerClassBucketTypesResponse, error) {
	bucketTypes, err := qs.GetLedgerClassBucketTypes(ctx, req.LedgerClassId)
	if err != nil {
		return nil, err
	}

	return &types.QueryLedgerClassBucketTypesResponse{
		BucketTypes: bucketTypes,
	}, nil
}

// Ledger returns the ledger for a given ledger key.
func (qs LedgerQueryServer) Ledger(goCtx context.Context, req *types.QueryLedgerRequest) (*types.QueryLedgerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	l, err := qs.GetLedger(ctx, req.Key)
	if err != nil {
		return nil, err
	}

	if l == nil {
		return nil, types.NewErrCodeNotFound("ledger")
	}

	resp := types.QueryLedgerResponse{
		Ledger: l,
	}

	return &resp, nil
}

// Ledgers returns a paginated list of all ledgers.
func (qs LedgerQueryServer) Ledgers(ctx context.Context, req *types.QueryLedgersRequest) (*types.QueryLedgersResponse, error) {
	ledgers, pageRes, err := qs.GetAllLedgers(ctx, req.Pagination)
	if err != nil {
		return nil, err
	}

	return &types.QueryLedgersResponse{
		Ledgers:    ledgers,
		Pagination: pageRes,
	}, nil
}

// LedgerEntries returns the entries for a given ledger key.
func (qs LedgerQueryServer) LedgerEntries(goCtx context.Context, req *types.QueryLedgerEntriesRequest) (*types.QueryLedgerEntriesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	entries, err := qs.ListLedgerEntries(ctx, req.Key)
	if err != nil {
		return nil, err
	}

	if len(entries) == 0 {
		return nil, types.NewErrCodeNotFound("ledger entries")
	}

	resp := types.QueryLedgerEntriesResponse{}

	// Add entries to the response.
	resp.Entries = append(resp.Entries, entries...)

	return &resp, nil
}

// LedgerEntry returns a specific ledger entry for an NFT.
func (qs LedgerQueryServer) LedgerEntry(ctx context.Context, req *types.QueryLedgerEntryRequest) (*types.QueryLedgerEntryResponse, error) {
	if !qs.HasLedger(sdk.UnwrapSDKContext(ctx), req.Key) {
		return nil, types.NewErrCodeNotFound("ledger")
	}

	entry, err := qs.GetLedgerEntry(ctx, req.Key, req.CorrelationId)
	if err != nil {
		return nil, err
	}

	if entry == nil {
		return nil, types.NewErrCodeNotFound("ledger entry")
	}

	return &types.QueryLedgerEntryResponse{
		Entry: entry,
	}, nil
}

// LedgerBalancesAsOf returns the balances for a specific NFT as of a given date.
func (qs LedgerQueryServer) LedgerBalancesAsOf(ctx context.Context, req *types.QueryLedgerBalancesAsOfRequest) (*types.QueryLedgerBalancesAsOfResponse, error) {
	if !qs.HasLedger(sdk.UnwrapSDKContext(ctx), req.Key) {
		return nil, types.NewErrCodeNotFound("ledger")
	}

	// Parse the date string
	asOfDate, err := time.Parse("2006-01-02", req.AsOfDate)
	if err != nil {
		return nil, types.NewErrCodeInvalidField("as-of-date", "invalid date format")
	}

	balances, err := qs.GetBalancesAsOf(ctx, req.Key, asOfDate)
	if err != nil {
		return nil, err
	}

	if balances == nil {
		return nil, types.NewErrCodeNotFound("balances")
	}

	return &types.QueryLedgerBalancesAsOfResponse{
		BucketBalances: balances,
	}, nil
}

// LedgerSettlements returns all settlements for a ledger.
func (qs LedgerQueryServer) LedgerSettlements(ctx context.Context, req *types.QueryLedgerSettlementsRequest) (*types.QueryLedgerSettlementsResponse, error) {
	// convert the ledger key to a string
	keyStr := req.Key.String()

	settlementLists, err := qs.GetAllSettlements(ctx, &keyStr)
	if err != nil {
		return nil, err
	}

	rv := &types.QueryLedgerSettlementsResponse{}
	for _, settlements := range settlementLists {
		rv.Settlements = append(rv.Settlements, settlements.SettlementInstructions...)
	}

	return rv, nil
}

// LedgerSettlementsByCorrelationID returns settlements by correlation id.
func (qs LedgerQueryServer) LedgerSettlementsByCorrelationID(ctx context.Context, req *types.QueryLedgerSettlementsByCorrelationIDRequest) (*types.QueryLedgerSettlementsByCorrelationIDResponse, error) {
	// convert the ledger key to a string
	keyStr := req.Key.String()

	settlements, err := qs.GetSettlements(ctx, &keyStr, req.CorrelationId)
	if err != nil {
		return nil, err
	}

	if settlements == nil {
		return nil, types.NewErrCodeNotFound("settlement instructions")
	}

	return &types.QueryLedgerSettlementsByCorrelationIDResponse{
		Settlements: settlements.SettlementInstructions,
	}, nil
}
