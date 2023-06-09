package keeper_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/testutil"
	"github.com/provenance-io/provenance/x/attribute/types"
	markertypes "github.com/provenance-io/provenance/x/marker/types"
	metadatatypes "github.com/provenance-io/provenance/x/metadata/types"
)

type QueryServerTestSuite struct {
	suite.Suite

	app         *simapp.App
	ctx         sdk.Context
	cfg         network.Config
	queryClient types.QueryClient

	privkey1   cryptotypes.PrivKey
	pubkey1    cryptotypes.PubKey
	owner1     string
	owner1Addr sdk.AccAddress
	acct1      authtypes.AccountI

	addresses []sdk.AccAddress
}

func (s *QueryServerTestSuite) SetupTest() {
	s.app = simapp.SetupQuerier(s.T())
	s.ctx = s.app.BaseApp.NewContext(true, tmproto.Header{})
	s.app.AccountKeeper.SetParams(s.ctx, authtypes.DefaultParams())
	s.app.BankKeeper.SetParams(s.ctx, banktypes.DefaultParams())
	s.cfg = testutil.DefaultTestNetworkConfig()
	queryHelper := baseapp.NewQueryServerTestHelper(s.ctx, s.app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, s.app.AttributeKeeper)
	s.queryClient = types.NewQueryClient(queryHelper)

	s.privkey1 = secp256k1.GenPrivKey()
	s.pubkey1 = s.privkey1.PubKey()
	s.owner1Addr = sdk.AccAddress(s.pubkey1.Address())
	s.owner1 = s.owner1Addr.String()
	acc := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, s.owner1Addr)
	s.app.AccountKeeper.SetAccount(s.ctx, acc)
}

func TestQuerierTestSuite(t *testing.T) {
	suite.Run(t, new(QueryServerTestSuite))
}

func (s *QueryServerTestSuite) TestAttributeAccountsQuery() {
	name1 := "example.attribute"
	name2 := "foo.example.attribute"
	s.Require().NoError(s.app.NameKeeper.SetNameRecord(s.ctx, name1, s.owner1Addr, false))
	s.Require().NoError(s.app.NameKeeper.SetNameRecord(s.ctx, name2, s.owner1Addr, false))
	accounts := make([]string, 100)
	for i := 0; i < 100; i++ {
		privkey := secp256k1.GenPrivKey()
		pubkey := privkey.PubKey()
		acctAddr := sdk.AccAddress(pubkey.Address())
		acct := acctAddr.String()
		accounts[i] = acct
		s.Require().NoError(s.app.AttributeKeeper.SetAttribute(s.ctx,
			types.Attribute{
				Name:          name1,
				Value:         []byte("0123456789"),
				Address:       acct,
				AttributeType: types.AttributeType_String,
			}, s.owner1Addr))
		// add a second attr with same name diff type for testing of reduced addresses
		s.Require().NoError(s.app.AttributeKeeper.SetAttribute(s.ctx, types.Attribute{
			Name:          name1,
			Value:         []byte("1"),
			Address:       acct,
			AttributeType: types.AttributeType_Int,
		}, s.owner1Addr))
		s.Require().NoError(s.app.AttributeKeeper.SetAttribute(s.ctx, types.Attribute{
			Name:          name2,
			Value:         []byte("1"),
			Address:       acct,
			AttributeType: types.AttributeType_Int,
		}, s.owner1Addr))
	}
	results, err := s.queryClient.AttributeAccounts(s.ctx, &types.QueryAttributeAccountsRequest{AttributeName: name1})
	s.Assert().NoError(err)
	s.Assert().Len(results.Accounts, 100)
	s.Assert().ElementsMatch(accounts, results.Accounts)

	results, err = s.queryClient.AttributeAccounts(s.ctx, &types.QueryAttributeAccountsRequest{AttributeName: name2})
	s.Assert().NoError(err)
	s.Assert().Len(results.Accounts, 100)
	s.Assert().ElementsMatch(accounts, results.Accounts)

	var allResults []string
	results, err = s.queryClient.AttributeAccounts(s.ctx, &types.QueryAttributeAccountsRequest{AttributeName: name2, Pagination: &query.PageRequest{Limit: 50}})
	s.Assert().NoError(err)
	s.Assert().Len(results.Accounts, 50)
	allResults = append(allResults, results.Accounts...)

	results, err = s.queryClient.AttributeAccounts(s.ctx, &types.QueryAttributeAccountsRequest{AttributeName: name2, Pagination: &query.PageRequest{
		Key:   results.Pagination.NextKey,
		Limit: 50}})
	s.Assert().NoError(err)
	s.Assert().Len(results.Accounts, 50)
	allResults = append(allResults, results.Accounts...)

	s.Assert().ElementsMatch(accounts, allResults)
}

func (s *QueryServerTestSuite) TestAccountData() {
	// Use GetModuleAccount to ensure that the account exists.
	attrModAcc := s.app.AccountKeeper.GetModuleAccount(s.ctx, types.ModuleName)
	attrModAddr := attrModAcc.GetAddress()

	err := s.app.NameKeeper.SetNameRecord(s.ctx, types.AccountDataName, attrModAddr, true)
	s.Require().NoError(err, "SetNameRecord(%q)", types.AccountDataName)

	addrWithoutData := sdk.AccAddress("addrWithoutData_____").String()
	addrWithData := sdk.AccAddress("addrWithData________").String()
	scopeIDWithData := metadatatypes.ScopeMetadataAddress(uuid.New()).String()
	markerAddrWithData := markertypes.MustGetMarkerAddress("mytestdenom").String()

	addrData := "this is some data"
	scopeIDData := "this is some scope data"
	markerData := "this is some marker data"

	err = s.app.AttributeKeeper.SetAccountData(s.ctx, addrWithData, addrData)
	s.Require().NoError(err, "Setup: SetAccountData addrWithData")
	err = s.app.AttributeKeeper.SetAccountData(s.ctx, scopeIDWithData, scopeIDData)
	s.Require().NoError(err, "Setup: SetAccountData scopeIDWithData")
	err = s.app.AttributeKeeper.SetAccountData(s.ctx, markerAddrWithData, markerData)
	s.Require().NoError(err, "Setup: SetAccountData markerAddrWithData")

	req := func(account string) *types.QueryAccountDataRequest {
		return &types.QueryAccountDataRequest{Account: account}
	}
	resp := func(value string) *types.QueryAccountDataResponse {
		return &types.QueryAccountDataResponse{Value: value}
	}

	tests := []struct {
		name   string
		req    *types.QueryAccountDataRequest
		resp   *types.QueryAccountDataResponse
		expErr string
	}{
		{
			name:   "nil request",
			req:    nil,
			resp:   nil,
			expErr: "rpc error: code = InvalidArgument desc = invalid request",
		},
		// Not sure how to cause GetAccountData to return an error.
		{
			name: "address without data",
			req:  req(addrWithoutData),
			resp: resp(""),
		},
		{
			name: "account with data",
			req:  req(addrWithData),
			resp: resp(addrData),
		},
		{
			name: "scope with data",
			req:  req(scopeIDWithData),
			resp: resp(scopeIDData),
		},
		{
			name: "marker with data",
			req:  req(markerAddrWithData),
			resp: resp(markerData),
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			actual, err := s.app.AttributeKeeper.AccountData(s.ctx, tc.req)
			if len(tc.expErr) > 0 {
				s.Require().EqualErrorf(err, tc.expErr, "AccountData error")
			} else {
				s.Require().NoError(err, "AccountData error")
			}
			s.Assert().Equal(tc.resp, actual, "AccountData response")
		})
	}
}
