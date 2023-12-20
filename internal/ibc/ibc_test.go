package ibc_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"

	"github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/internal/ibc"
	"github.com/provenance-io/provenance/x/ibcratelimit"
)

func TestNewEmitErrorAcknowledgement(t *testing.T) {
	testCases := []struct {
		name      string
		err       error
		errCtx    []string
		hasEvents bool
		ack       channeltypes.Acknowledgement
	}{
		{
			name:      "success - emits ibc error events",
			err:       ibcratelimit.ErrRateLimitExceeded,
			errCtx:    []string{"err ctx 1", "error ctx 2"},
			hasEvents: true,
			ack:       channeltypes.NewErrorAcknowledgement(ibcratelimit.ErrRateLimitExceeded),
		},
		{
			name:      "success - no ctx",
			err:       ibcratelimit.ErrRateLimitExceeded,
			errCtx:    []string{},
			hasEvents: true,
			ack:       channeltypes.NewErrorAcknowledgement(ibcratelimit.ErrRateLimitExceeded),
		},
		{
			name:      "success - nil ctx",
			err:       ibcratelimit.ErrRateLimitExceeded,
			errCtx:    nil,
			hasEvents: true,
			ack:       channeltypes.NewErrorAcknowledgement(ibcratelimit.ErrRateLimitExceeded),
		},
		{
			name:      "success - nil error",
			err:       nil,
			errCtx:    []string{"err ctx 1", "error ctx 2"},
			hasEvents: false,
			ack:       channeltypes.NewErrorAcknowledgement(nil),
		},
	}

	testApp := app.Setup(t)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := testApp.BaseApp.NewContext(false)
			ack := ibc.NewEmitErrorAcknowledgement(ctx, tc.err, tc.errCtx...)
			events := ctx.EventManager().Events()
			assert.Equal(t, tc.hasEvents, len(events) > 0, "should correctly decide when to emit events")
			assert.Equal(t, tc.ack, ack, "should return the correct ack")
		})
	}
}

func TestEmitIBCErrorEvents(t *testing.T) {
	testCases := []struct {
		name   string
		err    error
		errCtx []string
		events sdk.Events
	}{
		{
			name:   "success - emits ibc error events",
			err:    ibcratelimit.ErrRateLimitExceeded,
			errCtx: []string{"err ctx 1", "error ctx 2"},
			events: []sdk.Event{
				sdk.NewEvent(ibc.IbcAcknowledgementErrorType,
					sdk.NewAttribute("error", "rate limit exceeded"),
					sdk.NewAttribute("error-context", "err ctx 1"),
					sdk.NewAttribute("error-context", "error ctx 2"),
				),
			},
		},
		{
			name:   "success - no ctx",
			err:    ibcratelimit.ErrRateLimitExceeded,
			errCtx: []string{},
			events: []sdk.Event{
				sdk.NewEvent(ibc.IbcAcknowledgementErrorType,
					sdk.NewAttribute("error", "rate limit exceeded"),
				),
			},
		},
		{
			name:   "success - nil ctx",
			err:    ibcratelimit.ErrRateLimitExceeded,
			errCtx: nil,
			events: []sdk.Event{
				sdk.NewEvent(ibc.IbcAcknowledgementErrorType,
					sdk.NewAttribute("error", "rate limit exceeded"),
				),
			},
		},
		{
			name:   "success - nil error",
			err:    nil,
			errCtx: []string{"err ctx 1", "error ctx 2"},
			events: []sdk.Event{},
		},
	}

	testApp := app.Setup(t)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := testApp.BaseApp.NewContext(false)
			ibc.EmitIBCErrorEvents(ctx, tc.err, tc.errCtx)
			events := ctx.EventManager().Events()
			assert.Equal(t, tc.events, events, "should emit the correct events")
		})
	}
}

func TestIsAckError(t *testing.T) {
	testCases := []struct {
		name     string
		ack      channeltypes.Acknowledgement
		expected bool
	}{
		{
			name:     "success - should detect error ack",
			ack:      channeltypes.NewErrorAcknowledgement(ibcratelimit.ErrRateLimitExceeded),
			expected: true,
		},
		{
			name:     "failure - should detect result ack",
			ack:      channeltypes.NewResultAcknowledgement([]byte("garbage")),
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ack, err := json.Marshal(tc.ack.Response)
			assert.NoError(t, err, "should not fail when marshaling ack")
			isAck := ibc.IsAckError(ack)
			assert.Equal(t, tc.expected, isAck, "should return the correct value")
		})
	}
}
