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
	RewardProgramKeyPrefix        = []byte{0x01}
	RewardProgramIDKey            = []byte{0x02}
	RewardProgramBalanceKeyPrefix = []byte{0x03}

	RewardClaimKeyPrefix = []byte{0x04}

	ClaimPeriodRewardDistributionKeyPrefix = []byte{0x05}

	EligibilityCriteriaKeyPrefix = []byte{0x06}

	ActionKeyPrefix       = []byte{0x07}
	ShareKeyPrefix        = []byte{0x08}
	AccountStateKeyPrefix = []byte{0x09}

	ActionDelegateKey            = []byte("Delegate")
	ActionTransferDelegationsKey = []byte("TransferDelegations")
)

// GetRewardProgramKey converts a name into key format.
func GetRewardProgramKey(id uint64) []byte {
	idByte := []byte(strconv.FormatUint(id, 10))
	return append(RewardProgramKeyPrefix, idByte...)
}

func GetRewardProgramBalanceKey(rewardProgramID uint64) []byte {
	idByte := []byte(strconv.FormatUint(rewardProgramID, 10))
	return append(RewardProgramBalanceKeyPrefix, idByte...)
}

// GetShareKey converts a reward program id, subPeriod id, and address into a ShareKey
func GetShareKey(rewardID uint64, subPeriod uint64, addr []byte) []byte {
	key := ShareKeyPrefix
	rewardByte := []byte(strconv.FormatUint(rewardID, 10))
	subPeriodByte := []byte(strconv.FormatUint(subPeriod, 10))
	key = append(key, rewardByte...)
	key = append(key, subPeriodByte...)
	key = append(key, address.MustLengthPrefix(addr)...)
	return key
}

// GetAccountStateKey converts a reward program id, subPeriod id, and address into an AccountStateKey
func GetAccountStateKey(rewardID uint64, subPeriod uint64, addr []byte) []byte {
	key := AccountStateKeyPrefix
	rewardByte := []byte(strconv.FormatUint(rewardID, 10))
	subPeriodByte := []byte(strconv.FormatUint(subPeriod, 10))
	key = append(key, rewardByte...)
	key = append(key, subPeriodByte...)
	key = append(key, address.MustLengthPrefix(addr)...)
	return key
}

// GetAccountStateKeyPrefix converts a reward program id and sub period into a prefix for iterating
func GetAccountStateKeyPrefix(rewardID uint64, subPeriod uint64) []byte {
	key := AccountStateKeyPrefix
	rewardByte := []byte(strconv.FormatUint(rewardID, 10))
	subPeriodByte := []byte(strconv.FormatUint(subPeriod, 10))
	key = append(key, rewardByte...)
	key = append(key, subPeriodByte...)
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

// GetRewardSubPeriodShareKeyPrefix converts a reward program id and sub period into a prefix for iterating
func GetRewardSubPeriodShareKeyPrefix(rewardID uint64, subPeriod uint64) []byte {
	key := ShareKeyPrefix
	rewardByte := []byte(strconv.FormatUint(rewardID, 10))
	subPeriodByte := []byte(strconv.FormatUint(subPeriod, 10))
	key = append(key, rewardByte...)
	key = append(key, subPeriodByte...)
	return key
}

// GetRewardClaimsKey converts an reward claim
func GetRewardClaimsKey(addr []byte) []byte {
	return append(RewardClaimKeyPrefix, address.MustLengthPrefix(addr)...)
}

func GetClaimPeriodRewardDistributionKey(claimId string, rewardID string) []byte {
	key := ClaimPeriodRewardDistributionKeyPrefix
	key = append(key, []byte(claimId)...)
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
