package types

import (
	"context"
	"time"

	"cosmossdk.io/x/feegrant"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"

	attrtypes "github.com/provenance-io/provenance/x/attribute/types"
)

// AccountKeeper defines the auth/account functionality needed by the marker keeper.
type AccountKeeper interface {
	GetAllAccounts(context.Context) []sdk.AccountI
	NextAccountNumber(context.Context) uint64
	GetAccount(context.Context, sdk.AccAddress) sdk.AccountI
	SetAccount(context.Context, sdk.AccountI)
	NewAccount(context.Context, sdk.AccountI) sdk.AccountI
	RemoveAccount(context.Context, sdk.AccountI)
}

// AuthzKeeper defines the authz functionality needed by the marker keeper.
type AuthzKeeper interface {
	GetAuthorization(ctx context.Context, grantee sdk.AccAddress, granter sdk.AccAddress, msgType string) (authz.Authorization, *time.Time)
	DeleteGrant(ctx context.Context, grantee sdk.AccAddress, granter sdk.AccAddress, msgType string) error
	SaveGrant(ctx context.Context, grantee, granter sdk.AccAddress, authorization authz.Authorization, expiration *time.Time) error
}

// BankKeeper defines the bank functionality needed by the marker module.
type BankKeeper interface {
	GetAllBalances(context context.Context, addr sdk.AccAddress) sdk.Coins
	GetBalance(context context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	GetSupply(context context.Context, denom string) sdk.Coin
	DenomOwners(context context.Context, req *banktypes.QueryDenomOwnersRequest) (*banktypes.QueryDenomOwnersResponse, error)

	SendCoins(context context.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(context context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(context context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	MintCoins(context context.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(context context.Context, moduleName string, amt sdk.Coins) error

	AppendSendRestriction(restriction banktypes.SendRestrictionFn)
	BlockedAddr(addr sdk.AccAddress) bool

	GetDenomMetaData(context context.Context, denom string) (banktypes.Metadata, bool)
	SetDenomMetaData(context context.Context, denomMetaData banktypes.Metadata)
}

// FeeGrantKeeper defines the fee-grant functionality needed by the marker module.
type FeeGrantKeeper interface {
	GrantAllowance(ctx context.Context, granter, grantee sdk.AccAddress, feeAllowance feegrant.FeeAllowanceI) error
	IterateAllFeeAllowances(ctx context.Context, cb func(grant feegrant.Grant) bool) error
}

// Note: There is no IBCKeeper interface in here.
// The SendTransfer function takes in a checkRestrictionsHandler. That is defined in the
// ibc keeper package. Furthermore, checkRestrictionsHandler takes in an IBC Keeper anyway.
// So there's no way to remove the dependency on that ibc keeper package.

// AttrKeeper defines the attribute functionality needed by the marker module.
type AttrKeeper interface {
	GetMaxValueLength(ctx sdk.Context) uint32
	GetAllAttributesAddr(ctx sdk.Context, addr []byte) ([]attrtypes.Attribute, error)
	GetAccountData(ctx sdk.Context, addr string) (string, error)
	SetAccountData(ctx sdk.Context, addr string, value string) error
}

// NameKeeper defines the name keeper functionality needed by the marker module.
type NameKeeper interface {
	Normalize(ctx sdk.Context, name string) (string, error)
}

// IbcTransferMsgServer defines the message server functionality needed by the marker module.
type IbcTransferMsgServer interface {
	Transfer(goCtx context.Context, msg *transfertypes.MsgTransfer) (*transfertypes.MsgTransferResponse, error)
}

// GroupChecker defines the functionality for checking if an account is part of a group.
type GroupChecker interface {
	IsGroupAddress(sdk.Context, sdk.AccAddress) bool
}
