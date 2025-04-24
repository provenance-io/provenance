package keeper

import (
	"context"

	"cosmossdk.io/x/nft"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BankKeeper is an interface that allows the ledger keeper to send coins.
type BankKeeper interface {
	SendCoins(ctx context.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) error
	HasBalance(ctx context.Context, addr sdk.AccAddress, amt sdk.Coin) bool
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	SpendableCoin(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	HasSupply(ctx context.Context, denom string) bool
}

type MetaDataKeeper interface {
}

type NFTKeeper interface {
	// GetOwner returns the owner of an NFT
	GetOwner(ctx context.Context, classID, nftID string) sdk.AccAddress

	// HasClass checks if an NFT class exists
	HasClass(ctx context.Context, classID string) bool

	// GetClass returns an NFT class by ID
	GetClass(ctx context.Context, classID string) (nft.Class, bool)
}
