package keeper_test

import (
	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/provenance-io/provenance/x/name/keeper"
	"github.com/provenance-io/provenance/x/name/types"
	"github.com/stretchr/testify/suite"
)

type MigrationTestSuite struct {
	suite.Suite

	ctx       sdk.Context
	store     storetypes.StoreKey
	cdc       codec.BinaryCodec
	pubkey1   cryptotypes.PubKey
	user1     string
	user1Addr sdk.AccAddress
}

func (s *MigrationTestSuite) SetupTest() {
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	tKey := storetypes.NewTransientStoreKey("transient_test")
	s.store = storeKey
	s.ctx = testutil.DefaultContext(storeKey, tKey)

	// Initialize pubkey and address
	privKey := secp256k1.GenPrivKey()
	s.pubkey1 = privKey.PubKey()
	s.user1Addr = sdk.AccAddress(s.pubkey1.Address())
	s.user1 = s.user1Addr.String()

	interfaceRegistry := codectypes.NewInterfaceRegistry()
	s.cdc = codec.NewProtoCodec(interfaceRegistry)

	store := s.ctx.KVStore(storeKey)

	params := types.DefaultParams()
	store.Set(types.NameParamStoreKey, s.cdc.MustMarshal(&params))

	name := "test.provenance"
	nameKey, err := types.GetNameKeyBytes(name)
	s.Require().NoError(err)

	record := types.NameRecord{
		Name:       name,
		Address:    s.user1,
		Restricted: true,
	}
	store.Set(nameKey, s.cdc.MustMarshal(&record))

	addrPrefix, err := types.GetAddressKeyPrefix(s.user1Addr)
	s.Require().NoError(err)

	addrIndexKey := append(addrPrefix, nameKey...)
	store.Set(addrIndexKey, s.cdc.MustMarshal(&record))
}

func (s *MigrationTestSuite) TestMigration() {
	storeKey := s.store.(*storetypes.KVStoreKey)

	s.user1Addr = sdk.AccAddress(s.pubkey1.Address())
	s.user1 = s.user1Addr.String()
	name := "test.provenance"
	record := types.NewNameRecord(name, s.user1Addr, true)

	oldStore := s.ctx.KVStore(storeKey)
	nameKey, err := types.GetNameKeyBytes(name)
	s.Require().NoError(err, "failed to get name key bytes")

	recordBz, err := s.cdc.Marshal(&record)
	s.Require().NoError(err, "failed to marshal name record")
	oldStore.Set(nameKey, recordBz)

	addrIndexKey, err := types.GetAddressKeyBytes(s.user1Addr, nameKey[1:]) // Strip prefix byte
	s.Require().NoError(err, "failed to get address key bytes")
	oldStore.Set(addrIndexKey, []byte("old"))

	newKeeper := keeper.NewKeeper(s.cdc, runtime.NewKVStoreService(storeKey))
	migrator := keeper.NewMigrator(newKeeper)

	err = migrator.MigrateKVToCollections2to3(s.ctx)
	s.Require().NoError(err, "migration failed")

	params := newKeeper.GetParams(s.ctx)
	s.Require().Equal(types.DefaultParams(), params, "params mismatch after migration")

	migratedRecord, err := newKeeper.GetNameRecord(s.ctx, name)
	s.Require().NoError(err, "failed to get migrated name record")
	s.Require().Equal(name, migratedRecord.Name, "migrated record name mismatch")
	s.Require().Equal(s.user1Addr.String(), migratedRecord.Address, "migrated record address mismatch")

	normalized, err := newKeeper.Normalize(s.ctx, name)
	s.Require().NoError(err, "failed to normalize name")

	pair := collections.Join([]byte(s.user1Addr), normalized)
	iter, err := newKeeper.GetAddrIndex().MatchExact(s.ctx, pair)
	s.Require().NoError(err, "failed to get address index iterator")
	defer iter.Close()

	s.Require().True(iter.Valid(), "address index iterator should be valid")

	primaryKey, err := iter.PrimaryKey()
	s.Require().NoError(err, "failed to get primary key from iterator")

	indexRecord, err := newKeeper.GetNameRecord(s.ctx, primaryKey)
	s.Require().NoError(err, "failed to get record by primary key")
	s.Require().Equal(migratedRecord, indexRecord, "migrated record and index record do not match")

	res, err := newKeeper.ReverseLookup(s.ctx, &types.QueryReverseLookupRequest{Address: s.user1})
	s.Require().NoError(err, "reverse lookup query failed")
	s.Require().Len(res.Name, 1, "unexpected number of names in reverse lookup")
	s.Require().Equal(name, res.Name[0], "reverse lookup returned wrong name")
}
