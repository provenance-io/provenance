package keeper

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	metadataTypes "github.com/provenance-io/provenance/x/metadata/types"
)

// HasNFT checks if an NFT exists in either the metadata or nft module.
// If the assetClassId is a metadata scope, it will check if the scope exists.
// Otherwise, it will check if the NFT exists in the nft module.
func (k BaseRegistryKeeper) HasNFT(ctx sdk.Context, assetClassId, nftId *string) bool {
	metadataAddress, isMetadataScope := metadataScopeID(*nftId)
	if isMetadataScope {
		sdkCtx := sdk.UnwrapSDKContext(ctx)
		_, found := k.MetaDataKeeper.GetScope(sdkCtx, metadataAddress)
		return found
	} else {
		return k.NFTKeeper.HasNFT(ctx, *assetClassId, *nftId)
	}
}

// AssetClassExists checks if an asset class exists in either the metadata or nft module.
func (k BaseRegistryKeeper) AssetClassExists(ctx sdk.Context, assetClassId *string) bool {
	metadataAddress, isMetadataScope := metadataScopeID(*assetClassId)
	if isMetadataScope {
		sdkCtx := sdk.UnwrapSDKContext(ctx)
		_, found := k.MetaDataKeeper.GetScopeSpecification(sdkCtx, metadataAddress)
		return found
	} else {
		return k.NFTKeeper.HasClass(ctx, *assetClassId)
	}
}

// GetNFTOwner returns the owner of an NFT.
// If the assetClassId is a metadata scope, it will return the owner of the scope.
// Otherwise, it will return the owner of the NFT from the nft module.
func (k BaseRegistryKeeper) GetNFTOwner(ctx sdk.Context, assetClassId, nftId *string) sdk.AccAddress {
	metadataAddress, isMetadataScope := metadataScopeID(*nftId)
	if isMetadataScope {
		sdkCtx := sdk.UnwrapSDKContext(ctx)

		scope, found := k.MetaDataKeeper.GetScope(sdkCtx, metadataAddress)
		if !found {
			return nil
		}

		// Use the value owner address as the owner of the scope.
		accAddr, err := sdk.AccAddressFromBech32(scope.ValueOwnerAddress)
		if err != nil {
			return nil
		}
		return accAddr
	} else {
		return k.NFTKeeper.GetOwner(ctx, *assetClassId, *nftId)
	}
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
