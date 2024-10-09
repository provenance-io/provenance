package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/x/metadata/types"
)

// SignersWrapper stores the signers as strings and acc addresses.
// One is created by providing the strings. They are then converted to acc addresses
// if they're needed that way.
type SignersWrapper struct {
	signers    []string
	signerAccs []sdk.AccAddress
	converted  bool
}

func NewSignersWrapper(signers []string) *SignersWrapper {
	return &SignersWrapper{signers: signers}
}

// Strings gets the string versions of the signers.
func (s *SignersWrapper) Strings() []string {
	return s.signers
}

// Accs gets the sdk.AccAddress versions of the signers.
// Conversion happens if it hasn't already been done yet.
// Any strings that fail to convert are simply ignored.
func (s *SignersWrapper) Accs() []sdk.AccAddress {
	if !s.converted {
		s.signerAccs = safeBech32ToAccAddresses(s.signers)
		s.converted = true
	}
	return s.signerAccs
}

// UnwrapMetadataContext retrieves a Context from a context.Context instance attached with WrapSDKContext.
// It then adds an types.AuthzCache to it.
// It panics if a Context was not properly attached, or if the types.AuthzCache can't be added.
//
// This should be used for all Metadata msg server endpoints instead of sdk.UnwrapSDKContext.
// This should not be used outside of the Metadata module.
func UnwrapMetadataContext(goCtx context.Context) sdk.Context {
	return types.AddAuthzCacheToContext(sdk.UnwrapSDKContext(goCtx))
}

// safeBech32ToAccAddresses attempts to convert all provided strings to AccAddresses.
// Any that fail to convert are ignored.
func safeBech32ToAccAddresses(bech32s []string) []sdk.AccAddress {
	rv := make([]sdk.AccAddress, 0, len(bech32s))
	for _, bech32 := range bech32s {
		addr, err := sdk.AccAddressFromBech32(bech32)
		if err == nil {
			rv = append(rv, addr)
		}
	}
	return rv
}
