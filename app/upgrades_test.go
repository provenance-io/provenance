package app

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/suite"

	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	msgfeetypes "github.com/provenance-io/provenance/x/msgfees/types"
)

type UpgradeTestSuite struct {
	suite.Suite

	app *App
	ctx sdk.Context

	logBuffer bytes.Buffer
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
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{})
}

// GetLogOutput gets the log buffer contents. This (probably) also clears the log buffer.
func (s *UpgradeTestSuite) GetLogOutput(msg string, args ...interface{}) string {
	logOutput := s.logBuffer.String()
	s.T().Logf(msg+" log output:\n%s", append(args, logOutput)...)
	return logOutput
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

	if !s.Assert().Contains(upgrades, key, "defined upgrades map") {
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
	rv = s.AssertLogContents(logOutput, expInLog, expNotInLog, msgFormat) && rv

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
func (s *UpgradeTestSuite) AssertLogContents(logOutput string, expInLog, expNotInLog []string, msg string, args ...interface{}) bool {
	s.T().Helper()
	rv := s.assertLogLinesInOrder(logOutput, expInLog, msg+" log output", args...)
	rv = s.assertLogDoesNotContain(logOutput, expNotInLog, msg+" log output", args...) && rv
	return rv
}

// assertLogLinesInOrder asserts that the log output has whole lines matching each of the expInLog entries,
// and that they're in the same order (allowing for extra lines to be in the log output that aren't in expInLog).
// Designed for AssertLogContents, please use that.
func (s *UpgradeTestSuite) assertLogLinesInOrder(logOutput string, expInLog []string, msg string, args ...interface{}) bool {
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
		`INF Setting "accountdata" name record.`,
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
		`INF Setting "accountdata" name record.`,
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

	startingMsg := `INF Creating message fee for "` + v1TypeURL + `" if it doesn't already exist.`
	successMsg := func(amt string) string {
		return `INF Successfully set fee for "` + v1TypeURL + `" with amount "` + amt + `".`
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
				`INF Message fee for "` + v1TypeURL + `" already exists with amount "88foocoin". Nothing to do.`,
			},
			expAmt: *coin("foocoin", 88),
		},
		{
			name:    "v1beta1 exists",
			v1B1Amt: coin("betacoin", 99),
			expInLog: []string{
				startingMsg,
				`INF Copying "` + v1B1TypeURL + `" fee to "` + v1TypeURL + `".`,
				successMsg("99betacoin"),
			},
			expAmt: *coin("betacoin", 99),
		},
		{
			name: "brand new",
			expInLog: []string{
				startingMsg,
				`INF Creating "` + v1TypeURL + `" fee.`,
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
			didNotPanic := s.Assert().NotPanics(testFunc, "addGovV1SubmitFee")
			logOutput := s.GetLogOutput("addGovV1SubmitFee")
			if !didNotPanic {
				return
			}

			// Make sure the log has the expected lines.
			s.AssertLogContents(logOutput, tc.expInLog, nil, "addGovV1SubmitFee")

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
	startingMsg := `INF Removing message fee for "` + typeURL + `" if one exists.`

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
				`INF Message fee for "` + typeURL + `" already does not exist. Nothing to do.`,
			},
		},
		{
			name: "exists",
			amt:  coin("p8ecoin", 808),
			expInLog: []string{
				startingMsg,
				`INF Successfully removed message fee for "` + typeURL + `" with amount "808p8ecoin".`,
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

			// Reset the log buffer to clear out unrelated entries.
			s.logBuffer.Reset()
			// Call removeP8eMemorializeContractFee and relog its output (to help if things fail).
			testFunc := func() {
				removeP8eMemorializeContractFee(s.ctx, s.app)
			}
			didNotPanic := s.Assert().NotPanics(testFunc, "removeP8eMemorializeContractFee")
			logOutput := s.GetLogOutput("removeP8eMemorializeContractFee")
			if !didNotPanic {
				return
			}

			s.AssertLogContents(logOutput, tc.expInLog, nil, "removeP8eMemorializeContractFee")

			// Make sure there isn't a fee anymore.
			fee, err := s.app.MsgFeesKeeper.GetMsgFee(s.ctx, typeURL)
			s.Require().NoError(err, "GetMsgFee(%q) error", typeURL)
			s.Require().Nil(fee, "GetMsgFee(%q) value", typeURL)
		})
	}
}

func (s *UpgradeTestSuite) TestSetAccountDataNameRecord() {
	// Most of the testing should (hopefully) be done in the name module. So this is
	// just a superficial test that makes sure it's doing something.
	// Since the name is also created during InitGenesis, it should already be set up as needed.
	// So in this unit test, the logs will indicate that.
	expInLog := []string{
		`INF Setting "accountdata" name record.`,
		`INF The "accountdata" name record already exists as needed. Nothing to do.`,
	}
	// During an actual upgrade, that last line would instead be this:
	// `INF Successfully set "accountdata" name record.`

	// Reset the log buffer to clear out unrelated entries.
	s.logBuffer.Reset()
	// Call setAccountDataNameRecord and relog its output (to help if things fail).
	var err error
	testFunc := func() {
		err = setAccountDataNameRecord(s.ctx, s.app.AccountKeeper, &s.app.NameKeeper)
	}
	didNotPanic := s.Assert().NotPanics(testFunc, "setAccountDataNameRecord")
	logOutput := s.GetLogOutput("setAccountDataNameRecord")
	if !didNotPanic {
		return
	}
	s.Require().NoError(err, "setAccountDataNameRecord")
	s.AssertLogContents(logOutput, expInLog, nil, "setAccountDataNameRecord")
}
