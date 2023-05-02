package types

import (
	cerrs "cosmossdk.io/errors"
)

var (
	ErrTriggerNotFound = cerrs.Register(ModuleName, 2, "trigger not found")
)
