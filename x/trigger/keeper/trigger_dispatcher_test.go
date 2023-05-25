package keeper_test

import (
	"fmt"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/trigger/types"
)

func (s *KeeperTestSuite) TestProcessActions() {
	owner := s.accountAddresses[0].String()
	trigger1 := s.CreateTrigger(1, owner, &types.BlockHeightEvent{BlockHeight: uint64(s.ctx.BlockHeight())}, &types.MsgDestroyTriggerRequest{Id: 100, Authority: owner})
	trigger2 := s.CreateTrigger(2, owner, &types.BlockHeightEvent{BlockHeight: 130}, &types.MsgDestroyTriggerRequest{Id: 101, Authority: owner})
	emptyTrigger := s.CreateTrigger(3, owner, &types.BlockHeightEvent{BlockHeight: uint64(s.ctx.BlockHeight())}, &types.MsgDestroyTriggerRequest{Id: 102, Authority: owner})
	emptyTrigger.Actions = []*codectypes.Any{}
	multiActionTrigger := s.CreateTrigger(4, owner, &types.BlockHeightEvent{BlockHeight: uint64(s.ctx.BlockHeight())}, &types.MsgDestroyTriggerRequest{Id: 103, Authority: owner})
	multiActionTrigger.Actions = []*codectypes.Any{trigger1.Actions[0], trigger2.Actions[0]}

	existing1 := s.CreateTrigger(100, owner, &types.BlockHeightEvent{BlockHeight: 120}, &types.MsgDestroyTriggerRequest{Id: 100, Authority: owner})
	existing2 := s.CreateTrigger(101, owner, &types.BlockHeightEvent{BlockHeight: 120}, &types.MsgDestroyTriggerRequest{Id: 101, Authority: owner})

	event1, _ := sdk.TypedEventToEvent(&types.EventTriggerDestroyed{
		TriggerId: fmt.Sprintf("%d", existing1.GetId()),
	})
	event2, _ := sdk.TypedEventToEvent(&types.EventTriggerDestroyed{
		TriggerId: fmt.Sprintf("%d", existing2.GetId()),
	})

	tests := []struct {
		name     string
		panics   bool
		existing []types.Trigger
		queue    []types.QueuedTrigger
		gas      []uint64
		expected []types.Trigger
		events   sdk.Events
	}{
		{
			name:     "valid - no items in queue to run",
			panics:   false,
			existing: []types.Trigger{},
			queue:    []types.QueuedTrigger{},
			gas:      []uint64{},
			expected: []types.Trigger(nil),
			events:   []sdk.Event{},
		},
		{
			name:     "valid - one item in queue to run",
			panics:   false,
			existing: []types.Trigger{existing1},
			queue: []types.QueuedTrigger{
				{
					BlockHeight: uint64(s.ctx.BlockHeight()),
					Time:        s.ctx.BlockTime(),
					Trigger:     trigger1,
				},
			},
			gas:      []uint64{9999999999},
			expected: []types.Trigger(nil),
			events:   []sdk.Event{event1},
		},
		{
			name:   "invalid - trigger with missing gas limit",
			panics: true,
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
		},
		{
			name:     "valid - trigger with no action",
			panics:   false,
			existing: []types.Trigger{},
			queue: []types.QueuedTrigger{
				{
					BlockHeight: uint64(s.ctx.BlockHeight()),
					Time:        s.ctx.BlockTime(),
					Trigger:     emptyTrigger,
				},
			},
			gas:      []uint64{9999999999},
			expected: []types.Trigger(nil),
			events:   []sdk.Event{},
		},
		{
			name:     "valid - trigger with multiple actions",
			panics:   false,
			existing: []types.Trigger{existing1, existing2},
			queue: []types.QueuedTrigger{
				{
					BlockHeight: uint64(s.ctx.BlockHeight()),
					Time:        s.ctx.BlockTime(),
					Trigger:     multiActionTrigger,
				},
			},
			gas:      []uint64{9999999999},
			expected: []types.Trigger(nil),
			events:   []sdk.Event{event1, event2},
		},
		{
			name:     "valid - multiple triggers in queue",
			panics:   false,
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
			gas:      []uint64{9999999999, 9999999999},
			expected: []types.Trigger(nil),
			events:   []sdk.Event{event1, event2},
		},
		{
			name:     "invalid - trigger with single action runs out of gas",
			panics:   false,
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
			events:   []sdk.Event{},
		},
		{
			name:     "invalid - trigger with multiple actions runs out of gas",
			panics:   false,
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
			events:   []sdk.Event{},
		},
		{
			name:     "valid - multiple triggers in queue and one runs out of gas",
			panics:   false,
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
			gas:      []uint64{1, 9999999999},
			expected: []types.Trigger{existing1},
			events:   []sdk.Event{event2},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			for _, trigger := range tc.existing {
				s.app.TriggerKeeper.RegisterTrigger(s.ctx, trigger)
				s.ctx.GasMeter().RefundGas(999999999999, "testing")
			}

			for i, item := range tc.queue {
				s.app.TriggerKeeper.Enqueue(s.ctx, item)
				if len(tc.gas) > 0 {
					s.app.TriggerKeeper.SetGasLimit(s.ctx, item.Trigger.Id, tc.gas[i])
				}
			}
			s.ctx = s.ctx.WithEventManager(sdk.NewEventManager())

			if tc.panics {
				s.Panics(func() {
					s.app.TriggerKeeper.ProcessTriggers(s.ctx)
				})
			} else {
				s.app.TriggerKeeper.ProcessTriggers(s.ctx)

				remaining, err := s.app.TriggerKeeper.GetAllTriggers(s.ctx)
				s.NoError(err)
				s.Equal(tc.expected, remaining)

				events := s.ctx.EventManager().Events()
				s.Equal(tc.events, events)
			}
		})
	}
}
