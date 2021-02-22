package rest_test

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"google.golang.org/genproto/googleapis/rpc/status"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdktestutil "github.com/cosmos/cosmos-sdk/testutil"
	testnet "github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/provenance-io/provenance/testutil"

	"github.com/provenance-io/provenance/x/metadata/types"
	metadatatypes "github.com/provenance-io/provenance/x/metadata/types"

	"github.com/gogo/protobuf/proto"

	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
)

type IntegrationTestSuite struct {
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

	scope     metadatatypes.Scope
	scopeUUID uuid.UUID
	scopeID   types.MetadataAddress

	specUUID uuid.UUID
	specID   types.MetadataAddress
}

func (suite *IntegrationTestSuite) SetupSuite() {
	suite.accountKey = secp256k1.GenPrivKeyFromSecret([]byte("acc2"))
	addr, err := sdk.AccAddressFromHex(suite.accountKey.PubKey().Address().String())
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

	suite.scope = *metadatatypes.NewScope(suite.scopeID, suite.specID, []string{suite.user1}, []string{suite.user1}, suite.user1)
	// Configure Genesis data for metadata module

	var metadataData metadatatypes.GenesisState
	metadataData.Params = metadatatypes.DefaultParams()
	metadataData.Scopes = append(metadataData.Scopes, suite.scope)
	metadataDataBz, err := cfg.Codec.MarshalJSON(&metadataData)
	suite.Require().NoError(err)

	genesisState[metadatatypes.ModuleName] = metadataDataBz

	cfg.GenesisState = genesisState

	suite.cfg = cfg

	suite.testnet = testnet.New(suite.T(), cfg)

	_, err = suite.testnet.WaitForHeight(1)
	suite.Require().NoError(err)
}

func (suite *IntegrationTestSuite) TearDownSuite() {
	suite.testnet.WaitForNextBlock()
	suite.T().Log("tearing down integration test suite")
	suite.testnet.Cleanup()
}
func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (suite *IntegrationTestSuite) TestGRPCQueries() {
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
			"Get metadata params",
			fmt.Sprintf("%s/provenance/metadata/v1/params", baseURL),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "1",
			},
			false,
			&metadatatypes.QueryParamsResponse{},
			&metadatatypes.QueryParamsResponse{Params: metadatatypes.DefaultParams()},
		},
		{
			"Get metadata scope by id",
			fmt.Sprintf("%s/provenance/metadata/v1/scope/%s", baseURL, suite.scopeUUID),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "1",
			},
			false,
			&metadatatypes.ScopeResponse{},
			&metadatatypes.ScopeResponse{Scope: &suite.scope},
		},
		{
			"Unknown metadata scope id",
			fmt.Sprintf("%s/provenance/metadata/v1/scope/%s", baseURL, uuid.New()),
			map[string]string{
				grpctypes.GRPCBlockHeightHeader: "1",
			},
			true,
			&status.Status{},
			&status.Status{},
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			resp, err := sdktestutil.GetRequestWithHeaders(tc.url, tc.headers)
			suite.Require().NoError(err)
			err = val.ClientCtx.JSONMarshaler.UnmarshalJSON(resp, tc.respType)
			if tc.expErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				suite.Require().Equal(tc.expected.String(), tc.respType.String())
			}
		})
	}
}
