package simulation_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cosmos/cosmos-sdk/types/kv"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"

	"github.com/provenance-io/provenance/testutil/assertions"
	"github.com/provenance-io/provenance/x/ibcratelimit"
	ibcratelimitmodule "github.com/provenance-io/provenance/x/ibcratelimit/module"
	"github.com/provenance-io/provenance/x/ibcratelimit/simulation"
)

func TestDecodeStore(t *testing.T) {
	cdc := moduletestutil.MakeTestEncodingConfig(ibcratelimitmodule.AppModuleBasic{}).Codec
	dec := simulation.NewDecodeStore(cdc)
	params := func(contract string) []byte {
		p := ibcratelimit.NewParams("contract a")
		return cdc.MustMarshal(&p)
	}

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
			expPanic: "unexpected ratelimitedibc key 9A (\x9a)",
		},
		{
			name: "success - ParamsKey",
			kvA:  kv.Pair{Key: ibcratelimit.ParamsKey, Value: params("contract a")},
			kvB:  kv.Pair{Key: ibcratelimit.ParamsKey, Value: params("contract b")},
			exp:  "Params: A:[{contract a}] B:[{contract a}]\n",
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
