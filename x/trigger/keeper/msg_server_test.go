package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/trigger/types"
)

func (s *KeeperTestSuite) TestCreateTrigger() {
	owner := s.accountAddresses[0].String()
	var event types.TriggerEventI = &types.BlockHeightEvent{BlockHeight: 130}
	action := types.MsgDestroyTriggerRequest{Id: 100, Authority: owner}

	tests := []struct {
		name       string
		request    *types.MsgCreateTriggerRequest
		expectedId types.TriggerID
		err        string
	}{
		{
			name:       "valid - single trigger created",
			request:    types.NewCreateTriggerRequest(owner, event, []sdk.Msg{&action}),
			expectedId: 1,
			err:        "",
		},
		{
			name:       "valid - second trigger has incremented id",
			request:    types.NewCreateTriggerRequest(owner, event, []sdk.Msg{&action}),
			expectedId: 2,
			err:        "",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.ctx = s.ctx.WithGasMeter(sdk.NewGasMeter(9999999999))
			s.ctx = s.ctx.WithEventManager(sdk.NewEventManager())
			result, err := s.handler(s.ctx, tc.request)
			s.ctx = s.ctx.WithGasMeter(sdk.NewGasMeter(9999999999))

			if len(tc.err) == 0 {
				s.NoError(err, "handler should not throw an error")
				resultEvent, _ := sdk.TypedEventToEvent(&types.EventTriggerCreated{
					TriggerId: fmt.Sprintf("%d", tc.expectedId),
				})
				s.Equal(sdk.Events{resultEvent}, result.GetEvents())
				_, err = s.app.TriggerKeeper.GetEventListener(s.ctx, event.GetEventPrefix(), tc.expectedId)
				s.NoError(err, "event listener should exist")
				_, err = s.app.TriggerKeeper.GetTrigger(s.ctx, tc.expectedId)
				s.NoError(err, "trigger should exist")
				gasLimit := s.app.TriggerKeeper.GetGasLimit(s.ctx, tc.expectedId)
				s.Equal(uint64(2000000), gasLimit, "gas limit should be set for trigger")
			} else {
				s.EqualError(err, tc.err, "handler should throw an error")
			}

		})
	}
}

func (s *KeeperTestSuite) TestDestroyTrigger() {
	owner := s.accountAddresses[0].String()
	owner2 := s.accountAddresses[1].String()
	var event types.TriggerEventI = &types.BlockHeightEvent{BlockHeight: 130}
	action := types.MsgDestroyTriggerRequest{Id: 100, Authority: owner}

	setupRequests := []*types.MsgCreateTriggerRequest{
		types.NewCreateTriggerRequest(owner, event, []sdk.Msg{&action}),
		types.NewCreateTriggerRequest(owner, event, []sdk.Msg{&action}),
		types.NewCreateTriggerRequest(owner2, event, []sdk.Msg{&action}),
	}
	for _, request := range setupRequests {
		s.ctx = s.ctx.WithGasMeter(sdk.NewGasMeter(9999999999))
		s.handler(s.ctx, request)
	}

	tests := []struct {
		name    string
		request *types.MsgDestroyTriggerRequest
		err     string
	}{
		{
			name:    "valid - single trigger destroyed",
			request: types.NewDestroyTriggerRequest(owner, 1),
			err:     "",
		},
		{
			name:    "valid - multiple triggers destroyed",
			request: types.NewDestroyTriggerRequest(owner, 2),
			err:     "",
		},
		{
			name:    "invalid - destroy a non existant trigger",
			request: types.NewDestroyTriggerRequest(owner, 100),
			err:     "trigger not found",
		},
		{
			name:    "invalid - destroy a trigger that is not owned by the user",
			request: types.NewDestroyTriggerRequest(owner, 3),
			err:     "signer does not have authority to destroy trigger",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.ctx = s.ctx.WithGasMeter(sdk.NewGasMeter(9999999999))
			s.ctx = s.ctx.WithEventManager(sdk.NewEventManager())
			result, err := s.handler(s.ctx, tc.request)
			s.ctx = s.ctx.WithGasMeter(sdk.NewGasMeter(9999999999))

			if len(tc.err) == 0 {
				s.NoError(err, "handler should not throw an error")
				resultEvent, _ := sdk.TypedEventToEvent(&types.EventTriggerDestroyed{
					TriggerId: fmt.Sprintf("%d", tc.request.GetId()),
				})
				s.Equal(sdk.Events{resultEvent}, result.GetEvents(), "result events should match")
				_, err = s.app.TriggerKeeper.GetEventListener(s.ctx, event.GetEventPrefix(), tc.request.GetId())
				s.Error(err, "event listener should no longer exist")
				_, err = s.app.TriggerKeeper.GetTrigger(s.ctx, tc.request.GetId())
				s.Error(err, "trigger should no longer exist")
				s.Panics(func() {
					s.app.TriggerKeeper.GetGasLimit(s.ctx, tc.request.GetId())
				}, "gas limit should not exist")
			} else {
				s.EqualError(err, tc.err, "handler should throw error and match")
			}

		})
	}
}
