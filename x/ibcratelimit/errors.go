package ibcratelimit

import (
	cerrs "cosmossdk.io/errors"
)

var (
	ErrRateLimitExceeded = cerrs.Register(ModuleName, 2, "rate limit exceeded")
	ErrBadMessage        = cerrs.Register(ModuleName, 3, "bad message")
	ErrContractError     = cerrs.Register(ModuleName, 4, "contract error")
)
