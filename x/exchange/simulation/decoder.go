package simulation

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/kv"
)

// NewDecodeStore returns a new store decoder for the exchange state.
func NewDecodeStore(_ codec.Codec) func(kvA, kvB kv.Pair) string {
	return func(_, _ kv.Pair) string {
		// TODO[1658]: Write NewDecodeStore.
		return "not implemented"
	}
}
