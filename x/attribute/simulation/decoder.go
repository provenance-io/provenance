package simulation

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/kv"

	"github.com/provenance-io/provenance/x/attribute/types"
)

func NewDecodeStore(cdc codec.Codec) func(kvA, kvB kv.Pair) string {
	return func(kvA, kvB kv.Pair) string {
		switch {
		case bytes.Equal(kvA.Key[:1], types.AttributeKeyPrefix):
			var attribA, attribB types.Attribute
			unmarshalIfPresent(cdc, kvA.Value, &attribA)
			unmarshalIfPresent(cdc, kvB.Value, &attribB)
			return fmt.Sprintf("%v\n%v", attribA, attribB)

		case bytes.Equal(kvA.Key[:1], types.AttributeAddrLookupKeyPrefix):
			return fmt.Sprintf("%d\n%d", decodeUint64(kvA.Value), decodeUint64(kvB.Value))

		case bytes.Equal(kvA.Key[:1], types.AttributeExpirationKeyPrefix):
			return fmt.Sprintf("%X\n%X", kvA.Value, kvB.Value)

		case bytes.Equal(kvA.Key[:1], types.AttributeParamPrefix):
			var paramsA, paramsB types.Params
			unmarshalIfPresent(cdc, kvA.Value, &paramsA)
			unmarshalIfPresent(cdc, kvB.Value, &paramsB)
			return fmt.Sprintf("%v\n%v", paramsA, paramsB)

		default:
			panic(fmt.Sprintf("unexpected %s key %X (%s)", types.ModuleName, kvA.Key, kvA.Key))
		}
	}
}

func decodeUint64(b []byte) uint64 {
	if len(b) >= 8 {
		return binary.BigEndian.Uint64(b)
	}
	return 0
}

//nolint:staticcheck // codec.ProtoMarshaler required for compatibility.
func unmarshalIfPresent(cdc codec.Codec, b []byte, msg codec.ProtoMarshaler) {
	if len(b) > 0 {
		cdc.MustUnmarshal(b, msg)
	}
}
