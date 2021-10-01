package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/provenance-io/provenance/x/msgfees/types"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/gogo/protobuf/proto"
)

// MsgBasedFeeKeeperI Fee keeper calculates the additional fees to be charged
type MsgBasedFeeKeeperI interface {
	GetFeeRate(ctx sdk.Context) (feeRate sdk.Dec)
}

// Keeper of the Additional fee store
type Keeper struct {
	storeKey         sdk.StoreKey
	cdc              codec.BinaryCodec
	paramSpace       paramtypes.Subspace
	feeCollectorName string // name of the FeeCollector ModuleAccount
}

// NewKeeper returns a AdditionalFeeKeeper. It handles:
// CONTRACT: the parameter Subspace must have the param key table already initialized
func NewKeeper(
	cdc codec.BinaryCodec,
	key sdk.StoreKey,
	paramSpace paramtypes.Subspace,
	feeCollectorName string,
) Keeper {
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		storeKey:         key,
		cdc:              cdc,
		paramSpace:       paramSpace,
		feeCollectorName: feeCollectorName,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// SetMsgBasedFeeSchedule sets the additional fee schedule for a Msg
func (k Keeper) SetMsgBasedFeeSchedule(ctx sdk.Context, msgBasedFees types.MsgFees) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&msgBasedFees)
	proto.MessageName(msgBasedFees.Msg)
	store.Set(types.GetMsgBasedFeeKey(proto.MessageName(msgBasedFees.Msg)), bz)
}
