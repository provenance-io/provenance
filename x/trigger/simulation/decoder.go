package simulation

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/kv"
	"github.com/provenance-io/provenance/x/trigger/types"
)

// NewDecodeStore returns a decoder function closure that unmarshalls the KVPair's
// Value
func NewDecodeStore(cdc codec.Codec) func(kvA, kvB kv.Pair) string {
	return func(kvA, kvB kv.Pair) string {
		switch {
		case bytes.Equal(kvA.Key[:1], types.TriggerKeyPrefix):
			var attribA, attribB types.Trigger

			cdc.MustUnmarshal(kvA.Value, &attribA)
			cdc.MustUnmarshal(kvB.Value, &attribB)

			return fmt.Sprintf("%v\n%v", attribA, attribB)
		case bytes.Equal(kvA.Key[:1], types.EventListenerKeyPrefix):
			var attribA, attribB types.Trigger

			cdc.MustUnmarshal(kvA.Value, &attribA)
			cdc.MustUnmarshal(kvB.Value, &attribB)

			return fmt.Sprintf("%v\n%v", attribA, attribB)
		case bytes.Equal(kvA.Key[:1], types.QueueKeyPrefix):
			var attribA, attribB types.QueuedTrigger

			cdc.MustUnmarshal(kvA.Value, &attribA)
			cdc.MustUnmarshal(kvB.Value, &attribB)

			return fmt.Sprintf("%v\n%v", attribA, attribB)
		case bytes.Equal(kvA.Key[:1], types.GasLimitKeyPrefix):
			var attribA, attribB uint64
			attribA = types.GetGasLimitFromBytes(kvA.Value)
			attribB = types.GetGasLimitFromBytes(kvB.Value)

			return fmt.Sprintf("%v\n%v", attribA, attribB)
		case bytes.Equal(kvA.Key[:1], types.NextTriggerIDKey):
			var attribA, attribB uint64
			attribA = types.GetGasLimitFromBytes(kvA.Value)
			attribB = types.GetGasLimitFromBytes(kvB.Value)

			return fmt.Sprintf("%v\n%v", attribA, attribB)
		case bytes.Equal(kvA.Key[:1], types.QueueStartIndexKey):
			var attribA, attribB uint64
			attribA = types.GetGasLimitFromBytes(kvA.Value)
			attribB = types.GetGasLimitFromBytes(kvB.Value)

			return fmt.Sprintf("%v\n%v", attribA, attribB)
		case bytes.Equal(kvA.Key[:1], types.QueueLengthKey):
			var attribA, attribB uint64
			attribA = types.GetGasLimitFromBytes(kvA.Value)
			attribB = types.GetGasLimitFromBytes(kvB.Value)

			return fmt.Sprintf("%v\n%v", attribA, attribB)
		default:
			panic(fmt.Sprintf("unexpected %s key %X (%s)", types.ModuleName, kvA.Key, kvA.Key))
		}
	}
}
