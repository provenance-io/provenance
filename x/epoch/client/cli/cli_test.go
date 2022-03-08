package cli_test

import (
	"strings"
	"testing"

	"github.com/provenance-io/provenance/internal/antewrapper"
	"github.com/provenance-io/provenance/testutil"
	"github.com/provenance-io/provenance/x/epoch/client/cli"

	"github.com/stretchr/testify/suite"

	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	testnet "github.com/cosmos/cosmos-sdk/testutil/network"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg     testnet.Config
	testnet *testnet.Network
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupSuite() {

	s.T().Log("setting up integration test suite")

	cfg := testutil.DefaultTestNetworkConfig()
	cfg.NumValidators = 1

	s.cfg = cfg
	cfg.ChainID = antewrapper.SimAppChainID
	s.testnet = testnet.New(s.T(), cfg)

	_, err := s.testnet.WaitForHeight(1)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.testnet.WaitForNextBlock()
	s.T().Log("tearing down integration test suite")
	s.testnet.Cleanup()
}

func (s *IntegrationTestSuite) TestEpochInfosCmd() {

	cmd := cli.EpochInfosCmd()
	clientCtx := s.testnet.Validators[0].ClientCtx

	out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, "")
	s.Require().NoError(err)
	s.Require().Equal("", strings.TrimSpace(out.String()))
}

func (s *IntegrationTestSuite) TestCurrentEpochCmd() {

	testCases := []struct {
		name           string
		identifier     []string
		expectedOutput string
	}{
		{
			"query by identifier, should succeed",
			[]string{"day"},
			"",
		},
		{
			"query by identifier, should fail to find by identifier",
			[]string{"dne"},
			"",
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.CurrentEpochCmd()
			clientCtx := s.testnet.Validators[0].ClientCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.identifier)
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedOutput, strings.TrimSpace(out.String()))
		})
	}
}
