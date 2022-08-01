package types

import (
	"testing"
	time "time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type RewardMsgTypesTestSuite struct {
	suite.Suite
}

func TestRewardMsgTypesTestSuite(t *testing.T) {
	suite.Run(t, new(RewardMsgTypesTestSuite))
}

func (s *RewardMsgTypesTestSuite) SetupSuite() {
	s.T().Parallel()
}

func (s *RewardMsgTypesTestSuite) TestMsgCreateRewardProgramRequestValidateBasic() {
	dateTime := time.Date(2024, 12, 2, 0, 0, 0, 0, time.UTC)
	lTitle := make([]byte, MaxTitleLength+1)
	longTitle := string(lTitle)
	lDescription := make([]byte, MaxDescriptionLength+1)
	longDescription := string(lDescription)
	minimumDelegation := sdk.NewInt64Coin("nhash", 0)
	maximumDelegation := sdk.NewInt64Coin("nhash", 100)
	qualifyingActions := []QualifyingAction{
		{
			Type: &QualifyingAction_Delegate{
				Delegate: &ActionDelegate{
					MinimumActions:               0,
					MaximumActions:               0,
					MinimumDelegationAmount:      &minimumDelegation,
					MaximumDelegationAmount:      &maximumDelegation,
					MinimumActiveStakePercentile: sdk.NewDecWithPrec(0, 0),
					MaximumActiveStakePercentile: sdk.NewDecWithPrec(1, 0),
				},
			},
		},
	}
	tests := []struct {
		name                          string
		msgCreateRewardProgramRequest MsgCreateRewardProgramRequest
		want                          string
	}{
		{
			"valid",
			*NewMsgCreateRewardProgramRequest(
				"title",
				"description",
				"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
				sdk.NewInt64Coin("jackthecat", 1),
				sdk.NewInt64Coin("jackthecat", 1),
				dateTime,
				1,
				1,
				1,
				1,
				qualifyingActions,
			),
			"",
		},
		{
			"invalid - blank title",
			*NewMsgCreateRewardProgramRequest(
				"",
				"description",
				"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
				sdk.NewInt64Coin("jackthecat", 1),
				sdk.NewInt64Coin("jackthecat", 1),
				dateTime,
				1,
				1,
				1,
				1,
				qualifyingActions,
			),
			"reward program title cannot be blank",
		},
		{
			"invalid - title too long",
			*NewMsgCreateRewardProgramRequest(
				longTitle,
				"description",
				"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
				sdk.NewInt64Coin("jackthecat", 1),
				sdk.NewInt64Coin("jackthecat", 1),
				dateTime,
				1,
				1,
				1,
				1,
				qualifyingActions,
			),
			"reward program title is longer than max length of 140",
		},
		{
			"invalid - description is blank",
			*NewMsgCreateRewardProgramRequest(
				"title",
				"",
				"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
				sdk.NewInt64Coin("jackthecat", 1),
				sdk.NewInt64Coin("jackthecat", 1),
				dateTime,
				1,
				1,
				1,
				1,
				qualifyingActions,
			),
			"reward program description cannot be blank",
		},
		{
			"invalid - title is too long",
			*NewMsgCreateRewardProgramRequest(
				"title",
				longDescription,
				"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
				sdk.NewInt64Coin("jackthecat", 1),
				sdk.NewInt64Coin("jackthecat", 1),
				dateTime,
				1,
				1,
				1,
				1,
				qualifyingActions,
			),
			"reward program description is longer than max length of 10000",
		},
		{
			"invalid - address is incorrect",
			*NewMsgCreateRewardProgramRequest(
				"title",
				"description",
				"invalid",
				sdk.NewInt64Coin("jackthecat", 1),
				sdk.NewInt64Coin("jackthecat", 1),
				dateTime,
				1,
				1,
				1,
				1,
				qualifyingActions,
			),
			"invalid address for rewards program distribution from address: decoding bech32 failed: invalid bech32 string length 7",
		},
		{
			"invalid - coin is not positive",
			*NewMsgCreateRewardProgramRequest(
				"title",
				"description",
				"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
				sdk.NewInt64Coin("jackthecat", 0),
				sdk.NewInt64Coin("jackthecat", 2),
				dateTime,
				1,
				1,
				1,
				1,
				qualifyingActions,
			),
			"reward program requires total reward pool to be positive: 0jackthecat",
		},
		{
			"invalid - max coin per address is invalid",
			*NewMsgCreateRewardProgramRequest(
				"title",
				"description",
				"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
				sdk.NewInt64Coin("jackthecat", 1),
				sdk.NewInt64Coin("jackthecat", 0),
				dateTime,
				1,
				1,
				1,
				1,
				qualifyingActions,
			),
			"reward program requires positive max reward by address: 0jackthecat",
		},
		{
			"invalid - denoms differ",
			*NewMsgCreateRewardProgramRequest(
				"title",
				"description",
				"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
				sdk.NewInt64Coin("jackthecat", 1),
				sdk.NewInt64Coin("hotdog", 1),
				dateTime,
				1,
				1,
				1,
				1,
				qualifyingActions,
			),
			"coin denoms differ jackthecat : hotdog",
		},
		{
			"invalid - reward per address is greater than pool",
			*NewMsgCreateRewardProgramRequest(
				"title",
				"description",
				"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
				sdk.NewInt64Coin("jackthecat", 1),
				sdk.NewInt64Coin("jackthecat", 2),
				dateTime,
				0,
				1,
				1,
				1,
				qualifyingActions,
			),
			"max claims per address cannot be larger than pool 2 : 1",
		},
		{
			"invalid - number of claim periods is 0",
			*NewMsgCreateRewardProgramRequest(
				"title",
				"description",
				"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
				sdk.NewInt64Coin("jackthecat", 1),
				sdk.NewInt64Coin("jackthecat", 1),
				dateTime,
				1,
				0,
				1,
				1,
				qualifyingActions,
			),
			"claim periods (1), claim period days (0), and expire days (1) must be larger than 0",
		},
		{
			"invalid - number of claim periods is 0",
			*NewMsgCreateRewardProgramRequest(
				"title",
				"description",
				"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
				sdk.NewInt64Coin("jackthecat", 1),
				sdk.NewInt64Coin("jackthecat", 1),
				dateTime,
				1,
				1,
				1,
				0,
				qualifyingActions,
			),
			"claim periods (1), claim period days (1), and expire days (0) must be larger than 0",
		},
		{
			"invalid - reward has no qualifying actions",
			*NewMsgCreateRewardProgramRequest(
				"title",
				"description",
				"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
				sdk.NewInt64Coin("jackthecat", 1),
				sdk.NewInt64Coin("jackthecat", 1),
				dateTime,
				4,
				4,
				1,
				1,
				[]QualifyingAction{},
			),
			"reward program must contain qualifying actions",
		},
	}
	for _, tt := range tests {
		tt := tt
		s.T().Run(tt.name, func(t *testing.T) {
			err := tt.msgCreateRewardProgramRequest.ValidateBasic()
			if err != nil {
				assert.Equal(t, tt.want, err.Error())
			} else if len(tt.want) > 0 {
				t.Errorf("MsgCreateRewardProgramRequest ValidateBasic error = nil, expected: %s", tt.want)
			}
		})
	}
}

func (s *RewardMsgTypesTestSuite) TestMsgEndRewardProgramRequestValidateBasic() {
	tests := []struct {
		name                          string
		msgCreateRewardProgramRequest MsgEndRewardProgramRequest
		want                          string
	}{
		{
			"valid",
			*NewMsgEndRewardProgramRequest(
				1,
				"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
			),
			"",
		},
		{
			"invalid program id",
			*NewMsgEndRewardProgramRequest(
				0,
				"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
			),
			"invalid reward program id: 0",
		},
		{
			"valid program owner address",
			*NewMsgEndRewardProgramRequest(
				1,
				"invalid-address",
			),
			"invalid address for rewards program distribution from address: decoding bech32 failed: invalid separator index -1",
		},
	}
	for _, tt := range tests {
		tt := tt
		s.T().Run(tt.name, func(t *testing.T) {
			err := tt.msgCreateRewardProgramRequest.ValidateBasic()
			if err != nil {
				assert.Equal(t, tt.want, err.Error())
			} else if len(tt.want) > 0 {
				t.Errorf("MsgEndRewardProgramRequest ValidateBasic error = nil, expected: %s", tt.want)
			}
		})
	}
}

func (s *RewardMsgTypesTestSuite) TestMsgClaimRewardValidateBasic() {
	tests := []struct {
		name                  string
		MsgClaimRewardsRequest MsgClaimRewardsRequest
		want                  string
	}{
		{
			"valid",
			*NewMsgClaimRewardsRequest(1, "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h"),
			"",
		},
		{
			"invalid - address incorrect",
			*NewMsgClaimRewardsRequest(
				1,
				"invalid",
			),
			"invalid reward address : decoding bech32 failed: invalid bech32 string length 7",
		},
		{
			"invalid - incorrect reward id",
			*NewMsgClaimRewardsRequest(
				0,
				"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
			),
			"invalid rewards program id : 0",
		},
	}
	for _, tt := range tests {
		tt := tt
		s.T().Run(tt.name, func(t *testing.T) {
			err := tt.MsgClaimRewardsRequest.ValidateBasic()
			if err != nil {
				assert.Equal(t, tt.want, err.Error())
			} else if len(tt.want) > 0 {
				t.Errorf("MsgClaimRewardsRequest ValidateBasic error = nil, expected: %s", tt.want)
			}
		})
	}
}
