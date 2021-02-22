package cli_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"

	tmcli "github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client/flags"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	testnet "github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/testutil"

	"github.com/provenance-io/provenance/x/metadata/client/cli"
	"github.com/provenance-io/provenance/x/metadata/types"
	metadatatypes "github.com/provenance-io/provenance/x/metadata/types"
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

func ownerPartyList(addresses ...string) []types.Party {
	retval := make([]types.Party, len(addresses))
	for i, addr := range addresses {
		retval[i] = types.Party{Address: addr, Role: types.PartyType_PARTY_TYPE_OWNER}
	}
	return retval
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.accountKey = secp256k1.GenPrivKeyFromSecret([]byte("acc2"))
	addr, err := sdk.AccAddressFromHex(s.accountKey.PubKey().Address().String())
	s.Require().NoError(err)
	s.accountAddr = addr
	s.T().Log("setting up integration test suite")

	cfg := testutil.DefaultTestNetworkConfig()

	genesisState := cfg.GenesisState
	cfg.NumValidators = 1

	s.pubkey1 = secp256k1.GenPrivKey().PubKey()
	s.user1Addr = sdk.AccAddress(s.pubkey1.Address())
	s.user1 = s.user1Addr.String()

	s.pubkey2 = secp256k1.GenPrivKey().PubKey()
	s.user2Addr = sdk.AccAddress(s.pubkey2.Address())
	s.user2 = s.user2Addr.String()

	s.scopeUUID = uuid.New()
	s.scopeID = types.ScopeMetadataAddress(s.scopeUUID)
	s.specUUID = uuid.New()
	s.specID = types.ScopeSpecMetadataAddress(s.specUUID)

	s.scope = *metadatatypes.NewScope(s.scopeID, s.specID, ownerPartyList(s.user1), []string{s.user1}, s.user1)

	var metadataData metadatatypes.GenesisState
	metadataData.Params = metadatatypes.DefaultParams()
	metadataData.Scopes = append(metadataData.Scopes, s.scope)
	metadataDataBz, err := cfg.Codec.MarshalJSON(&metadataData)
	s.Require().NoError(err)
	genesisState[metadatatypes.ModuleName] = metadataDataBz

	// adding to auth genesis does not work due to cosmos sdk overwritting it
	var authData authtypes.GenesisState
	s.Require().NoError(cfg.Codec.UnmarshalJSON(genesisState[authtypes.ModuleName], &authData))
	genAccount, err := codectypes.NewAnyWithValue(authtypes.NewBaseAccount(s.accountAddr, s.accountKey.PubKey(), 1, 1))
	s.Require().NoError(err)
	authData.Accounts = append(authData.Accounts, genAccount)

	user1Account, err := codectypes.NewAnyWithValue(authtypes.NewBaseAccount(s.user1Addr, s.pubkey1, 2, 2))
	s.Require().NoError(err)
	authData.Accounts = append(authData.Accounts, user1Account)

	user2Account, err := codectypes.NewAnyWithValue(authtypes.NewBaseAccount(s.user2Addr, s.pubkey2, 3, 3))
	s.Require().NoError(err)
	authData.Accounts = append(authData.Accounts, user2Account)

	authDataBz, err := cfg.Codec.MarshalJSON(&authData)
	s.Require().NoError(err)
	genesisState[authtypes.ModuleName] = authDataBz

	cfg.GenesisState = genesisState

	s.cfg = cfg

	// TODO: This is overwritting some of our genesis states https://github.com/provenance-io/provenance/issues/81
	s.testnet = testnet.New(s.T(), cfg)

	_, err = s.testnet.WaitForHeight(1)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.testnet.WaitForNextBlock()
	s.T().Log("tearing down integration test suite")
	s.testnet.Cleanup()
}

func (s *IntegrationTestSuite) TestGetAttributeParamsCmd() {
	testCases := []struct {
		name           string
		args           []string
		expectedOutput string
	}{
		{
			"get params as json output",
			[]string{fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			"{}",
		},
		{
			"get params as text output",
			[]string{fmt.Sprintf("--%s=text", tmcli.OutputFlag)},
			"{}",
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetMetadataParamsCmd()
			clientCtx := s.testnet.Validators[0].ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedOutput, strings.TrimSpace(out.String()))
		})
	}
}

func (s *IntegrationTestSuite) TestGetMetadataScopeCmd() {
	testCases := []struct {
		name           string
		args           []string
		expectedOutput string
	}{
		{
			"get scope as json output",
			[]string{s.scopeUUID.String(), fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			fmt.Sprintf("{\"scope_id\":\"%s\",\"specification_id\":\"%s\",\"parties\":[{\"address\":\"%s\",\"role\":\"%s\"}],\"data_access\":[\"%s\"],\"value_owner_address\":\"%s\"}",
				s.scope.ScopeId,
				s.scope.SpecificationId.String(),
				s.scope.Parties[0].Address,
				s.scope.Parties[0].Role.String(),
				s.scope.DataAccess[0],
				s.scope.ValueOwnerAddress),
		},
		{
			"get scope as text output",
			[]string{s.scopeUUID.String(), fmt.Sprintf("--%s=text", tmcli.OutputFlag)},
			fmt.Sprintf(`data_access:
- %s
parties:
- address: %s
  role: %s
scope_id: %s
specification_id: %s
value_owner_address: %s`,
				s.scope.DataAccess[0],
				s.scope.Parties[0].Address,
				s.scope.Parties[0].Role.String(),
				s.scope.ScopeId,
				s.scope.SpecificationId.String(),
				s.scope.ValueOwnerAddress),
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetMetadataScopeCmd()
			clientCtx := s.testnet.Validators[0].ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedOutput, strings.TrimSpace(out.String()))
		})
	}
}

func (s *IntegrationTestSuite) TestAddMetadataScopeCmd() {

	scopeUUID := uuid.New().String()

	testCases := []struct {
		name         string
		cmd          *cobra.Command
		args         []string
		expectErr    bool
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"Should successfully add metadata scope",
			cli.AddMetadataScopeCmd(),
			[]string{
				scopeUUID,
				uuid.New().String(),
				s.user1,
				s.user1,
				s.user1,
				s.testnet.Validators[0].Address.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"Should fail to add metadata scope, incorrect scope uuid",
			cli.AddMetadataScopeCmd(),
			[]string{
				"not-a-uuid",
				uuid.New().String(),
				s.user1,
				s.user1,
				s.user1,
				s.testnet.Validators[0].Address.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, &sdk.TxResponse{}, 0,
		},
		{
			"Should fail to add metadata scope, incorrect scope spec uuid",
			cli.AddMetadataScopeCmd(),
			[]string{
				uuid.New().String(),
				"not-a-uuid",
				s.user1,
				s.user1,
				s.user1,
				s.testnet.Validators[0].Address.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, &sdk.TxResponse{}, 0,
		},
		{
			"Should fail to add metadata scope, incorrect scope spec uuid",
			cli.AddMetadataScopeCmd(),
			[]string{
				uuid.New().String(),
				"not-a-uuid",
				s.user1,
				s.user1,
				s.user1,
				s.testnet.Validators[0].Address.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, &sdk.TxResponse{}, 0,
		},
		{
			"Should fail to add metadata scope, incorrect owner address format",
			cli.AddMetadataScopeCmd(),
			[]string{
				uuid.New().String(),
				uuid.New().String(),
				"incorrect,incorrect",
				s.user1,
				s.user1,
				s.testnet.Validators[0].Address.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, &sdk.TxResponse{}, 0,
		},
		{
			"Should fail to add metadata scope, incorrect data access format",
			cli.AddMetadataScopeCmd(),
			[]string{
				uuid.New().String(),
				uuid.New().String(),
				s.user1,
				"incorrect,incorrect",
				s.user1,
				s.testnet.Validators[0].Address.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.user1),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, &sdk.TxResponse{}, 0,
		},
		{
			"Should fail to add metadata scope, incorrect value owner address",
			cli.AddMetadataScopeCmd(),
			[]string{
				uuid.New().String(),
				uuid.New().String(),
				s.user1,
				s.user1,
				"incorrect",
				s.testnet.Validators[0].Address.String(),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.user1),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, &sdk.TxResponse{}, 0,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			clientCtx := s.testnet.Validators[0].ClientCtx
			out, err := clitestutil.ExecTestCLICmd(clientCtx, tc.cmd, tc.args)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestRemoveMetadataScopeCmd() {

	userId := s.testnet.Validators[0].Address.String()
	scopeUUID := uuid.New().String()

	testCases := []struct {
		name         string
		cmd          *cobra.Command
		args         []string
		expectErr    bool
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"Should successfully add metadata scope for testing scope removal",
			cli.AddMetadataScopeCmd(),
			[]string{
				scopeUUID,
				uuid.New().String(),
				userId,
				userId,
				userId,
				userId,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
		{
			"Should fail to remove metadata scope, invalid scopeid",
			cli.RemoveMetadataScopeCmd(),
			[]string{
				"not-valid",
				userId,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, &sdk.TxResponse{}, 0,
		},
		{
			"Should fail to remove metadata scope, invalid userid",
			cli.RemoveMetadataScopeCmd(),
			[]string{
				scopeUUID,
				"not-valid",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			true, &sdk.TxResponse{}, 0,
		},
		{
			"Should remove metadata scope",
			cli.RemoveMetadataScopeCmd(),
			[]string{
				scopeUUID,
				userId,
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			false, &sdk.TxResponse{}, 0,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			clientCtx := s.testnet.Validators[0].ClientCtx
			out, err := clitestutil.ExecTestCLICmd(clientCtx, tc.cmd, tc.args)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code)
			}
		})
	}
}
