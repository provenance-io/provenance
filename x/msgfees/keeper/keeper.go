package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/provenance-io/provenance/x/msgfees/types"
)

// Fee keeper calculates the additional fees to be charged
type AdditionalFeeKeeper interface {
	GetFeeRate(ctx sdk.Context) (feeRate sdk.Dec)
}

type (
	Keeper struct {
		// The reference to the Paramstore to get and set account specific params
		paramSpace paramtypes.Subspace
		cdc        codec.BinaryCodec
		storeKey   sdk.StoreKey
		memKey     sdk.StoreKey
	}
)

// NewKeeper returns a marker keeper. It handles:

//
// CONTRACT: the parameter Subspace must have the param key table already initialized
func NewKeeper(
	cdc codec.BinaryCodec,
	key sdk.StoreKey,
	paramSpace paramtypes.Subspace,
) Keeper {
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		paramSpace: paramSpace,
		storeKey:   key,
		cdc:        cdc,
	}
}
