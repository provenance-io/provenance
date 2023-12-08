package rc1

import (
	store "github.com/cosmos/cosmos-sdk/store/types"
	icqtypes "github.com/cosmos/ibc-apps/modules/async-icq/v6/types"

	"github.com/provenance-io/provenance/app/upgrades"
	"github.com/provenance-io/provenance/x/exchange"
	"github.com/provenance-io/provenance/x/hold"
	ibchookstypes "github.com/provenance-io/provenance/x/ibchooks/types"
	oracletypes "github.com/provenance-io/provenance/x/oracle/types"
)

const UpgradeName = "saffron-rc1"

var Upgrade = upgrades.Upgrade{
	UpgradeName:     UpgradeName,
	UpgradeStrategy: UpgradeStrategy,
	StoreUpgrades: store.StoreUpgrades{
		Added: []string{
			icqtypes.ModuleName,
			oracletypes.ModuleName,
			ibchookstypes.StoreKey,
			hold.ModuleName,
			exchange.ModuleName,
		},
		Deleted: []string{},
	},
}
