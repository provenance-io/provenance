package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Compile time interface checks
var (
	_ sdk.Msg = &MsgAddExpirationRequest{}
	_ sdk.Msg = &MsgExtendExpirationRequest{}
	//_ sdk.Msg = &MsgDeleteExpirationRequest{}
	_ sdk.Msg = &MsgInvokeExpirationRequest{}
)

// MustAccAddressFromBech32 converts a Bech32 address to sdk.AccAddress
// Panics on error
func MustAccAddressFromBech32(s string) sdk.AccAddress {
	accAddress, err := sdk.AccAddressFromBech32(s)
	if err != nil {
		panic(err)
	}
	return accAddress
}

// private method to convert an array of strings into an array of Acc Addresses.
func stringsToAccAddresses(signers []string) []sdk.AccAddress {
	addresses := make([]sdk.AccAddress, len(signers))
	for i, address := range signers {
		addresses[i] = MustAccAddressFromBech32(address)
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

func NewMsgExtendExpirationRequest(expiration Expiration, signers []string) *MsgExtendExpirationRequest {
	return &MsgExtendExpirationRequest{
		Expiration: expiration,
		Signers:    signers,
	}
}

// ValidateBasic does a simple validation check that
// doesn't require access to any other information.
func (m *MsgExtendExpirationRequest) ValidateBasic() error {
	return validateBasic(m.Expiration, m.Signers)
}

// GetSigners returns the typed AccAddress of signers that must sign
func (m *MsgExtendExpirationRequest) GetSigners() []sdk.AccAddress {
	return stringsToAccAddresses(m.Signers)
}

// MsgTypeURL returns the TypeURL of a `sdk.Msg`
func (m *MsgExtendExpirationRequest) MsgTypeURL() string {
	return sdk.MsgTypeURL(m)
}

// ------------------  MsgDeleteExpirationRequest  ------------------

//func NewMsgDeleteExpirationRequest(moduleAssetID string, signers []string) *MsgDeleteExpirationRequest {
//	return &MsgDeleteExpirationRequest{
//		ModuleAssetId: moduleAssetID,
//		Signers:       signers,
//	}
//}
//
//// ValidateBasic does a simple validation check that
//// doesn't require access to any other information.
//func (m *MsgDeleteExpirationRequest) ValidateBasic() error {
//	if len(m.ModuleAssetId) == 0 {
//		return ErrEmptyModuleAssetID
//	}
//	if len(m.Signers) == 0 {
//		return ErrMissingSigners
//	}
//	return nil
//}
//
//// GetSigners returns the typed AccAddress of signers that must sign
//func (m *MsgDeleteExpirationRequest) GetSigners() []sdk.AccAddress {
//	return stringsToAccAddresses(m.Signers)
//}
//
//// MsgTypeURL returns the TypeURL of a `sdk.Msg`
//func (m *MsgDeleteExpirationRequest) MsgTypeURL() string {
//	return sdk.MsgTypeURL(m)
//}

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
