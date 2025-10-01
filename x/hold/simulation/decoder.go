package simulation

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/kv"

	"github.com/provenance-io/provenance/x/hold/keeper"
)

// NewDecodeStore returns a decoder function closure that unmarshals the KVPair's
// Value to the corresponding group type.
func NewDecodeStore(_ codec.Codec) func(kvA, kvB kv.Pair) string {
	return func(kvA, kvB kv.Pair) string {
		switch {
		case bytes.HasPrefix(kvA.Key, keeper.KeyPrefixHoldCoin):
			addr, denom := keeper.ParseHoldCoinKey(kvA.Key)
			valAMsg := holdCoinValueMsg(kvA.Value)
			valBMsg := holdCoinValueMsg(kvB.Value)
			return fmt.Sprintf("<HoldCoin><%s><%s>: A = %s, B = %s\n", addr, denom, valAMsg, valBMsg)

		default:
			panic(fmt.Sprintf("invalid hold key %X", kvA.Key))
		}
	}
}

// holdCoinValueMsg converts the given bytes into a hold coin entry value string.
func holdCoinValueMsg(value []byte) string {
	val, err := keeper.UnmarshalHoldCoinValue(value)
	if err != nil {
		return fmt.Sprintf("<invalid>: %v", value)
	}
	return `"` + val.String() + `"`
}
