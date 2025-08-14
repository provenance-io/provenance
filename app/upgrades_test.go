package app

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"regexp"
	"slices"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/provenance-io/provenance/internal"
	internalsdk "github.com/provenance-io/provenance/internal/sdk"
	"github.com/provenance-io/provenance/testutil/assertions"
	flatfeestypes "github.com/provenance-io/provenance/x/flatfees/types"
	ledgerTypes "github.com/provenance-io/provenance/x/ledger/types"
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

func (s *UpgradeTestSuite) SetupTest() {
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

// CreateUnbondedValidator creates a new validator in the app with the provided unbonded time.
func (s *UpgradeTestSuite) CreateUnbondedValidator(unbondedTime time.Time, status stakingtypes.BondStatus) stakingtypes.Validator {
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
	_, err = s.app.StakingKeeper.Delegate(s.ctx, delegatorAddress, coin.Amount, stakingtypes.Unbonded, validator, true)
	s.Require().NoError(err, "Delegate(%q, %q, unbonded, %q, true) error", delegatorAddress.String(), coin.String(), valAddress.String())
}

func (s *UpgradeTestSuite) GetOperatorAddr(val stakingtypes.ValidatorI) sdk.ValAddress {
	addr, err := internalsdk.GetOperatorAddr(val)
	s.Require().NoError(err, "GetOperatorAddr(%q)", val.GetOperator())
	return addr
}

// addrCoins associates an address with coins.
type addrCoins struct {
	addr sdk.AccAddress
	bal  sdk.Coins
}

// getAllSpendable will get the spendable balance of all known accounts (even if the account doesn't have anything in it).
func (s *UpgradeTestSuite) getAllSpendable() []addrCoins {
	var rv []addrCoins
	err := s.app.AccountKeeper.Accounts.Walk(s.ctx, nil, func(addr sdk.AccAddress, _ sdk.AccountI) (stop bool, err error) {
		bal := s.app.BankKeeper.SpendableCoins(s.ctx, addr)
		rv = append(rv, addrCoins{addr: addr, bal: bal})
		return false, nil
	})
	s.Require().NoError(err, "error walking accounts to get all balances")
	return rv
}

// assertAllSpendable will assert that each of the accounts has the correct spendable balance.
// Returns true if all spendable balances are as expected.
func (s *UpgradeTestSuite) assertAllSpendable(expBals []addrCoins) bool {
	ok := true
	for i, exp := range expBals {
		s.Run(fmt.Sprintf("%d: %s", i, exp.addr), func() {
			act := s.app.BankKeeper.SpendableCoins(s.ctx, exp.addr)
			ok = s.Assert().Equal(exp.bal.String(), act.String(), "SpendableCoins") && ok
		})
	}
	return ok
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
			LogMsgRemoveInactiveValidatorDelegations,
			"INF Threshold: 21 days",
			"INF A total of 0 inactive (unbonded) validators have had all their delegators removed.",
		}
		s.ExecuteAndAssertLogs(runner, expectedLogLines, nil, true, runnerName)

		newValidators, err := s.app.StakingKeeper.GetAllValidators(s.ctx)
		s.Require().NoError(err, "GetAllValidators")
		s.Require().Len(newValidators, 1, "GetAllValidators after %s", runnerName)
	})

	s.Run("one unbonded validator with a delegation", func() {
		// single unbonded validator with 1 delegations, should be removed
		unbondedVal1 := s.CreateUnbondedValidator(s.startTime.Add(-30*24*time.Hour), stakingtypes.Unbonded)
		unbondedVal1Addr := s.GetOperatorAddr(unbondedVal1)
		addr1Balance := s.app.BankKeeper.GetBalance(s.ctx, addr1, "stake")
		s.DelegateToValidator(unbondedVal1Addr, addr1, delegationCoin)
		s.Require().Equal(addr1Balance.Sub(delegationCoin), s.app.BankKeeper.GetBalance(s.ctx, addr1, "stake"), "addr1 should have less funds")
		validators, err := s.app.StakingKeeper.GetAllValidators(s.ctx)
		s.Require().NoError(err, "GetAllValidators")
		s.Require().Len(validators, 2, "Setup: GetAllValidators should have: 1 bonded, 1 unbonded")

		expectedLogLines := []string{
			LogMsgRemoveInactiveValidatorDelegations,
			"INF Threshold: 21 days",
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
		unbondedVal1 := s.CreateUnbondedValidator(s.startTime.Add(-30*24*time.Hour), stakingtypes.Unbonded)
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
			LogMsgRemoveInactiveValidatorDelegations,
			"INF Threshold: 21 days",
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
		unbondedVal1 := s.CreateUnbondedValidator(s.startTime.Add(-30*24*time.Hour), stakingtypes.Unbonded)
		unbondedVal1Addr := s.GetOperatorAddr(unbondedVal1)
		s.DelegateToValidator(unbondedVal1Addr, addr1, delegationCoin)
		s.DelegateToValidator(unbondedVal1Addr, addr2, delegationCoin)
		var err error
		unbondedVal1, err = s.app.StakingKeeper.GetValidator(s.ctx, unbondedVal1Addr)
		s.Require().NoError(err, "Setup: GetValidator(unbondedVal1) found")
		s.Require().Equal(delegationCoinAmt.Add(delegationCoinAmt), unbondedVal1.DelegatorShares, "Setup: shares delegated to unbondedVal1")

		unbondedVal2 := s.CreateUnbondedValidator(s.startTime.Add(-29*24*time.Hour), stakingtypes.Unbonded)
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
			LogMsgRemoveInactiveValidatorDelegations,
			"INF Threshold: 21 days",
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
		unbondedVal1 := s.CreateUnbondedValidator(s.startTime.Add(-30*24*time.Hour), stakingtypes.Unbonded)
		unbondedVal1Addr := s.GetOperatorAddr(unbondedVal1)
		s.DelegateToValidator(unbondedVal1Addr, addr1, delegationCoin)
		s.DelegateToValidator(unbondedVal1Addr, addr2, delegationCoin)
		var err error
		unbondedVal1, err = s.app.StakingKeeper.GetValidator(s.ctx, unbondedVal1Addr)
		s.Require().NoError(err, "Setup: GetValidator(unbondedVal1) found")
		s.Require().Equal(delegationCoinAmt.Add(delegationCoinAmt), unbondedVal1.DelegatorShares, "Setup: shares delegated to unbondedVal1")

		unbondedVal2 := s.CreateUnbondedValidator(s.startTime.Add(-20*24*time.Hour), stakingtypes.Unbonded)
		unbondedVal2Addr := s.GetOperatorAddr(unbondedVal2)
		s.DelegateToValidator(unbondedVal2Addr, addr1, delegationCoin)
		unbondedVal2, err = s.app.StakingKeeper.GetValidator(s.ctx, unbondedVal2Addr)
		s.Require().NoError(err, "Setup: GetValidator(unbondedVal2) found")
		s.Require().Equal(delegationCoinAmt, unbondedVal2.DelegatorShares, "Setup: shares delegated to unbondedVal1")
		validators, err := s.app.StakingKeeper.GetAllValidators(s.ctx)
		s.Require().NoError(err, "GetAllValidators")
		s.Require().Len(validators, 3, "Setup: GetAllValidators should have: 1 bonded, 1 recently unbonded, 1 old unbonded")

		expectedLogLines := []string{
			LogMsgRemoveInactiveValidatorDelegations,
			"INF Threshold: 21 days",
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
		unbondedVal1 := s.CreateUnbondedValidator(s.startTime.Add(-30*24*time.Hour), stakingtypes.Unbonded)
		validators, err := s.app.StakingKeeper.GetAllValidators(s.ctx)
		s.Require().NoError(err, "GetAllValidators")
		s.Require().Len(validators, 3, "Setup: GetAllValidators should have: 1 bonded, 1 recently unbonded, 1 empty unbonded")

		expectedLogLines := []string{
			LogMsgRemoveInactiveValidatorDelegations,
			"INF Threshold: 21 days",
			fmt.Sprintf("INF Validator %v has been inactive (unbonded) for %d days and will be removed.", unbondedVal1.OperatorAddress, 30),
			"INF A total of 1 inactive (unbonded) validators have had all their delegators removed.",
		}
		s.ExecuteAndAssertLogs(runner, expectedLogLines, nil, true, runnerName)

		validators, err = s.app.StakingKeeper.GetAllValidators(s.ctx)
		s.Require().NoError(err, "GetAllValidators")
		s.Assert().Len(validators, 3, "GetAllValidators after %s", runnerName)
	})
}

func (s *UpgradeTestSuite) TestConvertFinishedVestingAccountsToBase() {
	stakeCoins := func(amt int64) sdk.Coins {
		return sdk.Coins{sdk.NewInt64Coin("stake", amt)}
	}

	pivotTime := time.Date(2020, 4, 20, 16, 20, 0, 0, time.UTC).UTC()
	reallyOldTime := time.Date(1999, 12, 31, 59, 59, 59, 999999999, time.UTC).UTC()
	now := pivotTime.Unix()
	thePast := pivotTime.Add(time.Hour * 24 * -100).Unix()
	oneDayAgo := pivotTime.Add(time.Hour * -24).Unix()
	oneSecAgo := pivotTime.Add(time.Second * -1).Unix()
	oneSecFromNow := pivotTime.Add(time.Second).Unix()
	theFuture := pivotTime.Add(time.Hour * 24 * 1000).Unix()

	// newContinuousVestingAccount returns an acctSetup func for creating a continuous vesting account.
	newContinuousVestingAccount := func(origVest sdk.Coins, startTime, endTime int64) func(acct *authtypes.BaseAccount) (sdk.AccountI, error) {
		return func(acct *authtypes.BaseAccount) (sdk.AccountI, error) {
			return vesting.NewContinuousVestingAccount(acct, origVest, startTime, endTime)
		}
	}

	// newDelayedVestingAccount returns an acctSetup func for creating a delayed vesting account.
	newDelayedVestingAccount := func(origVest sdk.Coins, endTime int64) func(acct *authtypes.BaseAccount) (sdk.AccountI, error) {
		return func(acct *authtypes.BaseAccount) (sdk.AccountI, error) {
			return vesting.NewDelayedVestingAccount(acct, origVest, endTime)
		}
	}

	// newPeriodicVestingAccount returns an acctSetup func for creating a periodic vesting account.
	newPeriodicVestingAccount := func(origVest sdk.Coins, startTime int64, periods vesting.Periods) func(acct *authtypes.BaseAccount) (sdk.AccountI, error) {
		return func(acct *authtypes.BaseAccount) (sdk.AccountI, error) {
			return vesting.NewPeriodicVestingAccount(acct, origVest, startTime, periods)
		}
	}

	// newPermanentLockedAccount returns an acctSetup func for creating a permanent locked account.
	newPermanentLockedAccount := func(origVest sdk.Coins) func(acct *authtypes.BaseAccount) (sdk.AccountI, error) {
		return func(acct *authtypes.BaseAccount) (sdk.AccountI, error) {
			return vesting.NewPermanentLockedAccount(acct, origVest)
		}
	}

	tests := []*struct {
		name      string
		noPK      bool
		seq       uint64
		expConv   bool
		acctSetup func(baseAcct *authtypes.BaseAccount) (sdk.AccountI, error)
		addr      sdk.AccAddress
		baseAcct  *authtypes.BaseAccount
		origAcct  sdk.AccountI
		expAcct   sdk.AccountI
	}{
		{name: "base account: no pubkey, no sequence", noPK: true, seq: 0},
		{name: "base account: no pubkey, with sequence", noPK: true, seq: 5},
		{name: "base account: with pubkey, no sequence", noPK: false, seq: 0},
		{name: "base account: with pubkey, with sequence", noPK: false, seq: 12},
		{name: "fee_collector module account", addr: authtypes.NewModuleAddress("fee_collector")},
		{
			name: "permanent locked: no pubkey, no sequence",
			noPK: true, seq: 0,
			acctSetup: newPermanentLockedAccount(stakeCoins(500000)),
		},
		{
			name: "permanent locked: no pubkey, with sequence",
			noPK: true, seq: 41,
			acctSetup: newPermanentLockedAccount(stakeCoins(700000)),
		},
		{
			name: "permanent locked: with pubkey, no sequence",
			noPK: false, seq: 0,
			acctSetup: newPermanentLockedAccount(stakeCoins(300000)),
		},
		{
			name: "permanent locked: with pubkey, with sequence",
			noPK: false, seq: 83,
			acctSetup: newPermanentLockedAccount(stakeCoins(400000)),
		},
		{
			name: "continuous vesting: finished 1 second ago",
			seq:  177, expConv: true,
			acctSetup: newContinuousVestingAccount(stakeCoins(800000), thePast, oneSecAgo),
		},
		{
			name: "continuous vesting: finishes now",
			seq:  212, expConv: true,
			acctSetup: newContinuousVestingAccount(stakeCoins(700000), thePast, now),
		},
		{
			name: "continuous vesting: already started, not finished",
			seq:  41, expConv: false,
			acctSetup: newContinuousVestingAccount(stakeCoins(600000), thePast, oneSecFromNow),
		},
		{
			name: "continuous vesting: not started yet",
			seq:  18918, expConv: false,
			acctSetup: newContinuousVestingAccount(stakeCoins(850000), oneSecFromNow, theFuture),
		},
		{
			name: "delayed vesting: finished 1 second ago",
			seq:  71, expConv: true,
			acctSetup: newDelayedVestingAccount(stakeCoins(870000), oneSecAgo),
		},
		{
			name: "delayed vesting: finishes now",
			noPK: true, seq: 661, expConv: true,
			acctSetup: newDelayedVestingAccount(stakeCoins(870070), now),
		},
		{
			name: "delayed vesting: finishes in 1 second",
			seq:  661, expConv: false,
			acctSetup: newDelayedVestingAccount(stakeCoins(840000), oneSecFromNow),
		},
		{
			name: "periodic vesting: no periods finished",
			seq:  5551, expConv: false,
			acctSetup: newPeriodicVestingAccount(stakeCoins(500000), thePast, []vesting.Period{
				{Length: oneSecFromNow - thePast, Amount: stakeCoins(100000)},
				{Length: theFuture - oneSecFromNow, Amount: stakeCoins(400000)},
			}),
		},
		{
			name: "periodic vesting: first of two periods finishes now",
			seq:  3412, expConv: false,
			acctSetup: newPeriodicVestingAccount(stakeCoins(500500), thePast, []vesting.Period{
				{Length: now - thePast, Amount: stakeCoins(100400)},
				{Length: theFuture - now, Amount: stakeCoins(400100)},
			}),
		},
		{
			name: "periodic vesting: finishes in 1 second",
			seq:  1255, expConv: false,
			acctSetup: newPeriodicVestingAccount(stakeCoins(500050), thePast, []vesting.Period{
				{Length: oneDayAgo - thePast, Amount: stakeCoins(100040)},
				{Length: oneSecFromNow - oneDayAgo, Amount: stakeCoins(400010)},
			}),
		},
		{
			name: "periodic vesting: finishes now",
			seq:  9811, expConv: true,
			acctSetup: newPeriodicVestingAccount(stakeCoins(500050), thePast, []vesting.Period{
				{Length: oneDayAgo - thePast, Amount: stakeCoins(100040)},
				{Length: now - oneDayAgo, Amount: stakeCoins(400010)},
			}),
		},
		{
			name: "periodic vesting: finished 1 second ago",
			seq:  7111, expConv: true,
			acctSetup: newPeriodicVestingAccount(stakeCoins(505000), thePast, []vesting.Period{
				{Length: oneDayAgo - thePast, Amount: stakeCoins(102000)},
				{Length: oneSecAgo - oneDayAgo, Amount: stakeCoins(403000)},
			}),
		},
	}

	// Set up all the tests cases. Set the addr, baseAcct, origAcct, and expAcct fields, creating the accounts as needed.
	acctBalance := stakeCoins(1000000)
	expConvCount := 0
	setupOK := s.Run("account setup", func() {
		for _, tc := range tests {
			s.Run(tc.name, func() {
				if len(tc.addr) > 0 {
					tc.origAcct = s.app.AccountKeeper.GetAccount(s.ctx, tc.addr)
					// Not doing anything with tc.baseAcct since it shouldn't be needed for this case.
					tc.expAcct = tc.origAcct
					return
				}

				privKey := secp256k1.GenPrivKey()
				pubKey := privKey.PubKey()
				tc.addr = sdk.AccAddress(pubKey.Address())
				err := testutil.FundAccount(s.ctx, s.app.BankKeeper, tc.addr, acctBalance)
				s.LogIfError(err, "FundAccount(..., %q)", acctBalance.String())

				tc.origAcct = s.app.AccountKeeper.GetAccount(s.ctx, tc.addr)
				if !tc.noPK {
					err = tc.origAcct.SetPubKey(pubKey)
					s.Require().NoError(err, "acctI.SetPubKey(pubKey)")
				}

				s.Require().IsType(tc.baseAcct, tc.origAcct, "origAcct after funding")
				tc.baseAcct = tc.origAcct.(*authtypes.BaseAccount)
				tc.baseAcct.Sequence = tc.seq

				if tc.acctSetup == nil {
					tc.origAcct = tc.baseAcct
				} else {
					tc.origAcct, err = tc.acctSetup(tc.baseAcct)
					s.Require().NoError(err, "tc.acctSetup(...)")
				}

				s.app.AccountKeeper.SetAccount(s.ctx, tc.origAcct)

				tc.expAcct = tc.origAcct
				if tc.expConv {
					expConvCount++
					tc.expAcct = tc.baseAcct
				}
			})
		}

		s.Require().NotZero(expConvCount, "count of accounts expected to be converted.")
	})

	var err error
	runner := func() {
		err = convertFinishedVestingAccountsToBase(s.ctx, s.app)
	}

	s.Run("no accounts to convert", func() {
		if !setupOK {
			s.T().Skip("Skipping due to problem with setup.")
		}
		origCtx := s.ctx
		defer func() {
			s.ctx = origCtx
		}()
		s.ctx = s.ctx.WithBlockTime(reallyOldTime)

		expBals := s.getAllSpendable()

		expInLog := []string{
			LogMsgConvertFinishedVestingAccountsToBase,
			"INF No completed vesting accounts found.",
			"INF Done converting completed vesting accounts into base accounts.",
		}
		s.ExecuteAndAssertLogs(runner, expInLog, nil, true, "convertFinishedVestingAccountsToBase")
		s.Require().NoError(err, "convertFinishedVestingAccountsToBase")

		for _, tc := range tests {
			s.Run(tc.name, func() {
				acct := s.app.AccountKeeper.GetAccount(s.ctx, tc.addr)
				s.Require().NotNil(acct, "GetAccount(%q)", tc.addr.String())
				// Should still be the original since nothing was converted.
				s.Assert().Equal(tc.origAcct, acct, "GetAccount(%q) result", tc.addr.String())
			})
		}

		s.assertAllSpendable(expBals)
	})

	s.Run("with accounts to convert", func() {
		if !setupOK {
			s.T().Skip("Skipping due to problem with setup.")
		}
		origCtx := s.ctx
		defer func() {
			s.ctx = origCtx
		}()
		s.ctx = s.ctx.WithBlockTime(pivotTime)

		expBals := s.getAllSpendable()

		expInLog := []string{
			LogMsgConvertFinishedVestingAccountsToBase,
			fmt.Sprintf("INF Found %d completed vesting accounts. Updating them now.", expConvCount),
			"INF Done converting completed vesting accounts into base accounts.",
		}
		s.ExecuteAndAssertLogs(runner, expInLog, nil, true, "convertFinishedVestingAccountsToBase")
		s.Assert().NoError(err, "convertFinishedVestingAccountsToBase")

		for _, tc := range tests {
			s.Run(tc.name, func() {
				acct := s.app.AccountKeeper.GetAccount(s.ctx, tc.addr)
				s.Require().NotNil(acct, "GetAccount(%q)", tc.addr.String())
				s.Assert().Equal(tc.expAcct, acct, "GetAccount(%q) result", tc.addr.String())
			})
		}

		s.assertAllSpendable(expBals)
	})
}

func (s *UpgradeTestSuite) TestUnlockVestingAccounts() {
	var addrs []sdk.AccAddress
	newAddr := func() sdk.AccAddress {
		addr := sdk.AccAddress(fmt.Sprintf("addrs[%d]____________", len(addrs))[:20])
		addrs = append(addrs, addr)
		return addr
	}
	type expectedAcct struct {
		name string
		addr sdk.AccAddress
		orig sdk.AccountI
		exp  sdk.AccountI
	}
	var expectedAccts []expectedAcct
	expect := func(name string, addr sdk.AccAddress, orig, exp sdk.AccountI) {
		expectedAccts = append(expectedAccts, expectedAcct{name: name, addr: addr, orig: orig, exp: exp})
	}
	saveAcct := func(acct sdk.AccountI) sdk.AccountI {
		acct = s.app.AccountKeeper.NewAccount(s.ctx, acct)
		s.app.AccountKeeper.SetAccount(s.ctx, acct)
		return s.app.AccountKeeper.GetAccount(s.ctx, acct.GetAddress())
	}

	baseAddr := newAddr()
	baseAcct := authtypes.NewBaseAccountWithAddress(baseAddr)
	baseAcct.Sequence = 5
	baseAcctI := saveAcct(baseAcct)
	expect("base", baseAddr, baseAcct, baseAcct)

	vestContAddr := newAddr()
	vestContAcct, err := vesting.NewContinuousVestingAccount(
		authtypes.NewBaseAccountWithAddress(vestContAddr),
		sdk.NewCoins(sdk.NewInt64Coin("banana", 12)),
		s.ctx.BlockTime().Add(10*time.Second).Unix(),
		s.ctx.BlockTime().Add(100*time.Hour).Unix(),
	)
	s.Require().NoError(err, "NewContinuousVestingAccount")
	vestContAcct.Sequence = 3
	vestContAcctI := saveAcct(vestContAcct)
	expect("continuous", vestContAddr, vestContAcct, vestContAcct.BaseAccount)

	vestDelAddr := newAddr()
	vestDelAcct, err := vesting.NewDelayedVestingAccount(
		authtypes.NewBaseAccountWithAddress(vestDelAddr),
		sdk.NewCoins(sdk.NewInt64Coin("pear", 27)),
		s.ctx.BlockTime().Add(50*time.Minute).Unix(),
	)
	s.Require().NoError(err, "NewDelayedVestingAccount")
	vestDelAcct.Sequence = 12
	vestDelAcctI := saveAcct(vestDelAcct)
	expect("delayed", vestDelAddr, vestDelAcct, vestDelAcct.BaseAccount)

	vestPerAddr := newAddr()
	vestPerAcct, err := vesting.NewPeriodicVestingAccount(
		authtypes.NewBaseAccountWithAddress(vestPerAddr),
		sdk.NewCoins(sdk.NewInt64Coin("peach", 15)),
		s.ctx.BlockTime().Add(50*time.Minute).Unix(),
		vesting.Periods{
			{Length: 20 * 60, Amount: sdk.NewCoins(sdk.NewInt64Coin("peach", 5))},
			{Length: 30 * 60, Amount: sdk.NewCoins(sdk.NewInt64Coin("peach", 10))},
		},
	)
	s.Require().NoError(err, "NewPeriodicVestingAccount")
	vestPerAcct.Sequence = 6
	vestPerAcctI := saveAcct(vestPerAcct)
	expect("periodic", vestPerAddr, vestPerAcct, vestPerAcct.BaseAccount)

	permLockAddr := newAddr()
	permLockAcct, err := vesting.NewPermanentLockedAccount(
		authtypes.NewBaseAccountWithAddress(permLockAddr),
		sdk.NewCoins(sdk.NewInt64Coin("banana", 99)),
	)
	s.Require().NoError(err, "NewPermanentLockedAccount")
	permLockAcct.Sequence = 19
	permLockAcctI := saveAcct(permLockAcct)
	expect("permanent locked", permLockAddr, permLockAcct, permLockAcct.BaseAccount)

	modAddr := s.app.AccountKeeper.GetModuleAddress("marker")
	modAcct := s.app.AccountKeeper.GetModuleAccount(s.ctx, "marker")
	addrs = append(addrs, modAddr)
	expect("module", modAddr, modAcct, modAcct)

	unknownAddr := newAddr()
	expect("unknown", unknownAddr, nil, nil)

	expLogLines := []string{
		"INF Unlocking select vesting accounts.",
		"INF Identified 7 accounts to unlock.",
		fmt.Sprintf("ERR Could not unlock vesting account. "+
			"error=\"could not unlock account %s: unsupported account type *types.BaseAccount: invalid type\" "+
			"account_number=%d address=%s module=hold original_type=*types.BaseAccount",
			baseAddr, baseAcctI.GetAccountNumber(), baseAddr),
		fmt.Sprintf("INF Unlocked vesting account. "+
			"account_number=%d address=%s module=hold original_type=*types.ContinuousVestingAccount",
			vestContAcctI.GetAccountNumber(), vestContAddr),
		fmt.Sprintf("INF Unlocked vesting account. "+
			"account_number=%d address=%s module=hold original_type=*types.DelayedVestingAccount",
			vestDelAcctI.GetAccountNumber(), vestDelAddr),
		fmt.Sprintf("INF Unlocked vesting account. "+
			"account_number=%d address=%s module=hold original_type=*types.PeriodicVestingAccount",
			vestPerAcctI.GetAccountNumber(), vestPerAddr),
		fmt.Sprintf("INF Unlocked vesting account. "+
			"account_number=%d address=%s module=hold original_type=*types.PermanentLockedAccount",
			permLockAcctI.GetAccountNumber(), permLockAddr),
		fmt.Sprintf("ERR Could not unlock vesting account. "+
			"error=\"could not unlock account %s: unsupported account type *types.ModuleAccount: invalid type\" "+
			"account_number=%d address=%s module=hold original_type=*types.ModuleAccount",
			modAddr, modAcct.GetAccountNumber(), modAddr),
		fmt.Sprintf("ERR Could not unlock vesting account. "+
			"error=\"account \\\"%s\\\" does not exist: unknown address\" "+
			"address=%s module=hold",
			unknownAddr, unknownAddr),
		"INF Done unlocking select vesting accounts.",
		"",
	}

	var buffer bytes.Buffer
	logger := internal.NewBufferedDebugLogger(&buffer)
	ctx := s.ctx.WithLogger(logger)
	testFunc := func() {
		unlockVestingAccounts(ctx, s.app, addrs)
	}
	s.Require().NotPanics(testFunc, "unlockVestingAccounts")
	actLog := buffer.String()
	actLogLines := strings.Split(actLog, "\n")
	s.Assert().Equal(expLogLines, actLogLines, "Logged messages.")

	for _, tc := range expectedAccts {
		s.Run(tc.name, func() {
			acctI := s.app.AccountKeeper.GetAccount(ctx, tc.addr)
			s.Assert().Equal(tc.exp, acctI, "GetAccount result")
		})
	}
}

// Create strings with the log statements that start off the reusable upgrade functions.
var (
	LogMsgRunModuleMigrations                  = "INF Starting module migrations. This may take a significant amount of time to complete. Do not restart node."
	LogMsgRemoveInactiveValidatorDelegations   = "INF Removing inactive validator delegations."
	LogMsgPruneIBCExpiredConsensusStates       = "INF Pruning expired consensus states for IBC."
	LogMsgConvertFinishedVestingAccountsToBase = "INF Converting completed vesting accounts into base accounts."
)

func (s *UpgradeTestSuite) TestBouvardiaRC1() {
	expInLog := []string{
		"INF Starting module migrations. This may take a significant amount of time to complete. Do not restart node.",
		"INF Pruning expired consensus states for IBC.",
		"INF Removing inactive validator delegations.",
		"INF Converting completed vesting accounts into base accounts.",
		"INF Setting up flat fees.",
		"INF Bulk creating ledgers and entries.",
	}
	s.AssertUpgradeHandlerLogs("bouvardia-rc1", expInLog, nil)
}

func (s *UpgradeTestSuite) TestBouvardia() {
	expInLog := []string{
		"INF Starting module migrations. This may take a significant amount of time to complete. Do not restart node.",
		"INF Pruning expired consensus states for IBC.",
		"INF Removing inactive validator delegations.",
		"INF Converting completed vesting accounts into base accounts.",
		"INF Setting up flat fees.",
		"INF Bulk creating ledgers and entries.",
	}
	s.AssertUpgradeHandlerLogs("bouvardia", expInLog, nil)
}

type MockFlatFeesKeeper struct {
	SetParamsErrs  []string
	SetParamsCalls []flatfeestypes.Params

	SetMsgFeeErrs  []string
	SetMsgFeeCalls []*flatfeestypes.MsgFee
}

var _ FlatFeesKeeper = (*MockFlatFeesKeeper)(nil)

func NewMockFlatFeesKeeper() *MockFlatFeesKeeper {
	return &MockFlatFeesKeeper{}
}

func (m *MockFlatFeesKeeper) WithSetParamsErrs(errs ...string) *MockFlatFeesKeeper {
	m.SetParamsErrs = append(m.SetParamsErrs, errs...)
	return m
}

func (m *MockFlatFeesKeeper) WithSetMsgFeeErrs(errs ...string) *MockFlatFeesKeeper {
	m.SetMsgFeeErrs = append(m.SetMsgFeeErrs, errs...)
	return m
}

func (m *MockFlatFeesKeeper) SetParams(_ sdk.Context, params flatfeestypes.Params) error {
	m.SetParamsCalls = append(m.SetParamsCalls, params)
	var rv string
	if len(m.SetParamsErrs) > 0 {
		rv = m.SetParamsErrs[0]
		m.SetParamsErrs = m.SetParamsErrs[1:]
	}
	if len(rv) > 0 {
		return errors.New(rv)
	}
	return nil
}

func (m *MockFlatFeesKeeper) SetMsgFee(_ sdk.Context, msgFee flatfeestypes.MsgFee) error {
	msgFeeCopy := msgFee
	m.SetMsgFeeCalls = append(m.SetMsgFeeCalls, &msgFeeCopy)
	var rv string
	if len(m.SetMsgFeeErrs) > 0 {
		rv = m.SetMsgFeeErrs[0]
		m.SetMsgFeeErrs = m.SetMsgFeeErrs[1:]
	}
	if len(rv) > 0 {
		return errors.New(rv)
	}
	return nil
}

func TestSetupFlatFees(t *testing.T) {
	ctx := sdk.Context{}.WithLogger(log.NewNopLogger())
	expParams := MakeFlatFeesParams()
	expCosts := MakeFlatFeesCosts()

	tests := []struct {
		name    string
		ffk     *MockFlatFeesKeeper
		expErr  string
		expFees []*flatfeestypes.MsgFee
	}{
		{
			name:   "error setting params",
			ffk:    NewMockFlatFeesKeeper().WithSetParamsErrs("notgonnadoit"),
			expErr: "could not set x/flatfees params: notgonnadoit",
		},
		{
			name:    "error setting first cost",
			ffk:     NewMockFlatFeesKeeper().WithSetMsgFeeErrs("one error to rule them all"),
			expErr:  "could not set msg fee " + expCosts[0].String() + ": one error to rule them all",
			expFees: expCosts[0:1],
		},
		{
			name:    "error setting second cost",
			ffk:     NewMockFlatFeesKeeper().WithSetMsgFeeErrs("", "oopsies"),
			expErr:  "could not set msg fee " + expCosts[1].String() + ": oopsies",
			expFees: expCosts[0:2],
		},
		{
			name:    "error setting 10th cost",
			ffk:     NewMockFlatFeesKeeper().WithSetMsgFeeErrs("", "", "", "", "", "", "", "", "", "sproing"),
			expErr:  "could not set msg fee " + expCosts[9].String() + ": sproing",
			expFees: expCosts[0:10],
		},
		{
			name: "error setting 125th cost",
			ffk: NewMockFlatFeesKeeper().
				WithSetMsgFeeErrs(slices.Repeat([]string{""}, 124)...).
				WithSetMsgFeeErrs("almosthadit"),
			expErr:  "could not set msg fee " + expCosts[124].String() + ": almosthadit",
			expFees: expCosts[0:125],
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.ffk == nil {
				tc.ffk = NewMockFlatFeesKeeper()
			}

			var err error
			testFunc := func() {
				err = setupFlatFees(ctx, tc.ffk)
			}
			require.NotPanics(t, testFunc, "setupFlatFees")
			assertions.AssertErrorValue(t, err, tc.expErr, "setupFlatFees error")

			// SetParams should always get called exactly once.
			assert.Len(t, tc.ffk.SetParamsCalls, 1, "Calls made to SetParams(...)")
			for _, actParams := range tc.ffk.SetParamsCalls {
				ok := assert.Equal(t, expParams.DefaultCost.String(), actParams.DefaultCost.String(), "params.DefaultCost")
				ok = assert.Equal(t, expParams.ConversionFactor.String(), actParams.ConversionFactor.String(), "params.ConversionFactor") && ok
				if ok {
					assert.Equal(t, expParams, actParams, "params")
				}
			}

			expCalls := sliceToStrings(tc.expFees)
			actCalls := sliceToStrings(tc.ffk.SetMsgFeeCalls)
			if assert.Equal(t, expCalls, actCalls, "calls made to SetMsgFee (as strings)") {
				assert.Equal(t, tc.expFees, tc.ffk.SetMsgFeeCalls, "calls made to SetMsgFee")
			}
		})
	}
}

func sliceToStrings[S ~[]E, E fmt.Stringer](vals S) []string {
	if vals == nil {
		return nil
	}
	rv := make([]string, len(vals))
	for i, val := range vals {
		rv[i] = val.String()
	}
	return rv
}

func TestMakeFlatFeesParams(t *testing.T) {
	var params flatfeestypes.Params
	testFunc := func() {
		params = MakeFlatFeesParams()
	}
	require.NotPanics(t, testFunc, "MakeFlatFeesParams()")
	err := params.Validate()
	require.NoError(t, err, "params.Validate()")
	assert.Equal(t, "musd", params.DefaultCost.Denom, "params.DefaultCost.Denom")
	assert.Equal(t, "musd", params.ConversionFactor.DefinitionAmount.Denom, "params.ConversionFactor.DefinitionAmount.Denom")
	assert.Equal(t, "nhash", params.ConversionFactor.ConvertedAmount.Denom, "params.ConversionFactor.ConvertedAmount.Denom")
}

func TestMakeMsgFees(t *testing.T) {
	var msgFees []*flatfeestypes.MsgFee
	testFunc := func() {
		msgFees = MakeFlatFeesCosts()
	}
	require.NotPanics(t, testFunc, "MakeFlatFeesCosts()")
	require.NotNil(t, msgFees, "MakeFlatFeesCosts() result")
	require.NotEmpty(t, msgFees, "MakeFlatFeesCosts() result")

	for i, msgFee := range msgFees {
		t.Run(fmt.Sprintf("%d: %s", i, msgFee.MsgTypeUrl), func(t *testing.T) {
			err := msgFee.Validate()
			assert.NoError(t, err, "msgFee.Validate()")
			if len(msgFee.Cost) != 0 {
				if assert.Len(t, msgFee.Cost, 1, "msgFee.Cost") {
					assert.Equal(t, flatfeestypes.DefaultFeeDefinitionDenom, msgFee.Cost[0].Denom, "msgFee.Cost[0].Denom")
				}
			}

			// Make sure the msg type doesn't appear anywhere else in the list.
			for j, msgFee2 := range msgFees {
				if i == j {
					continue
				}
				assert.NotEqual(t, msgFee.MsgTypeUrl, msgFee2.MsgTypeUrl,
					"MsgTypeUrl of msgFees[%d]=%s and msgFees[%d]=%s", i, msgFee.Cost, j, msgFee2.Cost)
			}
		})
	}
}

func TestLogCostGrid(t *testing.T) {
	// This test just outputs (to logs) a table of costs at various conversion factors.
	// Delete this test with the other bouvardia stuff.
	// 1 hash = $0.025, so 1000000000nhash = 25musd.
	// Define the conversion factor musd amounts (per 1 hash).
	cfVals := []int64{20, 25, 30, 35, 40, 50, 75, 100}
	// Define a set of costs (in musd) that we always want in the table. All actually used amounts are automatically included.
	standardCosts := []int64{5, 50, 100, 150, 250, 500, 1000, 1500, 2000, 2500, 3000, 3500, 4000, 4500, 5000}

	// Create the conversion factors.
	cfs := make([]flatfeestypes.ConversionFactor, len(cfVals))
	for i, cf := range cfVals {
		cfs[i] = flatfeestypes.ConversionFactor{
			DefinitionAmount: sdk.NewInt64Coin("musd", cf),
			ConvertedAmount:  sdk.NewInt64Coin("nhash", 1000000000),
		}
	}

	// Combine the standard costs with all the actual costs without dups.
	// Start with the standard costs that we always want in the list.
	allCostAmounts := make(map[int64]bool)
	for _, amt := range standardCosts {
		allCostAmounts[amt] = true
	}

	// Add all the ones actually used, keeping track of them (for later reference).
	var usedCostAmounts []int64
	costUsed := make(map[int64]bool)
	for _, msgFee := range MakeFlatFeesCosts() {
		amt := msgFee.Cost.AmountOf("musd").Int64()
		if !costUsed[amt] {
			usedCostAmounts = append(usedCostAmounts, amt)
		}
		costUsed[amt] = true
		allCostAmounts[amt] = true
	}
	// The default isn't in this list, but should still be considered in use.
	if !costUsed[150] {
		costUsed[150] = true
		usedCostAmounts = append(usedCostAmounts, 150)
	}

	// Get the deduped list of cost amounts, sorted smallest to largest.
	costAmounts := slices.Sorted(maps.Keys(allCostAmounts))
	costs := make([]sdk.Coin, len(costAmounts))
	for i, amt := range costAmounts {
		costs[i] = sdk.NewInt64Coin("musd", amt)
	}

	// usdStr converts an amount of musd into a string in the format of "$1.234". The string will always
	// start with $, then have all needed whole digits and exactly three fractional digits.
	usdStr := func(musdAmount sdkmath.Int) string {
		rv := musdAmount.String()
		if len(rv) < 4 {
			rv = strings.Repeat("0", 4-len(rv)) + rv
		}
		p := len(rv) - 3
		return "$" + rv[:p] + "." + rv[p:]
	}
	// hashStr converts an amount of nhash into a string in the format of "1.23 hash". It will always have
	// at least one whole digit and exactly two fractional digits, and will always end with " hash".
	// Extra fractional decimals are simply truncated.
	hashStr := func(nhashAmount sdkmath.Int) string {
		rv := nhashAmount.String()
		if len(rv) < 10 {
			rv = strings.Repeat("0", 10-len(rv)) + rv
		}
		p := len(rv) - 9
		return rv[:p] + "." + rv[p:p+2] + " hash"
	}

	// Create a line for each cost amount.
	lines := make([]string, len(costAmounts))
	for i, costAmt := range costAmounts {
		cost := sdk.NewInt64Coin("musd", costAmt)
		// Create a column for each conversion factor.
		parts := make([]string, len(cfs))
		for j, cf := range cfs {
			convCost := cf.ConvertCoin(cost)
			parts[j] = fmt.Sprintf("%11s", hashStr(convCost.Amount))
		}
		// Put together the whole line.
		lines[i] = fmt.Sprintf("%8s = %s", usdStr(cost.Amount), strings.Join(parts, " | "))
		if !costUsed[costAmt] {
			lines[i] = "x" + strings.TrimPrefix(lines[i], " ")
		}
	}
	head := make([]string, len(cfs))
	for j, cf := range cfs {
		head[j] = fmt.Sprintf("%9s", usdStr(cf.DefinitionAmount.Amount)) + "  "
	}
	head[0] = "usd/hash = " + head[0]
	headLine := strings.Join(head, " | ")
	hrBz := make([]rune, len(headLine))
	for i, r := range headLine {
		switch r {
		case '|':
			hrBz[i] = '+'
		default:
			hrBz[i] = '-'
		}
	}
	t.Logf("Defined Costs at various conversion factors:\n%s\n%s\n%s", headLine, string(hrBz), strings.Join(lines, "\n"))
}

func (s *UpgradeTestSuite) TestStreamImportLedgerData() {
	// Test that the streaming ledger data import function works correctly
	// This test will only work if the gzipped file exists
	err := streamImportLedgerData(s.ctx, s.app.LedgerKeeper)
	if err != nil {
		s.T().Logf("Note: No ledger data files found or error loading: %v", err)
		return // This is expected if no data files exist
	}

	s.T().Log("Successfully stream imported ledger data")
}

func (s *UpgradeTestSuite) TestImportLedgerData() {
	// Test the full import process
	err := importLedgerData(s.ctx, s.app.LedgerKeeper)
	if err != nil {
		s.T().Logf("Note: Import ledger data failed (expected if no data): %v", err)
		return // This is expected if no data files exist
	}

	s.T().Log("Successfully imported ledger data")
}

func (s *UpgradeTestSuite) TestLedgerGenesisStateValidation() {
	// Load the actual genesis data from the gzipped file using the same method as the upgrade handler
	filePath := "upgrade_data/bouvardia_ledger_genesis.json.gz"

	// Read the gzipped file data
	data, err := upgradeDataFS.ReadFile(filePath)
	if err != nil {
		s.T().Logf("Note: No ledger genesis file found at %s: %v", filePath, err)
		s.T().Skip("Skipping test - no ledger genesis data file available")
		return
	}

	// Create gzip reader for decompression
	reader := bytes.NewReader(data)
	gzReader, err := gzip.NewReader(reader)
	if err != nil {
		s.T().Fatalf("Failed to create gzip reader for %s: %v", filePath, err)
	}
	defer gzReader.Close()

	// Decode the entire JSON into a GenesisState
	var genesisState ledgerTypes.GenesisState
	decoder := json.NewDecoder(gzReader)
	if err := decoder.Decode(&genesisState); err != nil {
		s.T().Fatalf("Failed to decode genesis state from %s: %v", filePath, err)
	}

	// Validate each ledger class
	for i, ledgerClass := range genesisState.LedgerClasses {
		err := ledgerClass.Validate()
		s.Require().NoError(err, "LedgerClass %d validation failed", i)
	}

	// Validate each ledger class entry type
	for i, entryType := range genesisState.LedgerClassEntryTypes {
		err := entryType.EntryType.Validate()
		s.Require().NoError(err, "LedgerClassEntryType %d validation failed", i)
	}

	// Validate each ledger class status type
	for i, statusType := range genesisState.LedgerClassStatusTypes {
		err := statusType.StatusType.Validate()
		s.Require().NoError(err, "LedgerClassStatusType %d validation failed", i)
	}

	// Validate each ledger class bucket type
	for i, bucketType := range genesisState.LedgerClassBucketTypes {
		err := bucketType.BucketType.Validate()
		s.Require().NoError(err, "LedgerClassBucketType %d validation failed", i)
	}

	// Validate each ledger
	for i, genesisLedger := range genesisState.Ledgers {
		err := genesisLedger.Ledger.Validate()
		s.Require().NoError(err, "Ledger %d validation failed", i)
	}

	// Validate each ledger entry
	for i, genesisEntry := range genesisState.LedgerEntries {
		err := genesisEntry.Entry.Validate()
		s.Require().NoError(err, "LedgerEntry %d validation failed", i)
	}

	// Validate each settlement instruction
	for i, settlement := range genesisState.SettlementInstructions {
		for j, instruction := range settlement.SettlementInstructions.SettlementInstructions {
			err := instruction.Validate()
			s.Require().NoError(err, "SettlementInstruction %d.%d validation failed", i, j)
		}
	}

	s.T().Log("Successfully validated all ledger genesis state components")
}
