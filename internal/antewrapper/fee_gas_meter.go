package antewrapper

import (
	"fmt"

	"github.com/armon/go-metrics"
	"github.com/cosmos/cosmos-sdk/telemetry"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tendermint/tendermint/libs/log"

	sdkgas "github.com/cosmos/cosmos-sdk/store/types"
)

type FeeGasMeter struct {
	// a context logger reference for info/debug output
	log log.Logger
	// the gas meter being wrapped
	base sdkgas.GasMeter
	// tracks amount used per purpose
	used map[string]uint64
	// tracks number of usages per purpose
	calls map[string]uint64

	usedFees map[string]sdk.Coin // map of msg fee type url --> fees charged

	// this is the base fee charged in decorator
	baseFeeCharged sdk.Coins

	simulate bool
}

// NewFeeGasMeterWrapper returns a reference to a new tracing gas meter that will track calls to the base gas meter
func NewFeeGasMeterWrapper(logger log.Logger, baseMeter sdkgas.GasMeter, isSimulate bool) sdkgas.GasMeter {
	return &FeeGasMeter{
		log:            logger,
		base:           baseMeter,
		used:           make(map[string]uint64),
		calls:          make(map[string]uint64),
		usedFees:       make(map[string]sdk.Coin),
		baseFeeCharged: sdk.Coins{},
		simulate:       isSimulate,
	}
}

var _ sdkgas.GasMeter = &FeeGasMeter{}

// GasConsumed reports the amount of gas consumed at Log.Info level
func (g *FeeGasMeter) GasConsumed() sdkgas.Gas {
	usage := "tracingGasMeter:\n  Purpose"
	for i, d := range g.used {
		usage = fmt.Sprintf("%s\n   - %s (x%d) = %d", usage, i, g.calls[i], d)
	}
	usage = fmt.Sprintf("%s\n  Total: %d gas", usage, g.base.GasConsumed())

	g.log.Info(usage)
	return g.base.GasConsumed()
}

// RefundGas refunds an amount of gas
func (g *FeeGasMeter) RefundGas(amount uint64, descriptor string) {
	g.base.RefundGas(amount, descriptor)
}

// GasConsumedToLimit will report the actual consumption or the meter limit, whichever is less.
func (g *FeeGasMeter) GasConsumedToLimit() sdkgas.Gas {
	return g.base.GasConsumedToLimit()
}

// Limit for amount of gas that can be consumed (if zero then unlimited)
func (g *FeeGasMeter) Limit() sdkgas.Gas {
	return g.base.Limit()
}

// ConsumeGas increments the amount of gas used on the meter associated with a given purpose.
func (g *FeeGasMeter) ConsumeGas(amount sdkgas.Gas, descriptor string) {
	cur := g.used[descriptor]
	g.used[descriptor] = cur + amount

	cur = g.calls[descriptor]
	g.calls[descriptor] = cur + 1

	telemetry.IncrCounterWithLabels([]string{"tx", "gas", "consumed"}, float32(amount), []metrics.Label{telemetry.NewLabel("purpose", descriptor)})

	g.base.ConsumeGas(amount, descriptor)
}

// IsPastLimit indicates consumption has passed the limit (if any)
func (g *FeeGasMeter) IsPastLimit() bool {
	return g.base.IsPastLimit()
}

// IsOutOfGas indicates the gas meter has tracked consumption at or above the limit
func (g *FeeGasMeter) IsOutOfGas() bool {
	return g.base.IsOutOfGas()
}

// String implements stringer interface
func (g *FeeGasMeter) String() string {
	return fmt.Sprintf("feeGasMeter:\n  limit: %d\n  consumed: %d fee consumed: %v", g.base.Limit(), g.base.GasConsumed(), g.FeeConsumed())
}

// ConsumeFee increments the amount of msg fee required by a msg type.
func (g *FeeGasMeter) ConsumeFee(amount sdk.Coin, msgType string) {
	cur := g.usedFees[msgType]
	if !cur.Amount.IsNil() {
		g.usedFees[msgType] = cur.Add(amount)
	} else {
		g.usedFees[msgType] = amount
	}
}

func (g *FeeGasMeter) FeeConsumedForType(msgType string) sdk.Coin {
	return g.usedFees[msgType]
}

// FeeConsumed returns total fee consumed in the current fee gas meter, is returned Sorted.
func (g *FeeGasMeter) FeeConsumed() sdk.Coins {
	consumedFees := make(sdk.Coins, len(g.usedFees))
	var i = 0
	for _, coin := range g.usedFees {
		consumedFees[i] = sdk.NewCoin(coin.Denom, coin.Amount)
		i++
	}
	return consumedFees.Sort()
}

// FeeConsumedByMsg total fee consumed for a particular MsgType
func (g *FeeGasMeter) FeeConsumedByMsg() map[string]sdk.Coin {
	return g.usedFees
}

func (g *FeeGasMeter) IsSimulate() bool {
	return g.simulate
}

func (g *FeeGasMeter) ConsumeBaseFee(amount sdk.Coins) sdk.Coins {
	g.baseFeeCharged = amount
	return g.baseFeeCharged
}

func (g *FeeGasMeter) BaseFeeConsumed() sdk.Coins {
	return g.baseFeeCharged
}
