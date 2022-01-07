package keeper_test

import (
	"bytes"
	"fmt"
	"math/rand"
	"sort"
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

	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

type MigrationsTestSuite struct {
	suite.Suite

	app   *simapp.App
	ctx   sdk.Context
	store sdk.KVStore

	newPrefixes        namedIndexList
	unaffectedPrefixes namedIndexList
	oldPrefixes        namedIndexList
	allPrefixes        namedIndexList
}

func (s *MigrationsTestSuite) SetupTest() {
	s.app = simapp.Setup(false)
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{}).WithLogger(log.TestingLogger())
	s.store = s.ctx.KVStore(s.app.GetKey(types.ModuleName))

	s.newPrefixes = namedIndexList{
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
	s.allPrefixes = append(s.allPrefixes, s.newPrefixes...)
	s.allPrefixes = append(s.allPrefixes, s.unaffectedPrefixes...)
	s.allPrefixes = append(s.allPrefixes, s.oldPrefixes...)
}

func TestMigrationsTestSuite(t *testing.T) {
	suite.Run(t, new(MigrationsTestSuite))
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

func (s *MigrationsTestSuite) newNamedIndex(name string, key []byte) namedIndex {
	for _, pre := range s.allPrefixes {
		if bytes.HasPrefix(key, pre.key) {
			if len(name) > 0 {
				return namedIndex{pre.name + ": " + name, key}
			}
			return namedIndex{pre.name, key}
		}
	}
	return namedIndex{name, key}
}

// namedIndexListSorter implements sort.Interface for namedIndexListSorter based on the key.
type namedIndexListSorter namedIndexList

func (l namedIndexListSorter) Len() int      { return len(l) }
func (l namedIndexListSorter) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
func (l namedIndexListSorter) Less(i, j int) bool {
	if len(l[i].key) < 4 || len(l[j].key) < 4 {
		return bytes.Compare(l[i].key, l[j].key) < 0
	}
	return bytes.Compare(l[i].key[2:], l[j].key[2:]) < 0
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
func (s *MigrationsTestSuite) assertDeleteAllIndexes(t *testing.T, prefixes namedIndexList) bool {
	if t == nil {
		t = s.T()
	}
	allPassed := true
	for _, pre := range prefixes {
		// Using assert here (instead of require) so it's easier to later decide if a test should continue.
		if !assert.NoError(t, clearStore(prefix.NewStore(s.store, pre.key)), "clearing %s indexes", pre.name) {
			allPassed = false
		}
	}
	return allPassed
}

// requireDeleteAllIndexes is the same as assertDeleteAllIndexes but ends the test if an error occurs.
// If t is nil, s.T() is used when a *testing.T is needed.
func (s *MigrationsTestSuite) requireDeleteAllIndexes(t *testing.T, prefixes namedIndexList) {
	if !(s.assertDeleteAllIndexes(t, prefixes)) {
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

// assertExactIndexList makes sure that the provided set of indexes is equal to the set of all indexes in the store.
func (s *MigrationsTestSuite) assertExactIndexList(t *testing.T, indexes namedIndexList) bool {
	if t == nil {
		t = s.T()
	}
	// Make sure all the provided entries exist.
	found := namedIndexList{}
	missing := namedIndexList{}
	for _, index := range indexes {
		if !s.store.Has(index.key) {
			missing = append(missing, index)
		} else {
			found = append(found, index)
		}
	}
	unknown := namedIndexList{}
	for _, pre := range s.allPrefixes {
		for _, key := range s.getAllStoreKeys(pre.key) {
			if !keyInList(key, indexes) {
				unknown = append(unknown, namedIndex{pre.name, key})
			}
		}
	}
	allFound := assert.Len(t, missing, 0, "missing entries")
	if !allFound {
		fmt.Printf("found entries (%d):\n%s\n", len(found), found)
	}
	allKnown := assert.Len(t, unknown, 0, "unexpected entries")
	if !allKnown {
		fmt.Printf("expected entries (%d):\n%s\n", len(indexes), indexes)
	}
	return allFound && allKnown
}

func (s *MigrationsTestSuite) getAllStoreKeys(pre []byte) [][]byte {
	rv := [][]byte{}
	pStore := prefix.NewStore(s.store, pre)
	iter := pStore.Iterator(nil, nil)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		fullKey := make([]byte, len(pre)+len(iter.Key()))
		copy(fullKey, pre)
		copy(fullKey[len(pre):], iter.Key())
		rv = append(rv, fullKey)
	}
	return rv
}

func keyInList(key []byte, indexes namedIndexList) bool {
	for _, index := range indexes {
		if bytes.Equal(key, index.key) {
			return true
		}
	}
	return false
}

func (s *MigrationsTestSuite) Test2To3() {
	s.T().Run("existing bad indexes are deleted", func(t *testing.T) {
		// makeIndexRand creates a random byte slice that starts with the given prefix.
		// These are bad addresses and need to be re-indexed (the 2nd byte will never be the proper addr length value).
		makeIndexRand := func(pre []byte, length int) []byte {
			rv := randomByteSlice(length)
			copy(rv, pre)
			// Make sure the length byte isn't accidentally correct (17 md addr bytes + 1 pre byte + 1 length byte).
			if len(pre) < 2 && rv[1] == byte(length-19) {
				rv[1] = byte(length - 20)
			}
			return rv
		}
		// makeGoodIndex creates a random key with length prefix + {length - 19} address bytes + 17 random bytes.
		// These are good addresses that don't need re-indexing.
		makeGoodIndex := func(pre []byte, length int) []byte {
			rv := makeIndexRand(pre, length)
			rv[1] = byte(length - 19)
			return rv
		}
		// makeIndexShort creates a random key with length prefix + {length - 20} address bytes + 16 random bytes.
		// These are bad addresses and need to be re-indexed.
		makeIndexShort := func(pre []byte, length int) []byte {
			rv := makeIndexRand(pre, length)
			rv[1] = byte(length - 20)
			return rv
		}
		// makeIndexShort creates a random key with length prefix + {length - 18} address bytes + 18 random bytes.
		// These are bad addresses and need to be re-indexed.
		makeIndexLong := func(pre []byte, length int) []byte {
			rv := makeIndexRand(pre, length)
			rv[1] = byte(length - 18)
			return rv
		}
		// makeIndexZeros creates a byte slice starting with a prefix followed by zeros until it has length 60.
		// These are bad addresses and need to be re-indexed; the address length byte is zero.
		makeIndexZeros := func(pre []byte, length int) []byte {
			rv := make([]byte, length)
			copy(rv, pre)
			return rv
		}

		// makeGoodIndexes makes indexes that are good and should not be deleted.
		makeGoodIndexes := func(name string, pre []byte) namedIndexList {
			rv := namedIndexList{}
			// note: 39 and 51 are the "real" good lengths, the rest just kind of check math stuff.
			for _, v := range []int{38, 39, 40, 50, 51, 52, 60, 68} {
				rv = append(rv, namedIndex{fmt.Sprintf("%s good %d", name, v), makeGoodIndex(pre, v)})
			}
			return rv
		}
		// makeBadIndexes makes indexes that are bad and should end up being deleted.
		makeBadIndexes := func(name string, pre []byte) namedIndexList {
			rv := namedIndexList{}
			for _, v := range []int{38, 39, 40, 50, 51, 52, 60, 68} {
				rv = append(rv, namedIndex{fmt.Sprintf("%s short %d", name, v), makeIndexShort(pre, v)})
				rv = append(rv, namedIndex{fmt.Sprintf("%s long %d", name, v), makeIndexLong(pre, v)})
				rv = append(rv, namedIndex{fmt.Sprintf("%s zeros %d", name, v), makeIndexZeros(pre, v)})
				rv = append(rv, namedIndex{fmt.Sprintf("%s rand %d", name, v), makeIndexRand(pre, v)})
			}
			return rv
		}

		indexes := namedIndexList{}
		expectedKeptIndexes := namedIndexList{}
		for _, pre := range s.newPrefixes {
			indexes = append(indexes, makeBadIndexes(pre.name, pre.key)...)
			expectedKeptIndexes = append(expectedKeptIndexes, makeGoodIndexes(pre.name, pre.key)...)
		}
		for _, pre := range s.unaffectedPrefixes {
			expectedKeptIndexes = append(expectedKeptIndexes, makeGoodIndexes(pre.name, pre.key)...)
			// Note: These are expected to be kept, even though they're "bad".
			// There is no scenario where they'd end up actually existing.
			// If they're involved in this test failing, the migration is doing more than it needs to.
			expectedKeptIndexes = append(expectedKeptIndexes, makeBadIndexes(pre.name, pre.key)...)
		}
		indexes = append(indexes, expectedKeptIndexes...)

		// Add them to the store.
		for _, index := range indexes {
			s.store.Set(index.key, []byte{0x01})
		}
		defer func() {
			for _, index := range indexes {
				s.store.Delete(index.key)
			}
		}()

		// Do the migration.
		migrator := keeper.NewMigrator(s.app.MetadataKeeper)
		require.NoError(t, migrator.Migrate2to3(s.ctx), "running migration")

		s.assertExactIndexList(t, expectedKeptIndexes)
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
			s.newNamedIndex("scope1 owner1", types.GetAddressScopeCacheKey(owner1.Addr, scope1.ScopeId)),
			s.newNamedIndex("scope1 ownerCommon", types.GetAddressScopeCacheKey(ownerCommon.Addr, scope1.ScopeId)),
			s.newNamedIndex("scope1 valueOwner1", types.GetAddressScopeCacheKey(valueOwner1.Addr, scope1.ScopeId)),
			s.newNamedIndex("scope1 valueOwner1", types.GetValueOwnerScopeCacheKey(valueOwner1.Addr, scope1.ScopeId)),
			s.newNamedIndex("scope1", types.GetScopeSpecScopeCacheKey(scope1.SpecificationId, scope1.ScopeId)),

			s.newNamedIndex("scope2 owner2", types.GetAddressScopeCacheKey(owner2.Addr, scope2.ScopeId)),
			s.newNamedIndex("scope2 ownerCommon", types.GetAddressScopeCacheKey(ownerCommon.Addr, scope2.ScopeId)),
			s.newNamedIndex("scope2 valueOwner2", types.GetAddressScopeCacheKey(valueOwner2.Addr, scope2.ScopeId)),
			s.newNamedIndex("scope2 valueOwner2", types.GetValueOwnerScopeCacheKey(valueOwner2.Addr, scope2.ScopeId)),
			s.newNamedIndex("scope2", types.GetScopeSpecScopeCacheKey(scope2.SpecificationId, scope2.ScopeId)),
		}

		// Set the scopes.
		s.app.MetadataKeeper.SetScope(s.ctx, scope1)
		s.app.MetadataKeeper.SetScope(s.ctx, scope2)
		defer func() {
			s.app.MetadataKeeper.RemoveScope(s.ctx, scope1.ScopeId)
			s.app.MetadataKeeper.RemoveScope(s.ctx, scope2.ScopeId)
		}()

		// Delete any indexes added for them.
		s.requireDeleteAllIndexes(t, s.newPrefixes)
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
			s.newNamedIndex("scopeSpec1 owner1", types.GetAddressScopeSpecCacheKey(owner1.Addr, scopeSpec1.SpecificationId)),
			s.newNamedIndex("scopeSpec1 ownerCommon", types.GetAddressScopeSpecCacheKey(ownerCommon.Addr, scopeSpec1.SpecificationId)),
			s.newNamedIndex("scopeSpec1 cSpec1", types.GetContractSpecScopeSpecCacheKey(cSpec1ID, scopeSpec1.SpecificationId)),
			s.newNamedIndex("scopeSpec1 cSpecCommon", types.GetContractSpecScopeSpecCacheKey(cSpecCommonID, scopeSpec1.SpecificationId)),

			s.newNamedIndex("scopeSpec2 owner2", types.GetAddressScopeSpecCacheKey(owner2.Addr, scopeSpec2.SpecificationId)),
			s.newNamedIndex("scopeSpec2 ownerCommon", types.GetAddressScopeSpecCacheKey(ownerCommon.Addr, scopeSpec2.SpecificationId)),
			s.newNamedIndex("scopeSpec2 cSpec2", types.GetContractSpecScopeSpecCacheKey(cSpec2ID, scopeSpec2.SpecificationId)),
			s.newNamedIndex("scopeSpec2 cSpecCommon", types.GetContractSpecScopeSpecCacheKey(cSpecCommonID, scopeSpec2.SpecificationId)),
		}

		// Set the scopes specs.
		s.app.MetadataKeeper.SetScopeSpecification(s.ctx, scopeSpec1)
		s.app.MetadataKeeper.SetScopeSpecification(s.ctx, scopeSpec2)
		defer func() {
			assert.NoError(t, s.app.MetadataKeeper.RemoveScopeSpecification(s.ctx, scopeSpec1.SpecificationId), "removing scopeSpec1")
			assert.NoError(t, s.app.MetadataKeeper.RemoveScopeSpecification(s.ctx, scopeSpec2.SpecificationId), "removing scopeSpec2")
		}()

		// Delete any indexes added for them.
		s.requireDeleteAllIndexes(t, s.newPrefixes)
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
			s.newNamedIndex("cSpec1 owner1", types.GetAddressContractSpecCacheKey(owner1.Addr, cSpec1.SpecificationId)),
			s.newNamedIndex("cSpec1 ownerCommon", types.GetAddressContractSpecCacheKey(ownerCommon.Addr, cSpec1.SpecificationId)),

			s.newNamedIndex("cSpec2 owner2", types.GetAddressContractSpecCacheKey(owner2.Addr, cSpec2.SpecificationId)),
			s.newNamedIndex("cSpec2 ownerCommon", types.GetAddressContractSpecCacheKey(ownerCommon.Addr, cSpec2.SpecificationId)),
		}

		// Set the contract specs.
		s.app.MetadataKeeper.SetContractSpecification(s.ctx, cSpec1)
		s.app.MetadataKeeper.SetContractSpecification(s.ctx, cSpec2)
		defer func() {
			assert.NoError(t, s.app.MetadataKeeper.RemoveContractSpecification(s.ctx, cSpec1.SpecificationId), "removing cSpec1")
			assert.NoError(t, s.app.MetadataKeeper.RemoveContractSpecification(s.ctx, cSpec2.SpecificationId), "removing cSpec2")
		}()

		// Delete any indexes added for them.
		s.requireDeleteAllIndexes(t, s.newPrefixes)
		// Run the migration
		migrator := keeper.NewMigrator(s.app.MetadataKeeper)
		require.NoError(t, migrator.Migrate2to3(s.ctx), "running migration")

		// Make sure the indexes are as expected.
		s.assertExactIndexList(t, indexes)
	})

	s.T().Run("good bad new full run all fixed", func(t *testing.T) {
		// This test mimics a state where there are three metadata entries of each type affected by the v1 to v2 migration.
		// 1) A "good" entry that was migrated from V1 to V2, and has been written since then (so it is correctly indexed).
		// 2) A "bad" entry that was migrated from V1 to V2, but has NOT been written since then.
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

		// getAllIndexes gets all the index keys from the store.
		getAllIndexes := func() namedIndexList {
			rv := namedIndexList{}
			for _, pre := range s.allPrefixes {
				for _, key := range s.getAllStoreKeys(pre.key) {
					rv = append(rv, namedIndex{pre.name + ": pre", key})
				}
			}
			sort.Sort(namedIndexListSorter(rv))
			return rv
		}

		// concatBz creates a new []byte containing the bytes in the three things provided.
		concatBz := func(pre, accAddr, mdAddr []byte) []byte {
			rv := make([]byte, len(pre)+len(accAddr)+len(mdAddr))
			copy(rv, pre)
			copy(rv[len(pre):], accAddr)
			copy(rv[len(pre)+len(accAddr):], mdAddr)
			return rv
		}

		// toBadIndex replicates what was done to the keys in the first migration.
		toBadIndex := func(oldKey []byte, indexer func(addr sdk.AccAddress) []byte) []byte {
			iterKey := oldKey[1:]
			legacyAddress := sdk.AccAddress(iterKey[1:21])
			newStoreKey := indexer(legacyAddress)
			metaaddress := iterKey[21:]
			return append(newStoreKey, metaaddress...)
		}

		// badAddrScopeInd creates the bad version of the address -> scope index keys.
		badAddrScopeInd := func(accAddr sdk.AccAddress, mdAddr types.MetadataAddress) []byte {
			return toBadIndex(
				concatBz(v042.AddressScopeCacheKeyPrefixLegacy, accAddr, mdAddr),
				types.GetAddressScopeCacheIteratorPrefix)
		}
		// badVOScopeInd creates the bad version of the value owner -> scope index keys.
		badVOScopeInd := func(accAddr sdk.AccAddress, mdAddr types.MetadataAddress) []byte {
			return toBadIndex(
				concatBz(v042.ValueOwnerScopeCacheKeyPrefixLegacy, accAddr, mdAddr),
				types.GetValueOwnerScopeCacheIteratorPrefix)
		}
		// badAddrCSpecInd creates the bad version of the address -> contract spec index keys.
		badAddrCSpecInd := func(accAddr sdk.AccAddress, mdAddr types.MetadataAddress) []byte {
			return toBadIndex(
				concatBz(v042.AddressContractSpecCacheKeyPrefixLegacy, accAddr, mdAddr),
				types.GetAddressContractSpecCacheIteratorPrefix)
		}
		// badAddrScopeSpecInd creates the bad version of the address -> scope spec index keys.
		badAddrScopeSpecInd := func(accAddr sdk.AccAddress, mdAddr types.MetadataAddress) []byte {
			return toBadIndex(
				concatBz(v042.AddressScopeSpecCacheKeyPrefixLegacy, accAddr, mdAddr),
				types.GetAddressScopeSpecCacheIteratorPrefix)
		}

		// expectedIndexes is all the keys that should exist for the metadata things above after the migration to V3.
		expectedIndexes := namedIndexList{
			s.newNamedIndex("scopeGood ownerGood", types.GetAddressScopeCacheKey(ownerGood.Addr, scopeGood.ScopeId)),
			s.newNamedIndex("scopeGood ownerCommon", types.GetAddressScopeCacheKey(ownerCommon.Addr, scopeGood.ScopeId)),
			s.newNamedIndex("scopeGood valueOwnerGood", types.GetAddressScopeCacheKey(valueOwnerGood.Addr, scopeGood.ScopeId)),
			s.newNamedIndex("scopeGood valueOwnerGood", types.GetValueOwnerScopeCacheKey(valueOwnerGood.Addr, scopeGood.ScopeId)),
			s.newNamedIndex("scopeGood", types.GetScopeSpecScopeCacheKey(scopeGood.SpecificationId, scopeGood.ScopeId)),

			s.newNamedIndex("scopeBad ownerBad", types.GetAddressScopeCacheKey(ownerBad.Addr, scopeBad.ScopeId)),
			s.newNamedIndex("scopeBad ownerCommon", types.GetAddressScopeCacheKey(ownerCommon.Addr, scopeBad.ScopeId)),
			s.newNamedIndex("scopeBad valueOwnerBad", types.GetAddressScopeCacheKey(valueOwnerBad.Addr, scopeBad.ScopeId)),
			s.newNamedIndex("scopeBad valueOwnerBad", types.GetValueOwnerScopeCacheKey(valueOwnerBad.Addr, scopeBad.ScopeId)),
			s.newNamedIndex("scopeBad", types.GetScopeSpecScopeCacheKey(scopeBad.SpecificationId, scopeBad.ScopeId)),

			s.newNamedIndex("scopeNew ownerNew", types.GetAddressScopeCacheKey(ownerNew.Addr, scopeNew.ScopeId)),
			s.newNamedIndex("scopeNew ownerCommon", types.GetAddressScopeCacheKey(ownerCommon.Addr, scopeNew.ScopeId)),
			s.newNamedIndex("scopeNew valueOwnerNew", types.GetAddressScopeCacheKey(valueOwnerNew.Addr, scopeNew.ScopeId)),
			s.newNamedIndex("scopeNew valueOwnerNew", types.GetValueOwnerScopeCacheKey(valueOwnerNew.Addr, scopeNew.ScopeId)),
			s.newNamedIndex("scopeNew", types.GetScopeSpecScopeCacheKey(scopeNew.SpecificationId, scopeNew.ScopeId)),

			s.newNamedIndex("cSpecGood ownerGood", types.GetAddressContractSpecCacheKey(ownerGood.Addr, cSpecGood.SpecificationId)),
			s.newNamedIndex("cSpecGood ownerCommon", types.GetAddressContractSpecCacheKey(ownerCommon.Addr, cSpecGood.SpecificationId)),

			s.newNamedIndex("cSpecBad ownerBad", types.GetAddressContractSpecCacheKey(ownerBad.Addr, cSpecBad.SpecificationId)),
			s.newNamedIndex("cSpecBad ownerCommon", types.GetAddressContractSpecCacheKey(ownerCommon.Addr, cSpecBad.SpecificationId)),

			s.newNamedIndex("cSpecNew ownerNew", types.GetAddressContractSpecCacheKey(ownerNew.Addr, cSpecNew.SpecificationId)),
			s.newNamedIndex("cSpecNew ownerCommon", types.GetAddressContractSpecCacheKey(ownerCommon.Addr, cSpecNew.SpecificationId)),

			s.newNamedIndex("scopeSpecGood ownerGood", types.GetAddressScopeSpecCacheKey(ownerGood.Addr, scopeSpecGood.SpecificationId)),
			s.newNamedIndex("scopeSpecGood ownerCommon", types.GetAddressScopeSpecCacheKey(ownerCommon.Addr, scopeSpecGood.SpecificationId)),
			s.newNamedIndex("scopeSpecGood cSpecGood", types.GetContractSpecScopeSpecCacheKey(cSpecGood.SpecificationId, scopeSpecGood.SpecificationId)),

			s.newNamedIndex("scopeSpecBad ownerBad", types.GetAddressScopeSpecCacheKey(ownerBad.Addr, scopeSpecBad.SpecificationId)),
			s.newNamedIndex("scopeSpecBad ownerCommon", types.GetAddressScopeSpecCacheKey(ownerCommon.Addr, scopeSpecBad.SpecificationId)),
			s.newNamedIndex("scopeSpecBad cSpecBad", types.GetContractSpecScopeSpecCacheKey(cSpecBad.SpecificationId, scopeSpecBad.SpecificationId)),

			s.newNamedIndex("scopeSpecNew ownerNew", types.GetAddressScopeSpecCacheKey(ownerNew.Addr, scopeSpecNew.SpecificationId)),
			s.newNamedIndex("scopeSpecNew ownerCommon", types.GetAddressScopeSpecCacheKey(ownerCommon.Addr, scopeSpecNew.SpecificationId)),
			s.newNamedIndex("scopeSpecNew cSpecNew", types.GetContractSpecScopeSpecCacheKey(cSpecNew.SpecificationId, scopeSpecNew.SpecificationId)),
		}

		// preExistingIndexes is all the keys that need to exist prior to the v1 to v2 migration.
		preExistingIndexes := namedIndexList{
			s.newNamedIndex("pe scopeGood ownerGood", concatBz(v042.AddressScopeCacheKeyPrefixLegacy, ownerGood.Addr, scopeGood.ScopeId)),
			s.newNamedIndex("pe scopeGood ownerCommon", concatBz(v042.AddressScopeCacheKeyPrefixLegacy, ownerCommon.Addr, scopeGood.ScopeId)),
			s.newNamedIndex("pe scopeGood valueOwnerGood", concatBz(v042.AddressScopeCacheKeyPrefixLegacy, valueOwnerGood.Addr, scopeGood.ScopeId)),
			s.newNamedIndex("pe scopeGood valueOwnerGood", concatBz(v042.ValueOwnerScopeCacheKeyPrefixLegacy, valueOwnerGood.Addr, scopeGood.ScopeId)),
			s.newNamedIndex("pe scopeGood", types.GetScopeSpecScopeCacheKey(scopeGood.SpecificationId, scopeGood.ScopeId)),

			s.newNamedIndex("pe scopeBad ownerBad", concatBz(v042.AddressScopeCacheKeyPrefixLegacy, ownerBad.Addr, scopeBad.ScopeId)),
			s.newNamedIndex("pe scopeBad ownerCommon", concatBz(v042.AddressScopeCacheKeyPrefixLegacy, ownerCommon.Addr, scopeBad.ScopeId)),
			s.newNamedIndex("pe scopeBad valueOwnerBad", concatBz(v042.AddressScopeCacheKeyPrefixLegacy, valueOwnerBad.Addr, scopeBad.ScopeId)),
			s.newNamedIndex("pe scopeBad valueOwnerBad", concatBz(v042.ValueOwnerScopeCacheKeyPrefixLegacy, valueOwnerBad.Addr, scopeBad.ScopeId)),
			s.newNamedIndex("pe scopeBad", types.GetScopeSpecScopeCacheKey(scopeBad.SpecificationId, scopeBad.ScopeId)),

			s.newNamedIndex("pe cSpecGood ownerGood", concatBz(v042.AddressContractSpecCacheKeyPrefixLegacy, ownerGood.Addr, cSpecGood.SpecificationId)),
			s.newNamedIndex("pe cSpecGood ownerCommon", concatBz(v042.AddressContractSpecCacheKeyPrefixLegacy, ownerCommon.Addr, cSpecGood.SpecificationId)),

			s.newNamedIndex("pe cSpecBad ownerBad", concatBz(v042.AddressContractSpecCacheKeyPrefixLegacy, ownerBad.Addr, cSpecBad.SpecificationId)),
			s.newNamedIndex("pe cSpecBad ownerCommon", concatBz(v042.AddressContractSpecCacheKeyPrefixLegacy, ownerCommon.Addr, cSpecBad.SpecificationId)),

			s.newNamedIndex("pe scopeSpecGood ownerGood", concatBz(v042.AddressScopeSpecCacheKeyPrefixLegacy, ownerGood.Addr, scopeSpecGood.SpecificationId)),
			s.newNamedIndex("pe scopeSpecGood ownerCommon", concatBz(v042.AddressScopeSpecCacheKeyPrefixLegacy, ownerCommon.Addr, scopeSpecGood.SpecificationId)),
			s.newNamedIndex("pe scopeSpecGood cSpecGood", types.GetContractSpecScopeSpecCacheKey(cSpecGood.SpecificationId, scopeSpecGood.SpecificationId)),

			s.newNamedIndex("pe scopeSpecBad ownerBad", concatBz(v042.AddressScopeSpecCacheKeyPrefixLegacy, ownerBad.Addr, scopeSpecBad.SpecificationId)),
			s.newNamedIndex("pe scopeSpecBad ownerCommon", concatBz(v042.AddressScopeSpecCacheKeyPrefixLegacy, ownerCommon.Addr, scopeSpecBad.SpecificationId)),
			s.newNamedIndex("pe scopeSpecBad cSpecBad", types.GetContractSpecScopeSpecCacheKey(cSpecBad.SpecificationId, scopeSpecBad.SpecificationId)),
		}

		// expectedIntermediateIndexes is all the keys expected to exist after the V1 to V2 migration.
		expectedIntermediateIndexes := namedIndexList{
			s.newNamedIndex("int scopeGood ownerGood", badAddrScopeInd(ownerGood.Addr, scopeGood.ScopeId)),
			s.newNamedIndex("int scopeGood ownerCommon", badAddrScopeInd(ownerCommon.Addr, scopeGood.ScopeId)),
			s.newNamedIndex("int scopeGood valueOwnerGood", badAddrScopeInd(valueOwnerGood.Addr, scopeGood.ScopeId)),
			s.newNamedIndex("int scopeGood valueOwnerGood", badVOScopeInd(valueOwnerGood.Addr, scopeGood.ScopeId)),
			s.newNamedIndex("int scopeGood", types.GetScopeSpecScopeCacheKey(scopeGood.SpecificationId, scopeGood.ScopeId)),

			s.newNamedIndex("int scopeBad ownerBad", badAddrScopeInd(ownerBad.Addr, scopeBad.ScopeId)),
			s.newNamedIndex("int scopeBad ownerCommon", badAddrScopeInd(ownerCommon.Addr, scopeBad.ScopeId)),
			s.newNamedIndex("int scopeBad valueOwnerBad", badAddrScopeInd(valueOwnerBad.Addr, scopeBad.ScopeId)),
			s.newNamedIndex("int scopeBad valueOwnerBad", badVOScopeInd(valueOwnerBad.Addr, scopeBad.ScopeId)),
			s.newNamedIndex("int scopeBad", types.GetScopeSpecScopeCacheKey(scopeBad.SpecificationId, scopeBad.ScopeId)),

			s.newNamedIndex("int cSpecGood ownerGood", badAddrCSpecInd(ownerGood.Addr, cSpecGood.SpecificationId)),
			s.newNamedIndex("int cSpecGood ownerCommon", badAddrCSpecInd(ownerCommon.Addr, cSpecGood.SpecificationId)),

			s.newNamedIndex("int cSpecBad ownerBad", badAddrCSpecInd(ownerBad.Addr, cSpecBad.SpecificationId)),
			s.newNamedIndex("int cSpecBad ownerCommon", badAddrCSpecInd(ownerCommon.Addr, cSpecBad.SpecificationId)),

			s.newNamedIndex("int scopeSpecGood ownerGood", badAddrScopeSpecInd(ownerGood.Addr, scopeSpecGood.SpecificationId)),
			s.newNamedIndex("int scopeSpecGood ownerCommon", badAddrScopeSpecInd(ownerCommon.Addr, scopeSpecGood.SpecificationId)),
			s.newNamedIndex("int scopeSpecGood cSpecGood", types.GetContractSpecScopeSpecCacheKey(cSpecGood.SpecificationId, scopeSpecGood.SpecificationId)),

			s.newNamedIndex("int scopeSpecBad ownerBad", badAddrScopeSpecInd(ownerBad.Addr, scopeSpecBad.SpecificationId)),
			s.newNamedIndex("int scopeSpecBad ownerCommon", badAddrScopeSpecInd(ownerCommon.Addr, scopeSpecBad.SpecificationId)),
			s.newNamedIndex("int scopeSpecBad cSpecBad", types.GetContractSpecScopeSpecCacheKey(cSpecBad.SpecificationId, scopeSpecBad.SpecificationId)),
		}

		// goodIndexes is all the keys that need to be written for the "good" metadata entries between the two migrations (to simulate that it was written).
		goodIndexes := namedIndexList{
			s.newNamedIndex("mid scopeGood ownerGood", types.GetAddressScopeCacheKey(ownerGood.Addr, scopeGood.ScopeId)),
			s.newNamedIndex("mid scopeGood ownerCommon", types.GetAddressScopeCacheKey(ownerCommon.Addr, scopeGood.ScopeId)),
			s.newNamedIndex("mid scopeGood valueOwnerGood", types.GetAddressScopeCacheKey(valueOwnerGood.Addr, scopeGood.ScopeId)),
			s.newNamedIndex("mid scopeGood valueOwnerGood", types.GetValueOwnerScopeCacheKey(valueOwnerGood.Addr, scopeGood.ScopeId)),
			s.newNamedIndex("mid scopeGood", types.GetScopeSpecScopeCacheKey(scopeGood.SpecificationId, scopeGood.ScopeId)),

			s.newNamedIndex("mid cSpecGood ownerGood", types.GetAddressContractSpecCacheKey(ownerGood.Addr, cSpecGood.SpecificationId)),
			s.newNamedIndex("mid cSpecGood ownerCommon", types.GetAddressContractSpecCacheKey(ownerCommon.Addr, cSpecGood.SpecificationId)),

			s.newNamedIndex("mid scopeSpecGood ownerGood", types.GetAddressScopeSpecCacheKey(ownerGood.Addr, scopeSpecGood.SpecificationId)),
			s.newNamedIndex("mid scopeSpecGood ownerCommon", types.GetAddressScopeSpecCacheKey(ownerCommon.Addr, scopeSpecGood.SpecificationId)),
			s.newNamedIndex("mid scopeSpecGood cSpecGood", types.GetContractSpecScopeSpecCacheKey(cSpecGood.SpecificationId, scopeSpecGood.SpecificationId)),
		}

		// newIndexes is all the keys that need to be written for the "new" metadata entry (written between migrations).
		newIndexes := namedIndexList{
			s.newNamedIndex("new scopeNew ownerNew", types.GetAddressScopeCacheKey(ownerNew.Addr, scopeNew.ScopeId)),
			s.newNamedIndex("new scopeNew ownerCommon", types.GetAddressScopeCacheKey(ownerCommon.Addr, scopeNew.ScopeId)),
			s.newNamedIndex("new scopeNew valueOwnerNew", types.GetAddressScopeCacheKey(valueOwnerNew.Addr, scopeNew.ScopeId)),
			s.newNamedIndex("new scopeNew valueOwnerNew", types.GetValueOwnerScopeCacheKey(valueOwnerNew.Addr, scopeNew.ScopeId)),
			s.newNamedIndex("new scopeNew", types.GetScopeSpecScopeCacheKey(scopeNew.SpecificationId, scopeNew.ScopeId)),

			s.newNamedIndex("new cSpecNew ownerNew", types.GetAddressContractSpecCacheKey(ownerNew.Addr, cSpecNew.SpecificationId)),
			s.newNamedIndex("new cSpecNew ownerCommon", types.GetAddressContractSpecCacheKey(ownerCommon.Addr, cSpecNew.SpecificationId)),

			s.newNamedIndex("new scopeSpecNew ownerNew", types.GetAddressScopeSpecCacheKey(ownerNew.Addr, scopeSpecNew.SpecificationId)),
			s.newNamedIndex("new scopeSpecNew ownerCommon", types.GetAddressScopeSpecCacheKey(ownerCommon.Addr, scopeSpecNew.SpecificationId)),
			s.newNamedIndex("new scopeSpecNew cSpecNew", types.GetContractSpecScopeSpecCacheKey(cSpecNew.SpecificationId, scopeSpecNew.SpecificationId)),
		}

		// Manually marshal and set the pre v1 to v2 items, bypassingthe keeper's auto-indexing stuff.
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

		preV1ToV2Indexes := getAllIndexes()
		defer func() {
			if t.Failed() {
				fmt.Printf("Pre V1 to V2 Indexes (%d):\n%s\n", len(preV1ToV2Indexes), preV1ToV2Indexes)
			}
		}()

		// Run the migration from v1 to v2.
		migrator1 := keeper.NewMigrator(s.app.MetadataKeeper)
		require.NoError(t, migrator1.Migrate1to2(s.ctx), "running migration v1 to v2")

		postV1ToV2Indexes := getAllIndexes()
		defer func() {
			if t.Failed() {
				fmt.Printf("Post V1 to V2 Indexes (%d):\n%s\n", len(postV1ToV2Indexes), postV1ToV2Indexes)
			}
		}()

		// Make assumptions about the V1 to V2 migration are correct.
		if !s.assertExactIndexList(t, expectedIntermediateIndexes) {
			t.Log("Unexpected state after first migration.")
			t.FailNow()
		}

		// Pretend the "good" stuff was written, and write those indexes.
		for _, index := range goodIndexes {
			s.store.Set(index.key, []byte{0x01})
		}

		// Store the new stuff same as before, bypassing the keeper's auto-indexing.
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

		//
		// Finally, all the setup is complete and we can do the thing we're actually trying to test here.
		//

		preV2ToV3Indexes := getAllIndexes()
		defer func() {
			if t.Failed() {
				fmt.Printf("Pre V2 to V3 Indexes (%d):\n%s\n", len(preV2ToV3Indexes), preV2ToV3Indexes)
			}
		}()

		// Run the migration.
		migrator2 := keeper.NewMigrator(s.app.MetadataKeeper)
		require.NoError(t, migrator2.Migrate2to3(s.ctx), "running migration v2 to v3")

		postV2ToV3Indexes := getAllIndexes()
		defer func() {
			if t.Failed() {
				fmt.Printf("Post V2 to V3 Indexes (%d):\n%s\n", len(postV2ToV3Indexes), postV2ToV3Indexes)
			}
		}()

		// Make sure the indexes are as expected!
		s.assertExactIndexList(t, expectedIndexes)
	})

	s.T().Run("make sure nothing is indexed anymore", func(t *testing.T) {
		// If this fails while none of the others fail, it means one of the other tests isn't properly identifying everything it should.
		s.assertExactIndexList(t, namedIndexList{})
	})

	s.T().Run("make sure an empty session is removed", func(t *testing.T) {
		scopeUUID := uuid.New()
		sessionUUID := uuid.New()
		sessionID := types.SessionMetadataAddress(scopeUUID, sessionUUID)
		owner := randomUser()
		session := types.Session{
			SessionId:       sessionID,
			SpecificationId: types.ContractSpecMetadataAddress(uuid.New()),
			Parties:         ownerPartyList(owner.Bech32),
			Name:            "deleteme",
		}
		s.store.Set(sessionID, s.app.AppCodec().MustMarshal(&session))

		migrator2 := keeper.NewMigrator(s.app.MetadataKeeper)
		require.NoError(t, migrator2.Migrate2to3(s.ctx), "running migration v2 to v3")

		hasSession := s.store.Has(sessionID)
		if !assert.False(t, hasSession, "session should not exist anymore because it was empty") {
			sessionStillThereBz := s.store.Get(sessionID)
			var sessionStillThere types.Session
			err := s.app.AppCodec().Unmarshal(sessionStillThereBz, &sessionStillThere)
			if assert.NoError(t, err, "unmarshalling session that should not exist anymore") {
				// Doing it both ways here to get it in the output and stop the test. One of these must fail.
				require.Equal(t, session, sessionStillThere, "session still there is different")
				require.NotEqual(t, session, sessionStillThere, "session still there is the same as what was written")
			}
		}
	})

	s.T().Run("make sure a non empty session is not removed", func(t *testing.T) {
		scopeUUID := uuid.New()
		sessionUUID := uuid.New()
		sessionID := types.SessionMetadataAddress(scopeUUID, sessionUUID)
		owner := randomUser()
		session := types.Session{
			SessionId:       sessionID,
			SpecificationId: types.ContractSpecMetadataAddress(uuid.New()),
			Parties:         ownerPartyList(owner.Bech32),
			Name:            "donotdeleteme",
		}
		s.store.Set(sessionID, s.app.AppCodec().MustMarshal(&session))
		defer s.store.Delete(sessionID)
		record := types.Record{
			Name:      "arecord",
			SessionId: sessionID,
			Process: types.Process{
				ProcessId: &types.Process_Hash{Hash: "recordprochash"},
				Name:      "recordproc",
				Method:    "recordprocmethod",
			},
			Inputs: []types.RecordInput{
				{
					Name:     "recordinput1name",
					Source:   &types.RecordInput_Hash{Hash: "recordinput1hash"},
					TypeName: "recordinput1type",
					Status:   types.RecordInputStatus_Proposed,
				},
			},
			Outputs: []types.RecordOutput{
				{
					Hash:   "recordout1",
					Status: types.ResultStatus_RESULT_STATUS_PASS,
				},
				{
					Hash:   "recordout2",
					Status: types.ResultStatus_RESULT_STATUS_PASS,
				},
			},
		}
		record.SpecificationId = types.RecordSpecMetadataAddress(uuid.New(), record.Name)
		recordID := types.RecordMetadataAddress(scopeUUID, record.Name)
		s.store.Set(recordID, s.app.AppCodec().MustMarshal(&record))
		defer s.store.Delete(recordID)

		migrator2 := keeper.NewMigrator(s.app.MetadataKeeper)
		require.NoError(t, migrator2.Migrate2to3(s.ctx), "running migration v2 to v3")

		hasSession := s.store.Has(sessionID)
		require.True(t, hasSession, "session should still exist because it was not empty")
	})

	s.T().Run("A little bit of everything", func(t *testing.T) {
		ownerCount := 5
		ScopeSpecCount := 2
		contractSpecCount := ScopeSpecCount * 2
		RecordSpecCount := contractSpecCount * 2
		scopeCount := ScopeSpecCount * 2 * 50
		sessionCount := scopeCount * 2
		recordCount := scopeCount * 2

		makeUUID := func(v uint8) uuid.UUID {
			rv, err := uuid.FromBytes(bytes.Repeat([]byte{v + 1}, 16))
			if err != nil {
				panic(err)
			}
			return rv
		}
		makeRecordSpecInputs := func(ri, count int) []*types.InputSpecification {
			rv := make([]*types.InputSpecification, count)
			for i := range rv {
				rv[i] = &types.InputSpecification{
					Name:     fmt.Sprintf("record%dInput%d", ri, i+1),
					TypeName: fmt.Sprintf("record%dInput%d.TypeName", ri, i+1),
					Source:   &types.InputSpecification_Hash{Hash: fmt.Sprintf("record%dInput%d.Source.Hash", ri, i+1)},
				}
			}
			return rv
		}
		makeRecordInputs := func(ri, count int) []types.RecordInput {
			rv := make([]types.RecordInput, count)
			for i := range rv {
				rv[i] = types.RecordInput{
					Name:     fmt.Sprintf("record%dInput%d", ri, i+1),
					Source:   &types.RecordInput_Hash{Hash: fmt.Sprintf("record%dInput%d.Source.Hash", ri, i+1)},
					TypeName: fmt.Sprintf("record%dInput%d.TypeName", ri, i+1),
					Status:   types.RecordInputStatus_Proposed,
				}
			}
			return rv
		}
		makeRecordOutputs := func(ri, count int) []types.RecordOutput {
			rv := make([]types.RecordOutput, count)
			for i := range rv {
				rv[i] = types.RecordOutput{
					Hash:   fmt.Sprintf("record%doutput%d", ri, i+1),
					Status: types.ResultStatus_RESULT_STATUS_PASS,
				}
			}
			return rv
		}

		owners := make([]user, ownerCount)
		for i := range owners {
			owners[i] = randomUser()
		}
		contractSpecs := make([]types.ContractSpecification, contractSpecCount)
		for i := range contractSpecs {
			contractSpecs[i] = types.ContractSpecification{
				SpecificationId: types.ContractSpecMetadataAddress(makeUUID(uint8(i))),
				Description:     nil,
				OwnerAddresses:  []string{owners[i].Bech32},
				PartiesInvolved: []types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
				Source:          types.NewContractSpecificationSourceHash(fmt.Sprintf("contractSpecs[%d].source.hash", i+1)),
				ClassName:       fmt.Sprintf("contractSpecs[%d]", i+1),
			}
		}
		recordSpecs := make([]types.RecordSpecification, RecordSpecCount)
		for i := range recordSpecs {
			name := fmt.Sprintf("record%d", i+1)
			recordSpecs[i] = types.RecordSpecification{
				SpecificationId:    contractSpecs[i/2].SpecificationId.MustGetAsRecordSpecAddress(name),
				Name:               name,
				Inputs:             makeRecordSpecInputs(i+1, 2),
				TypeName:           fmt.Sprintf("record%d.TypeName", i+1),
				ResultType:         types.DefinitionType_DEFINITION_TYPE_RECORD_LIST,
				ResponsibleParties: []types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
			}
		}
		scopeSpecs := make([]types.ScopeSpecification, ScopeSpecCount)
		for i := range scopeSpecs {
			scopeSpecs[i] = types.ScopeSpecification{
				SpecificationId: types.ScopeSpecMetadataAddress(makeUUID(uint8(i))),
				Description:     nil,
				OwnerAddresses:  []string{owners[i].Bech32},
				PartiesInvolved: []types.PartyType{types.PartyType_PARTY_TYPE_OWNER},
				ContractSpecIds: []types.MetadataAddress{contractSpecs[i*2].SpecificationId, contractSpecs[i*2+1].SpecificationId},
			}
		}
		scopes := make([]types.Scope, scopeCount)
		for i := range scopes {
			scopes[i] = types.Scope{
				ScopeId:           types.ScopeMetadataAddress(makeUUID(uint8(i))),
				SpecificationId:   scopeSpecs[i/2%ScopeSpecCount].SpecificationId,
				Owners:            ownerPartyList(owners[1+i%2].Bech32),
				DataAccess:        []string{owners[2-i%2].Bech32},
				ValueOwnerAddress: owners[0].Bech32,
			}
		}
		sessions := make([]types.Session, sessionCount)
		for i := range sessions {
			sessions[i] = types.Session{
				SessionId:       scopes[i/2].ScopeId.MustGetAsSessionAddress(makeUUID(uint8(i))),
				SpecificationId: contractSpecs[i%contractSpecCount].SpecificationId,
				Parties:         scopes[i/2].Owners,
				Name:            fmt.Sprintf("session[%d].Scope[%d]", i, i/2),
			}
		}
		records := make([]types.Record, recordCount)
		for i := range records {
			// Only used even numbered sessions.
			session := sessions[i/2*2]
			spec := recordSpecs[i%RecordSpecCount]
			records[i] = types.Record{
				Name:      spec.Name,
				SessionId: session.SessionId,
				Process: types.Process{
					ProcessId: &types.Process_Hash{Hash: fmt.Sprintf("record%d.Process.ProcessId.Hash", i+1)},
					Name:      spec.TypeName,
					Method:    fmt.Sprintf("record%d.Method", i+1),
				},
				Inputs:          makeRecordInputs(i+1, 2),
				Outputs:         makeRecordOutputs(i+1, 2),
				SpecificationId: spec.SpecificationId,
			}
		}

		for i, entry := range contractSpecs {
			bz, err := s.app.AppCodec().Marshal(&entry)
			require.NoError(t, err, "marshalling contract spec %d", i)
			s.store.Set(entry.SpecificationId, bz)
			keeper.IndexContractSpecBad(&s.store, &entry)
		}

		for i, entry := range recordSpecs {
			bz, err := s.app.AppCodec().Marshal(&entry)
			require.NoError(t, err, "marshalling record spec %d", i)
			s.store.Set(entry.SpecificationId, bz)
		}

		for i, entry := range scopeSpecs {
			bz, err := s.app.AppCodec().Marshal(&entry)
			require.NoError(t, err, "marshalling scope spec %d", i)
			s.store.Set(entry.SpecificationId, bz)
			keeper.IndexScopeSpecBad(&s.store, &entry)
		}

		for i, entry := range scopes {
			bz, err := s.app.AppCodec().Marshal(&entry)
			require.NoError(t, err, "marshalling scope %d", i)
			s.store.Set(entry.ScopeId, bz)
			keeper.IndexScopeBad(&s.store, &entry)
		}

		for i, entry := range sessions {
			bz, err := s.app.AppCodec().Marshal(&entry)
			require.NoError(t, err, "marshalling session %d", i)
			s.store.Set(entry.SessionId, bz)
		}

		for i, entry := range records {
			bz, err := s.app.AppCodec().Marshal(&entry)
			require.NoError(t, err, "marshalling record %d", i)
			s.store.Set(entry.GetRecordAddress(), bz)
		}

		migrator := keeper.NewMigrator(s.app.MetadataKeeper)
		require.NoError(t, migrator.Migrate2to3(s.ctx), "running migration v2 to v3")

		// The other parts have been tested in the other tests. This test just makes sure the migration runs
		// without error and with more data than the other tests (and doesn't deadlock).
		// The original migration process would deadlock while finding empty sessions
		// if there were enough entries in the store.
		// And "enough" wasn't that much. This unit test would pass with scopeCount := ScopeSpecCount * 2 * 7,
		// but it would deadlock with scopeCount := ScopeSpecCount * 2 * 8.
	})
}
