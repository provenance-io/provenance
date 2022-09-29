package antewrapper

import (
	"fmt"
	"testing"

	sdkgas "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"

	msgfeestype "github.com/provenance-io/provenance/x/msgfees/types"
)

func TestFeeGasMeter(t *testing.T) {
	casesFeeGas := []struct {
		limit sdkgas.Gas
		usage []sdkgas.Gas
		fees  map[string]sdk.Coin
	}{
		{10, []sdkgas.Gas{1, 2, 3, 4}, nil},
		{1000, []sdkgas.Gas{40, 30, 20, 10, 900}, map[string]sdk.Coin{"/cosmos.bank.v1beta1.MsgSend": sdk.NewCoin(msgfeestype.NhashDenom, sdk.NewInt(1000000)), "/provenance.marker.v1.MsgAddMarkerRequest": sdk.NewCoin("doge", sdk.NewInt(1000000))}},
		{100000, []sdkgas.Gas{99999, 1}, map[string]sdk.Coin{"/cosmos.bank.v1beta1.MsgSend": sdk.NewCoin(msgfeestype.NhashDenom, sdk.NewInt(1000000))}},
		{100000000, []sdkgas.Gas{50000000, 40000000, 10000000}, map[string]sdk.Coin{"/cosmos.bank.v1beta1.MsgSend": sdk.NewCoin(msgfeestype.NhashDenom, sdk.NewInt(5555))}},
		{65535, []sdkgas.Gas{32768, 32767}, nil},
		{65536, []sdkgas.Gas{32768, 32767, 1}, nil},
	}

	for tcnum, tc := range casesFeeGas {
		meter := NewFeeGasMeterWrapper(log.TestingLogger(), sdkgas.NewGasMeter(tc.limit), false).(*FeeGasMeter)
		used := uint64(0)
		var usedFee sdk.Coins

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
		require.Equal(t, meter.GasConsumedToLimit(), meter.Limit(), "sdkgas.Gas consumption (to limit) not match limit")
		require.Equal(t, meter.GasConsumed(), meter.Limit()+1, "sdkgas.Gas consumption not match limit+1")
		assert.Equal(t, meter.FeeConsumed().Sort(), usedFee.Sort(), "FeeConsumed does not match all Fees")
		meter2 := NewFeeGasMeterWrapper(log.TestingLogger(), sdkgas.NewGasMeter(100), false).(*FeeGasMeter)
		meter2.ConsumeGas(sdkgas.Gas(50), "consume half max")
		meter2.ConsumeFee(sdk.NewCoins(sdk.NewCoin(msgfeestype.NhashDenom, sdk.NewInt(1000000))), "/cosmos.bank.v1beta1.MsgSend", "")
		require.Equalf(t, "feeGasMeter:\n  limit: 100\n  consumed: 50 fee consumed: 1000000nhash", meter2.String(), "expect string output to match")
		meter2.RefundGas(uint64(20), "refund")
		require.Equalf(t, "feeGasMeter:\n  limit: 100\n  consumed: 30 fee consumed: 1000000nhash", meter2.String(), "expect string output to match")
		require.Equalf(t, "1000000nhash", meter2.FeeConsumed().String(), "expect string output to match")
		require.Panics(t, func() { meter2.ConsumeGas(sdkgas.Gas(70)+2, "panic") })
		meter2.ConsumeFee(sdk.NewCoins(sdk.NewCoin(msgfeestype.NhashDenom, sdk.NewInt(2000000))), "/cosmos.bank.v1beta1.MsgSend", "")
		require.Equalf(t, "feeGasMeter:\n  limit: 100\n  consumed: 102 fee consumed: 3000000nhash", meter2.String(), "expect string output to match")
		meter2.RefundGas(uint64(20), "refund")
		require.Equalf(t, "feeGasMeter:\n  limit: 100\n  consumed: 82 fee consumed: 3000000nhash", meter2.String(), "expect string output to match")
		require.Equalf(t, "3000000nhash", meter2.FeeConsumed().String(), "expect string output to match")
		require.Equalf(t, "map[/cosmos.bank.v1beta1.MsgSend:3000000nhash]", fmt.Sprintf("%v", meter2.FeeConsumedByMsg()), "expect string output to match")
		meter2.ConsumeFee(sdk.NewCoins(sdk.NewCoin("doge", sdk.NewInt(2000000))), "/provenance.marker.v1.MsgAddMarkerRequest", "")
		meter2.ConsumeFee(sdk.NewCoins(sdk.NewCoin("jackthecat", sdk.NewInt(420))), "/provenance.marker.v1.MsgAddMarkerRequest", "")
		meter2FeesConsumed := meter2.FeeConsumed()
		require.Equalf(t, "2000000doge,420jackthecat,3000000nhash", meter2FeesConsumed.String(), "expect string output to match")
		require.Equalf(t, "2000000doge,420jackthecat", meter2.FeeConsumedForType("/provenance.marker.v1.MsgAddMarkerRequest", "").String(), "expect string output to match")
		require.Equalf(t, "3000000nhash", meter2.FeeConsumedForType("/cosmos.bank.v1beta1.MsgSend", "").String(), "expect string output to match")
		require.Equalf(t, false, meter2.IsSimulate(), "simulate should be false")
	}
}

func TestWithSimulate(t *testing.T) {
	tests := []struct {
		from bool
		to   bool
	}{
		{false, false},
		{false, true},
		{true, false},
		{true, true},
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("%t to %t", tc.from, tc.to), func(tt *testing.T) {
			infGm := sdkgas.NewInfiniteGasMeter()
			gm1 := NewFeeGasMeterWrapper(log.TestingLogger(), infGm, tc.from).(*FeeGasMeter)
			require.Equal(tt, tc.from, gm1.IsSimulate(), "original IsSimulate")

			gm1.ConsumeGas(63, "testing")
			gm1.ConsumeGas(74, "testing")
			gm1.ConsumeGas(3, "other")
			gm1.ConsumeBaseFee(sdk.NewCoins(sdk.NewInt64Coin("banana", 99)))
			gm1.ConsumeFee(sdk.NewCoins(sdk.NewInt64Coin("orange", 22)), "some.msg", "")
			gm1.ConsumeFee(sdk.NewCoins(sdk.NewInt64Coin("orange", 33)), "another.msg", "someone")

			gm1Sim := gm1.IsSimulate()
			gm1Gas := gm1.GasConsumed()
			gm1BaseFee := gm1.BaseFeeConsumed()
			gm1Fee := gm1.FeeConsumed()
			gm1FeeDist := gm1.FeeConsumedDistributions()
			gm1String := gm1.String()
			gm1Events := gm1.EventFeeSummary()

			gm2 := gm1.WithSimulate(tc.to).(*FeeGasMeter)
			require.Equal(tt, tc.to, gm2.IsSimulate(), "new IsSimulate")

			tt.Run("original is unchanged", func(ttt *testing.T) {
				gm1Sim2 := gm1.IsSimulate()
				gm1Gas2 := gm1.GasConsumed()
				gm1BaseFee2 := gm1.BaseFeeConsumed()
				gm1Fee2 := gm1.FeeConsumed()
				gm1FeeDist2 := gm1.FeeConsumedDistributions()
				gm1String2 := gm1.String()
				gm1Events2 := gm1.EventFeeSummary()

				assert.Equal(ttt, gm1Sim, gm1Sim2, "IsSimulate")
				assert.Equal(ttt, int(gm1Gas), int(gm1Gas2), "GasConsumed")
				assert.Equal(ttt, gm1BaseFee, gm1BaseFee2, "BaseFeeConsumed")
				assert.Equal(ttt, gm1Fee, gm1Fee2, "FeeConsumed")
				assert.Equal(ttt, gm1FeeDist, gm1FeeDist2, "FeeConsumedDistributions")
				assert.Equal(ttt, gm1String, gm1String2, "String")
				assert.Equal(ttt, gm1Events, gm1Events2, "EventFeeSummary")
			})

			tt.Run("new is unchanged except for simulate", func(ttt *testing.T) {
				gm2Gas := gm2.GasConsumed()
				gm2BaseFee := gm2.BaseFeeConsumed()
				gm2Fee := gm2.FeeConsumed()
				gm2FeeDist := gm2.FeeConsumedDistributions()
				gm2String := gm2.String()
				gm2Events := gm2.EventFeeSummary()

				assert.Equal(ttt, int(gm1Gas), int(gm2Gas), "GasConsumed")
				assert.Equal(ttt, gm1BaseFee, gm2BaseFee, "BaseFeeConsumed")
				assert.Equal(ttt, gm1Fee, gm2Fee, "FeeConsumed")
				assert.Equal(ttt, gm1FeeDist, gm2FeeDist, "FeeConsumedDistributions")
				assert.Equal(ttt, gm1String, gm2String, "String")
				assert.Equal(ttt, gm1Events, gm2Events, "EventFeeSummary")
			})
		})
	}
}
