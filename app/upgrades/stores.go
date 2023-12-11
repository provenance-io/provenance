package upgrades

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/provenance-io/provenance/app/keepers"
)

func AttemptUpgradeStoreLoaders(app StoreLoaderUpgrader, k *keepers.AppKeepers, upgrades []Upgrade) {
	// Use the dump of $home/data/upgrade-info.json:{"name":"$plan","height":321654} to determine
	// if we load a store upgrade from the handlers. No file == no error from read func.
	upgradeInfo, err := k.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(err)
	}

	// Currently in an upgrade hold for this block.
	if upgradeInfo.Name != "" && upgradeInfo.Height == app.LastBlockHeight()+1 {
		if k.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
			app.Logger().Info("Skipping upgrade based on height",
				"plan", upgradeInfo.Name,
				"upgradeHeight", upgradeInfo.Height,
				"lastHeight", app.LastBlockHeight(),
			)
		} else {
			app.Logger().Info("Managing upgrade",
				"plan", upgradeInfo.Name,
				"upgradeHeight", upgradeInfo.Height,
				"lastHeight", app.LastBlockHeight(),
			)
			// See if we have a custom store loader to use for upgrades.
			storeLoader := GetUpgradeStoreLoader(app, upgradeInfo, upgrades)
			if storeLoader != nil {
				app.SetStoreLoader(storeLoader)
			}
		}
	}
}

// GetUpgradeStoreLoader creates an StoreLoader for use in an upgrade.
// Returns nil if no upgrade info is found or the upgrade doesn't need a store loader.
func GetUpgradeStoreLoader(app StoreLoaderUpgrader, info upgradetypes.Plan, upgrades []Upgrade) baseapp.StoreLoader {
	upgrade, found := FindUpgrade(info.Name, upgrades)
	if !found {
		return nil
	}

	if len(upgrade.StoreUpgrades.Renamed) == 0 && len(upgrade.StoreUpgrades.Deleted) == 0 && len(upgrade.StoreUpgrades.Added) == 0 {
		app.Logger().Info("No store upgrades required",
			"plan", info.Name,
			"height", info.Height,
		)
		return nil
	}

	app.Logger().Info("Store upgrades",
		"plan", info.Name,
		"height", info.Height,
		"upgrade.added", upgrade.StoreUpgrades.Added,
		"upgrade.deleted", upgrade.StoreUpgrades.Deleted,
		"upgrade.renamed", upgrade.StoreUpgrades.Renamed,
	)
	return upgradetypes.UpgradeStoreLoader(info.Height, &upgrade.StoreUpgrades)
}

func FindUpgrade(name string, upgrades []Upgrade) (*Upgrade, bool) {
	for _, upgrade := range upgrades {
		if upgrade.UpgradeName == name {
			return &upgrade, true
		}
	}
	return nil, false
}
