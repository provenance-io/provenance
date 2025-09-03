package ibcratelimit

import (
	cerrs "cosmossdk.io/errors"
)

var (
	ErrRateLimitExceeded = cerrs.Register(ModuleName, 2, "rate limit exceeded") //nolint:revive
	ErrBadMessage        = cerrs.Register(ModuleName, 3, "bad message")         //nolint:revive
	ErrContractError     = cerrs.Register(ModuleName, 4, "contract error")
)
