package keeper_test

import (
	"encoding/base64"
	"errors"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/gogo/protobuf/proto"
	"github.com/provenance-io/provenance/x/smartaccounts/utils"
	"github.com/stretchr/testify/assert"
	"testing"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/smartaccounts/types"
	"github.com/stretchr/testify/require"
)

func TestMsgServer(t *testing.T) {

	suite := &TestSuite{}
	// Run standard tests with the helper
	t.Run("TestParams", func(t *testing.T) {
		suite.SetupTest(t)
		suite.TestParams(t)
	})

	t.Run("TestRegisterFido2Credential", func(t *testing.T) {
		suite.SetupTest(t)
		suite.TestRegisterFido2Credential(t)
	})

	t.Run("TestRegisterFido2MultipleCredential", func(t *testing.T) {
		suite.SetupTest(t)
		suite.TestRegisterFido2MultipleCredential(t)
	})

	t.Run("TestRegisterFido2DuplicateCredential", func(t *testing.T) {
		suite.SetupTest(t)
		suite.TestRegisterFido2DuplicateCredential(t)
	})

	t.Run("TestRegisterCosmosCredential", func(t *testing.T) {
		suite.SetupTest(t)
		suite.TestRegisterCosmosCredential(t)
	})

	// Special case for test with custom parameters
	t.Run("TestRegisterFido2MaxCredentialCheck", func(t *testing.T) {
		suite = &TestSuite{maxCredentialAllowed: 4}
		suite.SetupTest(t)
		suite.TestRegisterFido2MaxCredentialCheck(t)
	})

	t.Run("TestRegisterFido2CredentialWithBadBase64", func(t *testing.T) {
		suite.SetupTest(t)
		suite.TestRegisterFido2CredentialWithBadBase64(t)
	})

	t.Run("TestRegisterCosmosCredentialDuplicate", func(t *testing.T) {
		suite.SetupTest(t)
		suite.TestRegisterCosmosCredentialDuplicate(t)
	})
}

func (s *TestSuite) TestParams(t *testing.T) {
	required := require.New(t)

	testCases := []struct {
		name    string
		request *types.MsgUpdateParams
		err     bool
	}{
		{
			name: "fail; invalid authority",
			request: &types.MsgUpdateParams{
				Authority: s.addrs[0].String(),
				Params:    types.DefaultParams(),
			},
			err: true,
		},
		{
			name: "success",
			request: &types.MsgUpdateParams{
				Authority: s.govModAddr,
				Params:    types.DefaultParams(),
			},
			err: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			_, err := s.msgServer.UpdateParams(s.ctx, tc.request)

			if tc.err {
				required.Error(err)
			} else {
				required.NoError(err)

				r, err := s.queryServer.Params(s.ctx, &types.QueryParamsRequest{})
				required.NoError(err)

				required.EqualValues(&tc.request.Params, r.Params)
			}

		})
	}
}

// TestRegisterFido2Credential tests the MsgRegisterFido2Credential function,
// this will only test the init function
func (s *TestSuite) TestRegisterFido2Credential(t *testing.T) {
	assertions := require.New(t)

	// Step 1: Create a new account
	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()
	addr := sdk.AccAddress(pubKey.Address())

	baseAcc := s.keeper.AccountKeeper.NewAccountWithAddress(s.ctx, addr)
	err := baseAcc.SetPubKey(pubKey)
	if err != nil {
		panic(err)
	}
	s.keeper.AccountKeeper.SetAccount(s.ctx, baseAcc)

	msg := &types.MsgRegisterFido2Credential{
		Sender:             addr.String(),
		EncodedAttestation: base64.RawURLEncoding.EncodeToString([]byte(utils.TestCredentialRequestResponses["success"])),
		UserIdentifier:     "example_username",
	}

	resp, err := s.msgServer.RegisterFido2Credential(s.ctx, msg)
	assertions.NoError(err)
	assertions.NotNil(resp, "response should not be nil")
	assertions.Equal(resp.CredentialNumber, uint64(0), "credential number should match")

	// Check events
	events := s.ctx.EventManager().Events()
	assertions.NotEmpty(events, "events should be emitted")
	expectedEvent := types.NewEventSmartAccountInit(addr.String(), 1)
	result := s.containsMessage(s.ctx.EventManager().ABCIEvents(), expectedEvent)
	assertions.True(result, "Expected typed event was not found in response.\n    Expected: %+v\n", expectedEvent)

	// Check for EventFido2CredentialAdd
	expectedCredEvent := &types.EventFido2CredentialAdd{
		Address:          addr.String(),
		CredentialNumber: 0,
		CredentialId:     resp.ProvenanceAccount.Credentials[0].GetFido2Authenticator().Id,
	}
	resultCred := s.containsMessage(s.ctx.EventManager().ABCIEvents(), expectedCredEvent)
	assertions.True(resultCred, "Expected EventFido2CredentialAdd was not found in response.\n    Expected: %+v\n", expectedCredEvent)

	// Fetch the smart account using the query server
	queryResp, err := s.queryServer.SmartAccount(s.ctx, &types.AccountQueryRequest{Address: addr.String()})
	assertions.NoError(err, "should be able to fetch the smart account")
	account := queryResp.Provenanceaccount
	assertions.Equal(addr.String(), account.Address, "account address should match")
	assertions.Len(account.Credentials, 1, "should have one credential")
	assertions.Equal(uint64(0), account.Credentials[0].CredentialNumber, "credential number should match")

}

// TestRegisterFido2MultipleCredential tests the registration of multiple FIDO2(webauthn) credentials
func (s *TestSuite) TestRegisterFido2MultipleCredential(t *testing.T) {
	assertions := require.New(t)

	// Step 1: Create a new account
	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()
	addr := sdk.AccAddress(pubKey.Address())

	baseAcc := s.keeper.AccountKeeper.NewAccountWithAddress(s.ctx, addr)
	err := baseAcc.SetPubKey(pubKey)
	if err != nil {
		panic(err)
	}
	s.keeper.AccountKeeper.SetAccount(s.ctx, baseAcc)

	// Create the MsgRegisterFido2Credential message
	msg := &types.MsgRegisterFido2Credential{
		Sender:             addr.String(),
		EncodedAttestation: "eyJpZCI6ImJiY1h1S3MxTWFYb3ZaYkxIWEljX1EiLCJyYXdJZCI6ImJiY1h1S3MxTWFYb3ZaYkxIWEljX1EiLCJyZXNwb25zZSI6eyJhdHRlc3RhdGlvbk9iamVjdCI6Im8yTm1iWFJrYm05dVpXZGhkSFJUZEcxMG9HaGhkWFJvUkdGMFlWaVVTWllONVlnT2pHaDBOQmNQWkhaZ1c0X2tycm1paGpMSG1Wenp1b01kbDJOZEFBQUFBT3FialdaTkFSMGhQT1MydEl5MWRkUUFFRzIzRjdpck5UR2w2TDJXeXgxeUhQMmxBUUlESmlBQklWZ2doNUpKTTZQNVpPTm82OFFNbnUybVQzQnBnYUtOUlJERGZkRVpDOEQwclo0aVdDQkR1M2tZUGM3a0o2QnVLTmdvQXMzZjVLdkVWZ0pTZG1LTDJpU1k0cy1pWHciLCJjbGllbnREYXRhSlNPTiI6ImV5SjBlWEJsSWpvaWQyVmlZWFYwYUc0dVkzSmxZWFJsSWl3aVkyaGhiR3hsYm1kbElqb2lUMmhuYWt4cGVEUk9ZV2xGWTJaelIxQnNTMG80Y0RSVFZuQjVhV05WTlZwWlEyeHZjemR1ZG1keVl5SXNJbTl5YVdkcGJpSTZJbWgwZEhBNkx5OXNiMk5oYkdodmMzUTZNVGd3T0RBaWZRIn0sInR5cGUiOiJwdWJsaWMta2V5IiwiYXV0aGVudGljYXRvckF0dGFjaG1lbnQiOiJwbGF0Zm9ybSJ9",
		UserIdentifier:     "example_username_1",
	}

	resp, err := s.msgServer.RegisterFido2Credential(s.ctx, msg)
	assertions.NoError(err)
	assertions.NotNil(resp, "response should not be nil")
	assertions.Equal(uint64(0), resp.CredentialNumber, "credential number should match")

	// Check events -- 2 events should be emitted one for init and one for credential add
	events := s.ctx.EventManager().Events()
	assertions.NotEmpty(events, "events should be emitted")
	expectedEvent := types.NewEventSmartAccountInit(addr.String(), 1)
	result := s.containsMessage(s.ctx.EventManager().ABCIEvents(), expectedEvent)
	assertions.True(result, "Expected typed event was not found in response.\n    Expected: %+v\n", expectedEvent)

	// Check for EventFido2CredentialAdd
	expectedCredEvent := &types.EventFido2CredentialAdd{
		Address:          addr.String(),
		CredentialNumber: 0,
		CredentialId:     resp.ProvenanceAccount.Credentials[0].GetFido2Authenticator().Id,
	}
	resultCred := s.containsMessage(s.ctx.EventManager().ABCIEvents(), expectedCredEvent)
	assertions.True(resultCred, "Expected EventFido2CredentialAdd was not found in response.\n    Expected: %+v\n", expectedCredEvent)

	// Create the MsgRegisterFido2Credential message
	msg = &types.MsgRegisterFido2Credential{
		Sender:             addr.String(),
		EncodedAttestation: "eyJpZCI6InAtOTNIaVpmRVpQX0ZYNURNY3dvaGciLCJyYXdJZCI6InAtOTNIaVpmRVpQX0ZYNURNY3dvaGciLCJyZXNwb25zZSI6eyJhdHRlc3RhdGlvbk9iamVjdCI6Im8yTm1iWFJrYm05dVpXZGhkSFJUZEcxMG9HaGhkWFJvUkdGMFlWaVVTWllONVlnT2pHaDBOQmNQWkhaZ1c0X2tycm1paGpMSG1Wenp1b01kbDJOZEFBQUFBT3FialdaTkFSMGhQT1MydEl5MWRkUUFFS2Z2ZHg0bVh4R1RfeFYtUXpITUtJYWxBUUlESmlBQklWZ2dWM2JBbjVaejJ1Z0JuRm9QVXIyR0RIaXZTaE50MjYxWmROaUpuaDVYV00waVdDQmlFblc0MGtzYThreFp6RmkxcV9RN2x0MmU5ZnhnOThXZDN0S0hDZ19tX1EiLCJjbGllbnREYXRhSlNPTiI6ImV5SjBlWEJsSWpvaWQyVmlZWFYwYUc0dVkzSmxZWFJsSWl3aVkyaGhiR3hsYm1kbElqb2lSbTFWVnpaRlVXUnRTMEpOWlMxNlJXRnRaR1ZhYzFOU1lub3RSekZxWWpOellYbDBXVWgxYzJzMlFTSXNJbTl5YVdkcGJpSTZJbWgwZEhBNkx5OXNiMk5oYkdodmMzUTZNVGd3T0RBaWZRIn0sInR5cGUiOiJwdWJsaWMta2V5IiwiYXV0aGVudGljYXRvckF0dGFjaG1lbnQiOiJwbGF0Zm9ybSJ9", // foo6
		UserIdentifier:     "example_username_2",
	}

	resp, err = s.msgServer.RegisterFido2Credential(s.ctx, msg)
	assertions.NoError(err)
	assertions.NotNil(resp, "response should not be nil")
	assertions.Equal(uint64(1), resp.CredentialNumber, "credential number should match")

	// Check for EventFido2CredentialAdd is an event emitted
	expectedCredEvent = &types.EventFido2CredentialAdd{
		Address:          addr.String(),
		CredentialNumber: 0,
		CredentialId:     resp.ProvenanceAccount.Credentials[0].GetFido2Authenticator().Id,
	}
	resultCred = s.containsMessage(s.ctx.EventManager().ABCIEvents(), expectedCredEvent)
	assertions.True(resultCred, "Expected EventFido2CredentialAdd was not found in response.\n    Expected: %+v\n", expectedCredEvent)

	// Fetch the smart account using the query server(now it should have two credentials)
	queryResp, err := s.queryServer.SmartAccount(s.ctx, &types.AccountQueryRequest{Address: addr.String()})
	assertions.NoError(err, "should be able to fetch the smart account")
	account := queryResp.Provenanceaccount
	assertions.Equal(addr.String(), account.Address, "account address should match")
	assertions.Len(account.Credentials, 2, "should have two credential")
	assertions.Equal(uint64(1), account.Credentials[1].CredentialNumber, "credential number should match")
	assertions.Equal(types.CredentialType_CREDENTIAL_TYPE_WEBAUTHN_UV, account.Credentials[0].BaseCredential.Variant, "credential variant should be UV in this case")
	assertions.Equal(types.CredentialType_CREDENTIAL_TYPE_WEBAUTHN_UV, account.Credentials[1].BaseCredential.Variant, "credential variant should be UV in this case")

}

// TestRegisterFido2DuplicateCredential tests the registration of multiple Fido2(Webauthn) credentials
func (s *TestSuite) TestRegisterFido2DuplicateCredential(t *testing.T) {
	assertions := require.New(t)

	// Step 1: Create a new account
	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()
	addr := sdk.AccAddress(pubKey.Address())

	baseAcc := s.keeper.AccountKeeper.NewAccountWithAddress(s.ctx, addr)
	err := baseAcc.SetPubKey(pubKey)
	if err != nil {
		panic(err)
	}
	s.keeper.AccountKeeper.SetAccount(s.ctx, baseAcc)

	// Create the MsgRegisterFido2Credential message
	msg := &types.MsgRegisterFido2Credential{
		Sender:             addr.String(),
		EncodedAttestation: "eyJpZCI6ImJiY1h1S3MxTWFYb3ZaYkxIWEljX1EiLCJyYXdJZCI6ImJiY1h1S3MxTWFYb3ZaYkxIWEljX1EiLCJyZXNwb25zZSI6eyJhdHRlc3RhdGlvbk9iamVjdCI6Im8yTm1iWFJrYm05dVpXZGhkSFJUZEcxMG9HaGhkWFJvUkdGMFlWaVVTWllONVlnT2pHaDBOQmNQWkhaZ1c0X2tycm1paGpMSG1Wenp1b01kbDJOZEFBQUFBT3FialdaTkFSMGhQT1MydEl5MWRkUUFFRzIzRjdpck5UR2w2TDJXeXgxeUhQMmxBUUlESmlBQklWZ2doNUpKTTZQNVpPTm82OFFNbnUybVQzQnBnYUtOUlJERGZkRVpDOEQwclo0aVdDQkR1M2tZUGM3a0o2QnVLTmdvQXMzZjVLdkVWZ0pTZG1LTDJpU1k0cy1pWHciLCJjbGllbnREYXRhSlNPTiI6ImV5SjBlWEJsSWpvaWQyVmlZWFYwYUc0dVkzSmxZWFJsSWl3aVkyaGhiR3hsYm1kbElqb2lUMmhuYWt4cGVEUk9ZV2xGWTJaelIxQnNTMG80Y0RSVFZuQjVhV05WTlZwWlEyeHZjemR1ZG1keVl5SXNJbTl5YVdkcGJpSTZJbWgwZEhBNkx5OXNiMk5oYkdodmMzUTZNVGd3T0RBaWZRIn0sInR5cGUiOiJwdWJsaWMta2V5IiwiYXV0aGVudGljYXRvckF0dGFjaG1lbnQiOiJwbGF0Zm9ybSJ9",
		UserIdentifier:     "example_username_1",
	}

	resp, err := s.msgServer.RegisterFido2Credential(s.ctx, msg)
	assertions.NoError(err)
	assertions.NotNil(resp, "response should not be nil")
	assertions.Equal(resp.CredentialNumber, uint64(0), "credential number should match")

	// Create the MsgRegisterFido2Credential message
	msg = &types.MsgRegisterFido2Credential{
		Sender:             addr.String(),
		EncodedAttestation: "eyJpZCI6ImJiY1h1S3MxTWFYb3ZaYkxIWEljX1EiLCJyYXdJZCI6ImJiY1h1S3MxTWFYb3ZaYkxIWEljX1EiLCJyZXNwb25zZSI6eyJhdHRlc3RhdGlvbk9iamVjdCI6Im8yTm1iWFJrYm05dVpXZGhkSFJUZEcxMG9HaGhkWFJvUkdGMFlWaVVTWllONVlnT2pHaDBOQmNQWkhaZ1c0X2tycm1paGpMSG1Wenp1b01kbDJOZEFBQUFBT3FialdaTkFSMGhQT1MydEl5MWRkUUFFRzIzRjdpck5UR2w2TDJXeXgxeUhQMmxBUUlESmlBQklWZ2doNUpKTTZQNVpPTm82OFFNbnUybVQzQnBnYUtOUlJERGZkRVpDOEQwclo0aVdDQkR1M2tZUGM3a0o2QnVLTmdvQXMzZjVLdkVWZ0pTZG1LTDJpU1k0cy1pWHciLCJjbGllbnREYXRhSlNPTiI6ImV5SjBlWEJsSWpvaWQyVmlZWFYwYUc0dVkzSmxZWFJsSWl3aVkyaGhiR3hsYm1kbElqb2lUMmhuYWt4cGVEUk9ZV2xGWTJaelIxQnNTMG80Y0RSVFZuQjVhV05WTlZwWlEyeHZjemR1ZG1keVl5SXNJbTl5YVdkcGJpSTZJbWgwZEhBNkx5OXNiMk5oYkdodmMzUTZNVGd3T0RBaWZRIn0sInR5cGUiOiJwdWJsaWMta2V5IiwiYXV0aGVudGljYXRvckF0dGFjaG1lbnQiOiJwbGF0Zm9ybSJ9", // foo6
		UserIdentifier:     "example_username_2",
	}

	resp, err = s.msgServer.RegisterFido2Credential(s.ctx, msg)
	assertions.Error(err)
	assertions.Nil(resp, "response should be nil")

	// Fetch the smart account using the query server(now it should have two credentials)
	queryResp, err := s.queryServer.SmartAccount(s.ctx, &types.AccountQueryRequest{Address: addr.String()})
	assertions.NoError(err, "should be able to fetch the smart account")
	account := queryResp.Provenanceaccount
	assertions.Equal(addr.String(), account.Address, "account address should match")
	assertions.Len(account.Credentials, 1, "should have two credential")
	assertions.Equal(uint64(0), account.Credentials[0].CredentialNumber, "credential number should match")
	assertions.Equal(types.CredentialType_CREDENTIAL_TYPE_WEBAUTHN_UV, account.Credentials[0].BaseCredential.Variant, "credential variant should be UV in this case")

	// try to register a duplicate credential with the same credential id and expect an error

	resp, err = s.msgServer.RegisterFido2Credential(s.ctx, msg)
	assertions.Error(err)
	// check the error message
	assertions.Contains(err.Error(), "credential already exists")
}

// TestRegisterFido2MaxCredentialCheck tests the registration of multiple FIDO2(WebAuthn) credentials
func (s *TestSuite) TestRegisterFido2MaxCredentialCheck(t *testing.T) {
	assertions := require.New(t)

	// Step 1: Create a new account
	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()
	addr := sdk.AccAddress(pubKey.Address())

	baseAcc := s.keeper.AccountKeeper.NewAccountWithAddress(s.ctx, addr)
	err := baseAcc.SetPubKey(pubKey)
	if err != nil {
		panic(err)
	}
	s.keeper.AccountKeeper.SetAccount(s.ctx, baseAcc)

	// Create the MsgRegisterFido2Credential message
	msg := &types.MsgRegisterFido2Credential{
		Sender:             addr.String(),
		EncodedAttestation: "eyJpZCI6ImJiY1h1S3MxTWFYb3ZaYkxIWEljX1EiLCJyYXdJZCI6ImJiY1h1S3MxTWFYb3ZaYkxIWEljX1EiLCJyZXNwb25zZSI6eyJhdHRlc3RhdGlvbk9iamVjdCI6Im8yTm1iWFJrYm05dVpXZGhkSFJUZEcxMG9HaGhkWFJvUkdGMFlWaVVTWllONVlnT2pHaDBOQmNQWkhaZ1c0X2tycm1paGpMSG1Wenp1b01kbDJOZEFBQUFBT3FialdaTkFSMGhQT1MydEl5MWRkUUFFRzIzRjdpck5UR2w2TDJXeXgxeUhQMmxBUUlESmlBQklWZ2doNUpKTTZQNVpPTm82OFFNbnUybVQzQnBnYUtOUlJERGZkRVpDOEQwclo0aVdDQkR1M2tZUGM3a0o2QnVLTmdvQXMzZjVLdkVWZ0pTZG1LTDJpU1k0cy1pWHciLCJjbGllbnREYXRhSlNPTiI6ImV5SjBlWEJsSWpvaWQyVmlZWFYwYUc0dVkzSmxZWFJsSWl3aVkyaGhiR3hsYm1kbElqb2lUMmhuYWt4cGVEUk9ZV2xGWTJaelIxQnNTMG80Y0RSVFZuQjVhV05WTlZwWlEyeHZjemR1ZG1keVl5SXNJbTl5YVdkcGJpSTZJbWgwZEhBNkx5OXNiMk5oYkdodmMzUTZNVGd3T0RBaWZRIn0sInR5cGUiOiJwdWJsaWMta2V5IiwiYXV0aGVudGljYXRvckF0dGFjaG1lbnQiOiJwbGF0Zm9ybSJ9",
		UserIdentifier:     "example_username_1",
	}

	resp, err := s.msgServer.RegisterFido2Credential(s.ctx, msg)
	assertions.NoError(err)
	assertions.NotNil(resp, "response should not be nil")
	assertions.Equal(uint64(0), resp.CredentialNumber, "credential number should match")

	// Create the MsgRegisterFido2Credential message
	msg = &types.MsgRegisterFido2Credential{
		Sender:             addr.String(),
		EncodedAttestation: "eyJpZCI6InAtOTNIaVpmRVpQX0ZYNURNY3dvaGciLCJyYXdJZCI6InAtOTNIaVpmRVpQX0ZYNURNY3dvaGciLCJyZXNwb25zZSI6eyJhdHRlc3RhdGlvbk9iamVjdCI6Im8yTm1iWFJrYm05dVpXZGhkSFJUZEcxMG9HaGhkWFJvUkdGMFlWaVVTWllONVlnT2pHaDBOQmNQWkhaZ1c0X2tycm1paGpMSG1Wenp1b01kbDJOZEFBQUFBT3FialdaTkFSMGhQT1MydEl5MWRkUUFFS2Z2ZHg0bVh4R1RfeFYtUXpITUtJYWxBUUlESmlBQklWZ2dWM2JBbjVaejJ1Z0JuRm9QVXIyR0RIaXZTaE50MjYxWmROaUpuaDVYV00waVdDQmlFblc0MGtzYThreFp6RmkxcV9RN2x0MmU5ZnhnOThXZDN0S0hDZ19tX1EiLCJjbGllbnREYXRhSlNPTiI6ImV5SjBlWEJsSWpvaWQyVmlZWFYwYUc0dVkzSmxZWFJsSWl3aVkyaGhiR3hsYm1kbElqb2lSbTFWVnpaRlVXUnRTMEpOWlMxNlJXRnRaR1ZhYzFOU1lub3RSekZxWWpOellYbDBXVWgxYzJzMlFTSXNJbTl5YVdkcGJpSTZJbWgwZEhBNkx5OXNiMk5oYkdodmMzUTZNVGd3T0RBaWZRIn0sInR5cGUiOiJwdWJsaWMta2V5IiwiYXV0aGVudGljYXRvckF0dGFjaG1lbnQiOiJwbGF0Zm9ybSJ9", // foo6
		UserIdentifier:     "example_username_2",
	}

	resp, err = s.msgServer.RegisterFido2Credential(s.ctx, msg)
	assertions.NoError(err)
	assertions.NotNil(resp, "response should not be nil")
	assertions.Equal(uint64(1), resp.CredentialNumber, "credential number should match")

	// Fetch the smart account using the query server(now it should have two credentials)
	queryResp, err := s.queryServer.SmartAccount(s.ctx, &types.AccountQueryRequest{Address: addr.String()})
	assertions.NoError(err, "should be able to fetch the smart account")
	account := queryResp.Provenanceaccount
	assertions.Equal(addr.String(), account.Address, "account address should match")
	assertions.Len(account.Credentials, 2, "should have two credential")
	assertions.Equal(uint64(1), account.Credentials[1].CredentialNumber, "credential number should match")
	assertions.Equal(types.CredentialType_CREDENTIAL_TYPE_WEBAUTHN_UV, account.Credentials[0].BaseCredential.Variant, "credential variant should be UV in this case")
	assertions.Equal(types.CredentialType_CREDENTIAL_TYPE_WEBAUTHN_UV, account.Credentials[1].BaseCredential.Variant, "credential variant should be UV in this case")

	// Create the MsgRegisterFido2Credential message
	msg = &types.MsgRegisterFido2Credential{
		Sender:             addr.String(),
		EncodedAttestation: "eyJpZCI6Ik94YVN5cVdoZFlEZmVycXhQSC1YV28tdWtYX0tpalVZdi0yRDRGbFVQZG8iLCJyYXdJZCI6Ik94YVN5cVdoZFlEZmVycXhQSC1YV28tdWtYX0tpalVZdi0yRDRGbFVQZG8iLCJyZXNwb25zZSI6eyJhdHRlc3RhdGlvbk9iamVjdCI6Im8yTm1iWFJrYm05dVpXZGhkSFJUZEcxMG9HaGhkWFJvUkdGMFlWaWtTWllONVlnT2pHaDBOQmNQWkhaZ1c0X2tycm1paGpMSG1Wenp1b01kbDJORkFBQUFBSzNPQUFJMXZNWUtaSXNMSmZId1ZRTUFJRHNXa3NxbG9YV0EzM3E2c1R4X2wxcVBycEZfeW9vMUdMX3RnLUJaVkQzYXBRRUNBeVlnQVNGWUlBRlV1OTZqeUxub2hoZGlTVWt1QUpXNTl3WnBBbnVZVVBfNExNaHkxTGFwSWxnZ3FhNXg1OGg1RDc3TjY5UlVudkFUTjloNGVwSnpUeEZCX29YQVhfUjFPSmciLCJjbGllbnREYXRhSlNPTiI6ImV5SjBlWEJsSWpvaWQyVmlZWFYwYUc0dVkzSmxZWFJsSWl3aVkyaGhiR3hsYm1kbElqb2lkMHBzZERabmFISjZMWFpZZVU1VlZsWlJWRVJYVDNCa2NFUmFNbXRUV21KdWNqQmlNVkJWT1VKTFRTSXNJbTl5YVdkcGJpSTZJbWgwZEhBNkx5OXNiMk5oYkdodmMzUTZNVGd3T0RBaWZRIn0sInR5cGUiOiJwdWJsaWMta2V5IiwiYXV0aGVudGljYXRvckF0dGFjaG1lbnQiOiJwbGF0Zm9ybSJ9", // foo6
		UserIdentifier:     "example_username_3",
	}

	resp, err = s.msgServer.RegisterFido2Credential(s.ctx, msg)
	assertions.NoError(err)
	assertions.NotNil(resp, "response should not be nil")
	assertions.Equal(uint64(2), resp.CredentialNumber, "credential number should match")

	// Create the MsgRegisterFido2Credential message
	msg = &types.MsgRegisterFido2Credential{
		Sender:             addr.String(),
		EncodedAttestation: "eyJpZCI6ImNpbWRnYkFydkhKWDdGcE90THg3bENwWHVSLUFxRXppWjdVTXphdVVKT2V0Nl9FM0lfc2RwN0NPTUwyVm9Ud1ciLCJyYXdJZCI6ImNpbWRnYkFydkhKWDdGcE90THg3bENwWHVSLUFxRXppWjdVTXphdVVKT2V0Nl9FM0lfc2RwN0NPTUwyVm9Ud1ciLCJyZXNwb25zZSI6eyJhdHRlc3RhdGlvbk9iamVjdCI6Im8yTm1iWFJrYm05dVpXZGhkSFJUZEcxMG9HaGhkWFJvUkdGMFlWakNTWllONVlnT2pHaDBOQmNQWkhaZ1c0X2tycm1paGpMSG1Wenp1b01kbDJQRkFBQUFBUUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFNSElwbllHd0s3eHlWLXhhVHJTOGU1UXFWN2tmZ0toTTRtZTFETTJybENUbnJldnhOeVA3SGFld2pqQzlsYUU4RnFVQkFnTW1JQUVoV0NCeUtaMkJzQ3U4Y2xmc1drNjBBRVJGNWJsWWpETVFVQzdleFVoNFVXakQyeUpZSUo0Z1JTQnpWVndSd2lDYzNmMmN3dHVKMnQ5eVV2bHRXNUxwSUllOFotMVNvV3RqY21Wa1VISnZkR1ZqZEFNIiwiY2xpZW50RGF0YUpTT04iOiJleUowZVhCbElqb2lkMlZpWVhWMGFHNHVZM0psWVhSbElpd2lZMmhoYkd4bGJtZGxJam9pVTB0RExWTnhORnBFYjNSTFZuVmtXRmhGYjNaMlNqWkZUa2QxYTBadWVXNVNaMWRsVUhwTmNFOU5UU0lzSW05eWFXZHBiaUk2SW1oMGRIQTZMeTlzYjJOaGJHaHZjM1E2TVRnd09EQWlmUSJ9LCJ0eXBlIjoicHVibGljLWtleSIsImF1dGhlbnRpY2F0b3JBdHRhY2htZW50IjoiY3Jvc3MtcGxhdGZvcm0ifQ", // foo6
		UserIdentifier:     "example_username_4",
	}

	resp, err = s.msgServer.RegisterFido2Credential(s.ctx, msg)
	assertions.NoError(err)
	assertions.NotNil(resp, "response should not be nil")
	assertions.Equal(uint64(3), resp.CredentialNumber, "credential number should match")

	// Create the same MsgRegisterFido2Credential message
	msg = &types.MsgRegisterFido2Credential{
		Sender:             addr.String(),
		EncodedAttestation: "eyJpZCI6IkZXc2Z0a29od0pjdXZOeTJDenEtdS04VWJIUmtUNWxmek50ZWxOZ0pGbXZ4V09ZVjVoODVaTFFMaXpXRVRLblQiLCJyYXdJZCI6IkZXc2Z0a29od0pjdXZOeTJDenEtdS04VWJIUmtUNWxmek50ZWxOZ0pGbXZ4V09ZVjVoODVaTFFMaXpXRVRLblQiLCJyZXNwb25zZSI6eyJhdHRlc3RhdGlvbk9iamVjdCI6Im8yTm1iWFJrYm05dVpXZGhkSFJUZEcxMG9HaGhkWFJvUkdGMFlWakNTWllONVlnT2pHaDBOQmNQWkhaZ1c0X2tycm1paGpMSG1Wenp1b01kbDJQRkFBQUFBUUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFNQlZySDdaS0ljQ1hMcnpjdGdzNnZydnZGR3gwWkUtWlg4emJYcFRZQ1JacjhWam1GZVlmT1dTMEM0czFoRXlwMDZVQkFnTW1JQUVoV0NBVmF4LTJTaUhBbHk2ODNMWUxrTS1nS253MlBtSlFJd2tPWTRpNzlERDQyaUpZSURGQUpIWXE5d1lmU093VExycTFYV3IwX3NuQkdTZGFDOWdYRnNXY3dTRWJvV3RqY21Wa1VISnZkR1ZqZEFNIiwiY2xpZW50RGF0YUpTT04iOiJleUowZVhCbElqb2lkMlZpWVhWMGFHNHVZM0psWVhSbElpd2lZMmhoYkd4bGJtZGxJam9pYzBoalZqYzRNR0pXT1hFMWMyeFVWbGxoVFVWdldWQlhTVkJGYTA5MlJHMVdVa1pJVm1SdlNrYzFTU0lzSW05eWFXZHBiaUk2SW1oMGRIQTZMeTlzYjJOaGJHaHZjM1E2TVRnd09EQWlmUSJ9LCJ0eXBlIjoicHVibGljLWtleSIsImF1dGhlbnRpY2F0b3JBdHRhY2htZW50IjoiY3Jvc3MtcGxhdGZvcm0ifQ", // foo6
		UserIdentifier:     "example_username_5",
	}

	resp, err = s.msgServer.RegisterFido2Credential(s.ctx, msg)
	assertions.Error(err)
	assertions.Contains(err.Error(), "maximum number of credentials")
}

// TestRegisterCosmosCredential tests the RegisterCosmosCredential function
func (s *TestSuite) TestRegisterCosmosCredential(t *testing.T) {
	assertions := require.New(t)

	// Generate keys for testing
	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()
	addr := sdk.AccAddress(pubKey.Address())

	// Create a base account for the sender
	s.createBaseAccount(pubKey)

	// Test 1: Register credential for a new smart account
	pubKeyAny, err := codectypes.NewAnyWithValue(pubKey)
	assertions.NoError(err)

	msg := &types.MsgRegisterCosmosCredential{
		Sender: addr.String(),
		Pubkey: pubKeyAny,
	}

	resp, err := s.msgServer.RegisterCosmosCredential(s.ctx, msg)
	assertions.NoError(err)
	assertions.NotNil(resp)
	assertions.Equal(uint64(0), resp.CredentialNumber)

	// Check events --2 events should be emitted one for init and one for credential add
	events := s.ctx.EventManager().Events()
	assertions.NotEmpty(events, "events should be emitted")

	// check for EventSmartAccountInit
	expectedEvent := types.NewEventSmartAccountInit(addr.String(), 1)
	result := s.containsMessage(s.ctx.EventManager().ABCIEvents(), expectedEvent)
	assertions.True(result, "Expected typed event was not found in response.\n    Expected: %+v\n", expectedEvent)
	// Check for EventFido2CredentialAdd
	expectedCredEvent := &types.EventCosmosCredentialAdd{
		Address:          addr.String(),
		CredentialNumber: 0,
	}
	resultCred := s.containsMessage(s.ctx.EventManager().ABCIEvents(), expectedCredEvent)
	assertions.True(resultCred, "Expected EventCosmosCredentialAdd was not found in response.\n    Expected: %+v\n", expectedCredEvent)

	// Verify the account was created with the credential
	smartAccount, err := s.keeper.LookupAccountByAddress(s.ctx, addr)
	assertions.NoError(err)
	assertions.Equal(1, len(smartAccount.Credentials))
	assertions.Equal(types.CredentialType_CREDENTIAL_TYPE_K256, smartAccount.Credentials[0].GetVariant())

	// Test 2: Register a second credential to the existing account
	privKey2 := secp256k1.GenPrivKey()
	pubKey2 := privKey2.PubKey()
	pubKeyAny2, err := codectypes.NewAnyWithValue(pubKey2)
	assertions.NoError(err)

	msg2 := &types.MsgRegisterCosmosCredential{
		Sender: addr.String(),
		Pubkey: pubKeyAny2,
	}

	resp2, err := s.msgServer.RegisterCosmosCredential(s.ctx, msg2)
	assertions.NoError(err)
	assertions.NotNil(resp2)
	assertions.Equal(uint64(1), resp2.CredentialNumber)

	// Check for EventFido2CredentialAdd
	expectedCredEvent2 := &types.EventCosmosCredentialAdd{
		Address:          addr.String(),
		CredentialNumber: 1,
	}
	resultCred2 := s.containsMessage(s.ctx.EventManager().ABCIEvents(), expectedCredEvent2)
	assertions.True(resultCred2, "Expected EventCosmosCredentialAdd was not found in response.\n    Expected: %+v\n", expectedCredEvent2)

	// Verify the second credential was added
	smartAccount, err = s.keeper.LookupAccountByAddress(s.ctx, addr)
	assertions.NoError(err)
	assertions.Equal(2, len(smartAccount.Credentials))

	// Test 3: Invalid address
	msg3 := &types.MsgRegisterCosmosCredential{
		Sender: "invalid-address",
		Pubkey: pubKeyAny,
	}

	_, err = s.msgServer.RegisterCosmosCredential(s.ctx, msg3)
	assertions.Error(err)
	assertions.Contains(err.Error(), "invalid sender address")
}

func (s *TestSuite) TestRegisterFido2CredentialWithBadBase64(t *testing.T) {
	assertions := require.New(t)

	// Create a test account
	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()
	addr := sdk.AccAddress(pubKey.Address())

	baseAcc := s.keeper.AccountKeeper.NewAccountWithAddress(s.ctx, addr)
	err := baseAcc.SetPubKey(pubKey)
	assertions.NoError(err)
	s.keeper.AccountKeeper.SetAccount(s.ctx, baseAcc)

	// Create message with malformed base64 attestation
	// Using characters that aren't valid in base64url encoding
	msg := &types.MsgRegisterFido2Credential{
		Sender:             addr.String(),
		EncodedAttestation: "This is not valid base64!@#$%^&*()_+",
		UserIdentifier:     "example_username",
	}

	// Call the message handler
	resp, err := s.msgServer.RegisterFido2Credential(s.ctx, msg)

	// Verify error handling
	assertions.Error(err, "should return an error for invalid base64")
	assertions.Nil(resp, "response should be nil when error occurs")
	assertions.Contains(err.Error(), "failed to decode base64", "error should mention base64 decoding")

	// Verify the error is wrapped with the expected error type
	assertions.True(errors.Is(err, types.ErrParseCredential),
		"error should be of type ErrParseCredential")

	// Verify no account was created
	_, err = s.queryServer.SmartAccount(s.ctx, &types.AccountQueryRequest{Address: addr.String()})
	assertions.Error(err, "no smart account should be created")
	assertions.Contains(err.Error(), "smart account does not exist")
}

func (s *TestSuite) TestRegisterCosmosCredentialDuplicate(t *testing.T) {
	assertions := assert.New(t)
	require := require.New(t)

	// Generate keys for testing
	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey()
	addr := sdk.AccAddress(pubKey.Address())

	pubKeyAny, err := codectypes.NewAnyWithValue(pubKey)
	require.NoError(err)

	// Create a base account for the sender
	s.createBaseAccount(pubKey)

	msg := types.NewMsgRegisterCosmosCredential(addr.String(), pubKeyAny)

	// First registration - should succeed
	resp, err := s.msgServer.RegisterCosmosCredential(s.ctx, msg)
	require.NoError(err)
	require.NotNil(resp)
	assertions.Equal(uint64(0), resp.CredentialNumber, "First credential should have number 0")

	// Verify account was created with the credential
	accountResp, err := s.queryServer.SmartAccount(s.ctx, &types.AccountQueryRequest{Address: addr.String()})
	require.NoError(err)
	require.NotNil(accountResp)

	// Check the credential was added correctly
	creds := accountResp.Provenanceaccount.Credentials
	require.Equal(1, len(creds))
	assertions.Equal(types.CredentialType_CREDENTIAL_TYPE_K256, creds[0].BaseCredential.Variant)
	assertions.Equal(pubKeyAny.TypeUrl, creds[0].BaseCredential.PublicKey.TypeUrl)
	assertions.Equal(pubKeyAny.Value, creds[0].BaseCredential.PublicKey.Value)

	// Try registering the same credential again - should fail with duplicate error
	duplicateMsg := types.NewMsgRegisterCosmosCredential(addr.String(), pubKeyAny)
	_, err = s.msgServer.RegisterCosmosCredential(s.ctx, duplicateMsg)
	assertions.Error(err)
	assertions.True(errors.Is(err, types.ErrDuplicateCredential),
		"Expected duplicate credential error but got: %v", err)

	// Generate a second key and verify we can add multiple different credentials
	privKey2 := secp256k1.GenPrivKey()
	pubKey2 := privKey2.PubKey()
	pubKeyAny2, err := codectypes.NewAnyWithValue(pubKey2)
	require.NoError(err)

	// Register second credential - should succeed
	msg2 := types.NewMsgRegisterCosmosCredential(addr.String(), pubKeyAny2)
	resp2, err := s.msgServer.RegisterCosmosCredential(s.ctx, msg2)
	require.NoError(err)
	assertions.Equal(uint64(1), resp2.CredentialNumber, "Second credential should have number 1")

	// Verify the account now has two credentials
	accountResp2, err := s.queryServer.SmartAccount(s.ctx, &types.AccountQueryRequest{Address: addr.String()})
	require.NoError(err)
	assertions.Equal(2, len(accountResp2.Provenanceaccount.Credentials))

	// Test invalid pubkey (nil)
	invalidMsg := types.NewMsgRegisterCosmosCredential(addr.String(), nil)
	_, err = s.msgServer.RegisterCosmosCredential(s.ctx, invalidMsg)
	assertions.Error(err)
}

func (s *TestSuite) containsMessage(events []abci.Event, msg proto.Message) bool {
	for _, event := range events {
		typeEvent, _ := sdk.ParseTypedEvent(event)
		if assert.ObjectsAreEqual(msg, typeEvent) {
			return true
		}
	}
	return false
}
