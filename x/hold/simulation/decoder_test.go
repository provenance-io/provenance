package simulation_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/kv"

	"github.com/provenance-io/provenance/testutil"
	"github.com/provenance-io/provenance/x/hold/keeper"
	"github.com/provenance-io/provenance/x/hold/simulation"
)

func TestDecodeStore(t *testing.T) {
	cdc := simapp.MakeTestEncodingConfig().Codec
	dec := simulation.NewDecodeStore(cdc)

	addr0 := sdk.AccAddress("addr0_______________")
	addr1 := sdk.AccAddress("addr1_______________")

	tests := []struct {
		name     string
		kvA      kv.Pair
		kvB      kv.Pair
		exp      string
		expPanic string
	}{
		{
			name: "HoldCoin",
			kvA:  kv.Pair{Key: keeper.CreateHoldCoinKey(addr0, "banana"), Value: []byte("99")},
			kvB:  kv.Pair{Key: keeper.CreateHoldCoinKey(addr1, "cherry"), Value: []byte("123")},
			exp:  "<HoldCoin><" + addr0.String() + "><banana>: A = \"99\", B = \"123\"\n",
		},
		{
			name:     "unknown",
			kvA:      kv.Pair{Key: []byte{0x9a}, Value: []byte{0x9b}},
			kvB:      kv.Pair{Key: []byte{0x9c}, Value: []byte{0x9d}},
			expPanic: "invalid hold key 9A",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual string
			testFunc := func() {
				actual = dec(tc.kvA, tc.kvB)
			}
			testutil.RequirePanicEquals(t, testFunc, tc.expPanic, "running decoder")
			assert.Equal(t, tc.exp, actual, "decoder result")
		})
	}
}
