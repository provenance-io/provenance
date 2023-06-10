package types

import (
	"testing"
	time "time"

	"github.com/stretchr/testify/assert"

	"github.com/cosmos/cosmos-sdk/types"
)

func TestNewGenesisState(t *testing.T) {
	request := NewCreateTriggerRequest("addr", &BlockHeightEvent{}, []types.Msg{&MsgDestroyTriggerRequest{}})
	trigger := NewTrigger(1, "owner", request.Event, request.Actions)
	state := NewGenesisState(1, 2, []Trigger{trigger}, []GasLimit{{TriggerId: 1, Amount: 1}, {TriggerId: 2, Amount: 2}}, []QueuedTrigger{{BlockHeight: 1, Time: time.Time{}, Trigger: trigger}})

	assert.Equal(t, uint64(1), state.TriggerId, "trigger ids should match in NewGenesisState")
	assert.Equal(t, uint64(2), state.QueueStart, "queue start should match in NewGenesisState")
	assert.Equal(t, []Trigger{trigger}, state.Triggers, "triggers should match in NewGenesisState")
	assert.Equal(t, []GasLimit{{TriggerId: 1, Amount: 1}, {TriggerId: 2, Amount: 2}}, state.GasLimits, "gas limits should match in NewGenesisState")
	assert.Equal(t, []QueuedTrigger{{BlockHeight: 1, Time: time.Time{}, Trigger: trigger}}, state.QueuedTriggers, "queud triggers should match in NewGenesisState")
}

func TestDefaultGenesis(t *testing.T) {
	state := DefaultGenesis()

	assert.Equal(t, uint64(1), state.TriggerId, "trigger ids should match in DefaultGenesis")
	assert.Equal(t, uint64(1), state.QueueStart, "queue start should match in DefaultGenesis")
	assert.Equal(t, []Trigger{}, state.Triggers, "triggers should be empty in DefaultGenesis")
	assert.Equal(t, []GasLimit{}, state.GasLimits, "gas limits should be empty in default DefaultGenesis")
	assert.Equal(t, []QueuedTrigger{}, state.QueuedTriggers, "queued triggers should be empty in default DefaultGenesis")
}

func TestGenesisStateValidate(t *testing.T) {
	request := NewCreateTriggerRequest("addr", &BlockHeightEvent{}, []types.Msg{&MsgDestroyTriggerRequest{Id: 1, Authority: "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma"}})
	badRequest := NewCreateTriggerRequest("addr", &TransactionEvent{Name: "", Attributes: []Attribute{}}, []types.Msg{&MsgDestroyTriggerRequest{Id: 1, Authority: ""}})
	trigger := NewTrigger(1, "owner", request.Event, request.Actions)
	trigger2 := NewTrigger(2, "owner", request.Event, request.Actions)

	tests := []struct {
		name   string
		state  *GenesisState
		modify func(*GenesisState)
		err    string
	}{
		{
			name:   "valid - genesis state validate",
			state:  DefaultGenesis(),
			modify: nil,
			err:    "",
		},
		{
			name: "valid - nil slices",
			state: &GenesisState{
				TriggerId:      1,
				QueueStart:     1,
				GasLimits:      nil,
				Triggers:       nil,
				QueuedTriggers: nil,
			},
			modify: nil,
			err:    "",
		},
		{
			name: "invalid - trigger id cannot be zero",
			state: &GenesisState{
				TriggerId:      0,
				QueueStart:     1,
				GasLimits:      []GasLimit{},
				Triggers:       []Trigger{},
				QueuedTriggers: []QueuedTrigger{},
			},
			modify: nil,
			err:    "invalid trigger id",
		},
		{
			name: "invalid - queue start cannot be zero",
			state: &GenesisState{
				TriggerId:      1,
				QueueStart:     0,
				GasLimits:      []GasLimit{},
				Triggers:       []Trigger{},
				QueuedTriggers: []QueuedTrigger{},
			},
			modify: nil,
			err:    "invalid queue start",
		},
		{
			name: "invalid - gas and trigger length mismatch",
			state: &GenesisState{
				TriggerId:      1,
				QueueStart:     1,
				GasLimits:      []GasLimit{{TriggerId: 1, Amount: 1}},
				Triggers:       []Trigger{},
				QueuedTriggers: []QueuedTrigger{},
			},
			modify: nil,
			err:    "gas limit list length must match sum of triggers and queued triggers length",
		},
		{
			name: "invalid - gas and queue length mismatch",
			state: &GenesisState{
				TriggerId:      1,
				QueueStart:     1,
				GasLimits:      []GasLimit{{TriggerId: 1, Amount: 1}},
				Triggers:       []Trigger{},
				QueuedTriggers: []QueuedTrigger{},
			},
			modify: nil,
			err:    "gas limit list length must match sum of triggers and queued triggers length",
		},
		{
			name: "invalid - gas and trigger + queue length mismatch",
			state: &GenesisState{
				TriggerId:      2,
				QueueStart:     1,
				GasLimits:      []GasLimit{{TriggerId: 1, Amount: 1}},
				Triggers:       []Trigger{trigger},
				QueuedTriggers: []QueuedTrigger{{BlockHeight: 1, Time: time.Time{}, Trigger: trigger2}},
			},
			modify: nil,
			err:    "gas limit list length must match sum of triggers and queued triggers length",
		},
		{
			name: "invalid - action must pass internal validate basic",
			state: &GenesisState{
				TriggerId:      2,
				QueueStart:     1,
				GasLimits:      []GasLimit{{TriggerId: 1, Amount: 1}, {TriggerId: 2, Amount: 1}},
				Triggers:       []Trigger{trigger},
				QueuedTriggers: []QueuedTrigger{{BlockHeight: 1, Time: time.Time{}, Trigger: trigger2}},
			},
			modify: func(gs *GenesisState) {
				gs.Triggers[0].Actions = badRequest.Actions
			},
			err: "trigger id: 1, msg: 0, err: invalid address for trigger authority from address: empty address string is not allowed",
		},
		{
			name: "invalid - A trigger's id cannot exceed state trigger id",
			state: &GenesisState{
				TriggerId:      2,
				QueueStart:     1,
				GasLimits:      []GasLimit{{TriggerId: 1, Amount: 1}, {TriggerId: 3, Amount: 1}},
				Triggers:       []Trigger{trigger},
				QueuedTriggers: []QueuedTrigger{{BlockHeight: 1, Time: time.Time{}, Trigger: trigger2}},
			},
			modify: func(gs *GenesisState) {
				gs.Triggers[0].Owner = "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma"
				gs.QueuedTriggers[0].Trigger.Owner = "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma"
				gs.Triggers[0].Id = 5
			},
			err: "trigger id 5 is invalid and cannot exceed 2",
		},
		{
			name: "invalid - queued trigger action must pass internal validate basic",
			state: &GenesisState{
				TriggerId:      2,
				QueueStart:     1,
				GasLimits:      []GasLimit{{TriggerId: 1, Amount: 1}, {TriggerId: 2, Amount: 1}},
				Triggers:       []Trigger{trigger},
				QueuedTriggers: []QueuedTrigger{{BlockHeight: 1, Time: time.Time{}, Trigger: trigger2}},
			},
			modify: func(gs *GenesisState) {
				gs.QueuedTriggers[0].Trigger.Actions = badRequest.Actions
			},
			err: "trigger id: 2, msg: 0, err: invalid address for trigger authority from address: empty address string is not allowed",
		},
		{
			name: "invalid - A queued trigger's id cannot exceed state trigger id",
			state: &GenesisState{
				TriggerId:      2,
				QueueStart:     1,
				GasLimits:      []GasLimit{{TriggerId: 1, Amount: 1}, {TriggerId: 2, Amount: 1}},
				Triggers:       []Trigger{trigger},
				QueuedTriggers: []QueuedTrigger{{BlockHeight: 1, Time: time.Time{}, Trigger: trigger2}},
			},
			modify: func(gs *GenesisState) {
				gs.Triggers[0].Owner = "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma"
				gs.QueuedTriggers[0].Trigger.Owner = "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma"
				gs.QueuedTriggers[0].Trigger.Id = 5
			},
			err: "trigger id 5 is invalid and cannot exceed 2",
		},
		{
			name: "invalid - A trigger's event must pass validation",
			state: &GenesisState{
				TriggerId:      1,
				QueueStart:     1,
				GasLimits:      []GasLimit{{TriggerId: 1, Amount: 1}, {TriggerId: 2, Amount: 1}},
				Triggers:       []Trigger{trigger},
				QueuedTriggers: []QueuedTrigger{{BlockHeight: 1, Time: time.Time{}, Trigger: trigger2}},
			},
			modify: func(gs *GenesisState) {
				gs.Triggers[0].Owner = "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma"
				gs.Triggers[0].Event = badRequest.Event
			},
			err: "could not validate event for trigger with id 1: empty event name",
		},
		{
			name: "invalid - A queued trigger's event must pass validation",
			state: &GenesisState{
				TriggerId:      2,
				QueueStart:     1,
				GasLimits:      []GasLimit{{TriggerId: 1, Amount: 1}, {TriggerId: 2, Amount: 1}},
				Triggers:       []Trigger{trigger},
				QueuedTriggers: []QueuedTrigger{{BlockHeight: 1, Time: time.Time{}, Trigger: trigger2}},
			},
			modify: func(gs *GenesisState) {
				gs.Triggers[0].Owner = "cosmos1w6t0l7z0yerj49ehnqwqaayxqpe3u7e23edgma"
				gs.QueuedTriggers[0].Trigger.Event = badRequest.Event
			},
			err: "could not validate event for trigger with id 2: empty event name",
		},
		{
			name: "invalid - Gas limits must match either a trigger or queued trigger",
			state: &GenesisState{
				TriggerId:      3,
				QueueStart:     1,
				GasLimits:      []GasLimit{{TriggerId: 3, Amount: 1}, {TriggerId: 2, Amount: 1}},
				Triggers:       []Trigger{trigger},
				QueuedTriggers: []QueuedTrigger{{BlockHeight: 1, Time: time.Time{}, Trigger: trigger2}},
			},
			modify: nil,
			err:    "trigger or queued trigger does not have a gas limit that matches it with id 1",
		},
		{
			name: "invalid - Gas limits ids must be unique",
			state: &GenesisState{
				TriggerId:      1,
				QueueStart:     1,
				GasLimits:      []GasLimit{{TriggerId: 1, Amount: 1}, {TriggerId: 1, Amount: 1}},
				Triggers:       []Trigger{trigger},
				QueuedTriggers: []QueuedTrigger{{BlockHeight: 1, Time: time.Time{}, Trigger: trigger}},
			},
			modify: nil,
			err:    "cannot have duplicate trigger id (1) in gas limits",
		},
		{
			name: "invalid - triggers cannot have duplicate id",
			state: &GenesisState{
				TriggerId:      1,
				QueueStart:     1,
				GasLimits:      []GasLimit{{TriggerId: 1, Amount: 1}, {TriggerId: 2, Amount: 1}, {TriggerId: 3, Amount: 1}},
				Triggers:       []Trigger{trigger, trigger},
				QueuedTriggers: []QueuedTrigger{{BlockHeight: 1, Time: time.Time{}, Trigger: trigger2}},
			},
			modify: nil,
			err:    "trigger id 1 is not unique within the set all triggers and queued triggers",
		},
		{
			name: "invalid - queued triggers cannot have duplicate id",
			state: &GenesisState{
				TriggerId:      2,
				QueueStart:     1,
				GasLimits:      []GasLimit{{TriggerId: 1, Amount: 1}, {TriggerId: 2, Amount: 1}, {TriggerId: 3, Amount: 1}},
				Triggers:       []Trigger{trigger},
				QueuedTriggers: []QueuedTrigger{{BlockHeight: 1, Time: time.Time{}, Trigger: trigger2}, {BlockHeight: 1, Time: time.Time{}, Trigger: trigger2}},
			},
			modify: nil,
			err:    "trigger id 2 is not unique within the set all triggers and queued triggers",
		},
		{
			name: "invalid - cannot have duplicate id between triggers and queued triggers",
			state: &GenesisState{
				TriggerId:      1,
				QueueStart:     1,
				GasLimits:      []GasLimit{{TriggerId: 1, Amount: 1}, {TriggerId: 2, Amount: 1}},
				Triggers:       []Trigger{trigger},
				QueuedTriggers: []QueuedTrigger{{BlockHeight: 1, Time: time.Time{}, Trigger: trigger}},
			},
			modify: nil,
			err:    "trigger id 1 is not unique within the set all triggers and queued triggers",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.modify != nil {
				tc.modify(tc.state)
			}
			res := tc.state.Validate()
			if len(tc.err) > 0 {
				assert.EqualError(t, res, tc.err, "Validate")
			} else {
				assert.NoError(t, res, "Validate")
			}
		})
	}
}
