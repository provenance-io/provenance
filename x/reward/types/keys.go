package types

import (
	"encoding/binary"
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
	RewardProgramIDKey     = []byte{0x02}

	RewardClaimKeyPrefix = []byte{0x03}

	EpochRewardDistributionKeyPrefix = []byte{0x04}

	EligibilityCriteriaKeyPrefix = []byte{0x05}

	ActionKeyPrefix       = []byte{0x06}
	ShareKeyPrefix        = []byte{0x07}
	AccountStateKeyPrefix = []byte{0x08}

	ActionDelegateKey            = []byte("Delegate")
	ActionTransferDelegationsKey = []byte("TransferDelegations")
)

// GetRewardProgramKey converts a name into key format.
func GetRewardProgramKey(id int64) []byte {
	idByte := []byte(strconv.FormatInt(id, 10))
	return append(RewardProgramKeyPrefix, idByte...)
}

// GetShareKey converts a reward program id, epoch id, and address into a ShareKey
func GetShareKey(rewardID uint64, epochID uint64, addr []byte) []byte {
	key := ShareKeyPrefix
	rewardByte := []byte(strconv.FormatUint(rewardID, 10))
	epochByte := []byte(strconv.FormatUint(epochID, 10))
	key = append(key, rewardByte...)
	key = append(key, epochByte...)
	key = append(key, address.MustLengthPrefix(addr)...)
	return key
}

// GetAccountStateKey converts a reward program id, epoch id, and address into an AccountStateKey
func GetAccountStateKey(rewardID uint64, epochID uint64, addr []byte) []byte {
	key := AccountStateKeyPrefix
	rewardByte := []byte(strconv.FormatUint(rewardID, 10))
	epochByte := []byte(strconv.FormatUint(epochID, 10))
	key = append(key, rewardByte...)
	key = append(key, epochByte...)
	key = append(key, address.MustLengthPrefix(addr)...)
	return key
}

// GetAccountStateKeyPrefix converts a reward program id and epoch id into a prefix for iterating
func GetAccountStateKeyPrefix(rewardID uint64, epochID uint64) []byte {
	key := AccountStateKeyPrefix
	rewardByte := []byte(strconv.FormatUint(rewardID, 10))
	epochByte := []byte(strconv.FormatUint(epochID, 10))
	key = append(key, rewardByte...)
	key = append(key, epochByte...)
	return key
}

// GetRewardShareKeyPrefix converts a reward program id into a prefix for iterating
func GetRewardShareKeyPrefix(rewardID uint64) []byte {
	key := ShareKeyPrefix
	rewardByte := []byte(strconv.FormatUint(rewardID, 10))
	key = append(key, rewardByte...)
	return key
}

// GetRewardProgramIDBytes returns the byte representation of the rewardprogramID
func GetRewardProgramIDBytes(rewardprogramID uint64) (rewardprogramIDBz []byte) {
	rewardprogramIDBz = make([]byte, 8)
	binary.BigEndian.PutUint64(rewardprogramIDBz, rewardprogramID)
	return
}

// GetRewardProgramIDFromBytes returns rewardprogramID in uint64 format from a byte array
func GetRewardProgramIDFromBytes(bz []byte) (rewardprogramID uint64) {
	return binary.BigEndian.Uint64(bz)
}

// GetRewardEpochShareKeyPrefix converts a reward program id and epoch id into a prefix for iterating
func GetRewardEpochShareKeyPrefix(rewardID uint64, epochID uint64) []byte {
	key := ShareKeyPrefix
	rewardByte := []byte(strconv.FormatUint(rewardID, 10))
	epochByte := []byte(strconv.FormatUint(epochID, 10))
	key = append(key, rewardByte...)
	key = append(key, epochByte...)
	return key
}

// GetRewardClaimsKey converts an reward claim
func GetRewardClaimsKey(addr []byte) []byte {
	return append(RewardClaimKeyPrefix, address.MustLengthPrefix(addr)...)
}

func GetEpochRewardDistributionKey(epochID string, rewardID string) []byte {
	key := append(EpochRewardDistributionKeyPrefix, []byte(epochID)...)
	return append(key, []byte(rewardID)...)
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
