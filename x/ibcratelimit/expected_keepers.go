package ibcratelimit

import sdk "github.com/cosmos/cosmos-sdk/types"

// PermissionedKeeper defines the expected interface for a keeper with permission checks.
type PermissionedKeeper interface {
	Sudo(ctx sdk.Context, contractAddress sdk.AccAddress, msg []byte) ([]byte, error)
}
