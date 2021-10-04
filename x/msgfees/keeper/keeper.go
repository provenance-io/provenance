package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/gogo/protobuf/proto"
	"github.com/provenance-io/provenance/x/msgfees/types"
	"github.com/tendermint/tendermint/libs/log"
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

func (k Keeper) GetMsgBasedFeeSchedule(ctx sdk.Context, msgType string) (*types.MsgFees, error) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetMsgBasedFeeKey(msgType)
	bz := store.Get(key)
	if len(bz) == 0 {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "msg-based-fees not found")
	}

	var msgBasedFee types.MsgFees
	if err := k.cdc.Unmarshal(bz, &msgBasedFee); err != nil {
		return nil, err
	}

	return &msgBasedFee, nil
}

type Handler func(record types.MsgFees) (stop bool)

// IterateMarkers  iterates all markers with the given handler function.
func (k Keeper) IterateMsgFees(ctx sdk.Context, handle func(msgFees types.MsgFees) (stop bool)) error {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.MsgBasedFeeKeyPrefix)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		record := types.MsgFees{}
		if err := k.cdc.Unmarshal(iterator.Value(), &record); err != nil {
			return err
		}
		if handle(record) {
			break
		}
	}
	return nil
}



