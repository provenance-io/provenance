package cli_test

import (
	"encoding/base64"
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	tmcli "github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	testnet "github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/internal/antewrapper"
	"github.com/provenance-io/provenance/internal/pioconfig"
	"github.com/provenance-io/provenance/testutil"
	namecli "github.com/provenance-io/provenance/x/name/client/cli"
	nametypes "github.com/provenance-io/provenance/x/name/types"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg     testnet.Config
	testnet *testnet.Network

	accountAddr sdk.AccAddress
	accountKey  *secp256k1.PrivKey

	account2Addr  sdk.AccAddress
	account2Key   *secp256k1.PrivKey
	acc2NameCount int
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.accountKey = secp256k1.GenPrivKeyFromSecret([]byte("acc2"))
	addr, err := sdk.AccAddressFromHexUnsafe(s.accountKey.PubKey().Address().String())
	s.Require().NoError(err)
	s.accountAddr = addr

	s.account2Key = secp256k1.GenPrivKeyFromSecret([]byte("acc22"))
	addr2, err2 := sdk.AccAddressFromHexUnsafe(s.account2Key.PubKey().Address().String())
	s.Require().NoError(err2)
	s.account2Addr = addr2
	s.acc2NameCount = 50

	s.T().Log("setting up integration test suite")
	pioconfig.SetProvenanceConfig("", 0)

	cfg := testutil.DefaultTestNetworkConfig()

	genesisState := cfg.GenesisState
	cfg.NumValidators = 1

	var nameData nametypes.GenesisState
	nameData.Params.AllowUnrestrictedNames = true
	nameData.Params.MaxNameLevels = 2
	nameData.Params.MaxSegmentLength = 32
	nameData.Params.MinSegmentLength = 1
	nameData.Bindings = append(nameData.Bindings, nametypes.NewNameRecord("attribute", s.accountAddr, false))
	nameData.Bindings = append(nameData.Bindings, nametypes.NewNameRecord("example.attribute", s.accountAddr, false))
	for i := 0; i < s.acc2NameCount; i++ {
		nameData.Bindings = append(nameData.Bindings, nametypes.NewNameRecord(toWritten(i), s.account2Addr, false))
	}
	nameDataBz, err := cfg.Codec.MarshalJSON(&nameData)
	s.Require().NoError(err)
	genesisState[nametypes.ModuleName] = nameDataBz

	cfg.GenesisState = genesisState

	cfg.ChainID = antewrapper.SimAppChainID
	s.cfg = cfg

	s.testnet, err = testnet.New(s.T(), s.T().TempDir(), cfg)
	s.Require().NoError(err, "creating testnet")

	_, err = s.testnet.WaitForHeight(1)
	s.Require().NoError(err, "waiting for height 1")
}

func (s *IntegrationTestSuite) TearDownSuite() {
	testutil.CleanUp(s.testnet, s.T())
}

// toWritten converts an integer to a written string version.
// Originally, this was the full written string, e.g. 38 => "thirtyEight" but that ended up being too long for
// an attribute name segment, so it got trimmed down, e.g. 115 => "onehun15".
func toWritten(i int) string {
	if i < 0 || i > 999 {
		panic("cannot convert negative numbers or numbers larger than 999 to written string")
	}
	switch i {
	case 0:
		return "zero"
	case 1:
		return "one"
	case 2:
		return "two"
	case 3:
		return "three"
	case 4:
		return "four"
	case 5:
		return "five"
	case 6:
		return "six"
	case 7:
		return "seven"
	case 8:
		return "eight"
	case 9:
		return "nine"
	case 10:
		return "ten"
	case 11:
		return "eleven"
	case 12:
		return "twelve"
	case 13:
		return "thirteen"
	case 14:
		return "fourteen"
	case 15:
		return "fifteen"
	case 16:
		return "sixteen"
	case 17:
		return "seventeen"
	case 18:
		return "eighteen"
	case 19:
		return "nineteen"
	case 20:
		return "twenty"
	case 30:
		return "thirty"
	case 40:
		return "forty"
	case 50:
		return "fifty"
	case 60:
		return "sixty"
	case 70:
		return "seventy"
	case 80:
		return "eighty"
	case 90:
		return "ninety"
	default:
		var r int
		var l string
		switch {
		case i < 100:
			r = i % 10
			l = toWritten(i - r)
		default:
			r = i % 100
			l = toWritten(i/100) + "hun"
		}
		if r == 0 {
			return l
		}
		return l + "." + fmt.Sprintf("%d", r)
	}
}

func limitArg(pageSize int) string {
	return fmt.Sprintf("--limit=%d", pageSize)
}

func pageKeyArg(nextKey string) string {
	return fmt.Sprintf("--page-key=%s", nextKey)
}

func (s *IntegrationTestSuite) TestGetNameParamsCmd() {
	testCases := []struct {
		name           string
		args           []string
		expectedOutput string
	}{
		{
		    name: 	"json output",
		    args: 	[]string{fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
		    expectedOutput: 	"{\"max_segment_length\":32,\"min_segment_length\":1,\"max_name_levels\":2,\"allow_unrestricted_names\":true}",
		},
		{
			name: "text output",
			args: []string{fmt.Sprintf("--%s=text", tmcli.OutputFlag)},
			expectedOutput: `allow_unrestricted_names: true
max_name_levels: 2
max_segment_length: 32
min_segment_length: 1`,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := namecli.QueryParamsCmd()
			clientCtx := s.testnet.Validators[0].ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedOutput, strings.TrimSpace(out.String()))
		})
	}
}

func (s *IntegrationTestSuite) TestResolveNameCommand() {
	testCases := []struct {
		name           string
		args           []string
		expectedOutput string
	}{
		{
			name: "query name, json output",
			args: []string{"attribute", fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			expectedOutput: fmt.Sprintf("{\"address\":\"%s\",\"restricted\":false}", s.accountAddr.String()),
		},
		{
			name: "query name, text output",
			args: []string{"attribute", fmt.Sprintf("--%s=text", tmcli.OutputFlag)},
			expectedOutput: fmt.Sprintf("address: %s\nrestricted: false", s.accountAddr.String()),
		},
		{
			name: "query name that does not exist, text output",
			args: []string{"doesnotexist", fmt.Sprintf("--%s=text", tmcli.OutputFlag)},
			expectedOutput: "",
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := namecli.ResolveNameCommand()
			clientCtx := s.testnet.Validators[0].ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedOutput, strings.TrimSpace(out.String()))
		})
	}
}

func (s *IntegrationTestSuite) TestReverseLookupCommand() {
	accountKey := secp256k1.GenPrivKeyFromSecret([]byte("nobindinginthisaccount"))
	addr, _ := sdk.AccAddressFromHexUnsafe(accountKey.PubKey().Address().String())
	testCases := []struct {
		name           string
		args           []string
		expectedOutput string
	}{
		{
			name: "query name, json output",
			args: []string{s.accountAddr.String(), fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			expectedOutput: "{\"name\":[\"example.attribute\",\"attribute\"],\"pagination\":{\"next_key\":null,\"total\":\"0\"}}",
		},
		{
			name: "query name, text output",
			args: []string{s.accountAddr.String(), fmt.Sprintf("--%s=text", tmcli.OutputFlag)},
			expectedOutput: "name:\n- example.attribute\n- attribute\npagination:\n  next_key: null\n  total: \"0\"",
		},
		{
			name: "query name that does not exist, text output",
			args: []string{addr.String(), fmt.Sprintf("--%s=text", tmcli.OutputFlag)},
			expectedOutput: "name: []\npagination:\n  next_key: null\n  total: \"0\"",
		},
		{
			name: "query name that does not exist, json output",
			args: []string{addr.String(), fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			expectedOutput: "{\"name\":[],\"pagination\":{\"next_key\":null,\"total\":\"0\"}}",
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := namecli.ReverseLookupCommand()
			clientCtx := s.testnet.Validators[0].ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedOutput, strings.TrimSpace(out.String()))
		})
	}
}

func (s *IntegrationTestSuite) TestGetBindNameCommand() {

	testCases := []struct {
		name         string
		cmd          *cobra.Command
		args         []string
		expectErr    bool
		respType     proto.Message
		expectedCode uint32
	}{
		{
			name: "should bind name to root name",
			cmd: namecli.GetBindNameCmd(),
			args: []string{"bindnew", s.testnet.Validators[0].Address.String(), "attribute",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr: false,respType:  &sdk.TxResponse{},expectedCode:  0,
		},
		{
			name: "should fail to bind name to empty root name",
			cmd: namecli.GetBindNameCmd(),
			args: []string{"bindnew", s.testnet.Validators[0].Address.String(), "",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr: true,respType:  &sdk.TxResponse{},expectedCode:  1,
		},
		{
			name: "should fail to bind name to root name that does exist",
			cmd: namecli.GetBindNameCmd(),
			args: []string{"bindnew", s.testnet.Validators[0].Address.String(), "dne",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr: false,respType:  &sdk.TxResponse{},expectedCode:  18,
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
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestGetDeleteNameCmd() {

	testCases := []struct {
		name         string
		cmd          *cobra.Command
		args         []string
		expectErr    bool
		respType     proto.Message
		expectedCode uint32
	}{
		{
		name: 	"bind name for deletion",
		cmd: 	namecli.GetBindNameCmd(),
		args: 	[]string{"todelete", s.testnet.Validators[0].Address.String(), "attribute",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr: false,respType:  &sdk.TxResponse{}, expectedCode: 0,
		},
		{
			name: "should delete name",
			cmd: namecli.GetDeleteNameCmd(),
			args: []string{"todelete.attribute",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr: false, respType: &sdk.TxResponse{},expectedCode:  0,
		},
		{
			name: "should fail to delete name that does exist",
			cmd: namecli.GetDeleteNameCmd(),
			args: []string{"dne",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr: false, respType: &sdk.TxResponse{},expectedCode:  18,
		},
		{
			name: "should fail to delete name, not authorized",
			cmd: namecli.GetDeleteNameCmd(),
			args: []string{"example.attribute",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr: false, respType: &sdk.TxResponse{},expectedCode:  4,
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
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestGetModifyNameCmd() {

	testCases := []struct {
		name         string
		cmd          *cobra.Command
		args         []string
		expectErr    bool
		respType     proto.Message
		expectedCode uint32
	}{
		{
			name: "bind name for modification",
			cmd: namecli.GetBindNameCmd(),
			args: []string{"tomodify", s.testnet.Validators[0].Address.String(), "attribute",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr: false,respType:  &sdk.TxResponse{},expectedCode:  0,
		},
		{
			name: "should modify name",
			cmd: namecli.GetModifyNameProposalCmd(),
			[]string{"tomodify",
			args: 	s.testnet.Validators[0].Address.String(),
				"--unrestrict",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr: false,respType:  &sdk.TxResponse{},expectedCode:  0,
		},
		{
			name: "should fail to delete empty name",
			cmd: namecli.GetModifyNameProposalCmd(),
			args: []string{"",
				s.testnet.Validators[0].Address.String(),
				"--unrestrict",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr: true, respType: &sdk.TxResponse{}, expectedCode: 0,
		},
		{
			name: "should fail on invalid owner",
			cmd: namecli.GetModifyNameProposalCmd(),
			args: []string{"tomodify",
				"",
				"--unrestrict",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr: true,respType:  &sdk.TxResponse{},expectedCode:  0,
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
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestPaginationWithPageKey() {
	asJson := fmt.Sprintf("--%s=json", tmcli.OutputFlag)

	s.T().Run("ReverseLookupCommand", func(t *testing.T) {
		// Choosing page size = 13 because it a) isn't the default, b) doesn't evenly divide 50.
		pageSize := 13
		expectedCount := s.acc2NameCount
		pageCount := expectedCount / pageSize
		if expectedCount%pageSize != 0 {
			pageCount++
		}
		pageSizeArg := limitArg(pageSize)

		results := make([]string, 0, expectedCount)
		var nextKey string

		// Only using the page variable here for error messages, not for the CLI args since that'll mess with the --page-key being tested.
		for page := 1; page <= pageCount; page++ {
			args := []string{s.account2Addr.String(), pageSizeArg, asJson}
			if page != 1 {
				args = append(args, pageKeyArg(nextKey))
			}
			iterID := fmt.Sprintf("page %d/%d, args: %v", page, pageCount, args)
			cmd := namecli.ReverseLookupCommand()
			clientCtx := s.testnet.Validators[0].ClientCtx
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
			require.NoErrorf(t, err, "cmd error %s", iterID)
			var result nametypes.QueryReverseLookupResponse
			merr := s.cfg.Codec.UnmarshalJSON(out.Bytes(), &result)
			require.NoErrorf(t, merr, "unmarshal error %s", iterID)
			resultNameCount := len(result.Name)
			if page != pageCount {
				require.Equalf(t, pageSize, resultNameCount, "page result count %s", iterID)
				require.NotEmptyf(t, result.Pagination.NextKey, "pagination next key %s", iterID)
			} else {
				require.GreaterOrEqualf(t, pageSize, resultNameCount, "last page result count %s", iterID)
				require.Emptyf(t, result.Pagination.NextKey, "pagination next key %s", iterID)
			}
			results = append(results, result.Name...)
			nextKey = base64.StdEncoding.EncodeToString(result.Pagination.NextKey)
		}

		// This can fail if the --page-key isn't encoded/decoded correctly resulting in an unexpected jump forward in the actual list.
		require.Equal(t, expectedCount, len(results), "total count of names returned")
		// Make sure none of the results are duplicates.
		// That can happen if the --page-key isn't encoded/decoded correctly resulting in an unexpected jump backward in the actual list.
		sort.Strings(results)
		for i := 1; i < len(results); i++ {
			require.NotEqual(t, results[i-1], results[i], "no two names should be equal here")
		}
	})
}

func (s *IntegrationTestSuite) TestCreateRootNameCmd() {

	testCases := []struct {
		name         string
		cmd          *cobra.Command
		args         []string
		expectErr    bool
		respType     proto.Message
		expectedCode uint32
	}{
		{
		name: 	"should create a root name proposal",
		cmd: 	namecli.GetRootNameProposalCmd(),
		args: 	[]string{"rootprop",
				fmt.Sprintf("--%s=%s", namecli.FlagTitle, "title"),
				fmt.Sprintf("--%s=%s", namecli.FlagDescription, "description"),
				fmt.Sprintf("--%s=%s%s", namecli.FlagDeposit, "10", s.cfg.BondDenom),
				fmt.Sprintf("--%s=%s", "owner", s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr: false,respType:  &sdk.TxResponse{},expectedCode:  0,
		},
		{
			name: "should succeed with missing deposit",
			cmd: namecli.GetRootNameProposalCmd(),
			args: []string{"rootprop",
				fmt.Sprintf("--%s=%s", namecli.FlagTitle, "title"),
				fmt.Sprintf("--%s=%s", namecli.FlagDescription, "description"),
				fmt.Sprintf("--%s=%s", "owner", s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr: false,respType:  &sdk.TxResponse{},expectedCode:  0,
		},
		{
			name: "should fail for bad deposit",
			cmd: namecli.GetRootNameProposalCmd(),
			args: []string{"rootprop",
				fmt.Sprintf("--%s=%s", namecli.FlagTitle, "title"),
				fmt.Sprintf("--%s=%s", namecli.FlagDescription, "description"),
				fmt.Sprintf("--%s=%s", namecli.FlagDeposit, "10"),
				fmt.Sprintf("--%s=%s", "owner", s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr: true, respType: &sdk.TxResponse{},expectedCode:  1,
		},
		{
			name: "should fail for empty title",
			cmd: namecli.GetRootNameProposalCmd(),
			args: []string{"rootprop",
				fmt.Sprintf("--%s=%s", namecli.FlagDescription, "description"),
				fmt.Sprintf("--%s=%s%s", namecli.FlagDeposit, "10", s.cfg.BondDenom),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr: true,respType:  &sdk.TxResponse{},expectedCode:  1,
		},
		{
			name: "should fail for empty description",
			cmd: namecli.GetRootNameProposalCmd(),
			args: []string{"rootprop",
				fmt.Sprintf("--%s=%s", namecli.FlagTitle, "title"),
				fmt.Sprintf("--%s=%s%s", namecli.FlagDeposit, "10", s.cfg.BondDenom),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr: true,respType:  &sdk.TxResponse{},expectedCode:  1,
		},
		{
			name: "should fail for bad owner",
			cmd: namecli.GetRootNameProposalCmd(),
			args: []string{"rootprop",
				fmt.Sprintf("--%s=%s", namecli.FlagTitle, "title"),
				fmt.Sprintf("--%s=%s%s", namecli.FlagDeposit, "10", s.cfg.BondDenom),

				fmt.Sprintf("--%s=%s", "owner", "asdf"),

				fmt.Sprintf("--%s=%s", flags.FlagFrom, s.testnet.Validators[0].Address.String()),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
				fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
			},
			expectErr: true,respType:  &sdk.TxResponse{},expectedCode:  1,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			clientCtx := s.testnet.Validators[0].ClientCtx
			// because the cmd runs inside of the gov cmd (which adds flags) we register here so we can use it directly.
			flags.AddTxFlagsToCmd(tc.cmd)
			out, err := clitestutil.ExecTestCLICmd(clientCtx, tc.cmd, tc.args)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code)
			}
		})
	}
}
