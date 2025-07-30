package types

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/url"
	"strings"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/protocol/webauthncose"

	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/codec"
	sdktypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// parseCredentialCreationResponse decodes and parses the stored credential creation response JSON.
func (authenticator *Fido2Authenticator) parseCredentialCreationResponse(ctx sdk.Context) (*protocol.ParsedCredentialCreationData, error) {
	rawAttestationObject, err := base64.RawURLEncoding.DecodeString(authenticator.GetCredentialCreationResponse())
	if err != nil {
		ctx.Logger().Debug("Failed to decode base64 credential creation response: %v", err)
		return nil, errorsmod.Wrap(ErrParseCredential, "failed to decode base64 credential creation response")
	}

	body := io.NopCloser(bytes.NewReader(rawAttestationObject))
	defer body.Close()

	parsed, err := protocol.ParseCredentialCreationResponseBody(body)
	if err != nil {
		ctx.Logger().Error("Failed to parse credential creation response body: %v", err)
		return nil, errorsmod.Wrap(ErrParseCredential, "failed to parse credential creation response body")
	}
	return parsed, nil
}

func (authenticator *Fido2Authenticator) VerifySignature(ctx sdk.Context, _ Credential, signBytes []byte, _ codec.Codec, signature []byte) error {
	// Compute the SHA-256 hash of the raw bytes and then have the user sign it via a FIDO2 device.
	fmt.Printf("the base 64 is %s", base64.RawURLEncoding.EncodeToString(signBytes))
	hash := sha256.Sum256(signBytes)
	txHash := strings.ToUpper(hex.EncodeToString(hash[:]))

	// The io.NopCloser function wraps the bytes.Reader with an io.ReadCloser interface, which adds a no-op Close method.
	bodyAssertion := io.NopCloser(bytes.NewReader(signature))
	par, errFromParseAssertion := protocol.ParseCredentialRequestResponseBody(bodyAssertion)
	if errFromParseAssertion != nil {
		return errorsmod.Wrapf(ErrParseCredential, "Failed to parse credential request response: %v", errFromParseAssertion)
	}
	challenge := base64.RawURLEncoding.EncodeToString([]byte(txHash))

	// Parse the stored credential creation data to get the public key needed for verification.
	actual, err := authenticator.parseCredentialCreationResponse(ctx)
	if err != nil {
		return err // Error is already wrapped in the helper function.
	}
	// Create empty slices for rpOrigins and rpTopOrigins
	rpOrigins := []string{authenticator.RpOrigin}
	rpTopOrigins := []string{}

	// Handle steps 4 through 16 of webauthn spec
	// example attestation object
	// None Attestation - MacOS TouchID.
	//	`success`: `{
	//	"id":"AI7D5q2P0LS-Fal9ZT7CHM2N5BLbUunF92T8b6iYC199bO2kagSuU05-5dZGqb1SP0A0lyTWng",
	//	"rawId":"AI7D5q2P0LS-Fal9ZT7CHM2N5BLbUunF92T8b6iYC199bO2kagSuU05-5dZGqb1SP0A0lyTWng",
	//	"clientExtensionResults":{"appID":"example.com"},
	//	"type":"public-key",
	//	"response":{
	//		"authenticatorData":"dKbqkhPJnC90siSSsyDPQCYqlMGpUKA5fyklC2CEHvBFXJJiGa3OAAI1vMYKZIsLJfHwVQMANwCOw-atj9C0vhWpfWU-whzNjeQS21Lpxfdk_G-omAtffWztpGoErlNOfuXWRqm9Uj9ANJck1p6lAQIDJiABIVggKAhfsdHcBIc0KPgAcRyAIK_-Vi-nCXHkRHPNaCMBZ-4iWCBxB8fGYQSBONi9uvq0gv95dGWlhJrBwCsj_a4LJQKVHQ",
	//		"clientDataJSON":"eyJjaGFsbGVuZ2UiOiJFNFBUY0lIX0hmWDFwQzZTaWdrMVNDOU5BbGdlenROMDQzOXZpOHpfYzlrIiwibmV3X2tleXNfbWF5X2JlX2FkZGVkX2hlcmUiOiJkbyBub3QgY29tcGFyZSBjbGllbnREYXRhSlNPTiBhZ2FpbnN0IGEgdGVtcGxhdGUuIFNlZSBodHRwczovL2dvby5nbC95YWJQZXgiLCJvcmlnaW4iOiJodHRwczovL3dlYmF1dGhuLmlvIiwidHlwZSI6IndlYmF1dGhuLmdldCJ9",
	//		"signature":"MEUCIBtIVOQxzFYdyWQyxaLR0tik1TnuPhGVhXVSNgFwLmN5AiEAnxXdCq0UeAVGWxOaFcjBZ_mEZoXqNboY5IkQDdlWZYc",
	//		"userHandle":"0ToAAAAAAAAAAA"}
	//	}
	//	}`,
	// Example challenge
	// expectedChallenge := "REbD5j1tq8h226bQ_vk2ROrToE2CyeuJ3faBTKhfnQk"

	// Encode the challenge using base64 URL encoding
	// encodedChallenge := base64.RawURLEncoding.EncodeToString([]byte(expectedChallenge))
	// typecast to ECDSA public key from any cred.GetPublicKey()
	// Example challenge
	// expectedChallenge := "REbD5j1tq8h226bQ_vk2ROrToE2CyeuJ3faBTKhfnQk"

	// Encode the challenge using base64 URL encoding
	// encodedChallenge := base64.RawURLEncoding.EncodeToString([]byte(expectedChallenge))
	// typecast to ECDSA public key from any cred.GetPublicKey()
	attestationPubKey := actual.Response.AttestationObject.AuthData.AttData.CredentialPublicKey
	// The Verify method signature changed in go-webauthn v0.13.0, adding the `verifyUserPresence` parameter.
	// We set both `verifyUser` and `verifyUserPresence` to true as is standard for FIDO2 UV credentials.
	validError := par.Verify(challenge, authenticator.RpId, rpOrigins, rpTopOrigins, protocol.TopOriginDefaultVerificationMode, "", true, true, attestationPubKey)

	if validError != nil {
		return errorsmod.Wrapf(ErrParseCredential, "Failed to verify credential: %v", validError)
	}
	return nil
}

// Default returns whether this is a default credential
func (cred *Credential) Default() bool {
	return false
}

// ParsePublicKey takes the attestation public key, parses it, and returns Any and an error.
func ParsePublicKey(attestationPubKey []byte) (*sdktypes.Any, error) {
	parsedKey, err := webauthncose.ParsePublicKey(attestationPubKey)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal COSE key: %w", err)
	}

	switch key := parsedKey.(type) {
	case webauthncose.EC2PublicKeyData:
		anyPubKey, err := sdktypes.NewAnyWithValue(&EC2PublicKeyData{
			PublicKeyData: &PublicKeyData{
				PublicKey: attestationPubKey,
				KeyType:   key.KeyType,
				Algorithm: key.Algorithm,
			},
			Curve:  key.Curve,
			XCoord: key.XCoord,
			YCoord: key.YCoord,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create Any with EC2PublicKeyData: %w", err)
		}
		return anyPubKey, nil

	case webauthncose.OKPPublicKeyData:
		anyPubKey, err := sdktypes.NewAnyWithValue(&EdDSAPublicKeyData{
			PublicKeyData: &PublicKeyData{
				PublicKey: attestationPubKey,
				KeyType:   key.KeyType,
				Algorithm: key.Algorithm,
			},
			Curve: key.Curve,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create Any with OKPPublicKeyData: %w", err)
		}
		return anyPubKey, nil

	default:
		return nil, fmt.Errorf("unsupported key type: %T", parsedKey)
	}
}

// MakeNewFidoCredential creates a new FIDO2 Credential object.
func MakeNewFidoCredential(ctx sdk.Context, credentialRequestResponsesJson string, username string, clientId string) (*Credential, error) {
	body := io.NopCloser(bytes.NewReader([]byte(credentialRequestResponsesJson)))
	actualParsedCredentialCreationResponseBody, err := protocol.ParseCredentialCreationResponseBody(body)
	if err != nil {
		return nil, err
	}
	rpid, err := url.Parse(actualParsedCredentialCreationResponseBody.Response.CollectedClientData.Origin)
	if err != nil {
		return nil, err
	}
	anyPubKey, err := ParsePublicKey(actualParsedCredentialCreationResponseBody.Response.AttestationObject.AuthData.AttData.CredentialPublicKey)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// Create base credential with only the base fields
	baseCredential := &BaseCredential{
		PublicKey: anyPubKey,
		Variant:   CredentialType_CREDENTIAL_TYPE_WEBAUTHN_UV,
		// Set credential_number and create_time as needed
		CreateTime: ctx.BlockTime().Unix(),
	}

	// Create FidoCredential with the FIDO-specific fields
	fido2Authenticator := &Fido2Authenticator{
		Id:                         actualParsedCredentialCreationResponseBody.ID,
		Username:                   username,
		Aaguid:                     actualParsedCredentialCreationResponseBody.Response.AttestationObject.AuthData.AttData.AAGUID,
		CredentialCreationResponse: base64.RawURLEncoding.EncodeToString([]byte(credentialRequestResponsesJson)),
		RpId:                       rpid.Hostname(),
		RpOrigin:                   actualParsedCredentialCreationResponseBody.Response.CollectedClientData.Origin,
	}

	credential := &Credential{
		BaseCredential: baseCredential,
		Authenticator: &Credential_Fido2Authenticator{
			Fido2Authenticator: fido2Authenticator,
		},
	}
	return credential, nil
}
