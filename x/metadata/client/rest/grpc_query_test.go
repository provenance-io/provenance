package rest_test

import (
	b64 "encoding/base64"
	"fmt"
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

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

func (s *IntegrationGRPCTestSuite) SetupSuite() {
	pioconfig.SetProvenanceConfig("atom", 0)
	s.accountKey = secp256k1.GenPrivKeyFromSecret([]byte("accountKey"))
	addr, err := sdk.AccAddressFromHexUnsafe(s.accountKey.PubKey().Address().String())
	s.Require().NoError(err)
	s.accountAddr = addr
	s.T().Log("setting up integration test suite")

	cfg := testutil.DefaultTestNetworkConfig()

	genesisState := cfg.GenesisState
	cfg.NumValidators = 1

	var authData authtypes.GenesisState
	s.Require().NoError(cfg.Codec.UnmarshalJSON(genesisState[authtypes.ModuleName], &authData))
	genAccount, err := codectypes.NewAnyWithValue(&authtypes.BaseAccount{
		Address:       s.accountAddr.String(),
		AccountNumber: 1,
		Sequence:      0,
	})
	s.Require().NoError(err)
	authData.Accounts = append(authData.Accounts, genAccount)

	s.pubkey1 = secp256k1.GenPrivKeyFromSecret([]byte("pubkey1")).PubKey()
	s.user1Addr = sdk.AccAddress(s.pubkey1.Address())
	s.user1 = s.user1Addr.String()

	s.pubkey2 = secp256k1.GenPrivKeyFromSecret([]byte("pubkey2")).PubKey()
	s.user2Addr = sdk.AccAddress(s.pubkey2.Address())
	s.user2 = s.user2Addr.String()

	s.scopeUUID = uuid.New()
	s.scopeID = types.ScopeMetadataAddress(s.scopeUUID)

	s.specUUID = uuid.New()
	s.specID = types.ScopeSpecMetadataAddress(s.specUUID)

	s.scope = *types.NewScope(s.scopeID, s.specID, ownerPartyList(s.user1), []string{s.user1}, s.user1, false)
	// Configure Genesis data for metadata module

	// add os locator
	s.ownerAddr = s.accountAddr
	s.uri = "http://foo.com"
	s.encryptionKey = sdk.AccAddress{}
	s.objectLocator = types.NewOSLocatorRecord(s.ownerAddr, s.encryptionKey, s.uri)

	s.ownerAddr1 = s.user1Addr
	s.uri1 = "http://bar.com"
	s.encryptionKey1 = s.ownerAddr
	s.objectLocator1 = types.NewOSLocatorRecord(s.ownerAddr1, s.encryptionKey1, s.uri1)

	var metadataData types.GenesisState
	metadataData.Params = types.DefaultParams()
	metadataData.OSLocatorParams = types.DefaultOSLocatorParams()
	metadataData.Scopes = append(metadataData.Scopes, s.scope)
	metadataData.ObjectStoreLocators = append(metadataData.ObjectStoreLocators, s.objectLocator, s.objectLocator1)
	metadataDataBz, err := cfg.Codec.MarshalJSON(&metadataData)
	s.Require().NoError(err)

	genesisState[types.ModuleName] = metadataDataBz

	cfg.GenesisState = genesisState

	s.cfg = cfg

	s.testnet, err = testnet.New(s.T(), s.T().TempDir(), cfg)
	s.Require().NoError(err, "creating testnet")

	_, err = s.testnet.WaitForHeight(1)
	s.Require().NoError(err, "waiting for height 1")
}

func (s *IntegrationGRPCTestSuite) TearDownSuite() {
	testutil.CleanUp(s.testnet, s.T())
}

func TestIntegrationGRPCTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationGRPCTestSuite))
}

func (s *IntegrationGRPCTestSuite) TestGRPCQueries() {
	val := s.testnet.Validators[0]
	baseURL := val.APIAddress
	unknownScopeUUID := uuid.New()
	unknownScopeAddr := types.ScopeMetadataAddress(unknownScopeUUID)

	testCases := []struct {
		name     string
		url      string
		respType proto.Message
		expected proto.Message
	}{
		{
			name:     "Get metadata params no request",
			url:      fmt.Sprintf("%s/provenance/metadata/v1/params", baseURL),
			respType: &types.QueryParamsResponse{},
			expected: &types.QueryParamsResponse{Params: types.DefaultParams()},
		},
		{
			name:     "Get metadata params with request",
			url:      fmt.Sprintf("%s/provenance/metadata/v1/params?include_request=true", baseURL),
			respType: &types.QueryParamsResponse{},
			expected: &types.QueryParamsResponse{
				Params:  types.DefaultParams(),
				Request: &types.QueryParamsRequest{IncludeRequest: true},
			},
		},
		{
			name:     "Get metadata scope by id no req no info",
			url:      fmt.Sprintf("%s/provenance/metadata/v1/scope/%s?exclude_id_info=true", baseURL, s.scopeUUID),
			respType: &types.ScopeResponse{},
			expected: &types.ScopeResponse{Scope: &types.ScopeWrapper{Scope: &s.scope}},
		},
		{
			name:     "Get metadata scope by id with req no info",
			url:      fmt.Sprintf("%s/provenance/metadata/v1/scope/%s?include_request=true&exclude_id_info=true", baseURL, s.scopeUUID),
			respType: &types.ScopeResponse{},
			expected: &types.ScopeResponse{
				Scope: &types.ScopeWrapper{Scope: &s.scope},
				Request: &types.ScopeRequest{
					ScopeId:        s.scopeUUID.String(),
					ExcludeIdInfo:  true,
					IncludeRequest: true,
				},
			},
		},
		{
			name:     "Get metadata scope by id no req with info",
			url:      fmt.Sprintf("%s/provenance/metadata/v1/scope/%s", baseURL, s.scopeUUID),
			respType: &types.ScopeResponse{},
			expected: &types.ScopeResponse{
				Scope: &types.ScopeWrapper{
					Scope:           &s.scope,
					ScopeIdInfo:     types.GetScopeIDInfo(s.scopeID),
					ScopeSpecIdInfo: types.GetScopeSpecIDInfo(s.specID),
				},
			},
		},
		{
			name:     "Get metadata scope by id with req and info",
			url:      fmt.Sprintf("%s/provenance/metadata/v1/scope/%s?include_request=true", baseURL, s.scopeUUID),
			respType: &types.ScopeResponse{},
			expected: &types.ScopeResponse{
				Scope: &types.ScopeWrapper{
					Scope:           &s.scope,
					ScopeIdInfo:     types.GetScopeIDInfo(s.scopeID),
					ScopeSpecIdInfo: types.GetScopeSpecIDInfo(s.specID),
				},
				Request: &types.ScopeRequest{
					ScopeId:        s.scopeUUID.String(),
					IncludeRequest: true,
				},
			},
		},
		{
			name:     "Unknown metadata scope id",
			url:      fmt.Sprintf("%s/provenance/metadata/v1/scope/%s", baseURL, unknownScopeUUID),
			respType: &types.ScopeResponse{},
			expected: &types.ScopeResponse{
				Scope: &types.ScopeWrapper{ScopeIdInfo: types.GetScopeIDInfo(unknownScopeAddr)},
			},
		},
		{
			name:     "Get metadata os locator params",
			url:      fmt.Sprintf("%s/provenance/metadata/v1/locator/params", baseURL),
			respType: &types.OSLocatorParamsResponse{},
			expected: &types.OSLocatorParamsResponse{Params: types.DefaultOSLocatorParams()},
		},
		{
			name:     "Get os locator from owner address no req",
			url:      fmt.Sprintf("%s/provenance/metadata/v1/locator/%s", baseURL, s.ownerAddr.String()),
			respType: &types.OSLocatorResponse{},
			expected: &types.OSLocatorResponse{Locator: &s.objectLocator},
		},
		{
			name:     "Get os locator from owner address with req",
			url:      fmt.Sprintf("%s/provenance/metadata/v1/locator/%s?include_request=true", baseURL, s.ownerAddr.String()),
			respType: &types.OSLocatorResponse{},
			expected: &types.OSLocatorResponse{
				Locator: &s.objectLocator,
				Request: &types.OSLocatorRequest{
					Owner:          s.ownerAddr.String(),
					IncludeRequest: true,
				},
			},
		},
		{
			name: "Get os locator from owner uri no req",
			// only way i could get around http url parse isseus for rest
			// This encodes/decodes using a URL-compatible base64
			// format.
			url:      fmt.Sprintf("%s/provenance/metadata/v1/locator/uri/%s", baseURL, b64.StdEncoding.EncodeToString([]byte(s.uri))),
			respType: &types.OSLocatorsByURIResponse{},
			expected: &types.OSLocatorsByURIResponse{
				Locators: []types.ObjectStoreLocator{{
					Owner:         s.ownerAddr.String(),
					LocatorUri:    s.uri,
					EncryptionKey: s.encryptionKey.String(),
				}},
				Pagination: &query.PageResponse{Total: 1},
			},
		},
		{
			name: "Get os locator from owner uri with req",
			// only way i could get around http url parse isseus for rest
			// This encodes/decodes using a URL-compatible base64
			// format.
			url:      fmt.Sprintf("%s/provenance/metadata/v1/locator/uri/%s?include_request=true", baseURL, b64.StdEncoding.EncodeToString([]byte(s.uri))),
			respType: &types.OSLocatorsByURIResponse{},
			expected: &types.OSLocatorsByURIResponse{
				Locators: []types.ObjectStoreLocator{{
					Owner:         s.ownerAddr.String(),
					LocatorUri:    s.uri,
					EncryptionKey: s.encryptionKey.String(),
				}},
				Request: &types.OSLocatorsByURIRequest{
					Uri:            b64.StdEncoding.EncodeToString([]byte(s.uri)),
					IncludeRequest: true,
				},
				Pagination: &query.PageResponse{Total: 1},
			},
		},
		{
			name:     "Get os locator's for given scope no req",
			url:      fmt.Sprintf("%s/provenance/metadata/v1/locator/scope/%s", baseURL, s.scopeUUID),
			respType: &types.OSLocatorsByScopeResponse{},
			expected: &types.OSLocatorsByScopeResponse{
				Locators: []types.ObjectStoreLocator{{
					Owner:         s.ownerAddr1.String(),
					LocatorUri:    s.uri1,
					EncryptionKey: s.encryptionKey1.String(),
				}},
			},
		},
		{
			name:     "Get os locator's for given scope with req",
			url:      fmt.Sprintf("%s/provenance/metadata/v1/locator/scope/%s?include_request=true", baseURL, s.scopeUUID),
			respType: &types.OSLocatorsByScopeResponse{},
			expected: &types.OSLocatorsByScopeResponse{
				Locators: []types.ObjectStoreLocator{{
					Owner:         s.ownerAddr1.String(),
					LocatorUri:    s.uri1,
					EncryptionKey: s.encryptionKey1.String(),
				}},
				Request: &types.OSLocatorsByScopeRequest{
					ScopeId:        s.scopeUUID.String(),
					IncludeRequest: true,
				},
			},
		},
		{
			name: "Get all os locators no req",
			// only way i could get around http url parse issues for rest
			// This encodes/decodes using a URL-compatible base64
			// format.
			url:      fmt.Sprintf("%s/provenance/metadata/v1/locators/all", baseURL),
			respType: &types.OSAllLocatorsResponse{},
			expected: &types.OSAllLocatorsResponse{
				Locators: []types.ObjectStoreLocator{
					{
						Owner:         s.ownerAddr.String(),
						EncryptionKey: s.encryptionKey.String(),
						LocatorUri:    s.uri,
					},
					{
						Owner:         s.ownerAddr1.String(),
						EncryptionKey: s.encryptionKey1.String(),
						LocatorUri:    s.uri1,
					},
				},
				Pagination: &query.PageResponse{Total: 2},
			},
		},
		{
			name: "Get all os locators with req",
			// only way i could get around http url parse issues for rest
			// This encodes/decodes using a URL-compatible base64
			// format.
			url:      fmt.Sprintf("%s/provenance/metadata/v1/locators/all?include_request=true", baseURL),
			respType: &types.OSAllLocatorsResponse{},
			expected: &types.OSAllLocatorsResponse{
				Locators: []types.ObjectStoreLocator{
					{
						Owner:         s.ownerAddr.String(),
						EncryptionKey: s.encryptionKey.String(),
						LocatorUri:    s.uri,
					},
					{
						Owner:         s.ownerAddr1.String(),
						EncryptionKey: s.encryptionKey1.String(),
						LocatorUri:    s.uri1,
					},
				},
				Request:    &types.OSAllLocatorsRequest{IncludeRequest: true},
				Pagination: &query.PageResponse{Total: 2},
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			headers := map[string]string{grpctypes.GRPCBlockHeightHeader: "1"}
			resp, err := sdktestutil.GetRequestWithHeaders(tc.url, headers)
			s.Require().NoError(err, "GetRequestWithHeaders")
			err = val.ClientCtx.Codec.UnmarshalJSON(resp, tc.respType)
			s.Require().NoError(err, "UnmarshalJSON:\n%s", string(resp))
			s.Require().Equal(tc.expected.String(), tc.respType.String())
		})
	}
}
