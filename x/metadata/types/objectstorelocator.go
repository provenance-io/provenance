package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewOSLocatorRecord creates a oslocator for a given address.
func NewOSLocatorRecord(ownerAddr sdk.AccAddress, uri string) ObjectStoreLocator { //nolint:interfacer
	return ObjectStoreLocator{
		Owner:      ownerAddr.String(),
		LocatorUri: uri,
	}
}

