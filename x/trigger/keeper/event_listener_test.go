package keeper_test

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"

	"github.com/provenance-io/provenance/x/trigger/types"
)

func (s *KeeperTestSuite) TestGetAndSetEventListener() {
	var event types.TriggerEventI = &types.BlockHeightEvent{BlockHeight: 1}
	actions, _ := sdktx.SetMsgs([]sdk.Msg{&types.MsgDestroyTriggerRequest{Id: 1, Authority: s.accountAddresses[0].String()}})
	eventAny, _ := codectypes.NewAnyWithValue(event)
	newTrigger := types.NewTrigger(1, s.accountAddresses[0].String(), eventAny, actions)
	s.app.TriggerKeeper.SetEventListener(s.ctx, newTrigger)
	s.app.TriggerKeeper.SetTrigger(s.ctx, newTrigger)

	tests := []struct {
		name     string
		id       types.TriggerID
		expected *types.Trigger
		prefix   string
		order    uint64
		err      string
	}{
		{
			name:     "valid - event listener",
			id:       1,
			expected: &newTrigger,
			prefix:   types.BlockHeightPrefix,
			order:    1,
			err:      "",
		},
		{
			name:     "invalid - event listener doesn't exist",
			id:       999,
			expected: nil,
			prefix:   types.BlockHeightPrefix,
			order:    1,
			err:      types.ErrEventNotFound.Error(),
		},
		{
			name:     "invalid - event listener doesn't exist with wrong prefix",
			id:       1,
			expected: nil,
			prefix:   types.BlockTimePrefix,
			order:    1,
			err:      types.ErrEventNotFound.Error(),
		},
		{
			name:     "invalid - event listener doesn't exist with wrong order",
			id:       1,
			expected: nil,
			prefix:   types.BlockTimePrefix,
			order:    3,
			err:      types.ErrEventNotFound.Error(),
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			listener, err := s.app.TriggerKeeper.GetEventListener(s.ctx, tc.prefix, tc.order, tc.id)
			if len(tc.err) > 0 {
				s.EqualError(err, tc.err, "should have correct error for invalid GetEventListener")
				s.Equal(0, int(listener.Id), "should have id of 0 for invalid GetEventListener")
			} else {
				s.NoError(err, "should have no error for valid GetEventListener")
				s.Equal(*tc.expected, listener, "should receive the correct listener from GetEventListener")
			}
		})
	}
}

func (s *KeeperTestSuite) TestIterateEventListeners() {
	var event types.TriggerEventI = &types.BlockHeightEvent{BlockHeight: 1}
	var event2 types.TriggerEventI = &types.BlockTimeEvent{Time: s.ctx.BlockTime()}
	actions, _ := sdktx.SetMsgs([]sdk.Msg{&types.MsgDestroyTriggerRequest{Id: 1, Authority: s.accountAddresses[0].String()}})
	eventAny, _ := codectypes.NewAnyWithValue(event)
	eventAny2, _ := codectypes.NewAnyWithValue(event2)

	newTrigger := types.NewTrigger(1, s.accountAddresses[0].String(), eventAny, actions)
	newTrigger2 := types.NewTrigger(2, s.accountAddresses[0].String(), eventAny, actions)
	newTrigger3 := types.NewTrigger(3, s.accountAddresses[0].String(), eventAny2, actions)

	tests := []struct {
		name      string
		shortStop bool
		add       []types.Trigger
		count     int
		prefix    string
	}{
		{
			name:      "valid - no event listeners",
			shortStop: false,
			add:       []types.Trigger{},
			count:     0,
			prefix:    types.BlockHeightPrefix,
		},
		{
			name:      "valid - one event listener",
			shortStop: false,
			add:       []types.Trigger{newTrigger},
			count:     1,
			prefix:    types.BlockHeightPrefix,
		},
		{
			name:      "valid - multiple event listeners witih prefix",
			shortStop: false,
			add:       []types.Trigger{newTrigger2, newTrigger3},
			count:     2,
			prefix:    types.BlockHeightPrefix,
		},
		{
			name:      "valid - do not iterate through all",
			shortStop: true,
			add:       []types.Trigger{},
			count:     1,
			prefix:    types.BlockHeightPrefix,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			count := 0
			for _, trigger := range tc.add {
				s.app.TriggerKeeper.SetEventListener(s.ctx, trigger)
				s.app.TriggerKeeper.SetTrigger(s.ctx, trigger)
			}
			err := s.app.TriggerKeeper.IterateEventListeners(s.ctx, tc.prefix, func(trigger types.Trigger) (stop bool, err error) {
				count += 1
				return tc.shortStop, nil
			})
			s.NoError(err, "should receive no error from IterateEventListeners")
			s.Equal(tc.count, count, "should iterate the correct number of times for IterateEventListeners")
		})
	}
}

func (s *KeeperTestSuite) TestRemoveEventListener() {
	var event types.TriggerEventI = &types.BlockHeightEvent{BlockHeight: 1}
	actions, _ := sdktx.SetMsgs([]sdk.Msg{&types.MsgDestroyTriggerRequest{Id: 1, Authority: s.accountAddresses[0].String()}})
	eventAny, _ := codectypes.NewAnyWithValue(event)

	trigger := types.NewTrigger(1, s.accountAddresses[0].String(), eventAny, actions)
	trigger2 := types.NewTrigger(2, s.accountAddresses[0].String(), eventAny, actions)
	s.app.TriggerKeeper.SetEventListener(s.ctx, trigger)

	tests := []struct {
		name     string
		trigger  types.Trigger
		expected bool
	}{
		{
			name:     "valid - event listener removal",
			trigger:  trigger,
			expected: true,
		},
		{
			name:     "invalid - event listener doesn't exist",
			trigger:  trigger2,
			expected: false,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			success, err := s.app.TriggerKeeper.RemoveEventListener(s.ctx, tc.trigger)
			s.Require().NoError(err, "RemoveEventListener should not return an error")
			s.Equal(tc.expected, success, "should return the correct result for RemoveEventListener")
		})
	}
}
func (s *KeeperTestSuite) TestCollectionsSetAndGetEventListener() {
	ctx := s.ctx
	k := s.app.TriggerKeeper

	// Create events and triggers
	blockHeightEvent := &types.BlockHeightEvent{BlockHeight: 100}
	blockTimeEvent := &types.BlockTimeEvent{Time: ctx.BlockTime()}

	actions, _ := sdktx.SetMsgs([]sdk.Msg{
		&types.MsgDestroyTriggerRequest{Id: 1, Authority: s.accountAddresses[0].String()},
	})

	eventAny1, _ := codectypes.NewAnyWithValue(blockHeightEvent)
	eventAny2, _ := codectypes.NewAnyWithValue(blockTimeEvent)

	trigger1 := types.NewTrigger(1, s.accountAddresses[0].String(), eventAny1, actions)
	trigger2 := types.NewTrigger(2, s.accountAddresses[0].String(), eventAny1, actions)
	trigger3 := types.NewTrigger(3, s.accountAddresses[0].String(), eventAny2, actions)

	s.Require().NoError(k.SetTrigger(ctx, trigger1), "failed to set trigger1")
	s.Require().NoError(k.SetTrigger(ctx, trigger2), "failed to set trigger2")
	s.Require().NoError(k.SetTrigger(ctx, trigger3), "failed to set trigger3")

	s.Require().NoError(k.SetEventListener(ctx, trigger1), "failed to set event listener for trigger1")
	s.Require().NoError(k.SetEventListener(ctx, trigger2), "failed to set event listener for trigger2")
	s.Require().NoError(k.SetEventListener(ctx, trigger3), "failed to set event listener for trigger3")

	tests := []struct {
		name      string
		trigger   types.Trigger
		expectErr bool
	}{
		{"valid listener 1", trigger1, false},
		{"valid listener 2", trigger2, false},
		{"valid listener 3", trigger3, false},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			eventI, err := tc.trigger.GetTriggerEventI()
			s.Require().NoError(err, "failed to get TriggerEventI for %s", tc.name)

			listener, err := k.GetEventListener(ctx, eventI.GetEventPrefix(), eventI.GetEventOrder(), tc.trigger.Id)
			s.Require().NoError(err, "expected no error when fetching listener for %s", tc.name)
			s.Require().Equal(tc.trigger.Id, listener.Id, "listener ID mismatch for %s", tc.name)
		})
	}

	count := 0
	err := k.IterateEventListeners(ctx, blockHeightEvent.GetEventPrefix(), func(trigger types.Trigger) (bool, error) {
		count++
		return false, nil
	})
	s.Require().NoError(err, "failed to iterate over event listeners for blockHeightEvent")
	s.Require().Equal(2, count, "expected to iterate over 2 blockHeightEvent listeners, got %d", count)

	removed, err := k.RemoveEventListener(ctx, trigger1)
	s.Require().NoError(err, "failed to remove trigger1 listener")
	s.Require().True(removed, "trigger1 listener should have been removed")

	eventI, _ := trigger1.GetTriggerEventI()
	_, err = k.GetEventListener(ctx, eventI.GetEventPrefix(), eventI.GetEventOrder(), trigger1.Id)
	s.Require().ErrorIs(err, types.ErrEventNotFound, "removed listener for trigger1 should not be found")
}

func (s *KeeperTestSuite) TestCollectionsRemoveAllEventListenersForTrigger() {
	ctx := s.ctx
	k := s.app.TriggerKeeper

	blockHeightEvent := &types.BlockHeightEvent{BlockHeight: 200}
	actions, _ := sdktx.SetMsgs([]sdk.Msg{
		&types.MsgDestroyTriggerRequest{Id: 1, Authority: s.accountAddresses[0].String()},
	})
	eventAny, _ := codectypes.NewAnyWithValue(blockHeightEvent)
	trigger := types.NewTrigger(10, s.accountAddresses[0].String(), eventAny, actions)

	s.Require().NoError(k.SetTrigger(ctx, trigger), "failed to set trigger for RemoveAllEventListeners test")
	s.Require().NoError(k.SetEventListener(ctx, trigger), "failed to set event listener for trigger")

	err := k.RemoveAllEventListenersForTrigger(ctx, trigger.Id)
	s.Require().NoError(err, "failed to remove all event listeners for trigger")

	eventI, _ := trigger.GetTriggerEventI()
	_, err = k.GetEventListener(ctx, eventI.GetEventPrefix(), eventI.GetEventOrder(), trigger.Id)
	s.Require().ErrorIs(err, types.ErrEventNotFound, "listener should have been removed after RemoveAllEventListenersForTrigger")
}
