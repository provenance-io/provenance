package keeper_test

import (
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
	nameKey, err := types.GetNameKeyPrefix(name)
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
	// Create new keeper with collections schema
	newKeeper := keeper.NewKeeper(
		s.cdc,
		storeKey,
		runtime.NewKVStoreService(storeKey),
	)

	migrator := keeper.NewMigrator(newKeeper)

	// Run migration
	err := migrator.MigrateKVToCollections2to3(s.ctx)
	s.Require().NoError(err)

	params := newKeeper.GetParams(s.ctx)
	s.Require().Equal(types.DefaultParams(), params)

	name := "test.provenance"
	nameKey, err := types.GetNameKeyPrefix(name)
	s.Require().NoError(err)

	record, err := newKeeper.GetNameRecord(s.ctx, nameKey)
	s.Require().NoError(err)
	s.Require().Equal(name, record.Name)

	recordPtr, err := newKeeper.GetRecordByName(s.ctx, name)
	s.Require().NoError(err)
	s.Require().Equal(name, recordPtr.Name)

	addr, err := sdk.AccAddressFromBech32(s.user1)
	s.Require().NoError(err)

	addrPrefix, err := types.GetAddressKeyPrefix(addr)
	s.Require().NoError(err)

	addrIndexKey := append(addrPrefix, nameKey...)
	indexRecord, err := newKeeper.GetAddrIndexRecord(s.ctx, addrIndexKey)
	s.Require().NoError(err)
	s.Require().Equal(*recordPtr, indexRecord)

	records, err := newKeeper.GetRecordsByAddress(s.ctx, addr)
	s.Require().NoError(err)
	s.Require().Len(records, 1)
	s.Require().Equal(name, records[0].Name)

	// Reverse lookup via query
	res, err := newKeeper.ReverseLookup(s.ctx, &types.QueryReverseLookupRequest{
		Address: s.user1,
	})
	s.Require().NoError(err)
	s.Require().Len(res.Name, 1)
	s.Require().Equal(name, res.Name[0])
}
