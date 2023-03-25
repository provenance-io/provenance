package rest_test

import (
	b64 "encoding/base64"
	"fmt"
	"strings"
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"google.golang.org/genproto/googleapis/rpc/status"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdktestutil "github.com/cosmos/cosmos-sdk/testutil"
	testnet "github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
	"github.com/cosmos/cosmos-sdk/types/query"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/provenance-io/provenance/internal/pioconfig"
	"github.com/provenance-io/provenance/testutil"
	"github.com/provenance-io/provenance/x/metadata/types"
)

type IntegrationGRPCTestSuite struct {
	suite.Suite

	cfg     testnet.Config
	testnet *testnet.Network

	accountAddr sdk.AccAddress
	accountKey  *secp256k1.PrivKey

	pubkey1   cryptotypes.PubKey
	user1     string
	user1Addr sdk.AccAddress

	pubkey2   cryptotypes.PubKey
	user2     string
	user2Addr sdk.AccAddress

	scope     types.Scope
	scopeUUID uuid.UUID
	scopeID   types.MetadataAddress

	specUUID uuid.UUID
	specID   types.MetadataAddress

	objectLocator types.ObjectStoreLocator
	ownerAddr     sdk.AccAddress
	encryptionKey sdk.AccAddress
	uri           string

	objectLocator1 types.ObjectStoreLocator
	ownerAddr1     sdk.AccAddress
	encryptionKey1 sdk.AccAddress
	uri1           string
}

func ownerPartyList(addresses ...string) []types.Party {
	retval := make([]types.Party, len(addresses))
	for i, addr := range addresses {
		retval[i] = types.Party{Address: addr, Role: types.PartyType_PARTY_TYPE_OWNER}
	}
	return retval
}

func (suite *IntegrationGRPCTestSuite) SetupSuite() {
	pioconfig.SetProvenanceConfig("atom", 0)
	suite.accountKey = secp256k1.GenPrivKeyFromSecret([]byte("acc2"))
	addr, err := sdk.AccAddressFromHexUnsafe(suite.accountKey.PubKey().Address().String())
	suite.Require().NoError(err)
	suite.accountAddr = addr
	suite.T().Log("setting up integration test suite")

	cfg := testutil.DefaultTestNetworkConfig()

	genesisState := cfg.GenesisState
	cfg.NumValidators = 1

	var authData authtypes.GenesisState
	suite.Require().NoError(cfg.Codec.UnmarshalJSON(genesisState[authtypes.ModuleName], &authData))
	genAccount, err := codectypes.NewAnyWithValue(&authtypes.BaseAccount{
		Address:       suite.accountAddr.String(),
		AccountNumber: 1,
		Sequence:      0,
	})
	suite.Require().NoError(err)
	authData.Accounts = append(authData.Accounts, genAccount)

	suite.pubkey1 = secp256k1.GenPrivKey().PubKey()
	suite.user1Addr = sdk.AccAddress(suite.pubkey1.Address())
	suite.user1 = suite.user1Addr.String()

	suite.pubkey2 = secp256k1.GenPrivKey().PubKey()
	suite.user2Addr = sdk.AccAddress(suite.pubkey2.Address())
	suite.user2 = suite.user2Addr.String()

	suite.scopeUUID = uuid.New()
	suite.scopeID = types.ScopeMetadataAddress(suite.scopeUUID)

	suite.specUUID = uuid.New()
	suite.specID = types.ScopeSpecMetadataAddress(suite.specUUID)

	suite.scope = *types.NewScope(suite.scopeID, suite.specID, ownerPartyList(suite.user1), []string{suite.user1}, suite.user1)
	// Configure Genesis data for metadata module

	// add os locator
	suite.ownerAddr = suite.accountAddr
	suite.uri = "http://foo.com"
	suite.encryptionKey = sdk.AccAddress{}
	suite.objectLocator = types.NewOSLocatorRecord(suite.ownerAddr, suite.encryptionKey, suite.uri)

	suite.ownerAddr1 = suite.user1Addr
	suite.uri1 = "http://bar.com"
	suite.encryptionKey1 = suite.ownerAddr
	suite.objectLocator1 = types.NewOSLocatorRecord(suite.ownerAddr1, suite.encryptionKey1, suite.uri1)

	var metadataData types.GenesisState
	metadataData.Params = types.DefaultParams()
	metadataData.OSLocatorParams = types.DefaultOSLocatorParams()
	metadataData.Scopes = append(metadataData.Scopes, suite.scope)
	metadataData.ObjectStoreLocators = append(metadataData.ObjectStoreLocators, suite.objectLocator, suite.objectLocator1)
	metadataDataBz, err := cfg.Codec.MarshalJSON(&metadataData)
	suite.Require().NoError(err)

	genesisState[types.ModuleName] = metadataDataBz

	cfg.GenesisState = genesisState

	suite.cfg = cfg

	suite.testnet, err = testnet.New(suite.T(), suite.T().TempDir(), cfg)
	suite.Require().NoError(err, "creating testnet")

	_, err = suite.testnet.WaitForHeight(1)
	suite.Require().NoError(err, "waiting for height 1")
}

func (suite *IntegrationGRPCTestSuite) TearDownSuite() {
	testutil.CleanUp(suite.testnet, suite.T())
}

func TestIntegrationGRPCTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationGRPCTestSuite))
}

func (suite *IntegrationGRPCTestSuite) TestGRPCQueries() {
	val := suite.testnet.Validators[0]
	baseURL := val.APIAddress

	testCases := []struct {
		name     string
		url      string
		headers  map[string]string
		expErr   bool
		respType proto.Message
		expected proto.Message
	}{
		{
			name: "Get metadata params",
			url : fmt.Sprintf("%s/provenance/metadata/v1/params", baseURL),
			headers : map[string]string{
				grpctypes.GRPCBlockHeightHeader: "1",
			},
			expErr: false,
			respType: &types.QueryParamsResponse{},
			expected: &types.QueryParamsResponse{
				Params:  types.DefaultParams(),
				Request: &types.QueryParamsRequest{},
			},
		},
		{
			name: "Get metadata scope by id",
			url : fmt.Sprintf("%s/provenance/metadata/v1/scope/%s", baseURL, suite.scopeUUID),
			headers :map[string]string{
				grpctypes.GRPCBlockHeightHeader: "1",
			},
			expErr: false,
			respType : &types.ScopeResponse{},
			expected: &types.ScopeResponse{
				Scope: &types.ScopeWrapper{
					Scope:           &suite.scope,
					ScopeIdInfo:     types.GetScopeIDInfo(suite.scopeID),
					ScopeSpecIdInfo: types.GetScopeSpecIDInfo(suite.specID),
				},
				Request: &types.ScopeRequest{ScopeId: suite.scopeUUID.String()},
			},
		},
		{
			name: "Unknown metadata scope id",
			url: fmt.Sprintf("%s/provenance/metadata/v1/scope/%s", baseURL, uuid.New()),
			headers: map[string]string{
				grpctypes.GRPCBlockHeightHeader: "1",
			},
			expErr: true,
			respType : &status.Status{},
			expected : &status.Status{},
		},
		{
			name: "Get metadata os locator params",
			url : fmt.Sprintf("%s/provenance/metadata/v1/locator/params", baseURL),
			headers : map[string]string{
				grpctypes.GRPCBlockHeightHeader: "1",
			},
			expErr: false,
			respType: &types.OSLocatorParamsResponse{},
			expected: &types.OSLocatorParamsResponse{
				Params:  types.DefaultOSLocatorParams(),
				Request: &types.OSLocatorParamsRequest{},
			},
		},
		{
			name : "Get os locator from owner address.",
			url : fmt.Sprintf("%s/provenance/metadata/v1/locator/%s", baseURL, suite.ownerAddr.String()),
			headers : map[string]string{
				grpctypes.GRPCBlockHeightHeader: "1",
			},
			expErr : false,
			respType : &types.OSLocatorResponse{},
			expected : &types.OSLocatorResponse{
				Locator: &suite.objectLocator,
				Request: &types.OSLocatorRequest{
					Owner: suite.ownerAddr.String(),
				},
			},
		},
		{
			name : "Get os locator from owner uri.",
			// only way i could get around http url parse isseus for rest
			// This encodes/decodes using a URL-compatible base64
			// format.
			url : fmt.Sprintf("%s/provenance/metadata/v1/locator/uri/%s", baseURL, b64.StdEncoding.EncodeToString([]byte(suite.uri))),
			headers : map[string]string{
				grpctypes.GRPCBlockHeightHeader: "1",
			},
			expErr : false,
			respType : &types.OSLocatorsByURIResponse{},
			expected : &types.OSLocatorsByURIResponse{
				Locators: []types.ObjectStoreLocator{{
					Owner:         suite.ownerAddr.String(),
					LocatorUri:    suite.uri,
					EncryptionKey: suite.encryptionKey.String(),
				}},
				Request: &types.OSLocatorsByURIRequest{
					Uri:        b64.StdEncoding.EncodeToString([]byte(suite.uri)),
					Pagination: nil,
				},
				Pagination: &query.PageResponse{
					NextKey: nil,
					Total:   1,
				},
			},
		},
		{
			name : "Get os locator's for given scope.",
			// only way i could get around http url parse issues for rest
			// This encodes/decodes using a URL-compatible base64
			// format.
			url : fmt.Sprintf("%s/provenance/metadata/v1/locator/scope/%s", baseURL, suite.scopeUUID),
			headers : map[string]string{
				grpctypes.GRPCBlockHeightHeader: "1",
			},
			expErr : false,
			respType : &types.OSLocatorsByScopeResponse{},
			expected : &types.OSLocatorsByScopeResponse{
				Locators: []types.ObjectStoreLocator{{
					Owner:         suite.ownerAddr1.String(),
					LocatorUri:    suite.uri1,
					EncryptionKey: suite.encryptionKey1.String(),
				}},
				Request: &types.OSLocatorsByScopeRequest{
					ScopeId: suite.scopeUUID.String(),
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			resp, err := sdktestutil.GetRequestWithHeaders(tc.url, tc.headers)
			suite.Require().NoError(err)
			err = val.ClientCtx.Codec.UnmarshalJSON(resp, tc.respType)
			if tc.expErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				suite.Require().Equal(tc.expected.String(), tc.respType.String())
			}
		})
	}
}

func (suite *IntegrationGRPCTestSuite) TestAllOSLocator() {
	val := suite.testnet.Validators[0]
	baseURL := val.APIAddress

	testCases := []struct {
		name     string
		url      string
		headers  map[string]string
		expErr   bool
		respType proto.Message
		expected proto.Message
	}{

		{
			name : "Get all os locator.",
			// only way i could get around http url parse issues for rest
			// This encodes/decodes using a URL-compatible base64
			// format.
			url : fmt.Sprintf("%s/provenance/metadata/v1/locators/all", baseURL),
			headers : map[string]string{
				grpctypes.GRPCBlockHeightHeader: "1",
			},
			expErr : false,
			respType : &types.OSAllLocatorsResponse{},
			expected : &types.OSAllLocatorsResponse{
				Locators: []types.ObjectStoreLocator{{
					Owner:         suite.ownerAddr1.String(),
					EncryptionKey: suite.encryptionKey1.String(),
					LocatorUri:    suite.uri1,
				}},
			},
		},
	}
	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			resp, err := sdktestutil.GetRequestWithHeaders(tc.url, tc.headers)
			suite.Require().NoError(err, "GetRequestWithHeaders err")
			err = val.ClientCtx.Codec.UnmarshalJSON(resp, tc.respType)
			if tc.expErr {
				suite.Require().Error(err, "UnmarshalJSON expected error")
			} else {
				suite.Require().NoError(err, "UnmarshalJSON unexpected error")
				suite.Require().True(strings.Contains(fmt.Sprint(tc.respType), fmt.Sprint(tc.expected)))
			}
		})
	}
}
