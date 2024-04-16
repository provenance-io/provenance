package quarantine

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// AccountKeeper defines the account/auth functionality needed from within the quarantine module.
type AccountKeeper interface {
	NewAccount(context.Context, sdk.AccountI) sdk.AccountI
	GetAccount(context.Context, sdk.AccAddress) sdk.AccountI
	SetAccount(context.Context, sdk.AccountI)
}

// BankKeeper defines the bank functionality needed from within the quarantine module.
type BankKeeper interface {
	AppendSendRestriction(restriction banktypes.SendRestrictionFn)
	GetAllBalances(context context.Context, addr sdk.AccAddress) sdk.Coins
	SendCoins(context context.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
	SpendableCoins(context context.Context, addr sdk.AccAddress) sdk.Coins
}
