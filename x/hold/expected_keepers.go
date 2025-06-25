package hold

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

type BankKeeper interface {
	AppendLockedCoinsGetter(getter banktypes.GetLockedCoinsFn)
	SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins
}
type AccountKeeper interface {
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	SetAccount(ctx context.Context, acc sdk.AccountI)
	HasAccount(ctx context.Context, addr sdk.AccAddress) bool
	NewAccount(ctx context.Context, acc sdk.AccountI) sdk.AccountI
}
