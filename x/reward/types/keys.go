package types

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/types/address"
)

const (
	// ModuleName defines the module name
	ModuleName = "reward"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName
)

var (
	RewardProgramKeyPrefix = []byte{0x01}

	RewardClaimKeyPrefix = []byte{0x02}

	EpochRewardDistributionKeyPrefix = []byte{0x03}

	EligibilityCriteriaKeyPrefix = []byte{0x04}

	ActionKeyPrefix = []byte{0x05}
	ShareKeyPrefix  = []byte{0x06}

	ActionDelegateKey            = []byte("Delegate")
	ActionTransferDelegationsKey = []byte("TransferDelegations")
)

// GetRewardProgramKey converts a name into key format.
func GetRewardProgramKey(id int64) []byte {
	idByte := []byte(strconv.FormatInt(id, 10))
	return append(RewardProgramKeyPrefix, idByte...)
}

// GetShareKey converts a reward program id, epoch id, and address into a ShareKey
func GetShareKey(rewardId uint64, epochId uint64, addr []byte) []byte {
	key := ShareKeyPrefix
	rewardByte := []byte(strconv.FormatUint(rewardId, 10))
	epochByte := []byte(strconv.FormatUint(epochId, 10))
	key = append(key, rewardByte...)
	key = append(key, epochByte...)
	key = append(key, address.MustLengthPrefix(addr)...)
	return key
}

// GetRewardShareKeyPrefix converts a reward program id into a prefix for iterating
func GetRewardShareKeyPrefix(rewardId uint64) []byte {
	key := ShareKeyPrefix
	rewardByte := []byte(strconv.FormatUint(rewardId, 10))
	key = append(key, rewardByte...)
	return key
}

// GetRewardEpochShareKeyPrefix converts a reward program id and epoch id into a prefix for iterating
func GetRewardEpochShareKeyPrefix(rewardId uint64, epochId uint64) []byte {
	key := ShareKeyPrefix
	rewardByte := []byte(strconv.FormatUint(rewardId, 10))
	epochByte := []byte(strconv.FormatUint(epochId, 10))
	key = append(key, rewardByte...)
	key = append(key, epochByte...)
	return key
}

// GetRewardClaimsKey converts an reward claim
func GetRewardClaimsKey(addr []byte) []byte {
	return append(RewardClaimKeyPrefix, address.MustLengthPrefix(addr)...)
}

func GetEpochRewardDistributionKey(epochId string, rewardId string) []byte {
	key := append(EpochRewardDistributionKeyPrefix, []byte(epochId)...)
	return append(key, []byte(rewardId)...)
}

func GetEligibilityCriteriaKey(name string) []byte {
	return append(EligibilityCriteriaKeyPrefix, []byte(name)...)
}

func GetActionKey(actionType string) []byte {
	return append(ActionKeyPrefix, []byte(actionType)...)
}

func GetActionDelegateKey() []byte {
	return append(ActionKeyPrefix, ActionDelegateKey...)
}

func GetActionTransferDelegationsKey() []byte {
	return append(ActionKeyPrefix, ActionTransferDelegationsKey...)
}
