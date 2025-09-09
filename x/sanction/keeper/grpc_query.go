package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/provenance-io/provenance/x/sanction"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ sanction.QueryServer = Keeper{}

// IsSanctioned checks if the given address is sanctioned.
func (k Keeper) IsSanctioned(goCtx context.Context, req *sanction.QueryIsSanctionedRequest) (*sanction.QueryIsSanctionedResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if len(req.Address) == 0 {
		return nil, status.Error(codes.InvalidArgument, "address cannot be empty")
	}

	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid address: %s", err.Error())
	}

	resp := &sanction.QueryIsSanctionedResponse{}
	resp.IsSanctioned = k.IsSanctionedAddr(goCtx, addr)
	return resp, nil
}

// SanctionedAddresses returns a list of sanctioned addresses.
func (k Keeper) SanctionedAddresses(goCtx context.Context, req *sanction.QuerySanctionedAddressesRequest) (*sanction.QuerySanctionedAddressesResponse, error) {
	var err error
	var pagination *query.PageRequest
	if req != nil {
		pagination = req.Pagination
	}

	resp := &sanction.QuerySanctionedAddressesResponse{}
	ctx := sdk.UnwrapSDKContext(goCtx)
	store := k.getSanctionedAddressPrefixStore(ctx)
	resp.Pagination, err = query.Paginate(
		store, pagination,
		func(key, _ []byte) error {
			addrBz, _ := ParseLengthPrefixedBz(key)
			resp.Addresses = append(resp.Addresses, sdk.AccAddress(addrBz).String())
			return nil
		},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return resp, nil
}

// TemporaryEntries returns the temporary sanction entries.
func (k Keeper) TemporaryEntries(goCtx context.Context, req *sanction.QueryTemporaryEntriesRequest) (*sanction.QueryTemporaryEntriesResponse, error) {
	var err error
	var pagination *query.PageRequest
	var addr sdk.AccAddress
	if req != nil {
		pagination = req.Pagination
		if len(req.Address) > 0 {
			addr, err = sdk.AccAddressFromBech32(req.Address)
			if err != nil {
				return nil, status.Errorf(codes.InvalidArgument, "invalid address: %s", err.Error())
			}
		}
	}

	resp := &sanction.QueryTemporaryEntriesResponse{}
	ctx := sdk.UnwrapSDKContext(goCtx)
	store, pre := k.getTemporaryEntryPrefixStore(ctx, addr)
	resp.Pagination, err = query.Paginate(
		store, pagination,
		func(key, value []byte) error {
			kAddr, propID := ParseTemporaryKey(ConcatBz(pre, key))
			entry := sanction.TemporaryEntry{
				Address:    kAddr.String(),
				ProposalId: propID,
				Status:     ToTempStatus(value),
			}
			resp.Entries = append(resp.Entries, &entry)
			return nil
		},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return resp, nil
}

// Params returns the current parameters of the sanction module.
func (k Keeper) Params(goCtx context.Context, _ *sanction.QueryParamsRequest) (*sanction.QueryParamsResponse, error) {
	resp := &sanction.QueryParamsResponse{}
	ctx := sdk.UnwrapSDKContext(goCtx)
	resp.Params = k.GetParams(ctx)
	return resp, nil
}
