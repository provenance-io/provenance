package keeper

import (
	"context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
)

// AuthKeeper is an interface with functions that the auth.Keeper has that are needed in this module.
type AuthKeeper interface {
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
}

// AuthzKeeper is an interface with functions that the authz.Keeper has that are needed in this module.
type AuthzKeeper interface {
	GetAuthorization(ctx context.Context, grantee, granter sdk.AccAddress, msgType string) (authz.Authorization, *time.Time)
	DeleteGrant(ctx context.Context, grantee, granter sdk.AccAddress, msgType string) error
	SaveGrant(ctx context.Context, grantee, granter sdk.AccAddress, authorization authz.Authorization, expiration *time.Time) error
}

// AttrKeeper defines the attribute functionality needed by the metadata module.
type AttrKeeper interface {
	GetAccountData(ctx sdk.Context, addr string) (string, error)
	SetAccountData(ctx sdk.Context, addr string, value string) error
}
