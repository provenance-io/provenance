package rc1

import (
	store "github.com/cosmos/cosmos-sdk/store/types"

	"github.com/provenance-io/provenance/app/upgrades"
	"github.com/provenance-io/provenance/x/ibcratelimit"
)

const UpgradeName = "tourmaline-rc1"

var Upgrade = upgrades.Upgrade{
	UpgradeName:     UpgradeName,
	UpgradeStrategy: UpgradeStrategy,
	StoreUpgrades: store.StoreUpgrades{
		Added: []string{
			ibcratelimit.ModuleName,
		},
		Deleted: []string{},
	},
}
