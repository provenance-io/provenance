package mainnet

import (
	"github.com/provenance-io/provenance/app/upgrades"
	v_1_17_0 "github.com/provenance-io/provenance/app/upgrades/mainnet/v1.17.0"
	v_1_17_0_rc1 "github.com/provenance-io/provenance/app/upgrades/mainnet/v1.17.0/rc1"
	v_1_17_0_rc2 "github.com/provenance-io/provenance/app/upgrades/mainnet/v1.17.0/rc2"
	v_1_17_0_rc3 "github.com/provenance-io/provenance/app/upgrades/mainnet/v1.17.0/rc3"
	v_1_18_0 "github.com/provenance-io/provenance/app/upgrades/mainnet/v1.18.0"
	v_1_18_0_rc1 "github.com/provenance-io/provenance/app/upgrades/mainnet/v1.18.0/rc1"
)

var (
	Upgrades = []upgrades.Upgrade{
		v_1_17_0_rc1.Upgrade,
		v_1_17_0_rc2.Upgrade,
		v_1_17_0_rc3.Upgrade,
		v_1_17_0.Upgrade,
		v_1_18_0_rc1.Upgrade,
		v_1_18_0.Upgrade,
	}
)
