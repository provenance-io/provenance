package errors

import sdkserrs "github.com/cosmos/cosmos-sdk/x/sanction/errors"

var (
	ErrInvalidParams      = sdkserrs.ErrInvalidParams
	ErrUnsanctionableAddr = sdkserrs.ErrUnsanctionableAddr
	ErrInvalidTempStatus  = sdkserrs.ErrInvalidTempStatus
	ErrSanctionedAccount  = sdkserrs.ErrSanctionedAccount
)

// TODO: Once we have an official version of the SDK without the sanction module:
//       1. Delete everything above this comment.
//       2. Uncomment everything below this comment.
//       3. Delete this comment.

// import cerrs "cosmossdk.io/errors"

// // sanctionCodespace is the codespace for all errors defined in sanction package
// const sanctionCodespace = "sanction"

// var (
// 	ErrInvalidParams      = cerrs.Register(sanctionCodespace, 2, "invalid params")
// 	ErrUnsanctionableAddr = cerrs.Register(sanctionCodespace, 3, "address cannot be sanctioned")
// 	ErrInvalidTempStatus  = cerrs.Register(sanctionCodespace, 4, "invalid temp status")
// 	ErrSanctionedAccount  = cerrs.Register(sanctionCodespace, 5, "account is sanctioned")
// )
