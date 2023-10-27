package ibcratelimit

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
)

const IbcAcknowledgementErrorType = "ibc-acknowledgement-error"

// NewEmitErrorAcknowledgement creates a new error acknowledgement after having emitted an event with the
// details of the error.
func NewEmitErrorAcknowledgement(ctx sdk.Context, err error, errorContexts ...string) channeltypes.Acknowledgement {
	EmitIBCErrorEvents(ctx, err, errorContexts)

	return channeltypes.NewErrorAcknowledgement(err)
}

// EmitIBCErrorEvents Emit and Log errors
func EmitIBCErrorEvents(ctx sdk.Context, err error, errorContexts []string) {
	logger := ctx.Logger().With("module", IbcAcknowledgementErrorType)

	attributes := make([]sdk.Attribute, len(errorContexts)+1)
	attributes[0] = sdk.NewAttribute("error", err.Error())
	for i, s := range errorContexts {
		attributes[i+1] = sdk.NewAttribute("error-context", s)
		logger.Error(fmt.Sprintf("error-context: %v", s))
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			IbcAcknowledgementErrorType,
			attributes...,
		),
	})
}

// IsAckError checks an IBC acknowledgement to see if it's an error.
// This is a replacement for ack.Success() which is currently not working on some circumstances
func IsAckError(acknowledgement []byte) bool {
	var ackErr channeltypes.Acknowledgement_Error
	if err := json.Unmarshal(acknowledgement, &ackErr); err == nil && len(ackErr.Error) > 0 {
		return true
	}
	return false
}
