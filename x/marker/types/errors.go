package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/marker module sentinel errors
var (
	ErrEmptyAccessGrantAddress = sdkerrors.Register(ModuleName, 2, "access grant address is empty")
	ErrAccessTypeInvalid       = sdkerrors.Register(ModuleName, 3, "invalid access type")
	ErrDuplicateAccessEntry    = sdkerrors.Register(ModuleName, 4, "access list contains duplicate entry")
	ErrInvalidMarkerStatus     = sdkerrors.Register(ModuleName, 5, "invalid marker status")
	ErrAccessTypeNotGranted    = sdkerrors.Register(ModuleName, 6, "access type not granted")
	ErrMarkerNotFound          = sdkerrors.Register(ModuleName, 7, "marker not found")
)
