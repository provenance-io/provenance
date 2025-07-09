package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/metadata/types"
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
	HasRole(ctx sdk.Context, key *registry.RegistryKey, role registry.RegistryRole, address string) (bool, error)
	GetRegistry(ctx sdk.Context, key *registry.RegistryKey) (*registry.RegistryEntry, error)
	CreateDefaultRegistry(ctx sdk.Context, authorityAddr sdk.AccAddress, key *registry.RegistryKey) error
	AssetClassExists(ctx sdk.Context, assetClassId *string) bool
	HasNFT(ctx sdk.Context, assetClassId, nftId *string) bool
	GetNFTOwner(ctx sdk.Context, assetClassId, nftId *string) sdk.AccAddress
}

type MetaDataKeeper interface {
	GetScopeSpecification(ctx sdk.Context, scopeSpecID types.MetadataAddress) (spec types.ScopeSpecification, found bool)
	GetScope(ctx sdk.Context, id types.MetadataAddress) (types.Scope, bool)
	SetScopeSpecification(ctx sdk.Context, spec types.ScopeSpecification)
	SetScope(ctx sdk.Context, scope types.Scope) error
}
