package types

import (
	"context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	attrtypes "github.com/provenance-io/provenance/x/attribute/types"
)

// AccountKeeper defines the auth/account functionality needed by the marker keeper.
type AccountKeeper interface {
	GetAllAccounts(ctx sdk.Context) (accounts []authtypes.AccountI)
	GetNextAccountNumber(ctx sdk.Context) uint64
	GetAccount(sdk.Context, sdk.AccAddress) authtypes.AccountI
	SetAccount(sdk.Context, authtypes.AccountI)
	NewAccount(sdk.Context, authtypes.AccountI) authtypes.AccountI
	RemoveAccount(ctx sdk.Context, acc authtypes.AccountI)
}

// AuthzKeeper defines the authz functionality needed by the marker keeper.
type AuthzKeeper interface {
	GetAuthorization(ctx sdk.Context, grantee sdk.AccAddress, granter sdk.AccAddress, msgType string) (authz.Authorization, *time.Time)
	DeleteGrant(ctx sdk.Context, grantee sdk.AccAddress, granter sdk.AccAddress, msgType string) error
	SaveGrant(ctx sdk.Context, grantee, granter sdk.AccAddress, authorization authz.Authorization, expiration *time.Time) error
}

// BankKeeper defines the bank functionality needed by the marker module.
type BankKeeper interface {
	GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	GetSupply(ctx sdk.Context, denom string) sdk.Coin
	DenomOwners(goCtx context.Context, req *banktypes.QueryDenomOwnersRequest) (*banktypes.QueryDenomOwnersResponse, error)

	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error

	AppendSendRestriction(restriction banktypes.SendRestrictionFn)
	BlockedAddr(addr sdk.AccAddress) bool

	GetDenomMetaData(ctx sdk.Context, denom string) (banktypes.Metadata, bool)
	SetDenomMetaData(ctx sdk.Context, denomMetaData banktypes.Metadata)

	// TODO: Delete the below entries when no longer needed.

	// IterateAllBalances only used in GetAllMarkerHolders used by the unneeded querier.
	// The Holding query just uses the DenomOwners query endpoint.
	IterateAllBalances(ctx sdk.Context, cb func(address sdk.AccAddress, coin sdk.Coin) (stop bool))
	// GetAllSendEnabledEntries only needed by RemoveIsSendEnabledEntries in the quicksilver upgrade.
	GetAllSendEnabledEntries(ctx sdk.Context) []banktypes.SendEnabled
	// DeleteSendEnabled only needed by RemoveIsSendEnabledEntries in the quicksilver upgrade.
	DeleteSendEnabled(ctx sdk.Context, denom string)
}

// FeeGrantKeeper defines the fee-grant functionality needed by the marker module.
type FeeGrantKeeper interface {
	GrantAllowance(ctx sdk.Context, granter, grantee sdk.AccAddress, feeAllowance feegrant.FeeAllowanceI) error
}

// GovKeeper defines the gov functionality needed from within the gov module.
type GovKeeper interface {
	GetProposal(ctx sdk.Context, proposalID uint64) (govtypes.Proposal, bool)
	GetDepositParams(ctx sdk.Context) govtypes.DepositParams
	GetVotingParams(ctx sdk.Context) govtypes.VotingParams
	GetProposalID(ctx sdk.Context) (uint64, error)
}

// AttrKeeper defines the attribute functionality needed by the marker module.
type AttrKeeper interface {
	GetMaxValueLength(ctx sdk.Context) uint32
	GetAllAttributesAddr(ctx sdk.Context, addr []byte) ([]attrtypes.Attribute, error)
}
