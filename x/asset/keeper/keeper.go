package keeper

import (
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/asset/types"
)

// Keeper of the asset store
type Keeper struct {
	cdc            codec.BinaryCodec
	router         baseapp.MessageRouter
	moduleAccount  sdk.AccAddress
	markerKeeper   types.MarkerKeeper
	nftKeeper      types.NFTKeeper
	registryKeeper types.BaseRegistryKeeper
}

// NewKeeper creates a new asset Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	router baseapp.MessageRouter,
	moduleAccount sdk.AccAddress,
	markerKeeper types.MarkerKeeper,
	nftKeeper types.NFTKeeper,
	registryKeeper types.BaseRegistryKeeper,
) Keeper {
	if nftKeeper == nil {
		panic("nft keeper is required for asset module")
	}

	if router == nil {
		panic("router is required for asset module")
	}

	return Keeper{
		cdc:            cdc,
		router:         router,
		moduleAccount:  moduleAccount,
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
