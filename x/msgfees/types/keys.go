package types

const (
	// ModuleName defines the module name
	ModuleName = "msgfees"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_msgfees"
)

// GetMsgFeeKey Takes in MsgName
func GetMsgFeeKey(p string) []byte {
	return append(MsgFeeKeyPrefix, []byte(p)...)
}

var (
	MsgFeeKeyPrefix = []byte{0x00}
)
