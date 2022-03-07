package keeper

import (
	"github.com/provenance-io/provenance/x/epoch/types"
)

var _ types.QueryServer = Keeper{}
