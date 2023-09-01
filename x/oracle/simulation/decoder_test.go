package simulation_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/types/kv"

	"github.com/provenance-io/provenance/testutil/assertions"
	"github.com/provenance-io/provenance/x/oracle/simulation"
	"github.com/provenance-io/provenance/x/oracle/types"
)

func TestDecodeStore(t *testing.T) {
	cdc := simapp.MakeTestEncodingConfig().Codec
	dec := simulation.NewDecodeStore(cdc)

	tests := []struct {
		name     string
		kvA      kv.Pair
		kvB      kv.Pair
		exp      string
		expPanic string
	}{
		{
			name:     "failure - unknown key type",
			kvA:      kv.Pair{Key: []byte{0x9a}, Value: []byte{0x9b}},
			kvB:      kv.Pair{Key: []byte{0x9c}, Value: []byte{0x9d}},
			expPanic: "unexpected oracle key 9A (\x9a)",
		},
		{
			name: "success - OracleStoreKey",
			kvA:  kv.Pair{Key: types.GetOracleStoreKey(), Value: []byte("99")},
			kvB:  kv.Pair{Key: types.GetOracleStoreKey(), Value: []byte("88")},
			exp:  "Oracle Address: A:[3939] B:[3838]\n",
		},
		{
			name: "success - PortStoreKey",
			kvA:  kv.Pair{Key: types.GetPortStoreKey(), Value: []byte("99")},
			kvB:  kv.Pair{Key: types.GetPortStoreKey(), Value: []byte("88")},
			exp:  "Port: A:[99] B:[88]\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual string
			testFunc := func() {
				actual = dec(tc.kvA, tc.kvB)
			}
			assertions.RequirePanicEquals(t, testFunc, tc.expPanic, "running decoder")
			assert.Equal(t, tc.exp, actual, "decoder result")
		})
	}
}
