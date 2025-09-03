// Package errors defines error types for the quarantine module.
package errors

import "cosmossdk.io/errors"

// quarantineCodespace is the codespace for all errors defined in quarantine package
const quarantineCodespace = "quarantine"

// ErrInvalidValue is returned when a provided value is invalid.
var ErrInvalidValue = errors.Register(quarantineCodespace, 2, "invalid value")
