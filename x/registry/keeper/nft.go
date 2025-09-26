package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/registry/types"
)

// HasNFT checks if an NFT exists in either the metadata or nft module.
// If the assetClassId is a metadata scope, it will check if the scope exists.
// Otherwise, it will check if the NFT exists in the nft module.
func (k Keeper) HasNFT(ctx context.Context, assetClassID, nftID *string) bool {
	metadataAddress, isMetadataScope := types.MetadataScopeID(*nftID)
	if isMetadataScope {
		sdkCtx := sdk.UnwrapSDKContext(ctx)
		_, found := k.MetadataKeeper.GetScope(sdkCtx, metadataAddress)
		return found
	}
	return k.NFTKeeper.HasNFT(ctx, *assetClassID, *nftID)
}

// AssetClassExists checks if an asset class exists in either the metadata or nft module.
func (k Keeper) AssetClassExists(ctx context.Context, assetClassID *string) bool {
	metadataAddress, isMetadataScope := types.MetadataScopeSpecID(*assetClassID)
	if isMetadataScope {
		sdkCtx := sdk.UnwrapSDKContext(ctx)
		_, found := k.MetadataKeeper.GetScopeSpecification(sdkCtx, metadataAddress)
		return found
	}
	return k.NFTKeeper.HasClass(ctx, *assetClassID)
}

// GetNFTOwner returns the owner of an NFT.
// If the assetClassId is a metadata scope, it will return the owner of the scope.
// Otherwise, it will return the owner of the NFT from the nft module.
func (k Keeper) GetNFTOwner(ctx context.Context, assetClassID, nftID *string) sdk.AccAddress {
	metadataAddress, isMetadataScope := types.MetadataScopeID(*nftID)
	if isMetadataScope {
		sdkCtx := sdk.UnwrapSDKContext(ctx)

		// Use the value owner address as the owner of the scope.
		accAddr, err := k.MetadataKeeper.GetScopeValueOwner(sdkCtx, metadataAddress)
		if err != nil {
			return nil
		}
		return accAddr
	}
	return k.NFTKeeper.GetOwner(ctx, *assetClassID, *nftID)
}
