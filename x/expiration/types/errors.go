package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrEmptyModuleAssetID  = sdkerrors.Register(ModuleName, 2, "empty module asset id")
	ErrEmptyOwnerAddress   = sdkerrors.Register(ModuleName, 3, "empty owner address")
	ErrNotFound            = sdkerrors.Register(ModuleName, 4, "expiration not found")
	ErrExtendExpiration    = sdkerrors.Register(ModuleName, 5, "failed to extend expiration")
	ErrNewOwnerNoMatch     = sdkerrors.Register(ModuleName, 6, "new owner doesn't match old owner")
	ErrTimeInPast          = sdkerrors.Register(ModuleName, 7, "expiration time is in the past")
	ErrInvalidDeposit      = sdkerrors.Register(ModuleName, 8, "invalid deposit amount")
	ErrMissingSigners      = sdkerrors.Register(ModuleName, 9, "at least one signer is required")
	ErrInvalidSigners      = sdkerrors.Register(ModuleName, 10, "invalid signers")
	ErrUnmarshal           = sdkerrors.Register(ModuleName, 12, "failed to unmarshal bytes")
	ErrInvalidKeyPrefix    = sdkerrors.Register(ModuleName, 13, "invalid key prefix")
	ErrInvalidMessage      = sdkerrors.Register(ModuleName, 14, "invalid expiration message")
	ErrInvoke              = sdkerrors.Register(ModuleName, 15, "failed to invoke expiration")
	ErrDurationValue       = sdkerrors.Register(ModuleName, 16, "invalid duration value")
	ErrInsufficientDeposit = sdkerrors.Register(ModuleName, 17, "insufficient funds for expiration deposit")
)
