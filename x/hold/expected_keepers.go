package hold

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	// banktypes "github.com/cosmos/cosmos-sdk/x/bank/types" // TODO[1760]: locked-coins
)

type BankKeeper interface {
	// AppendLockedCoinsGetter(getter banktypes.GetLockedCoinsFn) // TODO[1760]: locked-coins
	SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins
}
