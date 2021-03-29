package simulation

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/kv"
	"github.com/provenance-io/provenance/x/name/types"
)

// NewDecodeStore returns a decoder function closure that unmarshals the KVPair's
// Value
func NewDecodeStore(cdc codec.Marshaler) func(kvA, kvB kv.Pair) string {
	return func(kvA, kvB kv.Pair) string {
		switch {
		case bytes.Equal(kvA.Key[:1], types.NameKeyPrefixAmino):
			var nameA, nameB types.NameRecord

			cdc.MustUnmarshalBinaryBare(kvA.Value, &nameA)
			cdc.MustUnmarshalBinaryBare(kvB.Value, &nameB)

			return fmt.Sprintf("%v\n%v", nameA, nameB)
		case bytes.Equal(kvA.Key[:1], types.AddressKeyPrefixAmino):
			var nameA, nameB types.NameRecord

			cdc.MustUnmarshalBinaryBare(kvA.Value, &nameA)
			cdc.MustUnmarshalBinaryBare(kvB.Value, &nameB)

			return fmt.Sprintf("%v\n%v", nameA, nameB)
		default:
			panic(fmt.Sprintf("unexpected %s key %X (%s)", types.ModuleName, kvA.Key, kvA.Key))
		}
	}
}
