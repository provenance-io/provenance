package exchange

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	attrtypes "github.com/provenance-io/provenance/x/attribute/types"
)

type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI
	SetAccount(ctx sdk.Context, acc authtypes.AccountI)
	HasAccount(ctx sdk.Context, addr sdk.AccAddress) bool
	NewAccount(ctx sdk.Context, acc authtypes.AccountI) authtypes.AccountI
}

type NameKeeper interface {
	Normalize(ctx sdk.Context, name string) (string, error)
}

type AttributeKeeper interface {
	GetAllAttributesAddr(ctx sdk.Context, addr []byte) ([]attrtypes.Attribute, error)
}
