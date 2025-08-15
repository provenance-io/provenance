package keeper

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	metadataTypes "github.com/provenance-io/provenance/x/metadata/types"
)

// HasNFT checks if an NFT exists in either the metadata or nft module.
// If the assetClassId is a metadata scope, it will check if the scope exists.
// Otherwise, it will check if the NFT exists in the nft module.
func (k Keeper) HasNFT(ctx sdk.Context, assetClassID, nftID *string) bool {
	metadataAddress, isMetadataScope := metadataScopeID(*nftID)
	if isMetadataScope {
		sdkCtx := sdk.UnwrapSDKContext(ctx)
		_, found := k.MetadataKeeper.GetScope(sdkCtx, metadataAddress)
		return found
	}
	return k.NFTKeeper.HasNFT(ctx, *assetClassID, *nftID)
}

// AssetClassExists checks if an asset class exists in either the metadata or nft module.
func (k Keeper) AssetClassExists(ctx sdk.Context, assetClassID *string) bool {
	metadataAddress, isMetadataScope := metadataScopeID(*assetClassID)
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
func (k Keeper) GetNFTOwner(ctx sdk.Context, assetClassID, nftID *string) sdk.AccAddress {
	metadataAddress, isMetadataScope := metadataScopeID(*nftID)
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

// metadataScopeID returns the metadata address for a given bech32 string.
func metadataScopeID(bech32String string) (metadataTypes.MetadataAddress, bool) {
	// Do a bech32 decode if the prefix is "scope1" or "scopespec1"
	if strings.HasPrefix(bech32String, "scope1") || strings.HasPrefix(bech32String, "scopespec1") {
		metadataAddress, err := metadataTypes.MetadataAddressFromBech32(bech32String)
		if err != nil {
			return nil, false
		}
		return metadataAddress, true
	}
	return nil, false
}
