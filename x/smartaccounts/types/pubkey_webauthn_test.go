package types

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"testing"

	"github.com/go-webauthn/webauthn/protocol/webauthncose"
	"github.com/stretchr/testify/assert"
)

// TestEC2SignatureVerification verifies the signature for EC2 keys.
func TestEC2SignatureVerification(t *testing.T) {
	pubX, err := hex.DecodeString("f739f8c77b32f4d5f13265861febd76e7a9c61a1140d296b8c16302508870316")
	assert.Nil(t, err)
	pubY, err := hex.DecodeString("c24970ad7811ccd9da7f1b88f202bebac770663ef58ba68346186dd778200dd4")
	assert.Nil(t, err)

	key := EC2PublicKeyData{
		PublicKeyData: &PublicKeyData{
			KeyType:   2,  // EC.
			Algorithm: -7, // "ES256".
		},
		Curve:  1, // P-256.
		XCoord: pubX,
		YCoord: pubY,
	}

	data := []byte("webauthnFTW")

	validSig, err := hex.DecodeString("3045022053584980793ee4ec01d583f303604c4f85a7e87df3fe9551962c5ab69a5ce27b022100c801fd6186ca4681e87fbbb97c5cb659f039473995a75a9a9dffea2708d6f8fb")
	assert.Nil(t, err)

	ok := key.VerifySignature(data, validSig)
	assert.True(t, ok, "invalid EC signature")

	ok = key.VerifySignature([]byte("webauthnFTL"), validSig)
	assert.False(t, ok, "verification against bad data is successful!")
}

// TestEdDSASignatureVerification verifies the signature for EdDSA keys.
func TestEdDSASignatureVerification(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	assert.Nil(t, err)

	data := []byte("Sample data to sign")
	validSig := ed25519.Sign(priv, data)
	invalidSig := []byte("invalid")

	key := EdDSAPublicKeyData{
		PublicKeyData: &PublicKeyData{
			KeyType:   int64(webauthncose.OctetKey),
			Algorithm: int64(webauthncose.AlgEdDSA),
		},
		XCoord: pub,
	}

	ok := key.VerifySignature(data, validSig)
	assert.True(t, ok, "valid signature wasn't properly verified")

	ok = key.VerifySignature(data, invalidSig)
	assert.False(t, ok, "invalid signature was incorrectly verified")
}
