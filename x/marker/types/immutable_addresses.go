package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// ImmutableAccAddresses holds some AccAddresses in a way that they cannot be changed after creation.
type ImmutableAccAddresses struct {
	// addrsSlice is a deep copy of the originally provided addresses slice.
	addrsSlice []sdk.AccAddress
	// addrsMap contains the same addresses as addrsSlice for easier checking.
	// Each key is an sdk.AccAddress cast to a string (not converted to bech32).
	addrsMap map[string]bool
}

// NewImmutableAccAddresses creates a new ImmutableAccAddresses containing the provided addresses.
func NewImmutableAccAddresses(addrs []sdk.AccAddress) ImmutableAccAddresses {
	rv := ImmutableAccAddresses{
		addrsSlice: deepCopyAccAddresses(addrs),
		addrsMap:   make(map[string]bool),
	}
	for _, addr := range addrs {
		rv.addrsMap[string(addr)] = true
	}
	return rv
}

// GetSlice returns a copy of the addresses known to this ImmutableAccAddresses.
func (i ImmutableAccAddresses) GetSlice() []sdk.AccAddress {
	return deepCopyAccAddresses(i.addrsSlice)
}

// Has returns true if the provided address is known to this ImmutableAccAddresses.
func (i ImmutableAccAddresses) Has(addr sdk.AccAddress) bool {
	return i.addrsMap[string(addr)]
}

// deepCopyAccAddresses creates a deep copy of the provided slice of acc addresses.
// A copy of each entry is made and placed into a new slice.
func deepCopyAccAddresses(orig []sdk.AccAddress) []sdk.AccAddress {
	if orig == nil {
		return nil
	}
	rv := make([]sdk.AccAddress, len(orig))
	for i, addr := range orig {
		rv[i] = make(sdk.AccAddress, len(addr))
		copy(rv[i], addr)
	}
	return rv
}
