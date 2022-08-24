package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrEmptyModuleAssetID = sdkerrors.Register(ModuleName, 2, "empty module asset id")
	ErrEmptyOwnerAddress  = sdkerrors.Register(ModuleName, 3, "empty owner address")
	ErrExpirationNotFound = sdkerrors.Register(ModuleName, 4, "expiration not found")
	ErrExtendExpiration   = sdkerrors.Register(ModuleName, 5, "failed to extend expiration")
	ErrNewOwnerNoMatch    = sdkerrors.Register(ModuleName, 6, "new owner doesn't match old owner")
	ErrBlockHeightLteZero = sdkerrors.Register(ModuleName, 7, "block height must be greater than zero")
	ErrInvalidDeposit     = sdkerrors.Register(ModuleName, 8, "invalid deposit amount")
	ErrMissingSigners     = sdkerrors.Register(ModuleName, 9, "at least one signer is required")
	ErrInvalidSigners     = sdkerrors.Register(ModuleName, 10, "invalid signers")
	ErrInvalidBlockHeight = sdkerrors.Register(ModuleName, 11, "invalid block height")
	ErrUnmarshal          = sdkerrors.Register(ModuleName, 12, "failed to unmarshal bytes")
	ErrInvalidKeyPrefix   = sdkerrors.Register(ModuleName, 13, "invalid key prefix")
)
