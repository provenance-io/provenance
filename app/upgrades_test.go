package app

import (
	"bytes"
	"context"
	"embed"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	internalsdk "github.com/provenance-io/provenance/internal/sdk"
	metadatakeeper "github.com/provenance-io/provenance/x/metadata/keeper"
	metadatatypes "github.com/provenance-io/provenance/x/metadata/types"
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
	_, err = s.app.StakingKeeper.Delegate(s.ctx, delegatorAddress, coin.Amount, stakingtypes.Unbonded, validator, true)
	s.Require().NoError(err, "Delegate(%q, %q, unbonded, %q, true) error", delegatorAddress.String(), coin.String(), valAddress.String())
}

func (s *UpgradeTestSuite) GetOperatorAddr(val stakingtypes.ValidatorI) sdk.ValAddress {
	addr, err := internalsdk.GetOperatorAddr(val)
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
		"INF Migrating consensus params.",
		"INF Done migrating consensus params.",
		"INF Migrating bank params.",
		"INF Done migrating bank params.",
		"INF Migrating attribute params.",
		"INF Done migrating attribute params.",
		"INF Migrating marker params.",
		"INF Done migrating marker params.",
		"INF Migrating metadata os locator params.",
		"INF Done migrating metadata os locator params.",
		"INF Migrating msgfees params.",
		"INF Done migrating msgfees params.",
		"INF Migrating name params.",
		"INF Done migrating name params.",
		"INF Migrating ibchooks params.",
		"INF Done migrating ibchooks params.",
		"INF Starting module migrations. This may take a significant amount of time to complete. Do not restart node.",
		"INF Setting new gov params for testnet.",
		"INF Done setting new gov params for testnet.",
		"INF Updating IBC AllowedClients.",
		"INF Done updating IBC AllowedClients.",
		"INF Adding scope net asset values.",
		"INF Adding 1101 scope net asset value entries.",
		"INF Successfully added 0 of 1101 scope net asset value entries.",
		"INF Done adding scope net asset values.",
		"INF Removing inactive validator delegations.",
		"INF Threshold: 21 days",
		"INF A total of 0 inactive (unbonded) validators have had all their delegators removed.",
	}

	s.AssertUpgradeHandlerLogs("umber-rc1", expInLog, nil)
}

func (s *UpgradeTestSuite) TestUmberRC4() {
	key := "umber-rc4"
	s.Assert().Contains(upgrades, key, "%q defined upgrades map", key)

	entry := upgrades[key]
	s.Assert().NotNil(entry, "%q entry in the upgrades map", key)
	s.Assert().Empty(entry.Added, "%q.Added", key)
	s.Assert().Empty(entry.Deleted, "%q.Deleted", key)
	s.Assert().Empty(entry.Renamed, "%q.Renamed", key)
	s.Assert().Nil(entry.Handler, "%q.Handler", key)
}

func (s *UpgradeTestSuite) TestUmber() {
	expInLog := []string{
		"INF Pruning expired consensus states for IBC.",
		"INF Done pruning expired consensus states for IBC.",
		"INF Migrating consensus params.",
		"INF Done migrating consensus params.",
		"INF Migrating bank params.",
		"INF Done migrating bank params.",
		"INF Migrating attribute params.",
		"INF Done migrating attribute params.",
		"INF Migrating marker params.",
		"INF Done migrating marker params.",
		"INF Migrating metadata os locator params.",
		"INF Done migrating metadata os locator params.",
		"INF Migrating msgfees params.",
		"INF Done migrating msgfees params.",
		"INF Migrating name params.",
		"INF Done migrating name params.",
		"INF Migrating ibchooks params.",
		"INF Done migrating ibchooks params.",
		"INF Starting module migrations. This may take a significant amount of time to complete. Do not restart node.",
		"INF Setting new gov params for mainnet.",
		"INF Done setting new gov params for mainnet.",
		"INF Updating IBC AllowedClients.",
		"INF Done updating IBC AllowedClients.",
		"INF Adding scope net asset values.",
		"INF Adding 215558 scope net asset value entries.",
		"INF Successfully added 0 of 215558 scope net asset value entries.",
		"INF Done adding scope net asset values.",
		"INF Removing inactive validator delegations.",
		"INF Threshold: 21 days",
		"INF A total of 0 inactive (unbonded) validators have had all their delegators removed.",
		"INF Storing the Funding Trading Bridge smart contract.",
		"INF Done storing the Funding Trading Bridge smart contract.",
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
			"INF Removing inactive validator delegations.",
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
		unbondedVal1 := s.CreateValidator(s.startTime.Add(-30*24*time.Hour), stakingtypes.Unbonded)
		unbondedVal1Addr := s.GetOperatorAddr(unbondedVal1)
		addr1Balance := s.app.BankKeeper.GetBalance(s.ctx, addr1, "stake")
		s.DelegateToValidator(unbondedVal1Addr, addr1, delegationCoin)
		s.Require().Equal(addr1Balance.Sub(delegationCoin), s.app.BankKeeper.GetBalance(s.ctx, addr1, "stake"), "addr1 should have less funds")
		validators, err := s.app.StakingKeeper.GetAllValidators(s.ctx)
		s.Require().NoError(err, "GetAllValidators")
		s.Require().Len(validators, 2, "Setup: GetAllValidators should have: 1 bonded, 1 unbonded")

		expectedLogLines := []string{
			"INF Removing inactive validator delegations.",
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
			"INF Removing inactive validator delegations.",
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
			"INF Removing inactive validator delegations.",
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
			"INF Removing inactive validator delegations.",
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
		unbondedVal1 := s.CreateValidator(s.startTime.Add(-30*24*time.Hour), stakingtypes.Unbonded)
		validators, err := s.app.StakingKeeper.GetAllValidators(s.ctx)
		s.Require().NoError(err, "GetAllValidators")
		s.Require().Len(validators, 3, "Setup: GetAllValidators should have: 1 bonded, 1 recently unbonded, 1 empty unbonded")

		expectedLogLines := []string{
			"INF Removing inactive validator delegations.",
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

func (s *UpgradeTestSuite) TestSetNewGovParamsTestnet() {
	var runErr error
	runner := func() {
		runErr = setNewGovParamsTestnet(s.ctx, s.app)
	}
	runnerName := "setNewGovParamsTestnet"

	threeDays := time.Hour * 24 * 3
	fiveMinutes := time.Minute * 5

	iniParams := govv1.DefaultParams()
	iniParams.MinInitialDepositRatio = sdkmath.LegacyMustNewDecFromStr("0.25").String()
	iniParams.MinDepositRatio = sdkmath.LegacyMustNewDecFromStr("0.10").String()
	iniParams.ProposalCancelRatio = sdkmath.LegacyMustNewDecFromStr("0.83").String()
	iniParams.ProposalCancelDest = sdk.AccAddress("addr________________").String()
	iniParams.ExpeditedVotingPeriod = &threeDays
	iniParams.ExpeditedThreshold = sdkmath.LegacyMustNewDecFromStr("0.406").String()
	iniParams.ExpeditedMinDeposit = []sdk.Coin{sdk.NewInt64Coin("banana", 1000)}
	iniParams.BurnVoteQuorum = true
	iniParams.BurnProposalDepositPrevote = true
	iniParams.BurnVoteVeto = false

	err := s.app.GovKeeper.Params.Set(s.ctx, iniParams)
	s.Require().NoError(err, "Setting initial gov params")

	expParams := govv1.DefaultParams()
	expParams.MinInitialDepositRatio = sdkmath.LegacyMustNewDecFromStr("0.00002").String()
	expParams.MinDepositRatio = sdkmath.LegacyZeroDec().String()
	expParams.ProposalCancelRatio = sdkmath.LegacyZeroDec().String()
	expParams.ProposalCancelDest = ""
	expParams.ExpeditedVotingPeriod = &fiveMinutes
	expParams.ExpeditedThreshold = sdkmath.LegacyMustNewDecFromStr("0.667").String()
	expParams.ExpeditedMinDeposit = iniParams.MinDeposit
	expParams.BurnVoteQuorum = false
	expParams.BurnProposalDepositPrevote = false
	expParams.BurnVoteVeto = true

	expLogLines := []string{
		"INF Setting new gov params for testnet.",
		"INF Done setting new gov params for testnet.",
	}

	s.ExecuteAndAssertLogs(runner, expLogLines, nil, true, runnerName)
	s.Require().NoError(runErr, runnerName)

	actParams, err := s.app.GovKeeper.Params.Get(s.ctx)
	s.Require().NoError(err, "getting gov params after %s", runnerName)
	s.Assert().Equal(expParams, actParams, "resulting gov params")
}

func (s *UpgradeTestSuite) TestSetNewGovParamsMainnet() {
	var runErr error
	runner := func() {
		runErr = setNewGovParamsMainnet(s.ctx, s.app)
	}
	runnerName := "setNewGovParamsMainnet"

	threeDays := time.Hour * 24 * 3
	oneDay := time.Hour * 24

	iniParams := govv1.DefaultParams()
	iniParams.MinInitialDepositRatio = sdkmath.LegacyMustNewDecFromStr("0.25").String()
	iniParams.MinDepositRatio = sdkmath.LegacyMustNewDecFromStr("0.10").String()
	iniParams.ProposalCancelRatio = sdkmath.LegacyMustNewDecFromStr("0.83").String()
	iniParams.ProposalCancelDest = sdk.AccAddress("addr________________").String()
	iniParams.ExpeditedVotingPeriod = &threeDays
	iniParams.ExpeditedThreshold = sdkmath.LegacyMustNewDecFromStr("0.406").String()
	iniParams.ExpeditedMinDeposit = []sdk.Coin{sdk.NewInt64Coin("banana", 1000)}
	iniParams.BurnVoteQuorum = true
	iniParams.BurnProposalDepositPrevote = false
	iniParams.BurnVoteVeto = false

	err := s.app.GovKeeper.Params.Set(s.ctx, iniParams)
	s.Require().NoError(err, "Setting initial gov params")

	expParams := govv1.DefaultParams()
	expParams.MinInitialDepositRatio = sdkmath.LegacyMustNewDecFromStr("0.02").String()
	expParams.MinDepositRatio = sdkmath.LegacyZeroDec().String()
	expParams.ProposalCancelRatio = sdkmath.LegacyMustNewDecFromStr("0.5").String()
	expParams.ProposalCancelDest = ""
	expParams.ExpeditedVotingPeriod = &oneDay
	expParams.ExpeditedThreshold = sdkmath.LegacyMustNewDecFromStr("0.667").String()
	expParams.ExpeditedMinDeposit = iniParams.MinDeposit
	expParams.BurnVoteQuorum = false
	expParams.BurnProposalDepositPrevote = true
	expParams.BurnVoteVeto = true

	expLogLines := []string{
		"INF Setting new gov params for mainnet.",
		"INF Done setting new gov params for mainnet.",
	}

	s.ExecuteAndAssertLogs(runner, expLogLines, nil, true, runnerName)
	s.Require().NoError(runErr, runnerName)

	actParams, err := s.app.GovKeeper.Params.Get(s.ctx)
	s.Require().NoError(err, "getting gov params after %s", runnerName)
	s.Assert().Equal(expParams, actParams, "resulting gov params")
}

func (s *UpgradeTestSuite) TestAddScopeNAVsWithHeight() {
	usdNAV := func(amount int) *metadatatypes.NetAssetValue {
		return &metadatatypes.NetAssetValue{Price: sdk.NewInt64Coin(metadatatypes.UsdDenom, int64(amount))}
	}

	// Define some scopes and initial NAVs, and store them.
	address1 := sdk.AccAddress("address1")
	testScopes := []struct {
		scopeUUID string
		initNAV   *metadatatypes.NetAssetValue
	}{
		{scopeUUID: "11111111-1111-1111-1111-111111111111", initNAV: usdNAV(55)},
		{scopeUUID: "22222222-2222-2222-2222-222222222222"},
		{scopeUUID: "33333333-3333-3333-3333-333333333333"},
		{scopeUUID: "44444444-4444-4444-4444-444444444444"},
		{scopeUUID: "55555555-5555-5555-5555-555555555555", initNAV: usdNAV(73)},
	}

	for i, ts := range testScopes {
		uid, err := uuid.Parse(ts.scopeUUID)
		s.Require().NoError(err)
		scopeAddr := metadatatypes.ScopeMetadataAddress(uid)
		scope := metadatatypes.Scope{
			ScopeId: scopeAddr,
			Owners:  []metadatatypes.Party{{Address: address1.String(), Role: metadatatypes.PartyType_PARTY_TYPE_OWNER}},
		}
		s.app.MetadataKeeper.SetScope(s.ctx, scope)

		if ts.initNAV != nil {
			err = s.app.MetadataKeeper.SetNetAssetValueWithBlockHeight(s.ctx, scopeAddr, *ts.initNAV, "test", 100)
			s.Require().NoError(err, "[%d/%d]: SetNetAssetValueWithBlockHeight", i+1, len(testScopes))
		}
	}

	// Define the scope NAVs that we'll be giving to addScopeNAVsWithHeight.
	scopeNAVs := []ScopeNAV{
		{ScopeUUID: "22222222-2222-2222-2222-222222222222", NetAssetValue: *usdNAV(12345), Height: 406},
		{ScopeUUID: "12345678-90ab-cdef-123x-567890abcdef", NetAssetValue: *usdNAV(54321), Height: 407},
		{ScopeUUID: "12345678-90ab-cdef-1234-567890abcdef", NetAssetValue: *usdNAV(66666), Height: 408},
		{
			ScopeUUID: "44444444-4444-4444-4444-444444444444",
			NetAssetValue: metadatatypes.NetAssetValue{
				Price: sdk.Coin{Denom: metadatatypes.UsdDenom, Amount: sdkmath.NewInt(-9001)},
			},
			Height: 409,
		},
		{ScopeUUID: "55555555-5555-5555-5555-555555555555", NetAssetValue: *usdNAV(987654), Height: 410},
	}

	// Define the expected log output from addScopeNAVsWithHeight when given those NAVs.
	expLogOutput := []string{
		"INF Adding 5 scope net asset value entries.",
		"ERR [2/5]: Invalid UUID \"12345678-90ab-cdef-123x-567890abcdef\": invalid UUID format.",
		"ERR [3/5]: Unable to find scope with UUID \"12345678-90ab-cdef-1234-567890abcdef\".",
		"ERR [4/5]: Unable to set scope scope1qpzyg3zyg3zyg3zyg3zyg3zyg3zqyvqcrf (\"44444444-4444-4444-4444-444444444444\") net asset value -9001usd at height 409: negative coin amount: -9001.",
		"INF Successfully added 2 of 5 scope net asset value entries.",
	}

	// Define what the expected resulting NAV entries should be for each scope.
	tests := []struct {
		name      string
		scopeUUID string
		expNav    *metadatatypes.NetAssetValue
		expHeight uint64
	}{
		{
			name:      "previously set nav that was not provided",
			scopeUUID: "11111111-1111-1111-1111-111111111111",
			expNav:    usdNAV(55),
			expHeight: 100,
		},
		{
			name:      "new nav set",
			scopeUUID: "22222222-2222-2222-2222-222222222222",
			expNav:    usdNAV(12345),
			expHeight: 406,
		},
		{
			name:      "no nav set, was not provided",
			scopeUUID: "33333333-3333-3333-3333-333333333333",
			expNav:    nil,
		},
		{
			name:      "nav add fails for being invalid",
			scopeUUID: "44444444-4444-4444-4444-444444444444",
			expNav:    nil,
		},
		{
			name:      "nav set over existing entry",
			scopeUUID: "55555555-5555-5555-5555-555555555555",
			expNav:    usdNAV(987654),
			expHeight: 410,
		},
		{
			name:      "nav add fails for non-existent scope",
			scopeUUID: "12345678-90ab-cdef-1234-567890abcdef",
			expNav:    nil,
		},
	}

	// Call addScopeNAVsWithHeight.
	s.logBuffer.Reset()
	addScopeNAVsWithHeight(s.ctx, s.app, scopeNAVs)

	// Check the logs.
	s.Run("addScopeNAVsWithHeight logs", func() {
		actLogOutput := splitLogOutput(s.GetLogOutput("addScopeNAVsWithHeight"))
		s.Assert().Equal(expLogOutput, actLogOutput, "addScopeNAVsWithHeight")
	})

	// Check the NAVs.
	for _, tc := range tests {
		s.Run(tc.name, func() {
			uid, err := uuid.Parse(tc.scopeUUID)
			s.Require().NoError(err)
			scopeAddr := metadatatypes.ScopeMetadataAddress(uid)

			netAssetValues := []metadatatypes.NetAssetValue{}
			err = s.app.MetadataKeeper.IterateNetAssetValues(s.ctx, scopeAddr, func(state metadatatypes.NetAssetValue) (stop bool) {
				netAssetValues = append(netAssetValues, state)
				return false
			})
			s.Require().NoError(err, "IterateNetAssetValues err")

			if tc.expNav != nil {
				s.Assert().Len(netAssetValues, 1, "Should be 1 nav set for scope")
				s.Assert().Equal(tc.expNav.Price, netAssetValues[0].Price, "Net asset value price should equal default upgraded price")
				s.Assert().Equal(tc.expHeight, netAssetValues[0].UpdatedBlockHeight, "Net asset value updated block height should equal expected")
			} else {
				s.Assert().Len(netAssetValues, 0, "Scope not expected to have nav")
			}
		})
	}
}

// wrappedWasmMsgSrvr is a wasmtypes.MsgServer that lets us inject an error to return from StoreCode.
type wrappedWasmMsgSrvr struct {
	wasmMsgServer wasmtypes.MsgServer
	storeCodeErr  string
}

func newWrappedWasmMsgSrvr(wasmKeeper *wasmkeeper.Keeper) *wrappedWasmMsgSrvr {
	return &wrappedWasmMsgSrvr{wasmMsgServer: wasmkeeper.NewMsgServerImpl(wasmKeeper)}
}

func (w *wrappedWasmMsgSrvr) WithStoreCodeErr(str string) *wrappedWasmMsgSrvr {
	w.storeCodeErr = str
	return w
}

func (w *wrappedWasmMsgSrvr) StoreCode(ctx context.Context, msg *wasmtypes.MsgStoreCode) (*wasmtypes.MsgStoreCodeResponse, error) {
	if len(w.storeCodeErr) > 0 {
		return nil, errors.New(w.storeCodeErr)
	}
	return w.wasmMsgServer.StoreCode(ctx, msg)
}

func (s *UpgradeTestSuite) TestStoreWasmCode() {
	origUpgradeFiles := UpgradeFiles
	defer func() {
		UpgradeFiles = origUpgradeFiles
	}()

	tests := []struct {
		name         string
		upgradeFiles embed.FS
		expLogs      []string
	}{
		{
			name:         "success",
			upgradeFiles: UpgradeFiles,
			expLogs: []string{
				"INF Storing the Funding Trading Bridge smart contract.",
				"INF Smart contract stored with codeID: 1 and checksum: \"7e643e228169980aff5d75d576873d34b368d30a154dc617d2ed9b0093c97128\".",
				"INF Done storing the Funding Trading Bridge smart contract.",
			},
		},
		{
			name:         "failed to read file",
			upgradeFiles: embed.FS{},
			expLogs: []string{
				"INF Storing the Funding Trading Bridge smart contract.",
				"ERR Could not read smart contract. error=\"open upgrade_files/umber/funding_trading_bridge_smart_contract.wasm: file does not exist\"",
				"INF Done storing the Funding Trading Bridge smart contract.",
			},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			UpgradeFiles = tc.upgradeFiles
			s.logBuffer.Reset()
			testFunc := func() {
				storeWasmCode(s.ctx, s.app)
			}
			s.Require().NotPanics(testFunc, "storeWasmCode")
			actLogs := splitLogOutput(s.GetLogOutput("executeStoreCodeMsg"))
			s.Assert().Equal(tc.expLogs, actLogs, "log messages during executeStoreCodeMsg")
		})
	}
}

func (s *UpgradeTestSuite) TestExecuteStoreCodeMsg() {
	codeBz, err := UpgradeFiles.ReadFile("upgrade_files/umber/funding_trading_bridge_smart_contract.wasm")
	s.Require().NoError(err, "reading wasm file")
	msg := &wasmtypes.MsgStoreCode{
		Sender:                s.app.GovKeeper.GetAuthority(),
		WASMByteCode:          codeBz,
		InstantiatePermission: &wasmtypes.AccessConfig{Permission: wasmtypes.AccessTypeEverybody},
	}

	tests := []struct {
		name    string
		errMsg  string
		expLogs []string
	}{
		{
			name:    "storage fails",
			errMsg:  "dang no worky",
			expLogs: []string{"ERR Could not store smart contract. error=\"dang no worky\""},
		},
		{
			name:    "storage succeeds",
			expLogs: []string{"INF Smart contract stored with codeID: 1 and checksum: \"7e643e228169980aff5d75d576873d34b368d30a154dc617d2ed9b0093c97128\"."},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			msgServer := newWrappedWasmMsgSrvr(s.app.WasmKeeper).WithStoreCodeErr(tc.errMsg)
			s.logBuffer.Reset()
			testFunc := func() {
				executeStoreCodeMsg(s.ctx, msgServer, msg)
			}
			s.Require().NotPanics(testFunc, "executeStoreCodeMsg")
			actLogs := splitLogOutput(s.GetLogOutput("executeStoreCodeMsg"))
			s.Assert().Equal(tc.expLogs, actLogs, "log messages during executeStoreCodeMsg")
		})
	}
}

func (s *UpgradeTestSuite) TestViridianRC1() {
	expInLog := []string{
		"INF Pruning expired consensus states for IBC.",
		"INF Starting module migrations. This may take a significant amount of time to complete. Do not restart node.",
		"INF Removing inactive validator delegations.",
	}
	s.AssertUpgradeHandlerLogs("viridian-rc1", expInLog, nil)
}

func (s *UpgradeTestSuite) TestViridian() {
	expInLog := []string{
		"INF Pruning expired consensus states for IBC.",
		"INF Starting module migrations. This may take a significant amount of time to complete. Do not restart node.",
		"INF Removing inactive validator delegations.",
	}
	s.AssertUpgradeHandlerLogs("viridian", expInLog, nil)
}

func (s *UpgradeTestSuite) TestMetadataMigration() {
	// TODO: Delete this test with the rest of the viridian stuff.
	newAddr := func(name string) sdk.AccAddress {
		switch {
		case len(name) < 20:
			// If it's less than 19 bytes long, pad it to 20 chars.
			return sdk.AccAddress(name + strings.Repeat("_", 20-len(name)))
		case len(name) > 20 && len(name) < 32:
			// If it's 21 to 31 bytes long, pad it to 32 chars.
			return sdk.AccAddress(name + strings.Repeat("_", 32-len(name)))
		}
		// If the name is exactly 20 long already, or longer than 32, don't include any padding.
		return sdk.AccAddress(name)
	}
	addrs := make([]string, 1000)
	for i := range addrs {
		addrs[i] = newAddr(fmt.Sprintf("%03d", i)).String()
	}

	newUUID := func(i int) uuid.UUID {
		// Sixteen 9's is the largest number we can handle; one more and it's 17 digits.
		s.Require().LessOrEqual(i, 9999999999999999, "value provided to newScopeID")
		str := fmt.Sprintf("________________%d", i)
		str = str[len(str)-16:]
		rv, err := uuid.FromBytes([]byte(str))
		s.Require().NoError(err, "uuid.FromBytes([]byte(%q))", str)
		return rv
	}
	newScopeID := func(i int) metadatatypes.MetadataAddress {
		return metadatatypes.ScopeMetadataAddress(newUUID(i))
	}
	newSpecID := func(i int) metadatatypes.MetadataAddress {
		// The spec id shouldn't really matter in here, but I want it different from a scope's index.
		// So I do some math to make it seem kind of random, but is still deterministic.
		// 7, 39, and 79 were picked randomly and have no special meaning.
		// 4,999 was chosen so that there's 3,333 possible results.
		// This will overflow at 2,097,112, but it shouldn't get bigger than 300,000 in here, so we ignore that.
		j := ((i + 7) * (i + 39) * (i + 79)) % 49_999
		return metadatatypes.ScopeSpecMetadataAddress(newUUID(j))
	}
	newScope := func(i int) metadatatypes.Scope {
		rv := metadatatypes.Scope{
			ScopeId:            newScopeID(i),
			SpecificationId:    newSpecID(i),
			RequirePartyRollup: i%6 > 2, // 50% chance, three false, then three true, repeated.
		}

		// 1 in 7 does not have a value owner.
		incVO := i%7 != 0
		if incVO {
			rv.ValueOwnerAddress = addrs[i%len(addrs)]
		}

		// Include 1 to 5 owners and make one of them the value owner in just under 1 in 11 scopes.
		ownerCount := (i % 5) + 1
		incVOInOwners := incVO && i%11 == 0
		incVOAt := i % ownerCount
		rv.Owners = make([]metadatatypes.Party, ownerCount)
		for o := range rv.Owners {
			a := i + (o+1)*300
			if incVOInOwners && o == incVOAt {
				a = i
			}
			rv.Owners[o].Address = addrs[a%len(addrs)]
			rv.Owners[o].Role = metadatatypes.PartyType(1 + (i+o)%11) // 11 different roles, 1 to 11.
		}

		daCount := (i % 3) // 0 to 2.
		for d := 0; d < daCount; d++ {
			a := i + (d+1)*33
			rv.DataAccess = append(rv.DataAccess, addrs[a%len(addrs)])
		}

		return rv
	}
	voInOwners := func(scope metadatatypes.Scope) bool {
		if len(scope.ValueOwnerAddress) == 0 {
			return false
		}
		for _, owner := range scope.Owners {
			if owner.Address == scope.ValueOwnerAddress {
				return true
			}
		}
		return false
	}

	// Create 300,005 scopes and write them to state.
	// 6 of 7 have a value owner = 257_147 = 300_005 * 6 / 7 (truncated).
	// 1 of 7 do not = 42_858 = 300_005 - 257_147 .
	// There's 23,377 scopes that have the value owner in owners.
	// So we expect 490,917 keys to delete = 257,147 * 2 - 23,377.
	expCoin := make([]metadatatypes.MetadataAddress, 0, 257_147)
	expNoCoin := make([]metadatatypes.MetadataAddress, 0, 42_858)
	expDelInds := make([][]byte, 0, 490_917)
	expBals := make(map[string]sdk.Coins)
	t1 := time.Now()
	for i := 0; i < 300_005; i++ {
		scope := newScope(i)
		s.Require().NoError(s.app.MetadataKeeper.V3WriteNewScope(s.ctx, scope), "[%d]: V3WriteNewScope", i)

		if len(scope.ValueOwnerAddress) == 0 {
			expNoCoin = append(expNoCoin, scope.ScopeId)
			continue
		}

		expCoin = append(expCoin, scope.ScopeId)

		vo := scope.ValueOwnerAddress
		voAddr := sdk.MustAccAddressFromBech32(vo)
		expDelInds = append(expDelInds, metadatakeeper.GetValueOwnerScopeCacheKey(voAddr, scope.ScopeId))
		if !voInOwners(scope) {
			expDelInds = append(expDelInds, metadatatypes.GetAddressScopeCacheKey(voAddr, scope.ScopeId))
		}

		expBals[vo] = expBals[vo].Add(scope.ScopeId.Coin())
	}
	t2 := time.Now()
	s.T().Logf("setup took %s", t2.Sub(t1))
	s.T().Logf("len(expDelInds) = %d", len(expDelInds))
	s.T().Logf("len(expCoin) = %d", len(expCoin))
	s.T().Logf("len(expNoCoin) = %d", len(expNoCoin))

	mdStore := s.ctx.KVStore(s.app.GetKey(metadatatypes.ModuleName))
	for _, ind := range expDelInds {
		has := mdStore.Has(ind)
		s.Assert().True(has, "mdStore.Has(%v) before running the migrations", ind)
	}
	mdStore = nil

	expLogs := []string{
		"INF Starting module migrations. This may take a significant amount of time to complete. Do not restart node.",
		"INF Starting migration of x/metadata from 3 to 4. module=x/metadata",
		"INF Moving scope value owner data into x/bank ledger. module=x/metadata",
		"INF Progress update: module=x/metadata scopes=10000 value owners=8571",
		"INF Progress update: module=x/metadata scopes=20000 value owners=17143",
		"INF Progress update: module=x/metadata scopes=30000 value owners=25714",
		"INF Progress update: module=x/metadata scopes=40000 value owners=34286",
		"INF Progress update: module=x/metadata scopes=50000 value owners=42857",
		"INF Progress update: module=x/metadata scopes=60000 value owners=51428",
		"INF Progress update: module=x/metadata scopes=70000 value owners=60000",
		"INF Progress update: module=x/metadata scopes=80000 value owners=68571",
		"INF Progress update: module=x/metadata scopes=90000 value owners=77143",
		"INF Progress update: module=x/metadata scopes=100000 value owners=85714",
		"INF Progress update: module=x/metadata scopes=110000 value owners=94286",
		"INF Progress update: module=x/metadata scopes=120000 value owners=102857",
		"INF Progress update: module=x/metadata scopes=130000 value owners=111428",
		"INF Progress update: module=x/metadata scopes=140000 value owners=120000",
		"INF Progress update: module=x/metadata scopes=150000 value owners=128571",
		"INF Progress update: module=x/metadata scopes=160000 value owners=137143",
		"INF Progress update: module=x/metadata scopes=170000 value owners=145714",
		"INF Progress update: module=x/metadata scopes=180000 value owners=154286",
		"INF Progress update: module=x/metadata scopes=190000 value owners=162857",
		"INF Progress update: module=x/metadata scopes=200000 value owners=171428",
		"INF Progress update: module=x/metadata scopes=210000 value owners=180000",
		"INF Progress update: module=x/metadata scopes=220000 value owners=188572",
		"INF Progress update: module=x/metadata scopes=230000 value owners=197143",
		"INF Progress update: module=x/metadata scopes=240000 value owners=205714",
		"INF Progress update: module=x/metadata scopes=250000 value owners=214286",
		"INF Progress update: module=x/metadata scopes=260000 value owners=222857",
		"INF Progress update: module=x/metadata scopes=270000 value owners=231429",
		"INF Progress update: module=x/metadata scopes=280000 value owners=240000",
		"INF Progress update: module=x/metadata scopes=290000 value owners=248572",
		"INF Progress update: module=x/metadata scopes=300000 value owners=257143",
		"INF Done moving scope value owners into bank module. module=x/metadata scopes=300005 value owners=257147",
		"INF Done migrating x/metadata from 3 to 4. module=x/metadata",
		"INF Module migrations completed.",
	}

	vm, err := s.app.UpgradeKeeper.GetModuleVersionMap(s.ctx)
	s.Require().NoError(err, "GetModuleVersionMap")
	s.Require().Equal(4, int(vm[metadatatypes.ModuleName]), "%s module version", metadatatypes.ModuleName)
	// Drop it back to 3 so the migration runs.
	vm[metadatatypes.ModuleName] = 3

	runner := func() {
		t1 = time.Now()
		vm, err = runModuleMigrations(s.ctx, s.app, vm)
		t2 = time.Now()
	}
	s.ExecuteAndAssertLogs(runner, expLogs, nil, true, "runModuleMigrations")
	s.Assert().NoError(err, "error from runModuleMigrations")
	s.Assert().Equal(4, int(vm[metadatatypes.ModuleName]), "vm[metadatatypes.ModuleName]")
	s.T().Logf("runModuleMigrations took %s", t2.Sub(t1))

	for _, scopeID := range expCoin {
		denom := scopeID.Denom()
		supply := s.app.BankKeeper.GetSupply(s.ctx, denom)
		s.Assert().Equal("1"+denom, supply.String(), "GetSupply(%q)", denom)
	}

	for _, scopeID := range expNoCoin {
		denom := scopeID.Denom()
		supply := s.app.BankKeeper.GetSupply(s.ctx, denom)
		s.Assert().Equal("0"+denom, supply.String(), "GetSupply(%q)", denom)
	}

	for i, addr := range addrs {
		accAddr := sdk.MustAccAddressFromBech32(addr)
		actBal := s.app.BankKeeper.GetAllBalances(s.ctx, accAddr)
		for _, expCoin := range expBals[addr] {
			found, actCoin := actBal.Find(expCoin.Denom)
			if s.Assert().True(found, "[%d]%q: found bool from actBal.Find(%q)", i, addr, expCoin.Denom) {
				s.Assert().Equal(expCoin.String(), actCoin.String(), "[%d]%q: balance coin for %q", i, addr, expCoin.Denom)
			}
		}
	}

	mdStore = s.ctx.KVStore(s.app.GetKey(metadatatypes.ModuleName))
	for _, ind := range expDelInds {
		has := mdStore.Has(ind)
		s.Assert().False(has, "mdStore.Has(%v) after running the migrations", ind)
	}
}
