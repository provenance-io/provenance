package simulation

import (
	"bytes"
	"fmt"
	"github.com/cosmos/cosmos-sdk/types/kv"

	"github.com/provenance-io/provenance/x/epoch/types"

	"github.com/cosmos/cosmos-sdk/codec"
)

// NewDecodeStore returns a decoder function closure that unmarshalls the KVPair's
// Value
func NewDecodeStore(cdc codec.Codec) func(kvA, kvB kv.Pair) string {
	return func(kvA, kvB kv.Pair) string {
		switch {
		case bytes.Equal(kvA.Key[:1], types.KeyPrefixEpoch):
			var attribA, attribB types.EpochInfo

			cdc.MustUnmarshal(kvA.Value, &attribA)
			cdc.MustUnmarshal(kvB.Value, &attribB)

			return fmt.Sprintf("%v\n%v", attribA, attribB)
		default:
			panic(fmt.Sprintf("unexpected %s key %X (%s)", types.ModuleName, kvA.Key, kvA.Key))
		}
	}
}
