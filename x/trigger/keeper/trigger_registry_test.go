package keeper_test

import (
	"math"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/trigger/types"
)

func (s *KeeperTestSuite) TestRegisterTrigger() {
	owner := s.accountAddresses[0].String()

	tests := []struct {
		name     string
		trigger  types.Trigger
		meter    sdk.GasMeter
		expected int
		panic    *storetypes.ErrorOutOfGas
	}{
		{
			name:     "valid - register with infinite gas meter",
			meter:    sdk.NewInfiniteGasMeter(),
			trigger:  s.CreateTrigger(1, owner, &types.BlockHeightEvent{BlockHeight: uint64(s.ctx.BlockHeight())}, &types.MsgDestroyTriggerRequest{Id: 100, Authority: owner}),
			expected: 2000000,
		},
		{
			name:     "invalid - register with no gas",
			meter:    sdk.NewGasMeter(0),
			trigger:  s.CreateTrigger(1, owner, &types.BlockHeightEvent{BlockHeight: uint64(s.ctx.BlockHeight())}, &types.MsgDestroyTriggerRequest{Id: 100, Authority: owner}),
			expected: 0,
			panic:    &storetypes.ErrorOutOfGas{Descriptor: "WriteFlat"},
		},
		{
			name:     "valid - register with no gas for trigger",
			meter:    sdk.NewGasMeter(20130),
			trigger:  s.CreateTrigger(1, owner, &types.BlockHeightEvent{BlockHeight: uint64(s.ctx.BlockHeight())}, &types.MsgDestroyTriggerRequest{Id: 100, Authority: owner}),
			expected: 0,
		},
		{
			name:     "valid - register with gas",
			meter:    sdk.NewGasMeter(9999999999),
			trigger:  s.CreateTrigger(1, owner, &types.BlockHeightEvent{BlockHeight: uint64(s.ctx.BlockHeight())}, &types.MsgDestroyTriggerRequest{Id: 100, Authority: owner}),
			expected: 2000000,
		},
		{
			name:     "valid - register with maximum gas",
			meter:    sdk.NewGasMeter(math.MaxUint64),
			trigger:  s.CreateTrigger(1, owner, &types.BlockHeightEvent{BlockHeight: uint64(s.ctx.BlockHeight())}, &types.MsgDestroyTriggerRequest{Id: 100, Authority: owner}),
			expected: 2000000,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.ctx = s.ctx.WithGasMeter(tc.meter)

			if tc.panic == nil {
				s.app.TriggerKeeper.RegisterTrigger(s.ctx, tc.trigger)
				s.ctx = s.ctx.WithGasMeter(sdk.NewInfiniteGasMeter())

				trigger, err := s.app.TriggerKeeper.GetTrigger(s.ctx, tc.trigger.Id)
				s.NoError(err, "should add trigger to store in RegisterTrigger")
				s.Equal(tc.trigger, trigger, "should be correct trigger that is stored for RegisterTrigger")

				triggerEvent, _ := tc.trigger.GetTriggerEventI()
				listener, err := s.app.TriggerKeeper.GetEventListener(s.ctx, triggerEvent.GetEventPrefix(), triggerEvent.GetEventOrder(), tc.trigger.Id)
				s.NoError(err, "should add trigger to event listener store in RegisterTrigger")
				s.Equal(tc.trigger, listener, "should be correct trigger that is stored in RegisterTrigger")

				gasLimit := int(s.app.TriggerKeeper.GetGasLimit(s.ctx, tc.trigger.Id))
				s.Equal(tc.expected, gasLimit, "should store correct gas limit in RegisterTrigger")

				s.app.TriggerKeeper.UnregisterTrigger(s.ctx, trigger)
				s.app.TriggerKeeper.RemoveGasLimit(s.ctx, tc.trigger.Id)
			} else {
				s.PanicsWithValue(*tc.panic, func() {
					s.app.TriggerKeeper.RegisterTrigger(s.ctx, tc.trigger)
				})
			}

		})
	}
}

func (s *KeeperTestSuite) TestUnregisterTrigger() {
	owner := s.accountAddresses[0].String()

	tests := []struct {
		name    string
		exists  bool
		trigger types.Trigger
		err     error
	}{
		{
			name:    "valid - unregister existing trigger",
			exists:  true,
			trigger: s.CreateTrigger(1, owner, &types.BlockHeightEvent{BlockHeight: uint64(s.ctx.BlockHeight())}, &types.MsgDestroyTriggerRequest{Id: 100, Authority: owner}),
		},
		{
			name:    "invalid - unregister non existant trigger",
			exists:  false,
			trigger: s.CreateTrigger(1, owner, &types.BlockHeightEvent{BlockHeight: uint64(s.ctx.BlockHeight())}, &types.MsgDestroyTriggerRequest{Id: 100, Authority: owner}),
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			if tc.exists {
				s.app.TriggerKeeper.RegisterTrigger(s.ctx, tc.trigger)
				s.ctx.GasMeter().RefundGas(s.ctx.GasMeter().GasConsumed(), "testing")
			}
			s.app.TriggerKeeper.UnregisterTrigger(s.ctx, tc.trigger)

			_, err := s.app.TriggerKeeper.GetTrigger(s.ctx, tc.trigger.Id)
			s.EqualError(err, types.ErrTriggerNotFound.Error())
			triggerEvent, _ := tc.trigger.GetTriggerEventI()
			_, err = s.app.TriggerKeeper.GetEventListener(s.ctx, triggerEvent.GetEventPrefix(), triggerEvent.GetEventOrder(), tc.trigger.Id)
			s.EqualError(err, types.ErrEventNotFound.Error())
		})
	}
}
