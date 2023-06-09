package keeper_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v2"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/provenance-io/provenance/app"
	attrtypes "github.com/provenance-io/provenance/x/attribute/types"
	namekeeper "github.com/provenance-io/provenance/x/name/keeper"
	nametypes "github.com/provenance-io/provenance/x/name/types"
)

type KeeperTestSuite struct {
	suite.Suite

	app *app.App
	ctx sdk.Context

	pubkey1   cryptotypes.PubKey
	user1     string
	user1Addr sdk.AccAddress

	pubkey2   cryptotypes.PubKey
	user2     string
	user2Addr sdk.AccAddress

	msgSrvr nametypes.MsgServer
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest() {
	app := app.Setup(s.T())
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	s.app = app
	s.ctx = ctx
	s.pubkey1 = secp256k1.GenPrivKey().PubKey()
	s.user1Addr = sdk.AccAddress(s.pubkey1.Address())
	s.user1 = s.user1Addr.String()

	s.pubkey2 = secp256k1.GenPrivKey().PubKey()
	s.user2Addr = sdk.AccAddress(s.pubkey2.Address())
	s.user2 = s.user2Addr.String()

	var nameData nametypes.GenesisState
	nameData.Bindings = append(nameData.Bindings, nametypes.NewNameRecord("name", s.user1Addr, false))
	nameData.Bindings = append(nameData.Bindings, nametypes.NewNameRecord("example.name", s.user1Addr, false))
	nameData.Params.AllowUnrestrictedNames = false
	nameData.Params.MaxNameLevels = 16
	nameData.Params.MinSegmentLength = 2
	nameData.Params.MaxSegmentLength = 16

	app.NameKeeper.InitGenesis(ctx, nameData)

	s.app.AccountKeeper.SetAccount(s.ctx, s.app.AccountKeeper.NewAccountWithAddress(s.ctx, s.user1Addr))
	s.app.AccountKeeper.SetAccount(s.ctx, s.app.AccountKeeper.NewAccountWithAddress(s.ctx, s.user2Addr))
}

func (s *KeeperTestSuite) TestSetup() {
	s.Run("verify test setup params", func() {
		s.Require().False(s.app.NameKeeper.GetAllowUnrestrictedNames(s.ctx))
		s.Require().Equal(uint32(16), s.app.NameKeeper.GetMaxNameLevels(s.ctx))
		s.Require().Equal(uint32(2), s.app.NameKeeper.GetMinSegmentLength(s.ctx))
		s.Require().Equal(uint32(16), s.app.NameKeeper.GetMaxSegmentLength(s.ctx))
	})
	s.Run("verify get all test setup params", func() {
		p := s.app.NameKeeper.GetParams(s.ctx)
		s.Require().NotNil(p)
		s.Require().False(p.AllowUnrestrictedNames)
		s.Require().Equal(uint32(16), p.MaxNameLevels)
		s.Require().Equal(uint32(2), p.MinSegmentLength)
		s.Require().Equal(uint32(16), p.MaxSegmentLength)
	})

	expOut := fmt.Sprintf(`params:
  maxsegmentlength: 16
  minsegmentlength: 2
  maxnamelevels: 16
  allowunrestrictednames: false
bindings:
- name: name
  address: %[1]s
  restricted: false
- name: example.name
  address: %[1]s
  restricted: false
- name: %[2]s
  address: %[3]s
  restricted: true
`,
		s.user1Addr.String(), attrtypes.AccountDataName, authtypes.NewModuleAddress(attrtypes.ModuleName).String())

	gen := s.app.NameKeeper.ExportGenesis(s.ctx)
	out, err := yaml.Marshal(gen)
	s.Require().NoError(err)
	s.Require().Equal(expOut, string(out))
}

func (s *KeeperTestSuite) TestNameNormalization() {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// Valid names
		{"normalize upper case", args{name: "TEST.NORMALIZE.PIO"}, "test.normalize.pio", false},
		{"trim comp spaces", args{name: "test . normalize. pio "}, "test.normalize.pio", false},
		{"allow single dash per comp", args{name: "test-field.my-service.pio"}, "test-field.my-service.pio", false},
		{"allow digits", args{name: "test.normalize.v1.pio"}, "test.normalize.v1.pio", false},
		{"allow unicode chars", args{name: "tœst.nørmålize.v1.pio"}, "tœst.nørmålize.v1.pio", false},
		{"allow uuid as comp", args{name: "6443a1e8-ec9b-4ff1-b200-d639424bcba4.service.pb"},
			"6443a1e8-ec9b-4ff1-b200-d639424bcba4.service.pb", false},
		// Invalid names / components
		{"fail on empty name", args{name: ""}, "", true},
		{"fail when too short", args{name: "z"}, "", true},
		{"fail when too long", args{name: "too.looooooooooooooooooooooooooooooooooooooong.pio"}, "", true},
		{"fail on multiple dashes in comp", args{name: "fail-test-field.my-app.pio"}, "", true},
		{"fail on non-printable chars", args{name: "test.normalize" + string([]byte{0x01}) + ".pio"}, "", true},
		{"fail on too many components", args{name: "ab.bc.cd.de.ef.fg.gh.hi.ij.jk.kl.lm.mn.no.op.pq.qr"}, "", true},
		{"fail on unsupported chars", args{name: "fail_normalize.pio"}, "", true},
		{"fail on unsupported chars", args{name: "fail!normalize.pio"}, "", true},
		{"fail on unsupported chars", args{name: "fail|normalize.pio"}, "", true},
		{"fail on unsupported chars", args{name: "fail,normalize.pio"}, "", true},
		{"fail on unsupported chars", args{name: "fail~normalize.pio"}, "", true},
		{"fail on unsupported chars", args{name: "fail*normalize.pio"}, "", true},
		{"fail on unsupported chars", args{name: "fail&normalize.pio"}, "", true},
		{"fail on unsupported chars", args{name: "fail^normalize.pio"}, "", true},
		{"fail on unsupported chars", args{name: "fail@normalize.pio"}, "", true},
		{"fail on unsupported chars", args{name: "fail#normalize.pio"}, "", true},
		{"fail on unsupported chars", args{name: "fail=normalize.pio"}, "", true},
		{"fail on unsupported chars", args{name: "fail+normalize.pio"}, "", true},
		{"fail on unsupported chars", args{name: "fail`normalize.pio"}, "", true},
		{"fail on unsupported chars", args{name: "fail%normalize.pio"}, "", true},
		{"fail on invalid uuid", args{name: "6443a1e8-ec9b-4ff1-b200-d639424bcba4-deadbeef.service.pb"}, "", true},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			got, err := s.app.NameKeeper.Normalize(s.ctx, tt.args.name)
			if tt.wantErr {
				s.Error(err)
			} else {
				s.NoError(err)
				s.Require().Equal(tt.want, got)
			}
		})
	}
}

func (s *KeeperTestSuite) TestSetName() {
	cases := map[string]struct {
		recordName     string
		recordRestrict bool
		accAddr        sdk.AccAddress
		wantErr        bool
		errorMsg       string
	}{
		"should successfully add name": {
			recordName:     "new.name",
			recordRestrict: true,
			accAddr:        s.user1Addr,
			wantErr:        false,
			errorMsg:       "",
		},
		"invalid name": {
			recordName:     "fail!!.name",
			recordRestrict: true,
			accAddr:        s.user1Addr,
			wantErr:        true,
			errorMsg:       "value provided for name is invalid",
		},
		"bad address": {
			recordName:     "bad.address.name",
			recordRestrict: true,
			accAddr:        sdk.AccAddress{},
			wantErr:        true,
			errorMsg:       "addresses cannot be empty: unknown address: invalid account address",
		},
		"name already bound": {
			recordName:     "name",
			recordRestrict: true,
			accAddr:        s.user1Addr,
			wantErr:        true,
			errorMsg:       "name is already bound to an address",
		},
	}
	for n, tc := range cases {
		tc := tc

		s.Run(n, func() {
			err := s.app.NameKeeper.SetNameRecord(s.ctx, tc.recordName, tc.accAddr, tc.recordRestrict)
			if tc.wantErr {
				s.Error(err)
				s.Equal(tc.errorMsg, err.Error())
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGetName() {
	s.Run("get valid root name", func() {
		r, err := s.app.NameKeeper.GetRecordByName(s.ctx, "name")
		s.Require().NoError(err)
		s.Require().Equal("name", r.Name)
		s.Require().False(r.Restricted)
		s.Require().True(s.app.NameKeeper.ResolvesTo(s.ctx, "name", s.user1Addr))
		s.Require().True(s.app.NameKeeper.NameExists(s.ctx, "name"))
		s.Require().Equal(s.user1Addr.String(), r.Address)
	})
	s.Run("get valid sub name", func() {
		r, err := s.app.NameKeeper.GetRecordByName(s.ctx, "example.name")
		s.Require().NoError(err)
		s.Require().Equal("example.name", r.Name)
		s.Require().False(r.Restricted)
		s.Require().Equal(s.user1Addr.String(), r.Address)
	})
	s.Run("get invalid name", func() {
		r, err := s.app.NameKeeper.GetRecordByName(s.ctx, "undefined.name")
		s.Require().Error(err)
		s.Require().Nil(r)
		s.Require().Equal("no address bound to name", err.Error())
		s.Require().False(s.app.NameKeeper.NameExists(s.ctx, "undefined.name"))
		s.Require().False(s.app.NameKeeper.ResolvesTo(s.ctx, "undefined.name", s.user1Addr))
	})
	s.Run("get missing segment name", func() {
		r, err := s.app.NameKeeper.GetRecordByName(s.ctx, "..name")
		s.Require().Error(err)
		s.Require().Nil(r)
		s.Require().Equal("name segment cannot be empty: value provided for name is invalid", err.Error())
		s.Require().False(s.app.NameKeeper.NameExists(s.ctx, "..name"))
	})
}

func (s *KeeperTestSuite) TestGetAddress() {
	s.Run("get names by address", func() {
		r, err := s.app.NameKeeper.GetRecordsByAddress(s.ctx, s.user1Addr)
		s.Require().NoError(err)
		s.Require().NotEmpty(r)
	})
}

func (s *KeeperTestSuite) TestDeleteRecord() {
	s.Run("delete invalid name", func() {
		err := s.app.NameKeeper.DeleteRecord(s.ctx, "undefined.name")
		s.Require().Error(err)
		s.Require().Equal("no address bound to name", err.Error())
	})
	s.Run("delete valid root name", func() {
		err := s.app.NameKeeper.DeleteRecord(s.ctx, "name")
		s.Require().NoError(err)
	})
	s.Run("delete valid root sub name", func() {
		err := s.app.NameKeeper.DeleteRecord(s.ctx, "example.name")
		s.Require().NoError(err)
	})

}

func (s *KeeperTestSuite) TestModifyRecord() {
	jackthecat := "jackthecat"
	s.Run("update adds new name", func() {
		err := s.app.NameKeeper.UpdateNameRecord(s.ctx, jackthecat, s.user2Addr, true)
		s.Require().NoError(err, "UpdateNameRecord(%q, user2)", jackthecat)
		isUser2 := s.app.NameKeeper.ResolvesTo(s.ctx, jackthecat, s.user2Addr)
		s.Assert().True(isUser2, "ResolvesTo(%q, user2)", jackthecat)

		expUser2Recs := nametypes.NameRecords{
			{Name: jackthecat, Address: s.user2Addr.String(), Restricted: true},
		}
		addr2Recs, err := s.app.NameKeeper.GetRecordsByAddress(s.ctx, s.user2Addr)
		s.Require().NoError(err, "GetRecordsByAddress(user2)")
		s.Assert().Equal(expUser2Recs, addr2Recs, "GetRecordsByAddress(user2)")

	})
	s.Run("update to new owner", func() {
		err := s.app.NameKeeper.UpdateNameRecord(s.ctx, jackthecat, s.user1Addr, true)
		s.Require().NoError(err, "UpdateNameRecord(%q, user1)", jackthecat)
		isUser1 := s.app.NameKeeper.ResolvesTo(s.ctx, jackthecat, s.user1Addr)
		s.Assert().True(isUser1, "ResolvesTo(%q, user1)", jackthecat)
		isUser2 := s.app.NameKeeper.ResolvesTo(s.ctx, jackthecat, s.user2Addr)
		s.Assert().False(isUser2, "ResolvesTo(%q, user2)", jackthecat)

		expUser1Recs := nametypes.NameRecords{
			{Name: jackthecat, Address: s.user1Addr.String(), Restricted: true},
			{Name: "name", Address: s.user1Addr.String(), Restricted: false},
			{Name: "example.name", Address: s.user1Addr.String(), Restricted: false},
		}
		addr1Recs, err := s.app.NameKeeper.GetRecordsByAddress(s.ctx, s.user1Addr)
		s.Require().NoError(err, "GetRecordsByAddress(user1)")
		s.Assert().Equal(expUser1Recs, addr1Recs, "GetRecordsByAddress(user1)")

		expUser2Recs := nametypes.NameRecords{}
		addr2Recs, err := s.app.NameKeeper.GetRecordsByAddress(s.ctx, s.user2Addr)
		s.Require().NoError(err, "GetRecordsByAddress(user2)")
		s.Assert().Equal(expUser2Recs, addr2Recs, "GetRecordsByAddress(user2)")
	})
	s.Run("update has invalid address", func() {
		err := s.app.NameKeeper.UpdateNameRecord(s.ctx, "jackthecat", sdk.AccAddress{}, true)
		s.Require().Error(err)
		s.Require().Equal("addresses cannot be empty: unknown address: invalid account address", err.Error())
	})
	s.Run("update valid root name", func() {
		err := s.app.NameKeeper.UpdateNameRecord(s.ctx, "name", s.user2Addr, true)
		s.Require().NoError(err)
		record, err := s.app.NameKeeper.GetRecordByName(s.ctx, "name")
		s.Require().NoError(err)
		s.Require().Equal("name", record.Name)
		s.Require().Equal(s.user2Addr.String(), record.GetAddress())
		s.Require().Equal(true, record.Restricted)
	})
	s.Run("update valid root sub name", func() {
		err := s.app.NameKeeper.UpdateNameRecord(s.ctx, "example.name", s.user2Addr, true)
		s.Require().NoError(err)
		record, err := s.app.NameKeeper.GetRecordByName(s.ctx, "example.name")
		s.Require().NoError(err)
		s.Require().Equal("example.name", record.Name)
		s.Require().Equal(s.user2Addr.String(), record.GetAddress())
		s.Require().Equal(true, record.Restricted)
	})
}

func (s *KeeperTestSuite) TestGetAuthority() {
	s.Run("has correct authority", func() {
		authority := s.app.NameKeeper.GetAuthority()
		s.Require().Equal("cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn", authority)
	})
}

func (s *KeeperTestSuite) TestIterateRecord() {
	s.Run("iterate name's", func() {
		expRecords := nametypes.NameRecords{
			nametypes.NewNameRecord("name", s.user1Addr, false),
			nametypes.NewNameRecord("example.name", s.user1Addr, false),
			nametypes.NewNameRecord(attrtypes.AccountDataName, authtypes.NewModuleAddress(attrtypes.ModuleName), true),
		}
		records := nametypes.NameRecords{}
		// Callback func that adds records to genesis state.
		appendToRecords := func(record nametypes.NameRecord) error {
			records = append(records, record)
			return nil
		}
		// Collect and return genesis state.
		err := s.app.NameKeeper.IterateRecords(s.ctx, nametypes.NameKeyPrefix, appendToRecords)
		s.Require().NoError(err, "IterateRecords error")
		s.Require().Equal(expRecords, records, "records iterated over")
	})

}

func (s *KeeperTestSuite) TestSecp256r1KeyAlgo() {
	s.Run("should successfully add name for account with secp256r1 key", func() {
		err := s.app.NameKeeper.SetNameRecord(s.ctx, "secp256r1.name", s.user2Addr, true)
		s.NoError(err)
	})
}

func (s *KeeperTestSuite) TestAuthority() {
	require.EqualValues(s.T(), s.app.NameKeeper.GetAuthority(), "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn")
}

func (s *KeeperTestSuite) TestCreateRootName() {
	s.msgSrvr = namekeeper.NewMsgServerImpl(s.app.NameKeeper)
	msg := nametypes.MsgCreateRootNameRequest{
		Record: &nametypes.NameRecord{
			Name:       "swampmonster",
			Address:    "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
			Restricted: false,
		},
		Authority: "cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn",
	}

	s.Run("create valid root name", func() {
		_, err := s.msgSrvr.CreateRootName(s.ctx, &msg)
		s.Require().NoError(err)
	})

	s.Run("invalid authority", func() {
		msg.Authority = "..."
		_, err := s.msgSrvr.CreateRootName(s.ctx, &msg)
		s.Require().Error(err)
		s.Require().Equal("expected cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn got ...: expected gov account as only signer for proposal message", err.Error())
	})
}

func TestDeleteInvalidAddressIndexEntries(t *testing.T) {
	// Not using the suite here because:
	// a) this is only going to be around for a couple versions.
	// b) I don't want to worry about any of the name records automatically added for the suite runs.

	provApp := app.Setup(t)
	ctx := provApp.NewContext(false, tmproto.Header{})

	getRecordNames := func(records nametypes.NameRecords) []string {
		rv := make([]string, len(records))
		for i, record := range records {
			rv[i] = record.Name
		}
		return rv
	}

	// The point of this setup is that "two" will be saved (and indexed) to addr1.
	// It will then be updated (and indexed) to addr2, but the addr1 index will still exist.

	addr1 := sdk.AccAddress("addr1_______________")
	addr2 := sdk.AccAddress("addr2_______________")

	setups := []struct {
		id    string
		addr  sdk.AccAddress
		names []string
	}{
		{id: "addr1", addr: addr1, names: []string{"one", "sub.one", "two"}},
		{id: "addr2", addr: addr2, names: []string{"two", "sub.two"}},
	}

	for _, sc := range setups {
		for _, name := range sc.names {
			// Using the private addRecord method here to bypass the name-already-bound
			// check. This lets me mimic what used to happen in ModifyRecord.
			err := provApp.NameKeeper.AddRecord(ctx, name, sc.addr, false, true)
			require.NoError(t, err, "addRecord(%q, %s)", name, sc.id)
		}
	}

	// Defining these as full records because the address value is important here.
	expNameRecords := nametypes.NameRecords{
		{Name: "one", Address: addr1.String()},
		{Name: "sub.one", Address: addr1.String()},
		{Name: "two", Address: addr2.String()},
		{Name: "sub.two", Address: addr2.String()},
		{Name: attrtypes.AccountDataName, Address: authtypes.NewModuleAddress(attrtypes.ModuleName).String(), Restricted: true},
	}

	// For these, all we care about are the names.
	addr1ExpNames := []string{"one", "sub.one"}
	addr2ExpNames := []string{"two", "sub.two"}

	tests := []struct {
		name          string
		expLog        []string
		expAddr1Names []string
		expAddr2Names []string
	}{
		{
			// Sanity check. DeleteInvalidAddressIndexEntries isn't run on first test case.
			// Make sure there's a bad entry in the addr1 names.
			name:          "initial state sanity check",
			expAddr1Names: append(addr1ExpNames, "two"),
			expAddr2Names: addr2ExpNames,
		},
		{
			// DeleteInvalidAddressIndexEntries will be called first.
			// There should be one bad entry to delete (addr1 -> "two").
			name: "first run - deletes one",
			expLog: []string{
				"Checking address -> name index entries.",
				"Found 1 invalid address -> name index entries. Deleting them now.",
				fmt.Sprintf("Done checking address -> name index entries. Deleted 1 invalid entries and kept %d valid entries.", len(expNameRecords)),
			},
			expAddr1Names: addr1ExpNames,
			expAddr2Names: addr2ExpNames,
		},
		{
			// DeleteInvalidAddressIndexEntries will be called again.
			// This time, all is good, so there shouldn't be anything to delete.
			name: "second run - all ok already",
			expLog: []string{
				"Checking address -> name index entries.",
				fmt.Sprintf("Done checking address -> name index entries. All %d entries are valid", len(expNameRecords)),
			},
			expAddr1Names: addr1ExpNames,
			expAddr2Names: addr2ExpNames,
		},
	}

	for i, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// All the log lines are info level and have the module name at the end.
			for l, expLog := range tc.expLog {
				tc.expLog[l] = "INF " + expLog + " module=x/name"
			}
			if i != 0 {
				// Call the DeleteInvalidAddressIndexEntries function.
				// Use a custom logger that goes to a buffer, so I can see what was logged.
				var logBuffer bytes.Buffer
				lw := zerolog.ConsoleWriter{
					Out:          &logBuffer,
					NoColor:      true,
					PartsExclude: []string{"time"}, // Without this, each line starts with "<nil> "
				}
				// Error log lines will start with "ERR ".
				// Info log lines will start with "INF ".
				// Debug log lines are omitted, but would start with "DBG ".
				logger := server.ZeroLogWrapper{Logger: zerolog.New(lw).Level(zerolog.InfoLevel)}

				// And use a fresh event manager.
				em := sdk.NewEventManager()

				tctx := provApp.NewContext(false, tmproto.Header{}).WithEventManager(em).WithLogger(logger)
				testFunc := func() {
					provApp.NameKeeper.DeleteInvalidAddressIndexEntries(tctx)
				}
				require.NotPanics(t, testFunc, "DeleteInvalidAddressIndexEntries")

				// Get the log output and make sure it's as expected.
				logOut := logBuffer.String()
				t.Logf("DeleteInvalidAddressIndexEntries log output:\n%s", logOut)
				actLog1Lines := strings.Split(logOut, "\n")
				// Delete the last entry if it's just an empty string.
				if len(actLog1Lines[len(actLog1Lines)-1]) == 0 {
					actLog1Lines = actLog1Lines[:len(actLog1Lines)-1]
				}
				assert.Equal(t, tc.expLog, actLog1Lines, "logged output")

				// Make sure no events were emitted.
				events := em.Events()
				assert.Len(t, events, 0, "emitted events")
			}

			// Get all the records and make sure they're as expected.
			var allRecords nametypes.NameRecords
			err := provApp.NameKeeper.IterateRecords(ctx, nametypes.NameKeyPrefix, func(record nametypes.NameRecord) error {
				allRecords = append(allRecords, record)
				return nil
			})
			require.NoError(t, err, "IterateRecords by name")
			assert.ElementsMatch(t, expNameRecords, allRecords, "name records: expected (A) vs actual (B)")

			// Get all the addr1 name records and make sure they're as expected.
			addr1ActRecords, err := provApp.NameKeeper.GetRecordsByAddress(ctx, addr1)
			require.NoError(t, err, "GetRecordsByAddress(addr1)")
			addr1ActNames := getRecordNames(addr1ActRecords)
			require.ElementsMatch(t, tc.expAddr1Names, addr1ActNames, "addr1 names: expected (A) vs actual (B)")

			// Get all the addr2 name records and make sure they're as expected.
			addr2ActRecords, err := provApp.NameKeeper.GetRecordsByAddress(ctx, addr2)
			require.NoError(t, err, "GetRecordsByAddress(addr2)")
			addr2ActNames := getRecordNames(addr2ActRecords)
			require.ElementsMatch(t, tc.expAddr2Names, addr2ActNames, "addr2 names: expected (A) vs actual (B)")
		})
	}
}
