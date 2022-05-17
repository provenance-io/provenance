package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScopeKey(t *testing.T) {

	rewardProgramKey := GetRewardProgramKey(1)
	assert.EqualValues(t, RewardProgramKeyPrefix, rewardProgramKey[0:1])

	rewardProgramBalanceKey := GetRewardProgramBalanceKey(1)
	assert.EqualValues(t, RewardProgramBalanceKeyPrefix, rewardProgramBalanceKey[0:1])

}
