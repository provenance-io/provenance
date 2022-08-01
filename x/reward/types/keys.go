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

	ActionKeyPrefix                    = []byte{0x07}
	AccountStateAddressLookupKeyPrefix = []byte{0x08}
	AccountStateKeyPrefix              = []byte{0x09}

	ActionDelegateKey            = []byte("Delegate")
	ActionTransferDelegationsKey = []byte("TransferDelegations")
)

// GetRewardProgramKey converts a name into key format.
func GetRewardProgramKey(id uint64) []byte {
	idByte := []byte(strconv.FormatUint(id, 10))
	return append(RewardProgramKeyPrefix, idByte...)
}

// GetRewardAccountStateKey converts a reward program id, claim period id, and address into an AccountStateKey
func GetRewardAccountStateKey(rewardID uint64, rewardClaimPeriodID uint64, addr []byte) []byte {
	key := AccountStateKeyPrefix

	rewardBytes := make([]byte, 8)
	claimPeriodBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(rewardBytes, rewardID)
	binary.BigEndian.PutUint64(claimPeriodBytes, rewardClaimPeriodID)
	key = append(key, rewardBytes...)
	key = append(key, claimPeriodBytes...)
	key = append(key, address.MustLengthPrefix(addr)...)
	return key
}

// GetRewardAccountStateAddressLookupKey facilitates lookup of AccountState via address
func GetRewardAccountStateAddressLookupKey(addr []byte, rewardID uint64, rewardClaimPeriodID uint64) []byte {
	key := AccountStateAddressLookupKeyPrefix

	rewardBytes := make([]byte, 8)
	claimPeriodBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(rewardBytes, rewardID)
	binary.BigEndian.PutUint64(claimPeriodBytes, rewardClaimPeriodID)
	key = append(key, address.MustLengthPrefix(addr)...)
	key = append(key, rewardBytes...)
	key = append(key, claimPeriodBytes...)
	return key
}

// GetRewardAccountStateClaimPeriodKey converts a reward program id and claim period into a prefix for iterating
func GetRewardAccountStateClaimPeriodKey(rewardID uint64, rewardClaimPeriodID uint64) []byte {
	key := AccountStateKeyPrefix
	rewardBytes := make([]byte, 8)
	claimPeriodBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(rewardBytes, rewardID)
	binary.BigEndian.PutUint64(claimPeriodBytes, rewardClaimPeriodID)
	key = append(key, rewardBytes...)
	key = append(key, claimPeriodBytes...)
	return key
}

// GetRewardProgramRewardAccountStateKey returns the key to iterate over all RewardAccountStates for a RewardProgram
func GetRewardProgramRewardAccountStateKey(rewardID uint64) []byte {
	key := AccountStateKeyPrefix
	rewardBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(rewardBytes, rewardID)
	key = append(key, rewardBytes...)
	return key
}

// GetAllRewardAccountStateKey returns the key to iterate over all AccountStates
func GetAllRewardAccountStateKey() []byte {
	key := AccountStateKeyPrefix
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

// GetClaimPeriodRewardDistributionKey returns claim period reward distribution key
func GetClaimPeriodRewardDistributionKey(claimID uint64, rewardID uint64) []byte {
	claimBytes := make([]byte, 8)
	rewardBytes := make([]byte, 8)
	key := ClaimPeriodRewardDistributionKeyPrefix
	binary.BigEndian.PutUint64(claimBytes, claimID)
	binary.BigEndian.PutUint64(rewardBytes, rewardID)
	key = append(key, claimBytes...)
	return append(key, rewardBytes...)
}

// GetAllRewardAccountByAddressPartialKey returns the key to iterate over all AccountStateAddressLookup by address
func GetAllRewardAccountByAddressPartialKey(addr []byte) []byte {
	key := AccountStateAddressLookupKeyPrefix
	key = append(key, address.MustLengthPrefix(addr)...)
	return key
}

// GetAllRewardAccountByAddressAndRewardsIdPartialKey returns the key to iterate over all AccountStateAddressLookup by address and rewards id
func GetAllRewardAccountByAddressAndRewardsIdPartialKey(addr []byte, rewardID uint64) []byte {
	key := AccountStateAddressLookupKeyPrefix
	rewardBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(rewardBytes, rewardID)
	key = append(key, address.MustLengthPrefix(addr)...)
	key = append(key, rewardBytes...)
	return key
}
