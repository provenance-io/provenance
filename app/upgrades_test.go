package app

import (
	"bytes"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/provenance-io/provenance/x/exchange"
	markertypes "github.com/provenance-io/provenance/x/marker/types"
	msgfeestypes "github.com/provenance-io/provenance/x/msgfees/types"
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
	defer SetLoggerMaker(SetLoggerMaker(BufferedInfoLoggerMaker(&s.logBuffer)))
	s.app = Setup(s.T())
	s.logBuffer.Reset()
	s.startTime = time.Now()
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{Time: s.startTime})
}

// GetLogOutput gets the log buffer contents. This (probably) also clears the log buffer.
func (s *UpgradeTestSuite) GetLogOutput(msg string, args ...interface{}) string {
	logOutput := s.logBuffer.String()
	s.T().Logf(msg+" log output:\n%s", append(args, logOutput)...)
	return logOutput
}

// LogIfError logs an error if it's not nil.
// The error is automatically added to the format and args.
// Use this if there's a possible error that we probably don't care about (but might).
func (s *UpgradeTestSuite) LogIfError(err error, format string, args ...interface{}) {
	if err != nil {
		s.T().Logf(format+" error: %v", append(args, err)...)
	}
}

// AssertUpgradeHandlerLogs runs the Handler of the provided key and asserts that
// each entry in expInLog is in the log output, and each entry in expNotInLog
// is not in the log output.
//
// entries in expInLog are expected to be full lines.
// entries in expNotInLog can be parts of lines.
//
// This returns the log output and whether everything is as expected (true = all good).
func (s *UpgradeTestSuite) AssertUpgradeHandlerLogs(key string, expInLog, expNotInLog []string) (string, bool) {
	s.T().Helper()

	if !s.Assert().Contains(upgrades, key, "%q defined upgrades map", key) {
		return "", false // If the upgrades map doesn't have that key, there's nothing more to do in here.
	}
	handler := upgrades[key].Handler
	if !s.Assert().NotNil(handler, "upgrades[%q].Handler", key) {
		return "", false // If the entry doesn't have a .Handler, there's nothing more to do in here.
	}

	// This app was just created brand new, so it will create all the modules with their most
	// recent versions. As such, this GetModuleVersionMap will return all the most current
	// module versions info. That means that none of the module migrations will be running when we
	// call the handler, and that the log won't have any entries from any specific module migrations.
	// If you really want to include those in this run, you can try doing stuff with SetModuleVersionMap
	// prior to calling this AssertUpgradeHandlerLogs function. You'll probably also need to recreate
	// any old state for that module since anything added already was done using new stuff.
	origVersionMap := s.app.UpgradeKeeper.GetModuleVersionMap(s.ctx)

	msgFormat := fmt.Sprintf("upgrades[%q].Handler(...)", key)

	var versionMap module.VersionMap
	var err error
	s.logBuffer.Reset()
	testFunc := func() {
		versionMap, err = handler(s.ctx, s.app, origVersionMap)
	}
	didNotPanic := s.Assert().NotPanics(testFunc, msgFormat)
	logOutput := s.GetLogOutput(msgFormat)
	if !didNotPanic {
		// If the handler panicked, there's nothing more to check in here.
		return logOutput, false
	}

	// Checking for error cases should be done in individual function unit tests.
	// For this, always check for no error and a non-empty version map.
	rv := s.Assert().NoErrorf(err, msgFormat+" error")
	rv = s.Assert().NotEmptyf(versionMap, msgFormat+" version map") && rv
	rv = s.AssertLogContents(logOutput, expInLog, expNotInLog, true, msgFormat) && rv

	return logOutput, rv
}

// ExecuteAndAssertLogs executes the provided runner and makes sure the logs have the expected lines and don't have any unexpected substrings.
func (s *UpgradeTestSuite) ExecuteAndAssertLogs(runner func(), expInLog []string, expNotInLog []string, enforceOrder bool, msgFormat string, args ...interface{}) (string, bool) {
	s.logBuffer.Reset()
	didNotPanic := s.Assert().NotPanicsf(runner, msgFormat, args...)
	logOutput := s.GetLogOutput(msgFormat, args...)
	if !didNotPanic {
		return logOutput, false
	}

	rv := s.AssertLogContents(logOutput, expInLog, expNotInLog, enforceOrder, msgFormat, args...)

	return logOutput, rv
}

// splitLogOutput splits the provided log output into individual lines.
func splitLogOutput(logOutput string) []string {
	rv := strings.Split(logOutput, "\n")
	if len(rv[len(rv)-1]) == 0 {
		rv = rv[:len(rv)-1]
	}
	return rv
}

// AssertLogContents asserts that the provided log output string contains all expInLog lines in the provided order,
// and doesn't contain any string in the expNotInLog slice.
// Returns true if everything is as expected.
func (s *UpgradeTestSuite) AssertLogContents(logOutput string, expInLog, expNotInLog []string, enforceOrder bool, msg string, args ...interface{}) bool {
	s.T().Helper()
	rv := s.assertLogLinesInOrder(logOutput, expInLog, enforceOrder, msg+" log output", args...)
	rv = s.assertLogDoesNotContain(logOutput, expNotInLog, msg+" log output", args...) && rv
	return rv
}

// assertLogLinesInOrder asserts that the log output has whole lines matching each of the expInLog entries,
// and that they're in the same order (allowing for extra lines to be in the log output that aren't in expInLog).
// Designed for AssertLogContents, please use that.
func (s *UpgradeTestSuite) assertLogLinesInOrder(logOutput string, expInLog []string, enforceOrder bool, msg string, args ...interface{}) bool {
	if len(expInLog) == 0 {
		return true
	}

	logLines := splitLogOutput(logOutput)

	allThere := true
	// First, just make sure all the lines are there.
	// This gives a nicer failure message than if we try to do it while checking ordering.
	for _, exp := range expInLog {
		// I'm including the expected in the msgAndArgs here even though it's also included in the
		// failure message because the failure message puts everything on one line. So the expected
		// string is almost at the end of a very long line. Having it in the "message" portion of
		// the failure makes it easier to identify the problematic entry.
		if !s.Assert().Containsf(logLines, exp, msg+"\nExpecting: %q", append(args, exp)...) {
			allThere = false
		}
	}
	if !allThere {
		return false
	}
	if !enforceOrder {
		return true
	}

	// Now make sure they're in the same order, allowing for extra lines in the log that might not be expected.
	e := 0
	for _, logLine := range logLines {
		if expInLog[e] == logLine {
			e++
			if e == len(expInLog) {
				return true
			}
		}
	}

	// We didn't get to the end of the expected list. Issue a custom failure.
	failureLines := []string{
		"Log lines not in expected order.",
		fmt.Sprintf("End of log reached looking for [%d]: %q", e, expInLog[e]),
		// Note: We know that all expInLog lines are in the log output.
		// So the above loop will find at least the first entry, and e will be at least 1.
		fmt.Sprintf("Expected it to be after [%d]: %q", e-1, expInLog[e-1]),
	}
	// Putting the actual entries first to put it closer to the previous two lines,
	// which have the most important expected entries.
	failureLines = append(failureLines, "Actual:")
	for i, line := range logLines {
		failureLines = append(failureLines, fmt.Sprintf("  [%d]: %q", i, line))
	}
	// Then give all the expected entries for good form.
	failureLines = append(failureLines, "Expected:")
	for i, line := range expInLog {
		failureLines = append(failureLines, fmt.Sprintf("  [%d]: %q", i, line))
	}
	return s.Failf(strings.Join(failureLines, "\n"), msg, args...)
}

// assertLogDoesNotContain asserts that the log output does not contain any of the expNotInLog as substrings.
// Designed for AssertLogContents, please use that.
func (s *UpgradeTestSuite) assertLogDoesNotContain(logOutput string, expNotInLog []string, msg string, args ...interface{}) bool {
	noneThere := true
	for _, unexp := range expNotInLog {
		// I'm including the unexpected in the msgAndArgs here even though it's also included in the
		// failure message because the failure message puts everything on one line. So the expected
		// string is almost at the end of a very long line. Having it in the "message" portion of
		// the failure makes it easier to identify the problematic entry.
		if !s.Assert().NotContainsf(logOutput, unexp, msg+"\nNot Expecting: %q", append(args, unexp)...) {
			noneThere = false
		}
	}
	return noneThere
}

// CreateAndFundAccount creates a new account in the app and funds it with the provided coin.
func (s *UpgradeTestSuite) CreateAndFundAccount(coin sdk.Coin) sdk.AccAddress {
	key2 := secp256k1.GenPrivKey()
	pub2 := key2.PubKey()
	addr2 := sdk.AccAddress(pub2.Address())
	s.LogIfError(testutil.FundAccount(s.app.BankKeeper, s.ctx, addr2, sdk.Coins{coin}), "FundAccount(..., %q)", coin.String())
	return addr2
}

// CreateValidator creates a new validator in the app.
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

// DelegateToValidator delegates to a validator in the app.
func (s *UpgradeTestSuite) DelegateToValidator(valAddress sdk.ValAddress, delegatorAddress sdk.AccAddress, coin sdk.Coin) {
	validator, found := s.app.StakingKeeper.GetValidator(s.ctx, valAddress)
	s.Require().True(found, "GetValidator(%q)", valAddress.String())
	_, err := s.app.StakingKeeper.Delegate(s.ctx, delegatorAddress, coin.Amount, types.Unbonded, validator, true)
	s.Require().NoError(err, "Delegate(%q, %q, unbonded, %q, true) error", delegatorAddress.String(), coin.String(), valAddress.String())
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

func (s *UpgradeTestSuite) TestSaffronRC1() {
	// Each part is (hopefully) tested thoroughly on its own.
	// So for this test, just make sure there's log entries for each part being done.

	expInLog := []string{
		"INF Starting module migrations. This may take a significant amount of time to complete. Do not restart node.",
		"INF Removing all delegations from validators that have been inactive (unbonded) for 21 days.",
		"INF Updating ICQ params",
		"INF Done updating ICQ params",
		"INF Updating MaxSupply marker param",
		"INF Done updating MaxSupply marker param",
		"INF Ensuring exchange module params are set.",
	}

	s.AssertUpgradeHandlerLogs("saffron-rc1", expInLog, nil)
}

func (s *UpgradeTestSuite) TestSaffronRC2() {
	// Each part is (hopefully) tested thoroughly on its own.
	// So for this test, just make sure there's log entries for each part being done.

	expInLog := []string{
		"INF Updating ibc marker denom metadata",
		"INF Done updating ibc marker denom metadata",
	}

	s.AssertUpgradeHandlerLogs("saffron-rc2", expInLog, nil)
}

func (s *UpgradeTestSuite) TestSaffronRC3() {
	expInLog := []string{
		"INF Updating ibc marker denom metadata",
		"INF Done updating ibc marker denom metadata",
	}

	s.AssertUpgradeHandlerLogs("saffron-rc3", expInLog, nil)
}

func (s *UpgradeTestSuite) TestSaffron() {
	// Each part is (hopefully) tested thoroughly on its own.
	// So for this test, just make sure there's log entries for each part being done.

	expInLog := []string{
		"INF Starting module migrations. This may take a significant amount of time to complete. Do not restart node.",
		"INF Removing all delegations from validators that have been inactive (unbonded) for 21 days.",
		"INF Updating ICQ params",
		"INF Done updating ICQ params",
		"INF Updating MaxSupply marker param",
		"INF Done updating MaxSupply marker param",
		"INF Adding marker net asset values",
		"INF Done adding marker net asset values",
		"INF Ensuring exchange module params are set.",
		"INF Updating ibc marker denom metadata",
		"INF Done updating ibc marker denom metadata",
	}

	s.AssertUpgradeHandlerLogs("saffron", expInLog, nil)
}

func (s *UpgradeTestSuite) TestTourmalineRC1() {
	expInLog := []string{
		"INF Converting NAV units.",
	}
	expNotInLog := []string{
		"INF Setting MsgFees Params NhashPerUsdMil to 40000000.",
	}

	s.AssertUpgradeHandlerLogs("tourmaline-rc1", expInLog, expNotInLog)
}

func (s *UpgradeTestSuite) TestTourmalineRC2() {
	key := "tourmaline-rc2"
	s.Require().Contains(upgrades, key, "%q defined upgrades map", key)
	s.Require().Empty(upgrades[key], "upgrades[%q]", key)
}

func (s *UpgradeTestSuite) TestTourmalineRC3() {
	expInLog := []string{
		"INF Setting exchange module payment params to defaults.",
		"INF Done setting exchange module payment params to defaults.",
	}

	s.AssertUpgradeHandlerLogs("tourmaline-rc3", expInLog, nil)
}

func (s *UpgradeTestSuite) TestTourmaline() {
	expInLog := []string{
		"INF Starting module migrations. This may take a significant amount of time to complete. Do not restart node.",
		"INF Removing all delegations from validators that have been inactive (unbonded) for 21 days.",
		"INF Converting NAV units.",
		"INF Setting MsgFees Params NhashPerUsdMil to 40000000.",
		"INF Setting exchange module payment params to defaults.",
	}

	s.AssertUpgradeHandlerLogs("tourmaline", expInLog, nil)
}

func (s *UpgradeTestSuite) TestRemoveInactiveValidatorDelegations() {
	addr1 := s.CreateAndFundAccount(sdk.NewInt64Coin("stake", 1000000))
	addr2 := s.CreateAndFundAccount(sdk.NewInt64Coin("stake", 1000000))

	runner := func() {
		removeInactiveValidatorDelegations(s.ctx, s.app)
	}
	runnerName := "removeInactiveValidatorDelegations"

	delegationCoin := sdk.NewInt64Coin("stake", 10000)
	delegationCoinAmt := sdk.NewDec(delegationCoin.Amount.Int64())

	s.Run("just one bonded validator", func() {
		validators := s.app.StakingKeeper.GetAllValidators(s.ctx)
		s.Require().Len(validators, 1, "GetAllValidators after setup")

		expectedLogLines := []string{
			"INF Removing all delegations from validators that have been inactive (unbonded) for 21 days.",
			"INF A total of 0 inactive (unbonded) validators have had all their delegators removed.",
		}
		s.ExecuteAndAssertLogs(runner, expectedLogLines, nil, true, runnerName)

		newValidators := s.app.StakingKeeper.GetAllValidators(s.ctx)
		s.Require().Len(newValidators, 1, "GetAllValidators after %s", runnerName)
	})

	s.Run("one unbonded validator with a delegation", func() {
		// single unbonded validator with 1 delegations, should be removed
		unbondedVal1 := s.CreateValidator(s.startTime.Add(-30*24*time.Hour), stakingtypes.Unbonded)
		addr1Balance := s.app.BankKeeper.GetBalance(s.ctx, addr1, "stake")
		s.DelegateToValidator(unbondedVal1.GetOperator(), addr1, delegationCoin)
		s.Require().Equal(addr1Balance.Sub(delegationCoin), s.app.BankKeeper.GetBalance(s.ctx, addr1, "stake"), "addr1 should have less funds")
		validators := s.app.StakingKeeper.GetAllValidators(s.ctx)
		s.Require().Len(validators, 2, "Setup: GetAllValidators should have: 1 bonded, 1 unbonded")

		expectedLogLines := []string{
			"INF Removing all delegations from validators that have been inactive (unbonded) for 21 days.",
			fmt.Sprintf("INF Validator %v has been inactive (unbonded) for %d days and will be removed.", unbondedVal1.OperatorAddress, 30),
			fmt.Sprintf("INF Undelegate delegator %v from validator %v of all shares (%v).", addr1.String(), unbondedVal1.OperatorAddress, delegationCoinAmt),
			"INF A total of 1 inactive (unbonded) validators have had all their delegators removed.",
		}
		s.ExecuteAndAssertLogs(runner, expectedLogLines, nil, true, runnerName)

		validators = s.app.StakingKeeper.GetAllValidators(s.ctx)
		s.Assert().Len(validators, 1, "GetAllValidators after %s", runnerName)
		ubd, found := s.app.StakingKeeper.GetUnbondingDelegation(s.ctx, addr1, unbondedVal1.GetOperator())
		s.Assert().True(found, "GetUnbondingDelegation found")
		if s.Assert().Len(ubd.Entries, 1, "UnbondingDelegation entries") {
			s.Assert().Equal(delegationCoin.Amount, ubd.Entries[0].Balance, "UnbondingDelegation balance")
		}
	})

	s.Run("one unbonded validator with 2 delegations", func() {
		// single unbonded validator with 2 delegations
		unbondedVal1 := s.CreateValidator(s.startTime.Add(-30*24*time.Hour), stakingtypes.Unbonded)
		addr1Balance := s.app.BankKeeper.GetBalance(s.ctx, addr1, "stake")
		addr2Balance := s.app.BankKeeper.GetBalance(s.ctx, addr2, "stake")
		s.DelegateToValidator(unbondedVal1.GetOperator(), addr1, delegationCoin)
		s.DelegateToValidator(unbondedVal1.GetOperator(), addr2, delegationCoin)
		s.Require().Equal(addr1Balance.Sub(delegationCoin), s.app.BankKeeper.GetBalance(s.ctx, addr1, "stake"), "addr1 should have less funds after delegation")
		s.Require().Equal(addr2Balance.Sub(delegationCoin), s.app.BankKeeper.GetBalance(s.ctx, addr2, "stake"), "addr2 should have less funds after delegation")
		var found bool
		unbondedVal1, found = s.app.StakingKeeper.GetValidator(s.ctx, unbondedVal1.GetOperator())
		s.Require().True(found, "Setup: GetValidator(unbondedVal1) found")
		s.Require().Equal(delegationCoinAmt.Add(delegationCoinAmt), unbondedVal1.DelegatorShares, "Setup: unbondedVal1.DelegatorShares")
		validators := s.app.StakingKeeper.GetAllValidators(s.ctx)
		s.Require().Len(validators, 2, "Setup: GetAllValidators should have: 1 bonded, 1 unbonded")

		expectedLogLines := []string{
			"INF Removing all delegations from validators that have been inactive (unbonded) for 21 days.",
			fmt.Sprintf("INF Validator %v has been inactive (unbonded) for %d days and will be removed.", unbondedVal1.OperatorAddress, 30),
			fmt.Sprintf("INF Undelegate delegator %v from validator %v of all shares (%v).", addr1.String(), unbondedVal1.OperatorAddress, delegationCoinAmt),
			fmt.Sprintf("INF Undelegate delegator %v from validator %v of all shares (%v).", addr2.String(), unbondedVal1.OperatorAddress, delegationCoinAmt),
			"INF A total of 1 inactive (unbonded) validators have had all their delegators removed.",
		}
		s.ExecuteAndAssertLogs(runner, expectedLogLines, nil, false, runnerName)

		validators = s.app.StakingKeeper.GetAllValidators(s.ctx)
		s.Assert().Len(validators, 1, "GetAllValidators after %s", runnerName)
		ubd, found := s.app.StakingKeeper.GetUnbondingDelegation(s.ctx, addr1, unbondedVal1.GetOperator())
		s.Assert().True(found, "GetUnbondingDelegation(addr1) found")
		if s.Assert().Len(ubd.Entries, 1, "GetUnbondingDelegation(addr1) entries") {
			s.Assert().Equal(delegationCoin.Amount, ubd.Entries[0].Balance, "GetUnbondingDelegation(addr1) balance")
		}
		ubd, found = s.app.StakingKeeper.GetUnbondingDelegation(s.ctx, addr2, unbondedVal1.GetOperator())
		s.Assert().True(found, "GetUnbondingDelegation(addr2) found")
		if s.Assert().Len(ubd.Entries, 1, "GetUnbondingDelegation(addr2) entries") {
			s.Assert().Equal(delegationCoin.Amount, ubd.Entries[0].Balance, "GetUnbondingDelegation(addr2) balance")
		}
	})

	s.Run("two unbonded validators to be removed", func() {
		// 2 unbonded validators with delegations past inactive time, both should be removed
		unbondedVal1 := s.CreateValidator(s.startTime.Add(-30*24*time.Hour), stakingtypes.Unbonded)
		s.DelegateToValidator(unbondedVal1.GetOperator(), addr1, delegationCoin)
		s.DelegateToValidator(unbondedVal1.GetOperator(), addr2, delegationCoin)
		var found bool
		unbondedVal1, found = s.app.StakingKeeper.GetValidator(s.ctx, unbondedVal1.GetOperator())
		s.Require().True(found, "Setup: GetValidator(unbondedVal1) found")
		s.Require().Equal(delegationCoinAmt.Add(delegationCoinAmt), unbondedVal1.DelegatorShares, "Setup: shares delegated to unbondedVal1")

		unbondedVal2 := s.CreateValidator(s.startTime.Add(-29*24*time.Hour), stakingtypes.Unbonded)
		s.DelegateToValidator(unbondedVal2.GetOperator(), addr1, delegationCoin)
		s.DelegateToValidator(unbondedVal2.GetOperator(), addr2, delegationCoin)
		unbondedVal2, found = s.app.StakingKeeper.GetValidator(s.ctx, unbondedVal2.GetOperator())
		s.Require().True(found, "Setup: GetValidator(unbondedVal2) found")
		s.Require().Equal(delegationCoinAmt.Add(delegationCoinAmt), unbondedVal2.DelegatorShares, "Setup: shares delegated to unbondedVal2")
		validators := s.app.StakingKeeper.GetAllValidators(s.ctx)
		s.Require().Len(validators, 3, "Setup: GetAllValidators should have: 1 bonded, 2 unbonded")

		expectedLogLines := []string{
			"INF Removing all delegations from validators that have been inactive (unbonded) for 21 days.",
			fmt.Sprintf("INF Validator %v has been inactive (unbonded) for %d days and will be removed.", unbondedVal1.OperatorAddress, 30),
			fmt.Sprintf("INF Undelegate delegator %v from validator %v of all shares (%v).", addr1.String(), unbondedVal1.OperatorAddress, delegationCoinAmt),
			fmt.Sprintf("INF Undelegate delegator %v from validator %v of all shares (%v).", addr2.String(), unbondedVal1.OperatorAddress, delegationCoinAmt),
			fmt.Sprintf("INF Validator %v has been inactive (unbonded) for %d days and will be removed.", unbondedVal2.OperatorAddress, 29),
			fmt.Sprintf("INF Undelegate delegator %v from validator %v of all shares (%v).", addr1.String(), unbondedVal2.OperatorAddress, delegationCoinAmt),
			fmt.Sprintf("INF Undelegate delegator %v from validator %v of all shares (%v).", addr2.String(), unbondedVal2.OperatorAddress, delegationCoinAmt),
			"INF A total of 2 inactive (unbonded) validators have had all their delegators removed.",
		}
		s.ExecuteAndAssertLogs(runner, expectedLogLines, nil, false, runnerName)

		validators = s.app.StakingKeeper.GetAllValidators(s.ctx)
		s.Assert().Len(validators, 1, "GetAllValidators after %s", runnerName)
		ubd, found := s.app.StakingKeeper.GetUnbondingDelegation(s.ctx, addr1, unbondedVal1.GetOperator())
		s.Assert().True(found, "GetUnbondingDelegation(addr1) found")
		if s.Assert().Len(ubd.Entries, 1, "GetUnbondingDelegation(addr1) entries") {
			s.Assert().Equal(delegationCoin.Amount, ubd.Entries[0].Balance, "GetUnbondingDelegation(addr1) balance")
		}
		ubd, found = s.app.StakingKeeper.GetUnbondingDelegation(s.ctx, addr2, unbondedVal1.GetOperator())
		s.Assert().True(found, "GetUnbondingDelegation(addr2) found")
		if s.Assert().Len(ubd.Entries, 1, "GetUnbondingDelegation(addr2) entries") {
			s.Assert().Equal(delegationCoin.Amount, ubd.Entries[0].Balance, "GetUnbondingDelegation(addr2) balance")
		}
	})

	s.Run("two unbonded validators one too recently", func() {
		// 2 unbonded validators, 1 under the inactive day count, should only remove one
		unbondedVal1 := s.CreateValidator(s.startTime.Add(-30*24*time.Hour), stakingtypes.Unbonded)
		s.DelegateToValidator(unbondedVal1.GetOperator(), addr1, delegationCoin)
		s.DelegateToValidator(unbondedVal1.GetOperator(), addr2, delegationCoin)
		var found bool
		unbondedVal1, found = s.app.StakingKeeper.GetValidator(s.ctx, unbondedVal1.GetOperator())
		s.Require().True(found, "Setup: GetValidator(unbondedVal1) found")
		s.Require().Equal(delegationCoinAmt.Add(delegationCoinAmt), unbondedVal1.DelegatorShares, "Setup: shares delegated to unbondedVal1")

		unbondedVal2 := s.CreateValidator(s.startTime.Add(-20*24*time.Hour), stakingtypes.Unbonded)
		s.DelegateToValidator(unbondedVal2.GetOperator(), addr1, delegationCoin)
		unbondedVal2, found = s.app.StakingKeeper.GetValidator(s.ctx, unbondedVal2.GetOperator())
		s.Require().True(found, "Setup: GetValidator(unbondedVal2) found")
		s.Require().Equal(delegationCoinAmt, unbondedVal2.DelegatorShares, "Setup: shares delegated to unbondedVal1")
		validators := s.app.StakingKeeper.GetAllValidators(s.ctx)
		s.Require().Len(validators, 3, "Setup: GetAllValidators should have: 1 bonded, 1 recently unbonded, 1 old unbonded")

		expectedLogLines := []string{
			"INF Removing all delegations from validators that have been inactive (unbonded) for 21 days.",
			fmt.Sprintf("INF Validator %v has been inactive (unbonded) for %d days and will be removed.", unbondedVal1.OperatorAddress, 30),
			fmt.Sprintf("INF Undelegate delegator %v from validator %v of all shares (%v).", addr1.String(), unbondedVal1.OperatorAddress, delegationCoinAmt),
			fmt.Sprintf("INF Undelegate delegator %v from validator %v of all shares (%v).", addr2.String(), unbondedVal1.OperatorAddress, delegationCoinAmt),
			"INF A total of 1 inactive (unbonded) validators have had all their delegators removed.",
		}
		notExpectedLogLines := []string{
			fmt.Sprintf("Validator %v has been inactive (unbonded).", unbondedVal2.OperatorAddress),
			fmt.Sprintf("Undelegate delegator %v from validator %v", addr1.String(), unbondedVal2.OperatorAddress),
			fmt.Sprintf("Undelegate delegator %v from validator %v", addr2.String(), unbondedVal2.OperatorAddress),
		}
		s.ExecuteAndAssertLogs(runner, expectedLogLines, notExpectedLogLines, false, runnerName)

		validators = s.app.StakingKeeper.GetAllValidators(s.ctx)
		s.Assert().Len(validators, 2, "GetAllValidators after %s", runnerName)
		ubd, found := s.app.StakingKeeper.GetUnbondingDelegation(s.ctx, addr1, unbondedVal1.GetOperator())
		s.Assert().True(found, "GetUnbondingDelegation(addr1) found")
		if s.Assert().Len(ubd.Entries, 1, "GetUnbondingDelegation(addr1) entries") {
			s.Assert().Equal(delegationCoin.Amount, ubd.Entries[0].Balance, "GetUnbondingDelegation(addr1) balance")
		}
		ubd, found = s.app.StakingKeeper.GetUnbondingDelegation(s.ctx, addr2, unbondedVal1.GetOperator())
		s.Assert().True(found, "GetUnbondingDelegation(addr2) found")
		if s.Assert().Len(ubd.Entries, 1, "GetUnbondingDelegation(addr2) entries") {
			s.Assert().Equal(delegationCoin.Amount, ubd.Entries[0].Balance, "GetUnbondingDelegation(addr2) balance")
		}
	})

	s.Run("unbonded without delegators", func() {
		// create a unbonded validator with out delegators, should not remove
		unbondedVal1 := s.CreateValidator(s.startTime.Add(-30*24*time.Hour), stakingtypes.Unbonded)
		validators := s.app.StakingKeeper.GetAllValidators(s.ctx)
		s.Require().Len(validators, 3, "Setup: GetAllValidators should have: 1 bonded, 1 recently unbonded, 1 empty unbonded")

		expectedLogLines := []string{
			"INF Removing all delegations from validators that have been inactive (unbonded) for 21 days.",
			fmt.Sprintf("INF Validator %v has been inactive (unbonded) for %d days and will be removed.", unbondedVal1.OperatorAddress, 30),
			"INF A total of 1 inactive (unbonded) validators have had all their delegators removed.",
		}
		s.ExecuteAndAssertLogs(runner, expectedLogLines, nil, true, runnerName)

		validators = s.app.StakingKeeper.GetAllValidators(s.ctx)
		s.Assert().Len(validators, 3, "GetAllValidators after %s", runnerName)
	})
}

func (s *UpgradeTestSuite) TestAddMarkerNavs() {
	address1 := sdk.AccAddress("address1")
	testcoin := markertypes.NewEmptyMarkerAccount("testcoin",
		address1.String(),
		[]markertypes.AccessGrant{})
	testcoin.Supply = sdk.OneInt()
	s.Require().NoError(s.app.MarkerKeeper.AddMarkerAccount(s.ctx, testcoin), "AddMarkerAccount() error")

	testcoinInList := markertypes.NewEmptyMarkerAccount("testcoininlist",
		address1.String(),
		[]markertypes.AccessGrant{})
	testcoinInList.Supply = sdk.OneInt()
	s.Require().NoError(s.app.MarkerKeeper.AddMarkerAccount(s.ctx, testcoinInList), "AddMarkerAccount() error")

	nosupplycoin := markertypes.NewEmptyMarkerAccount("nosupplycoin",
		address1.String(),
		[]markertypes.AccessGrant{})
	s.Require().NoError(s.app.MarkerKeeper.AddMarkerAccount(s.ctx, nosupplycoin), "AddMarkerAccount() error")

	hasnavcoin := markertypes.NewEmptyMarkerAccount("hasnavcoin",
		address1.String(),
		[]markertypes.AccessGrant{})
	hasnavcoin.Supply = sdk.NewInt(100)
	s.Require().NoError(s.app.MarkerKeeper.AddMarkerAccount(s.ctx, hasnavcoin), "AddMarkerAccount() error")
	presentnav := markertypes.NewNetAssetValue(sdk.NewInt64Coin(markertypes.UsdDenom, int64(55)), uint64(100))
	s.Require().NoError(s.app.MarkerKeeper.AddSetNetAssetValues(s.ctx, hasnavcoin, []markertypes.NetAssetValue{presentnav}, "test"))

	addMarkerNavs(s.ctx, s.app, map[string]markertypes.NetAssetValue{
		"testcoininlist": {Price: sdk.NewInt64Coin(markertypes.UsdDenom, int64(12345)), Volume: uint64(1)},
	})

	tests := []struct {
		name       string
		markerAddr sdk.AccAddress
		expNav     *markertypes.NetAssetValue
	}{
		{
			name:       "already has nav",
			markerAddr: hasnavcoin.GetAddress(),
			expNav:     &markertypes.NetAssetValue{Price: sdk.NewInt64Coin(markertypes.UsdDenom, int64(55)), Volume: uint64(100)},
		},
		{
			name:       "nav add fails for coin",
			markerAddr: nosupplycoin.GetAddress(),
			expNav:     nil,
		},
		{
			name:       "nav set from custom config",
			markerAddr: testcoinInList.GetAddress(),
			expNav:     &markertypes.NetAssetValue{Price: sdk.NewInt64Coin(markertypes.UsdDenom, int64(12345)), Volume: uint64(1)},
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			netAssetValues := []markertypes.NetAssetValue{}
			err := s.app.MarkerKeeper.IterateNetAssetValues(s.ctx, tc.markerAddr, func(state markertypes.NetAssetValue) (stop bool) {
				netAssetValues = append(netAssetValues, state)
				return false
			})
			s.Require().NoError(err, "IterateNetAssetValues err")
			if tc.expNav != nil {
				s.Assert().Len(netAssetValues, 1, "Should be 1 nav set for testcoin")
				s.Assert().Equal(tc.expNav.Price, netAssetValues[0].Price, "Net asset value price should equal default upgraded price")
				s.Assert().Equal(tc.expNav.Volume, netAssetValues[0].Volume, "Net asset value volume should equal 1")
			} else {
				s.Assert().Len(netAssetValues, 0, "Marker not expected to have nav")
			}
		})
	}
}

func (s *UpgradeTestSuite) TestSetExchangeParams() {
	startMsg := "INF Ensuring exchange module params are set."
	noopMsg := "INF Exchange module params are already defined."
	updateMsg := "INF Setting exchange module params to defaults."
	doneMsg := "INF Done ensuring exchange module params are set."

	tests := []struct {
		name           string
		existingParams *exchange.Params
		expectedParams *exchange.Params
		expInLog       []string
	}{
		{
			name:           "no params set yet",
			existingParams: nil,
			expectedParams: exchange.DefaultParams(),
			expInLog:       []string{startMsg, updateMsg, doneMsg},
		},
		{
			name:           "params set with no splits and default zero",
			existingParams: &exchange.Params{DefaultSplit: 0},
			expectedParams: &exchange.Params{DefaultSplit: 0},
			expInLog:       []string{startMsg, noopMsg, doneMsg},
		},
		{
			name:           "params set with no splits and different default",
			existingParams: &exchange.Params{DefaultSplit: exchange.DefaultDefaultSplit + 100},
			expectedParams: &exchange.Params{DefaultSplit: exchange.DefaultDefaultSplit + 100},
			expInLog:       []string{startMsg, noopMsg, doneMsg},
		},
		{
			name: "params set with some splits",
			existingParams: &exchange.Params{
				DefaultSplit: exchange.DefaultDefaultSplit + 100,
				DenomSplits: []exchange.DenomSplit{
					{Denom: "peach", Split: 3000},
					{Denom: "plum", Split: 100},
				},
			},
			expectedParams: &exchange.Params{
				DefaultSplit: exchange.DefaultDefaultSplit + 100,
				DenomSplits: []exchange.DenomSplit{
					{Denom: "peach", Split: 3000},
					{Denom: "plum", Split: 100},
				},
			},
			expInLog: []string{startMsg, noopMsg, doneMsg},
		},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.app.ExchangeKeeper.SetParams(s.ctx, tc.existingParams)

			// Reset the log buffer and call the fun. Relog the output if it panics.
			s.logBuffer.Reset()
			testFunc := func() {
				setExchangeParams(s.ctx, s.app)
			}
			didNotPanic := s.Assert().NotPanics(testFunc, "setExchangeParams")
			logOutput := s.GetLogOutput("setExchangeParams")
			if !didNotPanic {
				return
			}

			// Make sure the log has the expected lines.
			s.AssertLogContents(logOutput, tc.expInLog, nil, true, "setExchangeParams")

			// Make sure the params are as expected now.
			params := s.app.ExchangeKeeper.GetParams(s.ctx)
			s.Assert().Equal(tc.expectedParams, params, "params after setExchangeParams")
		})
	}
}

func (s *UpgradeTestSuite) TestConvertNAVUnits() {
	tests := []struct {
		name       string
		markerNavs []sdk.Coins
		expected   []sdk.Coins
	}{
		{
			name:       "should work with no markers",
			markerNavs: []sdk.Coins{},
			expected:   []sdk.Coins{},
		},
		{
			name: "should work with one marker no usd denom",
			markerNavs: []sdk.Coins{
				sdk.NewCoins(sdk.NewInt64Coin("jackthecat", 1)),
			},
			expected: []sdk.Coins{
				sdk.NewCoins(sdk.NewInt64Coin("jackthecat", 1)),
			},
		},
		{
			name: "should work with multiple markers no usd denom",
			markerNavs: []sdk.Coins{
				sdk.NewCoins(sdk.NewInt64Coin("jackthecat", 1)),
				sdk.NewCoins(sdk.NewInt64Coin("georgethedog", 2)),
			},
			expected: []sdk.Coins{
				sdk.NewCoins(sdk.NewInt64Coin("jackthecat", 1)),
				sdk.NewCoins(sdk.NewInt64Coin("georgethedog", 2)),
			},
		},
		{
			name: "should work with one marker with usd denom",
			markerNavs: []sdk.Coins{
				sdk.NewCoins(sdk.NewInt64Coin("jackthecat", 1), sdk.NewInt64Coin(markertypes.UsdDenom, 2)),
			},
			expected: []sdk.Coins{
				sdk.NewCoins(sdk.NewInt64Coin("jackthecat", 1), sdk.NewInt64Coin(markertypes.UsdDenom, 20)),
			},
		},
		{
			name: "should work with multiple markers with usd denom",
			markerNavs: []sdk.Coins{
				sdk.NewCoins(sdk.NewInt64Coin("jackthecat", 1), sdk.NewInt64Coin(markertypes.UsdDenom, 3)),
				sdk.NewCoins(sdk.NewInt64Coin("georgethedog", 2), sdk.NewInt64Coin(markertypes.UsdDenom, 4)),
			},
			expected: []sdk.Coins{
				sdk.NewCoins(sdk.NewInt64Coin("jackthecat", 1), sdk.NewInt64Coin(markertypes.UsdDenom, 30)),
				sdk.NewCoins(sdk.NewInt64Coin("georgethedog", 2), sdk.NewInt64Coin(markertypes.UsdDenom, 40)),
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			// Create the marker
			for i, prices := range tc.markerNavs {
				address := sdk.AccAddress(fmt.Sprintf("marker%d", i))
				marker := markertypes.NewEmptyMarkerAccount(fmt.Sprintf("coin%d", i), address.String(), []markertypes.AccessGrant{})
				marker.Supply = sdk.OneInt()
				s.Require().NoError(s.app.MarkerKeeper.AddMarkerAccount(s.ctx, marker), "AddMarkerAccount() error")

				var navs []markertypes.NetAssetValue
				for _, price := range prices {
					navs = append(navs, markertypes.NewNetAssetValue(price, uint64(1)))
					navAddr := sdk.AccAddress(price.Denom)
					if acc, _ := s.app.MarkerKeeper.GetMarkerByDenom(s.ctx, price.Denom); acc == nil {
						navMarker := markertypes.NewEmptyMarkerAccount(price.Denom, navAddr.String(), []markertypes.AccessGrant{})
						navMarker.Supply = sdk.OneInt()
						s.Require().NoError(s.app.MarkerKeeper.AddMarkerAccount(s.ctx, navMarker), "AddMarkerAccount() error")
					}
				}
				s.Require().NoError(s.app.MarkerKeeper.AddSetNetAssetValues(s.ctx, marker, navs, "AddSetNetAssetValues() error"))
			}

			// Test Logic
			convertNavUnits(s.ctx, s.app)
			for i := range tc.markerNavs {
				marker, err := s.app.MarkerKeeper.GetMarkerByDenom(s.ctx, fmt.Sprintf("coin%d", i))
				s.Require().NoError(err, "GetMarkerByDenom() error")
				var prices sdk.Coins

				s.app.MarkerKeeper.IterateNetAssetValues(s.ctx, marker.GetAddress(), func(state markertypes.NetAssetValue) (stop bool) {
					prices = append(prices, state.Price)
					return false
				})
				s.Require().EqualValues(tc.expected[i], prices, "should update prices correctly for nav")
			}

			// Destroy the marker
			for i, prices := range tc.markerNavs {
				coin := fmt.Sprintf("coin%d", i)
				marker, err := s.app.MarkerKeeper.GetMarkerByDenom(s.ctx, coin)
				s.Require().NoError(err, "GetMarkerByDenom() error")
				s.app.MarkerKeeper.RemoveMarker(s.ctx, marker)

				// We need to remove the nav markers
				for _, price := range prices {
					if navMarker, _ := s.app.MarkerKeeper.GetMarkerByDenom(s.ctx, price.Denom); navMarker != nil {
						s.app.MarkerKeeper.RemoveMarker(s.ctx, navMarker)
					}
				}
			}
		})
	}
}

func (s *UpgradeTestSuite) TestUpdateMsgFeesNhashPerMil() {
	origParams := s.app.MsgFeesKeeper.GetParams(s.ctx)
	defer s.app.MsgFeesKeeper.SetParams(s.ctx, origParams)

	tests := []struct {
		name           string
		existingParams msgfeestypes.Params
	}{
		{
			name: "from 0",
			existingParams: msgfeestypes.Params{
				FloorGasPrice:      sdk.NewInt64Coin("cat", 50),
				NhashPerUsdMil:     0,
				ConversionFeeDenom: "horse",
			},
		},
		{
			name: "from 25,000,000",
			existingParams: msgfeestypes.Params{
				FloorGasPrice:      sdk.NewInt64Coin("dog", 12),
				NhashPerUsdMil:     25_000_000,
				ConversionFeeDenom: "pig",
			},
		},
		{
			name: "from 40,000,000",
			existingParams: msgfeestypes.Params{
				FloorGasPrice:      sdk.NewInt64Coin("fish", 83),
				NhashPerUsdMil:     40_000_000,
				ConversionFeeDenom: "cow",
			},
		},
		{
			name: "from 123,456,789",
			existingParams: msgfeestypes.Params{
				FloorGasPrice:      sdk.NewInt64Coin("hamster", 83),
				NhashPerUsdMil:     123_456_789,
				ConversionFeeDenom: "llama",
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.app.MsgFeesKeeper.SetParams(s.ctx, tc.existingParams)
			expParams := tc.existingParams
			expParams.NhashPerUsdMil = 40_000_000
			expInLog := []string{
				"INF Setting MsgFees Params NhashPerUsdMil to 40000000.",
				"INF Done setting MsgFees Params NhashPerUsdMil.",
			}

			// Reset the log buffer and call the func. Relog the output if it panics.
			s.logBuffer.Reset()
			testFunc := func() {
				updateMsgFeesNhashPerMil(s.ctx, s.app)
			}
			didNotPanic := s.Assert().NotPanics(testFunc, "updateMsgFeesNhashPerMil")
			logOutput := s.GetLogOutput("updateMsgFeesNhashPerMil")
			if !didNotPanic {
				return
			}
			s.AssertLogContents(logOutput, expInLog, nil, true, "updateMsgFeesNhashPerMil")

			actParams := s.app.MsgFeesKeeper.GetParams(s.ctx)
			s.Assert().Equal(expParams, actParams, "MsgFeesKeeper Params after updateMsgFeesNhashPerMil")
		})
	}
}

func (s *UpgradeTestSuite) TestSetExchangePaymentParamsToDefaults() {
	origParams := s.app.ExchangeKeeper.GetParams(s.ctx)
	defer s.app.ExchangeKeeper.SetParams(s.ctx, origParams)

	defaultParams := exchange.DefaultParams()

	tests := []struct {
		name           string
		existingParams *exchange.Params
		expectedParams *exchange.Params
	}{
		{
			name:           "no params set yet",
			existingParams: nil,
			expectedParams: defaultParams,
		},
		{
			name: "params set previously",
			existingParams: &exchange.Params{
				DefaultSplit:         333,
				DenomSplits:          []exchange.DenomSplit{{Denom: "date", Split: 99}},
				FeeCreatePaymentFlat: nil,
				FeeAcceptPaymentFlat: nil,
			},
			expectedParams: &exchange.Params{
				DefaultSplit:         333,
				DenomSplits:          []exchange.DenomSplit{{Denom: "date", Split: 99}},
				FeeCreatePaymentFlat: defaultParams.FeeCreatePaymentFlat,
				FeeAcceptPaymentFlat: defaultParams.FeeAcceptPaymentFlat,
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			s.app.ExchangeKeeper.SetParams(s.ctx, tc.existingParams)
			expInLog := []string{
				"INF Setting exchange module payment params to defaults.",
				"INF Done setting exchange module payment params to defaults.",
			}

			// Reset the log buffer and call the func. Relog the output if it panics.
			s.logBuffer.Reset()
			testFunc := func() {
				setExchangePaymentParamsToDefaults(s.ctx, s.app)
			}
			didNotPanic := s.Assert().NotPanics(testFunc, "setExchangePaymentParamsToDefaults")
			logOutput := s.GetLogOutput("setExchangePaymentParamsToDefaults")
			if !didNotPanic {
				return
			}
			s.AssertLogContents(logOutput, expInLog, nil, true, "setExchangePaymentParamsToDefaults")

			actParams := s.app.ExchangeKeeper.GetParams(s.ctx)
			s.Assert().Equal(tc.expectedParams, actParams, "Exchange Params after setExchangePaymentParamsToDefaults")
		})
	}
}
