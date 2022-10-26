package types

import (
	cerrs "cosmossdk.io/errors"
)

// x/marker module sentinel errors
var (
	ErrEmptyAccessGrantAddress  = cerrs.Register(ModuleName, 2, "access grant address is empty")
	ErrAccessTypeInvalid        = cerrs.Register(ModuleName, 3, "invalid access type")
	ErrDuplicateAccessEntry     = cerrs.Register(ModuleName, 4, "access list contains duplicate entry")
	ErrInvalidMarkerStatus      = cerrs.Register(ModuleName, 5, "invalid marker status")
	ErrAccessTypeNotGranted     = cerrs.Register(ModuleName, 6, "access type not granted")
	ErrMarkerNotFound           = cerrs.Register(ModuleName, 7, "marker not found")
	ErrReflectAccessTypeInvalid = cerrs.Register(ModuleName, 8, "access list contains mint and/or burn")
	ErrReflectSupplyFixed       = cerrs.Register(ModuleName, 9, "reflected marker can't have fixed supply")
)
