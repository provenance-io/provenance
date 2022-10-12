package types

import (
	cerrs "cosmossdk.io/errors"
)

var (
	ErrEmptyModuleAssetID  = cerrs.Register(ModuleName, 2, "empty module asset id")
	ErrEmptyOwnerAddress   = cerrs.Register(ModuleName, 3, "empty owner address")
	ErrNotFound            = cerrs.Register(ModuleName, 4, "expiration not found")
	ErrExtendExpiration    = cerrs.Register(ModuleName, 5, "failed to extend expiration")
	ErrNewOwnerNoMatch     = cerrs.Register(ModuleName, 6, "new owner doesn't match old owner")
	ErrExpirationTime      = cerrs.Register(ModuleName, 7, "invalid expiration time")
	ErrInvalidDeposit      = cerrs.Register(ModuleName, 8, "invalid deposit amount")
	ErrMissingSigners      = cerrs.Register(ModuleName, 9, "at least one signer is required")
	ErrInvalidSigners      = cerrs.Register(ModuleName, 10, "invalid signers")
	ErrUnmarshal           = cerrs.Register(ModuleName, 12, "failed to unmarshal bytes")
	ErrInvalidKey          = cerrs.Register(ModuleName, 13, "invalid key")
	ErrInvalidMessage      = cerrs.Register(ModuleName, 14, "invalid expiration message")
	ErrInvokeExpiration    = cerrs.Register(ModuleName, 15, "failed to invoke expiration")
	ErrDurationValue       = cerrs.Register(ModuleName, 16, "invalid duration value")
	ErrInsufficientDeposit = cerrs.Register(ModuleName, 17, "insufficient funds for expiration deposit")
	ErrMsgHandler          = cerrs.Register(ModuleName, 18, "invalid message handler")
	ErrSetExpiration       = cerrs.Register(ModuleName, 20, "failed set expiration")
	ErrResolveDepositor    = cerrs.Register(ModuleName, 21, "failed to resolve depositor")
)
