package types

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/stretchr/testify/require"
	"io"
	"log"
	"net/url"
	"testing"
)

var attestation = map[string]string{
	`success`: `
{
  "id": "bbcXuKs1MaXovZbLHXIc_Q",
  "rawId": "bbcXuKs1MaXovZbLHXIc_Q",
  "response": {
    "attestationObject": "o2NmbXRkbm9uZWdhdHRTdG10oGhhdXRoRGF0YViUSZYN5YgOjGh0NBcPZHZgW4_krrmihjLHmVzzuoMdl2NdAAAAAOqbjWZNAR0hPOS2tIy1ddQAEG23F7irNTGl6L2Wyx1yHP2lAQIDJiABIVggh5JJM6P5ZONo68QMnu2mT3BpgaKNRRDDfdEZC8D0rZ4iWCBDu3kYPc7kJ6BuKNgoAs3f5KvEVgJSdmKL2iSY4s-iXw",
    "clientDataJSON": "eyJ0eXBlIjoid2ViYXV0aG4uY3JlYXRlIiwiY2hhbGxlbmdlIjoiT2hnakxpeDROYWlFY2ZzR1BsS0o4cDRTVnB5aWNVNVpZQ2xvczdudmdyYyIsIm9yaWdpbiI6Imh0dHA6Ly9sb2NhbGhvc3Q6MTgwODAifQ"
  },
  "type": "public-key",
  "authenticatorAttachment": "platform"
}
`}

var assertionResponse = map[string]string{
	`success`: `
{
  "id" : "bbcXuKs1MaXovZbLHXIc_Q",
  "type" : "public-key",
  "rawId" : "bbcXuKs1MaXovZbLHXIc_Q",
  "response" : {
    "clientDataJSON" : "eyJ0eXBlIjoid2ViYXV0aG4uZ2V0IiwiY2hhbGxlbmdlIjoiTVVZMU9FWTNRekExTmpnd1FUUkZSRGd5TmtaRk5URTFORVkwTXpBeE9VVTJPVEV6TnpVMVFVSTBNRFk1TVVWRFJFSTRNekZDTWpjMVFqY3hPRVV5TUEiLCJvcmlnaW4iOiJodHRwOi8vbG9jYWxob3N0OjE4MDgwIiwiY3Jvc3NPcmlnaW4iOmZhbHNlfQ",
    "authenticatorData" : "SZYN5YgOjGh0NBcPZHZgW4_krrmihjLHmVzzuoMdl2MdAAAAAA",
    "signature" : "MEYCIQCiU8f5qzM2CEC34PwIvlcOIC8rh8FSFwVZumLGeJg6JgIhALnLbIE1y8YnOWXzBZ5NmoCK6YRzrw_iQiST8_NPNfvh",
    "userHandle" : "6-SG3dq4hPbuAQ"
  }
}
`}
var signBytes = `CokBCoYBChwvY29zbW9zLmJhbmsudjFiZXRhMS5Nc2dTZW5kEmYKKXRwMXc0MHEzcTd2MjZwZXR3Nmc1c3R6NWR0OXhzZXpnbnphbHhndzh4Eil0cDEwZmpsZXVldmM5amE1eDdrbGc1dWY3MzA2enE4eWxwMHNucTMzNhoOCgVuaGFzaBIFMTAwMDASJgoIEgQKAggBGBcSGgoUCgVuaGFzaBILMzg0MDAwMDAwMDAQgIl6Ggd0ZXN0aW5nIAo`

func TestVerifySignature(t *testing.T) {
	body := io.NopCloser(bytes.NewReader([]byte(attestation["success"])))
	actual, err := protocol.ParseCredentialCreationResponseBody(body)
	require.NoError(t, err)
	fmt.Printf("actual: %v\n", actual)
	//check and set the rpid
	rpid, err := url.Parse(actual.Response.CollectedClientData.Origin)
	require.NoError(t, err)

	rpIDHash := sha256.Sum256([]byte(rpid.Hostname()))
	if !bytes.Equal(rpIDHash[:], actual.Response.AttestationObject.AuthData.RPIDHash) {
		t.Fatalf("RPID hash mismatch")
	}
	attestationPubKey := actual.Response.AttestationObject.AuthData.AttData.CredentialPublicKey

	anyPubKey, err := ParsePublicKey(attestationPubKey)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// Create base credential with the common fields
	baseCredential := &BaseCredential{
		PublicKey: anyPubKey,
		Variant:   CredentialType_CREDENTIAL_TYPE_WEBAUTHN_UV,
	}

	// Create Fido2Authenticator with the FIDO-specific fields
	fido2Authenticator := &Fido2Authenticator{
		Id:                         actual.ID,
		Username:                   "foo3@bar.com",
		Aaguid:                     actual.Response.AttestationObject.AuthData.AttData.AAGUID,
		CredentialCreationResponse: base64.RawURLEncoding.EncodeToString([]byte(attestation["success"])),
		RpId:                       rpid.Hostname(),
		RpOrigin:                   actual.Response.CollectedClientData.Origin,
	}

	// Create the complete credential with proper structure
	credential := &Credential{
		BaseCredential: baseCredential,
		Authenticator: &Credential_Fido2Authenticator{
			Fido2Authenticator: fido2Authenticator,
		},
	}
	bodyAssertion := io.NopCloser(bytes.NewReader([]byte(assertionResponse["success"])))
	// Making sure object can be unmarshalled.
	par, err := protocol.ParseCredentialRequestResponseBody(bodyAssertion)
	require.NoError(t, err, "Failed to parse credential request response")

	// Making sure object can be unmarshalled.
	err = json.Unmarshal(par.Raw.AssertionResponse.ClientDataJSON, &par.Response.CollectedClientData)
	require.NoError(t, err, "Failed to unmarshal client data JSON")

	// Create a mock context for testing
	ctx := sdk.Context{} // or use a proper test context if available
	signBytes, err := base64.RawURLEncoding.DecodeString(signBytes)
	require.NoError(t, err)
	err = fido2Authenticator.VerifySignature(ctx, *credential, signBytes, codec.NewProtoCodec(nil), []byte(assertionResponse["success"]))
	require.NoError(t, err)
}
