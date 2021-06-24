package antewrapper

import (
	"testing"

	sdkgas "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/stretchr/testify/require"
)

func TestGasMeter(t *testing.T) {

	t.Parallel()
	cases := []struct {
		limit sdkgas.Gas
		usage []sdkgas.Gas
	}{
		{10, []sdkgas.Gas{1, 2, 3, 4}},
		{1000, []sdkgas.Gas{40, 30, 20, 10, 900}},
		{100000, []sdkgas.Gas{99999, 1}},
		{100000000, []sdkgas.Gas{50000000, 40000000, 10000000}},
		{65535, []sdkgas.Gas{32768, 32767}},
		{65536, []sdkgas.Gas{32768, 32767, 1}},
	}

	for tcnum, tc := range cases {
		meter := NewTracingMeterWrapper(log.TestingLogger(), sdkgas.NewGasMeter(tc.limit))
		used := uint64(0)

		for unum, usage := range tc.usage {
			usage := usage
			used += usage
			require.NotPanics(t, func() { meter.ConsumeGas(usage, "") }, "Not exceeded limit but panicked. tc #%d, usage #%d", tcnum, unum)
			require.Equal(t, used, meter.GasConsumed(), "sdkgas.Gas consumption not match. tc #%d, usage #%d", tcnum, unum)
			require.Equal(t, used, meter.GasConsumedToLimit(), "sdkgas.Gas consumption (to limit) not match. tc #%d, usage #%d", tcnum, unum)
			require.False(t, meter.IsPastLimit(), "Not exceeded limit but got IsPastLimit() true")
			if unum < len(tc.usage)-1 {
				require.False(t, meter.IsOutOfGas(), "Not yet at limit but got IsOutOfGas() true")
			} else {
				require.True(t, meter.IsOutOfGas(), "At limit but got IsOutOfGas() false")
			}
		}

		require.Panics(t, func() { meter.ConsumeGas(1, "") }, "Exceeded but not panicked. tc #%d", tcnum)
		require.Equal(t, meter.GasConsumedToLimit(), meter.Limit(), "sdkgas.Gas consumption (to limit) not match limit")
		require.Equal(t, meter.GasConsumed(), meter.Limit()+1, "sdkgas.Gas consumption not match limit+1")
		meter2 := NewTracingMeterWrapper(log.TestingLogger(), sdkgas.NewGasMeter(100))
		meter2.ConsumeGas(sdkgas.Gas(50), "consume half max")
		require.Equalf(t, "TracingGasMeter:\n  limit: 100\n  consumed: 50", meter2.String(), "expect string output to match")
		// TODO wait for cosmos team to say where this went
		// meter2.RefundGas(uint64(20), "refund")
		require.Equalf(t, "TracingGasMeter:\n  limit: 100\n  consumed: 30", meter2.String(), "expect string output to match")
		require.Panics(t, func() { meter2.ConsumeGas(sdkgas.Gas(70)+2, "panic") })
	}
}
