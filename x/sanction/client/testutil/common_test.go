package testutil

import (
	"fmt"

	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gov "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	"github.com/provenance-io/provenance/internal/pioconfig"
	"github.com/provenance-io/provenance/testutil"
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
	valAddr    sdk.AccAddress

	sanctionGenesis *sanction.GenesisState
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")
	pioconfig.SetProvenanceConfig(sdk.DefaultBondDenom, 0)
	govv1.DefaultMinDepositRatio = sdkmath.LegacyZeroDec()
	s.cfg = testutil.DefaultTestNetworkConfig()
	s.cfg.NumValidators = 5 // enough for voting and maybe sanctioning one.

	// Define some stuff in the santion genesis state.
	testutil.MutateGenesisState(s.T(), &s.cfg, sanction.ModuleName, &sanction.GenesisState{}, func(sanctionGen *sanction.GenesisState) *sanction.GenesisState {
		sanctionedAddr1 := sdk.AccAddress("1_sanctioned_address_")
		sanctionedAddr2 := sdk.AccAddress("2_sanctioned_address_")
		tempSanctAddr := sdk.AccAddress("temp_sanctioned_addr")
		tempUnsanctAddr := sdk.AccAddress("temp_unsanctioned___")

		sanctionGen.SanctionedAddresses = append(sanctionGen.SanctionedAddresses,
			sanctionedAddr1.String(),
			sanctionedAddr2.String(),
		)
		sanctionGen.TemporaryEntries = append(sanctionGen.TemporaryEntries,
			&sanction.TemporaryEntry{
				Address:    tempSanctAddr.String(),
				ProposalId: 1,
				Status:     sanction.TEMP_STATUS_SANCTIONED,
			},
			&sanction.TemporaryEntry{
				Address:    tempUnsanctAddr.String(),
				ProposalId: 1,
				Status:     sanction.TEMP_STATUS_UNSANCTIONED,
			},
		)
		sanctionGen.Params = &sanction.Params{
			ImmediateSanctionMinDeposit:   sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 52)),
			ImmediateUnsanctionMinDeposit: sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 133)),
		}

		s.sanctionGenesis = sanctionGen
		return sanctionGen
	})

	// Tweak the gov params too to make testing gov props easier.
	testutil.MutateGenesisState(s.T(), &s.cfg, gov.ModuleName, &govv1.GenesisState{}, func(govGen *govv1.GenesisState) *govv1.GenesisState {
		govGen.Params.MinDeposit = sdk.NewCoins(sdk.NewInt64Coin(s.cfg.BondDenom, 6)) // default is 10000000stake
		votingPeriod := s.cfg.TimeoutCommit * blocksPerVotingPeriod
		govGen.Params.MaxDepositPeriod = &votingPeriod // default is 48h
		govGen.Params.VotingPeriod = &votingPeriod     // default is 48h
		return govGen
	})

	var err error
	s.network, err = network.New(s.T(), s.T().TempDir(), s.cfg)
	s.Require().NoError(err, "network.New(...)")

	s.waitForHeight(1)

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
	testutil.Cleanup(s.network, s.T())
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
	acct := queries.GetModuleAccountByName(s.T(), s.network, "gov")
	return acct.GetAddress().String()
}

// logHeight outputs the current height to the test log.
func (s *IntegrationTestSuite) logHeight() int64 {
	height, err := testutil.LatestHeight(s.network)
	s.Require().NoError(err, "LatestHeight()")
	s.T().Logf("Current height: %d", height)
	return height
}

// waitForHeight waits for the requested height, logging the current height once we get there.
func (s *IntegrationTestSuite) waitForHeight(height int64) int64 {
	rv, err := testutil.WaitForHeight(s.network, height)
	s.Require().NoError(err, "WaitForHeight(%d)", height)
	s.T().Logf("Current height: %d", rv)
	return rv
}

// waitForNextBlock waits for the current height to be finished.
func (s *IntegrationTestSuite) waitForNextBlock(msgAndArgs ...interface{}) {
	if len(msgAndArgs) == 0 {
		msgAndArgs = append(msgAndArgs, "WaitForNextBlock")
	}
	s.Require().NoError(testutil.WaitForNextBlock(s.network), msgAndArgs...)
}
