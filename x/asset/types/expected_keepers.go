package types

import (
	"context"

	"cosmossdk.io/x/nft"

	sdk "github.com/cosmos/cosmos-sdk/types"

	markertypes "github.com/provenance-io/provenance/x/marker/types"
	registrytypes "github.com/provenance-io/provenance/x/registry/types"
)

// MarkerKeeper defines the expected marker keeper interface.
type MarkerKeeper interface {
	AddFinalizeAndActivateMarker(ctx sdk.Context, marker markertypes.MarkerAccountI) error
	GetMarkerByDenom(ctx sdk.Context, denom string) (markertypes.MarkerAccountI, error)
	SetMarker(ctx sdk.Context, marker markertypes.MarkerAccountI)
}

// NFTKeeper defines the expected NFT keeper interface.
type NFTKeeper interface {
	// SaveClass saves an NFT class.
	SaveClass(ctx context.Context, class nft.Class) error
	// Mint mints an NFT.
	Mint(ctx context.Context, token nft.NFT, receiver sdk.AccAddress) error
	// Owner returns the owner of an NFT.
	Owner(ctx context.Context, r *nft.QueryOwnerRequest) (*nft.QueryOwnerResponse, error)
	// Transfer transfers an NFT from one account to another.
	Transfer(ctx context.Context, classID, nftID string, receiver sdk.AccAddress) error
	// NFTs queries NFTs with NFT module pagination logic.
	NFTs(ctx context.Context, r *nft.QueryNFTsRequest) (*nft.QueryNFTsResponse, error)
	// Classes queries classes with NFT module pagination logic.
	Classes(ctx context.Context, r *nft.QueryClassesRequest) (*nft.QueryClassesResponse, error)
	// Class queries a single class by id using NFT module query server.
	Class(ctx context.Context, r *nft.QueryClassRequest) (*nft.QueryClassResponse, error)
	// NFT queries a single NFT by class id and nft id using NFT module query server.
	NFT(ctx context.Context, r *nft.QueryNFTRequest) (*nft.QueryNFTResponse, error)
}

// BaseRegistryKeeper defines the expected base registry keeper interface.
type BaseRegistryKeeper interface {
	CreateDefaultRegistry(ctx sdk.Context, authorityAddr string, key *registrytypes.RegistryKey) error
}
