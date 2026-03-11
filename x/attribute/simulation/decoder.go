package simulation

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/kv"

	"github.com/provenance-io/provenance/x/attribute/types"
)

// NewDecodeStore returns a decoder function closure that unmarshalls the KVPair's
// Value
func NewDecodeStore(cdc codec.Codec) func(kvA, kvB kv.Pair) string {
	return func(kvA, kvB kv.Pair) string {
		switch {
		case bytes.Equal(kvA.Key[:1], types.AttributeKeyPrefix):
			var attribA, attribB types.Attribute

			cdc.MustUnmarshal(kvA.Value, &attribA)
			cdc.MustUnmarshal(kvB.Value, &attribB)

			return fmt.Sprintf("%v\n%v", attribA, attribB)

		case bytes.Equal(kvA.Key[:1], types.AttributeAddrLookupKeyPrefix):
			countA := binary.BigEndian.Uint64(kvA.Value)
			countB := binary.BigEndian.Uint64(kvB.Value)
			return fmt.Sprintf("%d\n%d", countA, countB)

		case bytes.Equal(kvA.Key[:1], types.AttributeExpirationKeyPrefix):
			return fmt.Sprintf("%X\n%X", kvA.Value, kvB.Value)

		case bytes.Equal(kvA.Key[:1], types.AttributeParamPrefix):
			var paramsA, paramsB types.Params
			cdc.MustUnmarshal(kvA.Value, &paramsA)
			cdc.MustUnmarshal(kvB.Value, &paramsB)
			return fmt.Sprintf("%v\n%v", paramsA, paramsB)
		default:
			panic(fmt.Sprintf("unexpected %s key %X (%s)", types.ModuleName, kvA.Key, kvA.Key))
		}
	}
}
