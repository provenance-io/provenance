package keeper_test

import (
	"bytes"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/google/uuid"
	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/metadata/keeper"
	"github.com/provenance-io/provenance/x/metadata/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"math/rand"
	"testing"
)

type MigrationsTestSuite struct {
	suite.Suite

	app   *simapp.App
	ctx   sdk.Context
	store sdk.KVStore

	prefixes []namedIndex
}

type namedIndex struct {
	name string
	key  []byte
}

func (s *MigrationsTestSuite) SetupTest() {
	s.app = simapp.Setup(false)
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{})
	s.store = s.ctx.KVStore(s.app.GetKey(types.ModuleName))

	s.prefixes = []namedIndex{
		{"Address Scope", types.AddressScopeCacheKeyPrefix},
		{"Scope Spec Scope", types.ScopeSpecScopeCacheKeyPrefix},
		{"Value Owner Scope", types.ValueOwnerScopeCacheKeyPrefix},
		{"Address Scope Spec", types.AddressScopeSpecCacheKeyPrefix},
		{"Contract Spec Scope Spec", types.ContractSpecScopeSpecCacheKeyPrefix},
		{"Address Contract Spec", types.AddressContractSpecCacheKeyPrefix},
	}
}

func TestMigrationsTestSuite(t *testing.T) {
	suite.Run(t, new(MigrationsTestSuite))
}

func randomByteSlice(length int) []byte {
	rv := make([]byte, length)
	for i := 0; i < length; i++ {
		rv[i] = byte(rand.Int() & 0xFF)
	}
	return rv
}

// assertDeleteAllIndexes clears all the metadata index entries and asserts no errors occurred.
// This is similar to migrations -> deleteIndexes but that is private (and should stay private),
// and this is a little nicer for these unit test. I want to use this tests list of prefixes too.
// If t is nil, s.T() is used when a *testing.T is needed.
// Returns true if all assertions passed, false otherwise.
func (s *MigrationsTestSuite) assertDeleteAllIndexes(t *testing.T) bool {
	if t == nil {
		t = s.T()
	}
	allPassed := true
	for _, pre := range s.prefixes {
		// Using assert here (instead of require) so it's easier to later decide if a test should continue.
		if !assert.NoError(t, clearStore(prefix.NewStore(s.store, pre.key)), "clearing %s indexes", pre.name) {
			allPassed = false
		}
	}
	return allPassed
}

// requireDeleteAllIndexes is the same as assertDeleteAllIndexes but ends the test if an error occurs.
// If t is nil, s.T() is used when a *testing.T is needed.
func (s *MigrationsTestSuite) requireDeleteAllIndexes(t *testing.T) {
	if !(s.assertDeleteAllIndexes(t)) {
		if t != nil {
			t.FailNow()
		} else {
			s.T().FailNow()
		}
	}
}

// clearStore deletes all the entries in a store.
// This is a copy of migrations -> clearStore but that is private (and should stay private).
func clearStore(store sdk.KVStore) (err error) {
	iter := store.Iterator(nil, nil)
	defer func() {
		err = iter.Close()
	}()
	for ; iter.Valid(); iter.Next() {
		store.Delete(iter.Key())
	}
	return nil
}

func (s *MigrationsTestSuite) assertExactIndexList(t *testing.T, indexes []namedIndex) bool {
	if t == nil {
		t = s.T()
	}
	// Make sure all the provided entries exist.
	allFound := true
	for _, index := range indexes {
		if !assert.True(t, s.store.Has(index.key), "entry exists for %s", index.name) {
			allFound = false
		}
	}
	// Make sure there aren't any entries other than those.
	allKnown := true
	for _, pre := range s.prefixes {
		unknown := s.findUnknownWithPrefix(pre.key, indexes)
		if !assert.Len(t, unknown, 0, "unknown entries for %s", pre.name) {
			allKnown = false
		}
	}
	return allFound && allKnown
}

func (s *MigrationsTestSuite) findUnknownWithPrefix(pre []byte, indexes []namedIndex) [][]byte {
	rv := [][]byte{}
	pStore := prefix.NewStore(s.store, pre)
	iter := pStore.Iterator(nil, nil)
	for ; iter.Valid(); iter.Next() {
		fullKey := make([]byte, len(pre)+len(iter.Key()))
		copy(fullKey, pre)
		copy(fullKey[len(pre):], iter.Key())
		if !isKnownIndex(fullKey, indexes) {
			rv = append(rv, fullKey)
		}
	}
	return rv
}

func isKnownIndex(key []byte, indexes []namedIndex) bool {
	for _, index := range indexes {
		if bytes.Equal(key, index.key) {
			return true
		}
	}
	return false
}

// TODO: Delete this and use the one defined in scope_test.go during merge.
type user struct {
	PrivKey cryptotypes.PrivKey
	PubKey  cryptotypes.PubKey
	Addr    sdk.AccAddress
	Bech32  string
}

// TODO: Delete this and use the one defined in scope_test.go during merge.
func randomUser() user {
	rv := user{}
	rv.PrivKey = secp256k1.GenPrivKey()
	rv.PubKey = rv.PrivKey.PubKey()
	rv.Addr = sdk.AccAddress(rv.PubKey.Address())
	rv.Bech32 = rv.Addr.String()
	return rv
}

func (s *MigrationsTestSuite) Test2To3() {
	s.T().Run("existing indexes are deleted", func(t *testing.T) {
		// Address w/length 20 and length prefix + 17 random bytes
		makeIndexAddr20 := func(pre []byte) []byte {
			rv := make([]byte, 0, 39)
			rv = append(rv, pre...)
			rv = append(rv, byte(20))
			rv = append(rv, randomByteSlice(20)...)
			rv = append(rv, randomByteSlice(17)...)
			return rv
		}
		// Address w/length 32 and length prefix + 17 random bytes.
		makeIndexAddr32 := func(pre []byte) []byte {
			rv := make([]byte, 0, 51)
			rv = append(rv, pre...)
			rv = append(rv, byte(32))
			rv = append(rv, randomByteSlice(32)...)
			rv = append(rv, randomByteSlice(17)...)
			return rv
		}
		// Prefix + 37 random bytes.
		makeIndexLen38 := func(pre []byte) []byte {
			rv := make([]byte, 0, len(pre)+37)
			rv = append(rv, pre...)
			rv = append(rv, randomByteSlice(37)...)
			return rv
		}
		// Prefix + 49 random bytes.
		makeIndexLen49 := func(pre []byte) []byte {
			rv := make([]byte, 0, len(pre)+49)
			rv = append(rv, pre...)
			rv = append(rv, randomByteSlice(49)...)
			return rv
		}
		// Prefix + zeros until total length is 60.
		makeIndexZeros := func(pre []byte) []byte {
			rv := make([]byte, 60)
			copy(rv, pre)
			return rv
		}

		makeIndexes := func(name string, pre []byte) []namedIndex {
			rv := make([]namedIndex, 5)
			rv[0] = namedIndex{name + " 20", makeIndexAddr20(pre)}
			rv[1] = namedIndex{name + " 32", makeIndexAddr32(pre)}
			rv[2] = namedIndex{name + " 38", makeIndexLen38(pre)}
			rv[3] = namedIndex{name + " 49", makeIndexLen49(pre)}
			rv[4] = namedIndex{name + " zeros", makeIndexZeros(pre)}
			return rv
		}
		indexes := []namedIndex{}
		for _, pre := range s.prefixes {
			indexes = append(indexes, makeIndexes(pre.name, pre.key)...)
		}

		// Add them to the store.
		for _, index := range indexes {
			s.store.Set(index.key, []byte{0x01})
		}

		// Do the migration.
		migrator := keeper.NewMigrator(s.app.MetadataKeeper)
		require.NoError(t, migrator.Migrate2to3(s.ctx), "running migration")

		// Ensure none of those indexes exist anymore.
		for _, index := range indexes {
			assert.False(t, s.store.Has(index.key), index.name)
		}
	})

	s.T().Run("scopes are reindexed", func(t *testing.T) {
		owner1 := randomUser()
		owner2 := randomUser()
		ownerCommon := randomUser()
		valueOwner1 := randomUser()
		valueOwner2 := randomUser()
		scope1 := types.Scope{
			ScopeId:           types.ScopeMetadataAddress(uuid.New()),
			SpecificationId:   types.ScopeSpecMetadataAddress(uuid.New()),
			Owners:            ownerPartyList(owner1.Bech32, ownerCommon.Bech32),
			DataAccess:        nil,
			ValueOwnerAddress: valueOwner1.Bech32,
		}
		scope2 := types.Scope{
			ScopeId:           types.ScopeMetadataAddress(uuid.New()),
			SpecificationId:   types.ScopeSpecMetadataAddress(uuid.New()),
			Owners:            ownerPartyList(owner2.Bech32, ownerCommon.Bech32),
			DataAccess:        nil,
			ValueOwnerAddress: valueOwner2.Bech32,
		}

		indexes := []namedIndex{
			{"scope1 owner1 address", types.GetAddressScopeCacheKey(owner1.Addr, scope1.ScopeId)},
			{"scope1 ownerCommon address", types.GetAddressScopeCacheKey(ownerCommon.Addr, scope1.ScopeId)},
			{"scope1 valueOwner1 address", types.GetAddressScopeCacheKey(valueOwner1.Addr, scope1.ScopeId)},
			{"scope1 value owner", types.GetValueOwnerScopeCacheKey(valueOwner1.Addr, scope1.ScopeId)},
			{"scope1 scope spec", types.GetScopeSpecScopeCacheKey(scope1.SpecificationId, scope1.ScopeId)},

			{"scope2 owner2 address", types.GetAddressScopeCacheKey(owner2.Addr, scope2.ScopeId)},
			{"scope2 ownerCommon address", types.GetAddressScopeCacheKey(ownerCommon.Addr, scope2.ScopeId)},
			{"scope2 valueOwner2 address", types.GetAddressScopeCacheKey(valueOwner2.Addr, scope2.ScopeId)},
			{"scope2 value owner", types.GetValueOwnerScopeCacheKey(valueOwner2.Addr, scope2.ScopeId)},
			{"scope2 scope spec", types.GetScopeSpecScopeCacheKey(scope2.SpecificationId, scope2.ScopeId)},
		}

		// Set the scopes.
		s.app.MetadataKeeper.SetScope(s.ctx, scope1)
		s.app.MetadataKeeper.SetScope(s.ctx, scope2)
		defer func() {
			s.app.MetadataKeeper.RemoveScope(s.ctx, scope1.ScopeId)
			s.app.MetadataKeeper.RemoveScope(s.ctx, scope2.ScopeId)
		}()

		// Delete any indexes added for them.
		s.requireDeleteAllIndexes(t)
		// Run the migration
		migrator := keeper.NewMigrator(s.app.MetadataKeeper)
		require.NoError(t, migrator.Migrate2to3(s.ctx), "running migration")

		// Make sure the indexes are as expected.
		s.assertExactIndexList(t, indexes)
	})

	s.T().Run("scope specs are reindexed", func(t *testing.T) {
		owner1 := randomUser()
		owner2 := randomUser()
		ownerCommon := randomUser()
		cSpec1ID := types.ContractSpecMetadataAddress(uuid.New())
		cSpec2ID := types.ContractSpecMetadataAddress(uuid.New())
		cSpecCommonID := types.ContractSpecMetadataAddress(uuid.New())
		scopeSpec1 := types.ScopeSpecification{
			SpecificationId: types.ScopeSpecMetadataAddress(uuid.New()),
			Description:     nil,
			OwnerAddresses:  []string{owner1.Bech32, ownerCommon.Bech32},
			PartiesInvolved: nil,
			ContractSpecIds: []types.MetadataAddress{cSpec1ID, cSpecCommonID},
		}
		scopeSpec2 := types.ScopeSpecification{
			SpecificationId: types.ScopeSpecMetadataAddress(uuid.New()),
			Description:     nil,
			OwnerAddresses:  []string{owner2.Bech32, ownerCommon.Bech32},
			PartiesInvolved: nil,
			ContractSpecIds: []types.MetadataAddress{cSpec2ID, cSpecCommonID},
		}

		indexes := []namedIndex{
			{"scopeSpec1 owner1 address", types.GetAddressScopeSpecCacheKey(owner1.Addr, scopeSpec1.SpecificationId)},
			{"scopeSpec1 ownerCommon address", types.GetAddressScopeSpecCacheKey(ownerCommon.Addr, scopeSpec1.SpecificationId)},
			{"scopeSpec1 cSpec1 address", types.GetContractSpecScopeSpecCacheKey(cSpec1ID, scopeSpec1.SpecificationId)},
			{"scopeSpec1 cSpecCommon address", types.GetContractSpecScopeSpecCacheKey(cSpecCommonID, scopeSpec1.SpecificationId)},

			{"scopeSpec2 owner2 address", types.GetAddressScopeSpecCacheKey(owner2.Addr, scopeSpec2.SpecificationId)},
			{"scopeSpec2 ownerCommon address", types.GetAddressScopeSpecCacheKey(ownerCommon.Addr, scopeSpec2.SpecificationId)},
			{"scopeSpec2 cSpec2 address", types.GetContractSpecScopeSpecCacheKey(cSpec2ID, scopeSpec2.SpecificationId)},
			{"scopeSpec2 cSpecCommon address", types.GetContractSpecScopeSpecCacheKey(cSpecCommonID, scopeSpec2.SpecificationId)},
		}

		// Set the scopes specs.
		s.app.MetadataKeeper.SetScopeSpecification(s.ctx, scopeSpec1)
		s.app.MetadataKeeper.SetScopeSpecification(s.ctx, scopeSpec2)
		defer func() {
			assert.NoError(t, s.app.MetadataKeeper.RemoveScopeSpecification(s.ctx, scopeSpec1.SpecificationId), "removing scopeSpec1")
			assert.NoError(t, s.app.MetadataKeeper.RemoveScopeSpecification(s.ctx, scopeSpec2.SpecificationId), "removing scopeSpec2")
		}()

		// Delete any indexes added for them.
		s.requireDeleteAllIndexes(t)
		// Run the migration
		migrator := keeper.NewMigrator(s.app.MetadataKeeper)
		require.NoError(t, migrator.Migrate2to3(s.ctx), "running migration")

		// Make sure the indexes are as expected.
		s.assertExactIndexList(t, indexes)
	})

	s.T().Run("contract specs are reindexed", func(t *testing.T) {
		owner1 := randomUser()
		owner2 := randomUser()
		ownerCommon := randomUser()
		cSpec1 := types.ContractSpecification{
			SpecificationId: types.ContractSpecMetadataAddress(uuid.New()),
			Description:     nil,
			OwnerAddresses:  []string{owner1.Bech32, ownerCommon.Bech32},
			PartiesInvolved: nil,
			Source:          nil,
			ClassName:       "",
		}
		cSpec2 := types.ContractSpecification{
			SpecificationId: types.ContractSpecMetadataAddress(uuid.New()),
			Description:     nil,
			OwnerAddresses:  []string{owner2.Bech32, ownerCommon.Bech32},
			PartiesInvolved: nil,
			Source:          nil,
			ClassName:       "",
		}

		indexes := []namedIndex{
			{"cSpec1 owner1 address", types.GetAddressContractSpecCacheKey(owner1.Addr, cSpec1.SpecificationId)},
			{"cSpec1 ownerCommon address", types.GetAddressContractSpecCacheKey(ownerCommon.Addr, cSpec1.SpecificationId)},

			{"cSpec2 owner2 address", types.GetAddressContractSpecCacheKey(owner2.Addr, cSpec2.SpecificationId)},
			{"cSpec2 ownerCommon address", types.GetAddressContractSpecCacheKey(ownerCommon.Addr, cSpec2.SpecificationId)},
		}

		// Set the contract specs.
		s.app.MetadataKeeper.SetContractSpecification(s.ctx, cSpec1)
		s.app.MetadataKeeper.SetContractSpecification(s.ctx, cSpec2)
		defer func() {
			assert.NoError(t, s.app.MetadataKeeper.RemoveContractSpecification(s.ctx, cSpec1.SpecificationId), "removing cSpec1")
			assert.NoError(t, s.app.MetadataKeeper.RemoveContractSpecification(s.ctx, cSpec2.SpecificationId), "removing cSpec2")
		}()

		// Delete any indexes added for them.
		s.requireDeleteAllIndexes(t)
		// Run the migration
		migrator := keeper.NewMigrator(s.app.MetadataKeeper)
		require.NoError(t, migrator.Migrate2to3(s.ctx), "running migration")

		// Make sure the indexes are as expected.
		s.assertExactIndexList(t, indexes)
	})
}
