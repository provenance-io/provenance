package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewNameRecord creates a name record binding that is restricted for child updates to the owner.
func NewNameRecord(name string, address sdk.AccAddress, restricted bool) NameRecord { //nolint:interfacer
	return NameRecord{
		Name:       name,
		Address:    address.String(),
		Restricted: restricted,
	}
}

// implement fmt.Stringer
func (nr NameRecord) String() string {
	if nr.Restricted {
		return strings.TrimSpace(fmt.Sprintf(`%s: %s [restricted]`, nr.Name, nr.Address))
	}
	return strings.TrimSpace(fmt.Sprintf(`%s: %s`, nr.Name, nr.Address))
}

// ValidateBasic performs basic stateless validity checks.
func (nr NameRecord) ValidateBasic() error {
	if strings.TrimSpace(nr.Address) == "" {
		return ErrInvalidAddress
	}
	if strings.TrimSpace(nr.Name) == "" {
		return ErrNameSegmentTooShort
	}
	return nil
}
