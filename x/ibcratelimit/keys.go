package ibcratelimit

import "cosmossdk.io/collections"

const (
	// ModuleName defines the module name
	ModuleName = "ratelimitedibc"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName
)

var (
	// ParamsKey is the key to obtain the module's params.
	ParamsKey = []byte{0x01}
	// ParamsKeyPrefix is the collections.Prefix for the Params Item.
	ParamsKeyPrefix = collections.NewPrefix(ParamsKey)
)
