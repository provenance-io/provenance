package cli_test

import (
	"strings"
	"testing"
	"time"

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

func CleanUp(n *testnet.Network, t *testing.T) {
	t.Log("teardown waiting for next block")
	//nolint:errcheck // The test shouldn't fail because cleanup was a problem. So ignoring any error from this.
	n.WaitForNextBlock()
	t.Log("teardown cleaning up testnet")
	n.Cleanup()
	// Give things a chance to finish closing up. Hopefully will prevent things like address collisions. 100ms chosen randomly.
	time.Sleep(100 * time.Millisecond)
	t.Log("teardown done")
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.testnet.WaitForNextBlock()
	s.T().Log("tearing down integration test suite")
	s.testnet.Cleanup()
	time.Sleep(100 * time.Millisecond)
	s.T().Log("teardown done")
}

func (s *IntegrationTestSuite) TestEpochInfosCmd() {

	cmd := cli.EpochInfosCmd()
	clientCtx := s.testnet.Validators[0].ClientCtx

	out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, []string{})
	s.Require().NoError(err)
	s.Require().Equal("epochs:\n- current_epoch: \"1\"\n  current_epoch_start_height: \"1\"\n  duration: \"17280\"\n  epoch_counting_started: true\n  identifier: day\n  start_height: \"0\"\n- current_epoch: \"1\"\n  current_epoch_start_height: \"1\"\n  duration: \"518400\"\n  epoch_counting_started: true\n  identifier: month\n  start_height: \"0\"\n- current_epoch: \"1\"\n  current_epoch_start_height: \"1\"\n  duration: \"120960\"\n  epoch_counting_started: true\n  identifier: week\n  start_height: \"0\"", strings.TrimSpace(out.String()))
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
			"current_epoch: \"1\"",
		},
		{
			"query by identifier, should fail to find by identifier",
			[]string{"dne"},
			"current_epoch: \"0\"",
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
