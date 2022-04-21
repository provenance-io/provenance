package cli_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/testutil/network"
	"github.com/stretchr/testify/suite"

	"github.com/provenance-io/provenance/testutil"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	s.cfg = testutil.DefaultTestNetworkConfig()

	s.network = network.New(s.T(), s.cfg)

	_, err := s.network.WaitForHeight(1)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func (s *IntegrationTestSuite) TestQueryRewardPrograms() {
	s.Assert().FailNow("not implemented")
	//cli.QueryRewardProgramsCmd()
	// TODO Need a way to create a reward program before these can be implemented
}

func (s *IntegrationTestSuite) TestQueryRewardProgramsById() {
	s.Assert().FailNow("not implemented")
	//cli.RewardProgramByIdCmd()
	// TODO Need a way to create a reward program before these can be implemented
}

func (s *IntegrationTestSuite) TestQueryRewardClaims() {
	s.Assert().FailNow("not implemented")
	//cli.QueryRewardProgramsCmd()
	// TODO Need a way to create a reward claim before these can be implemented
}

func (s *IntegrationTestSuite) TestQueryRewardClaimsById() {
	s.Assert().FailNow("not implemented")
	//cli.QueryRewardProgramsCmd()
	// TODO Need a way to create a reward claim before these can be implemented
}
