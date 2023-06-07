package keeper_test

import (
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/provenance-io/provenance/x/trigger/types"
)

func (s *KeeperTestSuite) TestTriggerByID() {
	queryClient := s.queryClient
	owner := s.accountAddresses[0].String()
	trigger1 := s.CreateTrigger(1, owner, &types.BlockHeightEvent{BlockHeight: uint64(s.ctx.BlockHeight())}, &types.MsgDestroyTriggerRequest{Id: 100, Authority: owner})
	trigger2 := s.CreateTrigger(2, owner, &types.BlockHeightEvent{BlockHeight: 130}, &types.MsgDestroyTriggerRequest{Id: 101, Authority: owner})

	tests := []struct {
		name      string
		triggers  []types.Trigger
		expected  types.Trigger
		triggerID uint64
		err       string
	}{
		{
			name:      "invalid - test non existing",
			triggers:  []types.Trigger{},
			expected:  types.Trigger{},
			triggerID: 5,
			err:       "trigger not found",
		},
		{
			name:      "valid - test single",
			triggers:  []types.Trigger{trigger1},
			expected:  trigger1,
			triggerID: 1,
			err:       "",
		},
		{
			name:      "valid - test one of many",
			triggers:  []types.Trigger{trigger1, trigger2},
			expected:  trigger2,
			triggerID: 2,
			err:       "",
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			// Setup
			for _, trigger := range tc.triggers {
				s.app.TriggerKeeper.SetTrigger(s.ctx, trigger)
			}

			request := types.QueryTriggerByIDRequest{Id: tc.triggerID}
			response, err := queryClient.TriggerByID(s.ctx.Context(), &request)
			if len(tc.err) > 0 {
				s.EqualError(err, tc.err)
			} else {
				s.Assert().Nil(err)
				s.Assert().Equal(tc.expected.Id, response.Trigger.Id)
			}

			// Destroy
			for _, trigger := range tc.triggers {
				s.app.TriggerKeeper.RemoveTrigger(s.ctx, trigger.GetId())
			}
		})
	}
}

func (s *KeeperTestSuite) TestTriggers() {
	queryClient := s.queryClient
	pageRequest := &query.PageRequest{}
	pageRequest.Limit = 100
	pageRequest.CountTotal = true
	owner := s.accountAddresses[0].String()
	trigger1 := s.CreateTrigger(1, owner, &types.BlockHeightEvent{BlockHeight: uint64(s.ctx.BlockHeight())}, &types.MsgDestroyTriggerRequest{Id: 100, Authority: owner})
	trigger2 := s.CreateTrigger(2, owner, &types.BlockHeightEvent{BlockHeight: 130}, &types.MsgDestroyTriggerRequest{Id: 101, Authority: owner})

	tests := []struct {
		name        string
		triggers    []types.Trigger
		expected    []types.Trigger
		pageRequest *query.PageRequest
	}{
		{
			name:        "valid - test empty",
			triggers:    []types.Trigger{},
			expected:    []types.Trigger(nil),
			pageRequest: nil,
		},
		{
			name: "valid - test single",
			triggers: []types.Trigger{
				trigger1,
			},
			expected: []types.Trigger{
				trigger1,
			},
			pageRequest: nil,
		},
		{
			name: "valid - test multiple",
			triggers: []types.Trigger{
				trigger1,
				trigger2,
			},
			expected: []types.Trigger{
				trigger1,
				trigger2,
			},
			pageRequest: nil,
		},
		{
			name: "valid - test pagination complete set",
			triggers: []types.Trigger{
				trigger1,
				trigger2,
			},
			expected: []types.Trigger{
				trigger1,
				trigger2,
			},
			pageRequest: &query.PageRequest{Limit: 100, CountTotal: true},
		},
		{
			name: "valid - test pagination partial set",
			triggers: []types.Trigger{
				trigger1,
				trigger2,
			},
			expected: []types.Trigger{
				trigger1,
			},
			pageRequest: &query.PageRequest{Limit: 1, CountTotal: true},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			// Setup
			for _, trigger := range tc.triggers {
				s.app.TriggerKeeper.SetTrigger(s.ctx, trigger)
			}

			request := types.QueryTriggersRequest{}
			if tc.pageRequest != nil {
				request.Pagination = tc.pageRequest
			}
			response, err := queryClient.Triggers(s.ctx.Context(), &request)
			s.Assert().Nil(err, "query should not error")
			s.Assert().Equal(len(tc.expected), len(response.Triggers))

			// Destroy
			for _, trigger := range tc.triggers {
				s.app.TriggerKeeper.RemoveTrigger(s.ctx, trigger.GetId())
			}
		})
	}
}
