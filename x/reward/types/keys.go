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

	ActionKeyPrefix       = []byte{0x05}
	ShareKeyPrefix        = []byte{0x06}
	ShareCounterKeyPrefix = []byte("ShareCounterKey")

	ActionDelegateKey            = []byte("Delegate")
	ActionTransferDelegationsKey = []byte("TransferDelegations")
)

// GetRewardProgramKey converts a name into key format.
func GetRewardProgramKey(id int64) []byte {
	idByte := []byte(strconv.FormatInt(id, 10))
	return append(RewardProgramKeyPrefix, idByte...)
}

// GetShareKey converts a reward program id into a share key
func GetShareKey(rewardId uint64, shareId uint64) []byte {
	key := GetRewardProgramShareKeyPrefix(rewardId)
	shareIdByte := []byte(strconv.FormatUint(shareId, 10))
	return append(key, []byte(shareIdByte)...)
}

func GetRewardProgramShareKeyPrefix(rewardId uint64) []byte {
	rewardIdByte := []byte(strconv.FormatUint(rewardId, 10))
	return append(ShareKeyPrefix, rewardIdByte...)
}

// GetShareKey converts a reward program id into a share key
func GetShareCounterKey(rewardId uint64) []byte {
	rewardIdByte := []byte(strconv.FormatUint(rewardId, 10))
	return append(ShareCounterKeyPrefix, []byte(rewardIdByte)...)
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
