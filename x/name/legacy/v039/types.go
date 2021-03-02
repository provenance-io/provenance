package v039

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NameRecord is an address with a flag that indicates if the name is restricted
type NameRecord struct {
	// The bound name
	Name string `json:"name" yaml:"name"`
	// The address the name resolved to.
	Address sdk.AccAddress `json:"address" yaml:"address"`
	// Whether owner signature is required to add sub-names.
	Restricted bool `json:"restricted,omitempty" yaml:"restricted"`
	// The value exchange address
	Pointer sdk.AccAddress `json:"pointer,omitempty" yaml:"pointer"`
}

// NewNameRecord creates a name record binding that is restricted for child updates to the owner.
func NewNameRecord(name string, address sdk.AccAddress, restricted bool) NameRecord {
	return NameRecord{
		Name:       name,
		Address:    address,
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

// NameRecords within the GenesisState
type NameRecords []NameRecord

// GenesisState is the initial set of name -> address bindings.
type GenesisState struct {
	Bindings NameRecords `json:"bindings"`
}
