package app

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/group"
)

// GroupCheckerFunc convenient type to match the GroupChecker interface.
type GroupCheckerFunc func(sdk.Context, sdk.AccAddress) bool

// GroupPolicyQuerier provides functionality to query group policies.
type GroupPolicyQuerier interface {
	GroupPolicyInfo(goCtx context.Context, request *group.QueryGroupPolicyInfoRequest) (*group.QueryGroupPolicyInfoResponse, error)
}

// IsGroupAddress checks if the account is a group address.
func (t GroupCheckerFunc) IsGroupAddress(ctx sdk.Context, account sdk.AccAddress) bool {
	if account == nil {
		return false
	}
	return t(ctx, account)
}

// NewGroupCheckerFunc creates a new GroupChecker function for checking if an account is in a group.
func NewGroupCheckerFunc(querier GroupPolicyQuerier) GroupCheckerFunc {
	return GroupCheckerFunc(func(ctx sdk.Context, account sdk.AccAddress) bool {
		msg := &group.QueryGroupPolicyInfoRequest{Address: account.String()}
		goCtx := sdk.WrapSDKContext(ctx)
		_, err := querier.GroupPolicyInfo(goCtx, msg)
		return err == nil
	})
}
