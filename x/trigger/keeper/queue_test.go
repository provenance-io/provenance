package keeper_test

import "github.com/provenance-io/provenance/x/trigger/types"

func (s *KeeperTestSuite) TestQueueTrigger() {
	owner := s.accountAddresses[0].String()
	trigger1 := s.CreateTrigger(1, owner, &types.BlockHeightEvent{BlockHeight: 120}, &types.MsgDestroyTriggerRequest{Id: 1, Authority: owner})
	trigger2 := s.CreateTrigger(2, owner, &types.BlockHeightEvent{BlockHeight: 130}, &types.MsgDestroyTriggerRequest{Id: 1, Authority: owner})
	tests := []struct {
		name     string
		triggers []types.Trigger
		expected []types.QueuedTrigger
	}{
		{
			name:     "valid - Test queue single item",
			triggers: []types.Trigger{trigger1},
			expected: []types.QueuedTrigger{
				types.NewQueuedTrigger(trigger1, s.ctx.BlockTime(), uint64(s.ctx.BlockHeight())),
			},
		},
		{
			name:     "valid - Test queue multiple items",
			triggers: []types.Trigger{trigger1, trigger2},
			expected: []types.QueuedTrigger{
				types.NewQueuedTrigger(trigger1, s.ctx.BlockTime(), uint64(s.ctx.BlockHeight())),
				types.NewQueuedTrigger(trigger2, s.ctx.BlockTime(), uint64(s.ctx.BlockHeight())),
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			for _, trigger := range tc.triggers {
				s.app.TriggerKeeper.QueueTrigger(s.ctx, trigger)
			}
			for _, expected := range tc.expected {
				item := s.app.TriggerKeeper.QueuePeek(s.ctx)
				s.Equal(expected, *item, "should place correct item in front of queue for QueueTrigger")
				s.app.TriggerKeeper.Dequeue(s.ctx)
			}
		})
	}
}

func (s *KeeperTestSuite) TestQueuePeek() {
	owner := s.accountAddresses[0].String()
	trigger1 := s.CreateTrigger(1, owner, &types.BlockHeightEvent{BlockHeight: 120}, &types.MsgDestroyTriggerRequest{Id: 1, Authority: owner})
	trigger2 := s.CreateTrigger(2, owner, &types.BlockHeightEvent{BlockHeight: 130}, &types.MsgDestroyTriggerRequest{Id: 1, Authority: owner})
	tests := []struct {
		name     string
		triggers []types.Trigger
		expected *types.QueuedTrigger
	}{
		{
			name:     "valid - empty",
			triggers: []types.Trigger{},
			expected: nil,
		},
		{
			name:     "valid - single item",
			triggers: []types.Trigger{trigger1},
			expected: &types.QueuedTrigger{
				BlockHeight: uint64(s.ctx.BlockHeight()),
				Time:        s.ctx.BlockTime(),
				Trigger:     trigger1,
			},
		},
		{
			name:     "valid - multiple item should display only first",
			triggers: []types.Trigger{trigger1, trigger2},
			expected: &types.QueuedTrigger{
				BlockHeight: uint64(s.ctx.BlockHeight()),
				Time:        s.ctx.BlockTime(),
				Trigger:     trigger1,
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			for _, trigger := range tc.triggers {
				s.app.TriggerKeeper.QueueTrigger(s.ctx, trigger)
			}

			item := s.app.TriggerKeeper.QueuePeek(s.ctx)
			s.Equal(tc.expected, item, "should return correct value for QueuePeek")

			for range tc.triggers {
				s.app.TriggerKeeper.Dequeue(s.ctx)
			}
		})
	}
}

func (s *KeeperTestSuite) TestEnqueue() {
	owner := s.accountAddresses[0].String()
	trigger1 := s.CreateTrigger(1, owner, &types.BlockHeightEvent{BlockHeight: 120}, &types.MsgDestroyTriggerRequest{Id: 1, Authority: owner})
	trigger2 := s.CreateTrigger(2, owner, &types.BlockHeightEvent{BlockHeight: 130}, &types.MsgDestroyTriggerRequest{Id: 1, Authority: owner})
	tests := []struct {
		name     string
		triggers []types.QueuedTrigger
		expected []types.QueuedTrigger
	}{
		{
			name: "valid - Test queue single item",
			triggers: []types.QueuedTrigger{
				types.NewQueuedTrigger(trigger1, s.ctx.BlockTime(), uint64(s.ctx.BlockHeight())),
			},
			expected: []types.QueuedTrigger{
				types.NewQueuedTrigger(trigger1, s.ctx.BlockTime(), uint64(s.ctx.BlockHeight())),
			},
		},
		{
			name: "valid - Test queue multiple items",
			triggers: []types.QueuedTrigger{
				types.NewQueuedTrigger(trigger1, s.ctx.BlockTime(), uint64(s.ctx.BlockHeight())),
				types.NewQueuedTrigger(trigger2, s.ctx.BlockTime(), uint64(s.ctx.BlockHeight())),
			},
			expected: []types.QueuedTrigger{
				types.NewQueuedTrigger(trigger1, s.ctx.BlockTime(), uint64(s.ctx.BlockHeight())),
				types.NewQueuedTrigger(trigger2, s.ctx.BlockTime(), uint64(s.ctx.BlockHeight())),
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			for _, trigger := range tc.triggers {
				s.app.TriggerKeeper.Enqueue(s.ctx, trigger)
			}

			items, err := s.app.TriggerKeeper.GetAllQueueItems(s.ctx)
			s.NoError(err, "should not return an error when calling GetAllQueueItems")
			s.Equal(tc.expected, items, "should return all items when calling GetAllQueueItems")

			for range tc.expected {
				s.app.TriggerKeeper.Dequeue(s.ctx)
			}
		})
	}
}

func (s *KeeperTestSuite) TestDequeue() {
	owner := s.accountAddresses[0].String()
	trigger1 := s.CreateTrigger(1, owner, &types.BlockHeightEvent{BlockHeight: 120}, &types.MsgDestroyTriggerRequest{Id: 1, Authority: owner})
	trigger2 := s.CreateTrigger(2, owner, &types.BlockHeightEvent{BlockHeight: 130}, &types.MsgDestroyTriggerRequest{Id: 1, Authority: owner})
	tests := []struct {
		name     string
		triggers []types.Trigger
		expected []types.QueuedTrigger
		panic    string
	}{
		{
			name:     "valid - empty",
			triggers: []types.Trigger{},
			expected: []types.QueuedTrigger{},
			panic:    "unable to dequeue from empty queue.",
		},
		{
			name:     "valid - single item",
			triggers: []types.Trigger{trigger1},
			expected: []types.QueuedTrigger(nil),
		},
		{
			name:     "valid - multiple item should display only last",
			triggers: []types.Trigger{trigger1, trigger2},
			expected: []types.QueuedTrigger{{
				BlockHeight: uint64(s.ctx.BlockHeight()),
				Time:        s.ctx.BlockTime(),
				Trigger:     trigger2,
			}},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			for _, trigger := range tc.triggers {
				s.app.TriggerKeeper.QueueTrigger(s.ctx, trigger)
			}

			if len(tc.panic) > 0 {
				s.PanicsWithValue(tc.panic, func() {
					s.app.TriggerKeeper.Dequeue(s.ctx)
				})
			} else {
				s.app.TriggerKeeper.Dequeue(s.ctx)
				items, err := s.app.TriggerKeeper.GetAllQueueItems(s.ctx)
				s.NoError(err, "should not throw an error when obtaining GetAllQueueItems after Dequeue")
				s.Equal(tc.expected, items, "should have correct items in queue after Dequeue")
			}

			for !s.app.TriggerKeeper.QueueIsEmpty(s.ctx) {
				s.app.TriggerKeeper.Dequeue(s.ctx)
			}
		})
	}
}

func (s *KeeperTestSuite) TestQueueIsEmpty() {
	owner := s.accountAddresses[0].String()
	trigger1 := s.CreateTrigger(1, owner, &types.BlockHeightEvent{BlockHeight: 120}, &types.MsgDestroyTriggerRequest{Id: 1, Authority: owner})
	trigger2 := s.CreateTrigger(2, owner, &types.BlockHeightEvent{BlockHeight: 130}, &types.MsgDestroyTriggerRequest{Id: 1, Authority: owner})
	tests := []struct {
		name     string
		triggers []types.Trigger
		expected bool
	}{
		{
			name:     "valid - empty",
			triggers: []types.Trigger{},
			expected: true,
		},
		{
			name:     "valid - single item should not empty",
			triggers: []types.Trigger{trigger1},
			expected: false,
		},
		{
			name:     "valid - multiple item should not be empty",
			triggers: []types.Trigger{trigger1, trigger2},
			expected: false,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			for _, trigger := range tc.triggers {
				s.app.TriggerKeeper.QueueTrigger(s.ctx, trigger)
			}

			isEmpty := s.app.TriggerKeeper.QueueIsEmpty(s.ctx)
			s.Equal(tc.expected, isEmpty, "should return the correct value for QueueIsEmpty")

			for range tc.triggers {
				s.app.TriggerKeeper.Dequeue(s.ctx)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGetAllQueueItems() {
	owner := s.accountAddresses[0].String()
	trigger1 := s.CreateTrigger(1, owner, &types.BlockHeightEvent{BlockHeight: 120}, &types.MsgDestroyTriggerRequest{Id: 1, Authority: owner})
	trigger2 := s.CreateTrigger(2, owner, &types.BlockHeightEvent{BlockHeight: 130}, &types.MsgDestroyTriggerRequest{Id: 1, Authority: owner})
	tests := []struct {
		name     string
		triggers []types.Trigger
		expected []types.QueuedTrigger
	}{
		{
			name:     "valid - Test queue empty",
			triggers: []types.Trigger{},
			expected: []types.QueuedTrigger(nil),
		},
		{
			name:     "valid - Test queue single item",
			triggers: []types.Trigger{trigger1},
			expected: []types.QueuedTrigger{
				types.NewQueuedTrigger(trigger1, s.ctx.BlockTime(), uint64(s.ctx.BlockHeight())),
			},
		},
		{
			name:     "valid - Test queue multiple items",
			triggers: []types.Trigger{trigger1, trigger2},
			expected: []types.QueuedTrigger{
				types.NewQueuedTrigger(trigger1, s.ctx.BlockTime(), uint64(s.ctx.BlockHeight())),
				types.NewQueuedTrigger(trigger2, s.ctx.BlockTime(), uint64(s.ctx.BlockHeight())),
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			for _, trigger := range tc.triggers {
				s.app.TriggerKeeper.QueueTrigger(s.ctx, trigger)
			}

			items, err := s.app.TriggerKeeper.GetAllQueueItems(s.ctx)
			s.NoError(err, "should not throw an error when calling GetAllQueueItems")
			s.Equal(tc.expected, items, "should return all queue items when calling GetAllQueueItems")

			for range tc.expected {
				s.app.TriggerKeeper.Dequeue(s.ctx)
			}
		})
	}
}

// TestQueueCollectionsOperations tests queue operations with collections
func (s *KeeperTestSuite) TestQueueCollectionsOperations() {
	ctx := s.ctx
	k := s.app.TriggerKeeper

	s.Require().True(k.QueueIsEmpty(ctx), "Queue should initially be empty")

	trigger := s.CreateTrigger(1, s.accountAddr.String(), &types.BlockHeightEvent{BlockHeight: 100}, &types.MsgDestroyTriggerRequest{})
	k.QueueTrigger(ctx, trigger)

	s.Require().False(k.QueueIsEmpty(ctx), "Queue should not be empty after enqueue")

	item := k.QueuePeek(ctx)
	s.Require().NotNil(item, "QueuePeek should return the first queued item")
	s.Require().Equal(trigger.Id, item.Trigger.Id, "Peeked trigger ID should match enqueued trigger")

	k.Dequeue(ctx)

	s.Require().True(k.QueueIsEmpty(ctx), "Queue should be empty after dequeue")
}

func (s *KeeperTestSuite) TestQueueMultipleItems() {
	ctx := s.ctx
	k := s.app.TriggerKeeper

	for i := uint64(1); i <= 3; i++ {
		trigger := s.CreateTrigger(i, s.accountAddr.String(), &types.BlockHeightEvent{BlockHeight: 100 + i}, &types.MsgDestroyTriggerRequest{})
		k.QueueTrigger(ctx, trigger)
	}

	length, err := k.GetQueueLength(ctx) // use getQueueLength
	s.Require().NoError(err)
	s.Require().Equal(uint64(3), length, "Queue should have 3 items after enqueue")

	item := k.QueuePeek(ctx)
	s.Require().NotNil(item)
	s.Require().Equal(uint64(1), item.Trigger.Id, "First item should have ID 1 (FIFO)")

	k.Dequeue(ctx)

	item = k.QueuePeek(ctx)
	s.Require().NotNil(item)
	s.Require().Equal(uint64(2), item.Trigger.Id, "Second item should have ID 2 after dequeue")

	length, err = k.GetQueueLength(ctx)
	s.Require().NoError(err)
	s.Require().Equal(uint64(2), length, "Queue length should decrease to 2 after dequeue")
}

func (s *KeeperTestSuite) TestQueueLengthConsistency() {
	ctx := s.ctx
	k := s.app.TriggerKeeper

	length, err := k.GetQueueLength(ctx)
	s.Require().NoError(err)
	s.Require().Equal(uint64(0), length)

	for i := uint64(1); i <= 3; i++ {
		trigger := s.CreateTrigger(i, s.accountAddr.String(), &types.BlockHeightEvent{BlockHeight: 100 + i}, &types.MsgDestroyTriggerRequest{})
		k.QueueTrigger(ctx, trigger)
	}

	length, err = k.GetQueueLength(ctx)
	s.Require().NoError(err)
	s.Require().Equal(uint64(3), length, "Queue length should be 3 after enqueueing 3 items")

	k.Dequeue(ctx)

	length, err = k.GetQueueLength(ctx)
	s.Require().NoError(err)
	s.Require().Equal(uint64(2), length, "Queue length should decrease to 2 after one dequeue")
}
