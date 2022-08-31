package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "expiration"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName
)

var (
	// ModuleAssetKeyPrefix encoded objects use this key prefix
	ModuleAssetKeyPrefix = []byte{0x00}
)

// GetModuleAssetKeyPrefix returns the key prefix used by encoded objects stored in the kv store.
func GetModuleAssetKeyPrefix(moduleAssetID string) ([]byte, error) {
	key := ModuleAssetKeyPrefix
	accAddress, err := sdk.AccAddressFromBech32(moduleAssetID)
	if err != nil {
		return nil, err
	}
	key = append(key, accAddress.Bytes()...)
	return key, nil
}
