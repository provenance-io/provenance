package app

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/suite"

	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	msgfeetypes "github.com/provenance-io/provenance/x/msgfees/types"
)

type UpgradeTestSuite struct {
	suite.Suite

	app *App
	ctx sdk.Context

	logBuffer bytes.Buffer

	startTime time.Time
}

func TestUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (s *UpgradeTestSuite) SetupSuite() {
	// Alert: This function is SetupSuite. That means all tests in here
	// will use the same app with the same store and data.
	bufferedLoggerMaker := func() log.Logger {
		lw := zerolog.ConsoleWriter{
			Out:          &s.logBuffer,
			NoColor:      true,
			PartsExclude: []string{"time"}, // Without this, each line starts with "<nil> "
		}
		// Error log lines will start with "ERR ".
		// Info log lines will start with "INF ".
		// Debug log lines are omitted, but would start with "DBG ".
		logger := zerolog.New(lw).Level(zerolog.InfoLevel)
		return server.ZeroLogWrapper{Logger: logger}
	}
	defer SetLoggerMaker(SetLoggerMaker(bufferedLoggerMaker))
	s.app = Setup(s.T())
	s.logBuffer.Reset()
	s.startTime = time.Now()
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{Time: s.startTime})
}

// AssertUpgradeHandlerLogs runs the Handler of the provided key and asserts that
// each entry in expInLog is in the log output, and each entry in expNotInLog
// is not in the log output.
//
// entries in expInLog are expected to be full lines.
// entries in expNotInLog can be parts of lines.
//
// This returns the log output and whether all assertions passed (true = all passed,
// false = one or more problems were found).
func (s *UpgradeTestSuite) AssertUpgradeHandlerLogs(key string, expInLog, expNotInLog []string) (logOutput string, allPassed bool) {
	s.T().Helper()
	defer func() {
		allPassed = !s.T().Failed()
	}()

	if !s.Assert().Contains(upgrades, key, "defined upgrades map") {
		return // If the upgrades map doesn't have that key, there's nothing more to do in here.
	}
	handler := upgrades[key].Handler
	if !s.Assert().NotNil(handler, "upgrades[%q].Handler", key) {
		return // If the entry doesn't have a .Handler, there's nothing more to do in here.
	}

	// This app was just created brand new, so it will create all the modules with their most
	// recent versions. As such, this GetModuleVersionMap will return all the most current
	// module versions info. That means that none of the module migrations will be running when we
	// call the handler, and that the log won't have any entries from any specific module migrations.
	// If you really want to include those in this run, you can try doing stuff with SetModuleVersionMap
	// prior to calling this AssertUpgradeHandlerLogs function. You'll probably also need to recreate
	// any old state for that module since anything added already was done using new stuff.
	origVersionMap := s.app.UpgradeKeeper.GetModuleVersionMap(s.ctx)

	var versionMap module.VersionMap
	var err error
	s.logBuffer.Reset()
	testFunc := func() {
		versionMap, err = handler(s.ctx, s.app, origVersionMap)
	}
	if !s.Assert().NotPanics(testFunc, "upgrades[%q].Handler(...)", key) {
		return // If the handler panics, there's nothing more to check in here.
	}

	// Add the handler log output to the test log for easier troubleshooting.
	logOutput = s.logBuffer.String()
	s.T().Logf("upgrades[%q].Handler(...) log output:\n%s", key, logOutput)

	// Checking for error cases should be done in individual function unit tests.
	// For this, always check for no error and a non-empty version map.
	s.Assert().NoError(err, "upgrades[%q].Handler(...) error", key)
	s.Assert().NotEmpty(versionMap, "upgrades[%q].Handler(...) version map", key)

	logLines := strings.Split(logOutput, "\n")
	for _, exp := range expInLog {
		// I'm including the expected in the msgAndArgs here even though it's also included in the
		// failure message because the failure message puts everything on one line. So the expected
		// string is almost at the end of a very long line. Having it in the "message" portion of
		// the failure makes it easier to identify the problematic entry.
		s.Assert().Contains(logLines, exp, "upgrades[%q].Handler(...) log output.\nExpecting: %q", key, exp)
	}

	for _, unexp := range expNotInLog {
		// Same here with the msgAndArgs thing.
		s.Assert().NotContains(logOutput, unexp, "upgrades[%q].Handler(...) log output.\nNot Expecting: %q", key, unexp)
	}

	return
}

func (s *UpgradeTestSuite) TestKeysInHandlersMap() {
	// handlerKeys are all the keys in the upgrades map defined in upgrades.go.
	handlerKeys := make([]string, 0, len(upgrades))
	// colors are the different colors used in the upgrades map.
	colors := make([]string, 0, 2)
	// rcs is a map of "<color>" to "rc<number>" suffixes for that color.
	rcs := make(map[string][]string)

	// addColor adds an entry to colors if it isn't already in there.
	addColor := func(newColor string) {
		for _, color := range colors {
			if newColor == color {
				return
			}
		}
		colors = append(colors, newColor)
	}

	// get all the upgrade handler keys and identify the unique colors and their release candidates.
	for key := range upgrades {
		handlerKeys = append(handlerKeys, key)
		parts := strings.SplitN(key, "-", 2)
		addColor(parts[0])
		if len(parts) > 1 && len(parts[1]) > 0 {
			rcs[parts[0]] = append(rcs[parts[0]], parts[1])
		}
	}

	// Sort all that stuff so it's easier to read in output and easier to test against.
	sort.Strings(handlerKeys)
	sort.Strings(colors)
	for _, v := range rcs {
		sort.Strings(v)
	}

	s.Run("upgrade key format", func() {
		// Upgrade keys should either be "<color>" or have the format "<color>-rc<number>".
		// Non-letters should be removed from the color to form one word, e.g. "neoncarrot"
		// Only lowercase letters should be used.
		rx := regexp.MustCompile(`^[a-z]+(-rc[0-9]+)?$`)
		for _, upgradeKey := range handlerKeys {
			s.Assert().Regexp(rx, upgradeKey)
		}
	})

	s.Run("two or more colors exist", func() {
		// We always need the colors currently in use on mainnet and testnet.
		// The ones before that shouldn't be removed until we add new ones.
		// It's okay to not clean the old ones up immediately, though.
		// So we always want at least 2 different colors in there.
		s.Assert().GreaterOrEqual(len(colors), 2, "number of distinct colors: %q in %q", colors, handlerKeys)
		// If there are more than 3, we need to do some cleanup though.
		s.Assert().LessOrEqual(len(colors), 3, "number of distinct colors: %q in %q", colors, handlerKeys)
	})

	s.Run("no two colors start with same character", func() {
		// Little tricky here. i will go from 0 to len(colors) - 2 and the color will go from the 2nd to last.
		// So colors[i] in here will be the entry just before color.
		for i, color := range colors[1:] {
			s.Assert().NotEqual(string(colors[i][0]), string(color[0]), "first letters of colors %q and %q", colors[i], color)
		}
	})

	s.Run("rc exists for each color", func() {
		// We shouldn't delete any entries (rc or non) until all of the entries for that color can be deleted.
		// This helps maintain a complete picture of the upgrades involved with a version.
		for _, color := range colors {
			s.Assert().NotEmpty(rcs[color], "rc entries for %s in %q", color, handlerKeys)
		}
	})

	s.Run("rcs all exist sequentially", func() {
		// All of the rc entries should be present starting with 1 and increasing by 1 each time.
		// I.e. we shouldn't skip a number, and old rc entries shouldn't be deleted until all
		// entries for that color can be deleted.
		// The rc numbers for a color won't necessarily align with the software version rc numbers.
		// E.g. quicksilver-rc1 was for v1.15.0-rc2.

		// Make a slice of strings rc1, rc2, ... rc11.
		// If we end up with something larger than rc11, dry your eyes and bump this.
		expRCs := make([]string, 10)
		for i := range expRCs {
			expRCs[i] = fmt.Sprintf("rc%d", i+1)
		}
		for _, color := range colors {
			if len(rcs[color]) > 0 {
				exp := expRCs[:len(rcs[color])]
				s.Assert().Equal(exp, rcs[color], "rc suffixes for %s in %q", color, handlerKeys)
			}
		}
	})

	s.Run("non rc exists for each color", func() {
		// If an rc entry for a color exists, the non-rc entry needs to exist too.
		// That way, the mainnet upgrade handler is harder to
		// forget about as the rc entries are being written.
		for _, color := range colors {
			s.Assert().Contains(handlerKeys, color, "non-rc entry for %s in %q", color, handlerKeys)
		}
	})
}

func (s *UpgradeTestSuite) TestRustRC1() {
	// Each part is (hopefully) tested thoroughly on its own.
	// So for this test, just make sure there's log entries for each part being done.

	expInLog := []string{
		"INF Starting module migrations. This may take a significant amount of time to complete. Do not restart node.",
		`INF Creating message fee for "/cosmos.gov.v1.MsgSubmitProposal" if it doesn't already exist.`,
		`INF Removing message fee for "/provenance.metadata.v1.MsgP8eMemorializeContractRequest" if one exists.`,
		"INF Fixing name module store index entries.",
	}

	s.AssertUpgradeHandlerLogs("rust-rc1", expInLog, nil)
}

func (s *UpgradeTestSuite) TestRust() {
	// Each part is (hopefully) tested thoroughly on its own.
	// So for this test, just make sure there's log entries for each part being done.
	// And not a log entry for stuff done in rust-rc1 but not this one.

	expInLog := []string{
		"INF Starting module migrations. This may take a significant amount of time to complete. Do not restart node.",
		`INF Removing message fee for "/provenance.metadata.v1.MsgP8eMemorializeContractRequest" if one exists.`,
		"INF Fixing name module store index entries.",
	}
	expNotInLog := []string{
		`Creating message fee for "/cosmos.gov.v1.MsgSubmitProposal" if it doesn't already exist.`,
	}

	s.AssertUpgradeHandlerLogs("rust", expInLog, expNotInLog)
}

func (s *UpgradeTestSuite) TestAddGovV1SubmitFee() {
	v1TypeURL := "/cosmos.gov.v1.MsgSubmitProposal"
	v1B1TypeURL := "/cosmos.gov.v1beta1.MsgSubmitProposal"

	startingMsg := `Creating message fee for "` + v1TypeURL + `" if it doesn't already exist.`
	successMsg := func(amt string) string {
		return `Successfully set fee for "` + v1TypeURL + `" with amount "` + amt + `".`
	}

	coin := func(denom string, amt int64) *sdk.Coin {
		rv := sdk.NewInt64Coin(denom, amt)
		return &rv
	}

	tests := []struct {
		name     string
		v1Amt    *sdk.Coin
		v1B1Amt  *sdk.Coin
		expInLog []string
		expAmt   sdk.Coin
	}{
		{
			name:    "v1 fee already exists",
			v1Amt:   coin("foocoin", 88),
			v1B1Amt: coin("betacoin", 99),
			expInLog: []string{
				startingMsg,
				`Message fee for "` + v1TypeURL + `" already exists with amount "88foocoin". Nothing to do.`,
			},
			expAmt: *coin("foocoin", 88),
		},
		{
			name:    "v1beta1 exists",
			v1B1Amt: coin("betacoin", 99),
			expInLog: []string{
				startingMsg,
				`Copying "` + v1B1TypeURL + `" fee to "` + v1TypeURL + `".`,
				successMsg("99betacoin"),
			},
			expAmt: *coin("betacoin", 99),
		},
		{
			name: "brand new",
			expInLog: []string{
				startingMsg,
				`Creating "` + v1TypeURL + `" fee.`,
				successMsg("100000000000nhash"),
			},
			expAmt: *coin("nhash", 100_000_000_000),
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			// Set/unset the v1 fee.
			if tc.v1Amt != nil {
				fee := msgfeetypes.NewMsgFee(v1TypeURL, *tc.v1Amt, "", 0)
				s.Require().NoError(s.app.MsgFeesKeeper.SetMsgFee(s.ctx, fee), "SetMsgFee v1")
			} else {
				err := s.app.MsgFeesKeeper.RemoveMsgFee(s.ctx, v1TypeURL)
				if err != nil && !errors.Is(err, msgfeetypes.ErrMsgFeeDoesNotExist) {
					s.Require().NoError(err, "RemoveMsgFee v1")
				}
			}

			// Set/unset the v1beta1 fee.
			if tc.v1B1Amt != nil {
				fee := msgfeetypes.NewMsgFee(v1B1TypeURL, *tc.v1B1Amt, "", 0)
				s.Require().NoError(s.app.MsgFeesKeeper.SetMsgFee(s.ctx, fee), "SetMsgFee v1beta1")
			} else {
				err := s.app.MsgFeesKeeper.RemoveMsgFee(s.ctx, v1B1TypeURL)
				if err != nil && !errors.Is(err, msgfeetypes.ErrMsgFeeDoesNotExist) {
					s.Require().NoError(err, "RemoveMsgFee v1")
				}
			}

			// Reset the log buffer to clear out unrelated entries.
			s.logBuffer.Reset()
			// Call addGovV1SubmitFee and relog its output (to help if things fail).
			testFunc := func() {
				addGovV1SubmitFee(s.ctx, s.app)
			}
			s.Require().NotPanics(testFunc, "addGovV1SubmitFee")
			logOutput := s.logBuffer.String()
			s.T().Logf("addGovV1SubmitFee log output:\n%s", logOutput)

			// Make sure the log has the expected lines.
			for _, exp := range tc.expInLog {
				s.Assert().Contains(logOutput, exp, "addGovV1SubmitFee log output")
			}

			// Get the fee and make sure it's now as expected.
			fee, err := s.app.MsgFeesKeeper.GetMsgFee(s.ctx, v1TypeURL)
			s.Require().NoError(err, "GetMsgFee(%q) error", v1TypeURL)
			s.Require().NotNil(fee, "GetMsgFee(%q) value", v1TypeURL)
			actFeeAmt := fee.AdditionalFee
			s.Assert().Equal(tc.expAmt.String(), actFeeAmt.String(), "final %s fee amount", v1TypeURL)
		})
	}
}

func (s *UpgradeTestSuite) TestRemoveP8eMemorializeContractFee() {
	typeURL := "/provenance.metadata.v1.MsgP8eMemorializeContractRequest"
	startingMsg := `Removing message fee for "` + typeURL + `" if one exists.`

	coin := func(denom string, amt int64) *sdk.Coin {
		rv := sdk.NewInt64Coin(denom, amt)
		return &rv
	}

	tests := []struct {
		name     string
		amt      *sdk.Coin
		expInLog []string
	}{
		{
			name: "does not exist",
			expInLog: []string{
				startingMsg,
				`Message fee for "` + typeURL + `" already does not exist. Nothing to do.`,
			},
		},
		{
			name: "exists",
			amt:  coin("p8ecoin", 808),
			expInLog: []string{
				startingMsg,
				`Successfully removed message fee for "` + typeURL + `" with amount "808p8ecoin".`,
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			// set/unset the fee
			if tc.amt != nil {
				fee := msgfeetypes.NewMsgFee(typeURL, *tc.amt, "", 0)
				s.Require().NoError(s.app.MsgFeesKeeper.SetMsgFee(s.ctx, fee), "Setup: SetMsgFee(%q)", typeURL)
			} else {
				err := s.app.MsgFeesKeeper.RemoveMsgFee(s.ctx, typeURL)
				if err != nil && !errors.Is(err, msgfeetypes.ErrMsgFeeDoesNotExist) {
					s.Require().NoError(err, "Setup: RemoveMsgFee(%q)", typeURL)
				}
			}

			s.logBuffer.Reset()
			testFunc := func() {
				removeP8eMemorializeContractFee(s.ctx, s.app)
			}
			s.Require().NotPanics(testFunc, "removeP8eMemorializeContractFee")
			logOutput := s.logBuffer.String()
			s.T().Logf("removeP8eMemorializeContractFee log output:\n%s", logOutput)

			for _, exp := range tc.expInLog {
				s.Assert().Contains(logOutput, exp, "removeP8eMemorializeContractFee log output")
			}

			// Make sure there isn't a fee anymore.
			fee, err := s.app.MsgFeesKeeper.GetMsgFee(s.ctx, typeURL)
			s.Require().NoError(err, "GetMsgFee(%q) error", typeURL)
			s.Require().Nil(fee, "GetMsgFee(%q) value", typeURL)
		})
	}
}

func (s *UpgradeTestSuite) TestRemoveInactiveValidators() {
	addr1 := s.CreateAndFundAccount(sdk.NewInt64Coin("stake", 1000000))
	addr2 := s.CreateAndFundAccount(sdk.NewInt64Coin("stake", 1000000))

	//run with just one bonded validator
	validators := s.app.StakingKeeper.GetAllValidators(s.ctx)
	s.Require().Equal(1, len(validators), "should have one validator, original validator made on test setup")

	expectedLogLines := []string{
		"INF removing any validator that has been inactive (unbonded) for 21 hours",
		"INF a total of 0 inactive (unbonded) validators have been removed",
	}
	s.ExecuteAndAssertLogs(removeInactiveValidators, expectedLogLines)

	// create a unbonded validator with out delegators...this case should not happen irl (or from what I can see)
	unbondedVal1 := s.CreateValidator(s.startTime.Add(-30*24*time.Hour), stakingtypes.Unbonded)
	validators = s.app.StakingKeeper.GetAllValidators(s.ctx)
	s.Assert().Equal(2, len(validators), "should two validators, one bonded and one unbonded")

	expectedLogLines = []string{
		"INF removing any validator that has been inactive (unbonded) for 21 hours",
		fmt.Sprintf("INF validator %v has been inactive (unbonded) for %d days and will be removed", unbondedVal1.OperatorAddress, 30),
		"INF a total of 1 inactive (unbonded) validators have been removed",
	}
	s.ExecuteAndAssertLogs(removeInactiveValidators, expectedLogLines)

	validators = s.app.StakingKeeper.GetAllValidators(s.ctx)
	s.Assert().Equal(1, len(validators), "should have removed the unbonded delegator")

	// single unbonded validator with 1 delegations, should be removed
	coin := sdk.NewInt64Coin("stake", 10000)
	unbondedVal1 = s.CreateValidator(s.startTime.Add(-30*24*time.Hour), stakingtypes.Unbonded)
	addr1Balance := s.app.BankKeeper.GetBalance(s.ctx, addr1, "stake")
	s.DelegateToValidator(unbondedVal1.GetOperator(), sdk.AccAddress(addr1), coin)
	s.Require().Equal(addr1Balance.Sub(coin), s.app.BankKeeper.GetBalance(s.ctx, addr1, "stake"), "addr1 should have less funds")
	validators = s.app.StakingKeeper.GetAllValidators(s.ctx)
	s.Require().Equal(2, len(validators), "should two validators, one bonded and one unbonded")

	expectedLogLines = []string{
		"INF removing any validator that has been inactive (unbonded) for 21 hours",
		fmt.Sprintf("INF validator %v has been inactive (unbonded) for %d days and will be removed", unbondedVal1.OperatorAddress, 30),
		fmt.Sprintf("INF undelegate delegator %v from validator %v of all shares (%v)", addr1.String(), unbondedVal1.OperatorAddress, sdk.NewDec(coin.Amount.Int64())),
		"INF a total of 1 inactive (unbonded) validators have been removed",
	}
	s.ExecuteAndAssertLogs(removeInactiveValidators, expectedLogLines)

	validators = s.app.StakingKeeper.GetAllValidators(s.ctx)
	s.Assert().Equal(1, len(validators), "should have removed the unbonded delegator")
	ubd, found := s.app.StakingKeeper.GetUnbondingDelegation(s.ctx, addr1, unbondedVal1.GetOperator())
	s.Assert().True(found, "should have found addr1 unbonding delegation")
	s.Assert().Len(ubd.Entries, 1, "addr1 should have 1 unbonding entry")
	s.Assert().Equal(coin.Amount, ubd.Entries[0].Balance, "addr1 should have the total delegation amount in unbonding delegation")

	// single unbonded validator with 2 delegations
	unbondedVal1 = s.CreateValidator(s.startTime.Add(-30*24*time.Hour), stakingtypes.Unbonded)
	addr1Balance = s.app.BankKeeper.GetBalance(s.ctx, addr1, "stake")
	addr2Balance := s.app.BankKeeper.GetBalance(s.ctx, addr2, "stake")
	s.DelegateToValidator(unbondedVal1.GetOperator(), sdk.AccAddress(addr1), coin)
	s.DelegateToValidator(unbondedVal1.GetOperator(), sdk.AccAddress(addr2), coin)
	s.Require().Equal(addr1Balance.Sub(coin), s.app.BankKeeper.GetBalance(s.ctx, addr1, "stake"), "addr1 should have less funds after delegation")
	s.Require().Equal(addr2Balance.Sub(coin), s.app.BankKeeper.GetBalance(s.ctx, addr2, "stake"), "addr2 should have less funds after delegation")
	unbondedVal1, _ = s.app.StakingKeeper.GetValidator(s.ctx, unbondedVal1.GetOperator())
	s.Require().Equal(sdk.NewDec(20_000), unbondedVal1.DelegatorShares, "should have correct amount of shares from two delegations")
	validators = s.app.StakingKeeper.GetAllValidators(s.ctx)
	s.Require().Equal(2, len(validators), "should two validators, one bonded and one unbonded with delegations")

	expectedLogLines = []string{
		"INF removing any validator that has been inactive (unbonded) for 21 hours",
		fmt.Sprintf("INF validator %v has been inactive (unbonded) for %d days and will be removed", unbondedVal1.OperatorAddress, 30),
		"INF a total of 1 inactive (unbonded) validators have been removed"}
	s.ExecuteAndAssertLogs(removeInactiveValidators, expectedLogLines)

	validators = s.app.StakingKeeper.GetAllValidators(s.ctx)
	s.Assert().Equal(1, len(validators), "should have removed the unbonded delegator")
	ubd, found = s.app.StakingKeeper.GetUnbondingDelegation(s.ctx, addr1, unbondedVal1.GetOperator())
	s.Assert().True(found, "should have found addr1 unbonding delegation")
	s.Assert().Len(ubd.Entries, 1, "addr1 should have 1 unbonding entry")
	s.Assert().Equal(coin.Amount, ubd.Entries[0].Balance, "addr1 should have the total delegation amount in unbonding delegation")
	ubd, found = s.app.StakingKeeper.GetUnbondingDelegation(s.ctx, addr2, unbondedVal1.GetOperator())
	s.Assert().True(found, "should have found addr2 unbonding delegation")
	s.Assert().Len(ubd.Entries, 1, "addr2 should have 1 unbonding entry")
	s.Assert().Equal(coin.Amount, ubd.Entries[0].Balance, "addr2 should have the total delegation amount in unbonding delegation")

	// 2 unbonded validators with delegations past inactive time, both should be removed
	unbondedVal1 = s.CreateValidator(s.startTime.Add(-30*24*time.Hour), stakingtypes.Unbonded)
	s.DelegateToValidator(unbondedVal1.GetOperator(), sdk.AccAddress(addr1), coin)
	s.DelegateToValidator(unbondedVal1.GetOperator(), sdk.AccAddress(addr2), coin)
	unbondedVal1, _ = s.app.StakingKeeper.GetValidator(s.ctx, unbondedVal1.GetOperator())
	s.Require().Equal(sdk.NewDec(20_000), unbondedVal1.DelegatorShares, "shares should have updated")
	unbondedVal2 := s.CreateValidator(s.startTime.Add(-29*24*time.Hour), stakingtypes.Unbonded)
	s.DelegateToValidator(unbondedVal2.GetOperator(), sdk.AccAddress(addr1), coin)
	s.DelegateToValidator(unbondedVal2.GetOperator(), sdk.AccAddress(addr2), coin)
	unbondedVal2, _ = s.app.StakingKeeper.GetValidator(s.ctx, unbondedVal2.GetOperator())
	s.Require().Equal(sdk.NewDec(20_000), unbondedVal2.DelegatorShares, "shares should have updated after delegation")
	validators = s.app.StakingKeeper.GetAllValidators(s.ctx)
	s.Require().Equal(3, len(validators), "should three validators, one bonded and two unbonded with delegations past inactive time")

	expectedLogLines = []string{
		"INF removing any validator that has been inactive (unbonded) for 21 hours",
		fmt.Sprintf("INF validator %v has been inactive (unbonded) for %d days and will be removed", unbondedVal1.OperatorAddress, 30),
		fmt.Sprintf("INF undelegate delegator %v from validator %v of all shares (%v)", addr1.String(), unbondedVal1.OperatorAddress, sdk.NewDec(coin.Amount.Int64())),
		fmt.Sprintf("INF undelegate delegator %v from validator %v of all shares (%v)", addr2.String(), unbondedVal1.OperatorAddress, sdk.NewDec(coin.Amount.Int64())),
		fmt.Sprintf("INF validator %v has been inactive (unbonded) for %d days and will be removed", unbondedVal2.OperatorAddress, 29),
		fmt.Sprintf("INF undelegate delegator %v from validator %v of all shares (%v)", addr1.String(), unbondedVal2.OperatorAddress, sdk.NewDec(coin.Amount.Int64())),
		fmt.Sprintf("INF undelegate delegator %v from validator %v of all shares (%v)", addr2.String(), unbondedVal2.OperatorAddress, sdk.NewDec(coin.Amount.Int64())),
		"INF a total of 2 inactive (unbonded) validators have been removed",
	}
	s.ExecuteAndAssertLogs(removeInactiveValidators, expectedLogLines)

	validators = s.app.StakingKeeper.GetAllValidators(s.ctx)
	s.Assert().Equal(1, len(validators), "should have removed the unbonded delegator")
	ubd, found = s.app.StakingKeeper.GetUnbondingDelegation(s.ctx, addr1, unbondedVal1.GetOperator())
	s.Assert().True(found, "should have found addr1 unbonding delegation")
	s.Assert().Len(ubd.Entries, 1, "addr1 should have 1 unbonding entry")
	s.Assert().Equal(coin.Amount, ubd.Entries[0].Balance, "addr1 should have the total delegation amount in unbonding delegation")
	ubd, found = s.app.StakingKeeper.GetUnbondingDelegation(s.ctx, addr2, unbondedVal1.GetOperator())
	s.Assert().True(found, "should have found addr2 unbonding delegation")
	s.Assert().Len(ubd.Entries, 1, "addr2 should have 1 unbonding entry")
	s.Assert().Equal(coin.Amount, ubd.Entries[0].Balance, "addr2 should have the total delegation amount in unbonding delegation")

	// 2 unbonded validators, 1 under the inactive day count, should only remove one
	unbondedVal1 = s.CreateValidator(s.startTime.Add(-30*24*time.Hour), stakingtypes.Unbonded)
	s.DelegateToValidator(unbondedVal1.GetOperator(), sdk.AccAddress(addr1), coin)
	s.DelegateToValidator(unbondedVal1.GetOperator(), sdk.AccAddress(addr2), coin)
	unbondedVal1, _ = s.app.StakingKeeper.GetValidator(s.ctx, unbondedVal1.GetOperator())
	s.Require().Equal(sdk.NewDec(20_000), unbondedVal1.DelegatorShares, "shares should have updated after delegation")
	unbondedVal2 = s.CreateValidator(s.startTime.Add(-20*24*time.Hour), stakingtypes.Unbonded)
	s.DelegateToValidator(unbondedVal2.GetOperator(), sdk.AccAddress(addr1), coin)
	unbondedVal2, _ = s.app.StakingKeeper.GetValidator(s.ctx, unbondedVal2.GetOperator())
	s.Require().Equal(sdk.NewDec(10_000), unbondedVal2.DelegatorShares, "shares should have updated after delegation")
	validators = s.app.StakingKeeper.GetAllValidators(s.ctx)
	s.Require().Equal(3, len(validators), "should three validators, one bonded and one unbonded with delegations after inactive time and one not past inactive time")

	expectedLogLines = []string{
		"INF removing any validator that has been inactive (unbonded) for 21 hours",
		fmt.Sprintf("INF validator %v has been inactive (unbonded) for %d days and will be removed", unbondedVal1.OperatorAddress, 30),
		fmt.Sprintf("INF undelegate delegator %v from validator %v of all shares (%v)", addr1.String(), unbondedVal1.OperatorAddress, sdk.NewDec(coin.Amount.Int64())),
		fmt.Sprintf("INF undelegate delegator %v from validator %v of all shares (%v)", addr2.String(), unbondedVal1.OperatorAddress, sdk.NewDec(coin.Amount.Int64())),
		"INF a total of 1 inactive (unbonded) validators have been removed",
	}
	s.ExecuteAndAssertLogs(removeInactiveValidators, expectedLogLines)

	validators = s.app.StakingKeeper.GetAllValidators(s.ctx)
	s.Assert().Equal(2, len(validators), "should have removed the unbonded delegator")
	ubd, found = s.app.StakingKeeper.GetUnbondingDelegation(s.ctx, addr1, unbondedVal1.GetOperator())
	s.Assert().True(found, "should have found addr1 unbonding delegation")
	s.Assert().Len(ubd.Entries, 1, "addr1 should have 1 unbonding entry")
	s.Assert().Equal(coin.Amount, ubd.Entries[0].Balance, "addr1 should have the total delegation amount in unbonding delegation")
	ubd, found = s.app.StakingKeeper.GetUnbondingDelegation(s.ctx, addr2, unbondedVal1.GetOperator())
	s.Assert().True(found, "should have found addr2 unbonding delegation")
	s.Assert().Len(ubd.Entries, 1, "addr2 should have 1 unbonding entry")
	s.Assert().Equal(coin.Amount, ubd.Entries[0].Balance, "addr2 should have the total delegation amount in unbonding delegation")
}

func (s *UpgradeTestSuite) CreateAndFundAccount(coin sdk.Coin) sdk.AccAddress {
	key2 := secp256k1.GenPrivKey()
	pub2 := key2.PubKey()
	addr2 := sdk.AccAddress(pub2.Address())
	testutil.FundAccount(s.app.BankKeeper, s.ctx, addr2, sdk.Coins{coin})
	return addr2
}

func (s *UpgradeTestSuite) CreateValidator(unbondedTime time.Time, status stakingtypes.BondStatus) stakingtypes.Validator {
	key := secp256k1.GenPrivKey()
	pub := key.PubKey()
	addr := sdk.AccAddress(pub.Address())
	valAddr := sdk.ValAddress(addr)
	validator, err := stakingtypes.NewValidator(valAddr, pub, stakingtypes.NewDescription(valAddr.String(), "", "", "", ""))
	s.Require().NoError(err, "could not init new validator")
	validator.UnbondingTime = unbondedTime
	validator.Status = status
	s.app.StakingKeeper.SetValidator(s.ctx, validator)
	err = s.app.StakingKeeper.SetValidatorByConsAddr(s.ctx, validator)
	s.Require().NoError(err, "could not SetValidatorByConsAddr ")
	err = s.app.StakingKeeper.AfterValidatorCreated(s.ctx, validator.GetOperator())
	s.Require().NoError(err, "could not AfterValidatorCreated")
	return validator
}

func (s *UpgradeTestSuite) DelegateToValidator(valAddress sdk.ValAddress, delegatorAddress sdk.AccAddress, coin sdk.Coin) {
	validator, found := s.app.StakingKeeper.GetValidator(s.ctx, valAddress)
	s.Require().True(found, "validator does not exist, cannot delegate address: %v", valAddress.String())
	_, err := s.app.StakingKeeper.Delegate(s.ctx, delegatorAddress, coin.Amount, types.Unbonded, validator, true)
	s.Require().NoError(err, "error delegating to validator %v from %v", valAddress.String(), delegatorAddress.String())
}

func (s *UpgradeTestSuite) ExecuteAndAssertLogs(delegate func(ctx sdk.Context, app *App), expectedLogs []string) {
	s.logBuffer.Reset()
	delegate(s.ctx, s.app)
	logOutput := s.logBuffer.String()

	logLines := strings.Split(logOutput, "\n")
	for _, expectedLine := range expectedLogs {
		s.Assert().Contains(logLines, expectedLine, "Expecting: %q", expectedLine)
	}
}
