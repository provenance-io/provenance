package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ParamSubspace defines the expected Subspace interface for parameters (noalias)
type ParamSubspace interface {
	Get(ctx sdk.Context, key []byte, ptr interface{})
	Set(ctx sdk.Context, key []byte, param interface{})
}

// AttributeKeeper defines the expected attribute keeper interface (noalias)
type AttributeKeeper interface {
	DeleteAttribute(ctx sdk.Context, addr string, name string, value *[]byte, owner sdk.AccAddress) error
	AccountsByAttribute(ctx sdk.Context, name string) (addresses []sdk.AccAddress, err error)
}
