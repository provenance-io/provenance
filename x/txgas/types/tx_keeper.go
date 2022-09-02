package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

type TxKeeper struct {
	// The reference to the Paramstore to get and set attribute specific params
	paramSubspace paramtypes.Subspace
}

func (tk TxKeeper) GetParams(ctx sdk.Context) (params Params) {
	return Params {
		TxGasLimit: tk.GetTxGasLimit(ctx),
	}
}

// SetParams sets the account parameters to the param space.
func (tk TxKeeper) SetParams(ctx sdk.Context, params Params) {
	tk.paramSubspace.SetParamSet(ctx, &params)
}

func (tk TxKeeper) GetTxGasLimit(ctx sdk.Context) (txGasLimit uint64) {
	txGasLimit = DefaultTxGasLimit
	if tk.paramSubspace.Has(ctx, ParamStoreKeyTxGasLimit) {
		tk.paramSubspace.Get(ctx, ParamStoreKeyTxGasLimit, &txGasLimit)
	}
	return
}