package types

import (
	"context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	nametypes "github.com/provenance-io/provenance/x/name/types"
)

// AccountKeeper defines the expected account keeper used for simulations (noalias)
type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI
}

// NameKeeper defines the expected account keeper used for simulations (noalias)
type NameKeeper interface {
	Resolve(context.Context, *nametypes.QueryResolveRequest) (*nametypes.QueryResolveResponse, error)
	ResolvesTo(ctx sdk.Context, name string, addr sdk.AccAddress) bool
	Normalize(ctx sdk.Context, name string) (string, error)
}
