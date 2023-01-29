package types

import (
	cerrs "cosmossdk.io/errors"
)

// x/metadata module errors
var (
	// ErrOSLocatorAlreadyBound occurs when a bindoslocator request is made against an existing owner address
	ErrOSLocatorAlreadyBound = cerrs.Register(ModuleName, 2, "owner address is already bound to an uri")
	// ErrInvalidAddress indicates the address given does not match an existing account.
	ErrInvalidAddress      = cerrs.Register(ModuleName, 3, "address does not match an existing account")
	ErrAddressNotBound     = cerrs.Register(ModuleName, 4, "no locator bound to address")
	ErrOSLocatorURIToolong = cerrs.Register(ModuleName, 5, "uri length greater than allowed")
	ErrNoRecordsFound      = cerrs.Register(ModuleName, 6, "No records found.")
	ErrOSLocatorURIInvalid = cerrs.Register(ModuleName, 7, "uri is invalid")
	ErrScopeIdInvalid      = cerrs.Register(ModuleName, 8, "scope id cannot be empty")
	ErrScopeNotFound       = cerrs.Register(ModuleName, 9, "scope not found with id ")
)
