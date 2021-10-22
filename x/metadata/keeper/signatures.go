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

// IsMarkerAndHasAuthority checks that the address is a marker addr and that one of the signers has the given role.
// First return boolean is whether or not address a marker address.
// Second return boolean is whether or not one of the signers has the given role on that marker.
// If the first return boolean is false, they'll both be false.
func (k Keeper) IsMarkerAndHasAuthority(ctx sdk.Context, address string, signers []string, role markertypes.Access) (isMarker bool, hasAuth bool) {
	addr, err := sdk.AccAddressFromBech32(address)
	// if the value owner is invalid then it is not possible to have any authority for it. e.g. value owner is empty.
	if err != nil {
		return false, false
	}

	acc := k.authKeeper.GetAccount(ctx, addr)
	if acc == nil {
		return false, false
	}

	// Convert over to the actual underlying marker type, or not.
	marker, isMarker := acc.(*markertypes.MarkerAccount)
	if !isMarker {
		return false, false
	}
	for _, signer := range signers {
		saddr, serr := sdk.AccAddressFromBech32(signer)
		// If the signer address is okay, check it for the role. If it checks out, they've got auth and we're done.
		if serr == nil && marker.AddressHasAccess(saddr, role) {
			return true, true
		}
	}
	return true, false
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
