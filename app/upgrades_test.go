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

	attributetypes "github.com/provenance-io/provenance/x/attribute/types"
	msgfeetypes "github.com/provenance-io/provenance/x/msgfees/types"
	namekeeper "github.com/provenance-io/provenance/x/name/keeper"
	nametypes "github.com/provenance-io/provenance/x/name/types"
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
		`INF Setting "accountdata" name record.`,
		`INF Creating message fee for "/cosmos.gov.v1.MsgSubmitProposal" if it doesn't already exist.`,
		`INF Removing message fee for "/provenance.metadata.v1.MsgP8eMemorializeContractRequest" if one exists.`,
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

type mockNameKeeper struct {
	Parent                namekeeper.Keeper
	SetNameRecordError    string
	GetRecordNameError    string
	UpdateNameRecordError string
}

var _ nameKeeper = (*mockNameKeeper)(nil)

func (s *UpgradeTestSuite) newMockNameKeeper() *mockNameKeeper {
	return &mockNameKeeper{Parent: s.app.NameKeeper}
}

func (k *mockNameKeeper) WithSetNameRecordError(err string) *mockNameKeeper {
	k.SetNameRecordError = err
	return k
}

func (k *mockNameKeeper) WithGetRecordNameError(err string) *mockNameKeeper {
	k.GetRecordNameError = err
	return k
}

func (k *mockNameKeeper) WithUpdateNameRecord(err string) *mockNameKeeper {
	k.UpdateNameRecordError = err
	return k
}

func (k *mockNameKeeper) NameExists(ctx sdk.Context, name string) bool {
	return k.Parent.NameExists(ctx, name)
}

func (k *mockNameKeeper) GetRecordByName(ctx sdk.Context, name string) (record *nametypes.NameRecord, err error) {
	if len(k.GetRecordNameError) > 0 {
		return nil, errors.New(k.GetRecordNameError)
	}
	return k.Parent.GetRecordByName(ctx, name)
}

func (k *mockNameKeeper) SetNameRecord(ctx sdk.Context, name string, addr sdk.AccAddress, restrict bool) error {
	if len(k.SetNameRecordError) > 0 {
		return errors.New(k.SetNameRecordError)
	}
	return k.Parent.SetNameRecord(ctx, name, addr, restrict)
}

func (k *mockNameKeeper) UpdateNameRecord(ctx sdk.Context, name string, addr sdk.AccAddress, restrict bool) error {
	if len(k.UpdateNameRecordError) > 0 {
		return errors.New(k.UpdateNameRecordError)
	}
	return k.Parent.UpdateNameRecord(ctx, name, addr, restrict)
}

func (s *UpgradeTestSuite) TestSetAccountDataNameRecord() {
	recordName := attributetypes.AccountDataName
	// This is automatically added to all expInLog slices.
	startMsg := `Setting "` + recordName + `" name record.`
	// This is automatically added to expInLog for tests that expect an error.
	errMsg := `Error setting "` + recordName + `" name record.`

	attrModAccAddr := "cosmos14y4l6qky2zhsxcx7540ejqvtye7fr66d3vt0ye"
	otherAddr := "cosmos1da6xsetjtaskgerjv4ehxh6lta047h6l3cc9z2" // sdk.AccAddress("other_address_______").String()

	nr := func(addr string, rest bool) *nametypes.NameRecord {
		return &nametypes.NameRecord{
			Name:       "", // Field not used below anyway.
			Address:    addr,
			Restricted: rest,
		}
	}

	tests := []struct {
		name        string
		existing    *nametypes.NameRecord // The Name field is ignored in this.
		nameK       *mockNameKeeper
		expErr      string
		expInLog    []string
		expNotInLog []string
	}{
		{
			name:     "name doesn't exist yet",
			expInLog: []string{`Successfully set "` + recordName + `" name record.`},
		},
		{
			name:        "name doesn't exist yet but error writing it",
			nameK:       s.newMockNameKeeper().WithSetNameRecordError("no writo for you-oh"),
			expErr:      "no writo for you-oh",
			expNotInLog: []string{`Successfully set "` + recordName + `" name record.`},
		},
		{
			name:        "already exists but error getting it",
			existing:    nr(attrModAccAddr, true),
			nameK:       s.newMockNameKeeper().WithGetRecordNameError("cannot get that thing for you"),
			expErr:      "cannot get that thing for you",
			expNotInLog: []string{`The "` + recordName + `" name record already exists as needed. Nothing to do.`},
		},
		{
			name:     "existing is as needed",
			existing: nr(attrModAccAddr, true),
			expInLog: []string{`The "` + recordName + `" name record already exists as needed. Nothing to do.`},
		},
		{
			name:     "existing is unrestricted",
			existing: nr(attrModAccAddr, false),
			expInLog: []string{
				`Existing "` + recordName + `" name record is not restricted. It will be updated to be restricted.`,
				`Updating existing "` + recordName + `" name record.`,
				`Successfully updated "` + recordName + `" name record.`,
			},
			expNotInLog: []string{
				`It will be updated to the attribute module account address "` + attrModAccAddr + `"`,
			},
		},
		{
			name:     "existing has different address",
			existing: nr(otherAddr, true),
			expInLog: []string{
				`Existing "` + recordName + `" name record has address "` + otherAddr + `". It will be updated to the attribute module account address "` + attrModAccAddr + `"`,
				`Updating existing "` + recordName + `" name record.`,
				`Successfully updated "` + recordName + `" name record.`,
			},
			expNotInLog: []string{
				`Existing "` + recordName + `" name record is not restricted. It will be updated to be restricted.`,
			},
		},
		{
			name:     "existing is unrestricted and has different address",
			existing: nr(otherAddr, false),
			expInLog: []string{
				`Existing "` + recordName + `" name record is not restricted. It will be updated to be restricted.`,
				`Existing "` + recordName + `" name record has address "` + otherAddr + `". It will be updated to the attribute module account address "` + attrModAccAddr + `"`,
				`Updating existing "` + recordName + `" name record.`,
				`Successfully updated "` + recordName + `" name record.`,
			},
		},
		{
			name:     "existing needs update but error updating it",
			existing: nr(otherAddr, false),
			nameK:    s.newMockNameKeeper().WithUpdateNameRecord("that update is not going to be allowed here"),
			expErr:   "that update is not going to be allowed here",
			expInLog: []string{
				`Existing "` + recordName + `" name record is not restricted. It will be updated to be restricted.`,
				`Existing "` + recordName + `" name record has address "` + otherAddr + `". It will be updated to the attribute module account address "` + attrModAccAddr + `"`,
				`Updating existing "` + recordName + `" name record.`,
			},
			expNotInLog: []string{`Successfully updated "` + recordName + `" name record.`},
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			// If the name record already exists, delete it now.
			if s.app.NameKeeper.NameExists(s.ctx, recordName) {
				err := s.app.NameKeeper.DeleteRecord(s.ctx, recordName)
				s.Require().NoError(err, "DeleteRecord")
			}

			// Now, if the test calls for an existing entry, create it.
			if tc.existing != nil {
				exiAddr, err := sdk.AccAddressFromBech32(tc.existing.Address)
				s.Require().NoError(err, "converting %q to AccAddress", tc.existing.Address)
				err = s.app.NameKeeper.SetNameRecord(s.ctx, recordName, exiAddr, tc.existing.Restricted)
				s.Require().NoError(err, "SetNameRecord")
			}

			// Define what's expected/not expected in the log.
			expInLog := append([]string{startMsg}, tc.expInLog...)
			expNotInLog := make([]string, len(tc.expNotInLog), len(tc.expNotInLog)+1)
			copy(expNotInLog, tc.expNotInLog)
			if len(tc.expErr) > 0 {
				expInLog = append(expInLog, fmt.Sprintf("ERR %s error=%q", errMsg, tc.expErr))
			} else {
				expNotInLog = append(expNotInLog, errMsg)
			}

			// Use the mock name keeper if one is defined, otherwise, use the normal one.
			var nameK nameKeeper = s.app.NameKeeper
			if tc.nameK != nil {
				nameK = tc.nameK
			}

			// Call setAccountDataNameRecord, making sure it doesn't panic and copy log output to test output.
			s.logBuffer.Reset()
			var err error
			testFunc := func() {
				err = setAccountDataNameRecord(s.ctx, s.app.AccountKeeper, nameK)
			}
			s.Require().NotPanics(testFunc, "setAccountDataNameRecord")
			logOutput := s.logBuffer.String()
			s.T().Logf("setAccountDataNameRecord log output:\n%s", logOutput)

			// Require the error to be as expected before checking other parts.
			if len(tc.expErr) > 0 {
				s.Require().EqualError(err, tc.expErr, "setAccountDataNameRecord error")
			} else {
				s.Require().NoError(err, "setAccountDataNameRecord error")
			}

			// Make sure the log output is as expected.
			for _, exp := range tc.expInLog {
				s.Assert().Contains(logOutput, exp, "setAccountDataNameRecord log output")
			}
			for _, unexp := range expNotInLog {
				s.Assert().NotContains(logOutput, unexp, "setAccountDataNameRecord log output")
			}

			// Make sure the name record exists as needed.
			if len(tc.expErr) == 0 {
				nameRecord, err := s.app.NameKeeper.GetRecordByName(s.ctx, recordName)
				if s.Assert().NoError(err, "GetRecordByName after setAccountDataNameRecord") {
					s.Assert().Equal(recordName, nameRecord.Name, "NameRecord.Name after setAccountDataNameRecord")
					s.Assert().Equal(attrModAccAddr, nameRecord.Address, "NameRecord.Address after setAccountDataNameRecord")
					s.Assert().True(nameRecord.Restricted, "NameRecord.Restricted after setAccountDataNameRecord")
				}
			}
		})
	}
}
