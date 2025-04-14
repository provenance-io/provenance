package keeper

import (
	"context"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/ledger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	// Parse the date string
	asOfDate, err := time.Parse(time.RFC3339, req.AsOfDate)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid date format: %v", err))
	}

	balances, err := qs.k.GetBalancesAsOf(ctx, req.NftAddress, asOfDate)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &ledger.QueryBalancesAsOfResponse{
		Balances: balances,
	}, nil
}

// GetLedgerEntry returns a specific ledger entry for an NFT
func (qs LedgerQueryServer) GetLedgerEntry(ctx context.Context, req *ledger.QueryLedgerEntryRequest) (*ledger.QueryLedgerEntryResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	entry, err := qs.k.GetLedgerEntry(ctx, req.NftAddress, req.CorrelationId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &ledger.QueryLedgerEntryResponse{
		Entry: entry,
	}, nil
}
