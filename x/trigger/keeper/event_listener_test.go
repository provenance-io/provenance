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

	tests := []struct {
		name     string
		id       types.TriggerID
		expected *types.Trigger
		prefix   string
		err      string
	}{
		{
			name:     "valid - event listener",
			id:       1,
			expected: &newTrigger,
			prefix:   types.BlockHeightPrefix,
			err:      "",
		},
		{
			name:     "invalid - event listener doesn't exist",
			id:       999,
			expected: nil,
			prefix:   types.BlockHeightPrefix,
			err:      types.ErrEventNotFound.Error(),
		},
		{
			name:     "invalid - event listener doesn't exist with wrong prefix",
			id:       1,
			expected: nil,
			prefix:   types.BlockTimePrefix,
			err:      types.ErrEventNotFound.Error(),
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			listener, err := s.app.TriggerKeeper.GetEventListener(s.ctx, tc.prefix, tc.id)
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
			success := s.app.TriggerKeeper.RemoveEventListener(s.ctx, tc.trigger)
			s.Equal(tc.expected, success, "should return the correct result for RemoveEventListener")
		})
	}
}
