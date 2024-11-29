package keeper

import (
	"context"

	"github.com/provenance-io/provenance/x/nav"
)

type QueryServer struct {
	Keeper
}

func NewQueryServer(k Keeper) nav.QueryServer {
	return QueryServer{Keeper: k}
}

// GetNAV returns the single Net Asset Value entry requested.
func (q QueryServer) GetNAV(context.Context, *nav.QueryGetNAVRequest) (*nav.QueryGetNAVResponse, error) {
	panic(nav.NotYetImplemented)
}

// GetAllNAVs returns a page of all Net Asset Value entries, possibly limited
// to a single asset denom.
func (q QueryServer) GetAllNAVs(context.Context, *nav.QueryGetAllNAVsRequest) (*nav.QueryGetAllNAVsResponse, error) {
	panic(nav.NotYetImplemented)
}
