package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/name module errors
var (
	// ErrNameNotBound is a sentinel error returned when a name is not bound to an address.
	ErrNameNotBound = sdkerrors.Register(ModuleName, 2, "no address bound to name")
)
