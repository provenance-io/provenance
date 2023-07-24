package simulation

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/kv"
)

// NewDecodeStore returns a decoder function closure that unmarshals the KVPair's
// Value to the corresponding group type.
func NewDecodeStore(cdc codec.Codec) func(kvA, kvB kv.Pair) string {
	return func(kvA, kvB kv.Pair) string {
		// TODO[1607]: Implement the simulation NewDecodeStore function.
		// case bytes.HasPrefix(kvA.Key, escrow.CoinKey):
		_ = cdc
		return "decoder not implemented"
	}
}
