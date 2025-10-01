package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/flatfees/types"
)

// ExportGenesis returns a GenesisState for a given context.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	params := k.GetParams(ctx)
	msgFees := make([]*types.MsgFee, 0)
	err := k.msgFees.Walk(ctx, nil, func(_ string, msgFee types.MsgFee) (bool, error) {
		msgFees = append(msgFees, &msgFee)
		return false, nil
	})
	if err != nil {
		panic(err)
	}
	return types.NewGenesisState(params, msgFees)
}

// InitGenesis new x/flatfees genesis
func (k Keeper) InitGenesis(ctx sdk.Context, data *types.GenesisState) {
	if err := data.Validate(); err != nil {
		panic(err)
	}

	if err := k.SetParams(ctx, data.Params); err != nil {
		panic(err)
	}

	for _, msgFee := range data.MsgFees {
		if err := k.SetMsgFee(ctx, *msgFee); err != nil {
			panic(err)
		}
	}
}
