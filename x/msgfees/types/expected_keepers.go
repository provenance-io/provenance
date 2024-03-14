package types

import (
	"context"

	"cosmossdk.io/x/feegrant"

	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
)

// AccountKeeper defines the expected account keeper (noalias)
type AccountKeeper interface {
	IterateAccounts(ctx context.Context, process func(sdk.AccountI) (stop bool))
	GetAccount(context.Context, sdk.AccAddress) sdk.AccountI
	SetAccount(context.Context, sdk.AccountI)
	NewAccount(context.Context, sdk.AccountI) sdk.AccountI
}

// MsgFeesKeeper for additional msg fees.
type MsgFeesKeeper interface {
	GetMsgFee(ctx sdk.Context, msgType string) (*MsgFee, error)
	GetFeeCollectorName() string
	DeductFeesDistributions(bankKeeper bankkeeper.Keeper, ctx sdk.Context, acc sdk.AccountI, remainingFees sdk.Coins, fees map[string]sdk.Coins) error
	GetFloorGasPrice(ctx sdk.Context) sdk.Coin
	GetNhashPerUsdMil(ctx sdk.Context) uint64
	ConvertDenomToHash(ctx sdk.Context, coin sdk.Coin) (sdk.Coin, error)
	CalculateAdditionalFeesToBePaid(ctx sdk.Context, msgs ...sdk.Msg) (MsgFeesDistribution, error)
}

// FeegrantKeeper defines the expected feegrant keeper.
type FeegrantKeeper interface {
	GetAllowance(ctx context.Context, granter sdk.AccAddress, grantee sdk.AccAddress) (feegrant.FeeAllowanceI, error)
	UseGrantedFees(ctx context.Context, granter, grantee sdk.AccAddress, fee sdk.Coins, msgs []sdk.Msg) error
}
