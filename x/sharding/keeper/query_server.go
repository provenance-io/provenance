package keeper

import (
	"github.com/provenance-io/provenance/x/sharding/types"
)

var _ types.QueryServer = Keeper{}
