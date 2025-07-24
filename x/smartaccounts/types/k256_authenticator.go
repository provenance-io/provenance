package types

import (
	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (authenticator *K256Authenticator) VerifySignature(ctx sdk.Context, credential Credential, signBytes []byte, codec codec.Codec, signature []byte) error {
	// Get the Any-wrapped public key
	anyPubKey := credential.GetPublicKey()
	if anyPubKey == nil {
		return errorsmod.Wrapf(ErrParseCredential, "Public key is nil")
	}

	// Unpack the Any to get the concrete secp256k1 public key
	var pubKey cryptotypes.PubKey

	if err := codec.UnpackAny(anyPubKey, &pubKey); err != nil {
		return errorsmod.Wrapf(ErrParseCredential, "Failed to unpack public key: %v", err)
	}

	if !pubKey.VerifySignature(signBytes, signature) {
		return nil
	} else {
		return errorsmod.Wrapf(ErrParseCredential, "Signature verification failed")
	}
}
