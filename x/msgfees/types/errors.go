package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/msgfees module sentinel errors
var (
	ErrEmptyMsgType        = sdkerrors.Register(ModuleName, 2, "msg type is empty")
	ErrInvalidFee          = sdkerrors.Register(ModuleName, 3, "invalid fee amount")
	ErrMsgFeeAlreadyExists = sdkerrors.Register(ModuleName, 4, "fee for type already exists.")
	ErrMsgFeeDoesNotExist  = sdkerrors.Register(ModuleName, 5, "fee for type does not exist.")
	ErrInvalidFeeProposal  = sdkerrors.Register(ModuleName, 6, "invalid fee proposal")
)
