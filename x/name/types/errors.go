package types

import (
	cerrs "cosmossdk.io/errors"
)

// x/name module errors
var (
	// ErrNameNotBound is a sentinel error returned when a name is not bound to an address.
	ErrNameNotBound = cerrs.Register(ModuleName, 2, "no address bound to name")
	// ErrNameAlreadyBound occurs when a bind request is made against an existing name
	ErrNameAlreadyBound = cerrs.Register(ModuleName, 3, "name is already bound to an address")
	// ErrNameInvalid occurs when a name is invalid
	ErrNameInvalid = cerrs.Register(ModuleName, 4, "value provided for name is invalid")
	// ErrNameSegmentTooShort occurs when a segment of a name is shorter than the minimum length
	ErrNameSegmentTooShort = cerrs.Register(ModuleName, 5, "segment of name is too short")
	// ErrNameSegmentTooLong occurs when a segment of a name is longer than the maximum length
	ErrNameSegmentTooLong = cerrs.Register(ModuleName, 6, "segment of name is too long")
	// ErrNameHasTooManySegments occurs when a name has too many segments (names separated by a period)
	ErrNameHasTooManySegments = cerrs.Register(ModuleName, 7, "name has too many segments")
	// ErrInvalidAddress indicates the address given does not match an existing account.
	ErrInvalidAddress = cerrs.Register(ModuleName, 8, "invalid account address")
	// ErrNameContainsSegments indicates a multi-segment name in a single segment context.
	ErrNameContainsSegments = cerrs.Register(ModuleName, 9, "invalid name: \".\" is reserved")
)
