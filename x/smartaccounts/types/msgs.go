package types

import (
	"fmt"

	"cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.Msg = &MsgUpdateParams{}
	_ sdk.Msg = &MsgRegisterFido2Credential{}
)
var AllRequestMsgs = []sdk.Msg{
	(*MsgUpdateParams)(nil),
	(*MsgRegisterFido2Credential)(nil),
	(*MsgRegisterCosmosCredential)(nil),
	(*MsgDeleteCredential)(nil),
}

// NewMsgUpdateParams creates new instance of MsgUpdateParams
func NewMsgUpdateParams(
	sender sdk.Address,
	someValue bool,
) *MsgUpdateParams {
	return &MsgUpdateParams{
		Authority: sender.String(),
		Params: Params{
			Enabled: someValue,
		},
	}
}

// ValidateBasic does a sanity check on the provided data.
func (msg *MsgUpdateParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return errors.Wrap(err, "invalid authority address")
	}

	return msg.Params.Validate()
}

func NewMsgRegisterFido2Credential(sender, encodedAttestation, userIdentifier string) *MsgRegisterFido2Credential {
	return &MsgRegisterFido2Credential{
		Sender:             sender,
		EncodedAttestation: encodedAttestation,
		UserIdentifier:     userIdentifier,
	}
}

func (msg *MsgRegisterFido2Credential) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return errors.Wrap(err, "invalid sender address")
	}
	if msg.EncodedAttestation == "" {
		return errors.Wrap(sdkerrors.ErrInvalidRequest, "encoded attestation cannot be empty")
	}
	if msg.UserIdentifier == "" {
		return errors.Wrap(sdkerrors.ErrInvalidRequest, "user identifier cannot be empty")
	}
	return nil
}

// NewMsgDeleteCredential creates new instance of MsgDeleteCredential
func NewMsgDeleteCredential(sender string, credentialNumber uint64) *MsgDeleteCredential {
	return &MsgDeleteCredential{
		Sender:           sender,
		CredentialNumber: credentialNumber,
	}
}

// ValidateBasic does a sanity check on the provided data.
func (msg *MsgDeleteCredential) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return errors.Wrap(err, "invalid sender address")
	}
	return nil
}

// NewMsgRegisterCosmosCredential creates new instance of MsgRegisterCosmosCredential
func NewMsgRegisterCosmosCredential(sender string, pubkey *types.Any) *MsgRegisterCosmosCredential {
	return &MsgRegisterCosmosCredential{
		Sender: sender,
		Pubkey: pubkey,
	}
}

// ValidateBasic does a sanity check on the provided data.
func (msg *MsgRegisterCosmosCredential) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return errors.Wrap(err, "invalid sender address")
	}
	if msg.Pubkey == nil {
		return fmt.Errorf("pubkey cannot be nil")
	}
	return nil
}
