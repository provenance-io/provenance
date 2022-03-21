package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/reward/types"
)

const StoreKey = types.ModuleName

type (
	Keeper struct {
		cdc      codec.Codec
		storeKey sdk.StoreKey
	}
)

func NewKeeper(cdc codec.Codec, storeKey sdk.StoreKey) *Keeper {
	return &Keeper{
		cdc:      cdc,
		storeKey: storeKey,
	}
}
