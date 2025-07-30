package types

import (
	"cosmossdk.io/collections"

	"github.com/cosmos/cosmos-sdk/types/address"
)

const (
	ModuleName = "smartaccounts"

	StoreKey = ModuleName
)

var (
	ModuleAccountAddress = address.Module(ModuleName)
	// SmartAccountNumberStoreKeyPrefix prefix for smartaccount-by-id store
	SmartAccountNumberStoreKeyPrefix = collections.NewPrefix("SmartAccountNumber")
)
