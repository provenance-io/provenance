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
		case bytes.HasPrefix(kvA.Key, keeper.KeyPrefixEscrowCoin):
			addr, denom := keeper.ParseEscrowCoinKey(kvA.Key)
			valAMsg := escrowCoinValueMsg(kvA.Value)
			valBMsg := escrowCoinValueMsg(kvB.Value)
			return fmt.Sprintf("<EscrowCoin><%s><%s>: A = %s, B = %s\n", addr, denom, valAMsg, valBMsg)

		default:
			panic(fmt.Sprintf("invalid escrow key %X", kvA.Key))
		}
	}
}

// escrowCoinValueMsg converts the given bytes into an escrow coin entry value string.
func escrowCoinValueMsg(value []byte) string {
	val, err := keeper.UnmarshalEscrowCoinValue(value)
	if err != nil {
		return fmt.Sprintf("<invalid>: %v", value)
	}
	return `"` + val.String() + `"`
}
