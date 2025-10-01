package types

import (
	cerrs "cosmossdk.io/errors"
)

var (
	ErrInvalidPacketTimeout = cerrs.Register(ModuleName, 3, "invalid packet timeout")
	ErrInvalidVersion       = cerrs.Register(ModuleName, 4, "invalid version")
	ErrMissingOracleAddress = cerrs.Register(ModuleName, 5, "missing oracle address")
)
