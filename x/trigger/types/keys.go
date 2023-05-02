package types

const (
	// ModuleName defines the module name
	ModuleName = "trigger"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName
)

var (
	TriggerKeyPrefix = []byte{0x01}
)

// GetRewardProgramKey converts a name into key format.
/*func GetRewardProgramKey(id uint64) []byte {
	rewardIDBytes := make([]byte, RewardIDKeyLength)
	binary.BigEndian.PutUint64(rewardIDBytes, id)
	return append(RewardProgramKeyPrefix, rewardIDBytes...)
}*/
