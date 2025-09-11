package keeper

import (
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/provenance-io/provenance/x/asset/types"
)

// Keeper of the asset store
type Keeper struct {
	cdc            codec.BinaryCodec
	moduleAccount  sdk.AccAddress
	markerKeeper   types.MarkerKeeper
	nftKeeper      types.NFTKeeper
	registryKeeper types.BaseRegistryKeeper
}

// NewKeeper creates a new asset Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	markerKeeper types.MarkerKeeper,
	nftKeeper types.NFTKeeper,
	registryKeeper types.BaseRegistryKeeper,
) Keeper {
	return Keeper{
		cdc:            cdc,
		moduleAccount:  authtypes.NewModuleAddress(types.ModuleName),
		markerKeeper:   markerKeeper,
		nftKeeper:      nftKeeper,
		registryKeeper: registryKeeper,
	}
}

// Logger returns a module-specific logger
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// GetModuleAddress returns the module account address
func (k Keeper) GetModuleAddress() sdk.AccAddress {
	return k.moduleAccount
}
