package antewrapper

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/armon/go-metrics"
	"github.com/tendermint/tendermint/libs/log"

	sdkgas "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
)

type feeGasMeter struct {
	// a context logger reference for info/debug output
	log log.Logger
	// the gas meter being wrapped
	base sdkgas.GasMeter
	// tracks amount used per purpose
	used map[string]uint64
	// tracks number of usages per purpose
	calls map[string]uint64

	usedFees map[string]sdk.Coin // map of msg fee type url --> fees charged

	feesWanted map[string]sdk.Coin // map of msg fee type url --> fees wanted

	// in passing cases usedFees and feesWanted should match
}

// NewFeeTracingMeterWrapper returns a reference to a new tracing gas meter that will track calls to the base gas meter
func NewFeeTracingMeterWrapper(logger log.Logger, baseMeter sdkgas.GasMeter) sdkgas.GasMeter {
	return &feeGasMeter{
		log:        logger,
		base:       baseMeter,
		used:       make(map[string]uint64),
		calls:      make(map[string]uint64),
		usedFees:   make(map[string]sdk.Coin),
		feesWanted: make(map[string]sdk.Coin),
	}
}

var _ sdkgas.GasMeter = &feeGasMeter{}

// GasConsumed reports the amount of gas consumed at Log.Info level
func (g *feeGasMeter) GasConsumed() sdkgas.Gas {
	usage := "tracingGasMeter:\n  Purpose"
	for i, d := range g.used {
		usage = fmt.Sprintf("%s\n   - %s (x%d) = %d", usage, i, g.calls[i], d)
	}
	usage = fmt.Sprintf("%s\n  Total: %d gas", usage, g.base.GasConsumed())

	g.log.Info(usage)
	return g.base.GasConsumed()
}

// RefundGas refunds an amount of gas
func (g *feeGasMeter) RefundGas(amount uint64, descriptor string) {
	g.base.RefundGas(amount, descriptor)
}

// GasConsumedToLimit will report the actual consumption or the meter limit, whichever is less.
func (g *feeGasMeter) GasConsumedToLimit() sdkgas.Gas {
	return g.base.GasConsumedToLimit()
}

// Limit for amount of gas that can be consumed (if zero then unlimited)
func (g *feeGasMeter) Limit() sdkgas.Gas {
	return g.base.Limit()
}

// ConsumeGas increments the amount of gas used on the meter associated with a given purpose.
func (g *feeGasMeter) ConsumeGas(amount sdkgas.Gas, descriptor string) {
	cur := g.used[descriptor]
	g.used[descriptor] = cur + amount

	cur = g.calls[descriptor]
	g.calls[descriptor] = cur + 1

	telemetry.IncrCounterWithLabels([]string{"tx", "gas", "consumed"}, float32(amount), []metrics.Label{telemetry.NewLabel("purpose", descriptor)})

	g.base.ConsumeGas(amount, descriptor)
}

// IsPastLimit indicates consumption has passed the limit (if any)
func (g *feeGasMeter) IsPastLimit() bool {
	return g.base.IsPastLimit()
}

// IsOutOfGas indicates the gas meter has tracked consumption at or above the limit
func (g *feeGasMeter) IsOutOfGas() bool {
	return g.base.IsOutOfGas()
}

// String implements stringer interface
func (g *feeGasMeter) String() string {
	return fmt.Sprintf("tracingGasMeter:\n  limit: %d\n  consumed: %d", g.base.Limit(), g.base.GasConsumed())
}


// ConsumeFee increments the amount of gas used on the meter associated with a given purpose.
func (g *feeGasMeter) ConsumeFee(amount sdk.Coin, msgType string) {
	cur := g.usedFees[msgType]
	g.usedFees[msgType] = cur.Add(amount)
}


func (g *feeGasMeter) feeWanted(amount sdk.Coin, msgType string) {
	cur := g.feesWanted[msgType]
	g.feesWanted[msgType] = cur.Add(amount)
}
