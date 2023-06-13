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

	event1, _ := sdk.TypedEventToEvent(&types.EventTriggerDestroyed{
		TriggerId: fmt.Sprintf("%d", existing1.GetId()),
	})
	event2, _ := sdk.TypedEventToEvent(&types.EventTriggerDestroyed{
		TriggerId: fmt.Sprintf("%d", existing2.GetId()),
	})

	tests := []struct {
		name         string
		panic        string
		existing     []types.Trigger
		queue        []types.QueuedTrigger
		gas          []uint64
		expected     []types.Trigger
		events       sdk.Events
		remainingGas uint64
	}{
		{
			name:         "valid - no items in queue to run",
			existing:     []types.Trigger{},
			queue:        []types.QueuedTrigger{},
			gas:          []uint64{},
			expected:     []types.Trigger(nil),
			events:       []sdk.Event{},
			remainingGas: 3998973,
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
			gas:          []uint64{2000000},
			expected:     []types.Trigger(nil),
			events:       []sdk.Event{event1},
			remainingGas: 3981542,
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
			gas:          []uint64{},
			expected:     []types.Trigger{existing1},
			events:       []sdk.Event{},
			remainingGas: 3994214,
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
			gas:          []uint64{2000000},
			expected:     []types.Trigger(nil),
			events:       []sdk.Event{},
			remainingGas: 3981851,
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
			gas:          []uint64{2000000},
			expected:     []types.Trigger(nil),
			events:       []sdk.Event{event1, event2},
			remainingGas: 3981236,
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
			gas:          []uint64{1000000, 1000000},
			expected:     []types.Trigger(nil),
			events:       []sdk.Event{event1, event2},
			remainingGas: 3964108,
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
			gas:          []uint64{2000000, 1000000},
			expected:     []types.Trigger{existing2},
			events:       []sdk.Event{event1},
			remainingGas: 3976756,
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
			gas:          []uint64{100000, 100000, 100000, 100000, 100000, 100000},
			expected:     []types.Trigger{existing2},
			events:       []sdk.Event{event1},
			remainingGas: 3511806,
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
			gas:          []uint64{1},
			expected:     []types.Trigger{existing1},
			events:       []sdk.Event{},
			remainingGas: 3981541,
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
			gas:          []uint64{6000},
			expected:     []types.Trigger{existing1, existing2},
			events:       []sdk.Event{},
			remainingGas: 3975236,
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
			gas:          []uint64{1, 1000000},
			expected:     []types.Trigger{existing1},
			events:       []sdk.Event{event2},
			remainingGas: 3964107,
		},
		{
			name:     "invalid - consumes all gas if action fails",
			existing: []types.Trigger{existing2},
			queue: []types.QueuedTrigger{
				{
					BlockHeight: uint64(s.ctx.BlockHeight()),
					Time:        s.ctx.BlockTime(),
					Trigger:     trigger1,
				},
			},
			gas:          []uint64{2000000},
			expected:     []types.Trigger{existing2},
			events:       []sdk.Event{},
			remainingGas: 1981542,
		},
		{
			name:     "invalid - consumes all gas if one action fails",
			existing: []types.Trigger{existing1},
			queue: []types.QueuedTrigger{
				{
					BlockHeight: uint64(s.ctx.BlockHeight()),
					Time:        s.ctx.BlockTime(),
					Trigger:     multiActionTrigger,
				},
			},
			gas:          []uint64{2000000},
			expected:     []types.Trigger{existing1},
			events:       []sdk.Event{},
			remainingGas: 1981236,
		},
		{
			name:     "invalid - stops early if one action fails",
			existing: []types.Trigger{existing2},
			queue: []types.QueuedTrigger{
				{
					BlockHeight: uint64(s.ctx.BlockHeight()),
					Time:        s.ctx.BlockTime(),
					Trigger:     multiActionTrigger,
				},
			},
			gas:          []uint64{2000000},
			expected:     []types.Trigger{existing2},
			events:       []sdk.Event{},
			remainingGas: 1981236,
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
			s.ctx = s.ctx.WithGasMeter(sdk.NewGasMeter(4000000))

			if len(tc.panic) > 0 {
				s.PanicsWithValue(tc.panic, func() {
					s.app.TriggerKeeper.ProcessTriggers(s.ctx)
				})
			} else {
				s.app.TriggerKeeper.ProcessTriggers(s.ctx)
				s.Equal(tc.remainingGas, s.ctx.GasMeter().GasRemaining(), "should have correct gas usage")

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
