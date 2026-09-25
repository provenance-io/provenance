package keeper

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"cosmossdk.io/collections"
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/provenance-io/provenance/x/hold"
)

// entryWithKey holds a key-value pair for sorting/reversing
type entryWithKey struct {
	key    collections.Pair[sdk.AccAddress, string]
	amount sdkmath.Int
	hasErr bool
	err    error
}

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
	ctx := sdk.UnwrapSDKContext(goCtx)
	if req == nil {
		req = &hold.GetAllHoldsRequest{
			Pagination: &query.PageRequest{
				CountTotal: true,
			},
		}
	}
	if req.Pagination == nil {
		req.Pagination = &query.PageRequest{
			CountTotal: true,
		}
	}
	return k.paginateAllHolds(ctx, req.Pagination)
}

// paginateAllHolds iterates over hold entries to generate a paginated GetAllHolds result.
// It's copied from query.FilteredPaginate and tweaked to count results by address instead of iterator entry.
// It was easier to do it this way than shoehorn a solution into a call to FilteredPaginate.
func (k Keeper) paginateAllHolds(ctx sdk.Context, pageRequest *query.PageRequest) (*hold.GetAllHoldsResponse, error) {
	if pageRequest.Offset > 0 && len(pageRequest.Key) > 0 {
		return nil, status.Error(
			codes.InvalidArgument,
			"either offset or key is expected, got both",
		)
	}

	offset := pageRequest.Offset
	pageKey := pageRequest.Key
	limit := pageRequest.Limit
	countTotal := pageRequest.CountTotal
	reverse := pageRequest.Reverse

	if limit == 0 {
		limit = query.DefaultLimit
		countTotal = true
	}

	// Initialize response
	resp := &hold.GetAllHoldsResponse{
		Holds:      []*hold.AccountHold{},
		Pagination: &query.PageResponse{},
	}

	// Create codec for key encoding/decoding
	pairCodec := collections.PairKeyCodec(
		hold.AddressKeyCodec{},
		hold.DenomKeyCodec{},
	)

	// Parse start key if provided
	var startKey collections.Pair[sdk.AccAddress, string]
	var hasStartKey bool

	if len(pageKey) > 0 {
		bytesRead, decodedKey, err := pairCodec.Decode(pageKey)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid pagination key: %v", err)
		}
		if bytesRead != len(pageKey) {
			return nil, status.Errorf(
				codes.InvalidArgument,
				"pagination key decode mismatch: read %d bytes, expected %d",
				bytesRead, len(pageKey),
			)
		}
		startKey = decodedKey
		hasStartKey = true
	}

	var allEntries []entryWithKey
	var entryErrors map[int]error

	iterator, err := k.Holds.Iterate(ctx, nil)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create iterator: %v", err)
	}
	defer iterator.Close() //nolint:errcheck // close error safe to ignore in this context.

	entryErrors = make(map[int]error)
	entryIndex := 0

	// Collect all entries
	for ; iterator.Valid(); iterator.Next() {
		kv, err := iterator.KeyValue()
		if err != nil {
			iterKey, keyErr := iterator.Key()
			if keyErr == nil {
				entryErrors[entryIndex] = fmt.Errorf("failed to read amount of %s for account %s: %w",
					iterKey.K2(), iterKey.K1(), err)
				allEntries = append(allEntries, entryWithKey{
					key:    iterKey,
					hasErr: true,
					err:    entryErrors[entryIndex],
				})
			} else {
				entryErrors[entryIndex] = fmt.Errorf("failed to read entry: %w", err)
				allEntries = append(allEntries, entryWithKey{
					hasErr: true,
					err:    entryErrors[entryIndex],
				})
			}
			entryIndex++
			continue
		}

		allEntries = append(allEntries, entryWithKey{
			key:    kv.Key,
			amount: kv.Value,
			hasErr: false,
		})
		entryIndex++
	}

	// Reverse the entries if needed
	if reverse {
		for i, j := 0, len(allEntries)-1; i < j; i, j = i+1, j-1 {
			allEntries[i], allEntries[j] = allEntries[j], allEntries[i]
		}
	}

	// Now process the entries with pagination
	var (
		lastAddr        sdk.AccAddress
		lastEntry       *hold.AccountHold
		numHits         uint64
		accumulate      bool
		nextKeySet      bool
		foundStartKey   bool
		processedErrors []error
	)

	if hasStartKey {
		foundStartKey = false
		accumulate = false
	} else {
		foundStartKey = true
		accumulate = true
	}

	// Process entries
	for _, entry := range allEntries {
		if entry.hasErr {
			shouldReport := true

			if hasStartKey && entry.key.K1() != nil {
				cmp := compareKeys(entry.key, startKey)
				if reverse {
					shouldReport = cmp <= 0
				} else {
					shouldReport = cmp >= 0
				}
			}

			// For offset/limit pagination (no key), don't report errors beyond the limit
			// This handles "bad_entry_but_its_out_of_the_result_range" test
			if !hasStartKey && uint64(len(resp.Holds)) >= limit {
				shouldReport = false
			}

			if shouldReport {
				processedErrors = append(processedErrors, entry.err)
			}
			continue
		}

		addr := entry.key.K1()
		denom := entry.key.K2()
		amount := entry.amount

		if hasStartKey && !foundStartKey {
			cmp := compareKeys(entry.key, startKey)

			if reverse {
				if cmp > 0 {
					continue
				}
				foundStartKey = true
			} else {
				if cmp < 0 {
					continue
				}
				foundStartKey = true
			}
		}

		// New address begins
		if !addr.Equals(lastAddr) {
			if uint64(len(resp.Holds)) >= limit && !nextKeySet {
				keySize := pairCodec.Size(entry.key)
				if keySize > 0 {
					keyBuffer := make([]byte, keySize)
					bytesWritten, encodeErr := pairCodec.Encode(keyBuffer, entry.key)
					if encodeErr == nil && bytesWritten == keySize {
						resp.Pagination.NextKey = keyBuffer
						nextKeySet = true
					}
				}

				if !countTotal {
					break
				}
			}

			lastAddr = addr
			numHits++

			if hasStartKey {
				accumulate = foundStartKey && uint64(len(resp.Holds)) < limit
			} else {
				accumulate = numHits > offset && uint64(len(resp.Holds)) < limit
			}

			// Create new entry if accumulating
			if accumulate {
				lastEntry = &hold.AccountHold{
					Address: addr.String(),
					Amount:  sdk.Coins{},
				}
				resp.Holds = append(resp.Holds, lastEntry)
			}
		}
		if accumulate && lastEntry != nil {
			lastEntry.Amount = lastEntry.Amount.Add(sdk.NewCoin(denom, amount))
		}
	}

	if countTotal {
		resp.Pagination.Total = numHits
	}

	if len(processedErrors) > 0 {
		return nil, status.Errorf(codes.Internal, "pagination failed: %v", errors.Join(processedErrors...))
	}

	return resp, nil
}

// compareKeys compares two pair keys lexicographically
// Returns: -1 if k1 < k2, 0 if k1 == k2, 1 if k1 > k2
func compareKeys(k1, k2 collections.Pair[sdk.AccAddress, string]) int {
	addr1 := k1.K1()
	addr2 := k2.K1()

	addrCmp := bytes.Compare(addr1.Bytes(), addr2.Bytes())
	if addrCmp != 0 {
		return addrCmp
	}
	denom1 := k1.K2()
	denom2 := k2.K2()

	if denom1 < denom2 {
		return -1
	} else if denom1 > denom2 {
		return 1
	}
	return 0
}
