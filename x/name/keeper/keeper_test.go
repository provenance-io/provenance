package keeper_test

import (
	"fmt"
	"testing"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"gopkg.in/yaml.v2"

	"github.com/provenance-io/provenance/app"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

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
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest() {
	app := app.Setup(false)
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
	gen := s.app.NameKeeper.ExportGenesis(s.ctx)
	out, err := yaml.Marshal(gen)
	s.Require().NoError(err)
	s.Require().Equal(fmt.Sprintf(`params:
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
`, s.user1Addr.String()), string(out))
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

func (s *KeeperTestSuite) TestIterateRecord() {
	s.Run("iterate name's", func() {
		records := nametypes.NameRecords{}
		// Callback func that adds records to genesis state.
		appendToRecords := func(record nametypes.NameRecord) error {
			records = append(records, record)
			return nil
		}
		// Collect and return genesis state.
		err := s.app.NameKeeper.IterateRecords(s.ctx, nametypes.NameKeyPrefix, appendToRecords)
		s.Require().NoError(err)
		s.Require().Equal(2, len(records))
	})

}

func (s *KeeperTestSuite) TestSecp256r1KeyAlgo() {
	s.Run("should successfully add name for account with secp256r1 key", func() {
		err := s.app.NameKeeper.SetNameRecord(s.ctx, "secp256r1.name", s.user2Addr, true)
		s.NoError(err)
	})
}
