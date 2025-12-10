package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"cosmossdk.io/collections"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/provenance-io/provenance/x/sanction"
)

var _ sanction.QueryServer = Keeper{}

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

func (k Keeper) SanctionedAddresses(goCtx context.Context, req *sanction.QuerySanctionedAddressesRequest) (*sanction.QuerySanctionedAddressesResponse, error) {
	var err error
	var pagination *query.PageRequest
	if req != nil {
		pagination = req.Pagination
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	resp := &sanction.QuerySanctionedAddressesResponse{
		Addresses: []string{},
	}

	// Use Collections pagination
	results, pageRes, err := query.CollectionPaginate(
		ctx,
		k.SanctionedAddressesStore,
		pagination,
		func(key sdk.AccAddress, value []byte) (string, error) {
			return key.String(), nil
		},
	)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp.Addresses = results
	resp.Pagination = pageRes

	return resp, nil
}

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
	ctx := sdk.UnwrapSDKContext(goCtx)

	if len(addr) > 0 {
		return k.temporaryEntriesForAddress(ctx, addr, pagination)
	}

	// Unfiltered query - use standard pagination
	results, pageRes, err := query.CollectionPaginate(
		ctx,
		k.TemporaryEntriesStore,
		pagination,
		func(key collections.Pair[sdk.AccAddress, uint64], value []byte) (*sanction.TemporaryEntry, error) {
			return &sanction.TemporaryEntry{
				Address:    key.K1().String(),
				ProposalId: key.K2(),
				Status:     ToTempStatus(value),
			}, nil
		},
	)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &sanction.QueryTemporaryEntriesResponse{
		Entries:    results,
		Pagination: pageRes,
	}, nil
}

// temporaryEntriesForAddress handles filtered queries for a specific address
func (k Keeper) temporaryEntriesForAddress(
	ctx sdk.Context,
	addr sdk.AccAddress,
	pagination *query.PageRequest,
) (*sanction.QueryTemporaryEntriesResponse, error) {
	limit := uint64(100)
	offset := uint64(0)
	reverse := false
	countTotal := false
	var startPropID *uint64

	if pagination != nil {
		if pagination.Limit > 0 {
			limit = pagination.Limit
		}
		reverse = pagination.Reverse
		countTotal = pagination.CountTotal
		if len(pagination.Key) == 8 {
			pid := sdk.BigEndianToUint64(pagination.Key)
			startPropID = &pid
			offset = 0
		} else {
			offset = pagination.Offset
		}
	}

	// Construct prefix matching TemporaryKeyCodec: <len><addr>
	prefixBytes := make([]byte, 1+len(addr))
	prefixBytes[0] = byte(len(addr))
	copy(prefixBytes[1:], addr)

	// Calculate prefix end
	prefixEnd := prefixEndBytes(prefixBytes)

	// Collect ALL matching entries first
	iter, err := k.TemporaryEntriesStore.IterateRaw(ctx, prefixBytes, prefixEnd, collections.OrderAscending)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer iter.Close()

	items := []*sanction.TemporaryEntry{} // never nil
	var total uint64

	for ; iter.Valid(); iter.Next() {
		kv, err := iter.KeyValue()
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		pid := kv.Key.K2()

		// :white_check_mark: FIX: Handle both forward and reverse key pagination
		if startPropID != nil {
			if !reverse && pid < *startPropID {
				// Forward mode: skip entries before startPropID
				if countTotal {
					total++
				}
				continue
			}
			if reverse && pid > *startPropID {
				// Reverse mode: skip entries after startPropID
				if countTotal {
					total++
				}
				continue
			}
		}

		if countTotal {
			total++
		}

		items = append(items, &sanction.TemporaryEntry{
			Address:    kv.Key.K1().String(),
			ProposalId: pid,
			Status:     ToTempStatus(kv.Value),
		})
	}

	// Apply reverse AFTER collection
	if reverse {
		for i, j := 0, len(items)-1; i < j; i, j = i+1, j-1 {
			items[i], items[j] = items[j], items[i]
		}
	}

	// Apply offset/limit AFTER collection
	start := int(offset)
	if start > len(items) {
		start = len(items)
	}
	end := start + int(limit)
	if end > len(items) {
		end = len(items)
	}
	pageItems := items[start:end]

	resp := &sanction.QueryTemporaryEntriesResponse{
		Entries: pageItems,
		Pagination: &query.PageResponse{
			Total: total,
		},
	}

	if end < len(items) {
		nextPropID := items[end].ProposalId
		resp.Pagination.NextKey = sdk.Uint64ToBigEndian(nextPropID)
	}

	return resp, nil
}

func (k Keeper) Params(goCtx context.Context, _ *sanction.QueryParamsRequest) (*sanction.QueryParamsResponse, error) {
	resp := &sanction.QueryParamsResponse{}
	ctx := sdk.UnwrapSDKContext(goCtx)
	resp.Params = k.GetParams(ctx)
	return resp, nil
}

// prefixEndBytes returns the []byte that would end a range query for all []byte with a certain prefix
func prefixEndBytes(prefix []byte) []byte {
	if len(prefix) == 0 {
		return nil
	}

	end := make([]byte, len(prefix))
	copy(end, prefix)

	for i := len(end) - 1; i >= 0; i-- {
		end[i]++
		if end[i] != 0 {
			return end
		}
	}

	return nil
}
