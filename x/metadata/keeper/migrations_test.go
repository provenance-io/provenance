package keeper_test

import (
	"bytes"
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/metadata/keeper"
	v042 "github.com/provenance-io/provenance/x/metadata/legacy/v042"
	"github.com/provenance-io/provenance/x/metadata/types"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

type MigrationsTestSuite struct {
	suite.Suite

	app   *simapp.App
	ctx   sdk.Context
	store sdk.KVStore

	prefixes           namedIndexList
	unaffectedPrefixes namedIndexList
	oldPrefixes        namedIndexList
	allPrefixes        namedIndexList
}

type namedIndexList []namedIndex

func (l namedIndexList) String() string {
	var rv strings.Builder
	for _, n := range l {
		rv.WriteString(n.String())
		rv.WriteByte('\n')
	}
	return rv.String()
}

type namedIndex struct {
	name string
	key  []byte
}

func (n namedIndex) String() string {
	return fmt.Sprintf("%s: %X", n.name, n.key)
}

func (s *MigrationsTestSuite) SetupTest() {
	s.app = simapp.Setup(false)
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{})
	s.store = s.ctx.KVStore(s.app.GetKey(types.ModuleName))

	s.prefixes = namedIndexList{
		{"Address Scope", types.AddressScopeCacheKeyPrefix},
		{"Value Owner Scope", types.ValueOwnerScopeCacheKeyPrefix},
		{"Address Scope Spec", types.AddressScopeSpecCacheKeyPrefix},
		{"Address Contract Spec", types.AddressContractSpecCacheKeyPrefix},
	}

	s.unaffectedPrefixes = namedIndexList{
		{"Scope Spec Scope", types.ScopeSpecScopeCacheKeyPrefix},
		{"Contract Spec Scope Spec", types.ContractSpecScopeSpecCacheKeyPrefix},
	}

	s.oldPrefixes = namedIndexList{
		{"Legacy Address Scope", v042.AddressScopeCacheKeyPrefixLegacy},
		{"Legacy Value Owner Scope", v042.ValueOwnerScopeCacheKeyPrefixLegacy},
		{"Legacy Address ScopeSpec", v042.AddressScopeSpecCacheKeyPrefixLegacy},
		{"Legacy Address Contract Spec", v042.AddressContractSpecCacheKeyPrefixLegacy},
	}

	s.allPrefixes = namedIndexList{}
	s.allPrefixes = append(s.allPrefixes, s.prefixes...)
	s.allPrefixes = append(s.allPrefixes, s.unaffectedPrefixes...)
	s.allPrefixes = append(s.allPrefixes, s.oldPrefixes...)
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

func (s *MigrationsTestSuite) assertExactIndexList(t *testing.T, indexes namedIndexList) bool {
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
	for _, pre := range s.allPrefixes {
		unknown := s.findUnknownWithPrefix(pre.key, indexes)
		if !assert.Len(t, unknown, 0, "unknown entries for %s", pre.name) {
			allKnown = false
		}
	}
	return allFound && allKnown
}

func (s *MigrationsTestSuite) findUnknownWithPrefix(pre []byte, indexes namedIndexList) []string {
	rv := []string{}
	pStore := prefix.NewStore(s.store, pre)
	iter := pStore.Iterator(nil, nil)
	for ; iter.Valid(); iter.Next() {
		fullKey := make([]byte, len(pre)+len(iter.Key()))
		copy(fullKey, pre)
		copy(fullKey[len(pre):], iter.Key())
		if !isKnownIndex(fullKey, indexes) {
			rv = append(rv, fmt.Sprintf("%X", fullKey))
		}
	}
	return rv
}

func isKnownIndex(key []byte, indexes namedIndexList) bool {
	for _, index := range indexes {
		if bytes.Equal(key, index.key) {
			return true
		}
	}
	return false
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

		makeIndexes := func(name string, pre []byte) namedIndexList {
			rv := make(namedIndexList, 5)
			rv[0] = namedIndex{name + " 20", makeIndexAddr20(pre)}
			rv[1] = namedIndex{name + " 32", makeIndexAddr32(pre)}
			rv[2] = namedIndex{name + " 38", makeIndexLen38(pre)}
			rv[3] = namedIndex{name + " 49", makeIndexLen49(pre)}
			rv[4] = namedIndex{name + " zeros", makeIndexZeros(pre)}
			return rv
		}
		indexes := namedIndexList{}
		for _, pre := range s.prefixes {
			indexes = append(indexes, makeIndexes(pre.name, pre.key)...)
		}
		keptIndexes := namedIndexList{}
		for _, pre := range s.unaffectedPrefixes {
			keptIndexes = append(keptIndexes, makeIndexes(pre.name, pre.key)...)
		}

		// Add them to the store.
		for _, index := range indexes {
			s.store.Set(index.key, []byte{0x01})
		}
		for _, index := range keptIndexes {
			s.store.Set(index.key, []byte{0x01})
		}
		defer func() {
			for _, index := range indexes {
				s.store.Delete(index.key)
			}
			for _, index := range keptIndexes {
				s.store.Delete(index.key)
			}
		}()

		// Do the migration.
		migrator := keeper.NewMigrator(s.app.MetadataKeeper)
		require.NoError(t, migrator.Migrate2to3(s.ctx), "running migration")

		// Ensure none of those indexes exist anymore.
		for _, index := range indexes {
			assert.False(t, s.store.Has(index.key), index.name)
		}
		// Ensure these indexes still exist.
		for _, index := range keptIndexes {
			assert.True(t, s.store.Has(index.key), index.name)
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

		indexes := namedIndexList{
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

		indexes := namedIndexList{
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

		indexes := namedIndexList{
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
	
	s.T().Run("good bad new full run all fixed", func(t *testing.T) {
		// This test mimics a state where there are three metadata entries of each type affected by the v1 to v2 migration.
		// 1) A "good" entry that was migrated, and has been written since the v1 to v2 migration (so it is correctly indexed).
		// 2) A "bad" entry that was migrated, but has NOT been written since the v1 to v2 migration.
		// 3) A "new" entry that was written after the v1 to v2 migration (also correctly indexed).

		ownerGood := randomUser()
		ownerBad := randomUser()
		ownerNew := randomUser()
		ownerCommon := randomUser()
		valueOwnerGood := randomUser()
		valueOwnerBad := randomUser()
		valueOwnerNew := randomUser()
		scopeGood := types.Scope{
			ScopeId:           types.ScopeMetadataAddress(uuid.New()),
			SpecificationId:   types.ScopeSpecMetadataAddress(uuid.New()),
			Owners:            ownerPartyList(ownerGood.Bech32, ownerCommon.Bech32),
			DataAccess:        nil,
			ValueOwnerAddress: valueOwnerGood.Bech32,
		}
		scopeBad := types.Scope{
			ScopeId:           types.ScopeMetadataAddress(uuid.New()),
			SpecificationId:   types.ScopeSpecMetadataAddress(uuid.New()),
			Owners:            ownerPartyList(ownerBad.Bech32, ownerCommon.Bech32),
			DataAccess:        nil,
			ValueOwnerAddress: valueOwnerBad.Bech32,
		}
		scopeNew := types.Scope{
			ScopeId:           types.ScopeMetadataAddress(uuid.New()),
			SpecificationId:   types.ScopeSpecMetadataAddress(uuid.New()),
			Owners:            ownerPartyList(ownerNew.Bech32, ownerCommon.Bech32),
			DataAccess:        nil,
			ValueOwnerAddress: valueOwnerNew.Bech32,
		}
		cSpecGood := types.ContractSpecification{
			SpecificationId: types.ContractSpecMetadataAddress(uuid.New()),
			Description:     nil,
			OwnerAddresses:  []string{ownerGood.Bech32, ownerCommon.Bech32},
			PartiesInvolved: nil,
			Source:          nil,
			ClassName:       "",
		}
		cSpecBad := types.ContractSpecification{
			SpecificationId: types.ContractSpecMetadataAddress(uuid.New()),
			Description:     nil,
			OwnerAddresses:  []string{ownerBad.Bech32, ownerCommon.Bech32},
			PartiesInvolved: nil,
			Source:          nil,
			ClassName:       "",
		}
		cSpecNew := types.ContractSpecification{
			SpecificationId: types.ContractSpecMetadataAddress(uuid.New()),
			Description:     nil,
			OwnerAddresses:  []string{ownerNew.Bech32, ownerCommon.Bech32},
			PartiesInvolved: nil,
			Source:          nil,
			ClassName:       "",
		}
		scopeSpecGood := types.ScopeSpecification{
			SpecificationId: types.ScopeSpecMetadataAddress(uuid.New()),
			Description:     nil,
			OwnerAddresses:  []string{ownerGood.Bech32, ownerCommon.Bech32},
			PartiesInvolved: nil,
			ContractSpecIds: []types.MetadataAddress{cSpecGood.SpecificationId},
		}
		scopeSpecBad := types.ScopeSpecification{
			SpecificationId: types.ScopeSpecMetadataAddress(uuid.New()),
			Description:     nil,
			OwnerAddresses:  []string{ownerBad.Bech32, ownerCommon.Bech32},
			PartiesInvolved: nil,
			ContractSpecIds: []types.MetadataAddress{cSpecBad.SpecificationId},
		}
		scopeSpecNew := types.ScopeSpecification{
			SpecificationId: types.ScopeSpecMetadataAddress(uuid.New()),
			Description:     nil,
			OwnerAddresses:  []string{ownerNew.Bech32, ownerCommon.Bech32},
			PartiesInvolved: nil,
			ContractSpecIds: []types.MetadataAddress{cSpecNew.SpecificationId},
		}

		// Enumerate all the keys that should exist for the metadata things above.
		expectedIndexes := namedIndexList{
			{"scopeGood ownerGood address", types.GetAddressScopeCacheKey(ownerGood.Addr, scopeGood.ScopeId)},
			{"scopeGood ownerCommon address", types.GetAddressScopeCacheKey(ownerCommon.Addr, scopeGood.ScopeId)},
			{"scopeGood valueOwnerGood address", types.GetAddressScopeCacheKey(valueOwnerGood.Addr, scopeGood.ScopeId)},
			{"scopeGood value owner", types.GetValueOwnerScopeCacheKey(valueOwnerGood.Addr, scopeGood.ScopeId)},
			{"scopeGood scope spec", types.GetScopeSpecScopeCacheKey(scopeGood.SpecificationId, scopeGood.ScopeId)},

			{"scopeBad ownerBad address", types.GetAddressScopeCacheKey(ownerBad.Addr, scopeBad.ScopeId)},
			{"scopeBad ownerCommon address", types.GetAddressScopeCacheKey(ownerCommon.Addr, scopeBad.ScopeId)},
			{"scopeBad valueOwnerBad address", types.GetAddressScopeCacheKey(valueOwnerBad.Addr, scopeBad.ScopeId)},
			{"scopeBad value owner", types.GetValueOwnerScopeCacheKey(valueOwnerBad.Addr, scopeBad.ScopeId)},
			{"scopeBad scope spec", types.GetScopeSpecScopeCacheKey(scopeBad.SpecificationId, scopeBad.ScopeId)},

			{"scopeNew ownerNew address", types.GetAddressScopeCacheKey(ownerNew.Addr, scopeNew.ScopeId)},
			{"scopeNew ownerCommon address", types.GetAddressScopeCacheKey(ownerCommon.Addr, scopeNew.ScopeId)},
			{"scopeNew valueOwnerNew address", types.GetAddressScopeCacheKey(valueOwnerNew.Addr, scopeNew.ScopeId)},
			{"scopeNew value owner", types.GetValueOwnerScopeCacheKey(valueOwnerNew.Addr, scopeNew.ScopeId)},
			{"scopeNew scope spec", types.GetScopeSpecScopeCacheKey(scopeNew.SpecificationId, scopeNew.ScopeId)},

			{"cSpecGood ownerGood address", types.GetAddressContractSpecCacheKey(ownerGood.Addr, cSpecGood.SpecificationId)},
			{"cSpecGood ownerCommon address", types.GetAddressContractSpecCacheKey(ownerCommon.Addr, cSpecGood.SpecificationId)},

			{"cSpecBad ownerBad address", types.GetAddressContractSpecCacheKey(ownerBad.Addr, cSpecBad.SpecificationId)},
			{"cSpecBad ownerCommon address", types.GetAddressContractSpecCacheKey(ownerCommon.Addr, cSpecBad.SpecificationId)},

			{"cSpecNew ownerNew address", types.GetAddressContractSpecCacheKey(ownerNew.Addr, cSpecNew.SpecificationId)},
			{"cSpecNew ownerCommon address", types.GetAddressContractSpecCacheKey(ownerCommon.Addr, cSpecNew.SpecificationId)},

			{"scopeSpecGood ownerGood address", types.GetAddressScopeSpecCacheKey(ownerGood.Addr, scopeSpecGood.SpecificationId)},
			{"scopeSpecGood ownerCommon address", types.GetAddressScopeSpecCacheKey(ownerCommon.Addr, scopeSpecGood.SpecificationId)},
			{"scopeSpecGood cSpecGood address", types.GetContractSpecScopeSpecCacheKey(cSpecGood.SpecificationId, scopeSpecGood.SpecificationId)},

			{"scopeSpecBad ownerBad address", types.GetAddressScopeSpecCacheKey(ownerBad.Addr, scopeSpecBad.SpecificationId)},
			{"scopeSpecBad ownerCommon address", types.GetAddressScopeSpecCacheKey(ownerCommon.Addr, scopeSpecBad.SpecificationId)},
			{"scopeSpecBad cSpecBad address", types.GetContractSpecScopeSpecCacheKey(cSpecBad.SpecificationId, scopeSpecBad.SpecificationId)},

			{"scopeSpecNew ownerNew address", types.GetAddressScopeSpecCacheKey(ownerNew.Addr, scopeSpecNew.SpecificationId)},
			{"scopeSpecNew ownerCommon address", types.GetAddressScopeSpecCacheKey(ownerCommon.Addr, scopeSpecNew.SpecificationId)},
			{"scopeSpecNew cSpecNew address", types.GetContractSpecScopeSpecCacheKey(cSpecNew.SpecificationId, scopeSpecNew.SpecificationId)},
		}

		// concatBz creates a new []byte containing the bytes in the three things provided.
		concatBz := func(pre, accAddr, mdAddr []byte) []byte {
			rv := make([]byte, len(pre)+len(accAddr)+len(mdAddr))
			copy(rv, pre)
			copy(rv[len(pre):], accAddr)
			copy(rv[len(pre)+len(accAddr):], mdAddr)
			return rv
		}

		// Create a list of indexes that should have existed prior to the v1 to v2 migration.
		preExistingIndexes := namedIndexList{
			{"pre-existing scopeGood ownerGood address", concatBz(v042.AddressScopeCacheKeyPrefixLegacy, ownerGood.Addr, scopeGood.ScopeId)},
			{"pre-existing scopeGood ownerCommon address", concatBz(v042.AddressScopeCacheKeyPrefixLegacy, ownerCommon.Addr, scopeGood.ScopeId)},
			{"pre-existing scopeGood valueOwnerGood address", concatBz(v042.AddressScopeCacheKeyPrefixLegacy, valueOwnerGood.Addr, scopeGood.ScopeId)},
			{"pre-existing scopeGood value owner", concatBz(v042.ValueOwnerScopeCacheKeyPrefixLegacy, valueOwnerGood.Addr, scopeGood.ScopeId)},
			{"pre-existing scopeGood scope spec", types.GetScopeSpecScopeCacheKey(scopeGood.SpecificationId, scopeGood.ScopeId)},

			{"pre-existing scopeBad ownerBad address", concatBz(v042.AddressScopeCacheKeyPrefixLegacy, ownerBad.Addr, scopeBad.ScopeId)},
			{"pre-existing scopeBad ownerCommon address", concatBz(v042.AddressScopeCacheKeyPrefixLegacy, ownerCommon.Addr, scopeBad.ScopeId)},
			{"pre-existing scopeBad valueOwnerBad address", concatBz(v042.AddressScopeCacheKeyPrefixLegacy, valueOwnerBad.Addr, scopeBad.ScopeId)},
			{"pre-existing scopeBad value owner", concatBz(v042.ValueOwnerScopeCacheKeyPrefixLegacy, valueOwnerBad.Addr, scopeBad.ScopeId)},
			{"pre-existing scopeBad scope spec", types.GetScopeSpecScopeCacheKey(scopeBad.SpecificationId, scopeBad.ScopeId)},

			{"pre-existing cSpecGood ownerGood address", concatBz(v042.AddressContractSpecCacheKeyPrefixLegacy, ownerGood.Addr, cSpecGood.SpecificationId)},
			{"pre-existing cSpecGood ownerCommon address", concatBz(v042.AddressContractSpecCacheKeyPrefixLegacy, ownerCommon.Addr, cSpecGood.SpecificationId)},

			{"pre-existing cSpecBad ownerBad address", concatBz(v042.AddressContractSpecCacheKeyPrefixLegacy, ownerBad.Addr, cSpecBad.SpecificationId)},
			{"pre-existing cSpecBad ownerCommon address", concatBz(v042.AddressContractSpecCacheKeyPrefixLegacy, ownerCommon.Addr, cSpecBad.SpecificationId)},

			{"pre-existing scopeSpecGood ownerGood address", concatBz(v042.AddressScopeSpecCacheKeyPrefixLegacy, ownerGood.Addr, scopeSpecGood.SpecificationId)},
			{"pre-existing scopeSpecGood ownerCommon address", concatBz(v042.AddressScopeSpecCacheKeyPrefixLegacy, ownerCommon.Addr, scopeSpecGood.SpecificationId)},
			{"pre-existing scopeSpecGood cSpecGood address", types.GetContractSpecScopeSpecCacheKey(cSpecGood.SpecificationId, scopeSpecGood.SpecificationId)},

			{"pre-existing scopeSpecBad ownerBad address", concatBz(v042.AddressScopeSpecCacheKeyPrefixLegacy, ownerBad.Addr, scopeSpecBad.SpecificationId)},
			{"pre-existing scopeSpecBad ownerCommon address", concatBz(v042.AddressScopeSpecCacheKeyPrefixLegacy, ownerCommon.Addr, scopeSpecBad.SpecificationId)},
			{"pre-existing scopeSpecBad cSpecBad address", types.GetContractSpecScopeSpecCacheKey(cSpecBad.SpecificationId, scopeSpecBad.SpecificationId)},
		}

		newIndexes := namedIndexList{
			{"add after v2 scopeNew ownerNew address", types.GetAddressScopeCacheKey(ownerNew.Addr, scopeNew.ScopeId)},
			{"add after v2 scopeNew ownerCommon address", types.GetAddressScopeCacheKey(ownerCommon.Addr, scopeNew.ScopeId)},
			{"add after v2 scopeNew valueOwnerNew address", types.GetAddressScopeCacheKey(valueOwnerNew.Addr, scopeNew.ScopeId)},
			{"add after v2 scopeNew value owner", types.GetValueOwnerScopeCacheKey(valueOwnerNew.Addr, scopeNew.ScopeId)},
			{"add after v2 scopeNew scope spec", types.GetScopeSpecScopeCacheKey(scopeNew.SpecificationId, scopeNew.ScopeId)},

			{"add after v2 cSpecNew ownerNew address", types.GetAddressContractSpecCacheKey(ownerNew.Addr, cSpecNew.SpecificationId)},
			{"add after v2 cSpecNew ownerCommon address", types.GetAddressContractSpecCacheKey(ownerCommon.Addr, cSpecNew.SpecificationId)},

			{"add after v2 scopeSpecNew ownerNew address", types.GetAddressScopeSpecCacheKey(ownerNew.Addr, scopeSpecNew.SpecificationId)},
			{"add after v2 scopeSpecNew ownerCommon address", types.GetAddressScopeSpecCacheKey(ownerCommon.Addr, scopeSpecNew.SpecificationId)},
			{"add after v2 scopeSpecNew cSpecNew address", types.GetContractSpecScopeSpecCacheKey(cSpecNew.SpecificationId, scopeSpecNew.SpecificationId)},
		}

		// Manually marshal and set the pre v1 to v2 items to bypass the keeper's auto-indexing stuff.
		s.store.Set(scopeGood.ScopeId, s.app.AppCodec().MustMarshal(&scopeGood))
		s.store.Set(scopeBad.ScopeId, s.app.AppCodec().MustMarshal(&scopeBad))
		s.store.Set(cSpecGood.SpecificationId, s.app.AppCodec().MustMarshal(&cSpecGood))
		s.store.Set(cSpecBad.SpecificationId, s.app.AppCodec().MustMarshal(&cSpecBad))
		s.store.Set(scopeSpecGood.SpecificationId, s.app.AppCodec().MustMarshal(&scopeSpecGood))
		s.store.Set(scopeSpecBad.SpecificationId, s.app.AppCodec().MustMarshal(&scopeSpecBad))
		defer func() {
			s.app.MetadataKeeper.RemoveScope(s.ctx, scopeGood.ScopeId)
			s.app.MetadataKeeper.RemoveScope(s.ctx, scopeBad.ScopeId)
			s.Assert().NoError(s.app.MetadataKeeper.RemoveScopeSpecification(s.ctx, scopeSpecGood.SpecificationId), "scopeSpecGood")
			s.Assert().NoError(s.app.MetadataKeeper.RemoveScopeSpecification(s.ctx, scopeSpecBad.SpecificationId), "scopeSpecBad")
			s.Assert().NoError(s.app.MetadataKeeper.RemoveContractSpecification(s.ctx, cSpecGood.SpecificationId), "cSpecGood")
			s.Assert().NoError(s.app.MetadataKeeper.RemoveContractSpecification(s.ctx, cSpecBad.SpecificationId), "cSpecBad")
		}()

		// Store all the pre-existing indexes
		for _, index := range preExistingIndexes {
			s.store.Set(index.key, []byte{0x01})
		}

		// Run the migration from v1 to v2.
		migrator1 := keeper.NewMigrator(s.app.MetadataKeeper)
		require.NoError(t, migrator1.Migrate2to3(s.ctx), "running migration v1 to v2")

		s.store.Set(scopeNew.ScopeId, s.app.AppCodec().MustMarshal(&scopeNew))
		s.store.Set(cSpecNew.SpecificationId, s.app.AppCodec().MustMarshal(&cSpecNew))
		s.store.Set(scopeSpecNew.SpecificationId, s.app.AppCodec().MustMarshal(&scopeSpecNew))
		defer func() {
			s.app.MetadataKeeper.RemoveScope(s.ctx, scopeNew.ScopeId)
			s.Assert().NoError(s.app.MetadataKeeper.RemoveScopeSpecification(s.ctx, scopeSpecNew.SpecificationId), "scopeSpecNew")
			s.Assert().NoError(s.app.MetadataKeeper.RemoveContractSpecification(s.ctx, cSpecNew.SpecificationId), "cSpecNew")
		}()

		// Store all the newly added indexes
		for _, index := range newIndexes {
			s.store.Set(index.key, []byte{0x01})
		}

		// Run the migration.
		migrator2 := keeper.NewMigrator(s.app.MetadataKeeper)
		require.NoError(t, migrator2.Migrate2to3(s.ctx), "running migration v2 to v3")

		// Make sure the indexes are as expected!
		if !s.assertExactIndexList(t, expectedIndexes) {
			fmt.Printf("pre-existing indexes:\n%s", preExistingIndexes)
			fmt.Printf("new indexes:\n%s", newIndexes)
		}
	})
}
