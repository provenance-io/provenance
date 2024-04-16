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

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/provenance-io/provenance/internal/helpers"
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
	s.ctx = s.app.BaseApp.NewContextLegacy(false, cmtproto.Header{Time: s.startTime})
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
	origVersionMap, err := s.app.UpgradeKeeper.GetModuleVersionMap(s.ctx)
	s.Require().NoError(err, "GetModuleVersionMap")

	msgFormat := fmt.Sprintf("upgrades[%q].Handler(...)", key)

	var versionMap module.VersionMap
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
	s.LogIfError(testutil.FundAccount(s.ctx, s.app.BankKeeper, addr2, sdk.Coins{coin}), "FundAccount(..., %q)", coin.String())
	return addr2
}

// CreateValidator creates a new validator in the app.
func (s *UpgradeTestSuite) CreateValidator(unbondedTime time.Time, status stakingtypes.BondStatus) stakingtypes.Validator {
	key := secp256k1.GenPrivKey()
	pub := key.PubKey()
	addr := sdk.AccAddress(pub.Address())
	valAddr := sdk.ValAddress(addr)
	validator, err := stakingtypes.NewValidator(valAddr.String(), pub, stakingtypes.NewDescription(valAddr.String(), "", "", "", ""))
	s.Require().NoError(err, "could not init new validator")
	validator.UnbondingTime = unbondedTime
	validator.Status = status
	err = s.app.StakingKeeper.SetValidator(s.ctx, validator)
	s.Require().NoError(err, "could not SetValidator ")
	err = s.app.StakingKeeper.SetValidatorByConsAddr(s.ctx, validator)
	s.Require().NoError(err, "could not SetValidatorByConsAddr ")
	err = s.app.StakingKeeper.Hooks().AfterValidatorCreated(s.ctx, valAddr)
	s.Require().NoError(err, "could not AfterValidatorCreated")
	return validator
}

// DelegateToValidator delegates to a validator in the app.
func (s *UpgradeTestSuite) DelegateToValidator(valAddress sdk.ValAddress, delegatorAddress sdk.AccAddress, coin sdk.Coin) {
	validator, err := s.app.StakingKeeper.GetValidator(s.ctx, valAddress)
	s.Require().NoError(err, "GetValidator(%q)", valAddress.String())
	_, err = s.app.StakingKeeper.Delegate(s.ctx, delegatorAddress, coin.Amount, types.Unbonded, validator, true)
	s.Require().NoError(err, "Delegate(%q, %q, unbonded, %q, true) error", delegatorAddress.String(), coin.String(), valAddress.String())
}

func (s *UpgradeTestSuite) GetOperatorAddr(val stakingtypes.ValidatorI) sdk.ValAddress {
	addr, err := helpers.GetOperatorAddr(val)
	s.Require().NoError(err, "GetOperatorAddr(%q)", val.GetOperator())
	return addr
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

func (s *UpgradeTestSuite) TestUmberRC1() {
	expInLog := []string{
		"INF Pruning expired consensus states for IBC.",
		"INF Done pruning expired consensus states for IBC.",
		"INF Migrating legacy params.",
		"INF Done migrating legacy params.",
		"INF Migrating attribute params.",
		"INF Done migrating attribute params.",
		"INF Migrating marker params.",
		"INF Done migrating marker params.",
		"INF Migrating metadata os locator params.",
		"INF Done migrating metadata os locator params.",
		"INF Starting module migrations. This may take a significant amount of time to complete. Do not restart node.",
		"INF Updating IBC AllowedClients.",
		"INF Done updating IBC AllowedClients.",
		"INF Removing inactive validator delegations.",
		"INF Threshold: 21 days",
		"INF A total of 0 inactive (unbonded) validators have had all their delegators removed.",
	}

	s.AssertUpgradeHandlerLogs("umber-rc1", expInLog, nil)
}

func (s *UpgradeTestSuite) TestUmber() {
	expInLog := []string{
		"INF Pruning expired consensus states for IBC.",
		"INF Done pruning expired consensus states for IBC.",
		"INF Migrating legacy params.",
		"INF Done migrating legacy params.",
		"INF Migrating attribute params.",
		"INF Done migrating attribute params.",
		"INF Migrating marker params.",
		"INF Done migrating marker params.",
		"INF Migrating metadata os locator params.",
		"INF Done migrating metadata os locator params.",
		"INF Starting module migrations. This may take a significant amount of time to complete. Do not restart node.",
		"INF Updating IBC AllowedClients.",
		"INF Done updating IBC AllowedClients.",
		"INF Removing inactive validator delegations.",
		"INF Threshold: 21 days",
		"INF A total of 0 inactive (unbonded) validators have had all their delegators removed.",
	}

	s.AssertUpgradeHandlerLogs("umber", expInLog, nil)
}

func (s *UpgradeTestSuite) TestRemoveInactiveValidatorDelegations() {
	addr1 := s.CreateAndFundAccount(sdk.NewInt64Coin("stake", 1000000))
	addr2 := s.CreateAndFundAccount(sdk.NewInt64Coin("stake", 1000000))

	runner := func() {
		removeInactiveValidatorDelegations(s.ctx, s.app)
	}
	runnerName := "removeInactiveValidatorDelegations"

	delegationCoin := sdk.NewInt64Coin("stake", 10000)
	delegationCoinAmt := sdkmath.LegacyNewDec(delegationCoin.Amount.Int64())

	s.Run("just one bonded validator", func() {
		validators, err := s.app.StakingKeeper.GetAllValidators(s.ctx)
		s.Require().NoError(err, "GetAllValidators")
		s.Require().Len(validators, 1, "GetAllValidators after setup")

		expectedLogLines := []string{
			"INF Removing all delegations from validators that have been inactive (unbonded) for 21 days.",
			"INF A total of 0 inactive (unbonded) validators have had all their delegators removed.",
		}
		s.ExecuteAndAssertLogs(runner, expectedLogLines, nil, true, runnerName)

		newValidators, err := s.app.StakingKeeper.GetAllValidators(s.ctx)
		s.Require().NoError(err, "GetAllValidators")
		s.Require().Len(newValidators, 1, "GetAllValidators after %s", runnerName)
	})

	s.Run("one unbonded validator with a delegation", func() {
		// single unbonded validator with 1 delegations, should be removed
		unbondedVal1 := s.CreateValidator(s.startTime.Add(-30*24*time.Hour), stakingtypes.Unbonded)
		unbondedVal1Addr := s.GetOperatorAddr(unbondedVal1)
		addr1Balance := s.app.BankKeeper.GetBalance(s.ctx, addr1, "stake")
		s.DelegateToValidator(unbondedVal1Addr, addr1, delegationCoin)
		s.Require().Equal(addr1Balance.Sub(delegationCoin), s.app.BankKeeper.GetBalance(s.ctx, addr1, "stake"), "addr1 should have less funds")
		validators, err := s.app.StakingKeeper.GetAllValidators(s.ctx)
		s.Require().NoError(err, "GetAllValidators")
		s.Require().Len(validators, 2, "Setup: GetAllValidators should have: 1 bonded, 1 unbonded")

		expectedLogLines := []string{
			"INF Removing all delegations from validators that have been inactive (unbonded) for 21 days.",
			fmt.Sprintf("INF Validator %v has been inactive (unbonded) for %d days and will be removed.", unbondedVal1.OperatorAddress, 30),
			fmt.Sprintf("INF Undelegate delegator %v from validator %v of all shares (%v).", addr1.String(), unbondedVal1.OperatorAddress, delegationCoinAmt),
			"INF A total of 1 inactive (unbonded) validators have had all their delegators removed.",
		}
		s.ExecuteAndAssertLogs(runner, expectedLogLines, nil, true, runnerName)

		validators, err = s.app.StakingKeeper.GetAllValidators(s.ctx)
		s.Require().NoError(err, "GetAllValidators")
		s.Assert().Len(validators, 1, "GetAllValidators after %s", runnerName)
		ubd, err := s.app.StakingKeeper.GetUnbondingDelegation(s.ctx, addr1, unbondedVal1Addr)
		s.Assert().NoError(err, "GetUnbondingDelegation found")
		if s.Assert().Len(ubd.Entries, 1, "UnbondingDelegation entries") {
			s.Assert().Equal(delegationCoin.Amount, ubd.Entries[0].Balance, "UnbondingDelegation balance")
		}
	})

	s.Run("one unbonded validator with 2 delegations", func() {
		// single unbonded validator with 2 delegations
		unbondedVal1 := s.CreateValidator(s.startTime.Add(-30*24*time.Hour), stakingtypes.Unbonded)
		unbondedVal1Addr := s.GetOperatorAddr(unbondedVal1)
		addr1Balance := s.app.BankKeeper.GetBalance(s.ctx, addr1, "stake")
		addr2Balance := s.app.BankKeeper.GetBalance(s.ctx, addr2, "stake")
		s.DelegateToValidator(unbondedVal1Addr, addr1, delegationCoin)
		s.DelegateToValidator(unbondedVal1Addr, addr2, delegationCoin)
		s.Require().Equal(addr1Balance.Sub(delegationCoin), s.app.BankKeeper.GetBalance(s.ctx, addr1, "stake"), "addr1 should have less funds after delegation")
		s.Require().Equal(addr2Balance.Sub(delegationCoin), s.app.BankKeeper.GetBalance(s.ctx, addr2, "stake"), "addr2 should have less funds after delegation")
		var err error
		unbondedVal1, err = s.app.StakingKeeper.GetValidator(s.ctx, unbondedVal1Addr)
		s.Require().NoError(err, "Setup: GetValidator(unbondedVal1) found")
		s.Require().Equal(delegationCoinAmt.Add(delegationCoinAmt), unbondedVal1.DelegatorShares, "Setup: unbondedVal1.DelegatorShares")
		validators, err := s.app.StakingKeeper.GetAllValidators(s.ctx)
		s.Require().NoError(err, "GetAllValidators")
		s.Require().Len(validators, 2, "Setup: GetAllValidators should have: 1 bonded, 1 unbonded")

		expectedLogLines := []string{
			"INF Removing all delegations from validators that have been inactive (unbonded) for 21 days.",
			fmt.Sprintf("INF Validator %v has been inactive (unbonded) for %d days and will be removed.", unbondedVal1.OperatorAddress, 30),
			fmt.Sprintf("INF Undelegate delegator %v from validator %v of all shares (%v).", addr1.String(), unbondedVal1.OperatorAddress, delegationCoinAmt),
			fmt.Sprintf("INF Undelegate delegator %v from validator %v of all shares (%v).", addr2.String(), unbondedVal1.OperatorAddress, delegationCoinAmt),
			"INF A total of 1 inactive (unbonded) validators have had all their delegators removed.",
		}
		s.ExecuteAndAssertLogs(runner, expectedLogLines, nil, false, runnerName)

		validators, err = s.app.StakingKeeper.GetAllValidators(s.ctx)
		s.Require().NoError(err, "GetAllValidators")
		s.Assert().Len(validators, 1, "GetAllValidators after %s", runnerName)
		ubd, err := s.app.StakingKeeper.GetUnbondingDelegation(s.ctx, addr1, unbondedVal1Addr)
		s.Assert().NoError(err, "GetUnbondingDelegation(addr1) found")
		if s.Assert().Len(ubd.Entries, 1, "GetUnbondingDelegation(addr1) entries") {
			s.Assert().Equal(delegationCoin.Amount, ubd.Entries[0].Balance, "GetUnbondingDelegation(addr1) balance")
		}
		ubd, err = s.app.StakingKeeper.GetUnbondingDelegation(s.ctx, addr2, unbondedVal1Addr)
		s.Assert().NoError(err, "GetUnbondingDelegation(addr2) found")
		if s.Assert().Len(ubd.Entries, 1, "GetUnbondingDelegation(addr2) entries") {
			s.Assert().Equal(delegationCoin.Amount, ubd.Entries[0].Balance, "GetUnbondingDelegation(addr2) balance")
		}
	})

	s.Run("two unbonded validators to be removed", func() {
		// 2 unbonded validators with delegations past inactive time, both should be removed
		unbondedVal1 := s.CreateValidator(s.startTime.Add(-30*24*time.Hour), stakingtypes.Unbonded)
		unbondedVal1Addr := s.GetOperatorAddr(unbondedVal1)
		s.DelegateToValidator(unbondedVal1Addr, addr1, delegationCoin)
		s.DelegateToValidator(unbondedVal1Addr, addr2, delegationCoin)
		var err error
		unbondedVal1, err = s.app.StakingKeeper.GetValidator(s.ctx, unbondedVal1Addr)
		s.Require().NoError(err, "Setup: GetValidator(unbondedVal1) found")
		s.Require().Equal(delegationCoinAmt.Add(delegationCoinAmt), unbondedVal1.DelegatorShares, "Setup: shares delegated to unbondedVal1")

		unbondedVal2 := s.CreateValidator(s.startTime.Add(-29*24*time.Hour), stakingtypes.Unbonded)
		unbondedVal2Addr := s.GetOperatorAddr(unbondedVal2)
		s.DelegateToValidator(unbondedVal2Addr, addr1, delegationCoin)
		s.DelegateToValidator(unbondedVal2Addr, addr2, delegationCoin)
		unbondedVal2, err = s.app.StakingKeeper.GetValidator(s.ctx, unbondedVal2Addr)
		s.Require().NoError(err, "Setup: GetValidator(unbondedVal2) found")
		s.Require().Equal(delegationCoinAmt.Add(delegationCoinAmt), unbondedVal2.DelegatorShares, "Setup: shares delegated to unbondedVal2")
		validators, err := s.app.StakingKeeper.GetAllValidators(s.ctx)
		s.Require().NoError(err, "GetAllValidators")
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

		validators, err = s.app.StakingKeeper.GetAllValidators(s.ctx)
		s.Require().NoError(err, "GetAllValidators")
		s.Assert().Len(validators, 1, "GetAllValidators after %s", runnerName)
		ubd, err := s.app.StakingKeeper.GetUnbondingDelegation(s.ctx, addr1, unbondedVal1Addr)
		s.Assert().NoError(err, "GetUnbondingDelegation(addr1) found")
		if s.Assert().Len(ubd.Entries, 1, "GetUnbondingDelegation(addr1) entries") {
			s.Assert().Equal(delegationCoin.Amount, ubd.Entries[0].Balance, "GetUnbondingDelegation(addr1) balance")
		}
		ubd, err = s.app.StakingKeeper.GetUnbondingDelegation(s.ctx, addr2, unbondedVal1Addr)
		s.Assert().NoError(err, "GetUnbondingDelegation(addr2) found")
		if s.Assert().Len(ubd.Entries, 1, "GetUnbondingDelegation(addr2) entries") {
			s.Assert().Equal(delegationCoin.Amount, ubd.Entries[0].Balance, "GetUnbondingDelegation(addr2) balance")
		}
	})

	s.Run("two unbonded validators one too recently", func() {
		// 2 unbonded validators, 1 under the inactive day count, should only remove one
		unbondedVal1 := s.CreateValidator(s.startTime.Add(-30*24*time.Hour), stakingtypes.Unbonded)
		unbondedVal1Addr := s.GetOperatorAddr(unbondedVal1)
		s.DelegateToValidator(unbondedVal1Addr, addr1, delegationCoin)
		s.DelegateToValidator(unbondedVal1Addr, addr2, delegationCoin)
		var err error
		unbondedVal1, err = s.app.StakingKeeper.GetValidator(s.ctx, unbondedVal1Addr)
		s.Require().NoError(err, "Setup: GetValidator(unbondedVal1) found")
		s.Require().Equal(delegationCoinAmt.Add(delegationCoinAmt), unbondedVal1.DelegatorShares, "Setup: shares delegated to unbondedVal1")

		unbondedVal2 := s.CreateValidator(s.startTime.Add(-20*24*time.Hour), stakingtypes.Unbonded)
		unbondedVal2Addr := s.GetOperatorAddr(unbondedVal2)
		s.DelegateToValidator(unbondedVal2Addr, addr1, delegationCoin)
		unbondedVal2, err = s.app.StakingKeeper.GetValidator(s.ctx, unbondedVal2Addr)
		s.Require().NoError(err, "Setup: GetValidator(unbondedVal2) found")
		s.Require().Equal(delegationCoinAmt, unbondedVal2.DelegatorShares, "Setup: shares delegated to unbondedVal1")
		validators, err := s.app.StakingKeeper.GetAllValidators(s.ctx)
		s.Require().NoError(err, "GetAllValidators")
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

		validators, err = s.app.StakingKeeper.GetAllValidators(s.ctx)
		s.Require().NoError(err, "GetAllValidators")
		s.Assert().Len(validators, 2, "GetAllValidators after %s", runnerName)
		ubd, err := s.app.StakingKeeper.GetUnbondingDelegation(s.ctx, addr1, unbondedVal1Addr)
		s.Assert().NoError(err, "GetUnbondingDelegation(addr1) found")
		if s.Assert().Len(ubd.Entries, 1, "GetUnbondingDelegation(addr1) entries") {
			s.Assert().Equal(delegationCoin.Amount, ubd.Entries[0].Balance, "GetUnbondingDelegation(addr1) balance")
		}
		ubd, err = s.app.StakingKeeper.GetUnbondingDelegation(s.ctx, addr2, unbondedVal1Addr)
		s.Assert().NoError(err, "GetUnbondingDelegation(addr2) found")
		if s.Assert().Len(ubd.Entries, 1, "GetUnbondingDelegation(addr2) entries") {
			s.Assert().Equal(delegationCoin.Amount, ubd.Entries[0].Balance, "GetUnbondingDelegation(addr2) balance")
		}
	})

	s.Run("unbonded without delegators", func() {
		// create a unbonded validator with out delegators, should not remove
		unbondedVal1 := s.CreateValidator(s.startTime.Add(-30*24*time.Hour), stakingtypes.Unbonded)
		validators, err := s.app.StakingKeeper.GetAllValidators(s.ctx)
		s.Require().NoError(err, "GetAllValidators")
		s.Require().Len(validators, 3, "Setup: GetAllValidators should have: 1 bonded, 1 recently unbonded, 1 empty unbonded")

		expectedLogLines := []string{
			"INF Removing all delegations from validators that have been inactive (unbonded) for 21 days.",
			fmt.Sprintf("INF Validator %v has been inactive (unbonded) for %d days and will be removed.", unbondedVal1.OperatorAddress, 30),
			"INF A total of 1 inactive (unbonded) validators have had all their delegators removed.",
		}
		s.ExecuteAndAssertLogs(runner, expectedLogLines, nil, true, runnerName)

		validators, err = s.app.StakingKeeper.GetAllValidators(s.ctx)
		s.Require().NoError(err, "GetAllValidators")
		s.Assert().Len(validators, 3, "GetAllValidators after %s", runnerName)
	})
}
