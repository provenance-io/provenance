package testutil

import (
	"fmt"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/testutil/assertions"
	"github.com/provenance-io/provenance/testutil/queries"
	"github.com/provenance-io/provenance/x/sanction"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg       network.Config
	network   *network.Network
	clientCtx client.Context

	commonArgs []string
	val0       *network.Validator
	valAddr    sdk.AccAddress
	authority  string

	sanctionGenesis *sanction.GenesisState
}

func NewIntegrationTestSuite(cfg network.Config, sanctionGenesis *sanction.GenesisState) *IntegrationTestSuite {
	return &IntegrationTestSuite{
		cfg:             cfg,
		sanctionGenesis: sanctionGenesis,
	}
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	var err error
	s.network, err = network.New(s.T(), s.T().TempDir(), s.cfg)
	s.Require().NoError(err)

	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)

	s.val0 = s.network.Validators[0]
	s.clientCtx = s.network.Validators[0].ClientCtx
	s.valAddr = s.network.Validators[0].Address

	s.commonArgs = []string{
		fmt.Sprintf("--%s", flags.FlagFrom), s.valAddr.String(),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, s.bondCoins(10).String()),
	}

}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

// assertErrorContents calls AssertErrorContents using this suite's t.
func (s *IntegrationTestSuite) assertErrorContents(theError error, contains []string, msgAndArgs ...interface{}) bool {
	return assertions.AssertErrorContents(s.T(), theError, contains, msgAndArgs...)
}

// bondCoin creates an sdk.Coin with the bond-denom in the amount provided.
func (s *IntegrationTestSuite) bondCoin(amt int64) sdk.Coin {
	return sdk.NewInt64Coin(s.cfg.BondDenom, amt)
}

// bondCoins creates an sdk.Coins with the bond-denom in the amount provided.
func (s *IntegrationTestSuite) bondCoins(amt int64) sdk.Coins {
	return sdk.NewCoins(s.bondCoin(amt))
}

// appendCommonFlagsTo adds this suite's common flags to the end of the provided arguments.
func (s *IntegrationTestSuite) appendCommonArgsTo(args ...string) []string {
	return append(args, s.commonArgs...)
}

// getAuthority executes a query to get the address of the gov module account.
func (s *IntegrationTestSuite) getAuthority() string {
	acct := queries.GetModuleAccountByName(s.T(), s.val0, "gov")
	return acct.GetAddress().String()
}
