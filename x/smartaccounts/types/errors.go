package types

import (
	cerrs "cosmossdk.io/errors"
)

// x/smartaccount module errors
var (
	ErrParseCredential          = cerrs.Register(ModuleName, 1, "credential parsing error")
	ErrSmartAccountDoesNotExist = cerrs.Register(ModuleName, 2, "smart account does not exist")
		ErrDuplicateCredential            = errorsmod.Register(ModuleName, 10, "duplicate credential")
	ErrSmartAccountsNotEnabled        = errorsmod.Register(ModuleName, 11, "smart accounts are not enabled")
)
)
