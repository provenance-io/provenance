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
	// OwnerKeyPrefix is a prefix added to keys for indexing expiration records by owner address.
	OwnerKeyPrefix = []byte{0x01}
)

func GetModuleAssetKeyPrefix(moduleAssetId string) ([]byte, error) {
	key := ModuleAssetKeyPrefix
	accAddress, err := sdk.AccAddressFromBech32(moduleAssetId)
	if err != nil {
		return nil, err
	}
	key = append(key, accAddress.Bytes()...)
	return key, nil
}

// GetOwnerKeyIndexPrefix returns an owner indexing store key for an expiration record
func GetOwnerKeyIndexPrefix(moduleAssetId string, owner string) ([]byte, error) {
	moduleKey, err := GetModuleAssetKeyPrefix(moduleAssetId)
	if err != nil {
		return nil, err
	}
	ownerKey, err := getOwnerKeyPrefix(owner)
	if err != nil {
		return nil, err
	}
	return append(moduleKey, ownerKey...), nil
}

// GetOwnerKeyPrefix returns a store key for a name record address
func GetOwnerKeyPrefix(owner string) ([]byte, error) {
	return getOwnerKeyPrefix(owner)
}

func getOwnerKeyPrefix(owner string) ([]byte, error) {
	ownerAddress, err := sdk.AccAddressFromBech32(owner)
	if err != nil {
		return nil, err
	}
	key := OwnerKeyPrefix
	key = append(key, ownerAddress.Bytes()...)
	return key, nil
}
