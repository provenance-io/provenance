package keeper_test

import (
	"fmt"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/provenance-io/provenance/x/metadata/types"
	"testing"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"gopkg.in/yaml.v2"

	"github.com/provenance-io/provenance/app"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	nametypes "github.com/provenance-io/provenance/x/name/types"
	"github.com/stretchr/testify/suite"
)

type KeeperLegacyTestSuite struct {
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

func TestKeeperLegacyTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperLegacyTestSuite))
}

func (s *KeeperLegacyTestSuite) SetupTest() {
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

	//set records the old way, i.e legacy amino
	InitGenesisLegacy(ctx, nameData, s.app)
	//convert it over to proto
	// now below test's should be able to get these old records
	//export etc should work the same way
	// test's fail if you don't do the conversion.
	s.app.NameKeeper.ConvertLegacyAmino(ctx)
	s.app.AccountKeeper.SetAccount(s.ctx, s.app.AccountKeeper.NewAccountWithAddress(s.ctx, s.user1Addr))
}

// InitGenesis creates the initial genesis state for the name module.THIS IS ONLY FOR TESTING THIS RELEASE
func InitGenesisLegacy(ctx sdk.Context, data nametypes.GenesisState, app *app.App) {
	app.NameKeeper.SetParams(ctx, data.Params)
	for _, record := range data.Bindings {
		addr, err := sdk.AccAddressFromBech32(record.Address)
		if err != nil {
			panic(err)
		}
		if err := SetNameRecord(ctx, record.Name, addr, record.Restricted, app); err != nil {
			panic(err)
		}
	}
}

func (s *KeeperLegacyTestSuite) TestSetupLegacy() {
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

func (s *KeeperLegacyTestSuite) TestGetNameLegacy() {
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

func (s *KeeperLegacyTestSuite) TestGetAddressLegacy() {
	s.Run("get names by address", func() {
		r, err := s.app.NameKeeper.GetRecordsByAddress(s.ctx, s.user1Addr)
		s.Require().NoError(err)
		s.Require().NotEmpty(r)
	})
}

func (s *KeeperLegacyTestSuite) TestDeleteRecordLegacy() {
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

// make sure nothing bad happends when old name's were set in legacy and converted
func (s *KeeperLegacyTestSuite) TestSetName() {
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
			errorMsg:       "incorrect address length (expected: 20, actual: 0): invalid account address",
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

// THIS IS ONLY FOR TESTING LEGACY AMINO stored obh
func SetNameRecord(ctx sdk.Context, name string, addr sdk.AccAddress, restrict bool, app *app.App) error {
	var err error
	if name, err = app.NameKeeper.Normalize(ctx, name); err != nil {
		return err
	}
	if err = sdk.VerifyAddressFormat(addr); err != nil {
		return sdkerrors.Wrap(types.ErrInvalidAddress, err.Error())
	}
	key, err := nametypes.GetNameKeyPrefixLegacyAmino(name)
	if err != nil {
		return err
	}
	store := ctx.KVStore(app.GetKey(nametypes.ModuleName))
	if store.Has(key) {
		return nametypes.ErrNameAlreadyBound
	}
	record := nametypes.NewNameRecord(name, addr, restrict)
	if err = record.ValidateBasic(); err != nil {
		return err
	}
	bz, err := types.ModuleCdc.MarshalBinaryBare(&record)
	if err != nil {
		return err
	}
	store.Set(key, bz)
	// Now index by address
	addrPrefix, err := GetAddressKeyPrefixAmino(addr)
	if err != nil {
		return err
	}
	indexKey := append(addrPrefix, key...) // [0x04] :: [addr-bytes] :: [name-key-bytes]
	store.Set(indexKey, bz)
	return nil
}

// GetAddressKeyPrefix returns a store key for a name record address
// only for testing
func GetAddressKeyPrefixAmino(address sdk.AccAddress) (key []byte, err error) {
	err = sdk.VerifyAddressFormat(address.Bytes())
	if err == nil {
		key = nametypes.AddressKeyPrefixAmino
		key = append(key, address.Bytes()...)
	}
	return
}
