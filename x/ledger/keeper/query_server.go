package keeper

import (
	"context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/ledger/types"
)

var _ types.QueryServer = LedgerQueryServer{}

type LedgerQueryServer struct {
	k Keeper
}

func NewLedgerQueryServer(k Keeper) LedgerQueryServer {
	return LedgerQueryServer{
		k: k,
	}
}

func (qs LedgerQueryServer) Ledger(goCtx context.Context, req *types.QueryLedgerRequest) (*types.QueryLedgerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	l, err := qs.k.GetLedger(ctx, req.Key)
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

func (qs LedgerQueryServer) LedgerEntries(goCtx context.Context, req *types.QueryLedgerEntriesRequest) (*types.QueryLedgerEntriesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	entries, err := qs.k.ListLedgerEntries(ctx, req.Key)
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

// GetBalancesAsOf returns the balances for a specific NFT as of a given date
func (qs LedgerQueryServer) LedgerBalancesAsOf(ctx context.Context, req *types.QueryLedgerBalancesAsOfRequest) (*types.QueryLedgerBalancesAsOfResponse, error) {
	if !qs.k.HasLedger(sdk.UnwrapSDKContext(ctx), req.Key) {
		return nil, types.NewErrCodeNotFound("ledger")
	}

	// Parse the date string
	asOfDate, err := time.Parse("2006-01-02", req.AsOfDate)
	if err != nil {
		return nil, types.NewErrCodeInvalidField("as-of-date", "invalid date format")
	}

	balances, err := qs.k.GetBalancesAsOf(ctx, req.Key, asOfDate)
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

// GetLedgerEntry returns a specific ledger entry for an NFT
func (qs LedgerQueryServer) LedgerEntry(ctx context.Context, req *types.QueryLedgerEntryRequest) (*types.QueryLedgerEntryResponse, error) {
	if !qs.k.HasLedger(sdk.UnwrapSDKContext(ctx), req.Key) {
		return nil, types.NewErrCodeNotFound("ledger")
	}

	entry, err := qs.k.GetLedgerEntry(ctx, req.Key, req.CorrelationId)
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

func (qs LedgerQueryServer) LedgerClassEntryTypes(ctx context.Context, req *types.QueryLedgerClassEntryTypesRequest) (*types.QueryLedgerClassEntryTypesResponse, error) {
	entryTypes, err := qs.k.GetLedgerClassEntryTypes(ctx, req.LedgerClassId)
	if err != nil {
		return nil, err
	}

	return &types.QueryLedgerClassEntryTypesResponse{
		EntryTypes: entryTypes,
	}, nil
}

func (qs LedgerQueryServer) LedgerClassStatusTypes(ctx context.Context, req *types.QueryLedgerClassStatusTypesRequest) (*types.QueryLedgerClassStatusTypesResponse, error) {
	statusTypes, err := qs.k.GetLedgerClassStatusTypes(ctx, req.LedgerClassId)
	if err != nil {
		return nil, err
	}

	return &types.QueryLedgerClassStatusTypesResponse{
		StatusTypes: statusTypes,
	}, nil
}

func (qs LedgerQueryServer) LedgerClassBucketTypes(ctx context.Context, req *types.QueryLedgerClassBucketTypesRequest) (*types.QueryLedgerClassBucketTypesResponse, error) {
	bucketTypes, err := qs.k.GetLedgerClassBucketTypes(ctx, req.LedgerClassId)
	if err != nil {
		return nil, err
	}

	return &types.QueryLedgerClassBucketTypesResponse{
		BucketTypes: bucketTypes,
	}, nil
}

func (qs LedgerQueryServer) LedgerClass(ctx context.Context, req *types.QueryLedgerClassRequest) (*types.QueryLedgerClassResponse, error) {
	ledgerClass, err := qs.k.GetLedgerClass(ctx, req.LedgerClassId)
	if err != nil {
		return nil, err
	}

	return &types.QueryLedgerClassResponse{
		LedgerClass: ledgerClass,
	}, nil
}

func (qs LedgerQueryServer) LedgerSettlements(ctx context.Context, req *types.QueryLedgerSettlementsRequest) (*types.QueryLedgerSettlementsResponse, error) {
	// convert the ledger key to a string
	keyStr := req.Key.String()

	settlements, err := qs.k.GetAllSettlements(ctx, &keyStr)
	if err != nil {
		return nil, err
	}

	return &types.QueryLedgerSettlementsResponse{
		Settlements: settlements,
	}, nil
}

func (qs LedgerQueryServer) LedgerSettlementsByCorrelationId(ctx context.Context, req *types.QueryLedgerSettlementsByCorrelationIdRequest) (*types.QueryLedgerSettlementsByCorrelationIdResponse, error) {
	// convert the ledger key to a string
	keyStr := req.Key.String()

	settlements, err := qs.k.GetSettlements(ctx, &keyStr, req.CorrelationId)
	if err != nil {
		return nil, err
	}

	return &types.QueryLedgerSettlementsByCorrelationIdResponse{
		Settlements: settlements,
	}, nil
}
