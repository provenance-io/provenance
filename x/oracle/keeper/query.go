package keeper

import (
	gogotypes "github.com/gogo/protobuf/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/oracle/types"
)

// GetLastQueryPacketSeq return the id from the last query request
func (k Keeper) GetLastQueryPacketSeq(ctx sdk.Context) uint64 {
	bz := ctx.KVStore(k.storeKey).Get(types.GetLastQueryPacketSeqKey())
	uintV := gogotypes.UInt64Value{}
	k.cdc.MustUnmarshalLengthPrefixed(bz, &uintV)
	return uintV.GetValue()
}

// SetLastQueryPacketSeq saves the id from the last query request
func (k Keeper) SetLastQueryPacketSeq(ctx sdk.Context, packetSequence uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetLastQueryPacketSeqKey(),
		k.cdc.MustMarshalLengthPrefixed(&gogotypes.UInt64Value{Value: packetSequence}))
}
