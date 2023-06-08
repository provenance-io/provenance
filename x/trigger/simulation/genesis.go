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
	TriggerID  = "trigger_id"
	QueueStart = "queue_start"
)

// TriggerIDStartFn randomized starting trigger id
func TriggerIDStartFn(r *rand.Rand) uint64 {
	return uint64(simtypes.RandIntBetween(r, 1, 10000000000))
}

// QueueStartFn randomized Queue Start Index
func QueueStartFn(r *rand.Rand) uint64 {
	return uint64(r.Intn(10000000000))
}

// NewRandomGasLimit randomized Gas Limit
func NewRandomGasLimit(r *rand.Rand) uint64 {
	return uint64(r.Intn(1000000))
}

// NewRandomEvent returns a random event
func NewRandomEvent(r *rand.Rand, now time.Time) types.TriggerEventI {
	if r.Intn(2) > 0 {
		minimumTime := int(time.Second * 10)
		maximumTime := int(time.Minute * 5)
		randTime := now.Add(time.Duration(simtypes.RandIntBetween(r, minimumTime, maximumTime)))
		return &types.BlockTimeEvent{Time: randTime.UTC()}
	}
	height := uint64(simtypes.RandIntBetween(r, 10, 150))
	return &types.BlockHeightEvent{BlockHeight: height}
}

// NewRandomAction returns a random action
func NewRandomAction(r *rand.Rand, from string, to string) sdk.Msg {
	amount := int64(simtypes.RandIntBetween(r, 100, 1000))
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
		panic(fmt.Sprintf("NewRandomTrigger failed: %v", err))
	}
	event := NewRandomEvent(r, simState.GenTimestamp.UTC())
	action := NewRandomAction(r, raccs[0].Address.String(), raccs[1].Address.String())

	actions, err := sdktx.SetMsgs([]sdk.Msg{action})
	if err != nil {
		panic("SetMsgs failed for NewRandomTrigger")
	}
	anyEvent, err := codectypes.NewAnyWithValue(event)
	if err != nil {
		panic("NewAnyWithValue failed for NewRandomTrigger")
	}

	return types.NewTrigger(id, raccs[0].Address.String(), anyEvent, actions)
}

// NewRandomQueuedTrigger returns a random queued trigger
func NewRandomQueuedTrigger(r *rand.Rand, simState *module.SimulationState, accs []simtypes.Account, id types.TriggerID) types.QueuedTrigger {
	raccs, err := RandomAccs(r, accs, 2)
	if err != nil {
		panic(fmt.Sprintf("NewRandomQueuedTrigger failed: %v", err))
	}
	var event types.TriggerEventI = &types.BlockHeightEvent{BlockHeight: 0}
	action := NewRandomAction(r, raccs[0].Address.String(), raccs[1].Address.String())

	actions, err := sdktx.SetMsgs([]sdk.Msg{action})
	if err != nil {
		panic("SetMsgs failed for NewRandomQueuedTrigger")
	}
	anyEvent, err := codectypes.NewAnyWithValue(event)
	if err != nil {
		panic("NewAnyWithValue failed for NewRandomQueuedTrigger")
	}

	trigger := types.NewTrigger(id, raccs[0].Address.String(), anyEvent, actions)

	now := simState.GenTimestamp

	return types.NewQueuedTrigger(trigger, now.UTC(), 1)
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

	triggers := make([]types.Trigger, simState.Rand.Intn(5))
	for i := range triggers {
		triggers[i] = NewRandomTrigger(simState.Rand, simState, simState.Accounts, triggerID)
		triggerID++
	}
	queuedTriggers := make([]types.QueuedTrigger, simState.Rand.Intn(5))
	for i := range queuedTriggers {
		queuedTriggers[i] = NewRandomQueuedTrigger(simState.Rand, simState, simState.Accounts, triggerID)
		triggerID++
	}

	gasLimits := []types.GasLimit{}
	for _, trigger := range triggers {
		gasLimits = append(gasLimits, types.GasLimit{
			TriggerId: trigger.GetId(),
			Amount:    NewRandomGasLimit(simState.Rand),
		})
	}
	for _, item := range queuedTriggers {
		gasLimits = append(gasLimits, types.GasLimit{
			TriggerId: item.Trigger.Id,
			Amount:    NewRandomGasLimit(simState.Rand),
		})
	}

	genesis := types.NewGenesisState(triggerID, queueStart, triggers, gasLimits, queuedTriggers)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(genesis)

	bz, err := json.MarshalIndent(simState.GenState[types.ModuleName], "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Selected randomly generated triggers:\n%s\n", bz)
}
