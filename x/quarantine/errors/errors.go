package errors

import sdkqerrs "github.com/cosmos/cosmos-sdk/x/quarantine/errors"

var ErrInvalidValue = sdkqerrs.ErrInvalidValue

// TODO[1760]: Once we have an official version of the SDK without the quarantine module:
//             1. Delete everything above this comment.
//             2. Uncomment everything below this comment.
//             3. Delete this comment.

// package errors
//
// import cerrs "cosmossdk.io/errors"
//
// // quarantineCodespace is the codespace for all errors defined in quarantine package
// const quarantineCodespace = "quarantine"
//
// var ErrInvalidValue = cerrs.Register(quarantineCodespace, 2, "invalid value")
