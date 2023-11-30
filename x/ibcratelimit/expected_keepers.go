package ibcratelimit

import sdk "github.com/cosmos/cosmos-sdk/types"

type PermissionedKeeper interface {
	Sudo(ctx sdk.Context, contractAddress sdk.AccAddress, msg []byte) ([]byte, error)
}
