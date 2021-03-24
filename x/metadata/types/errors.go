package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/metadata module errors
var (
	// ErrOSLocatorAlreadyBound occurs when a bindoslocator request is made against an existing owner address
	ErrOSLocatorAlreadyBound = sdkerrors.Register(ModuleName, 2, "owner address is already bound to an uri")
	// ErrInvalidAddress indicates the address given does not match an existing account.
	ErrInvalidAddress      = sdkerrors.Register(ModuleName, 3, "address does not match an existing account")
	ErrAddressNotBound     = sdkerrors.Register(ModuleName, 4, "no locator bound to address")
	ErrOSLocatorURIToolong = sdkerrors.Register(ModuleName, 5, "uri length greater than allowed")
	ErrNoRecordsFound      = sdkerrors.Register(ModuleName, 6, "No records found.")
	ErrOSLocatorURIInvalid = sdkerrors.Register(ModuleName, 7, "uri is invalid")
)
