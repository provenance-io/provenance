package types

const (
	// ModuleName defines the module name
	ModuleName = "asset"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName
)

// KVStore prefix bytes
var (
	// AssetKeyPrefix is the prefix for storing individual Asset objects
	// Key: AssetKeyPrefix | asset_id (string) -> Value: Asset (protobuf marshaled)
	AssetKeyPrefix = []byte{0x01}

	// AssetCountKey is the key for storing the total count of assets
	// Key: AssetCountKey -> Value: uint64 (protobuf marshaled or binary encoded)
	AssetCountKey = []byte{0x02}

	// AssetByBorrowerPrefix is the key for storing assets by borrower
	AssetByBorrowerPrefix = []byte{0x11} // Key: AssetByBorrowerPrefix | borrower_addr | asset_id -> Value: []byte{} (presence marker)
)

// AssetKey returns the key for a specific asset ID in the store
func AssetKey(assetID string) []byte {
	return append(AssetKeyPrefix, []byte(assetID)...)
}
