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
		name        string
		rewardClaim RewardClaim
		want        string
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
			err := tt.rewardClaim.ValidateBasic()
			if err != nil {
				assert.Equal(t, tt.want, err.Error())
			} else if len(tt.want) > 0 {
				t.Errorf("RewardClaim ValidateBasic error = nil, expected: %s", tt.want)
			}
		})
	}
}

func (s *RewardTypesTestSuite) TestEpochRewardDistributionValidateBasic() {
	tests := []struct {
		name          string
		rewardProgram EpochRewardDistribution
		want          string
	}{
		{
			"invalid -  address format",
			NewEpochRewardDistribution("", 1, sdk.NewInt64Coin("jackthecat", 100), 0, false),
			"epoch reward distribution must have a epoch id",
		},
		{
			"invalid - reward program id",
			NewEpochRewardDistribution("day", 0, sdk.NewInt64Coin("jackthecat", 100), 0, false),
			"epoch reward distribution must have a valid reward program id",
		},
		{
			"invalid - total rewards needs to be positive",
			NewEpochRewardDistribution("day", 1, sdk.NewInt64Coin("jackthecat", 0), 0, false),
			"epoch reward distribution must have a reward pool",
		},
		{
			"should succeed validate basic",
			NewEpochRewardDistribution("day", 1, sdk.NewInt64Coin("jackthecat", 1), 0, false),
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

func (s *RewardTypesTestSuite) TestEligibilityCriteriaValidateBasic() {
	tests := []struct {
		name string
		ec   EligibilityCriteria
		want string
	}{
		{
			"invalid -  empty name",
			NewEligibilityCriteria("", &ActionDelegate{}),
			"eligibility criteria must have a name",
		},
		{
			"should succeed validate basic",
			NewEligibilityCriteria("name", &ActionDelegate{}),
			"",
		},
	}

	for _, tt := range tests {
		tt := tt
		s.T().Run(tt.name, func(t *testing.T) {
			err := tt.ec.ValidateBasic()
			if err != nil {
				assert.Equal(t, tt.want, err.Error())
			} else if len(tt.want) > 0 {
				t.Errorf("EligibilityCriteria ValidateBasic error = nil, expected: %s", tt.want)
			}
		})
	}
}

func (s *RewardTypesTestSuite) TestSharesPerEpochPerRewardsProgramValidateBasic() {
	tests := []struct {
		name                            string
		sharesPerEpochPerRewardsProgram SharesPerEpochPerRewardsProgram
		want                            string
	}{
		{
			"invalid -  rewards program id incorrect",
			NewSharesPerEpochPerRewardsProgram(
				0,
				0,
				0,
				1,
				false,
				false,
				sdk.NewInt64Coin("jackthecat", 1),
			),
			"shares per epoch must have a valid reward program id",
		},
		{
			"invalid -  last recorded epoch incorrect",
			NewSharesPerEpochPerRewardsProgram(
				1,
				0,
				0,
				0,
				false,
				false,
				sdk.NewInt64Coin("jackthecat", 1),
			),
			"latest recorded epoch cannot be less than 1",
		},
		{
			"should succeed validate basic",
			NewSharesPerEpochPerRewardsProgram(
				1,
				0,
				0,
				1,
				false,
				false,
				sdk.NewInt64Coin("jackthecat", 1),
			),
			"",
		},
	}

	for _, tt := range tests {
		tt := tt
		s.T().Run(tt.name, func(t *testing.T) {
			err := tt.sharesPerEpochPerRewardsProgram.ValidateBasic()
			if err != nil {
				assert.Equal(t, tt.want, err.Error())
			} else if len(tt.want) > 0 {
				t.Errorf("SharesPerEpochPerRewardsProgram ValidateBasic error = nil, expected: %s", tt.want)
			}
		})
	}
}
