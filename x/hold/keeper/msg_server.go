package keeper

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/types/errors"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/hold/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the hold MsgServer interface
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// UnlockVestingAccounts converts vesting accounts back to base accounts
// This is a governance-only endpoint for security
func (s msgServer) UnlockVestingAccounts(goCtx context.Context, req *types.MsgUnlockVestingAccountsRequest) (*types.MsgUnlockVestingAccountsResponse, error) {

	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := s.validateAuthority(req.Authority); err != nil {
		return nil, err
	}

	var unlockedAddresses []string
	var failedAddresses []*types.UnlockFailure
	var failedAddrs []string

	// Process each address - continue on failures as per requirement
	for _, addrStr := range req.Addresses {
		addr, err := sdk.AccAddressFromBech32(addrStr)
		if err != nil {
			failure := &types.UnlockFailure{
				Address: addrStr,
				Reason:  fmt.Sprintf("invalid address format: %s", err),
			}
			failedAddresses = append(failedAddresses, failure)
			failedAddrs = append(failedAddrs, addrStr)
			continue
		}

		// Attempt to unlock the account
		if err := s.unlockVestingAccount(ctx, addr); err != nil {
			failure := &types.UnlockFailure{
				Address: addrStr,
				Reason:  err.Error(),
			}
			failedAddresses = append(failedAddresses, failure)
			failedAddrs = append(failedAddrs, addrStr)
			continue
		}

		unlockedAddresses = append(unlockedAddresses, addrStr)
	}

	return &types.MsgUnlockVestingAccountsResponse{
		UnlockedAddresses: unlockedAddresses,
		FailedAddresses:   failedAddresses,
	}, nil
}

// validateAuthority ensures only governance module can call this function
func (s msgServer) validateAuthority(authority string) error {
	govModuleAddr := s.GetAuthority()
	if authority != govModuleAddr {
		return errors.ErrUnauthorized.Wrapf(
			"invalid authority; expected %s, got %s",
			govModuleAddr,
			authority,
		)
	}
	return nil
}

// unlockVestingAccount converts a vesting account back to a base account
func (s msgServer) unlockVestingAccount(ctx sdk.Context, addr sdk.AccAddress) error {
	// Get the account
	account := s.accountKeeper.GetAccount(ctx, addr)
	if account == nil {
		return fmt.Errorf("account not found: %s", addr.String())
	}

	// Check if it's a vesting account
	vestingAccount, ok := account.(*vestingtypes.BaseVestingAccount)
	if !ok {
		return fmt.Errorf("account %s is type %T, not vesting", addr.String(), account)
	}

	ctx.Logger().Info(
		"unlocking vesting account",
		"address", addr.String(),
		"original_type", fmt.Sprintf("%T", vestingAccount),
		"account_number", account.GetAccountNumber(),
	)

	// Get the base account inside the vesting wrapper
	baseAccount := vestingAccount.BaseAccount
	if baseAccount == nil {
		return fmt.Errorf("failed to extract base account from vesting account: %s", addr.String())
	}

	s.accountKeeper.SetAccount(ctx, baseAccount)

	ctx.Logger().Info(
		"successfully unlocked vesting account",
		"address", addr.String(),
		"account_number", baseAccount.GetAccountNumber(),
		"sequence", baseAccount.GetSequence(),
	)

	return nil
}
