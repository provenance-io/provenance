package keeper_test

import (
	"fmt"

	storetypes "cosmossdk.io/store/types"

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

	executed := func(triggerID uint64, owner string, success bool) sdk.Event {
		event, _ := sdk.TypedEventToEvent(&types.EventTriggerExecuted{
			TriggerId: fmt.Sprintf("%d", triggerID),
			Owner:     owner,
			Success:   success,
		})
		return event
	}

	tests := []struct {
		name     string
		panic    string
		existing []types.Trigger
		queue    []types.QueuedTrigger
		expected []types.Trigger
		events   sdk.Events
	}{
		{
			name:     "valid - no items in queue to run",
			existing: []types.Trigger{},
			queue:    []types.QueuedTrigger{},
			expected: []types.Trigger(nil),
			events:   []sdk.Event{},
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
			expected: []types.Trigger(nil),
			events:   []sdk.Event{destroyed(existing1.Id), executed(trigger1.Id, trigger1.Owner, true)},
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
			expected: []types.Trigger(nil),
			events:   []sdk.Event{executed(emptyTrigger.Id, emptyTrigger.Owner, true)},
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
			expected: []types.Trigger(nil),
			events:   []sdk.Event{destroyed(existing1.Id), destroyed(existing2.Id), executed(multiActionTrigger.Id, multiActionTrigger.Owner, true)},
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
			expected: []types.Trigger(nil),
			events:   []sdk.Event{destroyed(existing1.Id), executed(trigger1.Id, trigger1.Owner, true), destroyed(existing2.Id), executed(trigger2.Id, trigger2.Owner, true)},
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
			expected: []types.Trigger{existing2},
			events: []sdk.Event{
				destroyed(existing1.Id),
				executed(trigger1.Id, trigger1.Owner, true),
				executed(trigger3.Id, trigger3.Owner, false),
				executed(trigger4.Id, trigger4.Owner, false),
				executed(trigger5.Id, trigger5.Owner, false),
				executed(trigger6.Id, trigger6.Owner, false),
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			for _, trigger := range tc.existing {
				s.app.TriggerKeeper.RegisterTrigger(s.ctx, trigger)
				s.ctx.GasMeter().RefundGas(s.ctx.GasMeter().GasConsumed(), "testing")
			}

			for _, item := range tc.queue {
				s.app.TriggerKeeper.Enqueue(s.ctx, item)
			}
			s.ctx = s.ctx.WithEventManager(sdk.NewEventManager())
			s.ctx = s.ctx.WithBlockGasMeter(storetypes.NewGasMeter(60000000))

			if len(tc.panic) > 0 {
				s.PanicsWithValue(tc.panic, func() {
					s.app.TriggerKeeper.ProcessTriggers(s.ctx)
				})
			} else {
				s.app.TriggerKeeper.ProcessTriggers(s.ctx)

				remaining, err := s.app.TriggerKeeper.GetAllTriggers(s.ctx)
				s.NoError(err, "GetAllTriggers")
				s.Equal(tc.expected, remaining, "should still have remaining triggers after deletion actions from ProcessTriggers")

				for _, trigger := range remaining {
					s.app.TriggerKeeper.UnregisterTrigger(s.ctx, trigger)
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
