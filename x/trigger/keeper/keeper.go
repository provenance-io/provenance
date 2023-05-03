package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/trigger/types"
)

const StoreKey = types.ModuleName

type Keeper struct {
	storeKey storetypes.StoreKey
	cdc      codec.BinaryCodec
}

func NewKeeper(
	cdc codec.BinaryCodec,
	key storetypes.StoreKey,
) Keeper {
	return Keeper{
		storeKey: key,
		cdc:      cdc,
	}
}

func (k Keeper) SetTrigger(ctx sdk.Context, trigger types.Trigger) error {
	return nil
}

func (k Keeper) GetTrigger(ctx sdk.Context, id types.TriggerID) (*types.Trigger, error) {
	return nil, nil
}

func (k Keeper) GetNextTriggerID(ctx sdk.Context) (types.TriggerID, error) {
	return 0, nil
}

// Get
// Set
// Iterate
