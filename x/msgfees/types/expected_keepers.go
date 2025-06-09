package types

import (
	"context"

	flatfees "github.com/provenance-io/provenance/x/flatfees/types"
)

// FlatFeesQuerier has the query endpoints needed from the flat-fees module.
type FlatFeesQuerier interface {
	CalculateTxFees(ctx context.Context, req *flatfees.QueryCalculateTxFeesRequest) (*flatfees.QueryCalculateTxFeesResponse, error)
}
