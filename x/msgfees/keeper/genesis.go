package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/msgfees/types"
)

// ExportGenesis returns a GenesisState for a given context.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	msgFees := make([]types.MsgFee, 0)
	params := k.GetParams(ctx)
	msgFeeRecords := func(msgFee types.MsgFee) bool {
		msgFees = append(msgFees, msgFee)
		return false
	}
	if err := k.IterateMsgFees(ctx, msgFeeRecords); err != nil {
		panic(err)
	}
	return types.NewGenesisState(params, msgFees)
}

// InitGenesis new msgfees genesis
func (k Keeper) InitGenesis(ctx sdk.Context, data *types.GenesisState) {
	k.SetParams(ctx, data.Params)

	if err := data.Validate(); err != nil {
		panic(err)
	}
	for _, msgFee := range data.MsgFees {
		if err := k.SetMsgFee(ctx, msgFee); err != nil {
			panic(err)
		}
	}
}
