package keeper_test

import (
	"fmt"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/trigger/types"
)

func (s *KeeperTestSuite) TestProcessTriggers() {
	owner := s.accountAddresses[0].String()
	trigger1 := s.CreateTrigger(1, owner, &types.BlockHeightEvent{BlockHeight: uint64(s.ctx.BlockHeight())}, &types.MsgDestroyTriggerRequest{Id: 100, Authority: owner})
	trigger2 := s.CreateTrigger(2, owner, &types.BlockHeightEvent{BlockHeight: 130}, &types.MsgDestroyTriggerRequest{Id: 101, Authority: owner})
	trigger3 := s.CreateTrigger(4, owner, &types.BlockHeightEvent{BlockHeight: 130}, &types.MsgDestroyTriggerRequest{Id: 104, Authority: owner})
	trigger4 := s.CreateTrigger(5, owner, &types.BlockHeightEvent{BlockHeight: 130}, &types.MsgDestroyTriggerRequest{Id: 105, Authority: owner})
	trigger5 := s.CreateTrigger(6, owner, &types.BlockHeightEvent{BlockHeight: 130}, &types.MsgDestroyTriggerRequest{Id: 106, Authority: owner})
	trigger6 := s.CreateTrigger(7, owner, &types.BlockHeightEvent{BlockHeight: 130}, &types.MsgDestroyTriggerRequest{Id: 107, Authority: owner})
	emptyTrigger := s.CreateTrigger(3, owner, &types.BlockHeightEvent{BlockHeight: uint64(s.ctx.BlockHeight())}, &types.MsgDestroyTriggerRequest{Id: 102, Authority: owner})
	emptyTrigger.Actions = []*codectypes.Any{}
	multiActionTrigger := s.CreateTrigger(4, owner, &types.BlockHeightEvent{BlockHeight: uint64(s.ctx.BlockHeight())}, &types.MsgDestroyTriggerRequest{Id: 103, Authority: owner})
	multiActionTrigger.Actions = []*codectypes.Any{trigger1.Actions[0], trigger2.Actions[0]}

	existing1 := s.CreateTrigger(100, owner, &types.BlockHeightEvent{BlockHeight: 120}, &types.MsgDestroyTriggerRequest{Id: 100, Authority: owner})
	existing2 := s.CreateTrigger(101, owner, &types.BlockHeightEvent{BlockHeight: 120}, &types.MsgDestroyTriggerRequest{Id: 101, Authority: owner})

	destroyed := func(triggerID uint64) sdk.Event {
		event, _ := sdk.TypedEventToEvent(&types.EventTriggerDestroyed{
			TriggerId: fmt.Sprintf("%d", triggerID),
		})
		return event
	}

	executed := func(triggerID uint64, owner, err string) sdk.Event {
		event, _ := sdk.TypedEventToEvent(&types.EventTriggerExecuted{
			TriggerId: fmt.Sprintf("%d", triggerID),
			Owner:     owner,
			Error:     err,
		})
		return event
	}

	tests := []struct {
		name     string
		panic    string
		existing []types.Trigger
		queue    []types.QueuedTrigger
		gas      []uint64
		expected []types.Trigger
		events   sdk.Events
		blockGas uint64
	}{
		{
			name:     "valid - no items in queue to run",
			existing: []types.Trigger{},
			queue:    []types.QueuedTrigger{},
			gas:      []uint64{},
			expected: []types.Trigger(nil),
			events:   []sdk.Event{},
			blockGas: 0,
		},
		{
			name:     "valid - one item in queue to run",
			existing: []types.Trigger{existing1},
			queue: []types.QueuedTrigger{
				{
					BlockHeight: uint64(s.ctx.BlockHeight()),
					Time:        s.ctx.BlockTime(),
					Trigger:     trigger1,
				},
			},
			gas:      []uint64{2000000},
			expected: []types.Trigger(nil),
			events:   []sdk.Event{destroyed(existing1.Id), executed(trigger1.Id, trigger1.Owner, "")},
			blockGas: 2000000,
		},
		{
			name:  "invalid - trigger with missing gas limit",
			panic: "gas limit not found for trigger",
			queue: []types.QueuedTrigger{
				{
					BlockHeight: uint64(s.ctx.BlockHeight()),
					Time:        s.ctx.BlockTime(),
					Trigger:     trigger1,
				},
			},
			gas:      []uint64{},
			expected: []types.Trigger{existing1},
			events:   []sdk.Event{},
			blockGas: 0,
		},
		{
			name:     "valid - trigger with no action",
			existing: []types.Trigger{},
			queue: []types.QueuedTrigger{
				{
					BlockHeight: uint64(s.ctx.BlockHeight()),
					Time:        s.ctx.BlockTime(),
					Trigger:     emptyTrigger,
				},
			},
			gas:      []uint64{2000000},
			expected: []types.Trigger(nil),
			events:   []sdk.Event{executed(emptyTrigger.Id, emptyTrigger.Owner, "")},
			blockGas: 2000000,
		},
		{
			name:     "valid - trigger with multiple actions",
			existing: []types.Trigger{existing1, existing2},
			queue: []types.QueuedTrigger{
				{
					BlockHeight: uint64(s.ctx.BlockHeight()),
					Time:        s.ctx.BlockTime(),
					Trigger:     multiActionTrigger,
				},
			},
			gas:      []uint64{2000000},
			expected: []types.Trigger(nil),
			events:   []sdk.Event{destroyed(existing1.Id), destroyed(existing2.Id), executed(multiActionTrigger.Id, multiActionTrigger.Owner, "")},
			blockGas: 2000000,
		},
		{
			name:     "valid - multiple triggers in queue",
			existing: []types.Trigger{existing1, existing2},
			queue: []types.QueuedTrigger{
				{
					BlockHeight: uint64(s.ctx.BlockHeight()),
					Time:        s.ctx.BlockTime(),
					Trigger:     trigger1,
				},
				{
					BlockHeight: uint64(s.ctx.BlockHeight()),
					Time:        s.ctx.BlockTime(),
					Trigger:     trigger2,
				},
			},
			gas:      []uint64{1000000, 1000000},
			expected: []types.Trigger(nil),
			events:   []sdk.Event{destroyed(existing1.Id), executed(trigger1.Id, trigger1.Owner, ""), destroyed(existing2.Id), executed(trigger2.Id, trigger2.Owner, "")},
			blockGas: 2000000,
		},
		{
			name:     "valid - limit multiple triggers in queue by gas",
			existing: []types.Trigger{existing1, existing2},
			queue: []types.QueuedTrigger{
				{
					BlockHeight: uint64(s.ctx.BlockHeight()),
					Time:        s.ctx.BlockTime(),
					Trigger:     trigger1,
				},
				{
					BlockHeight: uint64(s.ctx.BlockHeight()),
					Time:        s.ctx.BlockTime(),
					Trigger:     trigger2,
				},
			},
			gas:      []uint64{2000000, 1000000},
			expected: []types.Trigger{existing2},
			events:   []sdk.Event{destroyed(existing1.Id), executed(trigger1.Id, trigger1.Owner, "")},
			blockGas: 2000000,
		},
		{
			name:     "valid - limit multiple triggers in queue",
			existing: []types.Trigger{existing1, existing2},
			queue: []types.QueuedTrigger{
				{
					BlockHeight: uint64(s.ctx.BlockHeight()),
					Time:        s.ctx.BlockTime(),
					Trigger:     trigger1,
				},
				{
					BlockHeight: uint64(s.ctx.BlockHeight()),
					Time:        s.ctx.BlockTime(),
					Trigger:     trigger3,
				},
				{
					BlockHeight: uint64(s.ctx.BlockHeight()),
					Time:        s.ctx.BlockTime(),
					Trigger:     trigger4,
				},
				{
					BlockHeight: uint64(s.ctx.BlockHeight()),
					Time:        s.ctx.BlockTime(),
					Trigger:     trigger5,
				},
				{
					BlockHeight: uint64(s.ctx.BlockHeight()),
					Time:        s.ctx.BlockTime(),
					Trigger:     trigger6,
				},
				{
					BlockHeight: uint64(s.ctx.BlockHeight()),
					Time:        s.ctx.BlockTime(),
					Trigger:     trigger2,
				},
			},
			gas:      []uint64{100000, 100000, 100000, 100000, 100000, 100000},
			expected: []types.Trigger{existing2},
			events: []sdk.Event{
				destroyed(existing1.Id),
				executed(trigger1.Id, trigger1.Owner, ""),
				executed(trigger3.Id, trigger3.Owner, "error processing message /provenance.trigger.v1.MsgDestroyTriggerRequest at position 0: trigger not found"),
				executed(trigger4.Id, trigger4.Owner, "error processing message /provenance.trigger.v1.MsgDestroyTriggerRequest at position 0: trigger not found"),
				executed(trigger5.Id, trigger5.Owner, "error processing message /provenance.trigger.v1.MsgDestroyTriggerRequest at position 0: trigger not found"),
				executed(trigger6.Id, trigger6.Owner, "error processing message /provenance.trigger.v1.MsgDestroyTriggerRequest at position 0: trigger not found"),
			},
			blockGas: 500000,
		},
		{
			name:     "invalid - trigger with single action runs out of gas",
			existing: []types.Trigger{existing1},
			queue: []types.QueuedTrigger{
				{
					BlockHeight: uint64(s.ctx.BlockHeight()),
					Time:        s.ctx.BlockTime(),
					Trigger:     trigger1,
				},
			},
			gas:      []uint64{1},
			expected: []types.Trigger{existing1},
			events: []sdk.Event{
				executed(trigger1.Id, trigger1.Owner, "error processing message /provenance.trigger.v1.MsgDestroyTriggerRequest at position 0: gas 1000 exceeded limit 1 for message \"/provenance.trigger.v1.MsgDestroyTriggerRequest\"")},
			blockGas: 1,
		},
		{
			name:     "invalid - trigger with multiple actions runs out of gas",
			existing: []types.Trigger{existing1, existing2},
			queue: []types.QueuedTrigger{
				{
					BlockHeight: uint64(s.ctx.BlockHeight()),
					Time:        s.ctx.BlockTime(),
					Trigger:     multiActionTrigger,
				},
			},
			gas:      []uint64{6000},
			expected: []types.Trigger{existing1, existing2},
			events: []sdk.Event{
				executed(multiActionTrigger.Id, multiActionTrigger.Owner, "error processing message /provenance.trigger.v1.MsgDestroyTriggerRequest at position 0: gas 6621 exceeded limit 6000 for message \"/provenance.trigger.v1.MsgDestroyTriggerRequest\""),
			},
			blockGas: 6000,
		},
		{
			name:     "valid - multiple triggers in queue and one runs out of gas",
			existing: []types.Trigger{existing1, existing2},
			queue: []types.QueuedTrigger{
				{
					BlockHeight: uint64(s.ctx.BlockHeight()),
					Time:        s.ctx.BlockTime(),
					Trigger:     trigger1,
				},
				{
					BlockHeight: uint64(s.ctx.BlockHeight()),
					Time:        s.ctx.BlockTime(),
					Trigger:     trigger2,
				},
			},
			gas:      []uint64{1, 1000000},
			expected: []types.Trigger{existing1},
			events: []sdk.Event{
				executed(trigger1.Id, trigger1.Owner, "error processing message /provenance.trigger.v1.MsgDestroyTriggerRequest at position 0: gas 1000 exceeded limit 1 for message \"/provenance.trigger.v1.MsgDestroyTriggerRequest\""),
				destroyed(existing2.Id),
				executed(trigger2.Id, trigger2.Owner, ""),
			},
			blockGas: 1000001,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			for _, trigger := range tc.existing {
				s.app.TriggerKeeper.RegisterTrigger(s.ctx, trigger)
				s.ctx.GasMeter().RefundGas(s.ctx.GasMeter().GasConsumed(), "testing")
			}

			for i, item := range tc.queue {
				s.app.TriggerKeeper.Enqueue(s.ctx, item)
				if len(tc.gas) > 0 {
					s.app.TriggerKeeper.SetGasLimit(s.ctx, item.Trigger.Id, tc.gas[i])
				}
			}
			s.ctx = s.ctx.WithEventManager(sdk.NewEventManager())
			s.ctx = s.ctx.WithBlockGasMeter(sdk.NewGasMeter(60000000))

			if len(tc.panic) > 0 {
				s.PanicsWithValue(tc.panic, func() {
					s.app.TriggerKeeper.ProcessTriggers(s.ctx)
				})
			} else {
				s.app.TriggerKeeper.ProcessTriggers(s.ctx)

				s.Equal(int(tc.blockGas), int(s.ctx.BlockGasMeter().GasConsumed()), "should consume the correct amount of block gas")

				remaining, err := s.app.TriggerKeeper.GetAllTriggers(s.ctx)
				s.NoError(err, "GetAllTriggers")
				s.Equal(tc.expected, remaining, "should still have remaining triggers after deletion actions from ProcessTriggers")

				for _, trigger := range remaining {
					s.app.TriggerKeeper.UnregisterTrigger(s.ctx, trigger)
					s.app.TriggerKeeper.RemoveGasLimit(s.ctx, trigger.Id)
				}

				events := s.ctx.EventManager().Events()
				s.Equal(tc.events, events, "should have matching events from successful actions in ProcessTriggers")
			}

			for !s.app.TriggerKeeper.QueueIsEmpty(s.ctx) {
				s.app.TriggerKeeper.Dequeue(s.ctx)
			}
		})
	}
}
