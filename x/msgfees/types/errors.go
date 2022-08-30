package types

import (
	cerrs "cosmossdk.io/errors"
)

// x/msgfees module sentinel errors
var (
	ErrEmptyMsgType        = cerrs.Register(ModuleName, 2, "msg type is empty")
	ErrInvalidFee          = cerrs.Register(ModuleName, 3, "invalid fee amount")
	ErrMsgFeeAlreadyExists = cerrs.Register(ModuleName, 4, "fee for type already exists.")
	ErrMsgFeeDoesNotExist  = cerrs.Register(ModuleName, 5, "fee for type does not exist.")
	ErrInvalidFeeProposal  = cerrs.Register(ModuleName, 6, "invalid fee proposal")
)
