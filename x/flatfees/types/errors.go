package types

import (
	cerrs "cosmossdk.io/errors"
)

// The x/flatfees module sentinel errors.

var (
	ErrMsgFeeDoesNotExist  = cerrs.Register(ModuleName, 5, "fee for type does not exist")
	ErrOracleAlreadyExists = cerrs.Register(ModuleName, 1100, "oracle address already exists")
	ErrOracleNotFound      = cerrs.Register(ModuleName, 1101, "oracle address not found")
	ErrInvalidOracleAddr   = cerrs.Register(ModuleName, 1103, "invalid oracle address")
)
