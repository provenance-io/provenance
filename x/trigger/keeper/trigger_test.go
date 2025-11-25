package keeper_test

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"

	"github.com/provenance-io/provenance/x/trigger/types"
)

func (s *KeeperTestSuite) TestGetAllTriggers() {
	var event types.TriggerEventI = &types.BlockHeightEvent{BlockHeight: 1}
	actions, _ := sdktx.SetMsgs([]sdk.Msg{&types.MsgDestroyTriggerRequest{Id: 1, Authority: s.accountAddresses[0].String()}})
	eventAny, _ := codectypes.NewAnyWithValue(event)

	tests := []struct {
		name     string
		triggers []types.Trigger
	}{
		{
			name:     "valid - no triggers",
			triggers: []types.Trigger(nil),
		},
		{
			name: "valid - one trigger",
			triggers: []types.Trigger{
				types.NewTrigger(1, s.accountAddresses[0].String(), eventAny, actions),
			},
		},
		{
			name: "valid - multiple triggers",
			triggers: []types.Trigger{
				types.NewTrigger(1, s.accountAddresses[0].String(), eventAny, actions),
				types.NewTrigger(2, s.accountAddresses[0].String(), eventAny, actions),
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			for _, trigger := range tc.triggers {
				s.app.TriggerKeeper.SetTrigger(s.ctx, trigger)
			}
			triggers, err := s.app.TriggerKeeper.GetAllTriggers(s.ctx)
			s.NoError(err, "should have no error for GetAllTriggers")
			s.Equal(tc.triggers, triggers, "should have matching trigger output for GetAllTriggers")
		})
	}
}

func (s *KeeperTestSuite) TestGetAndSetTrigger() {
	var event types.TriggerEventI = &types.BlockHeightEvent{BlockHeight: 1}
	actions, _ := sdktx.SetMsgs([]sdk.Msg{&types.MsgDestroyTriggerRequest{Id: 1, Authority: s.accountAddresses[0].String()}})
	eventAny, _ := codectypes.NewAnyWithValue(event)

	expected := types.NewTrigger(1, s.accountAddresses[0].String(), eventAny, actions)

	s.app.TriggerKeeper.SetTrigger(s.ctx, expected)

	tests := []struct {
		name     string
		id       types.TriggerID
		expected *types.Trigger
		err      string
	}{
		{
			name:     "valid - trigger",
			id:       1,
			expected: &expected,
			err:      "",
		},
		{
			name:     "invalid - trigger id doesn't exist",
			id:       2,
			expected: nil,
			err:      types.ErrTriggerNotFound.Error(),
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			trigger, err := s.app.TriggerKeeper.GetTrigger(s.ctx, tc.id)
			if len(tc.err) > 0 {
				s.EqualError(err, tc.err, "should have correct error for invalid GetTrigger")
				s.Equal(0, int(trigger.Id), "should have invalid id for invalid GetTrigger")
			} else {
				s.NoError(err, "should have no error for valid GetTrigger")
				s.Equal(*tc.expected, trigger, "should have correct output for valid GetTrigger")
			}
		})
	}
}

func (s *KeeperTestSuite) TestNewTriggerWithID() {
	var event types.TriggerEventI = &types.BlockHeightEvent{BlockHeight: 1}
	actions, _ := sdktx.SetMsgs([]sdk.Msg{&types.MsgDestroyTriggerRequest{Id: 1, Authority: s.accountAddresses[0].String()}})
	eventAny, _ := codectypes.NewAnyWithValue(event)

	tests := []struct {
		name     string
		expected types.Trigger
	}{
		{
			name:     "valid - trigger first id",
			expected: types.NewTrigger(1, s.accountAddresses[0].String(), eventAny, actions),
		},
		{
			name:     "invalid - trigger id increments",
			expected: types.NewTrigger(2, s.accountAddresses[0].String(), eventAny, actions),
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			trigger, _ := s.app.TriggerKeeper.NewTriggerWithID(s.ctx, tc.expected.Owner, tc.expected.Event, tc.expected.Actions)
			s.Equal(tc.expected, trigger, "should have correct trigger from NewTrigger")
		})
	}
}

func (s *KeeperTestSuite) TestIterateTriggers() {
	var event types.TriggerEventI = &types.BlockHeightEvent{BlockHeight: 1}
	actions, _ := sdktx.SetMsgs([]sdk.Msg{&types.MsgDestroyTriggerRequest{Id: 1, Authority: s.accountAddresses[0].String()}})
	eventAny, _ := codectypes.NewAnyWithValue(event)

	tests := []struct {
		name      string
		shortStop bool
		add       []types.Trigger
		count     int
	}{
		{
			name:      "valid - no triggers",
			shortStop: false,
			add:       []types.Trigger{},
			count:     0,
		},
		{
			name:      "valid - one trigger",
			shortStop: false,
			add:       []types.Trigger{types.NewTrigger(1, s.accountAddresses[0].String(), eventAny, actions)},
			count:     1,
		},
		{
			name:      "valid - multiple trigger",
			shortStop: false,
			add:       []types.Trigger{types.NewTrigger(2, s.accountAddresses[0].String(), eventAny, actions)},
			count:     2,
		},
		{
			name:      "valid - do not iterate through all",
			shortStop: true,
			add:       []types.Trigger{},
			count:     1,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			count := 0
			for _, trigger := range tc.add {
				s.app.TriggerKeeper.SetTrigger(s.ctx, trigger)
			}
			err := s.app.TriggerKeeper.IterateTriggers(s.ctx, func(trigger types.Trigger) (stop bool, err error) {
				count += 1
				return tc.shortStop, nil
			})
			s.NoError(err, "should have no error from IterateTriggers")
			s.Equal(tc.count, count, "should iterate the correct number of items in IterateTriggers")
		})
	}
}

func (s *KeeperTestSuite) TestRemoveTrigger() {
	var event types.TriggerEventI = &types.BlockHeightEvent{BlockHeight: 1}
	actions, _ := sdktx.SetMsgs([]sdk.Msg{&types.MsgDestroyTriggerRequest{Id: 1, Authority: s.accountAddresses[0].String()}})
	eventAny, _ := codectypes.NewAnyWithValue(event)

	expected := types.NewTrigger(1, s.accountAddresses[0].String(), eventAny, actions)

	s.app.TriggerKeeper.SetTrigger(s.ctx, expected)

	tests := []struct {
		name     string
		id       types.TriggerID
		expected bool
	}{
		{
			name:     "valid - trigger",
			id:       1,
			expected: true,
		},
		{
			name:     "invalid - trigger id doesn't exist",
			id:       2,
			expected: false,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			success := s.app.TriggerKeeper.RemoveTrigger(s.ctx, tc.id)
			s.Equal(tc.expected, success, "should have the correct output for RemoveTrigger")
		})
	}
}

// TestTriggerCollectionsOperations tests basic trigger operations with collections
func (s *KeeperTestSuite) TestTriggerCollectionsOperations() {
	ctx := s.ctx
	k := s.app.TriggerKeeper

	// Test SetTrigger
	trigger := s.CreateTrigger(1, s.accountAddr.String(), &types.BlockHeightEvent{BlockHeight: 100}, &types.MsgDestroyTriggerRequest{})
	err := k.SetTrigger(ctx, trigger)
	s.Require().NoError(err, "SetTrigger should not return an error")

	// Test GetTrigger
	retrieved, err := k.GetTrigger(ctx, 1)
	s.Require().NoError(err, "GetTrigger should not return an error for existing ID")
	s.Require().Equal(trigger.Id, retrieved.Id)
	s.Require().Equal(trigger.Owner, retrieved.Owner)
	s.Require().Equal(trigger.Event.TypeUrl, retrieved.Event.TypeUrl)
	s.Require().Equal(trigger.Event.Value, retrieved.Event.Value)

	s.Require().Equal(len(trigger.Actions), len(retrieved.Actions))
	s.Require().Equal(trigger.Actions[0].TypeUrl, retrieved.Actions[0].TypeUrl)

	// Test HasTrigger
	exists, err := k.HasTrigger(ctx, 1)
	s.Require().NoError(err, "HasTrigger should not error for existing ID")
	s.Require().True(exists, "HasTrigger should return true for an existing ID")

	exists, err = k.HasTrigger(ctx, 999)
	s.Require().NoError(err, "HasTrigger should not error for non-existing ID")
	s.Require().False(exists, "HasTrigger should return false for non-existing ID")

	// Test RemoveTrigger
	removed := k.RemoveTrigger(ctx, 1)
	s.Require().True(removed, "RemoveTrigger should indicate successful removal for existing ID")

	// Verify removal
	exists, err = k.HasTrigger(ctx, 1)
	s.Require().NoError(err, "HasTrigger after removal should not return an error")
	s.Require().False(exists, "HasTrigger after removal should report false")
}
