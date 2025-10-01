package types

import (
	"fmt"
	"net/url"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewOSLocatorRecord creates a oslocator for a given address.
func NewOSLocatorRecord(ownerAddr, encryptionKey sdk.AccAddress, uri string) ObjectStoreLocator {
	return ObjectStoreLocator{
		Owner:         ownerAddr.String(),
		LocatorUri:    uri,
		EncryptionKey: encryptionKey.String(),
	}
}

func (r ObjectStoreLocator) Validate() error {
	if strings.TrimSpace(r.Owner) == "" {
		return fmt.Errorf("owner address cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(r.Owner); err != nil {
		return fmt.Errorf("failed to add locator for a given owner address, invalid address: %s", r.Owner)
	}

	if strings.TrimSpace(r.LocatorUri) == "" {
		return fmt.Errorf("uri cannot be empty")
	}

	if _, err := url.Parse(r.LocatorUri); err != nil {
		return fmt.Errorf("failed to add locator for a given owner address, invalid uri: %s", r.LocatorUri)
	}

	if strings.TrimSpace(r.EncryptionKey) != "" {
		if _, err := sdk.AccAddressFromBech32(r.EncryptionKey); err != nil {
			return fmt.Errorf("failed to add locator for a given owner address: %s, invalid encryption key address: %s",
				r.Owner, r.EncryptionKey)
		}
	}
	return nil
}
