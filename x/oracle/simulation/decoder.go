package simulation

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/kv"

	"github.com/provenance-io/provenance/x/oracle/types"
)

// NewDecodeStore returns a decoder function closure that unmarshalls the KVPair's
// Value
func NewDecodeStore(cdc codec.Codec) func(kvA, kvB kv.Pair) string {
	return func(kvA, kvB kv.Pair) string {
		switch {
		case bytes.Equal(kvA.Key[:1], types.OracleStoreKey):
			var attribA, attribB sdk.AccAddress = kvA.Value, kvB.Value
			return fmt.Sprintf("Oracle Address: A:[%v] B:[%v]\n", attribA, attribB)
		case bytes.Equal(kvA.Key[:1], types.LastQueryPacketSeqKey):
			attribA := binary.BigEndian.Uint64(kvA.Value)
			attribB := binary.BigEndian.Uint64(kvB.Value)

			return fmt.Sprintf("Last Query Packet Sequence: A:[%v] B:[%v]\n", attribA, attribB)
		case bytes.Equal(kvA.Key[:1], types.PortStoreKey):
			attribA := string(kvA.Value)
			attribB := string(kvB.Value)

			return fmt.Sprintf("Port: A:[%v] B:[%v]\n", attribA, attribB)
		default:
			panic(fmt.Sprintf("unexpected %s key %X (%s)", types.ModuleName, kvA.Key, kvA.Key))
		}
	}
}
