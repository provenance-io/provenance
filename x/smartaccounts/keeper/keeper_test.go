package keeper_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttime "github.com/cometbft/cometbft/types/time"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/protocol/webauthncose"
	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/smartaccounts/keeper"
	smartaccounttypes "github.com/provenance-io/provenance/x/smartaccounts/types"
	"github.com/provenance-io/provenance/x/smartaccounts/utils"
	"github.com/stretchr/testify/require"
	"io"
	"log"
	"testing"
)

type TestSuite struct {
	keeper               keeper.Keeper
	ctx                  sdk.Context
	app                  *simapp.App
	addrs                []sdk.AccAddress
	govModAddr           string
	msgServer            smartaccounttypes.MsgServer
	queryServer          smartaccounttypes.QueryServer
	maxCredentialAllowed uint32
}

func (s *TestSuite) SetupTest(t *testing.T) {
	app := simapp.Setup(t)
	ctx := app.BaseApp.NewContext(false)

	now := cmttime.Now()
	ctx = ctx.WithBlockHeader(cmtproto.Header{Time: now})

	s.ctx = ctx
	s.app = app
	s.keeper = app.SmartAccountKeeper
	s.govModAddr = authtypes.NewModuleAddress(govtypes.ModuleName).String()
	s.addrs = simtestutil.CreateIncrementalAccounts(3)

	if s.maxCredentialAllowed == 0 {
		s.maxCredentialAllowed = 10
	}
	// Set custom params for tests
	customParams := smartaccounttypes.Params{
		Enabled:              true,
		MaxCredentialAllowed: s.maxCredentialAllowed, // Set custom value for testing
	}

	// Apply params directly to the keeper
	err := s.keeper.SetParams(ctx, customParams)
	require.NoError(t, err)

	s.msgServer = keeper.NewMsgServerImpl(s.keeper)
	s.queryServer = keeper.NewQuerier(s.keeper)
}

func TestKeeper(t *testing.T) {
	suite := &TestSuite{}
	suite.SetupTest(t)

	t.Run("InitWithEmptyCredentials", suite.InitWithEmptyCredentials)
	t.Run("InitWithValidPubKey", suite.InitWithValidPubKey)
	t.Run("TestCreateCredential", func(t *testing.T) {
		suite := &TestSuite{}
		suite.SetupTest(t)
		suite.TestCreateCredential(t)
	})
	t.Run("TestCreateCredentialAndDelete", func(t *testing.T) {
		suite := &TestSuite{}
		suite.SetupTest(t)
		suite.TestCreateCredentialAndDelete(t)
	})
}

// createBaseAccount creates a new base account with the provided public key and sets it in the account keeper
func (s *TestSuite) createBaseAccount(pubKey cryptotypes.PubKey) *authtypes.BaseAccount {
	accountNum := s.app.AccountKeeper.NextAccountNumber(s.ctx)
	baseAcc := authtypes.NewBaseAccount(sdk.AccAddress(pubKey.Address()), pubKey, accountNum, 0)
	s.app.AccountKeeper.SetAccount(s.ctx, baseAcc)
	return baseAcc
}

// This test simulates the creation of a smart account with no credentials, giving no error.
// This i think is a good hook for an existing account creating a smart account to add credentials later.
func (s *TestSuite) InitWithEmptyCredentials(t *testing.T) {
	privKey := secp256k1.GenPrivKey()
	pubKeyv := privKey.PubKey().(*secp256k1.PubKey)
	pubKey := &secp256k1.PubKey{
		Key: pubKeyv.Key,
	}

	// Use the helper function
	s.createBaseAccount(pubKey)
	msg := &smartaccounttypes.MsgInit{
		Sender:      sdk.AccAddress(pubKey.Address()).String(),
		Credentials: []*smartaccounttypes.Credential{},
	}

	resp, err := s.keeper.Init(s.ctx, msg)
	require.Error(t, err)
	require.Nil(t, resp)
}

// simulates using a secp256k1 public key to create a smart account
func (s *TestSuite) InitWithValidPubKey(t *testing.T) {
	privKey := secp256k1.GenPrivKey()
	pubKeyv := privKey.PubKey().(*secp256k1.PubKey)
	pubKey := &secp256k1.PubKey{
		Key: pubKeyv.Key,
	}
	pubKeyAny, err := types.NewAnyWithValue(pubKey)
	s.createBaseAccount(pubKey)
	baseCredential := smartaccounttypes.BaseCredential{
		PublicKey: pubKeyAny,
		Variant:   smartaccounttypes.CredentialType_CREDENTIAL_TYPE_K256,
	}

	credential := smartaccounttypes.Credential{
		BaseCredential: &baseCredential,
	}

	msg := &smartaccounttypes.MsgInit{
		Sender:      sdk.AccAddress(pubKey.Address()).String(),
		Credentials: []*smartaccounttypes.Credential{&credential},
	}

	resp, err := s.keeper.Init(s.ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 1, len(resp.Credentials))
}

// from https://github.com/go-webauthn/webauthn/blob/67139112f304e5b9bf38fdc5fe9438b785fe2d56/protocol/credential_test.go#L37
func (s *TestSuite) TestCreateCredential(t *testing.T) {

	body := io.NopCloser(bytes.NewReader([]byte(utils.TestCredentialRequestResponses["success"])))
	actual, err := protocol.ParseCredentialCreationResponseBody(body)
	require.NoError(t, err)

	byteID, _ := base64.RawURLEncoding.DecodeString("6xrtBhJQW6QU4tOaB4rrHaS2Ks0yDDL_q8jDC16DEjZ-VLVf4kCRkvl2xp2D71sTPYns-exsHQHTy3G-zJRK8g")
	byteAuthData, _ := base64.RawURLEncoding.DecodeString("o2NmbXRkbm9uZWdhdHRTdG10oGhhdXRoRGF0YVjEdKbqkhPJnC90siSSsyDPQCYqlMGpUKA5fyklC2CEHvBBAAAAAAAAAAAAAAAAAAAAAAAAAAAAQOsa7QYSUFukFOLTmgeK6x2ktirNMgwy_6vIwwtegxI2flS1X-JAkZL5dsadg-9bEz2J7PnsbB0B08txvsyUSvKlAQIDJiABIVggLKF5xS0_BntttUIrm2Z2tgZ4uQDwllbdIfrrBMABCNciWCDHwin8Zdkr56iSIh0MrB5qZiEzYLQpEOREhMUkY6q4Vw")
	byteRPIDHash := sha256.Sum256([]byte("https://webauthn.io"))
	byteCredentialPubKey := []byte{0x04, 0x88, 0x25, 0x8d, 0x1c, 0xa2, 0x9f, 0xc6, 0x5d, 0x92, 0xbe, 0x7a, 0x89, 0x22, 0x21, 0xd0, 0xca, 0xc1, 0xe6, 0xa6, 0x62, 0x13, 0x36, 0x0b, 0x42, 0x91, 0x0e, 0x44, 0x48, 0x4c, 0x52, 0x46, 0x3a, 0xab, 0x85, 0x70}
	byteClientDataJSON, _ := base64.RawURLEncoding.DecodeString("eyJjaGFsbGVuZ2UiOiJXOEd6RlU4cEdqaG9SYldyTERsYW1BZnFfeTRTMUNaRzFWdW9lUkxBUnJFIiwib3JpZ2luIjoiaHR0cHM6Ly93ZWJhdXRobi5pbyIsInR5cGUiOiJ3ZWJhdXRobi5jcmVhdGUifQ")
	byteAttObject, _ := base64.RawURLEncoding.DecodeString("o2NmbXRkbm9uZWdhdHRTdG10oGhhdXRoRGF0YVjEdKbqkhPJnC90siSSsyDPQCYqlMGpUKA5fyklC2CEHvBBAAAAAAAAAAAAAAAAAAAAAAAAAAAAQOsa7QYSUFukFOLTmgeK6x2ktirNMgwy_6vIwwtegxI2flS1X-JAkZL5dsadg-9bEz2J7PnsbB0B08txvsyUSvKlAQIDJiABIVggLKF5xS0_BntttUIrm2Z2tgZ4uQDwllbdIfrrBMABCNciWCDHwin8Zdkr56iSIh0MrB5qZiEzYLQpEOREhMUkY6q4Vw")

	expected := &protocol.ParsedCredentialCreationData{
		ParsedPublicKeyCredential: protocol.ParsedPublicKeyCredential{
			ParsedCredential: protocol.ParsedCredential{
				ID:   "6xrtBhJQW6QU4tOaB4rrHaS2Ks0yDDL_q8jDC16DEjZ-VLVf4kCRkvl2xp2D71sTPYns-exsHQHTy3G-zJRK8g",
				Type: "public-key",
			},
			RawID: byteID,
			ClientExtensionResults: protocol.AuthenticationExtensionsClientOutputs{
				"appid": true,
			},
			AuthenticatorAttachment: protocol.Platform,
		},
		Response: protocol.ParsedAttestationResponse{
			CollectedClientData: protocol.CollectedClientData{
				Type:      protocol.CeremonyType("webauthn.create"),
				Challenge: "W8GzFU8pGjhoRbWrLDlamAfq_y4S1CZG1VuoeRLARrE",
				Origin:    "https://webauthn.io",
			},
			AttestationObject: protocol.AttestationObject{
				Format:      "none",
				RawAuthData: byteAuthData,
				AuthData: protocol.AuthenticatorData{
					RPIDHash: byteRPIDHash[:],
					Counter:  0,
					Flags:    0x041,
					AttData: protocol.AttestedCredentialData{
						AAGUID:              make([]byte, 16),
						CredentialID:        byteID,
						CredentialPublicKey: byteCredentialPubKey,
					},
				},
			},
			Transports: []protocol.AuthenticatorTransport{protocol.USB, protocol.NFC, "fake"},
		},
		Raw: protocol.CredentialCreationResponse{
			PublicKeyCredential: protocol.PublicKeyCredential{
				Credential: protocol.Credential{
					Type: "public-key",
					ID:   "6xrtBhJQW6QU4tOaB4rrHaS2Ks0yDDL_q8jDC16DEjZ-VLVf4kCRkvl2xp2D71sTPYns-exsHQHTy3G-zJRK8g",
				},
				RawID: byteID,
				ClientExtensionResults: protocol.AuthenticationExtensionsClientOutputs{
					"appid": true,
				},
				AuthenticatorAttachment: "platform",
			},
			AttestationResponse: protocol.AuthenticatorAttestationResponse{
				AuthenticatorResponse: protocol.AuthenticatorResponse{
					ClientDataJSON: byteClientDataJSON,
				},
				AttestationObject: byteAttObject,
				Transports:        []string{"usb", "nfc", "fake"},
			},
		},
	}
	require.Equal(t, expected.ClientExtensionResults, actual.ClientExtensionResults)
	require.Equal(t, expected.ID, actual.ID)
	require.Equal(t, expected.Type, actual.Type)
	require.Equal(t, expected.ParsedCredential, actual.ParsedCredential)
	require.Equal(t, expected.ParsedPublicKeyCredential, actual.ParsedPublicKeyCredential)
	require.Equal(t, expected.ParsedPublicKeyCredential, actual.ParsedPublicKeyCredential)
	require.Equal(t, expected.Raw, actual.Raw)
	require.Equal(t, expected.RawID, actual.RawID)
	require.Equal(t, expected.Response.Transports, actual.Response.Transports)
	require.Equal(t, expected.Response.CollectedClientData, actual.Response.CollectedClientData)
	require.Equal(t, expected.Response.AttestationObject.AuthData.AttData.CredentialID, actual.Response.AttestationObject.AuthData.AttData.CredentialID)
	require.Equal(t, expected.Response.AttestationObject.Format, actual.Response.AttestationObject.Format)
	require.Equal(t, expected.Response.CollectedClientData.Origin, actual.Response.CollectedClientData.Origin)

	// Parse the COSE public key
	parsedKey, err := webauthncose.ParsePublicKey(actual.Response.AttestationObject.AuthData.AttData.CredentialPublicKey)
	if err != nil {
		log.Fatalf("Failed to unmarshal COSE key: %v", err)
	}
	// Type assertion to EC2PublicKeyData
	ec2Key, ok := parsedKey.(webauthncose.EC2PublicKeyData)
	if !ok {
		log.Fatalf("Failed to cast to EC2PublicKeyData")
	}

	attestationPubKey := actual.Response.AttestationObject.AuthData.AttData.CredentialPublicKey

	anyPubKey, err := types.NewAnyWithValue(&smartaccounttypes.EC2PublicKeyData{
		PublicKeyData: &smartaccounttypes.PublicKeyData{
			PublicKey: attestationPubKey,
			KeyType:   ec2Key.KeyType,
			Algorithm: ec2Key.Algorithm,
		},
		Curve:  ec2Key.Curve,
		XCoord: ec2Key.XCoord,
		YCoord: ec2Key.YCoord,
	})

	// Serialize the CredentialCreationResponse to bytes
	attestation, err := json.Marshal(actual)
	if err != nil {
		log.Fatalf("Failed to marshal CredentialCreationResponse: %v", err)
	}
	require.NoError(t, err)

	if (actual.Response.AttestationObject.AuthData.Flags & 0x04) == 0x04 {
		fmt.Println("UV flag is set")
	} else {
		fmt.Println("UV flag is not set")
	}

	baseCredential := smartaccounttypes.BaseCredential{
		PublicKey: anyPubKey,
		Variant:   smartaccounttypes.CredentialType_CREDENTIAL_TYPE_WEBAUTHN,
	}

	// Create Fido2Authenticator with the FIDO-specific fields
	fido2Authenticator := &smartaccounttypes.Fido2Authenticator{
		Id:                         actual.ID,
		Username:                   "example_username",
		Aaguid:                     actual.Response.AttestationObject.AuthData.AttData.AAGUID,
		CredentialCreationResponse: base64.RawURLEncoding.EncodeToString(attestation),
		RpId:                       "webauthn.io", // Extract from the origin
		RpOrigin:                   actual.Response.CollectedClientData.Origin,
	}

	credential := &smartaccounttypes.Credential{
		BaseCredential: &baseCredential,
		Authenticator: &smartaccounttypes.Credential_Fido2Authenticator{
			Fido2Authenticator: fido2Authenticator,
		},
	}

	require.NoError(t, err)

	// compare credential id and raw id
	require.Equal(t, "6xrtBhJQW6QU4tOaB4rrHaS2Ks0yDDL_q8jDC16DEjZ-VLVf4kCRkvl2xp2D71sTPYns-exsHQHTy3G-zJRK8g", fido2Authenticator.Id)

	privkey1 := secp256k1.GenPrivKey()
	pubkey1 := privkey1.PubKey()
	owner1Addr := sdk.AccAddress(pubkey1.Address())
	fmt.Printf("%s", owner1Addr.String())

	accountNum := s.app.AccountKeeper.NextAccountNumber(s.ctx)
	// Create and set the base account before initializing the smart account
	baseAcc := authtypes.NewBaseAccount(owner1Addr, pubkey1, accountNum, 0)
	s.app.AccountKeeper.SetAccount(s.ctx, baseAcc)

	msg := &smartaccounttypes.MsgInit{
		Sender:      owner1Addr.String(),
		Credentials: []*smartaccounttypes.Credential{credential},
	}

	resp, err := s.keeper.Init(s.ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 1, len(resp.Credentials))
	require.Equal(t, owner1Addr.String(), resp.Address)

	// Fetch the smart account to verify it was created
	account, err := s.keeper.LookupAccountByAddress(s.ctx, owner1Addr)
	require.NoError(t, err, "should be able to fetch the smart account")
	require.Equal(t, owner1Addr.String(), account.Address, "account address should match")
	require.Len(t, account.Credentials, 1, "should have one credential")
	require.Equal(t, uint64(0), account.Credentials[0].CredentialNumber, "credential number should match")

}

func (s *TestSuite) TestCreateCredentialAndDelete(t *testing.T) {

	body := io.NopCloser(bytes.NewReader([]byte(utils.TestCredentialRequestResponses["success"])))
	actual, err := protocol.ParseCredentialCreationResponseBody(body)
	require.NoError(t, err)

	// Parse the COSE public key
	parsedKey, err := webauthncose.ParsePublicKey(actual.Response.AttestationObject.AuthData.AttData.CredentialPublicKey)
	if err != nil {
		log.Fatalf("Failed to unmarshal COSE key: %v", err)
	}
	// Type assertion to EC2PublicKeyData
	ec2Key, ok := parsedKey.(webauthncose.EC2PublicKeyData)
	if !ok {
		log.Fatalf("Failed to cast to EC2PublicKeyData")
	}

	attestationPubKey := actual.Response.AttestationObject.AuthData.AttData.CredentialPublicKey

	anyPubKey, err := types.NewAnyWithValue(&smartaccounttypes.EC2PublicKeyData{
		PublicKeyData: &smartaccounttypes.PublicKeyData{
			PublicKey: attestationPubKey,
			KeyType:   ec2Key.KeyType,
			Algorithm: ec2Key.Algorithm,
		},
		Curve:  ec2Key.Curve,
		XCoord: ec2Key.XCoord,
		YCoord: ec2Key.YCoord,
	})

	// Serialize the CredentialCreationResponse to bytes
	attestation, err := json.Marshal(actual)
	if err != nil {
		log.Fatalf("Failed to marshal CredentialCreationResponse: %v", err)
	}
	require.NoError(t, err)

	if (actual.Response.AttestationObject.AuthData.Flags & 0x04) == 0x04 {
		fmt.Println("UV flag is set")
	} else {
		fmt.Println("UV flag is not set")
	}

	baseCredential := smartaccounttypes.BaseCredential{
		PublicKey: anyPubKey,
		Variant:   smartaccounttypes.CredentialType_CREDENTIAL_TYPE_WEBAUTHN,
	}

	// Create Fido2Authenticator with the FIDO-specific fields
	fido2Authenticator := &smartaccounttypes.Fido2Authenticator{
		Id:                         actual.ID,
		Username:                   "example_username",
		Aaguid:                     actual.Response.AttestationObject.AuthData.AttData.AAGUID,
		CredentialCreationResponse: base64.RawURLEncoding.EncodeToString(attestation),
		RpId:                       "webauthn.io", // Extract from the origin
		RpOrigin:                   actual.Response.CollectedClientData.Origin,
	}

	credential := &smartaccounttypes.Credential{
		BaseCredential: &baseCredential,
		Authenticator: &smartaccounttypes.Credential_Fido2Authenticator{
			Fido2Authenticator: fido2Authenticator,
		},
	}

	require.NoError(t, err)

	// compare credential id and raw id
	require.Equal(t, "6xrtBhJQW6QU4tOaB4rrHaS2Ks0yDDL_q8jDC16DEjZ-VLVf4kCRkvl2xp2D71sTPYns-exsHQHTy3G-zJRK8g", fido2Authenticator.Id)

	privkey1 := secp256k1.GenPrivKey()
	pubkey1 := privkey1.PubKey()
	owner1Addr := sdk.AccAddress(pubkey1.Address())
	fmt.Printf("%s", owner1Addr.String())

	accountNum := s.app.AccountKeeper.NextAccountNumber(s.ctx)
	// Create and set the base account before initializing the smart account
	baseAcc := authtypes.NewBaseAccount(owner1Addr, pubkey1, accountNum, 0)
	s.app.AccountKeeper.SetAccount(s.ctx, baseAcc)

	msg := &smartaccounttypes.MsgInit{
		Sender:      owner1Addr.String(),
		Credentials: []*smartaccounttypes.Credential{credential},
	}

	resp, err := s.keeper.Init(s.ctx, msg)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 1, len(resp.Credentials))
	require.Equal(t, owner1Addr.String(), resp.Address)

	// Fetch the smart account to verify it was created
	account, err := s.keeper.LookupAccountByAddress(s.ctx, owner1Addr)
	require.NoError(t, err, "should be able to fetch the smart account")
	require.Equal(t, owner1Addr.String(), account.Address, "account address should match")
	require.Len(t, account.Credentials, 1, "should have one credential")
	require.Equal(t, uint64(0), account.Credentials[0].CredentialNumber, "credential number should match")

	// Delete the credential
	respFromDelete, errFromDelete := s.keeper.DeleteCredential(s.ctx, &account, account.Credentials[0].BaseCredential.CredentialNumber)
	require.NoError(t, errFromDelete)
	require.NotNil(t, respFromDelete)

	// Fetch the smart account to verify it was created
	account, err = s.keeper.LookupAccountByAddress(s.ctx, owner1Addr)
	require.NoError(t, err, "should be able to fetch the smart account")
	require.Equal(t, owner1Addr.String(), account.Address, "account address should match")
	require.Len(t, account.Credentials, 0, "should not have any credential")
}
