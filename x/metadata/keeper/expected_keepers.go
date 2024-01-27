package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"

	markertypes "github.com/provenance-io/provenance/x/marker/types"
)

// AuthKeeper is an interface with functions that the auth.Keeper has that are needed in this module.
type AuthKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI
}

// AuthzKeeper is an interface with functions that the authz.Keeper has that are needed in this module.
type AuthzKeeper interface {
	GetAuthorization(ctx sdk.Context, grantee, granter sdk.AccAddress, msgType string) (authz.Authorization, *time.Time)
	DeleteGrant(ctx sdk.Context, grantee, granter sdk.AccAddress, msgType string) error
	SaveGrant(ctx sdk.Context, grantee, granter sdk.AccAddress, authorization authz.Authorization, expiration *time.Time) error
}

// AttrKeeper defines the attribute functionality needed by the metadata module.
type AttrKeeper interface {
	GetAccountData(ctx sdk.Context, addr string) (string, error)
	SetAccountData(ctx sdk.Context, addr string, value string) error
}

// MarkerKeeper defines the attribute functionality needed by the metadata module.
type MarkerKeeper interface {
	GetMarkerByDenom(ctx sdk.Context, denom string) (markertypes.MarkerAccountI, error)
}
