package simulation

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/provenance-io/provenance/internal/pioconfig"
	"github.com/provenance-io/provenance/x/trigger/types"
)

const (
	TriggerID        = "trigger_id"
	QueueStart       = "trigger_queue_start"
	NumIniTrig       = "num_ini_trig"
	NumIniTrigQueued = "num_ini_trig_queued"
	Triggers         = "triggers"
	QueuedTriggers   = "queued_triggers"
	GasLimits        = "trigger_gas_limits"
)

// TriggerIDStartFn randomized starting trigger id
func TriggerIDStartFn(r *rand.Rand) uint64 {
	// max 5 ids for the triggers max 5 ids for the queue = min of 10 here.
	return uint64(randIntBetween(r, 10, 10000000000))
}

// QueueStartFn randomized Queue Start Index
func QueueStartFn(r *rand.Rand) uint64 {
	return uint64(randIntBetween(r, 1, 10000000000))
}

// NewRandomEvent returns a random event
func NewRandomEvent(r *rand.Rand, now time.Time) types.TriggerEventI {
	if r.Intn(2) > 0 {
		minimumTime := int(time.Second * 10)
		maximumTime := int(time.Minute * 5)
		randTime := now.Add(time.Duration(randIntBetween(r, minimumTime, maximumTime)))
		return &types.BlockTimeEvent{Time: randTime.UTC()}
	}
	height := uint64(randIntBetween(r, 10, 150))
	return &types.BlockHeightEvent{BlockHeight: height}
}

// NewRandomAction returns a random action
func NewRandomAction(r *rand.Rand, from string, to string) sdk.Msg {
	amount := int64(randIntBetween(r, 100, 1000))
	return &banktypes.MsgSend{
		FromAddress: from,
		ToAddress:   to,
		Amount:      sdk.NewCoins(sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().BondDenom, amount)),
	}
}

// NewRandomTrigger returns a random trigger
func NewRandomTrigger(r *rand.Rand, simState *module.SimulationState, accs []simtypes.Account, id types.TriggerID) types.Trigger {
	raccs, err := RandomAccs(r, accs, 2)
	if err != nil {
		panic(fmt.Errorf("NewRandomTrigger failed: %w", err))
	}
	event := NewRandomEvent(r, simState.GenTimestamp.UTC())
	action := NewRandomAction(r, raccs[0].Address.String(), raccs[1].Address.String())

	actions, err := sdktx.SetMsgs([]sdk.Msg{action})
	if err != nil {
		panic(fmt.Errorf("SetMsgs failed for NewRandomTrigger: %w", err))
	}
	anyEvent, err := codectypes.NewAnyWithValue(event)
	if err != nil {
		panic(fmt.Errorf("NewAnyWithValue failed for NewRandomTrigger: %w", err))
	}

	return types.NewTrigger(id, raccs[0].Address.String(), anyEvent, actions)
}

// RandomNewTriggers generates a random trigger for each provided id.
func RandomNewTriggers(r *rand.Rand, simState *module.SimulationState, accs []simtypes.Account, triggerIDs []uint64) []types.Trigger {
	rv := make([]types.Trigger, len(triggerIDs))
	for i, id := range triggerIDs {
		rv[i] = NewRandomTrigger(r, simState, accs, id)
	}
	return rv
}

// NewRandomQueuedTrigger returns a random queued trigger
func NewRandomQueuedTrigger(r *rand.Rand, simState *module.SimulationState, accs []simtypes.Account, id types.TriggerID) types.QueuedTrigger {
	trigger := NewRandomTrigger(r, simState, accs, id)

	now := simState.GenTimestamp

	return types.NewQueuedTrigger(trigger, now.UTC(), 1)
}

// RandomNewQueuedTriggers generates a random queued trigger for each provided id.
func RandomNewQueuedTriggers(r *rand.Rand, simState *module.SimulationState, accs []simtypes.Account, queuedTriggerIds []uint64) []types.QueuedTrigger {
	rv := make([]types.QueuedTrigger, len(queuedTriggerIds))
	for i, id := range queuedTriggerIds {
		rv[i] = NewRandomQueuedTrigger(r, simState, accs, id)
	}
	return rv
}

// NewRandomGasLimit randomized Gas Limit
func NewRandomGasLimit(r *rand.Rand) uint64 {
	return uint64(r.Intn(1000000))
}

// RandomGasLimits generates random gas limits for all the provided triggers and queued triggers.
func RandomGasLimits(r *rand.Rand, triggers []types.Trigger, queuedTriggers []types.QueuedTrigger) []types.GasLimit {
	rv := make([]types.GasLimit, 0, len(triggers)+len(queuedTriggers))
	for _, trigger := range triggers {
		rv = append(rv, types.GasLimit{
			TriggerId: trigger.GetId(),
			Amount:    NewRandomGasLimit(r),
		})
	}
	for _, item := range queuedTriggers {
		rv = append(rv, types.GasLimit{
			TriggerId: item.Trigger.Id,
			Amount:    NewRandomGasLimit(r),
		})
	}
	r.Shuffle(len(rv), func(i, j int) {
		rv[i], rv[j] = rv[j], rv[i]
	})
	return rv
}

// randomTriggerAndQueueIDs generates random unique ids (1 to max inclusive) and divvies them up with the provided counts.
func randomTriggerAndQueueIDs(r *rand.Rand, triggerCount, queueCount, max int) ([]uint64, []uint64) {
	ids := randomTriggerIDs(r, triggerCount+queueCount, max)
	return ids[:triggerCount], ids[triggerCount:]
}

// randomTriggerIDs generates count different random trigger ids between 1 and max inclusive
// If count > max, only max will be returned.
func randomTriggerIDs(r *rand.Rand, count int, max int) []uint64 {
	if count > max {
		panic(fmt.Errorf("cannot generate %d unique trigger ids with a max of %d", count, max))
	}
	if count == 0 {
		return []uint64{}
	}
	rv := make([]uint64, count)
	// If count is 33+% of max, generate a permutation up to max (exclusive).
	// Then increment each of the first <count> of them (since we want 1 to max inclusive) and return those.
	if max/3 <= count {
		nums := r.Perm(max)
		for i := range rv {
			rv[i] = uint64(nums[i]) + 1
		}
		return rv
	}
	// Less than 33% of numbers are needed. Just pick randomly until we have count different ones.
	seen := make(map[uint64]bool)
	seen[0] = true
	for i := range rv {
		for seen[rv[i]] {
			rv[i] = uint64(r.Intn(max) + 1)
		}
		seen[rv[i]] = true
	}
	return rv
}

// RandomizedGenState generates a random GenesisState for trigger
func RandomizedGenState(simState *module.SimulationState) {
	var triggerID uint64
	simState.AppParams.GetOrGenerate(
		simState.Cdc, TriggerID, &triggerID, simState.Rand,
		func(r *rand.Rand) { triggerID = TriggerIDStartFn(r) },
	)

	var queueStart uint64
	simState.AppParams.GetOrGenerate(
		simState.Cdc, QueueStart, &queueStart, simState.Rand,
		func(r *rand.Rand) { queueStart = QueueStartFn(r) },
	)

	var numIniTrig int
	simState.AppParams.GetOrGenerate(
		simState.Cdc, NumIniTrig, &numIniTrig, simState.Rand,
		func(r *rand.Rand) { numIniTrig = r.Intn(6) },
	)

	var numIniTrigQueued int
	simState.AppParams.GetOrGenerate(
		simState.Cdc, NumIniTrigQueued, &numIniTrigQueued, simState.Rand,
		func(r *rand.Rand) { numIniTrigQueued = r.Intn(6) },
	)

	if triggerID < uint64(numIniTrig)+uint64(numIniTrigQueued) {
		triggerID = uint64(numIniTrig) + uint64(numIniTrigQueued)
	}

	triggerIDs, queueTriggerIDs := randomTriggerAndQueueIDs(simState.Rand, numIniTrig, numIniTrigQueued, int(triggerID))
	var triggers []types.Trigger
	simState.AppParams.GetOrGenerate(
		simState.Cdc, Triggers, &triggers, simState.Rand,
		func(r *rand.Rand) { triggers = RandomNewTriggers(r, simState, simState.Accounts, triggerIDs) },
	)

	var queuedTriggers []types.QueuedTrigger
	simState.AppParams.GetOrGenerate(
		simState.Cdc, QueuedTriggers, &queuedTriggers, simState.Rand,
		func(r *rand.Rand) {
			queuedTriggers = RandomNewQueuedTriggers(r, simState, simState.Accounts, queueTriggerIDs)
		},
	)

	var gasLimits []types.GasLimit
	simState.AppParams.GetOrGenerate(
		simState.Cdc, GasLimits, &gasLimits, simState.Rand,
		func(r *rand.Rand) { gasLimits = RandomGasLimits(r, triggers, queuedTriggers) },
	)

	genesis := types.NewGenesisState(triggerID, queueStart, triggers, gasLimits, queuedTriggers)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(genesis)

	bz, err := json.MarshalIndent(simState.GenState[types.ModuleName], "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Selected randomly generated trigger parameters:\n%s\n", bz)
}

// randIntBetween generates a random number between min and max inclusive.
func randIntBetween(r *rand.Rand, min, max int) int {
	return r.Intn(max-min+1) + min
}
