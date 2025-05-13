package keeper

import (
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/provenance-io/provenance/x/asset/types"
	ledgerkeeper "github.com/provenance-io/provenance/x/ledger/keeper"
	markerkeeper "github.com/provenance-io/provenance/x/marker/keeper"
)

// Keeper of the asset store
type Keeper struct {
	cdc            codec.BinaryCodec
	storeKey       storetypes.StoreKey
	nftKeeper      types.NFTKeeper
	router         baseapp.MessageRouter
	ledgerKeeper   ledgerkeeper.BaseKeeper
	registryKeeper types.BaseRegistryKeeper
	markerKeeper   markerkeeper.MarkerKeeperI
}

// NewKeeper creates a new asset Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	key storetypes.StoreKey,
	nftKeeper types.NFTKeeper,
	router baseapp.MessageRouter,
	ledgerKeeper ledgerkeeper.BaseKeeper,
	registryKeeper types.BaseRegistryKeeper,
	markerKeeper markerkeeper.MarkerKeeperI,
) Keeper {
	if nftKeeper == nil {
		panic("nft keeper is required for asset module")
	}

	if router == nil {
		panic("router is required for asset module")
	}

	return Keeper{
		cdc:            cdc,
		storeKey:       key,
		nftKeeper:      nftKeeper,
		router:         router,
		ledgerKeeper:   ledgerKeeper,
		registryKeeper: registryKeeper,
		markerKeeper:   markerKeeper,
	}
}

// Logger returns a module-specific logger
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// GetModuleAddress returns the module account address
func (k Keeper) GetModuleAddress() sdk.AccAddress {
	return authtypes.NewModuleAddress(types.ModuleName)
}
