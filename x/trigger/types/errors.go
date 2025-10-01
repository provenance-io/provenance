package types

import (
	cerrs "cosmossdk.io/errors"
)

var (
	ErrTriggerNotFound         = cerrs.Register(ModuleName, 2, "trigger not found")
	ErrEventNotFound           = cerrs.Register(ModuleName, 3, "event not found")
	ErrQueueIndexNotFound      = cerrs.Register(ModuleName, 4, "queue index not found")
	ErrQueueEmpty              = cerrs.Register(ModuleName, 5, "queue is empty")
	ErrGasLimitNotFound        = cerrs.Register(ModuleName, 6, "gas limit not found")
	ErrTriggerGasLimitExceeded = cerrs.Register(ModuleName, 7, "gas limit execeeded for trigger")
	ErrInvalidTriggerAuthority = cerrs.Register(ModuleName, 8, "signer does not have authority to destroy trigger")
	ErrNoTriggerEvent          = cerrs.Register(ModuleName, 9, "trigger does not have event")
	ErrInvalidBlockHeight      = cerrs.Register(ModuleName, 10, "block height has already passed")
	ErrInvalidBlockTime        = cerrs.Register(ModuleName, 11, "block time has already passed")
)
