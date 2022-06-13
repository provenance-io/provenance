package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
)

// AccountKeeper defines the expected account keeper (noalias)
type AccountKeeper interface {
	IterateAccounts(ctx sdk.Context, process func(authtypes.AccountI) (stop bool))
	GetAccount(sdk.Context, sdk.AccAddress) authtypes.AccountI
	SetAccount(sdk.Context, authtypes.AccountI)
	NewAccount(sdk.Context, authtypes.AccountI) authtypes.AccountI
}

// MsgFeesKeeper for additional msg fees.
type MsgFeesKeeper interface {
	GetMsgFee(ctx sdk.Context, msgType string) (*MsgFee, error)
	GetFeeCollectorName() string
	DeductFeesDistributions(bankKeeper bankkeeper.Keeper, ctx sdk.Context, acc authtypes.AccountI, remainingFees sdk.Coins, fees map[string]sdk.Coins) error
	GetFloorGasPrice(ctx sdk.Context) sdk.Coin
	GetNhashPerUsdMil(ctx sdk.Context) uint64
	ConvertDenomToHash(ctx sdk.Context, coin sdk.Coin) (sdk.Coin, error)
}

// FeegrantKeeper defines the expected feegrant keeper.
type FeegrantKeeper interface {
	GetAllowance(ctx sdk.Context, granter sdk.AccAddress, grantee sdk.AccAddress) (feegrant.FeeAllowanceI, error)
	UseGrantedFees(ctx sdk.Context, granter, grantee sdk.AccAddress, fee sdk.Coins, msgs []sdk.Msg) error
}
