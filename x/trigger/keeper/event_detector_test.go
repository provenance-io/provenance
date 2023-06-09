package keeper_test

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/provenance-io/provenance/x/trigger/types"
)

func (s *KeeperTestSuite) TestDetectBlockEvents() {
	tests := []struct {
		name       string
		triggers   []types.TriggerEventI
		registered []types.Trigger
		queued     []types.QueuedTrigger
	}{
		{
			name:       "valid - no triggers",
			triggers:   []types.TriggerEventI{},
			registered: []types.Trigger(nil),
			queued:     []types.QueuedTrigger(nil),
		},
		{
			name: "valid - triggers, but no events",
			triggers: []types.TriggerEventI{
				&types.BlockHeightEvent{BlockHeight: uint64(s.ctx.BlockHeight()) + 1},
			},
			registered: []types.Trigger{
				s.CreateTrigger(1, s.accountAddresses[0].String(), &types.BlockHeightEvent{BlockHeight: uint64(s.ctx.BlockHeight()) + 1}, &types.MsgDestroyTriggerRequest{Id: 1, Authority: s.accountAddresses[0].String()}),
			},
			queued: []types.QueuedTrigger(nil),
		},
		{
			name: "valid - 1 detected transaction event",
			triggers: []types.TriggerEventI{
				&types.TransactionEvent{Name: "event2"},
			},
			registered: []types.Trigger(nil),
			queued: []types.QueuedTrigger{
				{
					BlockHeight: uint64(s.ctx.BlockHeight()),
					Time:        s.ctx.BlockTime(),
					Trigger:     s.CreateTrigger(2, s.accountAddresses[0].String(), &types.TransactionEvent{Name: "event2"}, &types.MsgDestroyTriggerRequest{Id: 1, Authority: s.accountAddresses[0].String()}),
				},
			},
		},
		{
			name: "valid - 1 detected transaction event for multiple of same event",
			triggers: []types.TriggerEventI{
				&types.TransactionEvent{Name: "event1"},
			},
			registered: []types.Trigger(nil),
			queued: []types.QueuedTrigger{
				{
					BlockHeight: uint64(s.ctx.BlockHeight()),
					Time:        s.ctx.BlockTime(),
					Trigger:     s.CreateTrigger(3, s.accountAddresses[0].String(), &types.TransactionEvent{Name: "event1"}, &types.MsgDestroyTriggerRequest{Id: 1, Authority: s.accountAddresses[0].String()}),
				},
			},
		},
		{
			name: "valid - 1 detected block height event",
			triggers: []types.TriggerEventI{
				&types.BlockHeightEvent{BlockHeight: uint64(s.ctx.BlockHeader().Height)},
			},
			registered: []types.Trigger(nil),
			queued: []types.QueuedTrigger{
				{
					BlockHeight: uint64(s.ctx.BlockHeight()),
					Time:        s.ctx.BlockTime(),
					Trigger:     s.CreateTrigger(4, s.accountAddresses[0].String(), &types.BlockHeightEvent{BlockHeight: uint64(s.ctx.BlockHeight())}, &types.MsgDestroyTriggerRequest{Id: 1, Authority: s.accountAddresses[0].String()}),
				},
			},
		},
		{
			name: "valid - 1 detected time event",
			triggers: []types.TriggerEventI{
				&types.BlockTimeEvent{Time: s.ctx.BlockTime()},
			},
			registered: []types.Trigger(nil),
			queued: []types.QueuedTrigger{
				{
					BlockHeight: uint64(s.ctx.BlockHeight()),
					Time:        s.ctx.BlockTime(),
					Trigger:     s.CreateTrigger(5, s.accountAddresses[0].String(), &types.BlockTimeEvent{Time: s.ctx.BlockTime()}, &types.MsgDestroyTriggerRequest{Id: 1, Authority: s.accountAddresses[0].String()}),
				},
			},
		},
		{
			name: "valid - multiple detected transaction events",
			triggers: []types.TriggerEventI{
				&types.TransactionEvent{Name: "event1"},
				&types.TransactionEvent{Name: "event2"},
				&types.TransactionEvent{Name: "event2"},
			},
			registered: []types.Trigger(nil),
			queued: []types.QueuedTrigger{
				{
					BlockHeight: uint64(s.ctx.BlockHeight()),
					Time:        s.ctx.BlockTime(),
					Trigger:     s.CreateTrigger(6, s.accountAddresses[0].String(), &types.TransactionEvent{Name: "event1"}, &types.MsgDestroyTriggerRequest{Id: 1, Authority: s.accountAddresses[0].String()}),
				},
				{
					BlockHeight: uint64(s.ctx.BlockHeight()),
					Time:        s.ctx.BlockTime(),
					Trigger:     s.CreateTrigger(7, s.accountAddresses[0].String(), &types.TransactionEvent{Name: "event2"}, &types.MsgDestroyTriggerRequest{Id: 1, Authority: s.accountAddresses[0].String()}),
				},
				{
					BlockHeight: uint64(s.ctx.BlockHeight()),
					Time:        s.ctx.BlockTime(),
					Trigger:     s.CreateTrigger(8, s.accountAddresses[0].String(), &types.TransactionEvent{Name: "event2"}, &types.MsgDestroyTriggerRequest{Id: 1, Authority: s.accountAddresses[0].String()}),
				},
			},
		},
		{
			name: "valid - multiple detected block height events",
			triggers: []types.TriggerEventI{
				&types.BlockHeightEvent{BlockHeight: uint64(s.ctx.BlockHeight())},
				&types.BlockHeightEvent{BlockHeight: uint64(s.ctx.BlockHeight())},
			},
			registered: []types.Trigger(nil),
			queued: []types.QueuedTrigger{
				{
					BlockHeight: uint64(s.ctx.BlockHeight()),
					Time:        s.ctx.BlockTime(),
					Trigger:     s.CreateTrigger(9, s.accountAddresses[0].String(), &types.BlockHeightEvent{BlockHeight: uint64(s.ctx.BlockHeight())}, &types.MsgDestroyTriggerRequest{Id: 1, Authority: s.accountAddresses[0].String()}),
				},
				{
					BlockHeight: uint64(s.ctx.BlockHeight()),
					Time:        s.ctx.BlockTime(),
					Trigger:     s.CreateTrigger(10, s.accountAddresses[0].String(), &types.BlockHeightEvent{BlockHeight: uint64(s.ctx.BlockHeight())}, &types.MsgDestroyTriggerRequest{Id: 1, Authority: s.accountAddresses[0].String()}),
				},
			},
		},
		{
			name: "valid - multiple detected time events",
			triggers: []types.TriggerEventI{
				&types.BlockTimeEvent{Time: s.ctx.BlockTime()},
				&types.BlockTimeEvent{Time: s.ctx.BlockTime()},
			},
			registered: []types.Trigger(nil),
			queued: []types.QueuedTrigger{
				{
					BlockHeight: uint64(s.ctx.BlockHeight()),
					Time:        s.ctx.BlockTime(),
					Trigger:     s.CreateTrigger(11, s.accountAddresses[0].String(), &types.BlockTimeEvent{Time: s.ctx.BlockTime()}, &types.MsgDestroyTriggerRequest{Id: 1, Authority: s.accountAddresses[0].String()}),
				},
				{
					BlockHeight: uint64(s.ctx.BlockHeight()),
					Time:        s.ctx.BlockTime(),
					Trigger:     s.CreateTrigger(12, s.accountAddresses[0].String(), &types.BlockTimeEvent{Time: s.ctx.BlockTime()}, &types.MsgDestroyTriggerRequest{Id: 1, Authority: s.accountAddresses[0].String()}),
				},
			},
		},
		{
			name: "valid - different types detected",
			triggers: []types.TriggerEventI{
				&types.TransactionEvent{Name: "event1"},
				&types.TransactionEvent{Name: "non-existing-event"},
				&types.BlockHeightEvent{BlockHeight: uint64(s.ctx.BlockHeight())},
				&types.BlockTimeEvent{Time: s.ctx.BlockTime()},
			},
			registered: []types.Trigger{
				s.CreateTrigger(14, s.accountAddresses[0].String(), &types.TransactionEvent{Name: "non-existing-event"}, &types.MsgDestroyTriggerRequest{Id: 1, Authority: s.accountAddresses[0].String()}),
			},
			queued: []types.QueuedTrigger{
				{
					BlockHeight: uint64(s.ctx.BlockHeight()),
					Time:        s.ctx.BlockTime(),
					Trigger:     s.CreateTrigger(13, s.accountAddresses[0].String(), &types.TransactionEvent{Name: "event1"}, &types.MsgDestroyTriggerRequest{Id: 1, Authority: s.accountAddresses[0].String()}),
				},
				{
					BlockHeight: uint64(s.ctx.BlockHeight()),
					Time:        s.ctx.BlockTime(),
					Trigger:     s.CreateTrigger(15, s.accountAddresses[0].String(), &types.BlockHeightEvent{BlockHeight: uint64(s.ctx.BlockHeight())}, &types.MsgDestroyTriggerRequest{Id: 1, Authority: s.accountAddresses[0].String()}),
				},
				{
					BlockHeight: uint64(s.ctx.BlockHeight()),
					Time:        s.ctx.BlockTime(),
					Trigger:     s.CreateTrigger(16, s.accountAddresses[0].String(), &types.BlockTimeEvent{Time: s.ctx.BlockTime()}, &types.MsgDestroyTriggerRequest{Id: 1, Authority: s.accountAddresses[0].String()}),
				},
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			// Setup
			registered := []types.Trigger{}
			for _, event := range tc.triggers {
				actions, _ := sdktx.SetMsgs([]sdk.Msg{&types.MsgDestroyTriggerRequest{Id: 1, Authority: s.accountAddresses[0].String()}})
				any, _ := codectypes.NewAnyWithValue(event)
				trigger := s.app.TriggerKeeper.NewTriggerWithID(s.ctx, s.accountAddresses[0].String(), any, actions)
				s.app.TriggerKeeper.RegisterTrigger(s.ctx, trigger)
				s.ctx.GasMeter().RefundGas(s.ctx.GasMeter().GasConsumed(), "testing")
				registered = append(registered, trigger)
			}

			// Action
			s.app.TriggerKeeper.DetectBlockEvents(s.ctx)

			// Verify
			triggers, err := s.app.TriggerKeeper.GetAllTriggers(s.ctx)
			s.NoError(err, "GetAllTriggers")
			s.Equal(tc.registered, triggers, "should have the correct remaining triggers after DetectBlockEvents")
			for _, trigger := range triggers {
				event, err := trigger.GetTriggerEventI()
				s.NoError(err, "GetTriggerEventI")
				_, err = s.app.TriggerKeeper.GetEventListener(s.ctx, event.GetEventPrefix(), trigger.GetId())
				s.NoError(err, "GetEventListener")
			}

			items, err := s.app.TriggerKeeper.GetAllQueueItems(s.ctx)
			s.NoError(err, "GetAllQueueItems")
			s.Equal(tc.queued, items, "should have correct items in queue after DetectBlockEvents")

			// Cleanup
			for !s.app.TriggerKeeper.QueueIsEmpty(s.ctx) {
				s.app.TriggerKeeper.Dequeue(s.ctx)
			}
			for _, registered := range registered {
				s.app.TriggerKeeper.UnregisterTrigger(s.ctx, registered)
			}

		})
	}
}
