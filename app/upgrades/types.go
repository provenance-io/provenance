package upgrades

import (
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/baseapp"
	store "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/provenance-io/provenance/app/keepers"
)

type AppUpgrader interface {
	ModuleManager() *module.Manager
	Configurator() module.Configurator
	Keepers() *keepers.AppKeepers
}

type StoreLoaderUpgrader interface {
	Logger() log.Logger
	LastBlockHeight() int64
	SetStoreLoader(loader baseapp.StoreLoader)
}

type UpgradeStrategy = func(ctx sdk.Context, app AppUpgrader, vm module.VersionMap) (module.VersionMap, error)

// Upgrade defines a struct containing necessary fields that a SoftwareUpgradeProposal
// must have written, in order for the state migration to go smoothly.
// An upgrade must implement this struct, and then set it in the app.go.
// The app.go will then define the handler.
type Upgrade struct {
	// Upgrade version name, for the upgrade handler, e.g. `v7`
	UpgradeName string

	// CreateUpgradeHandler defines the function that creates an upgrade handler
	UpgradeStrategy UpgradeStrategy

	// Store upgrades, should be used for any new modules introduced, new modules deleted, or store names renamed.
	StoreUpgrades store.StoreUpgrades
}
