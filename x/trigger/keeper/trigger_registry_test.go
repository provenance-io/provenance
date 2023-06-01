package keeper_test

import (
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/trigger/types"
)

func (s *KeeperTestSuite) TestRegisterTrigger() {
	owner := s.accountAddresses[0].String()

	tests := []struct {
		name     string
		trigger  types.Trigger
		meter    sdk.GasMeter
		expected uint64
		panics   bool
	}{
		{
			name:     "valid - register with infinite gas meter",
			meter:    sdk.NewInfiniteGasMeter(),
			trigger:  s.CreateTrigger(1, owner, &types.BlockHeightEvent{BlockHeight: uint64(s.ctx.BlockHeight())}, &types.MsgDestroyTriggerRequest{Id: 100, Authority: owner}),
			expected: 2000000,
			panics:   false,
		},
		{
			name:     "invalid - register with no gas",
			meter:    sdk.NewGasMeter(0),
			trigger:  s.CreateTrigger(1, owner, &types.BlockHeightEvent{BlockHeight: uint64(s.ctx.BlockHeight())}, &types.MsgDestroyTriggerRequest{Id: 100, Authority: owner}),
			expected: 0,
			panics:   true,
		},
		{
			name:     "valid - register with no gas for trigger",
			meter:    sdk.NewGasMeter(19890),
			trigger:  s.CreateTrigger(1, owner, &types.BlockHeightEvent{BlockHeight: uint64(s.ctx.BlockHeight())}, &types.MsgDestroyTriggerRequest{Id: 100, Authority: owner}),
			expected: 0,
			panics:   false,
		},
		{
			name:     "valid - register with gas",
			meter:    sdk.NewGasMeter(9999999999),
			trigger:  s.CreateTrigger(1, owner, &types.BlockHeightEvent{BlockHeight: uint64(s.ctx.BlockHeight())}, &types.MsgDestroyTriggerRequest{Id: 100, Authority: owner}),
			expected: 2000000,
			panics:   false,
		},
		{
			name:     "valid - register with maximum gas",
			meter:    sdk.NewGasMeter(math.MaxUint64),
			trigger:  s.CreateTrigger(1, owner, &types.BlockHeightEvent{BlockHeight: uint64(s.ctx.BlockHeight())}, &types.MsgDestroyTriggerRequest{Id: 100, Authority: owner}),
			expected: 2000000,
			panics:   false,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.ctx = s.ctx.WithGasMeter(tc.meter)

			if !tc.panics {
				s.app.TriggerKeeper.RegisterTrigger(s.ctx, tc.trigger)
				s.ctx = s.ctx.WithGasMeter(sdk.NewInfiniteGasMeter())

				trigger, err := s.app.TriggerKeeper.GetTrigger(s.ctx, tc.trigger.Id)
				s.NoError(err)
				s.Equal(tc.trigger, trigger)

				listener, err := s.app.TriggerKeeper.GetEventListener(s.ctx, types.BlockHeightPrefix, tc.trigger.Id)
				s.NoError(err)
				s.Equal(tc.trigger, listener)

				gasLimit := s.app.TriggerKeeper.GetGasLimit(s.ctx, tc.trigger.Id)
				s.Equal(tc.expected, gasLimit)

				s.app.TriggerKeeper.UnregisterTrigger(s.ctx, trigger)
				s.app.TriggerKeeper.RemoveGasLimit(s.ctx, tc.trigger.Id)
			} else {
				s.Panics(func() {
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
			s.Error(err)
			_, err = s.app.TriggerKeeper.GetEventListener(s.ctx, types.BlockHeightPrefix, tc.trigger.Id)
			s.Error(err)
		})
	}
}
