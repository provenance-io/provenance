package keeper

import (
	"context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/authz"

	markertypes "github.com/provenance-io/provenance/x/marker/types"
	"github.com/provenance-io/provenance/x/metadata/types"
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

// MarkerKeeper defines the attribute functionality needed by the metadata module.
type MarkerKeeper interface {
	GetMarkerByDenom(ctx sdk.Context, denom string) (markertypes.MarkerAccountI, error)
}

type BankKeeper interface {
	BlockedAddr(addr sdk.AccAddress) bool
	MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	SendCoins(ctx context.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) error
	SpendableCoin(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin

	// These are methods not in the bank keeper, but that we add using our own MDBankKeeper.

	DenomOwner(ctx context.Context, denom string) (sdk.AccAddress, error)
	GetScopesForValueOwner(ctx context.Context, valueOwner sdk.AccAddress, pageReq *query.PageRequest) (types.AccMDLinks, *query.PageResponse, error)
}
