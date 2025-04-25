package types

import (
	"context"
	"cosmossdk.io/x/nft"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NFTKeeper defines the expected NFT keeper interface
type NFTKeeper interface {
	// SaveClass saves an NFT class
	SaveClass(ctx context.Context, class nft.Class) error
	// HasClass checks if an NFT class exists
	HasClass(ctx context.Context, classID string) bool
	// GetClass returns an NFT class by ID
	GetClass(ctx context.Context, classID string) (nft.Class, bool)
	// GetClasses returns all NFT classes
	GetClasses(ctx context.Context) []*nft.Class
	// GetNFTsOfClass returns all NFTs of a class
	GetNFTsOfClass(ctx context.Context, classID string) []nft.NFT
	// GetOwner returns the owner of an NFT
	GetOwner(ctx context.Context, classID, nftID string) sdk.AccAddress
	// Mint mints an NFT
	Mint(ctx context.Context, token nft.NFT, receiver sdk.AccAddress) error
	// GetNFT returns an NFT by class and ID
	GetNFT(ctx context.Context, classID, nftID string) (nft.NFT, bool)
}
