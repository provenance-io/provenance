package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/registry"
)

// BankKeeper is an interface that allows the ledger keeper to send coins.
type BankKeeper interface {
	SendCoins(ctx context.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) error
	HasBalance(ctx context.Context, addr sdk.AccAddress, amt sdk.Coin) bool
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	SpendableCoin(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	HasSupply(ctx context.Context, denom string) bool
}

type RegistryKeeper interface {
	HasRole(ctx sdk.Context, key *registry.RegistryKey, role string, address string) (bool, error)
	GetRegistry(ctx sdk.Context, key *registry.RegistryKey) (*registry.RegistryEntry, error)
	AssetClassExists(ctx sdk.Context, assetClassId *string) bool
	HasNFT(ctx sdk.Context, assetClassId, nftId *string) bool
	GetNFTOwner(ctx sdk.Context, assetClassId, nftId *string) sdk.AccAddress
}
