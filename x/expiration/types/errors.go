package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/expiration module sentinel errors
var (
	ErrEmptyModuleAssetId = sdkerrors.Register(ModuleName, 2, "empty module asset id.")
	ErrEmptyOwnerAddress  = sdkerrors.Register(ModuleName, 3, "empty owner address.")
	ErrExpirationNotFound = sdkerrors.Register(ModuleName, 4, "expiration for module asset does not exist.")
	ErrExtendExpiration   = sdkerrors.Register(ModuleName, 5, "failed to extend expiration")
	ErrNewOwnerNoMatch    = sdkerrors.Register(ModuleName, 6, "new owner doesn't match old owner")
	ErrBlockHeightLteZero = sdkerrors.Register(ModuleName, 7, "block height must be greater than zero")
	ErrBlockHeightInPast  = sdkerrors.Register(ModuleName, 8, "block height must be higher than current block height")
	ErrInvalidDeposit     = sdkerrors.Register(ModuleName, 8, "invalid deposit")
	ErrMissingSigners     = sdkerrors.Register(ModuleName, 9, "at least one signer is required")
)
