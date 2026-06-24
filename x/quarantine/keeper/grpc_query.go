package keeper

import (
	"bytes"
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"cosmossdk.io/collections"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/provenance-io/provenance/x/quarantine"
)

var _ quarantine.QueryServer = Keeper{}

func (k Keeper) IsQuarantined(goCtx context.Context, req *quarantine.QueryIsQuarantinedRequest) (*quarantine.QueryIsQuarantinedResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if len(req.ToAddress) == 0 {
		return nil, status.Error(codes.InvalidArgument, "to address cannot be empty")
	}

	toAddr, err := sdk.AccAddressFromBech32(req.ToAddress)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid to address: %s", err.Error())
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	resp := &quarantine.QueryIsQuarantinedResponse{
		IsQuarantined: k.IsQuarantinedAddr(ctx, toAddr),
	}

	return resp, nil
}

func (k Keeper) QuarantinedFunds(goCtx context.Context, req *quarantine.QueryQuarantinedFundsRequest) (*quarantine.QueryQuarantinedFundsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if len(req.FromAddress) > 0 && len(req.ToAddress) == 0 {
		return nil, status.Error(codes.InvalidArgument, "to address cannot be empty when from address is not")
	}

	var toAddr, fromAddr sdk.AccAddress
	var err error
	if len(req.ToAddress) > 0 {
		toAddr, err = sdk.AccAddressFromBech32(req.ToAddress)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid to address: %s", err.Error())
		}
	}
	if len(req.FromAddress) > 0 {
		fromAddr, err = sdk.AccAddressFromBech32(req.FromAddress)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid from address: %s", err.Error())
		}
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	resp := &quarantine.QueryQuarantinedFundsResponse{}

	if len(fromAddr) > 0 {
		// Not paginating here because it's assumed that there are few results.
		// Also, there's no way to use query.FilteredPaginate to iterate over just these specific entries.
		// So it'd be doing a lot of extra unneeded work.
		for _, qr := range k.GetQuarantineRecords(ctx, toAddr, fromAddr) {
			resp.QuarantinedFunds = append(resp.QuarantinedFunds, qr.AsQuarantinedFunds(toAddr))
		}
	} else {
		resp.Pagination, err = k.filteredRecordPaginate(
			ctx, toAddr, req.Pagination,
			func(kToAddr sdk.AccAddress, record *quarantine.QuarantineRecord) {
				resp.QuarantinedFunds = append(resp.QuarantinedFunds, record.AsQuarantinedFunds(kToAddr))
			},
		)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return resp, nil
}

func (k Keeper) AutoResponses(goCtx context.Context, req *quarantine.QueryAutoResponsesRequest) (*quarantine.QueryAutoResponsesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if len(req.ToAddress) == 0 {
		return nil, status.Error(codes.InvalidArgument, "to address cannot be empty")
	}

	toAddr, err := sdk.AccAddressFromBech32(req.ToAddress)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid to address: %s", err.Error())
	}

	var fromAddr sdk.AccAddress
	if len(req.FromAddress) > 0 {
		fromAddr, err = sdk.AccAddressFromBech32(req.FromAddress)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid from address: %s", err.Error())
		}
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	resp := &quarantine.QueryAutoResponsesResponse{}

	if len(fromAddr) > 0 {
		qar := k.GetAutoResponse(ctx, toAddr, fromAddr)
		resp.AutoResponses = append(resp.AutoResponses, quarantine.NewAutoResponseEntry(toAddr, fromAddr, qar))
	} else {
		// We use query.CollectionPaginate so that key-based (NextKey) cursor
		// pagination works correctly in addition to offset-based pagination.
		ptrs, pageRes, pErr := query.CollectionPaginate(
			ctx,
			k.autoResponses,
			req.Pagination,
			func(
				key collections.Pair[sdk.AccAddress, sdk.AccAddress],
				response quarantine.AutoResponse,
			) (*quarantine.AutoResponseEntry, error) {
				return quarantine.NewAutoResponseEntry(key.K1(), key.K2(), response), nil
			},
			query.WithCollectionPaginationPairPrefix[sdk.AccAddress, sdk.AccAddress](toAddr),
		)
		if pErr != nil {
			return nil, status.Error(codes.Internal, pErr.Error())
		}
		// query.CollectionPaginate returns nil *PageResponse when CountTotal is
		// false (the default).  The old query.Paginate always returned a non-nil
		// &PageResponse{},
		if pageRes == nil {
			pageRes = &query.PageResponse{}
		}
		resp.Pagination = pageRes
		for _, p := range ptrs {
			if p != nil {
				resp.AutoResponses = append(resp.AutoResponses, p)
			}
		}
	}

	return resp, nil
}

// filteredRecordPaginate iterates the records collection using offset-based
// pagination, skipping declined records.
func (k Keeper) filteredRecordPaginate(
	ctx sdk.Context,
	toAddr sdk.AccAddress,
	pageReq *query.PageRequest,
	onResult func(toAddr sdk.AccAddress, record *quarantine.QuarantineRecord),
) (*query.PageResponse, error) {
	if pageReq == nil {
		pageReq = &query.PageRequest{CountTotal: true}
	}
	limit := pageReq.Limit
	if limit == 0 {
		limit = query.DefaultLimit
	}

	offset := pageReq.Offset
	countTotal := pageReq.CountTotal
	reverse := pageReq.Reverse

	var ranger collections.Ranger[collections.Pair[sdk.AccAddress, sdk.AccAddress]]
	if reverse {
		rng := &collections.Range[collections.Pair[sdk.AccAddress, sdk.AccAddress]]{}
		rng.Descending()
		ranger = rng
	} else if len(toAddr) > 0 {
		ranger = collections.NewPrefixedPairRange[sdk.AccAddress, sdk.AccAddress](toAddr)
	}

	iter, err := k.records.Iterate(ctx, ranger)
	if err != nil {
		return nil, fmt.Errorf("quarantine: failed to open records iterator: %w", err)
	}

	defer iter.Close() //nolint:errcheck // ignoring close error on iterator: not critical for this context.

	hasCursor := len(pageReq.Key) > 0 && !reverse
	var (
		numSeen         uint64
		numAccum        uint64
		nextKey         []byte
		nextKeyCaptured bool
		cursorCleared   bool // flips to true the moment we reach the cursor.
	)

	for ; iter.Valid(); iter.Next() {
		kv, kvErr := iter.KeyValue()
		if kvErr != nil {
			return nil, fmt.Errorf("quarantine: failed to read record: %w", kvErr)
		}
		k1 := kv.Key.K1()
		if reverse && len(toAddr) > 0 {
			cmp := bytes.Compare(k1, toAddr)
			if cmp > 0 {
				continue
			}
			if cmp < 0 {
				break
			}
		}
		if hasCursor && !cursorCleared {
			var encoded []byte
			if len(toAddr) > 0 {
				encoded = address.MustLengthPrefix(kv.Key.K2())
			} else {
				encoded = quarantine.CreateRecordKey(k1, kv.Key.K2())[1:]
			}
			if bytes.Compare(encoded, pageReq.Key) < 0 {
				continue
			}
			cursorCleared = true
		}
		if kv.Value.Declined {
			continue
		}
		numSeen++
		if numSeen <= offset {
			continue
		}
		if numAccum < limit {
			record := kv.Value
			onResult(k1, &record)
			numAccum++
		} else {
			if !nextKeyCaptured {
				if len(toAddr) > 0 {
					nextKey = address.MustLengthPrefix(kv.Key.K2())
				} else {
					nextKey = quarantine.CreateRecordKey(k1, kv.Key.K2())[1:]
				}
				nextKeyCaptured = true
			}
			if !countTotal {
				break
			}
		}
	}

	pageRes := &query.PageResponse{NextKey: nextKey}
	if countTotal {
		pageRes.Total = numSeen
	}

	return pageRes, nil
}
