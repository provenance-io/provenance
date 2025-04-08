package errors

import (
	cerrs "cosmossdk.io/errors"
)

// sanctionCodespace is the codespace for all errors defined in sanction package
const sanctionCodespace = "sanction"

var (
	ErrInvalidParams      = cerrs.Register(sanctionCodespace, 2, "invalid params")
	ErrUnsanctionableAddr = cerrs.Register(sanctionCodespace, 3, "address cannot be sanctioned")
	ErrInvalidTempStatus  = cerrs.Register(sanctionCodespace, 4, "invalid temp status")
	ErrSanctionedAccount  = cerrs.Register(sanctionCodespace, 5, "account is sanctioned")
)
