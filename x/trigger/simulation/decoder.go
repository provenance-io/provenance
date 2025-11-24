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
		case bytes.Equal(kvA.Key, types.NextTriggerIDKey.Bytes()):
			var attribA, attribB uint64
			attribA = types.GetTriggerIDFromBytes(kvA.Value)
			attribB = types.GetTriggerIDFromBytes(kvB.Value)
			return fmt.Sprintf("TriggerID: A:[%v] B:[%v]\n", attribA, attribB)

		case bytes.Equal(kvA.Key, types.QueueStartIndexKey.Bytes()):
			var attribA, attribB uint64
			attribA = types.GetQueueIndexFromBytes(kvA.Value)
			attribB = types.GetQueueIndexFromBytes(kvB.Value)
			return fmt.Sprintf("QueueStart: A:[%v] B:[%v]\n", attribA, attribB)

		case bytes.Equal(kvA.Key, types.QueueLengthKey.Bytes()):
			var attribA, attribB uint64
			attribA = types.GetQueueIndexFromBytes(kvA.Value)
			attribB = types.GetQueueIndexFromBytes(kvB.Value)
			return fmt.Sprintf("QueueLength: A:[%v] B:[%v]\n", attribA, attribB)

		case bytes.Equal(kvA.Key[:1], types.TriggerKeyPrefix.Bytes()):
			if len(kvA.Key) < 1+types.TriggerIDLength {
				return fmt.Sprintf("invalid trigger key length: %d", len(kvA.Key))
			}
			var attribA, attribB types.Trigger
			cdc.MustUnmarshal(kvA.Value, &attribA)
			cdc.MustUnmarshal(kvB.Value, &attribB)
			return fmt.Sprintf("Trigger: A:[%v] B:[%v]\n", attribA, attribB)

		case bytes.Equal(kvA.Key[:1], types.EventListenerKeyPrefix.Bytes()):
			var attribA, attribB types.Trigger
			cdc.MustUnmarshal(kvA.Value, &attribA)
			cdc.MustUnmarshal(kvB.Value, &attribB)
			return fmt.Sprintf("EventListener: A:[%v] B:[%v]\n", attribA, attribB)

		case bytes.Equal(kvA.Key[:1], types.QueueKeyPrefix.Bytes()):
			if len(kvA.Key) < 1+types.QueueIndexLength {
				return fmt.Sprintf("invalid queue key length: %d", len(kvA.Key))
			}
			var attribA, attribB types.QueuedTrigger
			cdc.MustUnmarshal(kvA.Value, &attribA)
			cdc.MustUnmarshal(kvB.Value, &attribB)
			return fmt.Sprintf("QueuedTrigger: A:[%v] B:[%v]\n", attribA, attribB)

		default:
			panic(fmt.Sprintf("unexpected %s key %X (%s)", types.ModuleName, kvA.Key, kvA.Key))
		}
	}
}
