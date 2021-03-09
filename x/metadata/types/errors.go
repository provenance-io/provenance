package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/name module errors
var (
	// ErrOSLocatorAlreadyBound occurs when a bindoslocator request is made against an existing owner address
	ErrOSLocatorAlreadyBound = sdkerrors.Register(ModuleName, 2, "owner address is already bound to an uri")
	// ErrInvalidAddress indicates the address given does not match an existing account.
	ErrInvalidAddress = sdkerrors.Register(ModuleName, 8, "address does not match an existing account")
)
