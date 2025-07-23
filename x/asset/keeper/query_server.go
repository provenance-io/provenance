package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/provenance-io/provenance/x/asset/types"
)

type queryServer struct {
	Keeper
}

// NewQueryServerImpl returns an implementation of the asset QueryServer interface
func NewQueryServerImpl(keeper Keeper) types.QueryServer {
	return &queryServer{Keeper: keeper}
}

// ListAssets implements the Query/ListAssets RPC method
func (q queryServer) ListAssets(ctx context.Context, req *types.QueryListAssets) (*types.QueryListAssetsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Convert the address string to an sdk.AccAddress
	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, fmt.Errorf("invalid address: %w", err)
	}

	// Get pagination from request
	var pagination *query.PageRequest
	if req != nil {
		pagination = req.Pagination
	}

	// Get all NFT classes first
	classes := q.nftKeeper.GetClasses(sdkCtx)

	resp := &types.QueryListAssetsResponse{}
	var allAssets []*types.Asset
	var totalCount uint64

	// Collect all assets owned by the address
	for _, class := range classes {
		nfts := q.nftKeeper.GetNFTsOfClass(sdkCtx, class.Id)

		for _, nft := range nfts {
			owner := q.nftKeeper.GetOwner(sdkCtx, class.Id, nft.Id)
			if !owner.Equals(addr) {
				continue // Skip NFTs not owned by the address
			}

			totalCount++

			// Convert NFT to Asset
			asset := &types.Asset{
				ClassId: nft.ClassId,
				Id:      nft.Id,
				Uri:     nft.Uri,
				UriHash: nft.UriHash,
			}

			// If there's data, convert it to string
			if nft.Data != nil {
				strValue, err := types.AnyToString(q.cdc, nft.Data)
				if err == nil {
					asset.Data = strValue
				} else {
					continue // Skip NFTs with invalid data
				}
			}

			allAssets = append(allAssets, asset)
		}
	}

	// Apply pagination
	if pagination != nil {
		start := pagination.Offset
		end := start + pagination.Limit

		if start >= uint64(len(allAssets)) {
			start = uint64(len(allAssets))
		}
		if end > uint64(len(allAssets)) {
			end = uint64(len(allAssets))
		}

		resp.Assets = allAssets[start:end]
		resp.Pagination = &query.PageResponse{
			NextKey: nil, // No next key since we're not using key-based pagination
			Total:   totalCount,
		}
	} else {
		resp.Assets = allAssets
		resp.Pagination = &query.PageResponse{
			NextKey: nil,
			Total:   totalCount,
		}
	}

	return resp, nil
}

// ListAssetClasses implements the Query/ListAssetClasses RPC method
func (q queryServer) ListAssetClasses(ctx context.Context, req *types.QueryListAssetClasses) (*types.QueryListAssetClassesResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Get pagination from request
	var pagination *query.PageRequest
	if req != nil {
		pagination = req.Pagination
	}

	// Get all NFT classes
	classes := q.nftKeeper.GetClasses(sdkCtx)

	resp := &types.QueryListAssetClassesResponse{}

	// Convert NFT classes to Asset classes
	var allAssetClasses []*types.AssetClass
	for _, class := range classes {
		assetClass := &types.AssetClass{
			Id:          class.Id,
			Name:        class.Name,
			Symbol:      class.Symbol,
			Description: class.Description,
			Uri:         class.Uri,
			UriHash:     class.UriHash,
		}

		// If there's data, convert it to string
		if class.Data != nil {
			strValue, err := types.AnyToString(q.cdc, class.Data)
			if err == nil {
				assetClass.Data = strValue
			} else {
				assetClass.Data = "" // fallback to empty string if unpack fails
			}
		}

		allAssetClasses = append(allAssetClasses, assetClass)
	}

	// Apply pagination
	if pagination != nil {
		start := pagination.Offset
		end := start + pagination.Limit

		if start >= uint64(len(allAssetClasses)) {
			start = uint64(len(allAssetClasses))
		}
		if end > uint64(len(allAssetClasses)) {
			end = uint64(len(allAssetClasses))
		}

		resp.AssetClasses = allAssetClasses[start:end]
		resp.Pagination = &query.PageResponse{
			NextKey: nil, // No next key since we're not using key-based pagination
			Total:   uint64(len(allAssetClasses)),
		}
	} else {
		resp.AssetClasses = allAssetClasses
		resp.Pagination = &query.PageResponse{
			NextKey: nil,
			Total:   uint64(len(allAssetClasses)),
		}
	}

	return resp, nil
}

func (q queryServer) GetClass(ctx context.Context, req *types.QueryGetClass) (*types.QueryGetClassResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	nftClassResp, ok := q.nftKeeper.GetClass(sdkCtx, req.Id)
	if !ok {
		return nil, fmt.Errorf("class not found")
	}

	queryResp := &types.QueryGetClassResponse{
		AssetClass: &types.AssetClass{
			Id:          nftClassResp.Id,
			Name:        nftClassResp.Name,
			Symbol:      nftClassResp.Symbol,
			Description: nftClassResp.Description,
			Uri:         nftClassResp.Uri,
			UriHash:     nftClassResp.UriHash,
		},
	}

	if nftClassResp.Data != nil {
		dataString, err := types.AnyToString(q.cdc, nftClassResp.Data)
		if err != nil {
			return nil, fmt.Errorf("failed to convert Any to string: %w", err)
		}
		queryResp.AssetClass.Data = dataString
	}

	return queryResp, nil
}
