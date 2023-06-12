package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/trigger/types"
)

func (s *KeeperTestSuite) TestCreateTrigger() {
	owner := []string{s.accountAddresses[0].String()}
	var event types.TriggerEventI = &types.BlockHeightEvent{BlockHeight: 130}
	action := types.MsgDestroyTriggerRequest{Id: 100, Authority: owner[0]}

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
		},
		{
			name:       "valid - second trigger has incremented id",
			request:    types.NewCreateTriggerRequest(owner, event, []sdk.Msg{&action}),
			expectedId: 2,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			em := sdk.NewEventManager()
			ctx := s.ctx.WithGasMeter(sdk.NewGasMeter(9999999999)).WithEventManager(em)
			response, err := s.msgServer.CreateTrigger(ctx, tc.request)
			s.ctx = s.ctx.WithGasMeter(sdk.NewGasMeter(9999999999))

			if len(tc.err) == 0 {
				s.NoError(err, "should not throw an error for handler")
				expResp := &types.MsgCreateTriggerResponse{Id: tc.expectedId}
				resultEvent, _ := sdk.TypedEventToEvent(&types.EventTriggerCreated{
					TriggerId: fmt.Sprintf("%d", tc.expectedId),
				})
				s.Equal(expResp, response, "CreateTrigger response")
				s.Equal(sdk.Events{resultEvent}, em.Events(), "should have correct events for CreateTrigger")
				_, err = s.app.TriggerKeeper.GetEventListener(s.ctx, event.GetEventPrefix(), event.GetEventOrder(), tc.expectedId)
				s.NoError(err, "should have event listener for successful CreateTrigger")
				_, err = s.app.TriggerKeeper.GetTrigger(s.ctx, tc.expectedId)
				s.NoError(err, "should have trigger for successful CreateTrigger")
				gasLimit := s.app.TriggerKeeper.GetGasLimit(s.ctx, tc.expectedId)
				s.Equal(uint64(2000000), gasLimit, "should have correct gas limit for successful CreateTrigger")
			} else {
				s.EqualError(err, tc.err, "should throw an error on invalid handler")
			}

		})
	}
}

func (s *KeeperTestSuite) TestDestroyTrigger() {
	owner := []string{s.accountAddresses[0].String()}
	owner2 := []string{s.accountAddresses[1].String()}
	var event types.TriggerEventI = &types.BlockHeightEvent{BlockHeight: 130}
	action := types.MsgDestroyTriggerRequest{Id: 100, Authority: owner[0]}

	setupRequests := []*types.MsgCreateTriggerRequest{
		types.NewCreateTriggerRequest(owner, event, []sdk.Msg{&action}),
		types.NewCreateTriggerRequest(owner, event, []sdk.Msg{&action}),
		types.NewCreateTriggerRequest(owner2, event, []sdk.Msg{&action}),
	}
	for i, request := range setupRequests {
		s.ctx = s.ctx.WithGasMeter(sdk.NewGasMeter(9999999999))
		_, err := s.msgServer.CreateTrigger(s.ctx, request)
		s.Require().NoError(err, "Setup[%d]: CreateTrigger", i)
	}

	tests := []struct {
		name    string
		request *types.MsgDestroyTriggerRequest
		err     string
	}{
		{
			name:    "valid - single trigger destroyed",
			request: types.NewDestroyTriggerRequest(owner[0], 1),
			err:     "",
		},
		{
			name:    "valid - multiple triggers destroyed",
			request: types.NewDestroyTriggerRequest(owner[0], 2),
			err:     "",
		},
		{
			name:    "invalid - destroy a non existant trigger",
			request: types.NewDestroyTriggerRequest(owner[0], 100),
			err:     "trigger not found",
		},
		{
			name:    "invalid - destroy a trigger that is not owned by the user",
			request: types.NewDestroyTriggerRequest(owner[0], 3),
			err:     "signer does not have authority to destroy trigger",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			em := sdk.NewEventManager()
			ctx := s.ctx.WithGasMeter(sdk.NewGasMeter(9999999999)).WithEventManager(em)
			_, err := s.msgServer.DestroyTrigger(ctx, tc.request)
			s.ctx = s.ctx.WithGasMeter(sdk.NewGasMeter(9999999999))

			if len(tc.err) == 0 {
				s.NoError(err, "should not throw an error on valid call to handler for TriggerDestroyRequest")
				resultEvent, _ := sdk.TypedEventToEvent(&types.EventTriggerDestroyed{
					TriggerId: fmt.Sprintf("%d", tc.request.GetId()),
				})
				s.Equal(sdk.Events{resultEvent}, em.Events(), "should have correct events for TriggerDestroyRequest")
				_, err = s.app.TriggerKeeper.GetEventListener(s.ctx, event.GetEventPrefix(), event.GetEventOrder(), tc.request.GetId())
				s.Error(err, "should not have an event listener after handling TriggerDestroyRequest")
				_, err = s.app.TriggerKeeper.GetTrigger(s.ctx, tc.request.GetId())
				s.Error(err, "should not have a trigger after handling TriggerDestroyRequest")
				s.PanicsWithValue("gas limit not found for trigger", func() {
					s.app.TriggerKeeper.GetGasLimit(s.ctx, tc.request.GetId())
				})
			} else {
				s.EqualError(err, tc.err, "handler should throw error and match")
			}

		})
	}
}
