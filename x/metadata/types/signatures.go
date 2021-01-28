package types

import (
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/crypto/types/multisig"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
)

// ValidateContractSignature takes a sha512 hash of the contract bytes and verifies the signature instance is valid
// for it, returns the address of the user who created the signature and any encountered errors
func ValidateContractSignature(signature signing.SignatureDescriptor, sha512digest []byte) (addr sdk.AccAddress, err error) {
	var pubKey cryptotypes.PubKey
	if err = ModuleCdc.UnpackAny(signature.PublicKey, pubKey); err != nil {
		return
	}
	addr = sdk.AccAddress(pubKey.Address())

	sigData := signing.SignatureDataFromProto(signature.Data)
	switch data := sigData.(type) {
	case *signing.SingleSignatureData:
		if !pubKey.VerifySignature(sha512digest, data.Signature) {
			err = fmt.Errorf("unable to verify single signer signature")
		}
	case *signing.MultiSignatureData:
		multiPK, ok := pubKey.(multisig.PubKey)
		if !ok {
			err = fmt.Errorf("expected %T, got %T", (multisig.PubKey)(nil), pubKey)
		}
		err = multiPK.VerifyMultisignature(func(mode signing.SignMode) ([]byte, error) {
			// the signing mode is specific to this type of contract signature, we are not using the sequence
			// number construct here because these signatures do not count towards the committed signed tx count

			// as the sequence number and chain-id are not included the signing bytes are always the same digest
			return sha512digest, err
		}, data)
	default:
		err = fmt.Errorf("unexpected SignatureData %T", sigData)
	}
	return
}

// RecoverPublicKey recovers a tendermint secp256k1 public key from a signtaure and message hash.
func RecoverPublicKey(sig, hash []byte) (cryptotypes.PubKey, sdk.AccAddress, error) {
	// Recover public key
	pk, _, err := btcec.RecoverCompact(btcec.S256(), sig, hash)
	if err != nil {
		return nil, nil, err
	}
	// Create tendermint public key type and return with address.
	pubKey := secp256k1.PubKey{}
	copy(pubKey.Key, pk.SerializeCompressed())
	return &pubKey, pubKey.Address().Bytes(), nil
}
