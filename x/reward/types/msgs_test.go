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
	now := time.Now().UTC()
	lTitle := make([]byte, MaxTitleLength+1)
	longTitle := string(lTitle)
	lDescription := make([]byte, MaxDescriptionLength+1)
	longDescription := string(lDescription)
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
				sdk.NewInt64Coin("jackthecat", 2),
				now,
				"day",
				1,
				NewEligibilityCriteria("name", &ActionDelegate{}),
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
				sdk.NewInt64Coin("jackthecat", 2),
				now,
				"day",
				1,
				NewEligibilityCriteria("name", &ActionDelegate{}),
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
				sdk.NewInt64Coin("jackthecat", 2),
				now,
				"day",
				1,
				NewEligibilityCriteria("name", &ActionDelegate{}),
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
				sdk.NewInt64Coin("jackthecat", 2),
				now,
				"day",
				1,
				NewEligibilityCriteria("name", &ActionDelegate{}),
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
				sdk.NewInt64Coin("jackthecat", 2),
				now,
				"day",
				1,
				NewEligibilityCriteria("name", &ActionDelegate{}),
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
				sdk.NewInt64Coin("jackthecat", 2),
				now,
				"day",
				1,
				NewEligibilityCriteria("name", &ActionDelegate{}),
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
				now,
				"day",
				1,
				NewEligibilityCriteria("name", &ActionDelegate{}),
			),
			"reward program requires coins: 0jackthecat",
		},
		{
			"invalid - max coin per address is invalid",
			*NewMsgCreateRewardProgramRequest(
				"title",
				"description",
				"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
				sdk.NewInt64Coin("jackthecat", 1),
				sdk.NewInt64Coin("jackthecat", 0),
				now,
				"day",
				1,
				NewEligibilityCriteria("name", &ActionDelegate{}),
			),
			"reward program requires positive max reward by address: 0jackthecat",
		},
		{
			"invalid - epoch type not found",
			*NewMsgCreateRewardProgramRequest(
				"title",
				"description",
				"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
				sdk.NewInt64Coin("jackthecat", 1),
				sdk.NewInt64Coin("jackthecat", 2),
				now,
				"blah",
				1,
				NewEligibilityCriteria("name", &ActionDelegate{}),
			),
			"epoch type not found: blah",
		},
		{
			"invalid - number of epochs is 0",
			*NewMsgCreateRewardProgramRequest(
				"title",
				"description",
				"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
				sdk.NewInt64Coin("jackthecat", 1),
				sdk.NewInt64Coin("jackthecat", 2),
				now,
				"day",
				0,
				NewEligibilityCriteria("name", &ActionDelegate{}),
			),
			"reward program number of epochs must be larger than 0",
		},
		{
			"invalid - ec validation failure",
			*NewMsgCreateRewardProgramRequest(
				"title",
				"description",
				"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
				sdk.NewInt64Coin("jackthecat", 1),
				sdk.NewInt64Coin("jackthecat", 2),
				now,
				"day",
				1,
				NewEligibilityCriteria("", &ActionDelegate{}),
			),
			"eligibility criteria is not valid: eligibility criteria must have a name",
		},
	}
	for _, tt := range tests {
		tt := tt
		s.T().Run(tt.name, func(t *testing.T) {
			err := tt.msgCreateRewardProgramRequest.ValidateBasic()
			if err != nil {
				assert.Equal(t, tt.want, err.Error())
			} else if len(tt.want) > 0 {
				t.Errorf("RewardProgram ValidateBasic error = nil, expected: %s", tt.want)
			}
		})
	}
}
