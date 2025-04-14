package keeper

import (
	"context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/ledger"
)

var _ ledger.QueryServer = LedgerQueryServer{}

type LedgerQueryServer struct {
	k ViewKeeper
}

func NewLedgerQueryServer(k ViewKeeper) LedgerQueryServer {
	return LedgerQueryServer{
		k: k,
	}
}

func (qs LedgerQueryServer) Config(goCtx context.Context, req *ledger.QueryLedgerConfigRequest) (*ledger.QueryLedgerConfigResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	l, err := qs.k.GetLedger(ctx, req.NftAddress)
	if err != nil {
		return nil, err
	}

	if l == nil {
		return nil, NewLedgerCodedError(ErrCodeNotFound, "ledger")
	}

	resp := ledger.QueryLedgerConfigResponse{
		Ledger: l,
	}

	return &resp, nil
}

func (qs LedgerQueryServer) Entries(goCtx context.Context, req *ledger.QueryLedgerRequest) (*ledger.QueryLedgerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	entries, err := qs.k.ListLedgerEntries(ctx, req.NftAddress)
	if err != nil {
		return nil, err
	}

	if len(entries) == 0 {
		return nil, NewLedgerCodedError(ErrCodeNotFound, "ledger")
	}

	resp := ledger.QueryLedgerResponse{}

	// Add entries to the response.
	for _, entry := range entries {
		resp.Entries = append(resp.Entries, &entry)
	}

	return &resp, nil
}

// GetBalancesAsOf returns the balances for a specific NFT as of a given date
func (qs LedgerQueryServer) GetBalancesAsOf(ctx context.Context, req *ledger.QueryBalancesAsOfRequest) (*ledger.QueryBalancesAsOfResponse, error) {
	if req == nil {
		return nil, NewLedgerCodedError(ErrCodeInvalidField, "request")
	}

	if !qs.k.HasLedger(sdk.UnwrapSDKContext(ctx), req.NftAddress) {
		return nil, NewLedgerCodedError(ErrCodeNotFound, "ledger")
	}

	// Parse the date string
	asOfDate, err := time.Parse("2006-01-02", req.AsOfDate)
	if err != nil {
		return nil, NewLedgerCodedError(ErrCodeInvalidField, "as-of-date")
	}

	balances, err := qs.k.GetBalancesAsOf(ctx, req.NftAddress, asOfDate)
	if err != nil {
		return nil, err
	}

	if balances == nil {
		return nil, NewLedgerCodedError(ErrCodeNotFound, "balances")
	}

	return &ledger.QueryBalancesAsOfResponse{
		Balances: balances,
	}, nil
}

// GetLedgerEntry returns a specific ledger entry for an NFT
func (qs LedgerQueryServer) GetLedgerEntry(ctx context.Context, req *ledger.QueryLedgerEntryRequest) (*ledger.QueryLedgerEntryResponse, error) {
	if req == nil {
		return nil, NewLedgerCodedError(ErrCodeInvalidField, "request")
	}

	if !qs.k.HasLedger(sdk.UnwrapSDKContext(ctx), req.NftAddress) {
		return nil, NewLedgerCodedError(ErrCodeNotFound, "ledger")
	}

	entry, err := qs.k.GetLedgerEntry(ctx, req.NftAddress, req.CorrelationId)
	if err != nil {
		return nil, err
	}

	if entry == nil {
		return nil, NewLedgerCodedError(ErrCodeNotFound, "ledger entry")
	}

	return &ledger.QueryLedgerEntryResponse{
		Entry: entry,
	}, nil
}
