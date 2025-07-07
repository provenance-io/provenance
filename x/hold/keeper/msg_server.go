package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/hold"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the hold MsgServer interface
func NewMsgServerImpl(keeper Keeper) hold.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ hold.MsgServer = msgServer{}

// UnlockVestingAccounts converts vesting accounts back to base accounts
// This is a governance-only endpoint for security
func (s msgServer) UnlockVestingAccounts(goCtx context.Context, req *hold.MsgUnlockVestingAccountsRequest) (*hold.MsgUnlockVestingAccountsResponse, error) {
	if err := s.ValidateAuthority(req.Authority); err != nil {
		return nil, err
	}

	// If this endpoint were to return an error, it'd undo everything it did and mark the proposal as failed.
	// But we don't want that here. We want to keep any work that it did. So we can just ignore this error.
	_ = s.Keeper.UnlockVestingAccounts(sdk.UnwrapSDKContext(goCtx), req.Addresses)

	return &hold.MsgUnlockVestingAccountsResponse{}, nil
}
