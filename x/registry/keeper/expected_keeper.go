package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/metadata/types"
)

type MetaDataKeeper interface {
	GetScopeSpecification(ctx sdk.Context, scopeSpecID types.MetadataAddress) (spec types.ScopeSpecification, found bool)
	GetScope(ctx sdk.Context, id types.MetadataAddress) (types.Scope, bool)
	GetScopeValueOwner(ctx sdk.Context, id types.MetadataAddress) (sdk.AccAddress, error)
}

type NFTKeeper interface {
	// GetOwner returns the owner of an NFT
	GetOwner(ctx context.Context, classID, nftID string) sdk.AccAddress

	// HasNFT checks if an NFT exists
	HasNFT(ctx context.Context, classID, id string) bool

	// HasClass checks if an NFT class exists
	HasClass(ctx context.Context, classID string) bool
}
