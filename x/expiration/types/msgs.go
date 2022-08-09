package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Compile time interface checks
var (
	_ sdk.Msg = &MsgAddExpirationRequest{}
	_ sdk.Msg = &MsgExtendExpirationRequest{}
	_ sdk.Msg = &MsgDeleteExpirationRequest{}
)

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
	return mustAccAddressFromBech32(m.Signers)
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
	return mustAccAddressFromBech32(m.Signers)
}

// ------------------  MsgDeleteExpirationRequest  ------------------

func NewMsgDeleteExpirationRequest(moduleAssetId string, signers []string) *MsgDeleteExpirationRequest {
	return &MsgDeleteExpirationRequest{
		ModuleAssetId: moduleAssetId,
		Signers:       signers,
	}
}

// ValidateBasic does a simple validation check that
// doesn't require access to any other information.

func (m *MsgDeleteExpirationRequest) ValidateBasic() error {
	if len(m.ModuleAssetId) == 0 {
		return ErrEmptyModuleAssetId
	}
	if len(m.Signers) == 0 {
		return ErrMissingSigners
	}
	return nil
}

// GetSigners returns the typed AccAddress of signers that must sign
func (m *MsgDeleteExpirationRequest) GetSigners() []sdk.AccAddress {
	return mustAccAddressFromBech32(m.Signers)
}

// MustAccAddressFromBech32 converts a Bech32 address to sdk.AccAddress
// Panics on error
func mustAccAddressFromBech32(s []string) []sdk.AccAddress {
	addresses := make([]sdk.AccAddress, 0)
	for _, address := range s {
		accAddress, err := sdk.AccAddressFromBech32(address)
		if err != nil {
			panic(err)
		}
		addresses = append(addresses, accAddress)
	}
	return addresses
}

func validateBasic(expiration Expiration, signers []string) error {
	if err := expiration.ValidateBasic(); err != nil {
		return err
	}
	if len(signers) == 0 {
		return ErrMissingSigners
	}
	return nil
}
