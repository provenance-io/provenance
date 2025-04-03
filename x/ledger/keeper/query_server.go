package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/ledger"
)

var _ ledger.QueryServer = Keeper{}

func (k Keeper) Config(goCtx context.Context, req *ledger.QueryLedgerConfigRequest) (*ledger.QueryLedgerConfigResponse, error) {
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

func (k Keeper) Entries(context.Context, *ledger.QueryLedgerRequest) (*ledger.QueryLedgerResponse, error) {
	return nil, nil
}
