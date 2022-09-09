package types

import (
	"encoding/binary"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/stretchr/testify/assert"
)

func TestRewardModuleTypeKeys(t *testing.T) {
	addressFromSec256k1 := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	rewardProgramId := uint64(123456)
	claimPeriodId := uint64(7891011)

	rewardProgramKey := GetRewardProgramKey(rewardProgramId)
	assert.EqualValues(t, RewardProgramKeyPrefix, rewardProgramKey[0:1])
	assert.EqualValues(t, rewardProgramId, uint64(binary.BigEndian.Uint64(rewardProgramKey[1:9])))

	accountStateKey := GetRewardAccountStateKey(rewardProgramId, claimPeriodId, addressFromSec256k1)
	assert.EqualValues(t, AccountStateKeyPrefix, accountStateKey[0:1])
	assert.EqualValues(t, rewardProgramId, uint64(binary.BigEndian.Uint64(accountStateKey[1:9])))
	assert.EqualValues(t, claimPeriodId, uint64(binary.BigEndian.Uint64(accountStateKey[9:17])))
	assert.EqualValues(t, address.MustLengthPrefix(addressFromSec256k1), accountStateKey[17:])

	rewardAccountStateKey := GetRewardAccountStateClaimPeriodKey(rewardProgramId, claimPeriodId)
	assert.EqualValues(t, AccountStateKeyPrefix, rewardAccountStateKey[0:1])
	assert.EqualValues(t, rewardProgramId, binary.BigEndian.Uint64(rewardAccountStateKey[1:9]))
	assert.EqualValues(t, claimPeriodId, binary.BigEndian.Uint64(rewardAccountStateKey[9:17]))

	rewardProgramRewardAccountStateKey := GetRewardProgramRewardAccountStateKey(rewardProgramId)
	assert.EqualValues(t, AccountStateKeyPrefix, rewardProgramRewardAccountStateKey[0:1])
	assert.EqualValues(t, rewardProgramId, binary.BigEndian.Uint64(rewardProgramRewardAccountStateKey[1:9]))

	allRewardAccountStateKey := GetAllRewardAccountStateKey()
	assert.EqualValues(t, AccountStateKeyPrefix, allRewardAccountStateKey[0:1])

	rewardProgramIdBytes := GetRewardProgramIDBytes(rewardProgramId)
	assert.EqualValues(t, []uint8([]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0xe2, 0x40}), rewardProgramIdBytes)

	rewardIdResult := GetRewardProgramIDFromBytes([]uint8([]byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0xe2, 0x40}))
	assert.EqualValues(t, rewardProgramId, rewardIdResult)

	claimPeriodRewardDistributionKey := GetClaimPeriodRewardDistributionKey(claimPeriodId, rewardProgramId)
	assert.EqualValues(t, ClaimPeriodRewardDistributionKeyPrefix, claimPeriodRewardDistributionKey[0:1])
	assert.EqualValues(t, rewardProgramId, binary.BigEndian.Uint64(claimPeriodRewardDistributionKey[1:9]))
	assert.EqualValues(t, claimPeriodId, binary.BigEndian.Uint64(claimPeriodRewardDistributionKey[9:17]))

	accountStateAddressLookupKey := GetRewardAccountStateAddressLookupKey(addressFromSec256k1, rewardProgramId, claimPeriodId)
	assert.EqualValues(t, AccountStateAddressLookupKeyPrefix, accountStateAddressLookupKey[0:1])

	assert.Equal(t, byte(20), accountStateAddressLookupKey[1:2][0], "should be the length of key 20 for secp256k1")
	assert.EqualValues(t, addressFromSec256k1.Bytes(), accountStateAddressLookupKey[2:22])
	assert.EqualValues(t, rewardProgramId, binary.BigEndian.Uint64(accountStateAddressLookupKey[22:30]))
	assert.EqualValues(t, claimPeriodId, uint64(binary.BigEndian.Uint64(accountStateAddressLookupKey[30:38])))

	rewardAccountByAddressPartial := GetAllRewardAccountByAddressPartialKey(addressFromSec256k1)
	assert.Equal(t, byte(20), rewardAccountByAddressPartial[1:2][0], "should be the length of key 20 for secp256k1")
	assert.EqualValues(t, AccountStateAddressLookupKeyPrefix, rewardAccountByAddressPartial[0:1])
	assert.EqualValues(t, addressFromSec256k1.Bytes(), rewardAccountByAddressPartial[2:22])

	rewardAccountByAddressAndRewardsIDPartialKey := GetAllRewardAccountByAddressAndRewardsIDPartialKey(addressFromSec256k1, rewardProgramId)
	assert.Equal(t, byte(20), rewardAccountByAddressAndRewardsIDPartialKey[1:2][0], "should be the length of key 20 for secp256k1")
	assert.EqualValues(t, AccountStateAddressLookupKeyPrefix, rewardAccountByAddressAndRewardsIDPartialKey[0:1])
	assert.EqualValues(t, addressFromSec256k1.Bytes(), rewardAccountByAddressAndRewardsIDPartialKey[2:22])
	assert.EqualValues(t, rewardProgramId, binary.BigEndian.Uint64(rewardAccountByAddressAndRewardsIDPartialKey[22:30]))

}
