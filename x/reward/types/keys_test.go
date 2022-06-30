package types

import (
	"strconv"
	"testing"

	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/stretchr/testify/assert"
)

func TestScopeKey(t *testing.T) {

	rewardProgramKey := GetRewardProgramKey(1)
	assert.EqualValues(t, RewardProgramKeyPrefix, rewardProgramKey[0:1])

	rewardProgramBalanceKey := GetRewardProgramBalanceKey(1)
	assert.EqualValues(t, RewardProgramBalanceKeyPrefix, rewardProgramBalanceKey[0:1])

	// Test account state key
	accountStateKey := GetRewardAccountStateKey(1, 2, []byte("test"))
	assert.EqualValues(t, AccountStateKeyPrefix, accountStateKey[0:1])
	assert.EqualValues(t, []byte(strconv.FormatUint(1, 10)), accountStateKey[1:2])
	assert.EqualValues(t, []byte(strconv.FormatUint(2, 10)), accountStateKey[2:3])
	assert.EqualValues(t, address.MustLengthPrefix([]byte("test")), accountStateKey[3:])

	// Test get account state key prefix
	accountStateKeyPrefix := GetRewardAccountStateKeyPrefix(1, 2)
	assert.EqualValues(t, AccountStateKeyPrefix, accountStateKeyPrefix[0:1])
	assert.EqualValues(t, []byte(strconv.FormatUint(1, 10)), accountStateKeyPrefix[1:2])
	assert.EqualValues(t, []byte(strconv.FormatUint(2, 10)), accountStateKeyPrefix[2:3])
}
