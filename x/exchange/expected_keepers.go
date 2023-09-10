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

type AttributeKeeper interface {
	GetAllAttributesAddr(ctx sdk.Context, addr []byte) ([]attrtypes.Attribute, error)
}

type BankKeeper interface {
	SendCoins(ctx sdk.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
}

type NameKeeper interface {
	Normalize(ctx sdk.Context, name string) (string, error)
}

type HoldKeeper interface {
	AddHold(ctx sdk.Context, addr sdk.AccAddress, funds sdk.Coins, reason string) error
	ReleaseHold(ctx sdk.Context, addr sdk.AccAddress, funds sdk.Coins) error
}
