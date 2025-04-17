package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

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

	// Get all NFTs owned by the asset module
	moduleAddr := q.GetModuleAddress()
	if moduleAddr == nil {
		return nil, fmt.Errorf("asset module account not found")
	}

	// Get all NFT classes
	classes := q.nftKeeper.GetClasses(sdkCtx)

	// Collect all assets
	var assets []*types.Asset
	for _, class := range classes {
		// Get all NFTs in this class
		nfts := q.nftKeeper.GetNFTsOfClass(sdkCtx, class.Id)

		// Filter NFTs owned by the module
		for _, nft := range nfts {
			owner := q.nftKeeper.GetOwner(sdkCtx, class.Id, nft.Id)
			if owner.Equals(moduleAddr) {
				// Convert NFT to Asset
				asset := &types.Asset{
					ClassId: nft.ClassId,
					Id:      nft.Id,
					Uri:     nft.Uri,
					UriHash: nft.UriHash,
				}

				// If there's data, convert it to string
				if nft.Data != nil {
					// Try to extract string value from Any
					var strValue string
					if err := q.cdc.UnpackAny(nft.Data, &strValue); err == nil {
						asset.Data = strValue
					} else {
						// If we can't unpack as string, just use the raw data
						asset.Data = string(nft.Data.Value)
					}
				}

				assets = append(assets, asset)
			}
		}
	}

	return &types.QueryListAssetsResponse{
		Assets: assets,
	}, nil
}

// ListAssetClasses implements the Query/ListAssetClasses RPC method
func (q queryServer) ListAssetClasses(ctx context.Context, req *types.QueryListAssetClasses) (*types.QueryListAssetClassesResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Get all NFT classes
	classes := q.nftKeeper.GetClasses(sdkCtx)

	// Convert NFT classes to Asset classes
	var assetClasses []*types.AssetClass
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
			// Try to extract string value from Any
			var strValue string
			if err := q.cdc.UnpackAny(class.Data, &strValue); err == nil {
				assetClass.Data = strValue
			} else {
				// If we can't unpack as string, just use the raw data
				assetClass.Data = string(class.Data.Value)
			}
		}

		assetClasses = append(assetClasses, assetClass)
	}

	return &types.QueryListAssetClassesResponse{
		AssetClasses: assetClasses,
	}, nil
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

	// If there's data, convert it to string
	if nftClassResp.Data != nil {
		// Try to extract string value from Any
		var strValue string
		if err := q.cdc.UnpackAny(nftClassResp.Data, &strValue); err == nil {
			queryResp.AssetClass.Data = strValue
		} else {
			// If we can't unpack as string, just use the raw data
			queryResp.AssetClass.Data = string(nftClassResp.Data.Value)
		}
	}	

	return queryResp, nil
}
