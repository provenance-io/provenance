package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	registrytypes "github.com/provenance-io/provenance/x/registry/types"
)

// BankKeeper is an interface that allows the ledger keeper to send coins.
type BankKeeper interface {
	SendCoins(ctx context.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) error
	HasBalance(ctx context.Context, addr sdk.AccAddress, amt sdk.Coin) bool
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	SpendableCoin(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	HasSupply(ctx context.Context, denom string) bool
	BlockedAddr(addr sdk.AccAddress) bool
}

type RegistryKeeper interface {
	HasRole(ctx context.Context, key *registrytypes.RegistryKey, role registrytypes.RegistryRole, address string) (bool, error)
	GetRegistry(ctx context.Context, key *registrytypes.RegistryKey) (*registrytypes.RegistryEntry, error)
	AssetClassExists(ctx context.Context, assetClassID *string) bool
	HasNFT(ctx context.Context, assetClassID, nftID *string) bool
	GetNFTOwner(ctx context.Context, assetClassID, nftID *string) sdk.AccAddress
}
