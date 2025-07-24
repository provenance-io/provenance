package provenanceaccount

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/provenance-io/provenance/x/smartaccounts/types"
)

func TestValidateWebAuthnKey(t *testing.T) {
	// Generate a valid P-256 key for testing
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	validXCoord := privateKey.X.Bytes()
	validYCoord := privateKey.Y.Bytes()

	// Invalid coordinates (Y is not on the curve with X)
	invalidYCoord := new(big.Int).SetBytes(validYCoord)
	invalidYCoord.Add(invalidYCoord, big.NewInt(1))

	// Generate a valid EdDSA key for testing
	validEdDSAPubKey, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	testCases := []struct {
		name        string
		key         *types.EC2PublicKeyData
		expectErr   bool
		errContains string
	}{
		{
			name: "success - valid EC2 key on P-256 curve",
			key: &types.EC2PublicKeyData{
				PublicKeyData: &types.PublicKeyData{PublicKey: []byte("pubkey"), KeyType: 2, Algorithm: -7},
				Curve:         1, // P-256
				XCoord:        validXCoord,
				YCoord:        validYCoord,
			},
			expectErr: false,
		},
		{
			name: "failure - EC2 key not on curve",
			key: &types.EC2PublicKeyData{
				PublicKeyData: &types.PublicKeyData{PublicKey: []byte("pubkey"), KeyType: 2, Algorithm: -7},
				Curve:         1, // P-256
				XCoord:        validXCoord,
				YCoord:        invalidYCoord.Bytes(),
			},
			expectErr:   true,
			errContains: "public key coordinates are not on the specified curve",
		},
		{
			name: "failure - unsupported curve identifier",
			key: &types.EC2PublicKeyData{
				PublicKeyData: &types.PublicKeyData{PublicKey: []byte("pubkey"), KeyType: 2, Algorithm: -7},
				Curve:         999, // Invalid curve
				XCoord:        validXCoord,
				YCoord:        validYCoord,
			},
			expectErr:   true,
			errContains: "unsupported curve identifier: 999",
		},
		{
			name: "failure - missing public key data",
			key: &types.EC2PublicKeyData{
				PublicKeyData: nil,
				Curve:         1,
				XCoord:        validXCoord,
				YCoord:        validYCoord,
			},
			expectErr:   true,
			errContains: "EC2 public key data cannot be nil",
		},
		{
			name: "failure - missing x coordinate",
			key: &types.EC2PublicKeyData{
				PublicKeyData: &types.PublicKeyData{PublicKey: []byte("pubkey"), KeyType: 2, Algorithm: -7},
				Curve:         1,
				XCoord:        []byte{},
				YCoord:        validYCoord,
			},
			expectErr:   true,
			errContains: "EC2 x-coordinate cannot be empty",
		},
		{
			name: "failure - missing y coordinate",
			key: &types.EC2PublicKeyData{
				PublicKeyData: &types.PublicKeyData{PublicKey: []byte("pubkey"), KeyType: 2, Algorithm: -7},
				Curve:         1,
				XCoord:        validXCoord,
				YCoord:        []byte{},
			},
			expectErr:   true,
			errContains: "EC2 y-coordinate cannot be empty",
		},
		{
			name: "failure - missing public key bytes",
			key: &types.EC2PublicKeyData{
				PublicKeyData: &types.PublicKeyData{PublicKey: []byte{}, KeyType: 2, Algorithm: -7},
				Curve:         1,
				XCoord:        validXCoord,
				YCoord:        validYCoord,
			},
			expectErr:   true,
			errContains: "EC2 public key bytes cannot be empty",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateWebAuthnKey(tc.key)
			if tc.expectErr {
				require.Error(t, err)
				require.ErrorContains(t, err, tc.errContains)
			} else {
				require.NoError(t, err)
			}
		})
	}

	// Test EdDSA separately since the generic type is different
	t.Run("success - valid EdDSA key", func(t *testing.T) {
		key := &types.EdDSAPublicKeyData{
			PublicKeyData: &types.PublicKeyData{PublicKey: []byte("pubkey"), KeyType: 1, Algorithm: -8},
			Curve:         6, // Ed25519
			XCoord:        validEdDSAPubKey,
		}
		err := validateWebAuthnKey(key)
		require.NoError(t, err)
	})

	t.Run("failure - missing x coordinate for EdDSA", func(t *testing.T) {
		key := &types.EdDSAPublicKeyData{
			PublicKeyData: &types.PublicKeyData{PublicKey: []byte("pubkey"), KeyType: 1, Algorithm: -8},
			Curve:         6,
			XCoord:        []byte{},
		}
		err := validateWebAuthnKey(key)
		require.Error(t, err)
		require.ErrorContains(t, err, "EdDSA x-coordinate cannot be empty")
	})
}
