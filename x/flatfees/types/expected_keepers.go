package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AccountKeeper defines the expected account keeper (noalias)
type AccountKeeper interface {
	IterateAccounts(ctx context.Context, process func(sdk.AccountI) (stop bool))
	GetAccount(context.Context, sdk.AccAddress) sdk.AccountI
	SetAccount(context.Context, sdk.AccountI)
	NewAccount(context.Context, sdk.AccountI) sdk.AccountI
}

// BaseAppSimulateFunc is the SimulateProv method on the baseapp, used to simulate a tx for gas and costs.
type BaseAppSimulateFunc func(txBytes []byte) (sdk.GasInfo, *sdk.Result, sdk.Context, error)
