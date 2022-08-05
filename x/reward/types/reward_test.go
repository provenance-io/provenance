package types

import (
	"testing"
	time "time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type RewardTypesTestSuite struct {
	suite.Suite
	ctx sdk.Context
}

func TestRewardTypesTestSuite(t *testing.T) {
	suite.Run(t, new(RewardTypesTestSuite))
}

func (s *RewardTypesTestSuite) SetupSuite() {
	s.T().Parallel()
}

func (s *RewardTypesTestSuite) TestRewardProgramValidateBasic() {
	now := time.Now().UTC()
	lTitle := make([]byte, MaxTitleLength+1)
	longTitle := string(lTitle)
	lDescription := make([]byte, MaxDescriptionLength+1)
	longDescription := string(lDescription)
	minDelegation := sdk.NewInt64Coin("nhash", 4)
	maxDelegation := sdk.NewInt64Coin("nhash", 40)
	tests := []struct {
		name          string
		rewardProgram RewardProgram
		want          string
	}{
		{
			"valid",
			NewRewardProgram(
				"title",
				"description",
				1,
				"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
				sdk.NewInt64Coin("nhash", 1),
				sdk.NewInt64Coin("nhash", 2),
				now,
				1,
				1,
				0,
				1,
				[]QualifyingAction{
					{
						Type: &QualifyingAction_Vote{
							Vote: &ActionVote{
								MinimumActions:          0,
								MaximumActions:          1,
								MinimumDelegationAmount: minDelegation,
							},
						},
					},
					{
						Type: &QualifyingAction_Delegate{
							Delegate: &ActionDelegate{
								MinimumActions:               0,
								MaximumActions:               1,
								MinimumDelegationAmount:      &minDelegation,
								MaximumDelegationAmount:      &maxDelegation,
								MinimumActiveStakePercentile: sdk.NewDecWithPrec(0, 0),
								MaximumActiveStakePercentile: sdk.NewDecWithPrec(1, 0),
							},
						},
					},
				},
			),
			"",
		},
		{
			"invalid - title is blank",
			NewRewardProgram(
				"",
				"description",
				1,
				"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
				sdk.NewInt64Coin("nhash", 1),
				sdk.NewInt64Coin("nhash", 2),
				now,
				1,
				1,
				0,
				1,
				[]QualifyingAction{},
			),
			"reward program title cannot be blank",
		},
		{
			"invalid - title is too long",
			NewRewardProgram(
				longTitle,
				"description",
				1,
				"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
				sdk.NewInt64Coin("nhash", 1),
				sdk.NewInt64Coin("nhash", 2),
				now,
				1,
				1,
				0,
				1,
				[]QualifyingAction{},
			),
			"reward program title is longer than max length of 140",
		},
		{
			"invalid - description is blank",
			NewRewardProgram(
				"title",
				"",
				1,
				"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
				sdk.NewInt64Coin("nhash", 1),
				sdk.NewInt64Coin("nhash", 2),
				now,
				1,
				1,
				0,
				1,
				[]QualifyingAction{},
			),
			"reward program description cannot be blank",
		},
		{
			"invalid - description is too long",
			NewRewardProgram(
				"title",
				longDescription,
				1,
				"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
				sdk.NewInt64Coin("nhash", 1),
				sdk.NewInt64Coin("nhash", 2),
				now,
				1,
				1,
				0,
				1,
				[]QualifyingAction{},
			),
			"reward program description is longer than max length of 10000",
		},
		{
			"invalid - reward id must be greater than 0",
			NewRewardProgram(
				"title",
				"description",
				0,
				"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
				sdk.NewInt64Coin("nhash", 1),
				sdk.NewInt64Coin("nhash", 2),
				now,
				1,
				1,
				0,
				1,
				[]QualifyingAction{},
			),
			"reward program id must be larger than 0",
		},
		{
			"invalid - invalid address",
			NewRewardProgram(
				"title",
				"description",
				1,
				"invalid",
				sdk.NewInt64Coin("nhash", 1),
				sdk.NewInt64Coin("nhash", 2),
				now,
				1,
				1,
				0,
				1,
				[]QualifyingAction{},
			),
			"invalid address for rewards program distribution from address: decoding bech32 failed: invalid bech32 string length 7",
		},
		{
			"invalid - coin must be positive",
			NewRewardProgram(
				"title",
				"description",
				1,
				"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
				sdk.NewInt64Coin("nhash", 0),
				sdk.NewInt64Coin("nhash", 2),
				now,
				1,
				1,
				0,
				1,
				[]QualifyingAction{},
			),
			"reward program requires coins: 0nhash",
		},
		{
			"invalid - max reward must be positive",
			NewRewardProgram(
				"title",
				"description",
				1,
				"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
				sdk.NewInt64Coin("nhash", 1),
				sdk.NewInt64Coin("nhash", 0),
				now,
				1,
				1,
				0,
				1,
				[]QualifyingAction{},
			),
			"reward program requires positive max reward by address: 0nhash",
		},
		{
			"invalid - number of claim periods must be larger than 0",
			NewRewardProgram(
				"title",
				"description",
				1,
				"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
				sdk.NewInt64Coin("nhash", 1),
				sdk.NewInt64Coin("nhash", 2),
				now,
				1,
				0,
				0,
				1,
				[]QualifyingAction{},
			),
			"reward program number of claim periods must be larger than 0",
		},
	}

	for _, tt := range tests {
		tt := tt
		s.T().Run(tt.name, func(t *testing.T) {
			err := tt.rewardProgram.Validate()
			if err != nil {
				assert.Equal(t, tt.want, err.Error())
			} else if len(tt.want) > 0 {
				t.Errorf("RewardProgram ValidateBasic error = nil, expected: %s", tt.want)
			}
		})
	}
}

func (s *RewardTypesTestSuite) TestEpochRewardDistributionValidateBasic() {
	tests := []struct {
		name          string
		rewardProgram ClaimPeriodRewardDistribution
		want          string
	}{
		{
			"invalid -  claim period id",
			NewClaimPeriodRewardDistribution(0, 1, sdk.NewInt64Coin("nhash", 100), sdk.NewInt64Coin("nhash", 100), 0, false),
			"claim reward distribution has invalid claim id",
		},
		{
			"invalid - reward program id",
			NewClaimPeriodRewardDistribution(1, 0, sdk.NewInt64Coin("nhash", 100), sdk.NewInt64Coin("nhash", 100), 0, false),
			"claim reward distribution must have a valid reward program id",
		},
		{
			"invalid - total rewards needs to be positive",
			NewClaimPeriodRewardDistribution(1, 1, sdk.NewInt64Coin("nhash", 100), sdk.NewInt64Coin("nhash", 0), 0, false),
			"",
		},
		{
			"invalid - total rewards needs to be positive",
			NewClaimPeriodRewardDistribution(1, 1, sdk.NewInt64Coin("nhash", 0), sdk.NewInt64Coin("nhash", 100), 0, false),
			"claim reward distribution must have a reward pool which is positive",
		},
		{
			"should succeed validate basic",
			NewClaimPeriodRewardDistribution(1, 1, sdk.NewInt64Coin("nhash", 1), sdk.NewInt64Coin("nhash", 1), 0, false),
			"",
		},
	}

	for _, tt := range tests {
		tt := tt
		s.T().Run(tt.name, func(t *testing.T) {
			err := tt.rewardProgram.ValidateBasic()
			if err != nil {
				assert.Equal(t, tt.want, err.Error())
			} else if len(tt.want) > 0 {
				t.Errorf("EpochRewardDistribution ValidateBasic error = nil, expected: %s", tt.want)
			}
		})
	}
}

func (s *RewardTypesTestSuite) TesRewardAccountStateValidateBasic() {
	tests := []struct {
		name               string
		RewardAccountState RewardAccountState
		want               string
	}{
		{
			"valid",
			NewRewardAccountState(
				1,
				2,
				"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
				0,
				map[string]uint64{},
			),
			"",
		},
		{
			"invalid reward id",
			NewRewardAccountState(
				0,
				2,
				"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
				0,
				map[string]uint64{},
			),
			"reward program id must be greater than 0",
		},
		{
			"invalid - address is incorrect",
			NewRewardAccountState(
				1,
				2,
				"test",
				0,
				map[string]uint64{},
			),
			"invalid address for reward program balance: decoding bech32 failed: invalid bech32 string length 7",
		},
	}

	for _, tt := range tests {
		tt := tt
		s.T().Run(tt.name, func(t *testing.T) {
			err := tt.RewardAccountState.ValidateBasic()
			if err != nil {
				assert.Equal(t, tt.want, err.Error())
			} else if len(tt.want) > 0 {
				t.Errorf("RewardAccountState ValidateBasic error = nil, expected: %s", tt.want)
			}
		})
	}
}

func (s *RewardTypesTestSuite) TestActionDelegateCreation() {
	minDelegation := sdk.NewInt64Coin("nhash", 4)
	maxDelegation := sdk.NewInt64Coin("nhash", 40)

	action := ActionDelegate{
		MinimumActions:               0,
		MaximumActions:               1,
		MinimumDelegationAmount:      &minDelegation,
		MaximumDelegationAmount:      &maxDelegation,
		MinimumActiveStakePercentile: sdk.NewDecWithPrec(0, 0),
		MaximumActiveStakePercentile: sdk.NewDecWithPrec(1, 0),
	}
	s.Assert().Nil(action.Validate(), "validate basic must have no error")
	s.Assert().Equal("ActionDelegate", action.ActionType(), "must have appropriate action type")
	s.Assert().Equal(true, action.GetBuilder() != nil, "must have appropriate builder")
}

func (s *RewardTypesTestSuite) TestActionVoteCreation() {
	minDelegation := sdk.NewInt64Coin("nhash", 4)

	action := ActionVote{
		MinimumActions:          0,
		MaximumActions:          1,
		MinimumDelegationAmount: minDelegation,
	}
	s.Assert().Nil(action.Validate(), "validate basic must have no error")
	s.Assert().Equal("ActionVote", action.ActionType(), "must have appropriate action type")
	s.Assert().Equal(true, action.GetBuilder() != nil, "must have appropriate builder")
}

func (s *RewardTypesTestSuite) TestActionTransferCreation() {
	minDelegation := sdk.NewInt64Coin("nhash", 4)

	action := ActionTransfer{
		MinimumActions:          0,
		MaximumActions:          1,
		MinimumDelegationAmount: minDelegation,
	}
	s.Assert().Nil(action.Validate(), "validate basic must have no error")
	s.Assert().Equal("ActionTransfer", action.ActionType(), "must have appropriate action type")
	s.Assert().Equal(true, action.GetBuilder() != nil, "must have appropriate builder")
}

func (s *RewardTypesTestSuite) TestNewEventCriteria() {
	event1 := ABCIEvent{
		Type: "type1",
		Attributes: map[string][]byte{
			"attribute1": []byte("value1"),
		},
	}
	event2 := ABCIEvent{
		Type: "type2",
		Attributes: map[string][]byte{
			"attribute3": []byte("value3"),
			"attribute4": []byte("value4"),
		},
	}
	events := []ABCIEvent{event1, event2}

	criteria := NewEventCriteria(events)
	s.Assert().Equal(criteria.Events["type1"].Type, "type1", "element must exist and type must match")
	s.Assert().Equal(criteria.Events["type1"].Attributes["attribute1"], []byte("value1"), "the attribute should still exist and contain the correct value")
	s.Assert().Equal(criteria.Events["type2"].Type, "type2", "element must exist and type must match")
	s.Assert().Equal(criteria.Events["type2"].Attributes["attribute3"], []byte("value3"), "the attribute should still exist and contain the correct value")
	s.Assert().Equal(criteria.Events["type2"].Attributes["attribute4"], []byte("value4"), "the attribute should still exist and contain the correct value")

	nilCriteria := NewEventCriteria(nil)
	s.Assert().Nil(nilCriteria.Events, "events should be nil")
}

func (s *RewardTypesTestSuite) TestMatchesAttribute() {
	event := ABCIEvent{
		Type: "type1",
		Attributes: map[string][]byte{
			"attribute1": []byte("value1"),
			"attribute3": nil,
		},
	}
	s.Assert().True(event.MatchesAttribute("attribute1", []byte("value1")), "attribute and value must match")
	s.Assert().True(event.MatchesAttribute("attribute3", []byte("blah")), "only attribute needs to match")
	s.Assert().False(event.MatchesAttribute("attribute2", []byte("blah")), "must fail when attribute doesn't exist")
	s.Assert().False(event.MatchesAttribute("attribute1", []byte("value2")), "must fail when attribute's value doesn't match")
}

func (s *RewardTypesTestSuite) TestMatchesEvent() {
	event1 := ABCIEvent{
		Type: "type1",
		Attributes: map[string][]byte{
			"attribute1": []byte("value1"),
		},
	}
	event2 := ABCIEvent{
		Type: "type2",
		Attributes: map[string][]byte{
			"attribute3": []byte("value3"),
			"attribute4": []byte("value4"),
		},
	}

	nilCriteria := NewEventCriteria(nil)
	s.Assert().True(nilCriteria.MatchesEvent("blah"), "a nil criteria must match everything")

	events := []ABCIEvent{}
	criteria := NewEventCriteria(events)
	s.Assert().True(criteria.MatchesEvent("blah"), "a empty criteria must match everything")

	events = []ABCIEvent{event1, event2}
	criteria = NewEventCriteria(events)

	s.Assert().True(criteria.MatchesEvent("type1"), "criteria must match against valid event type")
	s.Assert().False(criteria.MatchesEvent("blah"), "criteria must fail against invalid event type")
}

func (s *RewardTypesTestSuite) TestDateOnOrAfter() {
	// compareConstant's use case is blockTime
	compareConstant, err := time.Parse(time.RFC3339, "2022-07-21T13:00:10Z")
	s.Require().NoError(err)

	past, err := time.Parse(time.RFC3339, "2022-07-21T13:00:09Z")
	s.Require().NoError(err)
	future, err := time.Parse(time.RFC3339, "2022-07-21T13:00:11Z")
	s.Require().NoError(err)

	s.Assert().False(TimeOnOrAfter(compareConstant, past))
	s.Assert().True(TimeOnOrAfter(compareConstant, future))
	s.Assert().True(TimeOnOrAfter(compareConstant, compareConstant))
}

func (s *RewardTypesTestSuite) TestCalculateExpectedEndTime() {
	compareConstant, err := time.Parse(time.RFC3339, "2022-01-21T15:00:00Z")
	s.Require().NoError(err)

	result := CalculateExpectedEndTime(compareConstant, uint64(DayInSeconds), 7)
	expected, err := time.Parse(time.RFC3339, "2022-01-28T15:00:00Z")
	s.Require().NoError(err)
	s.Assert().True(result.Equal(expected))

	result = CalculateExpectedEndTime(compareConstant, uint64(DayInSeconds), 0)
	s.Require().NoError(err)
	s.Assert().True(result.Equal(compareConstant))

}

func (s *RewardTypesTestSuite) TestCalculateEndTimeMax() {
	compareConstant, err := time.Parse(time.RFC3339, "2022-01-21T15:00:00Z")
	s.Require().NoError(err)

	result := CalculateEndTimeMax(compareConstant, uint64(DayInSeconds), 7, 0)
	expected, err := time.Parse(time.RFC3339, "2022-01-28T15:00:00Z")
	s.Require().NoError(err)
	s.Assert().True(result.Equal(expected))

	result = CalculateEndTimeMax(compareConstant, uint64(DayInSeconds), 7, 1)
	expected, err = time.Parse(time.RFC3339, "2022-01-29T15:00:00Z")
	s.Require().NoError(err)
	s.Assert().True(result.Equal(expected))

	result = CalculateEndTimeMax(compareConstant, uint64(DayInSeconds), 0, 0)
	s.Require().NoError(err)
	s.Assert().True(result.Equal(compareConstant))

}

func (s *RewardTypesTestSuite) TestIsStarting() {
	now := time.Now().UTC()
	program := NewRewardProgram(
		"title",
		"description",
		1,
		"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
		sdk.NewInt64Coin("nhash", 1),
		sdk.NewInt64Coin("nhash", 2),
		now,
		3600,
		1,
		0,
		1,
		[]QualifyingAction{},
	)
	s.ctx = s.ctx.WithBlockTime(now.Add(-1000))
	s.Assert().False(program.IsStarting(s.ctx))

	s.ctx = s.ctx.WithBlockTime(now.Add(2010))
	s.Assert().True(program.IsStarting(s.ctx))

}

func (s *RewardTypesTestSuite) TestIsExpiring() {
	now := time.Now().UTC()
	program := NewRewardProgram(
		"title",
		"description",
		1,
		"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
		sdk.NewInt64Coin("nhash", 1),
		sdk.NewInt64Coin("nhash", 2),
		now,
		3600,
		1,
		0,
		100,
		[]QualifyingAction{},
	)

	s.ctx = s.ctx.WithBlockTime(now)
	s.Assert().False(program.IsExpiring(s.ctx))

	program.ActualProgramEndTime = now
	program.State = RewardProgram_STATE_FINISHED
	s.Assert().False(program.IsExpiring(s.ctx))

	program.ExpirationOffset = 0
	s.Assert().True(program.IsExpiring(s.ctx))

}

func (s *RewardTypesTestSuite) TestIsEnding() {
	now := time.Now().UTC()
	program := NewRewardProgram(
		"title",
		"description",
		1,
		"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
		sdk.NewInt64Coin("nhash", 1),
		sdk.NewInt64Coin("nhash", 2),
		now,
		3600,
		1,
		1,
		100,
		[]QualifyingAction{},
	)

	s.ctx = s.ctx.WithBlockTime(now)
	s.Assert().False(program.IsEnding(s.ctx, sdk.NewInt64Coin("nhash", 100_000_000_001)))

	program.State = RewardProgram_STATE_STARTED
	program.ProgramEndTimeMax = now.Add(100000)
	s.Assert().False(program.IsEnding(s.ctx, sdk.NewInt64Coin("nhash", 100_000_000_000)))
	s.Assert().False(program.IsEnding(s.ctx, sdk.NewInt64Coin("nhash", 100_000_000_001)))
	s.Assert().True(program.IsEnding(s.ctx, sdk.NewInt64Coin("nhash", 10_000_000_000)))

	program.ProgramEndTimeMax = now.Add(-100000)
	s.Assert().True(program.IsEnding(s.ctx, sdk.NewInt64Coin("nhash", 100_000_000_001)))

}

func (s *RewardTypesTestSuite) TestIsEndingClaimPeriod() {
	now := time.Now().UTC()
	program := NewRewardProgram(
		"title",
		"description",
		1,
		"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
		sdk.NewInt64Coin("nhash", 1),
		sdk.NewInt64Coin("nhash", 2),
		now,
		3600,
		1,
		0,
		1,
		[]QualifyingAction{},
	)
	s.ctx = s.ctx.WithBlockTime(now.Add(100000))

	// RewardProgram_PENDING and blockTime after ClaimPeriodEndTime
	program.State = RewardProgram_STATE_PENDING
	s.Assert().False(program.IsEndingClaimPeriod(s.ctx))

	s.ctx = s.ctx.WithBlockTime(now.Add(-100000))
	// RewardProgram_STARTED and blockTime after ClaimPeriodEndTime
	program.State = RewardProgram_STATE_STARTED
	s.Assert().True(program.IsEndingClaimPeriod(s.ctx))

	// RewardProgram_FINISHED and blockTime after ClaimPeriodEndTime
	program.State = RewardProgram_STATE_FINISHED
	s.Assert().False(program.IsEndingClaimPeriod(s.ctx))

	// RewardProgram_EXPIRED and blockTime after ClaimPeriodEndTime
	program.State = RewardProgram_STATE_EXPIRED
	s.Assert().False(program.IsEndingClaimPeriod(s.ctx))

}
