package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/ledger"
)

var _ ledger.QueryServer = LedgerKeeper{}

func (k LedgerKeeper) Config(goCtx context.Context, req *ledger.QueryLedgerConfigRequest) (*ledger.QueryLedgerConfigResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	l, err := k.GetLedger(ctx, req.NftAddress)
	if err != nil {
		return nil, err
	}

	resp := ledger.QueryLedgerConfigResponse{
		Ledger: l,
	}

	return &resp, nil
}

func (k LedgerKeeper) Entries(goCtx context.Context, req *ledger.QueryLedgerRequest) (*ledger.QueryLedgerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	entries, err := k.ListLedgerEntries(ctx, req.NftAddress)
	if err != nil {
		return nil, err
	}

	resp := ledger.QueryLedgerResponse{}

	for _, entry := range entries {
		resp.Entries = append(resp.Entries, &entry)
	}

	return &resp, nil
}
