package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	metadatatypes "github.com/provenance-io/provenance/x/metadata/types"
)

// Compile time interface checks
var (
	_ sdk.Msg = &MsgAddExpirationRequest{}
	_ sdk.Msg = &MsgExtendExpirationRequest{}
	_ sdk.Msg = &MsgInvokeExpirationRequest{}
)

// private method to convert an array of strings into an array of Acc Addresses.
func stringsToAccAddresses(signers []string) []sdk.AccAddress {
	addresses := make([]sdk.AccAddress, len(signers))
	for i, address := range signers {
		addresses[i] = sdk.MustAccAddressFromBech32(address)
	}
	return addresses
}

// private method that does a simple validation check that
// doesn't require access to any other information.
func validateBasic(expiration Expiration, signers []string) error {
	if err := expiration.ValidateBasic(); err != nil {
		return err
	}
	if len(signers) == 0 {
		return ErrMissingSigners
	}
	return nil
}

// private method that checks an address string is bech32 or MetadataAddress type
func validateAddress(s string) error {
	if _, err := sdk.AccAddressFromBech32(s); err != nil {
		// check if we're dealing with a MetadataAddress
		if _, err2 := metadatatypes.MetadataAddressFromBech32(s); err2 != nil {
			return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress,
				fmt.Sprintf("invalid module asset id: %s", err.Error()))
		}
	}
	return nil
}

// ------------------  MsgAddExpirationRequest  ------------------

func NewMsgAddExpirationRequest(expiration Expiration, signers []string) *MsgAddExpirationRequest {
	return &MsgAddExpirationRequest{
		Expiration: expiration,
		Signers:    signers,
	}
}

// ValidateBasic does a simple validation check that
// doesn't require access to any other information.
func (m *MsgAddExpirationRequest) ValidateBasic() error {
	return validateBasic(m.Expiration, m.Signers)
}

// GetSigners returns the typed AccAddress of signers that must sign
func (m *MsgAddExpirationRequest) GetSigners() []sdk.AccAddress {
	return stringsToAccAddresses(m.Signers)
}

// MsgTypeURL returns the TypeURL of a `sdk.Msg`
func (m *MsgAddExpirationRequest) MsgTypeURL() string {
	return sdk.MsgTypeURL(m)
}

// ------------------  MsgExtendExpirationRequest  ------------------

func NewMsgExtendExpirationRequest(
	moduleAssetID string,
	duration string,
	signers []string,
) *MsgExtendExpirationRequest {
	return &MsgExtendExpirationRequest{
		ModuleAssetId: moduleAssetID,
		Duration:      duration,
		Signers:       signers,
	}
}

// ValidateBasic does a simple validation check that
// doesn't require access to any other information.
func (m *MsgExtendExpirationRequest) ValidateBasic() error {
	if len(m.ModuleAssetId) == 0 {
		return ErrEmptyModuleAssetID
	}
	if err := validateAddress(m.ModuleAssetId); err != nil {
		return err
	}
	_, err := ParseDuration(m.Duration)
	if err != nil {
		return err
	}
	if len(m.Signers) == 0 {
		return ErrMissingSigners
	}
	return nil
}

// GetSigners returns the typed AccAddress of signers that must sign
func (m *MsgExtendExpirationRequest) GetSigners() []sdk.AccAddress {
	return stringsToAccAddresses(m.Signers)
}

// MsgTypeURL returns the TypeURL of a `sdk.Msg`
func (m *MsgExtendExpirationRequest) MsgTypeURL() string {
	return sdk.MsgTypeURL(m)
}

// ------------------  MsgInvokeExpirationRequest  ------------------

func NewMsgInvokeExpirationRequest(moduleAssetID string, signers []string) *MsgInvokeExpirationRequest {
	return &MsgInvokeExpirationRequest{
		ModuleAssetId: moduleAssetID,
		Signers:       signers,
	}
}

// ValidateBasic does a simple validation check that
// doesn't require access to any other information.
func (m *MsgInvokeExpirationRequest) ValidateBasic() error {
	if len(m.ModuleAssetId) == 0 {
		return ErrEmptyModuleAssetID
	}
	if err := validateAddress(m.ModuleAssetId); err != nil {
		return err
	}
	if len(m.Signers) == 0 {
		return ErrMissingSigners
	}
	return nil
}

// GetSigners returns the typed AccAddress of signers that must sign
func (m *MsgInvokeExpirationRequest) GetSigners() []sdk.AccAddress {
	return stringsToAccAddresses(m.Signers)
}

// MsgTypeURL returns the TypeURL of a `sdk.Msg`
func (m *MsgInvokeExpirationRequest) MsgTypeURL() string {
	return sdk.MsgTypeURL(m)
}
