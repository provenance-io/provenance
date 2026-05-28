//nolint:staticcheck // SA1019: quarantine grpc API is deprecated; retained for compatibility.
package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/provenance-io/provenance/x/quarantine"
)

var _ quarantine.QueryServer = Keeper{}

func (k Keeper) IsQuarantined(_ context.Context, _ *quarantine.QueryIsQuarantinedRequest) (*quarantine.QueryIsQuarantinedResponse, error) {
	return nil, status.Error(codes.Unimplemented, errQuarantineRemoved)
}

func (k Keeper) QuarantinedFunds(_ context.Context, _ *quarantine.QueryQuarantinedFundsRequest) (*quarantine.QueryQuarantinedFundsResponse, error) {
	return nil, status.Error(codes.Unimplemented, errQuarantineRemoved)
}

func (k Keeper) AutoResponses(_ context.Context, _ *quarantine.QueryAutoResponsesRequest) (*quarantine.QueryAutoResponsesResponse, error) {
	return nil, status.Error(codes.Unimplemented, errQuarantineRemoved)
}
