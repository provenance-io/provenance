package types

import (
	"github.com/go-webauthn/webauthn/protocol/webauthncose"
)

// VerifySignature verifies the signature for the given message.
func (key *EC2PublicKeyData) VerifySignature(msg, sig []byte) bool {
	// Create an instance of webauthncose.EC2PublicKeyData
	webauthnKey := webauthncose.EC2PublicKeyData{
		PublicKeyData: webauthncose.PublicKeyData{
			KeyType:   key.PublicKeyData.KeyType,
			Algorithm: key.PublicKeyData.Algorithm,
		},
		Curve:  key.Curve,
		XCoord: key.XCoord,
		YCoord: key.YCoord,
	}

	// Use the Verify method from webauthncose to validate the signature
	valid, err := webauthnKey.Verify(msg, sig)
	if err != nil {
		return false
	}
	return valid
}

// VerifySignature verifies the signature for the given message.
func (key *EdDSAPublicKeyData) VerifySignature(msg, sig []byte) bool {
	// Create an instance of webauthncose.OKPPublicKeyData
	webauthnKey := webauthncose.OKPPublicKeyData{
		PublicKeyData: webauthncose.PublicKeyData{
			KeyType:   key.PublicKeyData.KeyType,
			Algorithm: key.PublicKeyData.Algorithm,
		},
		Curve:  key.Curve,
		XCoord: key.XCoord,
	}

	// Use the Verify method from webauthncose to validate the signature
	valid, err := webauthnKey.Verify(msg, sig)
	if err != nil {
		return false
	}
	return valid
}
