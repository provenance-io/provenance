package keeper

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	db "github.com/tendermint/tm-db"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/provenance-io/provenance/x/hold"
)

// GetHolds looks up the funds that are on hold for an address.
func (k Keeper) GetHolds(goCtx context.Context, req *hold.GetHoldsRequest) (*hold.GetHoldsResponse, error) {
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

	ctx := sdk.UnwrapSDKContext(goCtx)
	resp := &hold.GetHoldsResponse{}
	resp.Amount, err = k.GetHoldCoins(ctx, addr)
	if err != nil {
		return nil, err
	}
	return resp, err
}

// GetAllHolds returns all addresses with funds on hold, and the amount held.
func (k Keeper) GetAllHolds(goCtx context.Context, req *hold.GetAllHoldsRequest) (*hold.GetAllHoldsResponse, error) {
	var pageReq *query.PageRequest
	if req != nil {
		pageReq = req.Pagination
	}

	return k.paginateAllHolds(sdk.UnwrapSDKContext(goCtx), pageReq)
}

// paginateAllHolds iterates over hold entries to generate a paginated GetAllHolds result.
// It's copied from query.FilteredPaginate and tweaked to count results by address instead of iterator entry.
// It was easier to do it this way than shoehorn a solution into a call to FilteredPaginate.
func (k Keeper) paginateAllHolds(ctx sdk.Context, pageRequest *query.PageRequest) (*hold.GetAllHoldsResponse, error) {
	// if the PageRequest is nil, use default PageRequest
	if pageRequest == nil {
		pageRequest = &query.PageRequest{}
	}

	offset := pageRequest.Offset
	key := pageRequest.Key
	limit := pageRequest.Limit
	countTotal := pageRequest.CountTotal
	reverse := pageRequest.Reverse

	if offset > 0 && key != nil {
		return nil, status.Errorf(codes.InvalidArgument, "either offset or key is expected, got both")
	}

	if limit == 0 {
		limit = query.DefaultLimit

		// count total results when the limit is zero/not supplied
		countTotal = true
	}

	var lastAddr sdk.AccAddress
	var lastEntry *hold.AccountHold
	resp := &hold.GetAllHoldsResponse{Pagination: &query.PageResponse{}}
	prefixStore := k.getAllHoldCoinPrefixStore(ctx)

	if len(key) != 0 {
		iterator := getIterator(prefixStore, key, reverse)
		defer iterator.Close()

		for ; iterator.Valid(); iterator.Next() {
			if err := iterator.Error(); err != nil {
				return nil, err
			}

			ikey := iterator.Key()
			addr, denom := ParseHoldCoinKeyUnprefixed(ikey)
			if !addr.Equals(lastAddr) {
				if uint64(len(resp.Holds)) >= limit {
					resp.Pagination.NextKey = ikey
					break
				}
				lastAddr = addr
				lastEntry = &hold.AccountHold{Address: addr.String()}
				resp.Holds = append(resp.Holds, lastEntry)
			}
			ival := iterator.Value()
			amount, err := UnmarshalHoldCoinValue(ival)
			if err != nil {
				return nil, fmt.Errorf("failed to read amount of %s for account %s: %w", denom, addr, err)
			}
			lastEntry.Amount = lastEntry.Amount.Add(sdk.Coin{Denom: denom, Amount: amount})
		}

		return resp, nil
	}

	iterator := getIterator(prefixStore, nil, reverse)
	defer iterator.Close()

	accumulate := false
	var numHits uint64

	for ; iterator.Valid(); iterator.Next() {
		if err := iterator.Error(); err != nil {
			return nil, err
		}

		ikey := iterator.Key()
		addr, denom := ParseHoldCoinKeyUnprefixed(ikey)
		if !addr.Equals(lastAddr) {
			if uint64(len(resp.Holds)) >= limit && len(resp.Pagination.NextKey) == 0 {
				resp.Pagination.NextKey = ikey
				if !countTotal {
					break
				}
			}
			lastAddr = addr

			numHits++
			accumulate = numHits > offset && uint64(len(resp.Holds)) < limit
			if accumulate {
				lastEntry = &hold.AccountHold{Address: addr.String()}
				resp.Holds = append(resp.Holds, lastEntry)
			}
		}

		if accumulate {
			ival := iterator.Value()
			amount, err := UnmarshalHoldCoinValue(ival)
			if err != nil {
				return nil, fmt.Errorf("failed to read amount of %s for account %s: %w", denom, addr, err)
			}
			lastEntry.Amount = lastEntry.Amount.Add(sdk.Coin{Denom: denom, Amount: amount})
		}
	}

	if countTotal {
		resp.Pagination.Total = numHits
	}

	return resp, nil
}

// getIterator creates an iterator on the provided store with the provided start and direction.
// It's copied from query.pagination.go.
func getIterator(prefixStore storetypes.KVStore, start []byte, reverse bool) db.Iterator {
	if reverse {
		var end []byte
		if start != nil {
			itr := prefixStore.Iterator(start, nil)
			defer itr.Close()
			if itr.Valid() {
				itr.Next()
				end = itr.Key()
			}
		}
		return prefixStore.ReverseIterator(nil, end)
	}
	return prefixStore.Iterator(start, nil)
}
