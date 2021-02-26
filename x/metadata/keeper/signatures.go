package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/crypto/types/multisig"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"

	markertypes "github.com/provenance-io/provenance/x/marker/types"
)

// AccountIsMarker determines if account is marker
func (k Keeper) AccountIsMarker(ctx sdk.Context, address string) bool {
	addr, err := sdk.AccAddressFromBech32(address)
	// if the value owner is invalid then it is not possible to have any authority for it. e.g. value owner is empty.
	if err != nil {
		return false
	}

	mac := k.authKeeper.GetAccount(ctx, addr)
	if mac == nil {
		return false
	}

	// Convert over to the actual underlying marker type, or not.
	_, isMarker := mac.(*markertypes.MarkerAccount)

	return isMarker
}

// HasSignerWithMarkerValueAuthority checks the list of signers for any that have the requested role.
func (k Keeper) HasSignerWithMarkerValueAuthority(ctx sdk.Context, valueOwner string, signers []string, role markertypes.Access) bool {
	valueOwnerAddr, err := sdk.AccAddressFromBech32(valueOwner)
	// if the value owner is invalid then it is not possible to have any authority for it. e.g. value owner is empty.
	if err != nil {
		return false
	}

	mac := k.authKeeper.GetAccount(ctx, valueOwnerAddr)
	if mac == nil {
		return false
	}

	// Convert over to the actual underlying marker type, or not.
	macc, isMarker := mac.(*markertypes.MarkerAccount)
	if isMarker {
		for _, signer := range signers {
			address, err := sdk.AccAddressFromBech32(signer)
			if err != nil {
				continue // invalid address, loop to next.
			}
			// since this is a marker, check for the role and return true if found.
			if macc.AddressHasAccess(address, role) {
				return true
			}
		}
	}
	return false
}

// ValidateRawSignature takes a given message and verifies the signature instance is valid
// for it directly without calculating a signing structure to wrap it. ValidateRawSignature returns the address of the
// user who created the signature and any encountered errors.
func (k Keeper) ValidateRawSignature(signature signing.SignatureDescriptor, message []byte) (addr sdk.AccAddress, err error) {
	var pubKey cryptotypes.PubKey
	if err = k.cdc.UnpackAny(signature.PublicKey, &pubKey); err != nil {
		return
	}
	addr = sdk.AccAddress(pubKey.Address().Bytes())

	sigData := signing.SignatureDataFromProto(signature.Data)
	switch data := sigData.(type) {
	case *signing.SingleSignatureData:
		if !pubKey.VerifySignature(message, data.Signature) {
			err = fmt.Errorf("unable to verify single signer signature")
		}
	case *signing.MultiSignatureData:
		multiPK, ok := pubKey.(multisig.PubKey)
		if !ok {
			err = fmt.Errorf("expected %T, got %T", (multisig.PubKey)(nil), pubKey)
			return
		}
		err = multiPK.VerifyMultisignature(func(mode signing.SignMode) ([]byte, error) {
			// no special adjustments need to be made to the signing bytes based on signing mode
			return message, nil
		}, data)
	default:
		err = fmt.Errorf("unexpected SignatureData %T", sigData)
	}
	return
}

// CreateRawSignature creates a standard TX signature but uses the message bytes as provided instead of the typical approach
// of building a signing structure with sequence, chain-id, and account number.  This approach is required for independent
// signatures like those used in the contract memorialize process which are independent of blockchain tx and their replay protection.
func (k Keeper) CreateRawSignature(txf clienttx.Factory, name string, txBuilder client.TxBuilder, message []byte, appendSignature bool) error {
	key, err := txf.Keybase().Key(name)
	if err != nil {
		return err
	}

	pubKey := key.GetPubKey()
	var prevSignatures []signing.SignatureV2
	if appendSignature {
		prevSignatures, err = txBuilder.GetTx().GetSignaturesV2()
		if err != nil {
			return err
		}
	}

	sigBytes, _, err := txf.Keybase().Sign(name, message)
	if err != nil {
		return err
	}

	// Construct the SignatureV2 struct
	sig := signing.SignatureV2{
		PubKey: pubKey,
		Data: &signing.SingleSignatureData{
			SignMode:  signing.SignMode_SIGN_MODE_UNSPECIFIED, // We are performing a custom signature that can't be validated in the normal way
			Signature: sigBytes,
		},
		Sequence: txf.Sequence(),
	}

	if !appendSignature {
		return txBuilder.SetSignatures(sig)
	}
	prevSignatures = append(prevSignatures, sig)
	return txBuilder.SetSignatures(prevSignatures...)
}
