package app

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/group"
	groupkeeper "github.com/cosmos/cosmos-sdk/x/group/keeper"
)

// GroupCheckerFunc convenient type to match the GroupChecker interface.
type GroupCheckerFunc func(sdk.Context, sdk.AccAddress) bool

// IsGroupAddress checks if the account is a group address.
func (t GroupCheckerFunc) IsGroupAddress(ctx sdk.Context, account sdk.AccAddress) bool {
	return t(ctx, account)
}

// NewGroupChecker creates a new GroupChecker function for checking if an account is in a group.
func NewGroupCheckerFunc(keeper groupkeeper.Keeper) GroupCheckerFunc {
	return GroupCheckerFunc(func(ctx sdk.Context, account sdk.AccAddress) bool {
		msg := &group.QueryGroupPolicyInfoRequest{Address: account.String()}
		goCtx := sdk.WrapSDKContext(ctx)
		_, err := keeper.GroupPolicyInfo(goCtx, msg)
		return err == nil
	})
}
