package rc1

import (
	store "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/provenance-io/provenance/app/upgrades"

	triggertypes "github.com/provenance-io/provenance/x/trigger/types"
)

const UpgradeName = "rust-rc1"

var Upgrade = upgrades.Upgrade{
	UpgradeName:     UpgradeName,
	UpgradeStrategy: UpgradeStrategy,
	StoreUpgrades: store.StoreUpgrades{
		Added: []string{
			triggertypes.ModuleName,
		},
		Deleted: []string{},
	},
}
