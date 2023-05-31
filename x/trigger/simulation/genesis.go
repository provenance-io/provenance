package simulation

import (
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

// TriggerIDFn randomized MinActions
func TriggerIDFn(r *rand.Rand) uint64 {
	return uint64(simtypes.RandIntBetween(r, 2, 10000000000))
}

// QueueStartFn randomized MinActions
func QueueStartFn(r *rand.Rand) uint64 {
	return uint64(simtypes.RandIntBetween(r, 0, 10000000000))
}

// GetRandomTrigger returns a random event
func GetRandomEvent(r *rand.Rand, now time.Time) types.TriggerEventI {
	rand := simtypes.RandIntBetween(r, 0, 1)
	if rand > 0 {
		height := uint64(simtypes.RandIntBetween(r, 100000, 200000))
		return &types.BlockHeightEvent{BlockHeight: height}
	}
	randTime := now.Add(time.Hour * time.Duration(simtypes.RandIntBetween(r, 1, 10)))
	return &types.BlockTimeEvent{Time: randTime}
}

// GetRandomAction returns a random action
func GetRandomAction(r *rand.Rand, from string, to string) sdk.Msg {
	amount := int64(simtypes.RandIntBetween(r, 100, 1000))
	return &banktypes.MsgSend{
		FromAddress: from,
		ToAddress:   to,
		Amount:      sdk.NewCoins(sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().BondDenom, amount)),
	}
}

// GetRandomTrigger returns a random trigger
func GetRandomTrigger(r *rand.Rand, simState *module.SimulationState) types.Trigger {
	account := int64(simtypes.RandIntBetween(r, 1, 2))
	event := GetRandomEvent(r, simState.GenTimestamp.UTC())
	action := GetRandomAction(r, simState.Accounts[0].Address.String(), simState.Accounts[account].Address.String())

	actions, _ := sdktx.SetMsgs([]sdk.Msg{action})
	anyEvent, _ := codectypes.NewAnyWithValue(event)
	return types.NewTrigger(1, simState.Accounts[0].Address.String(), anyEvent, actions)
}

// GetRandomQueuedTrigger returns a random queued trigger
func GetRandomQueuedTrigger(_ *rand.Rand, simState *module.SimulationState) types.QueuedTrigger {
	var event types.TriggerEventI = &types.BlockHeightEvent{BlockHeight: 100000}
	var action sdk.Msg = &banktypes.MsgSend{
		FromAddress: simState.Accounts[0].Address.String(),
		ToAddress:   simState.Accounts[1].Address.String(),
		Amount:      sdk.NewCoins(sdk.NewInt64Coin(pioconfig.GetProvenanceConfig().BondDenom, 100)),
	}

	actions, _ := sdktx.SetMsgs([]sdk.Msg{action})
	anyEvent, _ := codectypes.NewAnyWithValue(event)
	trigger := types.NewTrigger(2, simState.Accounts[0].Address.String(), anyEvent, actions)

	now := simState.GenTimestamp

	return types.NewQueuedTrigger(trigger, now.UTC(), 1)
}

// RandomizedGenState generates a random GenesisState for trigger
func RandomizedGenState(simState *module.SimulationState) {
	var triggerID uint64
	simState.AppParams.GetOrGenerate(
		simState.Cdc, TriggerID, &triggerID, simState.Rand,
		func(r *rand.Rand) { triggerID = TriggerIDFn(r) },
	)

	var queueStart uint64
	simState.AppParams.GetOrGenerate(
		simState.Cdc, QueueStart, &queueStart, simState.Rand,
		func(r *rand.Rand) { queueStart = QueueStartFn(r) },
	)

	triggers := []types.Trigger{
		GetRandomTrigger(simState.Rand, simState),
	}
	queuedTriggers := []types.QueuedTrigger{
		GetRandomQueuedTrigger(simState.Rand, simState),
	}

	gasLimits := []types.GasLimit{}
	for _, trigger := range triggers {
		gasLimits = append(gasLimits, types.GasLimit{
			TriggerId: trigger.GetId(),
			Amount:    1000000,
		})
	}
	for _, item := range queuedTriggers {
		gasLimits = append(gasLimits, types.GasLimit{
			TriggerId: item.Trigger.Id,
			Amount:    1000000,
		})
	}

	// This is for params
	/*bz, err := json.MarshalIndent(&triggers, "", " ")
	if err != nil {
		panic(err)
	}*/

	genesis := types.NewGenesisState(triggerID, queueStart, triggers, gasLimits, queuedTriggers)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(genesis)
}
