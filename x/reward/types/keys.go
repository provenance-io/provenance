package types

import (
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
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

	RewardIDKeyLength   = 8
	ClaimPeriodIDLength = 8
)

var (
	RewardProgramKeyPrefix                 = []byte{0x01}
	RewardProgramIDKey                     = []byte{0x02}
	ClaimPeriodRewardDistributionKeyPrefix = []byte{0x03}
	AccountStateAddressLookupKeyPrefix     = []byte{0x04}
	AccountStateKeyPrefix                  = []byte{0x05}
)

// GetRewardProgramKey converts a name into key format.
func GetRewardProgramKey(id uint64) []byte {
	rewardIDBytes := make([]byte, RewardIDKeyLength)
	binary.BigEndian.PutUint64(rewardIDBytes, id)
	return append(RewardProgramKeyPrefix, rewardIDBytes...)
}

// GetRewardAccountStateKey converts a reward program id, claim period id, and address into an AccountStateKey
func GetRewardAccountStateKey(rewardID uint64, rewardClaimPeriodID uint64, addr sdk.AccAddress) []byte {
	key := AccountStateKeyPrefix
	rewardBytes := make([]byte, RewardIDKeyLength)
	claimPeriodBytes := make([]byte, ClaimPeriodIDLength)
	binary.BigEndian.PutUint64(rewardBytes, rewardID)
	binary.BigEndian.PutUint64(claimPeriodBytes, rewardClaimPeriodID)
	key = append(key, rewardBytes...)
	key = append(key, claimPeriodBytes...)
	key = append(key, address.MustLengthPrefix(addr)...)
	return key
}

// GetRewardAccountStateAddressLookupKey facilitates lookup of AccountState via address
func GetRewardAccountStateAddressLookupKey(addr sdk.AccAddress, rewardID uint64, rewardClaimPeriodID uint64) []byte {
	key := AccountStateAddressLookupKeyPrefix
	rewardBytes := make([]byte, RewardIDKeyLength)
	claimPeriodBytes := make([]byte, ClaimPeriodIDLength)
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
	rewardBytes := make([]byte, RewardIDKeyLength)
	claimPeriodBytes := make([]byte, ClaimPeriodIDLength)
	binary.BigEndian.PutUint64(rewardBytes, rewardID)
	binary.BigEndian.PutUint64(claimPeriodBytes, rewardClaimPeriodID)
	key = append(key, rewardBytes...)
	key = append(key, claimPeriodBytes...)
	return key
}

// GetRewardProgramRewardAccountStateKey returns the key to iterate over all RewardAccountStates for a RewardProgram
func GetRewardProgramRewardAccountStateKey(rewardID uint64) []byte {
	key := AccountStateKeyPrefix
	rewardBytes := make([]byte, RewardIDKeyLength)
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
	rewardprogramIDBz = make([]byte, RewardIDKeyLength)
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
func GetAllRewardAccountByAddressPartialKey(addr sdk.AccAddress) []byte {
	key := AccountStateAddressLookupKeyPrefix
	key = append(key, address.MustLengthPrefix(addr)...)
	return key
}

// GetAllRewardAccountByAddressAndRewardsIDPartialKey returns the key to iterate over all AccountStateAddressLookup by address and rewards id
func GetAllRewardAccountByAddressAndRewardsIDPartialKey(addr sdk.AccAddress, rewardID uint64) []byte {
	key := AccountStateAddressLookupKeyPrefix
	rewardBytes := make([]byte, RewardIDKeyLength)
	binary.BigEndian.PutUint64(rewardBytes, rewardID)
	key = append(key, address.MustLengthPrefix(addr.Bytes())...)
	key = append(key, rewardBytes...)
	return key
}

// MustAccAddressFromBech32 converts a Bech32 address to sdk.AccAddress
// Panics on error
func MustAccAddressFromBech32(s string) sdk.AccAddress {
	accAddress, err := sdk.AccAddressFromBech32(s)
	if err != nil {
		panic(err)
	}
	return accAddress
}
