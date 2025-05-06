package simulation

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"time"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/provenance-io/provenance/internal/pioconfig"
	internalrand "github.com/provenance-io/provenance/internal/rand"
	"github.com/provenance-io/provenance/x/trigger/types"
)

const (
	TriggerID        = "trigger_id"
	QueueStart       = "trigger_queue_start"
	NumIniTrig       = "num_ini_trig"
	NumIniTrigQueued = "num_ini_trig_queued"
	Triggers         = "triggers"
	QueuedTriggers   = "queued_triggers"
)

// TriggerIDStartFn randomized starting trigger id
func TriggerIDStartFn(r *rand.Rand) uint64 {
	// max 5 ids for the triggers max 5 ids for the queue = min of 10 here.
	return uint64(internalrand.IntBetween(r, 10, 10000000000))
}

// QueueStartFn randomized Queue Start Index
func QueueStartFn(r *rand.Rand) uint64 {
	return uint64(internalrand.IntBetween(r, 1, 10000000000))
}

// NewRandomEvent returns a random event
func NewRandomEvent(r *rand.Rand, now time.Time) types.TriggerEventI {
	if r.Intn(2) > 0 {
		minimumTime := int(time.Second * 10)
		maximumTime := int(time.Minute * 5)
		randTime := now.Add(time.Duration(internalrand.IntBetween(r, minimumTime, maximumTime)))
		return &types.BlockTimeEvent{Time: randTime.UTC()}
	}
	height := uint64(internalrand.IntBetween(r, 10, 150))
	return &types.BlockHeightEvent{BlockHeight: height}
}

// NewRandomAction returns a random action
func NewRandomAction(r *rand.Rand, from string, to string) sdk.Msg {
	amount := int64(internalrand.IntBetween(r, 100, 1000))
	return &banktypes.MsgSend{
		FromAddress: from,
		ToAddress:   to,
		Amount:      sdk.NewCoins(sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().BondDenom, amount)),
	}
}

// NewRandomTrigger returns a random trigger
func NewRandomTrigger(r *rand.Rand, simState *module.SimulationState, accs []simtypes.Account, id types.TriggerID) types.Trigger {
	raccs, err := internalrand.SelectAccounts(r, accs, 2)
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
func RandomNewQueuedTriggers(r *rand.Rand, simState *module.SimulationState, accs []simtypes.Account, queuedTriggerIDs []uint64) []types.QueuedTrigger {
	rv := make([]types.QueuedTrigger, len(queuedTriggerIDs))
	for i, id := range queuedTriggerIDs {
		rv[i] = NewRandomQueuedTrigger(r, simState, accs, id)
	}
	return rv
}

// randomTriggerAndQueueIDs generates random unique ids (1 to maxCount inclusive) and divvies them up with the provided counts.
func randomTriggerAndQueueIDs(r *rand.Rand, triggerCount, queueCount, maxCount int) ([]uint64, []uint64) {
	ids := randomTriggerIDs(r, triggerCount+queueCount, maxCount)
	return ids[:triggerCount], ids[triggerCount:]
}

// randomTriggerIDs generates count different random trigger ids between 1 and maxCount inclusive
// If count > maxCount, only maxCount will be returned.
func randomTriggerIDs(r *rand.Rand, count int, maxCount int) []uint64 {
	if count > maxCount {
		panic(fmt.Errorf("cannot generate %d unique trigger ids with a max of %d", count, maxCount))
	}
	if count == 0 {
		return []uint64{}
	}
	rv := make([]uint64, count)
	// If count is 33+% of maxCount, generate a permutation up to maxCount (exclusive).
	// Then increment each of the first <count> of them (since we want 1 to maxCount inclusive) and return those.
	if maxCount/3 <= count {
		nums := r.Perm(maxCount)
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
			rv[i] = uint64(r.Intn(maxCount) + 1)
		}
		seen[rv[i]] = true
	}
	return rv
}

// RandomizedGenState generates a random GenesisState for trigger
func RandomizedGenState(simState *module.SimulationState) {
	var triggerID uint64
	simState.AppParams.GetOrGenerate(
		TriggerID, &triggerID, simState.Rand,
		func(r *rand.Rand) { triggerID = TriggerIDStartFn(r) },
	)

	var queueStart uint64
	simState.AppParams.GetOrGenerate(
		QueueStart, &queueStart, simState.Rand,
		func(r *rand.Rand) { queueStart = QueueStartFn(r) },
	)

	var numIniTrig int
	simState.AppParams.GetOrGenerate(
		NumIniTrig, &numIniTrig, simState.Rand,
		func(r *rand.Rand) { numIniTrig = r.Intn(6) },
	)

	var numIniTrigQueued int
	simState.AppParams.GetOrGenerate(
		NumIniTrigQueued, &numIniTrigQueued, simState.Rand,
		func(r *rand.Rand) { numIniTrigQueued = r.Intn(6) },
	)

	if triggerID < uint64(numIniTrig)+uint64(numIniTrigQueued) {
		triggerID = uint64(numIniTrig) + uint64(numIniTrigQueued)
	}

	if triggerID > uint64(math.MaxInt) {
		// This should only ever happen if using a pre-defined value for the trigger id.
		panic(fmt.Errorf("cannot run sims with a %s param [%d] larger than max int (%d)", TriggerID, triggerID, math.MaxInt))
	}
	triggerIDInt := int(triggerID) //nolint:gosec // G115: Overflow handled above.
	triggerIDs, queueTriggerIDs := randomTriggerAndQueueIDs(simState.Rand, numIniTrig, numIniTrigQueued, triggerIDInt)
	var triggers []types.Trigger
	simState.AppParams.GetOrGenerate(
		Triggers, &triggers, simState.Rand,
		func(r *rand.Rand) { triggers = RandomNewTriggers(r, simState, simState.Accounts, triggerIDs) },
	)

	var queuedTriggers []types.QueuedTrigger
	simState.AppParams.GetOrGenerate(
		QueuedTriggers, &queuedTriggers, simState.Rand,
		func(r *rand.Rand) {
			queuedTriggers = RandomNewQueuedTriggers(r, simState, simState.Accounts, queueTriggerIDs)
		},
	)

	genesis := types.NewGenesisState(triggerID, queueStart, triggers, queuedTriggers)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(genesis)

	bz, err := json.MarshalIndent(simState.GenState[types.ModuleName], "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Selected randomly generated trigger parameters:\n%s\n", bz)
}
