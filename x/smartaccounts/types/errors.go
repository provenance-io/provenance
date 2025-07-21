package types

import (
	cerrs "cosmossdk.io/errors"
)

// x/smartaccount module errors
var (
	ErrParseCredential          = cerrs.Register(ModuleName, 1, "credential parsing error")
	ErrSmartAccountDoesNotExist = cerrs.Register(ModuleName, 2, "smart account does not exist")
	ErrDuplicateCredential      = cerrs.Register(ModuleName, 3, "duplicate credential")
)
