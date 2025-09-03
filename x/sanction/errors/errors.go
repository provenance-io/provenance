// Package errors defines custom error types for the sanction module.
package errors

import (
	cerrs "cosmossdk.io/errors"
)

// sanctionCodespace is the codespace for all errors defined in sanction package
const sanctionCodespace = "sanction"

var (
	ErrInvalidParams      = cerrs.Register(sanctionCodespace, 2, "invalid params") //nolint:revive
	ErrUnsanctionableAddr = cerrs.Register(sanctionCodespace, 3, "address cannot be sanctioned")
	ErrInvalidTempStatus  = cerrs.Register(sanctionCodespace, 4, "invalid temp status")
	ErrSanctionedAccount  = cerrs.Register(sanctionCodespace, 5, "account is sanctioned")
)
