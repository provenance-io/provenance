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

// GetShareKey converts a reward program id, claim period id, and address into a ShareKey
func GetShareKey(rewardID uint64, rewardClaimPeriodId uint64, addr []byte) []byte {
	key := ShareKeyPrefix
	rewardByte := []byte(strconv.FormatUint(rewardID, 10))
	claimPeriodByte := []byte(strconv.FormatUint(rewardClaimPeriodId, 10))
	key = append(key, rewardByte...)
	key = append(key, claimPeriodByte...)
	key = append(key, address.MustLengthPrefix(addr)...)
	return key
}

// GetRewardAccountStateKey converts a reward program id, claim period id, and address into an AccountStateKey
func GetRewardAccountStateKey(rewardID uint64, rewardClaimPeriodId uint64, addr []byte) []byte {
	key := AccountStateKeyPrefix
	rewardByte := []byte(strconv.FormatUint(rewardID, 10))
	claimPeriodByte := []byte(strconv.FormatUint(rewardClaimPeriodId, 10))
	key = append(key, rewardByte...)
	key = append(key, claimPeriodByte...)
	key = append(key, address.MustLengthPrefix(addr)...)
	return key
}

// GetRewardAccountStateKeyPrefix converts a reward program id and claim period into a prefix for iterating
func GetRewardAccountStateKeyPrefix(rewardID uint64, rewardClaimPeriodId uint64) []byte {
	key := AccountStateKeyPrefix
	rewardByte := []byte(strconv.FormatUint(rewardID, 10))
	claimPeriodByte := []byte(strconv.FormatUint(rewardClaimPeriodId, 10))
	key = append(key, rewardByte...)
	key = append(key, claimPeriodByte...)
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

// GetRewardClaimPeriodShareKeyPrefix converts a reward program id and claim period into a prefix for iterating
func GetRewardClaimPeriodShareKeyPrefix(rewardID uint64, rewardClaimPeriodId uint64) []byte {
	key := ShareKeyPrefix
	rewardByte := []byte(strconv.FormatUint(rewardID, 10))
	claimPeriodByte := []byte(strconv.FormatUint(rewardClaimPeriodId, 10))
	key = append(key, rewardByte...)
	key = append(key, claimPeriodByte...)
	return key
}

// GetRewardClaimsKey converts an reward claim
func GetRewardClaimsKey(addr []byte) []byte {
	return append(RewardClaimKeyPrefix, address.MustLengthPrefix(addr)...)
}

func GetClaimPeriodRewardDistributionKey(claimId uint64, rewardID uint64) []byte {
	key := ClaimPeriodRewardDistributionKeyPrefix
	key = append(key, []byte(strconv.FormatUint(claimId, 10))...)
	return append(key, []byte(strconv.FormatUint(rewardID, 10))...)
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
