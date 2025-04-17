package simulation

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/kv"
)

// NewDecodeStore returns a new store decoder for the asset state.
func NewDecodeStore(_ codec.Codec) func(kvA, kvB kv.Pair) string {
	return func(_, _ kv.Pair) string {
		// TODO write NewDecodeStore
		return "not implemented"
	}
}
