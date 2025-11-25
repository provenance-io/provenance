package simulation

import (
	"bytes"
	"encoding/binary"
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
			attribA := safeDecodeUint64(kvA.Value)
			attribB := safeDecodeUint64(kvB.Value)
			return fmt.Sprintf("NextTriggerID: A:[%v] B:[%v]", attribA, attribB)

		case bytes.Equal(kvA.Key, types.QueueStartIndexKey.Bytes()):
			attribA := safeDecodeUint64(kvA.Value)
			attribB := safeDecodeUint64(kvB.Value)
			return fmt.Sprintf("QueueStartIndex: A:[%v] B:[%v]", attribA, attribB)

		case bytes.Equal(kvA.Key, types.QueueLengthKey.Bytes()):
			attribA := safeDecodeUint64(kvA.Value)
			attribB := safeDecodeUint64(kvB.Value)
			return fmt.Sprintf("QueueLength: A:[%v] B:[%v]", attribA, attribB)

		case bytes.Equal(kvA.Key[:1], types.TriggerKeyPrefix.Bytes()):
			if len(kvA.Key) < 1+types.TriggerIDLength {
				return fmt.Sprintf("invalid Trigger key length: %d", len(kvA.Key))
			}

			var triggerA, triggerB types.Trigger
			if len(kvA.Value) > 0 {
				if err := cdc.Unmarshal(kvA.Value, &triggerA); err != nil {
					return fmt.Sprintf("ERROR unmarshaling Trigger A: %v", err)
				}
			} else {
				return "Trigger A has empty value"
			}

			if len(kvB.Value) > 0 {
				if err := cdc.Unmarshal(kvB.Value, &triggerB); err != nil {
					return fmt.Sprintf("ERROR unmarshaling Trigger B: %v", err)
				}
			} else {
				return "Trigger B has empty value"
			}

			return fmt.Sprintf("Trigger: A:[%v] B:[%v]", triggerA, triggerB)

		case bytes.Equal(kvA.Key[:1], types.EventListenerKeyPrefix.Bytes()):
			triggerIDA := extractTriggerIDFromEventListenerKey(kvA.Key)
			triggerIDB := extractTriggerIDFromEventListenerKey(kvB.Key)

			if len(kvA.Key) < 49 {
				return fmt.Sprintf("invalid EventListener key A length: %d", len(kvA.Key))
			}
			if len(kvB.Key) < 49 {
				return fmt.Sprintf("invalid EventListener key B length: %d", len(kvB.Key))
			}

			return fmt.Sprintf("EventListener: TriggerIDA:[%v] TriggerIDB:[%v]", triggerIDA, triggerIDB)

		case bytes.Equal(kvA.Key[:1], types.QueueKeyPrefix.Bytes()):
			if len(kvA.Key) < 1+types.QueueIndexLength {
				return fmt.Sprintf("invalid Queue key length: %d", len(kvA.Key))
			}
			var queueA, queueB types.QueuedTrigger
			if len(kvA.Value) > 0 {
				if err := cdc.Unmarshal(kvA.Value, &queueA); err != nil {
					return fmt.Sprintf("ERROR unmarshaling QueuedTrigger A: %v", err)
				}
			} else {
				return "QueuedTrigger A has empty value"
			}
			if len(kvB.Value) > 0 {
				if err := cdc.Unmarshal(kvB.Value, &queueB); err != nil {
					return fmt.Sprintf("ERROR unmarshaling QueuedTrigger B: %v", err)
				}
			} else {
				return "QueuedTrigger B has empty value"
			}

			return fmt.Sprintf("QueuedTrigger: A:[%v] B:[%v]", queueA, queueB)

		default:
			panic(fmt.Sprintf("unexpected %s key %X (%s)", types.ModuleName, kvA.Key, kvA.Key))
		}
	}
}

func extractTriggerIDFromEventListenerKey(key []byte) uint64 {
	if len(key) < 49 { // 1 + 32 + 8 + 8
		return 0
	}
	triggerIDBytes := key[41:49]
	if len(triggerIDBytes) < 8 {
		return 0
	}
	return binary.BigEndian.Uint64(triggerIDBytes)
}

func safeDecodeUint64(bz []byte) uint64 {
	if len(bz) < 8 {
		return 0
	}
	return binary.BigEndian.Uint64(bz)
}
