package antewrapper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
)

// MsgTypeBlacklistContextDecorator is (hopefully) a temporary hard-coded antehandler that disallows certain messages.
// Once the circuit breaker module is added to provenance, this should be removed.
type MsgTypeBlacklistContextDecorator struct {
	Blacklist []string
}

func NewMsgTypeBlacklistContextDecorator() MsgTypeBlacklistContextDecorator {
	return MsgTypeBlacklistContextDecorator{
		Blacklist: []string{
			// Disallow vesting account creation due to barberry: https://forum.cosmos.network/t/cosmos-sdk-security-advisory-barberry/10825
			// Once that fix is in the SDK that we pull in, these can be removed.
			// MsgCreatePeriodicVestingAccount is specific to barberry, the other two are due to extra caution.
			sdk.MsgTypeURL(&vestingtypes.MsgCreatePeriodicVestingAccount{}),
			sdk.MsgTypeURL(&vestingtypes.MsgCreateVestingAccount{}),
			sdk.MsgTypeURL(&vestingtypes.MsgCreatePermanentLockedAccount{}),
		},
	}
}

var _ sdk.AnteDecorator = MsgTypeBlacklistContextDecorator{}

func (b MsgTypeBlacklistContextDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	for _, msg := range tx.GetMsgs() {
		msgT := sdk.MsgTypeURL(msg)
		for _, nope := range b.Blacklist {
			if msgT == nope {
				return ctx, sdkerrors.ErrInvalidRequest.Wrapf("%s messages are not allowed", msgT)
			}
		}
	}
	return next(ctx, tx, simulate)
}
