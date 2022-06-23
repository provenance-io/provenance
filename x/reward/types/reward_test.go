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
				sdk.NewInt64Coin("jackthecat", 1),
				sdk.NewInt64Coin("jackthecat", 2),
				now,
				60*24,
				1,
				[]QualifyingAction{},
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
				sdk.NewInt64Coin("jackthecat", 1),
				sdk.NewInt64Coin("jackthecat", 2),
				now,
				60*24,
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
				sdk.NewInt64Coin("jackthecat", 1),
				sdk.NewInt64Coin("jackthecat", 2),
				now,
				60*24,
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
				sdk.NewInt64Coin("jackthecat", 1),
				sdk.NewInt64Coin("jackthecat", 2),
				now,
				60*24,
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
				sdk.NewInt64Coin("jackthecat", 1),
				sdk.NewInt64Coin("jackthecat", 2),
				now,
				60*24,
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
				sdk.NewInt64Coin("jackthecat", 1),
				sdk.NewInt64Coin("jackthecat", 2),
				now,
				60*24,
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
				sdk.NewInt64Coin("jackthecat", 1),
				sdk.NewInt64Coin("jackthecat", 2),
				now,
				60*24,
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
				sdk.NewInt64Coin("jackthecat", 0),
				sdk.NewInt64Coin("jackthecat", 2),
				now,
				60*24,
				1,
				[]QualifyingAction{},
			),
			"reward program requires coins: 0jackthecat",
		},
		{
			"invalid - max reward must be positive",
			NewRewardProgram(
				"title",
				"description",
				1,
				"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
				sdk.NewInt64Coin("jackthecat", 1),
				sdk.NewInt64Coin("jackthecat", 0),
				now,
				60*24,
				1,
				[]QualifyingAction{},
			),
			"reward program requires positive max reward by address: 0jackthecat",
		},
		{
			"invalid - number of sub periods must be larger than 0",
			NewRewardProgram(
				"title",
				"description",
				1,
				"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
				sdk.NewInt64Coin("jackthecat", 1),
				sdk.NewInt64Coin("jackthecat", 2),
				now,
				60*24,
				0,
				[]QualifyingAction{},
			),
			"reward program number of sub periods must be larger than 0",
		},
	}

	for _, tt := range tests {
		tt := tt
		s.T().Run(tt.name, func(t *testing.T) {
			err := tt.rewardProgram.ValidateBasic()
			if err != nil {
				assert.Equal(t, tt.want, err.Error())
			} else if len(tt.want) > 0 {
				t.Errorf("RewardProgram ValidateBasic error = nil, expected: %s", tt.want)
			}
		})
	}
}

func (s *RewardTypesTestSuite) TestRewardProgramBalanceValidateBasic() {
	tests := []struct {
		name                 string
		rewardProgramBalance RewardProgramBalance
		want                 string
	}{
		{
			"valid",
			NewRewardProgramBalance(
				1,
				sdk.NewInt64Coin("jackthecat", 1),
			),
			"",
		},
		{
			"invalid - reward id is 0",
			NewRewardProgramBalance(
				0,
				sdk.NewInt64Coin("jackthecat", 1),
			),
			"reward program id must be larger than 0",
		},
	}

	for _, tt := range tests {
		tt := tt
		s.T().Run(tt.name, func(t *testing.T) {
			err := tt.rewardProgramBalance.ValidateBasic()
			if err != nil {
				assert.Equal(t, tt.want, err.Error())
			} else if len(tt.want) > 0 {
				t.Errorf("RewardProgramBalance ValidateBasic error = nil, expected: %s", tt.want)
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
			NewClaimPeriodRewardDistribution(0, 1, sdk.NewInt64Coin("jackthecat", 100), sdk.NewInt64Coin("jackthecat", 100), 0, false),
			"claim reward distribution has invalid claim id",
		},
		{
			"invalid - reward program id",
			NewClaimPeriodRewardDistribution(1, 0, sdk.NewInt64Coin("jackthecat", 100), sdk.NewInt64Coin("jackthecat", 100), 0, false),
			"claim reward distribution must have a valid reward program id",
		},
		{
			"invalid - total rewards needs to be positive",
			NewClaimPeriodRewardDistribution(1, 1, sdk.NewInt64Coin("jackthecat", 100), sdk.NewInt64Coin("jackthecat", 0), 0, false),
			"claim reward distribution must have a total reward pool",
		},
		{
			"invalid - total rewards needs to be positive",
			NewClaimPeriodRewardDistribution(1, 1, sdk.NewInt64Coin("jackthecat", 0), sdk.NewInt64Coin("jackthecat", 100), 0, false),
			"claim reward distribution must have a reward pool",
		},
		{
			"should succeed validate basic",
			NewClaimPeriodRewardDistribution(1, 1, sdk.NewInt64Coin("jackthecat", 1), sdk.NewInt64Coin("jackthecat", 1), 0, false),
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

func (s *RewardTypesTestSuite) TesAccountStateValidateBasic() {
	tests := []struct {
		name         string
		accountState AccountState
		want         string
	}{
		{
			"valid",
			NewAccountState(
				1,
				2,
				"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
			),
			"",
		},
		{
			"invalid reward id",
			NewAccountState(
				0,
				2,
				"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
			),
			"reward program id must be greater than 0",
		},
		{
			"invalid - address is incorrect",
			NewAccountState(
				1,
				2,
				"test",
			),
			"invalid address for reward program balance: decoding bech32 failed: invalid bech32 string length 7",
		},
	}

	for _, tt := range tests {
		tt := tt
		s.T().Run(tt.name, func(t *testing.T) {
			err := tt.accountState.ValidateBasic()
			if err != nil {
				assert.Equal(t, tt.want, err.Error())
			} else if len(tt.want) > 0 {
				t.Errorf("AccountState ValidateBasic error = nil, expected: %s", tt.want)
			}
		})
	}
}

func (s *RewardTypesTestSuite) TestActionDelegateCreation() {
	action := NewActionDelegate()
	s.Assert().Nil(action.ValidateBasic(), "validate basic must have no error")
	s.Assert().Equal("ActionDelegate", action.ActionType(), "must have appropriate action type")
	//s.Assert().Equal("message", action.GetEventCriteria().EventType, "must have correct event type criteria")
	//s.Assert().Equal("staking", action.GetEventCriteria().Attributes, "must have correct attribute criteria")
	//s.Assert().Equal("sender", action.GetEventCriteria().AttributeValue, "must have correct attribute value criteria")
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
