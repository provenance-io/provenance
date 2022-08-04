package types

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
)

type RewardGenesisTypesTestSuite struct {
	suite.Suite
}

func TestRewardGenesisTypesTestSuite(t *testing.T) {
	suite.Run(t, new(RewardGenesisTypesTestSuite))
}

func (s *RewardGenesisTypesTestSuite) SetupSuite() {
	s.T().Parallel()
}

func (s *RewardGenesisTypesTestSuite) TestGenesisValidate() {
	now := time.Now()
	claimPeriodSeconds := uint64(1000)
	claimPeriods := uint64(100)
	maxRolloverPeriods := uint64(10)
	expireClaimPeriods := uint64(100000)

	rewardProgram := NewRewardProgram(
		"title",
		"description",
		1,
		"",
		sdk.NewInt64Coin("nhash", 100),
		sdk.NewInt64Coin("nhash", 100),
		now,
		claimPeriodSeconds,
		claimPeriods,
		maxRolloverPeriods,
		expireClaimPeriods,
		[]QualifyingAction{
			{
				Type: &QualifyingAction_Vote{
					Vote: &ActionVote{
						MinimumActions:          uint64(1),
						MaximumActions:          uint64(10),
						MinimumDelegationAmount: sdk.NewInt64Coin("nhash", 100),
					},
				},
			},
		},
	)
	s.Assert().Error(NewGenesisState(10, []RewardProgram{rewardProgram}, []ClaimPeriodRewardDistribution{}, []RewardAccountState{}).Validate(), "should fail on validation of reward program")

	// make reward program valid
	rewardProgram.DistributeFromAddress = "cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h"
	s.Assert().Error(NewGenesisState(1, []RewardProgram{rewardProgram}, []ClaimPeriodRewardDistribution{}, []RewardAccountState{}).Validate(), "should fail on validation of reward program id")

	invalidClaimPeriod := NewClaimPeriodRewardDistribution(0, 1, sdk.NewInt64Coin("nhash", 100), sdk.NewInt64Coin("nhash", 100), 0, false)
	s.Assert().Error(NewGenesisState(10, []RewardProgram{rewardProgram}, []ClaimPeriodRewardDistribution{invalidClaimPeriod}, []RewardAccountState{}).Validate(), "should fail on validation on claim period")

	validClaimPeriod := NewClaimPeriodRewardDistribution(1, 1, sdk.NewInt64Coin("nhash", 1), sdk.NewInt64Coin("nhash", 1), 0, false)
	invalidRewardState := NewRewardAccountState(
		1,
		2,
		"test",
		0,
		map[string]uint64{},
	)
	s.Assert().Error(NewGenesisState(10, []RewardProgram{rewardProgram}, []ClaimPeriodRewardDistribution{validClaimPeriod}, []RewardAccountState{invalidRewardState}).Validate(), "should fail on validation on reward state")

	validRewardState := NewRewardAccountState(
		1,
		2,
		"cosmos1v57fx2l2rt6ehujuu99u2fw05779m5e2ux4z2h",
		0,
		map[string]uint64{},
	)
	s.Assert().NoError(NewGenesisState(10, []RewardProgram{rewardProgram}, []ClaimPeriodRewardDistribution{validClaimPeriod}, []RewardAccountState{validRewardState}).Validate(), "should pass all validations")

}
