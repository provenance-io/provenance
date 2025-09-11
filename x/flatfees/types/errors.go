package types

import (
	cerrs "cosmossdk.io/errors"
)

// The x/flatfees module sentinel errors.

var (
	ErrMsgFeeDoesNotExist = cerrs.Register(ModuleName, 5, "fee for type does not exist")
)
