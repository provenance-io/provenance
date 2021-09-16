package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/provenance-io/provenance/x/marker/types"
)

var _ types.QueryServer = Keeper{}

// Params queries params of distribution module
func (k Keeper) Params(c context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	var params types.Params
	k.paramSpace.GetParamSet(ctx, &params)

	return &types.QueryParamsResponse{Params: params}, nil
}

// AllMarkers returns a list of all markers on the blockchain
func (k Keeper) AllMarkers(c context.Context, req *types.QueryAllMarkersRequest) (*types.QueryAllMarkersResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	markers := make([]*codectypes.Any, 0)
	store := ctx.KVStore(k.storeKey)
	markerStore := prefix.NewStore(store, types.MarkerStoreKeyPrefix)
	pageRes, err := query.Paginate(markerStore, req.Pagination, func(key []byte, value []byte) error {
		result, err := k.GetMarker(ctx, sdk.AccAddress(value))
		if err == nil {
			any, anyErr := codectypes.NewAnyWithValue(result)
			if anyErr != nil {
				return status.Errorf(codes.Internal, anyErr.Error())
			}
			markers = append(markers, any)
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	return &types.QueryAllMarkersResponse{Markers: markers, Pagination: pageRes}, nil
}

// Marker query for a single marker by denom or address
func (k Keeper) Marker(c context.Context, req *types.QueryMarkerRequest) (*types.QueryMarkerResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	marker, err := accountForDenomOrAddress(ctx, k, req.Id)
	if err != nil {
		return nil, err
	}
	any, err := codectypes.NewAnyWithValue(marker)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return &types.QueryMarkerResponse{Marker: any}, nil
}

// Holding query for all accounts holding the given marker coins
func (k Keeper) Holding(c context.Context, req *types.QueryHoldingRequest) (*types.QueryHoldingResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	marker, err := accountForDenomOrAddress(ctx, k, req.Id)
	if err != nil {
		return nil, err
	}

	denom := marker.GetDenom()
	balancesStore := prefix.NewStore(ctx.KVStore(k.bankKeeperStoreKey), banktypes.BalancesPrefix)
	var balances []types.Balance
	pageRes, perr := query.FilteredPaginate(balancesStore, req.Pagination, func(key []byte, value []byte, accumulate bool) (bool, error) {
		var coin sdk.Coin
		if cerr := k.cdc.Unmarshal(value, &coin); cerr != nil {
			return false, cerr
		}
		if coin.Denom != denom || coin.Amount.IsZero() {
			return false, nil
		}
		if accumulate {
			address, aerr := banktypes.AddressFromBalancesStore(key)
			if aerr != nil {
				k.Logger(ctx).With("key", key, "err", aerr).Error("failed to get address from balances store")
				return true, aerr
			}
			balances = append(balances,
				types.Balance{
					Address: address.String(),
					Coins:   sdk.NewCoins(coin),
				})
		}
		return true, nil
	})
	if perr != nil {
		return nil, perr
	}

	return &types.QueryHoldingResponse{
		Balances:   balances,
		Pagination: pageRes,
	}, nil
}

// Supply query for supply of coin on a marker account
func (k Keeper) Supply(c context.Context, req *types.QuerySupplyRequest) (*types.QuerySupplyResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	marker, err := accountForDenomOrAddress(ctx, k, req.Id)
	if err != nil {
		return nil, err
	}
	return &types.QuerySupplyResponse{Amount: marker.GetSupply()}, nil
}

// Escrow query for coins on a marker account
func (k Keeper) Escrow(c context.Context, req *types.QueryEscrowRequest) (*types.QueryEscrowResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	marker, err := accountForDenomOrAddress(ctx, k, req.Id)
	if err != nil {
		return nil, err
	}
	escrow := k.bankKeeper.GetAllBalances(ctx, marker.GetAddress())

	return &types.QueryEscrowResponse{Escrow: escrow}, nil
}

// Access query for access records on an account
func (k Keeper) Access(c context.Context, req *types.QueryAccessRequest) (*types.QueryAccessResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)
	marker, err := accountForDenomOrAddress(ctx, k, req.Id)
	if err != nil {
		return nil, err
	}
	return &types.QueryAccessResponse{Accounts: marker.GetAccessList()}, nil
}

// DenomMetadata query for metadata on denom
func (k Keeper) DenomMetadata(c context.Context, req *types.QueryDenomMetadataRequest) (*types.QueryDenomMetadataResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if err := sdk.ValidateDenom(req.Denom); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid denom")
	}

	ctx := sdk.UnwrapSDKContext(c)

	metadata, _ := k.bankKeeper.GetDenomMetaData(ctx, req.Denom)

	return &types.QueryDenomMetadataResponse{Metadata: metadata}, nil
}
