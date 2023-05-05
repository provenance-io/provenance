package types

import (
	cerrs "cosmossdk.io/errors"
)

var (
	ErrTriggerNotFound    = cerrs.Register(ModuleName, 2, "trigger not found")
	ErrEventNotFound      = cerrs.Register(ModuleName, 3, "event not found")
	ErrQueueIndexNotFound = cerrs.Register(ModuleName, 4, "queue index not found")
)
