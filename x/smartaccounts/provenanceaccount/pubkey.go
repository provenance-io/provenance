package provenanceaccount

import (
	"crypto/elliptic"
	"fmt"
	"math/big"

	// decred library was already a dependency in cosmos-sdk
	dcrdsecp256k1 "github.com/decred/dcrd/dcrec/secp256k1/v4"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogoproto "github.com/cosmos/gogoproto/proto"

	smartaccounttypes "github.com/provenance-io/provenance/x/smartaccounts/types"
)

// this file implements a general mechanism to plugin public keys to a baseaccount

// PubKey defines a generic pubkey.
type PubKey interface {
	sdk.Msg
	VerifySignature(msg, sig []byte) bool
}

type PubKeyG[T any] interface {
	*T
	PubKey
}

type pubKeyImpl struct {
	Decode   func(b []byte) (PubKey, error)
	Validate func(key PubKey) error
}

func WithSecp256K1PubKey() Option {
	return WithPubKeyWithValidationFunc(func(pt *secp256k1.PubKey) error {
		if _, err := dcrdsecp256k1.ParsePubKey(pt.Key); err != nil {
			return fmt.Errorf("invalid secp256k1 public key: %w", err)
		}
		return nil
	})
}

// getCurve returns the elliptic curve corresponding to the given COSE curve identifier.
func getCurve(curveID int64) (elliptic.Curve, error) {
	switch curveID {
	case 1: // COSE P-256
		return elliptic.P256(), nil
	case 2: // COSE P-384
		return elliptic.P384(), nil
	case 3: // COSE P-521
		return elliptic.P521(), nil
	default:
		return nil, fmt.Errorf("unsupported curve identifier: %d", curveID)
	}
}

// validateWebAuthnKey validates the public key for a WebAuthn credential.
func validateWebAuthnKey[T any, PT PubKeyG[T]](key PT) error {
	switch k := any(key).(type) {
	case *smartaccounttypes.EC2PublicKeyData:
		if k.GetPublicKeyData() == nil {
			return fmt.Errorf("EC2 public key data cannot be nil")
		}
		if len(k.GetPublicKeyData().GetPublicKey()) == 0 {
			return fmt.Errorf("EC2 public key bytes cannot be empty")
		}
		if k.GetPublicKeyData().GetKeyType() == 0 {
			return fmt.Errorf("EC2 public key type cannot be 0")
		}
		if k.GetPublicKeyData().GetAlgorithm() == 0 {
			return fmt.Errorf("EC2 public key algorithm cannot be 0")
		}
		if k.GetCurve() == 0 {
			return fmt.Errorf("EC2 curve cannot be 0")
		}
		if len(k.GetXCoord()) == 0 {
			return fmt.Errorf("EC2 x-coordinate cannot be empty")
		}
		if len(k.GetYCoord()) == 0 {
			return fmt.Errorf("EC2 y-coordinate cannot be empty")
		}
		curve, err := getCurve(k.GetCurve())
		if err != nil {
			return err
		}
		x := new(big.Int).SetBytes(k.GetXCoord())
		y := new(big.Int).SetBytes(k.GetYCoord())
		if !curve.IsOnCurve(x, y) {
			return fmt.Errorf("public key coordinates are not on the specified curve")
		}
	case *smartaccounttypes.EdDSAPublicKeyData:
		if k.GetPublicKeyData() == nil {
			return fmt.Errorf("EdDSA public key data cannot be nil")
		}
		if len(k.GetPublicKeyData().GetPublicKey()) == 0 {
			return fmt.Errorf("EdDSA public key bytes cannot be empty")
		}
		if k.GetPublicKeyData().GetKeyType() == 0 {
			return fmt.Errorf("EdDSA public key type cannot be 0")
		}
		if k.GetPublicKeyData().GetAlgorithm() == 0 {
			return fmt.Errorf("EdDSA public key algorithm cannot be 0")
		}
		if k.GetCurve() == 0 {
			return fmt.Errorf("EdDSA curve cannot be 0")
		}
		if len(k.GetXCoord()) == 0 {
			return fmt.Errorf("EdDSA x-coordinate cannot be empty")
		}
	default:
		return fmt.Errorf("unsupported webauthn key type: %T", k)
	}
	return nil
}

// WithWebAuthnPubKey is an option to register a WebAuthn public key type.
func WithWebAuthnPubKey[T any, PT PubKeyG[T]]() Option {
	return WithPubKeyWithValidationFunc[T, PT](validateWebAuthnKey[T, PT])
}

func WithPubKeyWithValidationFunc[T any, PT PubKeyG[T]](validateFn func(PT) error) Option {
	pkImpl := pubKeyImpl{
		Decode: func(b []byte) (PubKey, error) {
			key := PT(new(T))
			err := gogoproto.Unmarshal(b, key)
			if err != nil {
				return nil, err
			}
			return key, nil
		},
		Validate: func(k PubKey) error {
			concrete, ok := k.(PT)
			if !ok {
				return fmt.Errorf("invalid pubkey type passed for validation, wanted: %T, got: %T", concrete, k)
			}
			return validateFn(concrete)
		},
	}
	return func(a *ProvenanceSmartAccountHandler) {
		a.SupportedPubKeys[gogoproto.MessageName(PT(new(T)))] = pkImpl
	}
}
