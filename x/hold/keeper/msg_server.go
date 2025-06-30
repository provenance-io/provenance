package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"

	"github.com/provenance-io/provenance/x/hold"
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
	accountsToSave := make([]sdk.AccountI, 0, len(req.Addresses))
	addressToIndexMap := make(map[string]int)

	for _, addrStr := range req.Addresses {
		addr, err := sdk.AccAddressFromBech32(addrStr)
		if err != nil {
			failedAddresses = append(failedAddresses, &types.UnlockFailure{
				Address: addrStr,
				Reason:  fmt.Sprintf("invalid address format: %s", err),
			})
			continue
		}

		baseAccount, err := s.unlockVestingAccount(ctx, addr)
		if err != nil {
			failedAddresses = append(failedAddresses, &types.UnlockFailure{
				Address: addrStr,
				Reason:  err.Error(),
			})
			continue
		}

		accountsToSave = append(accountsToSave, baseAccount)
		addressToIndexMap[addrStr] = len(accountsToSave) - 1
		unlockedAddresses = append(unlockedAddresses, addrStr)
	}

	for _, acc := range accountsToSave {
		s.accountKeeper.SetAccount(ctx, acc)
	}
	if err := ctx.EventManager().EmitTypedEvent(hold.NewEventUnlockVestingAccounts(sdk.AccAddress(s.authority), uint32(len(unlockedAddresses)), uint32(len(failedAddresses)))); err != nil {
		return nil, err
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
		return sdkerrors.ErrUnauthorized.Wrapf("invalid authority; expected %s, got %s",
			govModuleAddr,
			authority)
	}
	return nil
}

// unlockVestingAccount converts a vesting account back to a base account
func (s msgServer) unlockVestingAccount(ctx sdk.Context, addr sdk.AccAddress) (*authtypes.BaseAccount, error) {
	account := s.accountKeeper.GetAccount(ctx, addr)
	if account == nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrap(addr.String())
	}
	// Extract base account directly
	var baseVestAcct *vesting.BaseVestingAccount
	switch acct := account.(type) {
	case *vesting.ContinuousVestingAccount:
		baseVestAcct = acct.BaseVestingAccount
	case *vesting.DelayedVestingAccount:
		baseVestAcct = acct.BaseVestingAccount
	case *vesting.PeriodicVestingAccount:
		baseVestAcct = acct.BaseVestingAccount
	default:
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("account %s is type %T, not a supported vesting account type", addr.String(), account)
	}
	if baseVestAcct == nil {
		return nil, sdkerrors.ErrInvalidType.Wrapf("failed to extract BaseVestingAccount from vesting account: %s", addr.String())
	}

	if baseVestAcct.BaseAccount == nil {
		return nil, sdkerrors.ErrInvalidType.Wrapf("BaseVestingAccount.BaseAccount is nil for account: %s", addr.String())
	}
	ctx.Logger().Info("Unlocking vesting account",
		"address", addr.String(),
		"original_type", fmt.Sprintf("%T", account),
		"account_number", baseVestAcct.BaseAccount.AccountNumber,
		"sequence", baseVestAcct.BaseAccount.Sequence)
	return baseVestAcct.BaseAccount, nil
}
