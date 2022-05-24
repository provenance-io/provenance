package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkTypes "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/reward/types"
)

/*func (suite *KeeperTestSuite) TestNewShare() {
	suite.SetupTest()

	time := timestamppb.Now().AsTime()
	share := types.NewShare(
		1,
		2,
		"test",
		true,
		time,
		5,
	)

	suite.Assert().Equal(uint64(1), share.GetRewardProgramId(), "reward program id must match")
	suite.Assert().Equal(uint64(2), share.GetEpochId(), "epoch id must match")
	suite.Assert().Equal("test", share.GetAddress(), "address must match")
	suite.Assert().Equal(true, share.GetClaimed(), "claim status must match")
	suite.Assert().Equal(time, share.GetExpireTime(), "expiration time must match")
	suite.Assert().Equal(int64(5), share.GetAmount(), "share amount must match")
}*/

func setupEventHistory(suite *KeeperTestSuite) {
	attributes1 := []sdkTypes.Attribute{
		sdkTypes.NewAttribute("key1", "value1"),
		sdkTypes.NewAttribute("key2", "value2"),
		sdkTypes.NewAttribute("key3", "value3"),
	}
	attributes2 := []sdkTypes.Attribute{
		sdkTypes.NewAttribute("key1", "value1"),
		sdkTypes.NewAttribute("key3", "value2"),
		sdkTypes.NewAttribute("key4", "value3"),
	}
	event1 := sdkTypes.NewEvent("event1", attributes1...)
	event2 := sdkTypes.NewEvent("event2", attributes2...)
	event3 := sdkTypes.NewEvent("event1", attributes1...)
	loggedEvents := sdkTypes.Events{
		event1,
		event2,
		event3,
	}
	eventManagerStub := sdkTypes.NewEventManagerWithHistory(loggedEvents.ToABCIEvents())
	suite.ctx = suite.ctx.WithEventManager(eventManagerStub)
}

func setupEventHistoryWithDelegates(suite *KeeperTestSuite) {
	attributes1 := []sdkTypes.Attribute{
		sdkTypes.NewAttribute("validator", "cosmosvaloper15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqh6tjun"),
		sdkTypes.NewAttribute("amount", "1000000000000000nhash"),
	}
	attributes2 := []sdkTypes.Attribute{
		sdkTypes.NewAttribute("module", "staking"),
		sdkTypes.NewAttribute("sender", "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h"),
	}
	attributes3 := []sdkTypes.Attribute{
		sdkTypes.NewAttribute("validator", "cosmosvaloper15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqh6tjun"),
		sdkTypes.NewAttribute("amount", "50000000000000nhash"),
		sdkTypes.NewAttribute("new_shares", "50000000000000.000000000000000000"),
	}
	event1 := sdkTypes.NewEvent("create_validator", attributes1...)
	event2 := sdkTypes.NewEvent("message", attributes2...)
	event3 := sdkTypes.NewEvent("delegate", attributes3...)
	event4 := sdkTypes.NewEvent("message", attributes2...)
	loggedEvents := sdkTypes.Events{
		event1,
		event2,
		event3,
		event4,
	}
	eventManagerStub := sdkTypes.NewEventManagerWithHistory(loggedEvents.ToABCIEvents())
	suite.ctx = suite.ctx.WithEventManager(eventManagerStub)
}

func (suite *KeeperTestSuite) TestIterateABCIEventsWildcard() {
	suite.SetupTest()
	setupEventHistory(suite)
	events := []types.ABCIEvent{}
	criteria := types.NewEventCriteria(events)
	counter := 0
	suite.app.RewardKeeper.IterateABCIEvents(suite.ctx, criteria, func(name string, attributes *map[string][]byte) error {
		counter += 1
		return nil
	})
	suite.Assert().Equal(3, counter, "should iterate for each event")
}

func (suite *KeeperTestSuite) TestIterateABCIEventsByEventType() {
	suite.SetupTest()
	setupEventHistory(suite)
	counter := 0
	events := []types.ABCIEvent{
		{
			Type: "event1",
		},
	}
	criteria := types.NewEventCriteria(events)
	suite.app.RewardKeeper.IterateABCIEvents(suite.ctx, criteria, func(name string, attributes *map[string][]byte) error {
		counter += 1
		return nil
	})
	suite.Assert().Equal(2, counter, "should iterate only for specified event types")
}

func (suite *KeeperTestSuite) TestIterateABCIEventsByEventTypeAndAttributeName() {
	suite.SetupTest()
	setupEventHistory(suite)
	counter := 0
	events := []types.ABCIEvent{
		{
			Type: "event1",
			Attributes: map[string][]byte{
				"key1": nil,
			},
		},
	}
	criteria := types.NewEventCriteria(events)
	suite.app.RewardKeeper.IterateABCIEvents(suite.ctx, criteria, func(name string, attributes *map[string][]byte) error {
		counter += 1
		return nil
	})
	suite.Assert().Equal(2, counter, "should iterate only for specified event types with matching attributes")
}

func (suite *KeeperTestSuite) TestIterateABCIEventsByEventTypeAndAttributeNameAndValue() {
	suite.SetupTest()
	setupEventHistory(suite)
	counter := 0
	events := []types.ABCIEvent{
		{
			Type: "event1",
			Attributes: map[string][]byte{
				"key1": []byte("value1"),
			},
		},
	}
	criteria := types.NewEventCriteria(events)
	suite.app.RewardKeeper.IterateABCIEvents(suite.ctx, criteria, func(name string, attributes *map[string][]byte) error {
		counter += 1
		return nil
	})
	suite.Assert().Equal(2, counter, "should iterate only for specified event types with matching attributes")
}

func (suite *KeeperTestSuite) TestIterateABCIEventsNonExistantEventType() {
	suite.SetupTest()
	setupEventHistory(suite)
	counter := 0
	events := []types.ABCIEvent{
		{
			Type:       "event5",
			Attributes: map[string][]byte{},
		},
	}
	criteria := types.NewEventCriteria(events)
	suite.app.RewardKeeper.IterateABCIEvents(suite.ctx, criteria, func(name string, attributes *map[string][]byte) error {
		counter += 1
		return nil
	})
	suite.Assert().Equal(0, counter, "should not iterate if event doesn't exist")
}

func (suite *KeeperTestSuite) TestIterateABCIEventsNonExistantAttributeName() {
	suite.SetupTest()
	setupEventHistory(suite)
	counter := 0
	events := []types.ABCIEvent{
		{
			Type: "event1",
			Attributes: map[string][]byte{
				"blah": []byte("value5"),
			},
		},
	}
	criteria := types.NewEventCriteria(events)
	suite.app.RewardKeeper.IterateABCIEvents(suite.ctx, criteria, func(name string, attributes *map[string][]byte) error {
		counter += 1
		return nil
	})
	suite.Assert().Equal(0, counter, "should not iterate if attribute doesn't match")
}

func (suite *KeeperTestSuite) TestIterateABCIEventsNonAttributeValueMatch() {
	suite.SetupTest()
	setupEventHistory(suite)
	counter := 0
	events := []types.ABCIEvent{
		{
			Type: "event1",
			Attributes: map[string][]byte{
				"key1": []byte("value5"),
			},
		},
	}
	criteria := types.NewEventCriteria(events)
	suite.app.RewardKeeper.IterateABCIEvents(suite.ctx, criteria, func(name string, attributes *map[string][]byte) error {
		counter += 1
		return nil
	})
	suite.Assert().Equal(0, counter, "should not iterate if attribute doesn't match")
}

func (suite *KeeperTestSuite) TestGetMatchingEventsWithDelegates() {
	suite.SetupTest()
	setupEventHistoryWithDelegates(suite)
	criteria := types.NewEventCriteria([]types.ABCIEvent{
		{
			Type: "message",
			Attributes: map[string][]byte{
				"module": []byte("staking"),
			},
		},
		{
			Type:       "delegate",
			Attributes: map[string][]byte{},
		},
		{
			Type:       "create_validator",
			Attributes: map[string][]byte{},
		},
	})

	events, err := suite.app.RewardKeeper.GetMatchingEvents(suite.ctx, criteria)
	suite.Assert().NoError(err, "should throw no error when handling no events")
	suite.Assert().Equal(2, len(events), "should find the two delegate events")
	for _, event := range events {
		suite.Assert().Equal(event.Shares, int64(1), "shares must be 1")
		suite.Assert().Equal(event.Delegator.String(), "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", "delegator address must be correct")
		suite.Assert().Equal(event.Validator.String(), "cosmosvaloper15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqh6tjun", "validator address must be correct")
	}
}

func (suite *KeeperTestSuite) TestGetMatchingEventsWithoutDelegates() {
	suite.SetupTest()
	criteria := types.NewEventCriteria([]types.ABCIEvent{
		{
			Type: "message",
			Attributes: map[string][]byte{
				"module": []byte("staking"),
			},
		},
		{
			Type:       "delegate",
			Attributes: map[string][]byte{},
		},
		{
			Type:       "create_validator",
			Attributes: map[string][]byte{},
		},
	})
	events, err := suite.app.RewardKeeper.GetMatchingEvents(suite.ctx, criteria)
	suite.Assert().NoError(err, "should throw no error when handling no events")
	suite.Assert().Equal(0, len(events), "should have no events when no delegates are made")
}

// FindQualifyingActions

type MockAction struct {
	PassEvaluate bool
}

func (m MockAction) ActionType() string {
	return ""
}

func (m MockAction) Evaluate(ctx sdk.Context, provider types.KeeperProvider, state types.AccountState, event types.EvaluationResult) bool {
	return m.PassEvaluate
}

func (m MockAction) GetEventCriteria() *types.EventCriteria {
	return nil
}

func (suite *KeeperTestSuite) TestFindQualifyingActionsWithNoAbciEvents() {
	suite.SetupTest()
	program := types.RewardProgram{}
	action := MockAction{PassEvaluate: false}
	results := suite.app.RewardKeeper.FindQualifyingActions(suite.ctx, &program, action, []types.EvaluationResult{})
	suite.Assert().Equal(0, len(results), "should have no results for empty list of abci events")
}

func (suite *KeeperTestSuite) TestFindQualifyingActionsWithNoMatchingResults() {
	suite.SetupTest()
	program := types.RewardProgram{}
	action := MockAction{PassEvaluate: false}
	results := suite.app.RewardKeeper.FindQualifyingActions(suite.ctx, &program, action, []types.EvaluationResult{
		{},
		{},
	})
	suite.Assert().Equal(0, len(results), "should have empty lists when no results match the evaluation")
}

func (suite *KeeperTestSuite) TestFindQualifyingActionsWithMatchingResults() {
	suite.SetupTest()
	program := types.RewardProgram{}
	action := MockAction{PassEvaluate: true}
	results := suite.app.RewardKeeper.FindQualifyingActions(suite.ctx, &program, action, []types.EvaluationResult{
		{},
		{},
	})
	suite.Assert().Equal(2, len(results), "should have all results that evaluate to true")
}

func (suite *KeeperTestSuite) TestFindQualifyingActionsWithNil() {
	suite.SetupTest()
	results := suite.app.RewardKeeper.FindQualifyingActions(suite.ctx, nil, nil, nil)
	suite.Assert().Equal(0, len(results), "should handle nil input")
}
