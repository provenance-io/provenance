package antewrapper

import (
	"fmt"
	"math"

	"github.com/armon/go-metrics"

	"github.com/tendermint/tendermint/libs/log"

	sdkgas "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	msgfeestypes "github.com/provenance-io/provenance/x/msgfees/types"
)

type FeeGasMeter struct {
	// a context logger reference for info/debug output
	log log.Logger
	// tracks amount used per purpose
	used map[string]uint64
	// tracks number of usages per purpose
	calls map[string]uint64

	// tracks number of msg fee calls by msg type url
	feeCalls map[string]uint64
	// tracks the total amount of fees per msg type url
	usedFees map[string]sdk.Coins

	// this is the base fee charged in decorator
	baseFeeCharged sdk.Coins

	// Idea is that these two below fields are used to track consumption but panic only if greater than 4 m gas.
	// gas tracker, this is the gas usage tracker for the tx
	gasConsumed sdkgas.Gas

	// gas limit should be set to 4 mil
	gasLimit sdkgas.Gas

	simulate bool
}

// NewFeeGasMeterWrapper returns a reference to a new tracing gas meter that will track calls to the base gas meter
func NewFeeGasMeterWrapper(logger log.Logger, isSimulate bool) sdkgas.GasMeter {
	return &FeeGasMeter{
		log:            logger,
		used:           make(map[string]uint64),
		calls:          make(map[string]uint64),
		feeCalls:       make(map[string]uint64),
		usedFees:       make(map[string]sdk.Coins),
		baseFeeCharged: sdk.Coins{},
		simulate:       isSimulate,
		gasLimit:       50000,
	}
}

var _ sdkgas.GasMeter = &FeeGasMeter{}

// GasConsumed reports the amount of gas consumed at Log.Info level
func (g *FeeGasMeter) GasConsumed() sdkgas.Gas {
	usage := "tracingGasMeter:\n  Purpose"
	for i, d := range g.used {
		usage = fmt.Sprintf("%s\n   - %s (x%d) = %d", usage, i, g.calls[i], d)
	}
	usage = fmt.Sprintf("%s\n  Total: %d gas", usage, g.gasConsumed)

	g.log.Info(usage)
	// don't consume the gas
	return g.gasConsumed
}

// RefundGas will deduct the given amount from the gas consumed. If the amount is greater than the
// gas consumed, the function will panic.
//
// Use case: This functionality enables refunding gas to the transaction or block gas pools so that
// EVM-compatible chains can fully support the go-ethereum StateDb interface.
// See https://github.com/cosmos/cosmos-sdk/pull/9403 for reference.
func (g *FeeGasMeter) RefundGas(amount sdkgas.Gas, descriptor string) {
	if g.gasConsumed < amount {
		panic(sdkgas.ErrorNegativeGasConsumed{Descriptor: descriptor})
	}

	g.gasConsumed -= amount
}

// GasConsumedToLimit will report the actual consumption or the meter limit, whichever is less.
func (g *FeeGasMeter) GasConsumedToLimit() sdkgas.Gas {
	if g.IsPastLimit() {
		return g.gasLimit
	}
	return g.gasConsumed
}

// GasRemaining returns the gas left in the GasMeter.
func (g *FeeGasMeter) GasRemaining() sdkgas.Gas {
	if g.IsPastLimit() {
		return 0
	}
	return g.gasLimit - g.gasConsumed
}

// Limit for amount of gas that can be consumed (if zero then unlimited)
func (g *FeeGasMeter) Limit() sdkgas.Gas {
	return g.gasLimit
}

// ConsumeGas increments the amount of gas used on the meter associated with a given purpose.
func (g *FeeGasMeter) ConsumeGas(amount sdkgas.Gas, descriptor string) {
	cur := g.used[descriptor]
	g.used[descriptor] = cur + amount

	cur = g.calls[descriptor]
	g.calls[descriptor] = cur + 1

	telemetry.IncrCounterWithLabels([]string{"tx", "gas", "consumed"}, float32(amount), []metrics.Label{telemetry.NewLabel("purpose", descriptor)})

	g.ConsumeGasWithOutLimitCheck(amount, descriptor)
}

// ConsumeGasWithOutLimitCheck adds the given amount of gas to the gas consumed and panics if it overflows the limit or out of gas.
func (g *FeeGasMeter) ConsumeGasWithOutLimitCheck(amount sdkgas.Gas, descriptor string) {
	var overflow bool
	g.gasConsumed, overflow = addUint64Overflow(g.gasConsumed, amount)
	if overflow {
		g.gasConsumed = math.MaxUint64
		panic(sdkgas.ErrorGasOverflow{Descriptor: descriptor})
	}

	// check only the 4m (or param) limit
	if g.gasConsumed > g.gasLimit {
		panic(sdkgas.ErrorOutOfGas{Descriptor: descriptor})
	}
}

// addUint64Overflow performs the addition operation on two uint64 integers and
// returns a boolean on whether or not the result overflows.
func addUint64Overflow(a, b uint64) (uint64, bool) {
	if math.MaxUint64-a < b {
		return 0, true
	}

	return a + b, false
}

// IsPastLimit indicates consumption has passed the limit (if any)
func (g *FeeGasMeter) IsPastLimit() bool {
	return g.gasConsumed > g.gasLimit
}

// IsOutOfGas indicates the gas meter has tracked consumption at or above the limit
func (g *FeeGasMeter) IsOutOfGas() bool {
	return g.gasConsumed >= g.gasLimit
}

// String implements stringer interface
func (g *FeeGasMeter) String() string {
	return fmt.Sprintf("feeGasMeter:\n  limit: %d\n  consumed: %d fee consumed: %v", g.Limit(), g.GasConsumed(), g.FeeConsumed())
}

// ConsumeFee increments the amount of msg fee required by a msg type.
func (g *FeeGasMeter) ConsumeFee(amount sdk.Coins, msgType string, recipient string) {
	key := msgfeestypes.GetCompositeKey(msgType, recipient)
	cur := g.usedFees[key]
	if cur.Empty() {
		g.usedFees[key] = sdk.NewCoins(amount...)
	} else {
		g.usedFees[key] = cur.Add(amount...)
	}
	g.feeCalls[key]++
}

func (g *FeeGasMeter) FeeConsumedForType(msgType string, recipient string) sdk.Coins {
	return g.usedFees[msgfeestypes.GetCompositeKey(msgType, recipient)]
}

// FeeConsumed returns total fee consumed in the current fee gas meter, is returned Sorted.
func (g *FeeGasMeter) FeeConsumed() sdk.Coins {
	var consumedFees sdk.Coins
	for _, coins := range g.usedFees {
		consumedFees = consumedFees.Add(coins...)
	}
	return consumedFees.Sort()
}

// FeeConsumedDistributions returns fees by distribution either to fee module when key is empty or address
func (g *FeeGasMeter) FeeConsumedDistributions() map[string]sdk.Coins {
	additionalFeeDistributions := make(map[string]sdk.Coins)
	for key, coins := range g.usedFees {
		_, addressKey := msgfeestypes.SplitCompositeKey(key)
		additionalFeeDistributions[addressKey] = additionalFeeDistributions[addressKey].Add(coins...)
	}
	return additionalFeeDistributions
}

// FeeConsumedByMsg total fee consumed for a particular MsgType
func (g *FeeGasMeter) FeeConsumedByMsg() map[string]sdk.Coins {
	consumedByMsg := make(map[string]sdk.Coins)
	for msg, coins := range g.usedFees {
		consumedByMsg[msg] = sdk.NewCoins(coins...)
	}
	return consumedByMsg
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

// EventFeeSummary returns total fee consumed in the current fee gas meter, is returned Sorted.
func (g *FeeGasMeter) EventFeeSummary() *msgfeestypes.EventMsgFees {
	return msgfeestypes.NewEventMsgs(g.feeCalls, g.usedFees)
}
