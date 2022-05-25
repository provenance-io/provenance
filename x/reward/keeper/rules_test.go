package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkTypes "github.com/cosmos/cosmos-sdk/types"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/provenance-io/provenance/x/reward/types"
)

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

// Test ActionDelegate Evaluate

type MockKeeperProvider struct {
}

func (m MockKeeperProvider) GetDistributionKeeper() types.DistributionKeeper {
	return nil
}

func (m MockKeeperProvider) GetStakingKeeper() types.StakingKeeper {
	return MockStakingKeeper{}
}

type MockStakingKeeper struct {
}

func (m MockStakingKeeper) GetAllDelegatorDelegations(ctx sdk.Context, delegator sdk.AccAddress) []stakingtypes.Delegation {
	return nil
}

func (m MockStakingKeeper) GetValidatorDelegations(ctx sdk.Context, valAddr sdk.ValAddress) (delegations []stakingtypes.Delegation) {
	delegations = []stakingtypes.Delegation{
		{
			DelegatorAddress: "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
			ValidatorAddress: "cosmosvaloper15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqh6tjun",
			Shares:           sdkTypes.NewDec(1),
		},
		{
			DelegatorAddress: "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
			ValidatorAddress: "cosmosvaloper15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqh6tjun",
			Shares:           sdkTypes.NewDec(1),
		},
		{
			DelegatorAddress: "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
			ValidatorAddress: "cosmosvaloper15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqh6tjun",
			Shares:           sdkTypes.NewDec(1),
		},
	}

	return delegations
}

func (suite *KeeperTestSuite) TestActionDelegateEvaluatePasses() {
	suite.SetupTest()

	action := types.NewActionDelegate()
	action.MinimumActions = 1
	action.MaximumActions = 2
	action.MinimumDelegationAmount = sdkTypes.NewDec(2).BigInt().Uint64()
	action.MaximumDelegationAmount = sdkTypes.NewDec(10).BigInt().Uint64()
	action.MinimumDelegationPercentage = .5
	action.MaximumDelegationPercentage = 1.0

	keeperProvider := MockKeeperProvider{}
	state := types.NewAccountState(0, 0, "")
	state.ActionCounter += 1

	validator, _ := sdk.ValAddressFromBech32("cosmosvaloper15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqh6tjun")
	delegator, _ := sdk.AccAddressFromBech32("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h")
	event := types.EvaluationResult{
		Validator: validator,
		Delegator: delegator,
	}

	passed := action.Evaluate(suite.ctx, keeperProvider, state, event)
	suite.Assert().True(passed, "evaluate should pass when criteria are met")
}

func (suite *KeeperTestSuite) TestActionDelegateEvaluateFailsWhenMinimumActionsNotMet() {
	suite.SetupTest()

	action := types.NewActionDelegate()
	action.MinimumActions = 1
	action.MaximumActions = 2
	action.MinimumDelegationAmount = sdkTypes.NewDec(2).BigInt().Uint64()
	action.MaximumDelegationAmount = sdkTypes.NewDec(10).BigInt().Uint64()
	action.MinimumDelegationPercentage = .5
	action.MaximumDelegationPercentage = 1.0

	keeperProvider := MockKeeperProvider{}
	state := types.NewAccountState(0, 0, "")

	validator, _ := sdk.ValAddressFromBech32("cosmosvaloper15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqh6tjun")
	delegator, _ := sdk.AccAddressFromBech32("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h")
	event := types.EvaluationResult{
		Validator: validator,
		Delegator: delegator,
	}

	passed := action.Evaluate(suite.ctx, keeperProvider, state, event)
	suite.Assert().False(passed, "test should fail when minimum actions not met")
}

func (suite *KeeperTestSuite) TestActionDelegateEvaluateFailsWhenMaximumActionsNotMet() {
	suite.SetupTest()

	action := types.NewActionDelegate()
	action.MinimumActions = 1
	action.MaximumActions = 2
	action.MinimumDelegationAmount = sdkTypes.NewDec(2).BigInt().Uint64()
	action.MaximumDelegationAmount = sdkTypes.NewDec(10).BigInt().Uint64()
	action.MinimumDelegationPercentage = .5
	action.MaximumDelegationPercentage = 1.0

	keeperProvider := MockKeeperProvider{}
	state := types.NewAccountState(0, 0, "")
	state.ActionCounter += 3

	validator, _ := sdk.ValAddressFromBech32("cosmosvaloper15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqh6tjun")
	delegator, _ := sdk.AccAddressFromBech32("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h")
	event := types.EvaluationResult{
		Validator: validator,
		Delegator: delegator,
	}

	passed := action.Evaluate(suite.ctx, keeperProvider, state, event)
	suite.Assert().False(passed, "test should fail when maximum actions not met")
}

func (suite *KeeperTestSuite) TestActionDelegateEvaluateFailsWhenMaximumDelegationAmountNotMet() {
	suite.SetupTest()

	action := types.NewActionDelegate()
	action.MinimumActions = 1
	action.MaximumActions = 2
	action.MinimumDelegationAmount = sdkTypes.NewDec(1).BigInt().Uint64()
	action.MaximumDelegationAmount = sdkTypes.NewDec(1).BigInt().Uint64()
	action.MinimumDelegationPercentage = .5
	action.MaximumDelegationPercentage = 1.0

	keeperProvider := MockKeeperProvider{}
	state := types.NewAccountState(0, 0, "")
	state.ActionCounter += 1

	validator, _ := sdk.ValAddressFromBech32("cosmosvaloper15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqh6tjun")
	delegator, _ := sdk.AccAddressFromBech32("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h")
	event := types.EvaluationResult{
		Validator: validator,
		Delegator: delegator,
	}

	passed := action.Evaluate(suite.ctx, keeperProvider, state, event)
	suite.Assert().False(passed, "test should fail when maximum delegation amount not met")
}

func (suite *KeeperTestSuite) TestActionDelegateEvaluateFailsWhenMinimumDelegationAmountNotMet() {
	suite.SetupTest()

	action := types.NewActionDelegate()
	action.MinimumActions = 1
	action.MaximumActions = 2
	action.MinimumDelegationAmount = sdkTypes.NewDec(5).BigInt().Uint64()
	action.MaximumDelegationAmount = sdkTypes.NewDec(10).BigInt().Uint64()
	action.MinimumDelegationPercentage = .5
	action.MaximumDelegationPercentage = 1.0

	keeperProvider := MockKeeperProvider{}
	state := types.NewAccountState(0, 0, "")
	state.ActionCounter += 1

	validator, _ := sdk.ValAddressFromBech32("cosmosvaloper15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqh6tjun")
	delegator, _ := sdk.AccAddressFromBech32("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h")
	event := types.EvaluationResult{
		Validator: validator,
		Delegator: delegator,
	}

	passed := action.Evaluate(suite.ctx, keeperProvider, state, event)
	suite.Assert().False(passed, "test should fail when minimum delegation amount not met")
}

func (suite *KeeperTestSuite) TestActionDelegateEvaluateFailsWhenMinimumDelegationPercentageNotMet() {
	suite.SetupTest()

	action := types.NewActionDelegate()
	action.MinimumActions = 1
	action.MaximumActions = 2
	action.MinimumDelegationAmount = sdkTypes.NewDec(1).BigInt().Uint64()
	action.MaximumDelegationAmount = sdkTypes.NewDec(10).BigInt().Uint64()
	action.MinimumDelegationPercentage = 1.1
	action.MaximumDelegationPercentage = 1.0

	keeperProvider := MockKeeperProvider{}
	state := types.NewAccountState(0, 0, "")
	state.ActionCounter += 1

	validator, _ := sdk.ValAddressFromBech32("cosmosvaloper15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqh6tjun")
	delegator, _ := sdk.AccAddressFromBech32("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h")
	event := types.EvaluationResult{
		Validator: validator,
		Delegator: delegator,
	}

	passed := action.Evaluate(suite.ctx, keeperProvider, state, event)
	suite.Assert().False(passed, "test should fail when minimum delegation percentage not met")
}

func (suite *KeeperTestSuite) TestActionDelegateEvaluateFailsWhenMaximumDelegationPercentageNotMet() {
	suite.SetupTest()

	action := types.NewActionDelegate()
	action.MinimumActions = 1
	action.MaximumActions = 2
	action.MinimumDelegationAmount = sdkTypes.NewDec(1).BigInt().Uint64()
	action.MaximumDelegationAmount = sdkTypes.NewDec(10).BigInt().Uint64()
	action.MinimumDelegationPercentage = 0.0
	action.MaximumDelegationPercentage = 0.5

	keeperProvider := MockKeeperProvider{}
	state := types.NewAccountState(0, 0, "")
	state.ActionCounter += 1

	validator, _ := sdk.ValAddressFromBech32("cosmosvaloper15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqh6tjun")
	delegator, _ := sdk.AccAddressFromBech32("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h")
	event := types.EvaluationResult{
		Validator: validator,
		Delegator: delegator,
	}

	passed := action.Evaluate(suite.ctx, keeperProvider, state, event)
	suite.Assert().False(passed, "test should fail when maximum delegation percentage not met")
}

// Test GetRewardAction

func (suite *KeeperTestSuite) TestGetRewardActionHandlesUnsupportedActions() {
	suite.SetupTest()
	qa := types.QualifyingAction{}
	_, err := qa.GetRewardAction(suite.ctx)
	suite.Assert().Error(err, "should throw error when an action is not supported")
}

func (suite *KeeperTestSuite) TestGetRewardActionHandlesActionDelegate() {
	suite.SetupTest()
	delegate := types.QualifyingAction_Delegate{}
	qa := types.QualifyingAction{
		Type: &delegate,
	}
	action, err := qa.GetRewardAction(suite.ctx)
	suite.Assert().NoError(err, "should not throw error when action is supported")
	suite.Assert().Equal(types.ActionTypeDelegate, action.ActionType(), "should return the correct action type")
}

// Test DetectQualifyingActions
func (suite *KeeperTestSuite) TestDetectQualifyingActionsWith1QualifyingAction() {
	suite.SetupTest()
	setupEventHistoryWithDelegates(suite)
	suite.app.RewardKeeper.SetStakingKeeper(MockStakingKeeper{})

	rewardProgram := types.NewRewardProgram(
		"title",
		"description",
		1,
		"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
		sdkTypes.NewInt64Coin("hotdog", 10000),
		sdkTypes.NewInt64Coin("hotdog", 10000),
		time.Now(),
		5,
		5,
		types.NewEligibilityCriteria("reward-action", &types.ActionDelegate{}),
	)
	rewardProgram.QualifyingActions = append(rewardProgram.QualifyingActions, types.QualifyingAction{
		Type: &types.QualifyingAction_Delegate{
			Delegate: &types.ActionDelegate{
				MinimumActions:              0,
				MaximumActions:              1,
				MinimumDelegationAmount:     sdkTypes.NewDec(0).BigInt().Uint64(),
				MaximumDelegationAmount:     sdkTypes.NewDec(10).BigInt().Uint64(),
				MinimumDelegationPercentage: 0,
				MaximumDelegationPercentage: 1,
			},
		},
	})
	qualifyingActions, err := suite.app.RewardKeeper.DetectQualifyingActions(suite.ctx, &rewardProgram)
	suite.Assert().NoError(err, "must not error")
	suite.Assert().Equal(2, len(qualifyingActions), "must find two qualifying actions")
}

func (suite *KeeperTestSuite) TestDetectQualifyingActionsWith2QualifyingAction() {
	suite.SetupTest()
	setupEventHistoryWithDelegates(suite)
	suite.app.RewardKeeper.SetStakingKeeper(MockStakingKeeper{})

	rewardProgram := types.NewRewardProgram(
		"title",
		"description",
		1,
		"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
		sdkTypes.NewInt64Coin("hotdog", 10000),
		sdkTypes.NewInt64Coin("hotdog", 10000),
		time.Now(),
		5,
		5,
		types.NewEligibilityCriteria("reward-action", &types.ActionDelegate{}),
	)
	rewardProgram.QualifyingActions = append(rewardProgram.QualifyingActions,
		types.QualifyingAction{
			Type: &types.QualifyingAction_Delegate{
				Delegate: &types.ActionDelegate{
					MinimumActions:              0,
					MaximumActions:              1,
					MinimumDelegationAmount:     sdkTypes.NewDec(0).BigInt().Uint64(),
					MaximumDelegationAmount:     sdkTypes.NewDec(10).BigInt().Uint64(),
					MinimumDelegationPercentage: 0,
					MaximumDelegationPercentage: 1,
				},
			},
		},
		types.QualifyingAction{
			Type: &types.QualifyingAction_Delegate{
				Delegate: &types.ActionDelegate{
					MinimumActions:              0,
					MaximumActions:              1,
					MinimumDelegationAmount:     sdkTypes.NewDec(0).BigInt().Uint64(),
					MaximumDelegationAmount:     sdkTypes.NewDec(10).BigInt().Uint64(),
					MinimumDelegationPercentage: 0,
					MaximumDelegationPercentage: 1,
				},
			},
		})
	qualifyingActions, err := suite.app.RewardKeeper.DetectQualifyingActions(suite.ctx, &rewardProgram)
	suite.Assert().NoError(err, "must not error")
	suite.Assert().Equal(4, len(qualifyingActions), "must find two qualifying actions")
}

func (suite *KeeperTestSuite) TestDetectQualifyingActionsWithNoQualifyingAction() {
	suite.SetupTest()
	setupEventHistoryWithDelegates(suite)
	suite.app.RewardKeeper.SetStakingKeeper(MockStakingKeeper{})

	rewardProgram := types.NewRewardProgram(
		"title",
		"description",
		1,
		"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
		sdkTypes.NewInt64Coin("hotdog", 10000),
		sdkTypes.NewInt64Coin("hotdog", 10000),
		time.Now(),
		5,
		5,
		types.NewEligibilityCriteria("reward-action", &types.ActionDelegate{}),
	)

	qualifyingActions, err := suite.app.RewardKeeper.DetectQualifyingActions(suite.ctx, &rewardProgram)
	suite.Assert().NoError(err, "must not error")
	suite.Assert().Equal(0, len(qualifyingActions), "must find two qualifying actions")
}

func (suite *KeeperTestSuite) TestDetectQualifyingActionsWithNoMatchingQualifyingAction() {
	suite.SetupTest()
	setupEventHistoryWithDelegates(suite)
	suite.app.RewardKeeper.SetStakingKeeper(MockStakingKeeper{})

	rewardProgram := types.NewRewardProgram(
		"title",
		"description",
		1,
		"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
		sdkTypes.NewInt64Coin("hotdog", 10000),
		sdkTypes.NewInt64Coin("hotdog", 10000),
		time.Now(),
		5,
		5,
		types.NewEligibilityCriteria("reward-action", &types.ActionDelegate{}),
	)
	rewardProgram.QualifyingActions = append(rewardProgram.QualifyingActions, types.QualifyingAction{
		Type: &types.QualifyingAction_Delegate{
			Delegate: &types.ActionDelegate{
				MinimumActions:              1000,
				MaximumActions:              1000,
				MinimumDelegationAmount:     sdkTypes.NewDec(0).BigInt().Uint64(),
				MaximumDelegationAmount:     sdkTypes.NewDec(10).BigInt().Uint64(),
				MinimumDelegationPercentage: 0,
				MaximumDelegationPercentage: 1,
			},
		},
	})
	qualifyingActions, err := suite.app.RewardKeeper.DetectQualifyingActions(suite.ctx, &rewardProgram)
	suite.Assert().NoError(err, "must not error")
	suite.Assert().Equal(0, len(qualifyingActions), "must find two qualifying actions")
}

// Test RewardShares
func (suite *KeeperTestSuite) TestRewardSharesSingle() {
	suite.SetupTest()

	rewardProgram := types.NewRewardProgram(
		"title",
		"description",
		1,
		"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
		sdkTypes.NewInt64Coin("hotdog", 10000),
		sdkTypes.NewInt64Coin("hotdog", 10000),
		time.Now(),
		5,
		5,
		types.NewEligibilityCriteria("reward-action", &types.ActionDelegate{}),
	)

	validator, _ := sdk.ValAddressFromBech32("cosmosvaloper15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqh6tjun")
	delegator, _ := sdk.AccAddressFromBech32("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h")
	results := []types.EvaluationResult{
		{
			EventTypeToSearch: "delegate",
			AttributeKey:      "attribute",
			Shares:            1,
			Address:           delegator,
			Validator:         validator,
			Delegator:         delegator,
		},
	}

	err := suite.app.RewardKeeper.RewardShares(suite.ctx, &rewardProgram, results)
	share, _ := suite.app.RewardKeeper.GetShare(suite.ctx, rewardProgram.GetId(), rewardProgram.GetCurrentEpoch(), delegator.String())
	suite.Assert().NoError(err, "should return no error on success")
	suite.Assert().Equal(int64(1), share.Amount, "share amount should increment")
	suite.Assert().Equal(rewardProgram.GetId(), share.GetRewardProgramId(), "reward program id should match")
	suite.Assert().Equal(rewardProgram.GetCurrentEpoch(), share.GetEpochId(), "epoch id should match")
	suite.Assert().Equal(delegator.String(), share.GetAddress(), "address should match delegator")
	suite.Assert().Equal(false, share.GetClaimed(), "claimed should be set to false")
	suite.Assert().Equal(rewardProgram.GetEpochEndTime().Add(time.Duration(rewardProgram.GetShareExpirationOffset())), share.GetExpireTime(), "expiration time should match epoch end time + offset")
}

func (suite *KeeperTestSuite) TestRewardSharesMultiple() {
	suite.SetupTest()

	rewardProgram := types.NewRewardProgram(
		"title",
		"description",
		1,
		"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
		sdkTypes.NewInt64Coin("hotdog", 10000),
		sdkTypes.NewInt64Coin("hotdog", 10000),
		time.Now(),
		5,
		5,
		types.NewEligibilityCriteria("reward-action", &types.ActionDelegate{}),
	)

	validator, _ := sdk.ValAddressFromBech32("cosmosvaloper15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqh6tjun")
	delegator, _ := sdk.AccAddressFromBech32("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h")
	results := []types.EvaluationResult{
		{
			EventTypeToSearch: "delegate",
			AttributeKey:      "attribute",
			Shares:            1,
			Address:           delegator,
			Validator:         validator,
			Delegator:         delegator,
		},
		{
			EventTypeToSearch: "delegate",
			AttributeKey:      "attribute",
			Shares:            1,
			Address:           delegator,
			Validator:         validator,
			Delegator:         delegator,
		},
	}

	err := suite.app.RewardKeeper.RewardShares(suite.ctx, &rewardProgram, results)
	share, _ := suite.app.RewardKeeper.GetShare(suite.ctx, rewardProgram.GetId(), rewardProgram.GetCurrentEpoch(), delegator.String())
	suite.Assert().NoError(err, "should return no error on success")
	suite.Assert().Equal(int64(2), share.Amount, "share amount should increment")
	suite.Assert().Equal(rewardProgram.GetId(), share.GetRewardProgramId(), "reward program id should match")
	suite.Assert().Equal(rewardProgram.GetCurrentEpoch(), share.GetEpochId(), "epoch id should match")
	suite.Assert().Equal(delegator.String(), share.GetAddress(), "address should match delegator")
	suite.Assert().Equal(false, share.GetClaimed(), "claimed should be set to false")
	suite.Assert().Equal(rewardProgram.GetEpochEndTime().Add(time.Duration(rewardProgram.GetShareExpirationOffset())), share.GetExpireTime(), "expiration time should match epoch end time + offset")
}

func (suite *KeeperTestSuite) TestRewardSharesInvalidRewardProgram() {
	suite.SetupTest()

	rewardProgram := types.NewRewardProgram(
		"title",
		"description",
		0,
		"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
		sdkTypes.NewInt64Coin("hotdog", 10000),
		sdkTypes.NewInt64Coin("hotdog", 10000),
		time.Now(),
		5,
		5,
		types.NewEligibilityCriteria("reward-action", &types.ActionDelegate{}),
	)

	validator, _ := sdk.ValAddressFromBech32("cosmosvaloper15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqh6tjun")
	delegator, _ := sdk.AccAddressFromBech32("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h")
	results := []types.EvaluationResult{
		{
			EventTypeToSearch: "delegate",
			AttributeKey:      "attribute",
			Shares:            1,
			Address:           delegator,
			Validator:         validator,
			Delegator:         delegator,
		},
	}

	err := suite.app.RewardKeeper.RewardShares(suite.ctx, &rewardProgram, results)
	share, _ := suite.app.RewardKeeper.GetShare(suite.ctx, rewardProgram.GetId(), rewardProgram.GetCurrentEpoch(), delegator.String())
	suite.Assert().Error(err, "should return an error on invalid program")
	suite.Assert().Equal(int64(0), share.Amount, "share amount should not increment")
}

func (suite *KeeperTestSuite) TestRewardSharesInvalidAddress() {
	suite.SetupTest()

	rewardProgram := types.NewRewardProgram(
		"title",
		"description",
		1,
		"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
		sdkTypes.NewInt64Coin("hotdog", 10000),
		sdkTypes.NewInt64Coin("hotdog", 10000),
		time.Now(),
		5,
		5,
		types.NewEligibilityCriteria("reward-action", &types.ActionDelegate{}),
	)

	validator, _ := sdk.ValAddressFromBech32("blah")
	delegator, _ := sdk.AccAddressFromBech32("blah")
	results := []types.EvaluationResult{
		{
			EventTypeToSearch: "delegate",
			AttributeKey:      "attribute",
			Shares:            1,
			Address:           delegator,
			Validator:         validator,
			Delegator:         delegator,
		},
	}

	err := suite.app.RewardKeeper.RewardShares(suite.ctx, &rewardProgram, results)
	share, _ := suite.app.RewardKeeper.GetShare(suite.ctx, rewardProgram.GetId(), rewardProgram.GetCurrentEpoch(), delegator.String())
	suite.Assert().NoError(err, "should return no error on invalid address")
	suite.Assert().Equal(int64(1), share.Amount, "share amount should not increment")
}
