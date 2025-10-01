package simulation

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/kv"

	"github.com/provenance-io/provenance/x/ibcratelimit"
)

// NewDecodeStore returns a decoder function closure that unmarshalls the KVPair's
// Value
func NewDecodeStore(cdc codec.Codec) func(kvA, kvB kv.Pair) string {
	return func(kvA, kvB kv.Pair) string {
		switch {
		case bytes.Equal(kvA.Key[:1], ibcratelimit.ParamsKey):
			var attribA, attribB ibcratelimit.Params

			cdc.MustUnmarshal(kvA.Value, &attribA)
			cdc.MustUnmarshal(kvB.Value, &attribB)

			return fmt.Sprintf("Params: A:[%v] B:[%v]\n", attribA, attribB)
		default:
			panic(fmt.Sprintf("unexpected %s key %X (%s)", ibcratelimit.ModuleName, kvA.Key, kvA.Key))
		}
	}
}
