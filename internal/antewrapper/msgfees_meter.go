package antewrapper

import (
	"fmt"

	"github.com/armon/go-metrics"
	"github.com/tendermint/tendermint/libs/log"

	sdkgas "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/telemetry"

	msgbasedfeetypes "github.com/provenance-io/provenance/x/msgfees/types"
)

type msgBasedFeeGasMeter struct {
	// a context logger reference for info/debug output
	log log.Logger
	// the gas meter being wrapped
	base sdkgas.GasMeter
	// tracks amount used per purpose
	used map[string]uint64
	// tracks number of usages per purpose
	calls map[string]uint64

	msgfeesKeeper msgbasedfeetypes.MsgBasedFeeKeeper
}

// NewTracingMeterWrapper returns a reference to a new tracing gas meter that will track calls to the base gas meter
func NewMsgBasedFeeMeterWrapper(logger log.Logger, baseMeter sdkgas.GasMeter, msgBasedFeeKeeper msgbasedfeetypes.MsgBasedFeeKeeper) sdkgas.GasMeter {
	return &msgBasedFeeGasMeter{
		log:           logger,
		base:          baseMeter,
		used:          make(map[string]uint64),
		calls:         make(map[string]uint64),
		msgfeesKeeper: msgBasedFeeKeeper,
	}
}

var _ sdkgas.GasMeter = &msgBasedFeeGasMeter{}

// GasConsumed reports the amount of gas consumed at Log.Info level
func (g *msgBasedFeeGasMeter) GasConsumed() sdkgas.Gas {
	g.log.Info("msg fee meter !!!!!!!!!!!!!!!!!")
	return g.base.GasConsumed()
}

// RefundGas refunds an amount of gas
func (g *msgBasedFeeGasMeter) RefundGas(amount uint64, descriptor string) {
	g.base.RefundGas(amount, descriptor)
}

// GasConsumedToLimit will report the actual consumption or the meter limit, whichever is less.
func (g *msgBasedFeeGasMeter) GasConsumedToLimit() sdkgas.Gas {
	return g.base.GasConsumedToLimit()
}

// Limit for amount of gas that can be consumed (if zero then unlimited)
func (g *msgBasedFeeGasMeter) Limit() sdkgas.Gas {
	return g.base.Limit()
}

// ConsumeGas increments the amount of gas used on the meter associated with a given purpose.
func (g *msgBasedFeeGasMeter) ConsumeGas(amount sdkgas.Gas, descriptor string) {
	cur := g.used[descriptor]
	g.used[descriptor] = cur + amount

	cur = g.calls[descriptor]
	g.calls[descriptor] = cur + 1

	telemetry.IncrCounterWithLabels([]string{"tx", "gas", "consumed"}, float32(amount), []metrics.Label{telemetry.NewLabel("purpose", descriptor)})

	g.base.ConsumeGas(amount, descriptor)
}

// IsPastLimit indicates consumption has passed the limit (if any)
func (g *msgBasedFeeGasMeter) IsPastLimit() bool {
	return g.base.IsPastLimit()
}

// IsOutOfGas indicates the gas meter has tracked consumption at or above the limit
func (g *msgBasedFeeGasMeter) IsOutOfGas() bool {
	return g.base.IsOutOfGas()
}

// String implements stringer interface
func (g *msgBasedFeeGasMeter) String() string {
	return fmt.Sprintf("tracingGasMeter:\n  limit: %d\n  consumed: %d", g.base.Limit(), g.base.GasConsumed())
}
