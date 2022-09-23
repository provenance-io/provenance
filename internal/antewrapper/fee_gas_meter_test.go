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
			require.NotPanics(t, func() { meter.ConsumeFee(usageFee, msgType, "") }, "panicked on adding fees")
		}

		require.Panics(t, func() { meter.ConsumeGas(1, "") }, "Exceeded but not panicked. tc #%d", tcnum)
		require.Equal(t, meter.GasConsumedToLimit(), meter.Limit(), "sdkgas.Gas consumption (to limit) not match limit")
		require.Equal(t, meter.GasConsumed(), meter.Limit()+1, "sdkgas.Gas consumption not match limit+1")
		assert.Equal(t, meter.FeeConsumed().Sort(), usedFee.Sort(), "FeeConsumed does not match all Fees")
		meter2 := NewFeeGasMeterWrapper(log.TestingLogger(), sdkgas.NewGasMeter(100), false).(*FeeGasMeter)
		meter2.ConsumeGas(sdkgas.Gas(50), "consume half max")
		meter2.ConsumeFee(sdk.NewCoin(msgfeestype.NhashDenom, sdk.NewInt(1000000)), "/cosmos.bank.v1beta1.MsgSend", "")
		require.Equalf(t, "feeGasMeter:\n  limit: 100\n  consumed: 50 fee consumed: 1000000nhash", meter2.String(), "expect string output to match")
		meter2.RefundGas(uint64(20), "refund")
		require.Equalf(t, "feeGasMeter:\n  limit: 100\n  consumed: 30 fee consumed: 1000000nhash", meter2.String(), "expect string output to match")
		require.Equalf(t, "1000000nhash", meter2.FeeConsumed().String(), "expect string output to match")
		require.Panics(t, func() { meter2.ConsumeGas(sdkgas.Gas(70)+2, "panic") })
		meter2.ConsumeFee(sdk.NewCoin(msgfeestype.NhashDenom, sdk.NewInt(2000000)), "/cosmos.bank.v1beta1.MsgSend", "")
		require.Equalf(t, "feeGasMeter:\n  limit: 100\n  consumed: 102 fee consumed: 3000000nhash", meter2.String(), "expect string output to match")
		meter2.RefundGas(uint64(20), "refund")
		require.Equalf(t, "feeGasMeter:\n  limit: 100\n  consumed: 82 fee consumed: 3000000nhash", meter2.String(), "expect string output to match")
		require.Equalf(t, "3000000nhash", meter2.FeeConsumed().String(), "expect string output to match")
		require.Equalf(t, "map[/cosmos.bank.v1beta1.MsgSend:3000000nhash]", fmt.Sprintf("%v", meter2.FeeConsumedByMsg()), "expect string output to match")
		meter2.ConsumeFee(sdk.NewCoin("doge", sdk.NewInt(2000000)), "/provenance.marker.v1.MsgAddMarkerRequest", "")
		meter2.ConsumeFee(sdk.NewCoin("jackthecat", sdk.NewInt(420)), "/provenance.marker.v1.MsgAddMarkerRequest", "")
		meter2FeesConsumed := meter2.FeeConsumed()
		require.Equalf(t, "2000000doge,420jackthecat,3000000nhash", meter2FeesConsumed.String(), "expect string output to match")
		require.Equalf(t, "2000000doge,420jackthecat", meter2.FeeConsumedForType("/provenance.marker.v1.MsgAddMarkerRequest", "").String(), "expect string output to match")
		require.Equalf(t, "3000000nhash", meter2.FeeConsumedForType("/cosmos.bank.v1beta1.MsgSend", "").String(), "expect string output to match")
		require.Equalf(t, false, meter2.IsSimulate(), "simulate should be false")
	}
}
