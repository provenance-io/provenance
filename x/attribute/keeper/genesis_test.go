package keeper_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/provenance-io/provenance/x/attribute/keeper"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/attribute/types"
	nametypes "github.com/provenance-io/provenance/x/name/types"
)

type GenesisTestSuite struct {
	suite.Suite

	app *app.App
	ctx sdk.Context

	logBuffer bytes.Buffer
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}

func (s *GenesisTestSuite) SetupSuite() {
	// Alert: This function is SetupSuite. That means all tests in here
	// will use the same app with the same store and data.
	defer app.SetLoggerMaker(app.SetLoggerMaker(app.BufferedInfoLoggerMaker(&s.logBuffer)))
	s.app = app.Setup(s.T())
	s.logBuffer.Reset()
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{})
}

// newMockNameKeeper creates a mockNameKeeper backed by this suite's app's name keeper.
func (s *GenesisTestSuite) newMockNameKeeper() *mockNameKeeper {
	return newMockNameKeeper(&s.app.NameKeeper)
}

func (s *GenesisTestSuite) TestInitGenesisModAcctAndNameRecord() {
	// Note: InitGenesis will have already been called during app setup.
	// So this is basically making sure it can be called again without blowing things up.
	// It also ensures that TestEnsureModuleAccountAndAccountDataNameRecord is called
	// during InitGenesis, but without anything being logged.

	s.Run("InitGenesis does not panic", func() {
		s.logBuffer.Reset()
		testFunc := func() {
			s.app.AttributeKeeper.InitGenesis(s.ctx, types.DefaultGenesisState())
		}
		s.Require().NotPanics(testFunc, "attribute module InitGenesis")
		logged := s.logBuffer.String()
		s.Assert().Empty(logged, "stuff logged during InitGenesis")
	})

	s.Run("attribute module account exists", func() {
		modAddr := authtypes.NewModuleAddress(types.ModuleName)
		acct := s.app.AccountKeeper.GetAccount(s.ctx, modAddr)
		s.Require().NotNil(acct, "GetAccount(%q) (%s module account address)", modAddr.String(), types.ModuleName)
		modAcct, isModAcct := acct.(authtypes.ModuleAccountI)
		s.Require().True(isModAcct, "can cast %T to authtypes.ModuleAccountI", acct)
		s.Assert().Equal(types.ModuleName, modAcct.GetName(), "module account name")
	})

	s.Run("account data name exists", func() {
		modAddr := authtypes.NewModuleAddress(types.ModuleName).String()
		record, err := s.app.NameKeeper.GetRecordByName(s.ctx, types.AccountDataName)
		s.Require().NoError(err, "GetRecordByName(%q) error", types.AccountDataName)
		s.Require().NotNil(record, "GetRecordByName(%q) record", types.AccountDataName)
		s.Assert().Equal(types.AccountDataName, record.Name, "record.Name")
		s.Assert().Equal(modAddr, record.Address, "record.Address")
		s.Assert().True(record.Restricted, "record.Restricted")
	})
}

func (s *GenesisTestSuite) TestEnsureModuleAccountAndAccountDataNameRecord() {
	recordName := types.AccountDataName
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
			nameK:       s.newMockNameKeeper().WithGetRecordByNameError("cannot get that thing for you"),
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
			nameK:    s.newMockNameKeeper().WithUpdateNameRecordError("that update is not going to be allowed here"),
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
			var nameK types.NameKeeper = &s.app.NameKeeper
			if tc.nameK != nil {
				nameK = tc.nameK
			}

			// Call setAccountDataNameRecord, making sure it doesn't panic and copy log output to test output.
			s.logBuffer.Reset()
			var err error
			testFunc := func() {
				err = keeper.EnsureModuleAccountAndAccountDataNameRecord(s.ctx, s.app.AccountKeeper, nameK)
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
