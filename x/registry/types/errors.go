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
	cerrs "cosmossdk.io/errors"
	"github.com/provenance-io/provenance/x/registry"
)

type ErrCode string

const (
	ErrCodeRegistryAlreadyExists ErrCode = "REGISTRY_ALREADY_EXISTS"
	ErrCodeNFTNotFound           ErrCode = "NFT_NOT_FOUND"
	ErrCodeUnauthorized          ErrCode = "UNAUTHORIZED"
	ErrCodeInvalidRole           ErrCode = "INVALID_ROLE"
	ErrCodeRegistryNotFound      ErrCode = "REGISTRY_NOT_FOUND"
	ErrCodeAddressAlreadyHasRole ErrCode = "ADDRESS_ALREADY_HAS_ROLE"
	ErrCodeInvalidHrp            ErrCode = "INVALID_HRP"
	ErrCodeInvalidKey            ErrCode = "INVALID_KEY"
)

var (
	ErrRegistryAlreadyExists = cerrs.Register(registry.ModuleName, 1, "registry already exists")
	ErrNFTNotFound           = cerrs.Register(registry.ModuleName, 2, "NFT does not exist")
	ErrUnauthorized          = cerrs.Register(registry.ModuleName, 3, "unauthorized")
	ErrInvalidRole           = cerrs.Register(registry.ModuleName, 4, "invalid role")
	ErrRegistryNotFound      = cerrs.Register(registry.ModuleName, 5, "registry not found")
	ErrAddressAlreadyHasRole = cerrs.Register(registry.ModuleName, 6, "address already has role")
	ErrInvalidHrp            = cerrs.Register(registry.ModuleName, 7, "invalid hrp")
	ErrInvalidKey            = cerrs.Register(registry.ModuleName, 8, "invalid key")
)

func NewErrCodeRegistryAlreadyExists(key string) error {
	return cerrs.Wrapf(ErrRegistryAlreadyExists, "registry already exists for key: %s", key)
}

func NewErrCodeNFTNotFound(nftId string) error {
	return cerrs.Wrapf(ErrNFTNotFound, "NFT does not exist: %s", nftId)
}

func NewErrCodeUnauthorized(why string) error {
	return cerrs.Wrapf(ErrUnauthorized, "unauthorized access: %s", why)
}

func NewErrCodeInvalidRole(role string) error {
	return cerrs.Wrapf(ErrInvalidRole, "invalid role: %s", role)
}

func NewErrCodeRegistryNotFound(key string) error {
	return cerrs.Wrapf(ErrRegistryNotFound, "registry not found for key: %s", key)
}

func NewErrCodeAddressAlreadyHasRole(address, role string) error {
	return cerrs.Wrapf(ErrAddressAlreadyHasRole, "address %s already has role %s", address, role)
}

func NewErrCodeInvalidHrp(hrp string) error {
	return cerrs.Wrapf(ErrInvalidHrp, "invalid hrp: %s", hrp)
}

func NewErrCodeInvalidKey(key string) error {
	return cerrs.Wrapf(ErrInvalidKey, "invalid key: %s", key)
}
