package keeper_test

import (
	"context"
	"errors"
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	teststaking "github.com/cosmos/cosmos-sdk/x/staking/testutil"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/reward/types"
)

const delegatorAddr = "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h"

var (
	testValidators = []stakingtypes.Validator{}
)

func setupEventHistory(s *KeeperTestSuite) {
	attributes1 := []sdk.Attribute{
		sdk.NewAttribute("key1", "value1"),
		sdk.NewAttribute("key2", "value2"),
		sdk.NewAttribute("key3", "value3"),
	}
	attributes2 := []sdk.Attribute{
		sdk.NewAttribute("key1", "value1"),
		sdk.NewAttribute("key3", "value2"),
		sdk.NewAttribute("key4", "value3"),
	}
	event1 := sdk.NewEvent("event1", attributes1...)
	event2 := sdk.NewEvent("event2", attributes2...)
	event3 := sdk.NewEvent("event1", attributes1...)
	loggedEvents := sdk.Events{
		event1,
		event2,
		event3,
	}
	SetupEventHistory(s, loggedEvents)
}

func SetupEventHistory(s *KeeperTestSuite, events sdk.Events) {
	// TODO[1760]: event-history: Put this back once the event history stuff is back in the SDK.
	// eventManagerStub := sdk.NewEventManagerWithHistory(events.ToABCIEvents())
	// s.ctx = s.ctx.WithEventManager(eventManagerStub)
}

// with delegate
func SetupEventHistoryWithDelegates(s *KeeperTestSuite) {
	attributes1 := []sdk.Attribute{
		sdk.NewAttribute("validator", "cosmosvaloper15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqh6tjun"),
		sdk.NewAttribute("amount", "1000000000000000nhash"),
	}
	attributes2 := []sdk.Attribute{
		sdk.NewAttribute("module", "staking"),
		sdk.NewAttribute("sender", delegatorAddr),
	}
	attributes3 := []sdk.Attribute{
		sdk.NewAttribute("validator", "cosmosvaloper15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqh6tjun"),
		sdk.NewAttribute("amount", "50000000000000nhash"),
		sdk.NewAttribute("new_shares", "50000000000000.000000000000000000"),
	}
	event1 := sdk.NewEvent("create_validator", attributes1...)
	event2 := sdk.NewEvent("message", attributes2...)
	event3 := sdk.NewEvent("delegate", attributes3...)
	event4 := sdk.NewEvent("message", attributes2...)
	loggedEvents := sdk.Events{
		event1,
		event2,
		event3,
		event4,
	}
	SetupEventHistory(s, loggedEvents)
}

func (s *KeeperTestSuite) TestProcessTransactions() {
	now := time.Unix(1681964400, 0) // roughly Provenance Blockchain's 2-year anniversary.
	curHeader := s.ctx.BlockHeader()
	curHeader.Time = now
	s.ctx = s.ctx.WithBlockHeader(curHeader)
	s.app.RewardKeeper.SetStakingKeeper(MockStakingKeeper{})

	// Create a reward program
	s.Require().NoError(testutil.FundAccount(s.ctx, s.app.BankKeeper, s.accountAddr,
		sdk.NewCoins(sdk.NewInt64Coin("nhash", 1000000))), "funding accountAddr")
	minDel := sdk.NewInt64Coin("nhash", 1)
	maxDel := sdk.NewInt64Coin("nhash", 40)
	program := types.NewRewardProgram(
		"title",
		"description",
		1,
		s.accountAddr.String(),
		sdk.NewInt64Coin("nhash", 100000),
		sdk.NewInt64Coin("nhash", 1000),
		now.Add(1*time.Second),
		60*60,
		3,
		0,
		0,
		[]types.QualifyingAction{
			{
				Type: &types.QualifyingAction_Delegate{
					Delegate: &types.ActionDelegate{
						MinimumActions:               0,
						MaximumActions:               1,
						MinimumDelegationAmount:      &minDel,
						MaximumDelegationAmount:      &maxDel,
						MinimumActiveStakePercentile: sdkmath.LegacyNewDecWithPrec(0, 0),
						MaximumActiveStakePercentile: sdkmath.LegacyNewDecWithPrec(1, 0),
					},
				},
			},
		},
	)
	s.Require().NoError(s.app.RewardKeeper.CreateRewardProgram(s.ctx, program), "CreateRewardProgram")

	// Start the reward program.
	curHeader = s.ctx.BlockHeader()
	curHeader.Time = program.ProgramStartTime.Add(1 * time.Second)
	s.ctx = s.ctx.WithBlockHeader(curHeader)
	s.app.RewardKeeper.UpdateUnexpiredRewardsProgram(s.ctx)

	// make sure the reward program has started.
	var err error
	program, err = s.app.RewardKeeper.GetRewardProgram(s.ctx, program.GetId())
	s.Require().NoError(err, "GetRewardProgram")
	s.Require().Equal(types.RewardProgram_STATE_STARTED.String(), program.State.String(), "program state")

	// Fake some events and process them.
	SetupEventHistoryWithDelegates(s)
	s.Require().NotPanics(func() {
		s.app.RewardKeeper.ProcessTransactions(s.ctx)
	}, "ProcessTransactions")

	// Make sure the share was recorded in both the reward account state for the delegator and claim period distribution.
	state, err := s.app.RewardKeeper.GetRewardAccountState(s.ctx, program.GetId(), program.GetCurrentClaimPeriod(), delegatorAddr)
	s.Require().NoError(err, "GetRewardAccountState")
	s.Assert().Equal(1, int(state.SharesEarned), "state.SharesEarned")
	claimPeriodRewardDistribution, err := s.app.RewardKeeper.GetClaimPeriodRewardDistribution(s.ctx, program.GetCurrentClaimPeriod(), program.GetId())
	s.Require().NoError(err, "GetClaimPeriodRewardDistribution")
	s.Assert().Equal(1, int(claimPeriodRewardDistribution.GetTotalShares()), "claimPeriodRewardDistribution.GetTotalShares()")
}

func (s *KeeperTestSuite) TestIterateABCIEventsWildcard() {
	setupEventHistory(s)
	var events []types.ABCIEvent
	criteria := types.NewEventCriteria(events)
	counter := 0
	err := s.app.RewardKeeper.IterateABCIEvents(s.ctx, criteria, func(name string, attributes *map[string][]byte) error {
		counter += 1
		return nil
	})
	s.Assert().NoError(err, "IterateABCIEvents")
	s.Assert().Equal(3, counter, "should iterate for each event")
}

func (s *KeeperTestSuite) TestIterateABCIEventsByEventType() {
	setupEventHistory(s)
	counter := 0
	events := []types.ABCIEvent{
		{
			Type: "event1",
		},
	}
	criteria := types.NewEventCriteria(events)
	err := s.app.RewardKeeper.IterateABCIEvents(s.ctx, criteria, func(name string, attributes *map[string][]byte) error {
		counter += 1
		return nil
	})
	s.Assert().NoError(err, "IterateABCIEvents")
	s.Assert().Equal(2, counter, "should iterate only for specified event types")
}

func (s *KeeperTestSuite) TestIterateABCIEventsByEventTypeAndAttributeName() {
	setupEventHistory(s)
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
	err := s.app.RewardKeeper.IterateABCIEvents(s.ctx, criteria, func(name string, attributes *map[string][]byte) error {
		counter += 1
		return nil
	})
	s.Assert().NoError(err, "IterateABCIEvents")
	s.Assert().Equal(2, counter, "should iterate only for specified event types with matching attributes")
}

func (s *KeeperTestSuite) TestIterateABCIEventsByEventTypeAndAttributeNameAndValue() {
	setupEventHistory(s)
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
	err := s.app.RewardKeeper.IterateABCIEvents(s.ctx, criteria, func(name string, attributes *map[string][]byte) error {
		counter += 1
		return nil
	})
	s.Assert().NoError(err, "IterateABCIEvents")
	s.Assert().Equal(2, counter, "should iterate only for specified event types with matching attributes")
}

func (s *KeeperTestSuite) TestIterateABCIEventsNonExistantEventType() {
	setupEventHistory(s)
	counter := 0
	events := []types.ABCIEvent{
		{
			Type:       "event5",
			Attributes: map[string][]byte{},
		},
	}
	criteria := types.NewEventCriteria(events)
	err := s.app.RewardKeeper.IterateABCIEvents(s.ctx, criteria, func(name string, attributes *map[string][]byte) error {
		counter += 1
		return nil
	})
	s.Assert().NoError(err, "IterateABCIEvents")
	s.Assert().Equal(0, counter, "should not iterate if event doesn't exist")
}

func (s *KeeperTestSuite) TestIterateABCIEventsNonExistantAttributeName() {
	setupEventHistory(s)
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
	err := s.app.RewardKeeper.IterateABCIEvents(s.ctx, criteria, func(name string, attributes *map[string][]byte) error {
		counter += 1
		return nil
	})
	s.Assert().NoError(err, "IterateABCIEvents")
	s.Assert().Equal(0, counter, "should not iterate if attribute doesn't match")
}

func (s *KeeperTestSuite) TestIterateABCIEventsNonAttributeValueMatch() {
	setupEventHistory(s)
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
	err := s.app.RewardKeeper.IterateABCIEvents(s.ctx, criteria, func(name string, attributes *map[string][]byte) error {
		counter += 1
		return nil
	})
	s.Assert().NoError(err, "IterateABCIEvents")
	s.Assert().Equal(0, counter, "should not iterate if attribute doesn't match")
}

func (s *KeeperTestSuite) TestIterateABCIEventsHandlesError() {
	setupEventHistory(s)
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
	err := s.app.RewardKeeper.IterateABCIEvents(s.ctx, criteria, func(name string, attributes *map[string][]byte) error {
		counter += 1
		return errors.New("error")
	})
	s.Assert().Error(err, "should throw error when internal function errors")
}

func (s *KeeperTestSuite) TestFindQualifyingActionsWithDelegates() {
	SetupEventHistoryWithDelegates(s)
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

	action := MockAction{Criteria: criteria, Builder: &types.DelegateActionBuilder{}}
	events, err := s.app.RewardKeeper.FindQualifyingActions(s.ctx, action)
	s.Assert().NoError(err, "should throw no error when handling no events")
	s.Assert().Equal(2, len(events), "should find the two delegate events")
	for _, event := range events {
		s.Assert().Equal(event.Shares, int64(1), "shares must be 1")
		s.Assert().Equal(event.Delegator.String(), "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", "delegator address must be correct")
		s.Assert().Equal(event.Validator.String(), "cosmosvaloper15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqh6tjun", "validator address must be correct")
	}
}

func (s *KeeperTestSuite) TestFindQualifyingActionsWithoutDelegates() {
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
	action := MockAction{Criteria: criteria, Builder: &types.DelegateActionBuilder{}}
	events, err := s.app.RewardKeeper.FindQualifyingActions(s.ctx, action)
	s.Assert().NoError(err, "should throw no error when handling no events")
	s.Assert().Equal(0, len(events), "should have no events when no delegates are made")
}

// FindQualifyingActions

type MockAction struct {
	PassEvaluate bool
	Criteria     *types.EventCriteria
	Builder      types.ActionBuilder
}

func (m MockAction) ActionType() string {
	return ""
}

func (m MockAction) Evaluate(ctx sdk.Context, provider types.KeeperProvider, state types.RewardAccountState, event types.EvaluationResult) bool {
	return m.PassEvaluate
}

func (m MockAction) GetEventCriteria() *types.EventCriteria {
	return m.Criteria
}

func (m MockAction) GetBuilder() types.ActionBuilder {
	return m.Builder
}

func (m MockAction) PreEvaluate(ctx sdk.Context, provider types.KeeperProvider, state types.RewardAccountState) bool {
	return true
	// Any action specific thing that we need to do before evaluation
}

func (m MockAction) PostEvaluate(ctx sdk.Context, provider types.KeeperProvider, state types.RewardAccountState, evaluationResult types.EvaluationResult) (bool, types.EvaluationResult) {
	return true, evaluationResult
	// Any action specific thing that we need to do after evaluation
}

func (s *KeeperTestSuite) TestProcessQualifyingActionsWithNoAbciEvents() {
	program := types.RewardProgram{Id: 1, CurrentClaimPeriod: 1}
	action := MockAction{PassEvaluate: false}
	results := s.app.RewardKeeper.ProcessQualifyingActions(s.ctx, &program, action, []types.EvaluationResult{})
	s.Assert().Equal(0, len(results), "should have no results for empty list of abci events")
}

func (s *KeeperTestSuite) TestProcessQualifyingActionsCreatesState() {
	program := types.RewardProgram{Id: 1, CurrentClaimPeriod: 1}
	action := MockAction{PassEvaluate: true}
	address1, _ := sdk.AccAddressFromBech32("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h")
	s.app.RewardKeeper.ProcessQualifyingActions(s.ctx, &program, action, []types.EvaluationResult{
		{
			Address: address1,
		},
	})
	state, _ := s.app.RewardKeeper.GetRewardAccountState(s.ctx, program.GetId(), program.GetCurrentClaimPeriod(), address1.String())
	s.Assert().Equal(program.GetId(), state.RewardProgramId, "claim period should be created when missing")
}

func (s *KeeperTestSuite) TestProcessQualifyingActionsWithNoMatchingResults() {
	program := types.RewardProgram{Id: 1, CurrentClaimPeriod: 1}
	action := MockAction{PassEvaluate: false}
	results := s.app.RewardKeeper.ProcessQualifyingActions(s.ctx, &program, action, []types.EvaluationResult{
		{Address: types.MustAccAddressFromBech32("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h")},
		{Address: types.MustAccAddressFromBech32("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h")},
	})
	s.Assert().Equal(0, len(results), "should have empty lists when no results match the evaluation")
}

func (s *KeeperTestSuite) TestProcessQualifyingActionsWithMatchingResults() {
	program := types.RewardProgram{Id: 1, CurrentClaimPeriod: 1}
	action := MockAction{PassEvaluate: true}
	results := s.app.RewardKeeper.ProcessQualifyingActions(s.ctx, &program, action, []types.EvaluationResult{
		{Address: types.MustAccAddressFromBech32("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h")},
		{Address: types.MustAccAddressFromBech32("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h")},
	})
	s.Assert().Equal(2, len(results), "should have all results that evaluate to true")
}

func (s *KeeperTestSuite) TestFindQualifyingActionsWithNil() {
	results := s.app.RewardKeeper.ProcessQualifyingActions(s.ctx, nil, nil, nil)
	s.Assert().Equal(0, len(results), "should handle nil input")
}

// Test ActionDelegate Evaluate

type MockKeeperProvider struct {
}

func (m MockKeeperProvider) GetStakingKeeper() types.StakingKeeper {
	return MockStakingKeeper{}
}

func (m MockKeeperProvider) GetAccountKeeper() types.AccountKeeper {
	return MockAccountKeeper{}
}

type MockStakingKeeper struct {
}

type MockAccountKeeper struct {
}

func (m MockAccountKeeper) GetModuleAddress(moduleName string) sdk.AccAddress {
	return nil
}

func (m MockStakingKeeper) GetAllDelegatorDelegations(ctx context.Context, delegator sdk.AccAddress) ([]stakingtypes.Delegation, error) {
	if delegator.String() == "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h" || delegator.String() == getOperatorBech32AddressForTestValidator().String() {
		return []stakingtypes.Delegation{
			{
				DelegatorAddress: "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
				ValidatorAddress: "cosmosvaloper15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqh6tjun",
				Shares:           sdkmath.LegacyNewDec(3),
			},
		}, nil
	}
	return []stakingtypes.Delegation{
		{},
	}, fmt.Errorf("mock error: delegator %s not found", delegator)
}

func (m MockStakingKeeper) GetDelegation(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) (delegation stakingtypes.Delegation, err error) {
	if delAddr.String() != "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h" || valAddr.String() != "cosmosvaloper15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqh6tjun" {
		return stakingtypes.Delegation{}, fmt.Errorf("mock error: delegator %s not found for %s", delAddr, valAddr)
	}

	return stakingtypes.Delegation{
		DelegatorAddress: "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
		ValidatorAddress: "cosmosvaloper15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqh6tjun",
		Shares:           sdkmath.LegacyNewDec(3),
	}, nil
}

func (m MockStakingKeeper) GetLastValidatorPower(ctx context.Context, operator sdk.ValAddress) (power int64, err error) {
	validators, _ := m.GetBondedValidatorsByPower(ctx)
	for i, v := range validators {
		power := int64(len(validators) - i)
		if v.GetOperator() == operator.String() {
			return power, nil
		}
	}
	return 0, fmt.Errorf("mock error: operator not found %s", operator)
}

func (s *KeeperTestSuite) createTestValidators(amount int) {
	addrDels := simapp.AddTestAddrsIncremental(s.app, s.ctx, amount, sdkmath.NewInt(10000))
	valAddrs := simtestutil.ConvertAddrsToValAddrs(addrDels)

	bondDenom, err := s.app.StakingKeeper.BondDenom(s.ctx)
	s.Require().NoError(err, "app.StakingKeeper.BondDenom(ctx)")

	totalSupply := sdk.NewCoins(sdk.NewInt64Coin(bondDenom, 1_000_000_000))
	notBondedPool := s.app.StakingKeeper.GetNotBondedPool(s.ctx)
	s.app.AccountKeeper.SetModuleAccount(s.ctx, notBondedPool)

	s.Require().NoError(testutil.FundModuleAccount(s.ctx, s.app.BankKeeper, notBondedPool.GetName(), totalSupply),
		"funding %s with %s", notBondedPool.GetName(), totalSupply)

	var validators []stakingtypes.Validator
	for i := 0; i < amount; i++ {
		validator := teststaking.NewValidator(s.T(), valAddrs[i], PKs[i])
		tokens := s.app.StakingKeeper.TokensFromConsensusPower(s.ctx, int64(1+amount-i))
		validator, _ = validator.AddTokensFromDel(tokens)
		validator = keeper.TestingUpdateValidator(&s.app.StakingKeeper, s.ctx, validator, true)
		validators = append(validators, validator)

		// Create the delegations
		bond := stakingtypes.NewDelegation(addrDels[i].String(), valAddrs[i].String(), sdkmath.LegacyNewDec(int64((i+1)*10)))
		err = s.app.StakingKeeper.SetDelegation(s.ctx, bond)
		s.Require().NoError(err, "SetDelegation %s", bond)

		// We want even validators to be bonded
		if i%2 == 0 {
			_, err := s.app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(s.ctx)
			s.Require().NoError(err, "ApplyAndReturnValidatorSetUpdates")
			validator.ABCIValidatorUpdate(s.app.StakingKeeper.PowerReduction(s.ctx))
		}
	}

	_, err = s.app.StakingKeeper.EndBlocker(s.ctx)
	s.Require().NoError(err, "staking end blocker")

	testValidators = validators
}

func getTestValidators(start, end int) []stakingtypes.Validator {
	var validators []stakingtypes.Validator
	for i := start; i <= end; i++ {
		validators = append(validators, testValidators[i])
	}
	return validators
}

func (m MockStakingKeeper) GetBondedValidatorsByPower(ctx context.Context) ([]stakingtypes.Validator, error) {
	validatorAddress, _ := sdk.ValAddressFromBech32("cosmosvaloper15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqh6tjun")
	validator, _ := stakingtypes.NewValidator(
		validatorAddress.String(),
		PKs[100],
		stakingtypes.Description{},
	)
	validators := getTestValidators(1, 3)
	validators = append(validators, validator)
	validators = append(validators, getTestValidators(4, 6)...)

	return validators, nil
}

func (m MockStakingKeeper) GetValidator(ctx context.Context, addr sdk.ValAddress) (validator stakingtypes.Validator, err error) {
	validators := getTestValidators(0, 9)
	addrStr := addr.String()

	for _, v := range validators {
		if addrStr == v.GetOperator() {
			return v, nil
		}
	}

	return stakingtypes.Validator{}, fmt.Errorf("mock error: validator not fount %s", addr)
}

func (s *KeeperTestSuite) TestActionDelegateEvaluatePasses() {
	action := types.NewActionDelegate()
	action.MinimumActions = 1
	action.MaximumActions = 2

	minDelegation := sdk.NewInt64Coin("nhash", 2)
	maxDelegation := sdk.NewInt64Coin("nhash", 10)
	action.MinimumDelegationAmount = &minDelegation
	action.MaximumDelegationAmount = &maxDelegation
	action.MinimumActiveStakePercentile = sdkmath.LegacyNewDecWithPrec(4, 1)
	action.MaximumActiveStakePercentile = sdkmath.LegacyNewDecWithPrec(1, 0)

	keeperProvider := MockKeeperProvider{}
	state := types.NewRewardAccountState(0, 0, "", 0, []*types.ActionCounter{})
	state.ActionCounter = types.IncrementActionCount(state.ActionCounter, action.ActionType())

	validator, _ := sdk.ValAddressFromBech32("cosmosvaloper15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqh6tjun")
	delegator, _ := sdk.AccAddressFromBech32("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h")
	event := types.EvaluationResult{
		Validator: validator,
		Delegator: delegator,
	}

	passed := action.Evaluate(s.ctx, keeperProvider, state, event)
	s.Assert().True(passed, "evaluate should pass when criteria are met")
}

func (s *KeeperTestSuite) TestActionDelegateEvaluateFailsWhenMinimumActionsNotMet() {
	action := types.NewActionDelegate()
	action.MinimumActions = 1
	action.MaximumActions = 2
	minDelegation := sdk.NewInt64Coin("nhash", 2)
	maxDelegation := sdk.NewInt64Coin("nhash", 10)
	action.MinimumDelegationAmount = &minDelegation
	action.MaximumDelegationAmount = &maxDelegation
	action.MinimumActiveStakePercentile = sdkmath.LegacyNewDecWithPrec(5, 1)
	action.MaximumActiveStakePercentile = sdkmath.LegacyNewDecWithPrec(1, 0)

	keeperProvider := MockKeeperProvider{}
	state := types.NewRewardAccountState(0, 0, "", 0, []*types.ActionCounter{})

	validator, _ := sdk.ValAddressFromBech32("cosmosvaloper15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqh6tjun")
	delegator, _ := sdk.AccAddressFromBech32("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h")
	event := types.EvaluationResult{
		Validator: validator,
		Delegator: delegator,
	}

	passed := action.Evaluate(s.ctx, keeperProvider, state, event)
	s.Assert().False(passed, "test should fail when minimum actions not met")
}

func (s *KeeperTestSuite) TestActionDelegateEvaluateFailsWhenMaximumActionsNotMet() {
	action := types.NewActionDelegate()
	action.MinimumActions = 1
	action.MaximumActions = 2
	minDelegation := sdk.NewInt64Coin("nhash", 2)
	maxDelegation := sdk.NewInt64Coin("nhash", 10)
	action.MinimumDelegationAmount = &minDelegation
	action.MaximumDelegationAmount = &maxDelegation
	action.MinimumActiveStakePercentile = sdkmath.LegacyNewDecWithPrec(5, 1)
	action.MaximumActiveStakePercentile = sdkmath.LegacyNewDecWithPrec(1, 0)

	keeperProvider := MockKeeperProvider{}
	state := types.NewRewardAccountState(0, 0, "", 0, []*types.ActionCounter{})
	state.ActionCounter = types.IncrementActionCount(state.ActionCounter, action.ActionType())
	state.ActionCounter = types.IncrementActionCount(state.ActionCounter, action.ActionType())
	state.ActionCounter = types.IncrementActionCount(state.ActionCounter, action.ActionType())

	validator, _ := sdk.ValAddressFromBech32("cosmosvaloper15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqh6tjun")
	delegator, _ := sdk.AccAddressFromBech32("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h")
	event := types.EvaluationResult{
		Validator: validator,
		Delegator: delegator,
	}

	passed := action.Evaluate(s.ctx, keeperProvider, state, event)
	s.Assert().False(passed, "test should fail when maximum actions not met")
}

func (s *KeeperTestSuite) TestActionDelegateEvaluateFailsWhenMaximumDelegationAmountNotMet() {
	action := types.NewActionDelegate()
	action.MinimumActions = 1
	action.MaximumActions = 2
	minDelegation := sdk.NewInt64Coin("nhash", 1)
	maxDelegation := sdk.NewInt64Coin("nhash", 1)
	action.MinimumDelegationAmount = &minDelegation
	action.MaximumDelegationAmount = &maxDelegation
	action.MinimumActiveStakePercentile = sdkmath.LegacyNewDecWithPrec(5, 1)
	action.MaximumActiveStakePercentile = sdkmath.LegacyNewDecWithPrec(1, 0)

	keeperProvider := MockKeeperProvider{}
	state := types.NewRewardAccountState(0, 0, "", 0, []*types.ActionCounter{})
	state.ActionCounter = types.IncrementActionCount(state.ActionCounter, action.ActionType())

	validator, _ := sdk.ValAddressFromBech32("cosmosvaloper15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqh6tjun")
	delegator, _ := sdk.AccAddressFromBech32("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h")
	event := types.EvaluationResult{
		Validator: validator,
		Delegator: delegator,
	}

	passed := action.Evaluate(s.ctx, keeperProvider, state, event)
	s.Assert().False(passed, "test should fail when maximum delegation amount not met")
}

func (s *KeeperTestSuite) TestActionDelegateEvaluateFailsWhenMinimumDelegationAmountNotMet() {
	action := types.NewActionDelegate()
	action.MinimumActions = 1
	action.MaximumActions = 2
	minDelegation := sdk.NewInt64Coin("nhash", 5)
	maxDelegation := sdk.NewInt64Coin("nhash", 10)
	action.MinimumDelegationAmount = &minDelegation
	action.MaximumDelegationAmount = &maxDelegation
	action.MinimumActiveStakePercentile = sdkmath.LegacyNewDecWithPrec(5, 1)
	action.MaximumActiveStakePercentile = sdkmath.LegacyNewDecWithPrec(1, 0)

	keeperProvider := MockKeeperProvider{}
	state := types.NewRewardAccountState(0, 0, "", 0, []*types.ActionCounter{})
	state.ActionCounter = types.IncrementActionCount(state.ActionCounter, action.ActionType())

	validator, _ := sdk.ValAddressFromBech32("cosmosvaloper15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqh6tjun")
	delegator, _ := sdk.AccAddressFromBech32("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h")
	event := types.EvaluationResult{
		Validator: validator,
		Delegator: delegator,
	}

	passed := action.Evaluate(s.ctx, keeperProvider, state, event)
	s.Assert().False(passed, "test should fail when minimum delegation amount not met")
}

func (s *KeeperTestSuite) TestActionDelegateEvaluateFailsWhenMinimumActiveStakePercentileNotMet() {
	action := types.NewActionDelegate()
	action.MinimumActions = 1
	action.MaximumActions = 2
	minDelegation := sdk.NewInt64Coin("nhash", 1)
	maxDelegation := sdk.NewInt64Coin("nhash", 10)
	action.MinimumDelegationAmount = &minDelegation
	action.MaximumDelegationAmount = &maxDelegation
	action.MinimumActiveStakePercentile = sdkmath.LegacyNewDecWithPrec(11, 1)
	action.MaximumActiveStakePercentile = sdkmath.LegacyNewDecWithPrec(1, 0)

	keeperProvider := MockKeeperProvider{}
	state := types.NewRewardAccountState(0, 0, "", 0, []*types.ActionCounter{})
	state.ActionCounter = types.IncrementActionCount(state.ActionCounter, action.ActionType())

	validator, _ := sdk.ValAddressFromBech32("cosmosvaloper15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqh6tjun")
	delegator, _ := sdk.AccAddressFromBech32("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h")
	event := types.EvaluationResult{
		Validator: validator,
		Delegator: delegator,
	}

	passed := action.Evaluate(s.ctx, keeperProvider, state, event)
	s.Assert().False(passed, "test should fail when minimum delegation percentage not met")
}

func (s *KeeperTestSuite) TestActionDelegateEvaluateFailsWhenMaximumDelegationPercentageNotMet() {
	action := types.NewActionDelegate()
	action.MinimumActions = 1
	action.MaximumActions = 2
	minDelegation := sdk.NewInt64Coin("nhash", 1)
	maxDelegation := sdk.NewInt64Coin("nhash", 10)
	action.MinimumDelegationAmount = &minDelegation
	action.MaximumDelegationAmount = &maxDelegation
	action.MinimumActiveStakePercentile = sdkmath.LegacyNewDecWithPrec(5, 1)
	action.MaximumActiveStakePercentile = sdkmath.LegacyNewDecWithPrec(1, 0)

	keeperProvider := MockKeeperProvider{}
	state := types.NewRewardAccountState(0, 0, "", 0, []*types.ActionCounter{})
	state.ActionCounter = types.IncrementActionCount(state.ActionCounter, action.ActionType())

	validator, _ := sdk.ValAddressFromBech32("cosmosvaloper15ky9du8a2wlstz6fpx3p4mqpjyrm5cgqh6tjun")
	delegator, _ := sdk.AccAddressFromBech32("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h")
	event := types.EvaluationResult{
		Validator: validator,
		Delegator: delegator,
	}

	passed := action.Evaluate(s.ctx, keeperProvider, state, event)
	s.Assert().False(passed, "test should fail when maximum delegation percentage not met")
}

// Test GetRewardAction

func (s *KeeperTestSuite) TestGetRewardActionHandlesUnsupportedActions() {
	qa := types.QualifyingAction{}
	_, err := qa.GetRewardAction(s.ctx)
	s.Assert().Error(err, "should throw error when an action is not supported")
}

func (s *KeeperTestSuite) TestGetRewardActionHandlesActionDelegate() {
	delegate := types.QualifyingAction_Delegate{}
	qa := types.QualifyingAction{
		Type: &delegate,
	}
	action, err := qa.GetRewardAction(s.ctx)
	s.Assert().NoError(err, "should not throw error when action is supported")
	s.Assert().Equal(types.ActionTypeDelegate, action.ActionType(), "should return the correct action type")
}

// Test DetectQualifyingActions
func (s *KeeperTestSuite) TestDetectQualifyingActionsWith1QualifyingAction() {
	SetupEventHistoryWithDelegates(s)
	s.app.RewardKeeper.SetStakingKeeper(MockStakingKeeper{})
	minDelegation := sdk.NewInt64Coin("nhash", 0)
	maxDelegation := sdk.NewInt64Coin("nhash", 10)

	rewardProgram := types.NewRewardProgram(
		"title",
		"description",
		1,
		"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
		sdk.NewInt64Coin("hotdog", 10000),
		sdk.NewInt64Coin("hotdog", 10000),
		time.Now(),
		5,
		5,
		0,
		0,
		[]types.QualifyingAction{
			{
				Type: &types.QualifyingAction_Delegate{
					Delegate: &types.ActionDelegate{
						MinimumActions:               0,
						MaximumActions:               1,
						MinimumDelegationAmount:      &minDelegation,
						MaximumDelegationAmount:      &maxDelegation,
						MinimumActiveStakePercentile: sdkmath.LegacyNewDecWithPrec(0, 0),
						MaximumActiveStakePercentile: sdkmath.LegacyNewDecWithPrec(1, 0),
					},
				},
			},
		},
	)
	rewardProgram.CurrentClaimPeriod = 1
	qualifyingActions, err := s.app.RewardKeeper.DetectQualifyingActions(s.ctx, &rewardProgram)
	s.Assert().NoError(err, "must not error")
	s.Assert().Equal(1, len(qualifyingActions), "must find one qualifying actions")
}

func (s *KeeperTestSuite) TestDetectQualifyingActionsWith2QualifyingAction() {
	SetupEventHistoryWithDelegates(s)
	s.app.RewardKeeper.SetStakingKeeper(MockStakingKeeper{})
	minDelegation := sdk.NewInt64Coin("nhash", 0)
	maxDelegation := sdk.NewInt64Coin("nhash", 10)

	rewardProgram := types.NewRewardProgram(
		"title",
		"description",
		1,
		"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
		sdk.NewInt64Coin("hotdog", 10000),
		sdk.NewInt64Coin("hotdog", 10000),
		time.Now(),
		5,
		5,
		0,
		0,
		[]types.QualifyingAction{
			{
				Type: &types.QualifyingAction_Delegate{
					Delegate: &types.ActionDelegate{
						MinimumActions:               0,
						MaximumActions:               4,
						MinimumDelegationAmount:      &minDelegation,
						MaximumDelegationAmount:      &maxDelegation,
						MinimumActiveStakePercentile: sdkmath.LegacyNewDecWithPrec(0, 0),
						MaximumActiveStakePercentile: sdkmath.LegacyNewDecWithPrec(1, 0),
					},
				},
			},
			{
				Type: &types.QualifyingAction_Delegate{
					Delegate: &types.ActionDelegate{
						MinimumActions:               0,
						MaximumActions:               4,
						MinimumDelegationAmount:      &minDelegation,
						MaximumDelegationAmount:      &maxDelegation,
						MinimumActiveStakePercentile: sdkmath.LegacyNewDecWithPrec(0, 0),
						MaximumActiveStakePercentile: sdkmath.LegacyNewDecWithPrec(1, 0),
					},
				},
			},
		},
	)
	rewardProgram.CurrentClaimPeriod = 1
	qualifyingActions, err := s.app.RewardKeeper.DetectQualifyingActions(s.ctx, &rewardProgram)
	s.Assert().NoError(err, "must not error")
	s.Assert().Equal(4, len(qualifyingActions), "must find four qualifying actions")
}

func (s *KeeperTestSuite) TestDetectQualifyingActionsWithNoQualifyingAction() {
	SetupEventHistoryWithDelegates(s)
	s.app.RewardKeeper.SetStakingKeeper(MockStakingKeeper{})

	rewardProgram := types.NewRewardProgram(
		"title",
		"description",
		1,
		"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
		sdk.NewInt64Coin("hotdog", 10000),
		sdk.NewInt64Coin("hotdog", 10000),
		time.Now(),
		5,
		5,
		0,
		0,
		[]types.QualifyingAction{},
	)
	rewardProgram.CurrentClaimPeriod = 1

	qualifyingActions, err := s.app.RewardKeeper.DetectQualifyingActions(s.ctx, &rewardProgram)
	s.Assert().NoError(err, "must not error")
	s.Assert().Equal(0, len(qualifyingActions), "must find no qualifying actions")
}

func (s *KeeperTestSuite) TestDetectQualifyingActionsWithNoMatchingQualifyingAction() {
	SetupEventHistoryWithDelegates(s)
	s.app.RewardKeeper.SetStakingKeeper(MockStakingKeeper{})
	minDelegation := sdk.NewInt64Coin("nhash", 0)
	maxDelegation := sdk.NewInt64Coin("nhash", 10)

	rewardProgram := types.NewRewardProgram(
		"title",
		"description",
		1,
		"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
		sdk.NewInt64Coin("hotdog", 10000),
		sdk.NewInt64Coin("hotdog", 10000),
		time.Now(),
		5,
		5,
		0,
		0,
		[]types.QualifyingAction{
			{
				Type: &types.QualifyingAction_Delegate{
					Delegate: &types.ActionDelegate{
						MinimumActions:               1000,
						MaximumActions:               1000,
						MinimumDelegationAmount:      &minDelegation,
						MaximumDelegationAmount:      &maxDelegation,
						MinimumActiveStakePercentile: sdkmath.LegacyNewDecWithPrec(0, 0),
						MaximumActiveStakePercentile: sdkmath.LegacyNewDecWithPrec(1, 0),
					},
				},
			},
		},
	)
	rewardProgram.CurrentClaimPeriod = 1
	qualifyingActions, err := s.app.RewardKeeper.DetectQualifyingActions(s.ctx, &rewardProgram)
	s.Assert().NoError(err, "must not error")
	s.Assert().Equal(0, len(qualifyingActions), "must find no qualifying actions")
}

// Test RewardShares
func (s *KeeperTestSuite) TestRewardSharesSingle() {
	rewardProgram := types.NewRewardProgram(
		"title",
		"description",
		1,
		"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
		sdk.NewInt64Coin("hotdog", 10000),
		sdk.NewInt64Coin("hotdog", 10000),
		time.Now(),
		5,
		5,
		0,
		0,
		[]types.QualifyingAction{},
	)
	rewardProgram.CurrentClaimPeriod = 1

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

	state := types.NewRewardAccountState(rewardProgram.GetId(), rewardProgram.GetCurrentClaimPeriod(), delegator.String(), 0, []*types.ActionCounter{})
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state)
	claimPeriodRewardDistribution := types.NewClaimPeriodRewardDistribution(rewardProgram.GetCurrentClaimPeriod(),
		rewardProgram.GetId(),
		sdk.NewInt64Coin("nhash", 100),
		sdk.NewInt64Coin("nhash", 100),
		0,
		false,
	)
	s.app.RewardKeeper.SetClaimPeriodRewardDistribution(s.ctx, claimPeriodRewardDistribution)

	err := s.app.RewardKeeper.RewardShares(s.ctx, &rewardProgram, results)

	state, _ = s.app.RewardKeeper.GetRewardAccountState(s.ctx, rewardProgram.GetId(), rewardProgram.GetCurrentClaimPeriod(), delegator.String())
	claimPeriodRewardDistribution, _ = s.app.RewardKeeper.GetClaimPeriodRewardDistribution(s.ctx, rewardProgram.GetCurrentClaimPeriod(), rewardProgram.GetId())
	s.Assert().NoError(err, "should return no error on success")
	s.Assert().Equal(uint64(1), state.GetSharesEarned(), "share amount should increment")
	s.Assert().Equal(int64(1), claimPeriodRewardDistribution.GetTotalShares(), "total share amount should increment")
	s.Assert().Equal(rewardProgram.GetId(), state.GetRewardProgramId(), "reward program id should match")
	s.Assert().Equal(rewardProgram.GetCurrentClaimPeriod(), state.GetClaimPeriodId(), "reward claim period id should match")
	s.Assert().Equal(delegator.String(), state.GetAddress(), "address should match delegator")
}

func (s *KeeperTestSuite) TestRewardSharesInvalidClaimPeriodRewardDistribution() {
	rewardProgram := types.NewRewardProgram(
		"title",
		"description",
		1,
		"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
		sdk.NewInt64Coin("hotdog", 10000),
		sdk.NewInt64Coin("hotdog", 10000),
		time.Now(),
		5,
		5,
		0,
		0,
		[]types.QualifyingAction{},
	)
	rewardProgram.CurrentClaimPeriod = 1

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

	state := types.NewRewardAccountState(rewardProgram.GetId(), rewardProgram.GetCurrentClaimPeriod(), delegator.String(), 0, []*types.ActionCounter{})
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state)
	claimPeriodRewardDistribution := types.NewClaimPeriodRewardDistribution(rewardProgram.GetCurrentClaimPeriod(),
		rewardProgram.GetId(),
		sdk.NewInt64Coin("nhash", 0),
		sdk.NewInt64Coin("nhash", 0),
		0,
		false,
	)
	s.app.RewardKeeper.SetClaimPeriodRewardDistribution(s.ctx, claimPeriodRewardDistribution)

	err := s.app.RewardKeeper.RewardShares(s.ctx, &rewardProgram, results)
	s.Assert().Error(err, "should return an error on invalid claim period reward distribution")
}

func (s *KeeperTestSuite) TestRewardSharesMultiple() {
	rewardProgram := types.NewRewardProgram(
		"title",
		"description",
		1,
		"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
		sdk.NewInt64Coin("hotdog", 10000),
		sdk.NewInt64Coin("hotdog", 10000),
		time.Now(),
		5,
		5,
		0,
		0,
		[]types.QualifyingAction{},
	)
	rewardProgram.CurrentClaimPeriod = 1

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

	state := types.NewRewardAccountState(rewardProgram.GetId(), rewardProgram.GetCurrentClaimPeriod(), delegator.String(), 0, []*types.ActionCounter{})
	s.app.RewardKeeper.SetRewardAccountState(s.ctx, state)
	claimPeriodRewardDistribution := types.NewClaimPeriodRewardDistribution(rewardProgram.GetCurrentClaimPeriod(),
		rewardProgram.GetId(),
		sdk.NewInt64Coin("nhash", 100),
		sdk.NewInt64Coin("nhash", 100),
		0,
		false,
	)
	s.app.RewardKeeper.SetClaimPeriodRewardDistribution(s.ctx, claimPeriodRewardDistribution)

	err := s.app.RewardKeeper.RewardShares(s.ctx, &rewardProgram, results)

	claimPeriodRewardDistribution, _ = s.app.RewardKeeper.GetClaimPeriodRewardDistribution(s.ctx, rewardProgram.GetCurrentClaimPeriod(), rewardProgram.GetId())
	state, _ = s.app.RewardKeeper.GetRewardAccountState(s.ctx, rewardProgram.GetId(), rewardProgram.GetCurrentClaimPeriod(), delegator.String())
	s.Assert().NoError(err, "should return no error on success")
	s.Assert().Equal(uint64(2), state.GetSharesEarned(), "share amount should increment")
	s.Assert().Equal(int64(2), claimPeriodRewardDistribution.GetTotalShares(), "total share amount should increment")
	s.Assert().Equal(rewardProgram.GetId(), state.GetRewardProgramId(), "reward program id should match")
	s.Assert().Equal(rewardProgram.GetCurrentClaimPeriod(), state.GetClaimPeriodId(), "reward claim period id should match")
	s.Assert().Equal(delegator.String(), state.GetAddress(), "address should match delegator")
}

func (s *KeeperTestSuite) TestRewardSharesInvalidRewardProgram() {
	rewardProgram := types.NewRewardProgram(
		"title",
		"description",
		0,
		"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
		sdk.NewInt64Coin("hotdog", 10000),
		sdk.NewInt64Coin("hotdog", 10000),
		time.Now(),
		5,
		5,
		0,
		0,
		[]types.QualifyingAction{},
	)
	rewardProgram.CurrentClaimPeriod = 1

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
	claimPeriodRewardDistribution := types.NewClaimPeriodRewardDistribution(rewardProgram.GetCurrentClaimPeriod(),
		rewardProgram.GetId(),
		sdk.NewInt64Coin("nhash", 100),
		sdk.NewInt64Coin("nhash", 100),
		0,
		false,
	)
	s.app.RewardKeeper.SetClaimPeriodRewardDistribution(s.ctx, claimPeriodRewardDistribution)

	err := s.app.RewardKeeper.RewardShares(s.ctx, nil, results)
	state, _ := s.app.RewardKeeper.GetRewardAccountState(s.ctx, rewardProgram.GetId(), rewardProgram.GetCurrentClaimPeriod(), delegator.String())

	s.Assert().Error(err, "should return an error on invalid program")
	s.Assert().Equal(uint64(0), state.GetSharesEarned(), "share amount should increment")
}

// with transfer
func SetupEventHistoryWithTransfers(s *KeeperTestSuite) {
	sender := "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h"
	recipient := "cosmos1tnh2q55v8wyygtt9srz5safamzdengsnqeycj3"
	attributes1 := []sdk.Attribute{
		sdk.NewAttribute("sender", sender),
		sdk.NewAttribute("recipient", recipient),
		sdk.NewAttribute("amount", "1000000000000000nhash"),
	}
	attributes2 := []sdk.Attribute{
		sdk.NewAttribute("module", "bank"),
		sdk.NewAttribute("sender", sender),
		sdk.NewAttribute("action", "/cosmos.bank.v1beta1.MsgSend"),
	}

	event1 := sdk.NewEvent("transfer", attributes1...)
	event2 := sdk.NewEvent("message", attributes2...)
	loggedEvents := sdk.Events{
		event1,
		event2,
	}
	SetupEventHistory(s, loggedEvents)
}

// with vote
func SetupEventHistoryWithVotes(s *KeeperTestSuite, sender string) {
	attributes1 := []sdk.Attribute{
		sdk.NewAttribute("action", "/cosmos.gov.v1beta1.MsgVote"),
	}
	attributes2 := []sdk.Attribute{
		sdk.NewAttribute("option", "{\"option\":1,\"weight\":\"1.000000000000000000\"}"),
		sdk.NewAttribute("proposal_id", "1"),
	}
	attributes3 := []sdk.Attribute{
		sdk.NewAttribute("module", "governance"),
		sdk.NewAttribute("sender", sender),
	}

	event1 := sdk.NewEvent("message", attributes1...)
	event2 := sdk.NewEvent("proposal_vote", attributes2...)
	event3 := sdk.NewEvent("message", attributes3...)
	loggedEvents := sdk.Events{
		event1,
		event2,
		event3,
	}
	// TODO[1760]: event-history: Put this back once the event history stuff is back in the SDK.
	_ = loggedEvents
	// newEvents := loggedEvents.ToABCIEvents()
	// newEvents = append(newEvents, s.ctx.EventManager().GetABCIEventHistory()...)
	// eventManagerStub := sdk.NewEventManagerWithHistory(newEvents)
	// s.ctx = s.ctx.WithEventManager(eventManagerStub)
}

// transfer
func (s *KeeperTestSuite) TestFindQualifyingActionsWithTransfers() {
	SetupEventHistoryWithTransfers(s)
	criteria := types.NewEventCriteria([]types.ABCIEvent{
		{
			Type:       banktypes.EventTypeTransfer,
			Attributes: map[string][]byte{},
		},
	})

	action := MockAction{Criteria: criteria, Builder: &types.TransferActionBuilder{}}
	events, err := s.app.RewardKeeper.FindQualifyingActions(s.ctx, action)
	s.Assert().NoError(err, "should throw no error when handling no events")
	s.Assert().Equal(1, len(events), "should find the one transfer event")
	for _, event := range events {
		s.Assert().Equal(event.Shares, int64(1), "shares must be 1")
		s.Assert().Equal(event.Address.String(), "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", "address must be correct")
	}
}

// vote
func (s *KeeperTestSuite) TestFindQualifyingActionsWithVotes() {
	SetupEventHistoryWithVotes(s, "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h")
	criteria := types.NewEventCriteria([]types.ABCIEvent{
		{
			Type:       sdk.EventTypeMessage,
			Attributes: map[string][]byte{sdk.AttributeKeyModule: []byte("governance")}, // TODO[1760]: reward: Ensure "governance" is still correct here.
		},
	})

	action := MockAction{Criteria: criteria, Builder: &types.VoteActionBuilder{}}
	events, err := s.app.RewardKeeper.FindQualifyingActions(s.ctx, action)
	s.Assert().NoError(err, "should throw no error when handling no events")
	s.Assert().Equal(1, len(events), "should find the one transfer event")
	for _, event := range events {
		s.Assert().Equal(event.Shares, int64(1), "shares must be 1")
		s.Assert().Equal(event.Address.String(), "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", "address must be correct")
	}
}

func (s *KeeperTestSuite) TestDetectQualifyingActionsWith1VotingQualifyingAction() {
	SetupEventHistoryWithVotes(s, "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h")
	s.app.RewardKeeper.SetStakingKeeper(MockStakingKeeper{})
	minDelegation := sdk.NewInt64Coin("nhash", 3)

	rewardProgram := types.NewRewardProgram(
		"title",
		"description",
		1,
		"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
		sdk.NewInt64Coin("hotdog", 10000),
		sdk.NewInt64Coin("hotdog", 10000),
		time.Now(),
		5,
		5,
		0,
		0,
		[]types.QualifyingAction{
			{
				Type: &types.QualifyingAction_Vote{
					Vote: &types.ActionVote{
						MinimumActions:          0,
						MaximumActions:          1,
						MinimumDelegationAmount: minDelegation,
					},
				},
			},
		},
	)
	rewardProgram.CurrentClaimPeriod = 1
	qualifyingActions, err := s.app.RewardKeeper.DetectQualifyingActions(s.ctx, &rewardProgram)
	s.Assert().NoError(err, "must not error")
	s.Assert().Equal(1, len(qualifyingActions), "must find one qualifying actions")
}

func (s *KeeperTestSuite) TestDetectQualifyingActionsWith1VotingQualifyingActionMultiplierPresent() {
	s.SetupTest()
	SetupEventHistoryWithVotes(s, getOperatorBech32AddressForTestValidator().String())
	s.app.RewardKeeper.SetStakingKeeper(MockStakingKeeper{})
	minDelegation := sdk.NewInt64Coin("nhash", 0)

	rewardProgram := types.NewRewardProgram(
		"title",
		"description",
		1,
		"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
		sdk.NewInt64Coin("hotdog", 10000),
		sdk.NewInt64Coin("hotdog", 10000),
		time.Now(),
		5,
		5,
		0,
		0,
		[]types.QualifyingAction{
			{
				Type: &types.QualifyingAction_Vote{
					Vote: &types.ActionVote{
						MinimumActions:          0,
						MaximumActions:          1,
						MinimumDelegationAmount: minDelegation,
						ValidatorMultiplier:     10,
					},
				},
			},
		},
	)
	rewardProgram.CurrentClaimPeriod = 1
	qualifyingActions, err := s.app.RewardKeeper.DetectQualifyingActions(s.ctx, &rewardProgram)
	s.Assert().NoError(err, "must not error")
	s.Assert().Equal(1, len(qualifyingActions), "must find one qualifying actions")
	s.Assert().Equal(int64(10), qualifyingActions[0].Shares, "shares should be 10")
}

func (s *KeeperTestSuite) TestDetectQualifyingActionsWith1VotingQualifyingActionMultiplierPresentAndDelegationRequired() {
	s.SetupTest()
	SetupEventHistoryWithVotes(s, getOperatorBech32AddressForTestValidator().String())
	s.app.RewardKeeper.SetStakingKeeper(MockStakingKeeper{})
	minDelegation := sdk.NewInt64Coin("nhash", 3)

	rewardProgram := types.NewRewardProgram(
		"title",
		"description",
		1,
		"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
		sdk.NewInt64Coin("hotdog", 10000),
		sdk.NewInt64Coin("hotdog", 10000),
		time.Now(),
		5,
		5,
		0,
		0,
		[]types.QualifyingAction{
			{
				Type: &types.QualifyingAction_Vote{
					Vote: &types.ActionVote{
						MinimumActions:          0,
						MaximumActions:          1,
						MinimumDelegationAmount: minDelegation,
						ValidatorMultiplier:     10,
					},
				},
			},
		},
	)
	rewardProgram.CurrentClaimPeriod = 1
	qualifyingActions, err := s.app.RewardKeeper.DetectQualifyingActions(s.ctx, &rewardProgram)
	s.Assert().NoError(err, "must not error")
	s.Assert().Equal(1, len(qualifyingActions), "must find one qualifying actions")
	s.Assert().Equal(int64(10), qualifyingActions[0].Shares, "shares should be 10")
}

func getOperatorBech32AddressForTestValidator() sdk.AccAddress {
	validatorAddress, _ := sdk.ValAddressFromBech32(getTestValidators(0, 1)[0].OperatorAddress)
	bz, err := sdk.GetFromBech32(validatorAddress.String(), sdk.GetConfig().GetBech32ValidatorAddrPrefix())
	if err != nil {
		panic(err)
	}
	accountAddr := sdk.AccAddress(bz)
	return accountAddr
}

func (s *KeeperTestSuite) TestDetectQualifyingActionsWith1VotingQualifyingActionDelegationNotMet() {
	SetupEventHistoryWithVotes(s, "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h")
	s.app.RewardKeeper.SetStakingKeeper(MockStakingKeeper{})
	minDelegation := sdk.NewInt64Coin("nhash", 4)

	rewardProgram := types.NewRewardProgram(
		"title",
		"description",
		1,
		"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
		sdk.NewInt64Coin("hotdog", 10000),
		sdk.NewInt64Coin("hotdog", 10000),
		time.Now(),
		5,
		5,
		0,
		0,
		[]types.QualifyingAction{
			{
				Type: &types.QualifyingAction_Vote{
					Vote: &types.ActionVote{
						MinimumActions:          0,
						MaximumActions:          1,
						MinimumDelegationAmount: minDelegation,
					},
				},
			},
		},
	)
	rewardProgram.CurrentClaimPeriod = 1
	qualifyingActions, err := s.app.RewardKeeper.DetectQualifyingActions(s.ctx, &rewardProgram)
	s.Assert().NoError(err, "must not error")
	s.Assert().Equal(0, len(qualifyingActions), "must find zero qualifying actions")
}

func (s *KeeperTestSuite) TestDetectQualifyingActionsWith1VotingNoQualifyingAction() {
	SetupEventHistoryWithDelegates(s)
	s.app.RewardKeeper.SetStakingKeeper(MockStakingKeeper{})
	minDelegation := sdk.NewInt64Coin("nhash", 0)

	rewardProgram := types.NewRewardProgram(
		"title",
		"description",
		1,
		"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
		sdk.NewInt64Coin("hotdog", 10000),
		sdk.NewInt64Coin("hotdog", 10000),
		time.Now(),
		5,
		5,
		0,
		0,
		[]types.QualifyingAction{
			{
				Type: &types.QualifyingAction_Vote{
					Vote: &types.ActionVote{
						MinimumActions:          0,
						MaximumActions:          1,
						MinimumDelegationAmount: minDelegation,
					},
				},
			},
		},
	)
	rewardProgram.CurrentClaimPeriod = 1
	qualifyingActions, err := s.app.RewardKeeper.DetectQualifyingActions(s.ctx, &rewardProgram)
	s.Assert().NoError(err, "must not error")
	s.Assert().Equal(0, len(qualifyingActions), "must find one qualifying actions")
}

func (s *KeeperTestSuite) TestDetectQualifyingActionsWith1VotingDelegateQualifyingAction() {
	SetupEventHistoryWithDelegates(s)
	s.app.RewardKeeper.SetStakingKeeper(MockStakingKeeper{})
	minDelegation := sdk.NewInt64Coin("nhash", 0)
	maxDelegation := sdk.NewInt64Coin("nhash", 10)

	rewardProgram := types.NewRewardProgram(
		"title",
		"description",
		1,
		"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
		sdk.NewInt64Coin("hotdog", 10000),
		sdk.NewInt64Coin("hotdog", 10000),
		time.Now(),
		5,
		5,
		0,
		0,
		[]types.QualifyingAction{
			{
				Type: &types.QualifyingAction_Vote{
					Vote: &types.ActionVote{
						MinimumActions:          0,
						MaximumActions:          1,
						MinimumDelegationAmount: minDelegation,
					},
				},
			},
			{
				Type: &types.QualifyingAction_Delegate{
					Delegate: &types.ActionDelegate{
						MinimumActions:               0,
						MaximumActions:               1,
						MinimumDelegationAmount:      &minDelegation,
						MaximumDelegationAmount:      &maxDelegation,
						MinimumActiveStakePercentile: sdkmath.LegacyNewDecWithPrec(0, 0),
						MaximumActiveStakePercentile: sdkmath.LegacyNewDecWithPrec(1, 0),
					},
				},
			},
		},
	)
	rewardProgram.CurrentClaimPeriod = 1
	qualifyingActions, err := s.app.RewardKeeper.DetectQualifyingActions(s.ctx, &rewardProgram)
	s.Assert().NoError(err, "must not error")
	s.Assert().Equal(1, len(qualifyingActions), "must find one qualifying actions")
}

func (s *KeeperTestSuite) TestDetectQualifyingActionsWith1Voting1DelegateQualifyingAction() {
	SetupEventHistoryWithDelegates(s)
	SetupEventHistoryWithVotes(s, "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h")
	s.app.RewardKeeper.SetStakingKeeper(MockStakingKeeper{})
	minDelegation := sdk.NewInt64Coin("nhash", 0)
	maxDelegation := sdk.NewInt64Coin("nhash", 10)

	rewardProgram := types.NewRewardProgram(
		"title",
		"description",
		1,
		"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
		sdk.NewInt64Coin("hotdog", 10000),
		sdk.NewInt64Coin("hotdog", 10000),
		time.Now(),
		5,
		5,
		0,
		0,
		[]types.QualifyingAction{
			{
				Type: &types.QualifyingAction_Vote{
					Vote: &types.ActionVote{
						MinimumActions:          0,
						MaximumActions:          1,
						MinimumDelegationAmount: minDelegation,
					},
				},
			},
			{
				Type: &types.QualifyingAction_Delegate{
					Delegate: &types.ActionDelegate{
						MinimumActions:               0,
						MaximumActions:               1,
						MinimumDelegationAmount:      &minDelegation,
						MaximumDelegationAmount:      &maxDelegation,
						MinimumActiveStakePercentile: sdkmath.LegacyNewDecWithPrec(0, 0),
						MaximumActiveStakePercentile: sdkmath.LegacyNewDecWithPrec(1, 0),
					},
				},
			},
		},
	)
	rewardProgram.CurrentClaimPeriod = 1
	qualifyingActions, err := s.app.RewardKeeper.DetectQualifyingActions(s.ctx, &rewardProgram)
	s.Assert().NoError(err, "must not error")
	s.Assert().Equal(2, len(qualifyingActions), "must find one qualifying actions")
}

func (s *KeeperTestSuite) TestGetAccountKeeper() {
	s.Assert().NotNil(s.app.RewardKeeper.GetAccountKeeper())
}
