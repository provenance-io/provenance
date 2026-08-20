package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/provenance-io/provenance/app"
	namekeeper "github.com/provenance-io/provenance/x/name/keeper"
	nametypes "github.com/provenance-io/provenance/x/name/types"
)

type MigrationsTestSuite struct {
	suite.Suite

	app *app.App
	ctx sdk.Context
	cdc *codec.ProtoCodec

	user1Addr sdk.AccAddress
	user2Addr sdk.AccAddress
}

func TestMigrationsTestSuite(t *testing.T) {
	suite.Run(t, new(MigrationsTestSuite))
}

func (s *MigrationsTestSuite) SetupTest() {
	s.app = app.Setup(s.T())
	s.cdc = codec.NewProtoCodec(s.app.GetEncodingConfig().InterfaceRegistry)
	s.ctx = s.app.BaseApp.NewContext(false)
	s.user1Addr = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	s.user2Addr = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
}

func (s *MigrationsTestSuite) store() storetypes.KVStore {
	return s.ctx.KVStore(s.app.GetKey(nametypes.StoreKey))
}

// setLegacyNameRecord writes a name record and its address index entry using the keys
// that the name module used before version 3.
func (s *MigrationsTestSuite) setLegacyNameRecord(name string, addr sdk.AccAddress, restricted bool) {
	record := nametypes.NewNameRecord(name, addr, restricted)
	bz, err := s.cdc.Marshal(&record)
	s.Require().NoError(err, "Marshal(%q)", name)

	nameKey, err := nametypes.GetLegacyNameKeyPrefix(name)
	s.Require().NoError(err, "GetLegacyNameKeyPrefix(%q)", name)
	addrKey, err := nametypes.GetAddressKeyPrefix(addr)
	s.Require().NoError(err, "GetAddressKeyPrefix(%q)", addr)

	store := s.store()
	store.Set(nameKey, bz)
	store.Set(append(addrKey, nameKey...), bz)
}

func (s *MigrationsTestSuite) TestMigrate2to3() {
	multiSegment := []string{"leaf.mid.root", "example.name", "test.root"}
	singleSegment := []string{"root", "name"}
	allNames := append(append([]string{}, multiSegment...), singleSegment...)

	for _, name := range allNames {
		s.setLegacyNameRecord(name, s.user1Addr, false)
	}
	s.setLegacyNameRecord("other.root", s.user2Addr, true)

	// The multi-segment records are only reachable by their legacy key until they're migrated.
	for _, name := range multiSegment {
		_, err := s.app.NameKeeper.GetRecordByName(s.ctx, name)
		s.Assert().ErrorIs(err, nametypes.ErrNameNotBound, "GetRecordByName(%q) before the migration", name)
	}

	err := namekeeper.NewMigrator(s.app.NameKeeper).Migrate2to3(s.ctx)
	s.Require().NoError(err, "Migrate2to3")

	store := s.store()
	for _, name := range allNames {
		record, err := s.app.NameKeeper.GetRecordByName(s.ctx, name)
		if s.Assert().NoError(err, "GetRecordByName(%q) after the migration", name) {
			s.Assert().Equal(name, record.Name, "record name")
			s.Assert().Equal(s.user1Addr.String(), record.Address, "record address")
		}

		addrKey, err := nametypes.GetAddressKeyPrefix(s.user1Addr)
		s.Require().NoError(err, "GetAddressKeyPrefix")
		nameKey, err := nametypes.GetNameKeyPrefix(name)
		s.Require().NoError(err, "GetNameKeyPrefix(%q)", name)
		s.Assert().True(store.Has(append(addrKey, nameKey...)), "address index entry exists for %q", name)
	}

	for _, name := range multiSegment {
		legacyKey, err := nametypes.GetLegacyNameKeyPrefix(name)
		s.Require().NoError(err, "GetLegacyNameKeyPrefix(%q)", name)
		s.Assert().False(store.Has(legacyKey), "legacy record still exists for %q", name)

		addrKey, err := nametypes.GetAddressKeyPrefix(s.user1Addr)
		s.Require().NoError(err, "GetAddressKeyPrefix")
		s.Assert().False(store.Has(append(addrKey, legacyKey...)), "legacy address index entry still exists for %q", name)
	}

	records, err := s.app.NameKeeper.GetRecordsByAddress(s.ctx, s.user1Addr)
	s.Require().NoError(err, "GetRecordsByAddress")
	names := make([]string, 0, len(records))
	for _, record := range records {
		names = append(names, record.Name)
	}
	s.Assert().ElementsMatch(allNames, names, "names bound to the address")

	record, err := s.app.NameKeeper.GetRecordByName(s.ctx, "other.root")
	s.Require().NoError(err, `GetRecordByName("other.root")`)
	s.Assert().Equal(s.user2Addr.String(), record.Address, "record address")
	s.Assert().True(record.Restricted, "record restricted")
}

func (s *MigrationsTestSuite) TestMigrate2to3IsRepeatable() {
	s.setLegacyNameRecord("leaf.mid.root", s.user1Addr, false)

	migrator := namekeeper.NewMigrator(s.app.NameKeeper)
	s.Require().NoError(migrator.Migrate2to3(s.ctx), "Migrate2to3 first run")
	s.Require().NoError(migrator.Migrate2to3(s.ctx), "Migrate2to3 second run")

	record, err := s.app.NameKeeper.GetRecordByName(s.ctx, "leaf.mid.root")
	s.Require().NoError(err, `GetRecordByName("leaf.mid.root")`)
	s.Assert().Equal(s.user1Addr.String(), record.Address, "record address")
}
