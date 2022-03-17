package cli_test

import (
	"fmt"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	"github.com/gogo/protobuf/proto"
	"github.com/provenance-io/provenance/testutil"
	"github.com/provenance-io/provenance/x/epoch/cli"
	"github.com/provenance-io/provenance/x/epoch/types"
	"github.com/stretchr/testify/suite"
	tmcli "github.com/tendermint/tendermint/libs/cli"
	"testing"
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

func (s *IntegrationTestSuite) TestGetCmdCurrentEpoch() {
	val := s.network.Validators[0]

	testCases := []struct {
		name       string
		identifier string
		expectErr  bool
		respType   proto.Message
	}{
		{
			"query weekly epoch number",
			"weekly",
			false, &types.QueryCurrentEpochResponse{},
		},
		{
			"query unavailable epoch number",
			"unavailable",
			false, &types.QueryCurrentEpochResponse{},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdCurrentEpoch()
			clientCtx := val.ClientCtx

			args := []string{
				tc.identifier,
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			}

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.JSONCodec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestGetCmdEpochsInfos() {
	val := s.network.Validators[0]

	testCases := []struct {
		name      string
		expectErr bool
		respType  proto.Message
	}{
		{
			"query epoch infos",
			false, &types.QueryEpochsInfoResponse{},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetCmdEpochsInfos()
			clientCtx := val.ClientCtx

			args := []string{
				fmt.Sprintf("--%s=json", tmcli.OutputFlag),
			}

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, args)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err, out.String())
				//s.Require().NoError(clientCtx.JSONCodec.UnmarshalJSON(out.Bytes(), &sdk.TxResponse{}), out.String())
				s.Require().NoError(clientCtx.JSONCodec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
			}
		})
	}
}
