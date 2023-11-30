package v1_16_0

import (
	store "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/provenance-io/provenance/app/upgrades"
	triggertypes "github.com/provenance-io/provenance/x/trigger/types"
)

const UpgradeName = "rust"

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
