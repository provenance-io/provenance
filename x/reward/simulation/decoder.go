package simulation

import (
	"bytes"
	"fmt"

	"github.com/provenance-io/provenance/x/msgfees/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/kv"
)

// NewDecodeStore returns a decoder function closure that unmarshalls the KVPair's
// Value
func NewDecodeStore(cdc codec.Codec) func(kvA, kvB kv.Pair) string {
	return func(kvA, kvB kv.Pair) string {
		switch {
		case bytes.Equal(kvA.Key[:1], types.MsgFeeKeyPrefix):
			var attribA, attribB types.MsgFee

			cdc.MustUnmarshal(kvA.Value, &attribA)
			cdc.MustUnmarshal(kvB.Value, &attribB)

			return fmt.Sprintf("%v\n%v", attribA, attribB)
		default:
			panic(fmt.Sprintf("unexpected %s key %X (%s)", types.ModuleName, kvA.Key, kvA.Key))
		}
	}
}
