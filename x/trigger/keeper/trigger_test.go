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
			s.NoError(err)
			s.Equal(tc.triggers, triggers)
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
				s.Error(err)
				s.EqualError(err, tc.err)
				s.Equal(uint64(0), trigger.Id)
			} else {
				s.NoError(err)
				s.Equal(*tc.expected, trigger)
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
			trigger := s.app.TriggerKeeper.NewTriggerWithID(s.ctx, tc.expected.Owner, tc.expected.Event, tc.expected.Actions)
			s.Equal(tc.expected, trigger)
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
		count     uint64
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
			count := uint64(0)
			for _, trigger := range tc.add {
				s.app.TriggerKeeper.SetTrigger(s.ctx, trigger)
			}
			err := s.app.TriggerKeeper.IterateTriggers(s.ctx, func(trigger types.Trigger) (stop bool, err error) {
				count += 1
				return tc.shortStop, nil
			})
			s.NoError(err)
			s.Equal(tc.count, count)
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
			s.Equal(tc.expected, success)
		})
	}
}
