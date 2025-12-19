package types

import (
	cerrs "cosmossdk.io/errors"
)

// The x/flatfees module sentinel errors.

var (
	// ErrMsgFeeDoesNotExist is returned when a message fee entry does not exist for a given message type.
	ErrMsgFeeDoesNotExist = cerrs.Register(ModuleName, 5, "fee for type does not exist")
)
