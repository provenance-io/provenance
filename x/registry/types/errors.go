// Package types provides error definitions and utilities for the registry module.
//
// This file defines the error codes and error creation functions used throughout
// the registry module. It provides a centralized way to create consistent error
// messages with proper error codes for the registry module.
//
// Error codes are defined as constants and registered with the Cosmos SDK error
// system. Each error type has a corresponding constructor function that wraps
// the base error with additional context.
//
// Example usage:
//
//	err := NewErrCodeRegistryAlreadyExists("registry_key")
//	err := NewErrCodeNFTNotFound("nft_id")
//	err := NewErrCodeUnauthorized("authority does not own the NFT")
//	err := NewErrCodeInvalidRole("role_type")
//	err := NewErrCodeRegistryNotFound("registry_key")
//	err := NewErrCodeAddressAlreadyHasRole("address", "role")
//	err := NewErrCodeInvalidHrp("hrp_value")
//	err := NewErrCodeInvalidKey("key_value")
package types

import (
	"fmt"

	cerrs "cosmossdk.io/errors"
)

type ErrCode string

const (
	ErrCodeRegistryAlreadyExists  ErrCode = "REGISTRY_ALREADY_EXISTS"
	ErrCodeNFTNotFound            ErrCode = "NFT_NOT_FOUND"
	ErrCodeUnauthorized           ErrCode = "UNAUTHORIZED"
	ErrCodeInvalidRole            ErrCode = "INVALID_ROLE"
	ErrCodeRegistryNotFound       ErrCode = "REGISTRY_NOT_FOUND"
	ErrCodeAddressAlreadyHasRole  ErrCode = "ADDRESS_ALREADY_HAS_ROLE"
	ErrCodeInvalidKey             ErrCode = "INVALID_KEY"
	ErrCodeAddressDoesNotHaveRole ErrCode = "ADDRESS_DOES_NOT_HAVE_ROLE"
	ErrCodeInvalidField           ErrCode = "INVALID_FIELD"
)

var (
	ErrRegistryAlreadyExists  = cerrs.Register(ModuleName, 1, "registry already exists")
	ErrNFTNotFound            = cerrs.Register(ModuleName, 2, "NFT does not exist")
	ErrUnauthorized           = cerrs.Register(ModuleName, 3, "unauthorized")
	ErrInvalidRole            = cerrs.Register(ModuleName, 4, "invalid role")
	ErrRegistryNotFound       = cerrs.Register(ModuleName, 5, "registry not found")
	ErrAddressAlreadyHasRole  = cerrs.Register(ModuleName, 6, "address already has role")
	ErrInvalidKey             = cerrs.Register(ModuleName, 7, "invalid key")
	ErrAddressDoesNotHaveRole = cerrs.Register(ModuleName, 8, "address does not have role")
	ErrInvalidField           = cerrs.Register(ModuleName, 9, "invalid field")
)

func NewErrCodeRegistryAlreadyExists(key string) error {
	return cerrs.Wrapf(ErrRegistryAlreadyExists, "registry already exists for key: %q", key)
}

func NewErrCodeNFTNotFound(nftID string) error {
	return cerrs.Wrapf(ErrNFTNotFound, "NFT does not exist: %q", nftID)
}

func NewErrCodeUnauthorized(why string) error {
	return cerrs.Wrapf(ErrUnauthorized, "unauthorized access: %s", why)
}

func NewErrCodeInvalidRole(role string) error {
	return cerrs.Wrapf(ErrInvalidRole, "invalid role: %q", role)
}

func NewErrCodeRegistryNotFound(key string) error {
	return cerrs.Wrapf(ErrRegistryNotFound, "registry not found for key: %q", key)
}

func NewErrCodeAddressAlreadyHasRole(address, role string) error {
	return cerrs.Wrapf(ErrAddressAlreadyHasRole, "address %q already has role %q", address, role)
}

func NewErrCodeAddressDoesNotHaveRole(address, role string) error {
	return cerrs.Wrapf(ErrAddressDoesNotHaveRole, "address %q does not have role %q", address, role)
}

func NewErrCodeInvalidField(field, format string, args ...interface{}) error {
	return cerrs.Wrapf(ErrInvalidField, "invalid %s: %s", field, fmt.Sprintf(format, args...))
}
