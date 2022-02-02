package simulation_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/types/kv"

	"github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/attribute/simulation"
	"github.com/provenance-io/provenance/x/attribute/types"
)

func TestDecodeStore(t *testing.T) {
	cdc := app.MakeEncodingConfig().Marshaler
	dec := simulation.NewDecodeStore(cdc)

	testAttributeRecord := types.NewAttribute("test", "", types.AttributeType_Int, []byte{1})

	kvPairs := kv.Pairs{
		Pairs: []kv.Pair{
			{Key: types.AttributeKeyPrefix, Value: cdc.MustMarshal(&testAttributeRecord)},
			{Key: []byte{0x99}, Value: []byte{0x99}},
		},
	}

	tests := []struct {
		attribute   string
		expectedLog string
	}{
		{"Attribute Record", fmt.Sprintf("%v\n%v", testAttributeRecord, testAttributeRecord)},
		{"other", ""},
	}

	for i, tt := range tests {
		i, tt := i, tt
		t.Run(tt.attribute, func(t *testing.T) {
			switch i {
			case len(tests) - 1:
				require.Panics(t, func() { dec(kvPairs.Pairs[i], kvPairs.Pairs[i]) }, tt.attribute)
			default:
				require.Equal(t, tt.expectedLog, dec(kvPairs.Pairs[i], kvPairs.Pairs[i]), tt.attribute)
			}
		})
	}
}
