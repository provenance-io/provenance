package antewrapper

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/internal/pioconfig"
)

func TestFeeGasMeter(t *testing.T) {
	pioconfig.SetProvenanceConfig("", 0)
	casesFeeGas := []struct {
		limit storetypes.Gas
		usage []storetypes.Gas
		fees  map[string]sdk.Coin
	}{
		{limit: 10, usage: []storetypes.Gas{1, 2, 3, 4}, fees: nil},
		{limit: 1000, usage: []storetypes.Gas{40, 30, 20, 10, 900}, fees: map[string]sdk.Coin{"/cosmos.bank.v1beta1.MsgSend": sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 1_000_000), "/provenance.marker.v1.MsgAddMarkerRequest": sdk.NewInt64Coin("doge", 1_000_000)}},
		{limit: 100_000, usage: []storetypes.Gas{99999, 1}, fees: map[string]sdk.Coin{"/cosmos.bank.v1beta1.MsgSend": sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 1_000_000)}},
		{limit: 100_000_000, usage: []storetypes.Gas{50_000_000, 40_000_000, 10_000_000}, fees: map[string]sdk.Coin{"/cosmos.bank.v1beta1.MsgSend": sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 5555)}},
		{limit: 65535, usage: []storetypes.Gas{32768, 32767}, fees: nil},
		{limit: 65536, usage: []storetypes.Gas{32768, 32767, 1}, fees: nil},
	}

	for tcnum, tc := range casesFeeGas {
		meter := NewFeeGasMeterWrapper(log.NewTestLogger(t), storetypes.NewGasMeter(tc.limit), false).(*FeeGasMeter)
		used := uint64(0)
		var usedFee sdk.Coins

		for unum, usage := range tc.usage {
			usage := usage
			used += usage
			require.NotPanics(t, func() { meter.ConsumeGas(usage, "") }, "Not exceeded limit but panicked. tc #%d, usage #%d", tcnum, unum)
			require.Equal(t, used, meter.GasConsumed(), "storetypes.Gas consumption not match. tc #%d, usage #%d", tcnum, unum)
			require.Equal(t, used, meter.GasConsumedToLimit(), "storetypes.Gas consumption (to limit) not match. tc #%d, usage #%d", tcnum, unum)
			require.False(t, meter.IsPastLimit(), "Not exceeded limit but got IsPastLimit() true")
			if unum < len(tc.usage)-1 {
				require.False(t, meter.IsOutOfGas(), "Not yet at limit but got IsOutOfGas() true")
			} else {
				require.True(t, meter.IsOutOfGas(), "At limit but got IsOutOfGas() false")
			}
		}
		// fees
		for msgType, fee := range tc.fees {
			usageFee := fee
			if usedFee.Empty() {
				usedFee = sdk.NewCoins(usageFee)
			} else {
				usedFee = usedFee.Add(usageFee)
			}
			require.NotPanics(t, func() { meter.ConsumeFee(sdk.NewCoins(usageFee), msgType, "") }, "panicked on adding fees")
		}

		require.Panics(t, func() { meter.ConsumeGas(1, "") }, "Exceeded but not panicked. tc #%d", tcnum)
		require.Equal(t, meter.GasConsumedToLimit(), meter.Limit(), "storetypes.Gas consumption (to limit) not match limit")
		require.Equal(t, meter.GasConsumed(), meter.Limit()+1, "storetypes.Gas consumption not match limit+1")
		assert.Equal(t, meter.FeeConsumed().Sort(), usedFee.Sort(), "FeeConsumed does not match all Fees")
		meter2 := NewFeeGasMeterWrapper(log.NewTestLogger(t), storetypes.NewGasMeter(100), false).(*FeeGasMeter)
		meter2.ConsumeGas(storetypes.Gas(50), "consume half max")
		meter2.ConsumeFee(sdk.NewCoins(sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 1_000_000)), "/cosmos.bank.v1beta1.MsgSend", "")
		require.Equalf(t, "feeGasMeter:\n  limit: 100\n  consumed: 50 fee consumed: 1000000nhash", meter2.String(), "expect string output to match")
		meter2.RefundGas(uint64(20), "refund")
		require.Equalf(t, "feeGasMeter:\n  limit: 100\n  consumed: 30 fee consumed: 1000000nhash", meter2.String(), "expect string output to match")
		require.Equalf(t, "1000000nhash", meter2.FeeConsumed().String(), "expect string output to match")
		require.Panics(t, func() { meter2.ConsumeGas(storetypes.Gas(70)+2, "panic") })
		meter2.ConsumeFee(sdk.NewCoins(sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().FeeDenom, 2_000_000)), "/cosmos.bank.v1beta1.MsgSend", "")
		require.Equalf(t, "feeGasMeter:\n  limit: 100\n  consumed: 102 fee consumed: 3000000nhash", meter2.String(), "expect string output to match")
		meter2.RefundGas(uint64(20), "refund")
		require.Equalf(t, "feeGasMeter:\n  limit: 100\n  consumed: 82 fee consumed: 3000000nhash", meter2.String(), "expect string output to match")
		require.Equalf(t, "3000000nhash", meter2.FeeConsumed().String(), "expect string output to match")
		require.Equalf(t, "map[/cosmos.bank.v1beta1.MsgSend:3000000nhash]", fmt.Sprintf("%v", meter2.FeeConsumedByMsg()), "expect string output to match")
		meter2.ConsumeFee(sdk.NewCoins(sdk.NewInt64Coin("doge", 2_000_000)), "/provenance.marker.v1.MsgAddMarkerRequest", "")
		meter2.ConsumeFee(sdk.NewCoins(sdk.NewInt64Coin("jackthecat", 420)), "/provenance.marker.v1.MsgAddMarkerRequest", "")
		meter2FeesConsumed := meter2.FeeConsumed()
		require.Equalf(t, "2000000doge,420jackthecat,3000000nhash", meter2FeesConsumed.String(), "expect string output to match")
		require.Equalf(t, "2000000doge,420jackthecat", meter2.FeeConsumedForType("/provenance.marker.v1.MsgAddMarkerRequest", "").String(), "expect string output to match")
		require.Equalf(t, "3000000nhash", meter2.FeeConsumedForType("/cosmos.bank.v1beta1.MsgSend", "").String(), "expect string output to match")
		require.Equalf(t, false, meter2.IsSimulate(), "simulate should be false")
	}
}
