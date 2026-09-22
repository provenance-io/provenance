package keeper_test

import (
	"testing"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/provenance-io/provenance/x/name/keeper"
	"github.com/provenance-io/provenance/x/name/types"
	"github.com/stretchr/testify/suite"
)

type MigrationTestSuite struct {
	suite.Suite

	ctx       sdk.Context
	storeKey  *storetypes.KVStoreKey
	cdc       codec.BinaryCodec
	user1Addr sdk.AccAddress
}

// NEW (#12): Without this, none of the tests in this suite run.
func TestMigrationTestSuite(t *testing.T) {
	suite.Run(t, new(MigrationTestSuite))
}

func (s *MigrationTestSuite) SetupTest() {
	s.storeKey = storetypes.NewKVStoreKey(types.StoreKey)
	s.ctx = testutil.DefaultContext(s.storeKey, storetypes.NewTransientStoreKey("transient_test"))
	s.cdc = codec.NewProtoCodec(codectypes.NewInterfaceRegistry())
	s.user1Addr = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
}

// seedLegacyRecord writes a name record and its address index entry using the pre-collections (v2) layout.
func (s *MigrationTestSuite) seedLegacyRecord(record types.NameRecord) {
	store := s.ctx.KVStore(s.storeKey)
	bz := s.cdc.MustMarshal(&record)

	nameKey, err := keeper.LegacyGetNameKeyBytes(record.Name)
	s.Require().NoError(err, "LegacyGetNameKeyBytes(%q)", record.Name)
	store.Set(nameKey, bz)

	addr, err := sdk.AccAddressFromBech32(record.Address)
	s.Require().NoError(err, "AccAddressFromBech32(%q)", record.Address)
	addrKey := append(append([]byte{}, keeper.LegacyAddressKeyPrefix...), address.MustLengthPrefix(addr)...)
	store.Set(addrKey, bz)
}

// seedLegacyParams writes the params using the pre-collections (v2) layout.
func (s *MigrationTestSuite) seedLegacyParams(params types.Params) {
	s.ctx.KVStore(s.storeKey).Set(keeper.LegacyNameParamStoreKey, s.cdc.MustMarshal(&params))
}

// migrate creates a keeper on the test store and runs the 2->3 migration.
func (s *MigrationTestSuite) migrate() keeper.Keeper {
	k := keeper.NewKeeper(s.cdc, runtime.NewKVStoreService(s.storeKey))
	s.Require().NoError(keeper.NewMigrator(k).MigrateKVToCollections2to3(s.ctx), "MigrateKVToCollections2to3")
	return k
}

// requireNoLegacyData asserts that nothing is left in the pre-collections (v2) layout.
func (s *MigrationTestSuite) requireNoLegacyData() {
	store := s.ctx.KVStore(s.storeKey)
	for _, prefix := range [][]byte{keeper.LegacyNameKeyPrefix, keeper.LegacyAddressKeyPrefix} {
		iter := storetypes.KVStorePrefixIterator(store, prefix)
		hasData := iter.Valid()
		s.Require().NoError(iter.Close(), "closing the iterator for prefix %X", prefix)
		s.Require().False(hasData, "legacy prefix %X still has entries", prefix)
	}
	s.Require().False(store.Has(keeper.LegacyNameParamStoreKey), "the legacy params entry still exists")
}

func (s *MigrationTestSuite) TestMigration() {
	records := []types.NameRecord{
		types.NewNameRecord("provenance", s.user1Addr, true),
		types.NewNameRecord("test.provenance", s.user1Addr, false),
	}
	for _, record := range records {
		s.seedLegacyRecord(record)
	}
	legacyParams := types.DefaultParams()
	s.seedLegacyParams(legacyParams)

	k := s.migrate()

	s.Require().Equal(legacyParams, k.GetParams(s.ctx), "params after migration")
	for _, exp := range records {
		got, err := k.GetRecordByName(s.ctx, exp.Name)
		s.Require().NoError(err, "GetRecordByName(%q)", exp.Name)
		s.Require().Equal(exp, *got, "record %q after migration", exp.Name)
	}

	byAddr, err := k.GetRecordsByAddress(s.ctx, s.user1Addr)
	s.Require().NoError(err, "GetRecordsByAddress")
	s.Require().ElementsMatch(records, byAddr, "records by address after migration")

	s.requireNoLegacyData()
}

// Covers review comment #13: starting from DefaultParams would turn a stored false into true.
func (s *MigrationTestSuite) TestMigrationKeepsAllowUnrestrictedNamesFalse() {
	legacyParams := types.DefaultParams()
	legacyParams.AllowUnrestrictedNames = false
	s.seedLegacyParams(legacyParams)

	k := s.migrate()

	s.Require().Equal(legacyParams, k.GetParams(s.ctx), "params after migration")
	s.requireNoLegacyData()
}

func (s *MigrationTestSuite) TestMigrationEmptyStore() {
	k := s.migrate()

	s.Require().Equal(types.DefaultParams(), k.GetParams(s.ctx), "params after migrating an empty store")
	s.requireNoLegacyData()
}
