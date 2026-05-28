package keeper

import (
	"context"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/provenance-io/provenance/x/quarantine"
)

var _ quarantine.MsgServer = Keeper{}

// errQuarantineRemoved is returned by all msg and query handlers.
// The quarantine module has been deactivated and its endpoints are no longer available.
const errQuarantineRemoved = "quarantine module has been removed"

func (k Keeper) OptIn(_ context.Context, _ *quarantine.MsgOptIn) (*quarantine.MsgOptInResponse, error) {
	return nil, sdkerrors.ErrNotSupported.Wrap(errQuarantineRemoved)
}

func (k Keeper) OptOut(_ context.Context, _ *quarantine.MsgOptOut) (*quarantine.MsgOptOutResponse, error) {
	return nil, sdkerrors.ErrNotSupported.Wrap(errQuarantineRemoved)
}

func (k Keeper) Accept(_ context.Context, _ *quarantine.MsgAccept) (*quarantine.MsgAcceptResponse, error) {
	return nil, sdkerrors.ErrNotSupported.Wrap(errQuarantineRemoved)
}

func (k Keeper) Decline(_ context.Context, _ *quarantine.MsgDecline) (*quarantine.MsgDeclineResponse, error) {
	return nil, sdkerrors.ErrNotSupported.Wrap(errQuarantineRemoved)
}

func (k Keeper) UpdateAutoResponses(_ context.Context, _ *quarantine.MsgUpdateAutoResponses) (*quarantine.MsgUpdateAutoResponsesResponse, error) {
	return nil, sdkerrors.ErrNotSupported.Wrap(errQuarantineRemoved)
}
