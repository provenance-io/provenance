package types

import (
	"testing"

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
	tests := []struct {
		name          string
		rewardProgram RewardProgram
		want          string
	}{
		{
			"invalid - wrong distribute from address format",
			NewRewardProgram(
				1,
				"invalid-address",
				sdk.NewInt64Coin("jackthecat", 1),
				sdk.NewInt64Coin("jackthecat", 2),
				"day",
				1,
				1,
				NewEligibilityCriteria("action-name", &ActionDelegate{}),
				false,
				1,
				2,
			),
			"invalid address for rewards program distribution from address: decoding bech32 failed: invalid separator index -1",
		},
		{
			"invalid - epoch id is empty",
			NewRewardProgram(
				1,
				"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
				sdk.NewInt64Coin("jackthecat", 1),
				sdk.NewInt64Coin("jackthecat", 2),
				"",
				1,
				1,
				NewEligibilityCriteria("action-name", &ActionDelegate{}),
				false,
				1,
				2,
			),
			"epoch id cannot be empty",
		},
		{
			"invalid - validate basic on ec fail",
			NewRewardProgram(
				1,
				"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
				sdk.NewInt64Coin("jackthecat", 1),
				sdk.NewInt64Coin("jackthecat", 2),
				"day",
				1,
				1,
				NewEligibilityCriteria("", &ActionDelegate{}),
				false,
				1,
				2,
			),
			"eligibility criteria is not valid: eligibility criteria must have a name",
		},
		{
			"invalid - coin amount must be positive",
			NewRewardProgram(
				1,
				"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
				sdk.NewInt64Coin("jackthecat", 0),
				sdk.NewInt64Coin("jackthecat", 2),
				"day",
				1,
				1,
				NewEligibilityCriteria("action-name", &ActionDelegate{}),
				false,
				1,
				2,
			),
			"reward program requires coins: 0jackthecat",
		},
		{
			"invalid - MaxRewardByAddress amount must be positive",
			NewRewardProgram(
				1,
				"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
				sdk.NewInt64Coin("jackthecat", 1),
				sdk.NewInt64Coin("jackthecat", 0),
				"day",
				1,
				1,
				NewEligibilityCriteria("action-name", &ActionDelegate{}),
				false,
				1,
				2,
			),
			"reward program requires positive max reward by address: 0jackthecat",
		},
		{
			"invalid - Maximum must be larger than 0",
			NewRewardProgram(
				1,
				"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
				sdk.NewInt64Coin("jackthecat", 1),
				sdk.NewInt64Coin("jackthecat", 2),
				"day",
				1,
				1,
				NewEligibilityCriteria("action-name", &ActionDelegate{}),
				false,
				1,
				0,
			),
			"maximum must be larger than 0",
		},
		{
			"invalid - maximum cannot be smaller than minimum",
			NewRewardProgram(
				1,
				"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
				sdk.NewInt64Coin("jackthecat", 1),
				sdk.NewInt64Coin("jackthecat", 2),
				"day",
				1,
				1,
				NewEligibilityCriteria("action-name", &ActionDelegate{}),
				false,
				100,
				2,
			),
			"minimum (100) cannot be larger than the maximum (2)",
		},
		{
			"valid",
			NewRewardProgram(
				1,
				"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
				sdk.NewInt64Coin("jackthecat", 1),
				sdk.NewInt64Coin("jackthecat", 2),
				"day",
				1,
				1,
				NewEligibilityCriteria("action-name", &ActionDelegate{}),
				false,
				1,
				2,
			),
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
				t.Errorf("RewardProgram ValidateBasic error = nil, expected: %s", tt.want)
			}
		})
	}
}

func (s *RewardTypesTestSuite) TestRewardClaimValidateBasic() {
	tests := []struct {
		name          string
		rewardProgram RewardClaim
		want          string
	}{
		{
			"invalid -  address format",
			NewRewardClaim("invalid", []SharesPerEpochPerRewardsProgram{}),
			"invalid address for reward claim address: decoding bech32 failed: invalid bech32 string length 7",
		},
		{
			"should succeed validate basic",
			NewRewardClaim("cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h", []SharesPerEpochPerRewardsProgram{}),
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
				t.Errorf("RewardProgram ValidateBasic error = nil, expected: %s", tt.want)
			}
		})
	}
}
