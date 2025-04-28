package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type MetaDataKeeper interface {
}

type NFTKeeper interface {
	// GetOwner returns the owner of an NFT
	GetOwner(ctx context.Context, classID, nftID string) sdk.AccAddress

	// HasNFT checks if an NFT exists
	HasNFT(ctx context.Context, classID, id string) bool
}
