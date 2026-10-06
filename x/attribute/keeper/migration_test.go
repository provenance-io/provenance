package keeper_test

import (
	"crypto/sha256"
	"encoding/binary"
	"strings"
	"testing"
	"time"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/suite"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/provenance-io/provenance/app"
	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/attribute/keeper"
	"github.com/provenance-io/provenance/x/attribute/types"
	nametypes "github.com/provenance-io/provenance/x/name/types"
)

type MigrationTestSuite struct {
	suite.Suite

	app *app.App
	ctx sdk.Context

	ownerPubKey cryptotypes.PubKey
	ownerAddr   sdk.AccAddress
	ownerBech32 string
}

func TestMigrationTestSuite(t *testing.T) {
	suite.Run(t, new(MigrationTestSuite))
}

func (s *MigrationTestSuite) SetupTest() {
	s.app = simapp.Setup(s.T())
	s.ctx = s.app.BaseApp.NewContextLegacy(false, cmtproto.Header{Time: time.Now()})

	s.ownerPubKey = secp256k1.GenPrivKey().PubKey()
	s.ownerAddr = sdk.AccAddress(s.ownerPubKey.Address())
	s.ownerBech32 = s.ownerAddr.String()

	// Bind the attribute name to the owner via the NameKeeper genesis,
	// same pattern used by KeeperTestSuite.SetupTest.
	var nameData nametypes.GenesisState
	nameData.Bindings = append(nameData.Bindings, nametypes.NewNameRecord("kyc", s.ownerAddr, false))
	nameData.Bindings = append(nameData.Bindings, nametypes.NewNameRecord("provenance.kyc", s.ownerAddr, false))
	nameData.Params.AllowUnrestrictedNames = false
	nameData.Params.MaxNameLevels = 3
	nameData.Params.MinSegmentLength = 3
	nameData.Params.MaxSegmentLength = 12
	s.app.NameKeeper.InitGenesis(s.ctx, nameData)

	// Materialize the owner account so SetAttribute's GetAccount check passes.
	s.app.AccountKeeper.SetAccount(
		s.ctx,
		s.app.AccountKeeper.NewAccountWithAddress(s.ctx, s.ownerAddr),
	)
}

// state written by the old binary must still be readable by the new one.
func (s *MigrationTestSuite) TestMigrate2to3_IsNoOp_AndPreservesState() {
	attr := types.Attribute{
		Name:          "provenance.kyc",
		Value:         []byte("verified"),
		AttributeType: types.AttributeType_String,
		Address:       s.ownerBech32,
	}
	s.Require().NoError(
		s.app.AttributeKeeper.SetAttribute(s.ctx, attr, s.ownerAddr),
		"SetAttribute must succeed during seeding",
	)

	pre, err := s.app.AttributeKeeper.GetAttributes(s.ctx, s.ownerBech32, attr.Name)
	s.Require().NoError(err)
	s.Require().Len(pre, 1, "expected one attribute before migration")

	m := keeper.NewMigrator(s.app.AttributeKeeper)
	s.Require().NoError(m.Migrate2to3(s.ctx), "Migrate2to3 must not error")

	post, err := s.app.AttributeKeeper.GetAttributes(s.ctx, s.ownerBech32, attr.Name)
	s.Require().NoError(err)
	s.Require().Len(post, 1, "expected one attribute after migration")
	s.Require().Equal(pre[0], post[0], "attribute contents must be byte-identical before/after migration")
}

// legacyNameKeyBytes reproduces the pre-migration GetNameKeyBytes: the name is normalized and
// reversed, then sha256'd into a fixed 32 bytes. The new code stores the reversed name itself,
// so this can't be built from the production helper any more.
func legacyNameKeyBytes(name string) []byte {
	comps := strings.Split(strings.ToLower(strings.TrimSpace(name)), ".")
	for i, j := 0, len(comps)-1; i < j; i, j = i+1, j-1 {
		comps[i], comps[j] = comps[j], comps[i]
	}
	sum := sha256.Sum256([]byte(strings.Join(comps, ".")))
	return sum[:]
}

// legacyAttrKey builds the pre-migration attribute key:
// [0x02][len(addr)][addr][sha256(reversed name)][sha256(value)].
func legacyAttrKey(attr types.Attribute) []byte {
	key := append([]byte(nil), types.AttributeKeyPrefix...)
	key = append(key, address.MustLengthPrefix(attr.GetAddressBytes())...)
	key = append(key, legacyNameKeyBytes(attr.Name)...)
	return append(key, attr.Hash()...)
}

// legacyNameAddrKey builds the pre-migration counter key:
// [0x03][sha256(reversed name)][len(addr)][addr].
func legacyNameAddrKey(name string, addrBz []byte) []byte {
	key := append([]byte(nil), types.AttributeAddrLookupKeyPrefix...)
	key = append(key, legacyNameKeyBytes(name)...)
	return append(key, address.MustLengthPrefix(addrBz)...)
}

// legacyExpireKey builds the pre-migration expiration key:
// [0x04][epoch][len(addr)][addr][sha256(reversed name)][sha256(value)].
func legacyExpireKey(attr types.Attribute) []byte {
	if attr.ExpirationDate == nil {
		return nil
	}
	key := append([]byte(nil), types.GetAttributeExpireTimePrefix(*attr.ExpirationDate)...)
	key = append(key, address.MustLengthPrefix(attr.GetAddressBytes())...)
	key = append(key, legacyNameKeyBytes(attr.Name)...)
	return append(key, attr.Hash()...)
}

// seedLegacy writes the attributes into the store using the old hashed-name layout, exactly as
// the pre-migration binary would have, bypassing the keeper entirely.
func (s *MigrationTestSuite) seedLegacy(attrs []types.Attribute) {
	store := s.ctx.KVStore(s.app.GetKey(types.StoreKey))
	counts := map[string]uint64{}

	for _, attr := range attrs {
		bz, err := s.app.AppCodec().Marshal(&attr)
		s.Require().NoError(err, "marshal %q", attr.Name)
		store.Set(legacyAttrKey(attr), bz)

		counts[attr.Name+"|"+attr.Address]++

		if attr.ExpirationDate != nil {
			store.Set(legacyExpireKey(attr), []byte{})
		}
	}

	for _, attr := range attrs {
		key := legacyNameAddrKey(attr.Name, attr.GetAddressBytes())
		val := make([]byte, 8)
		binary.BigEndian.PutUint64(val, counts[attr.Name+"|"+attr.Address])
		store.Set(key, val)
	}
}

// countPrefix returns how many store entries sit under the provided prefix.
func (s *MigrationTestSuite) countPrefix(prefix []byte) int {
	store := s.ctx.KVStore(s.app.GetKey(types.StoreKey))
	iter := storetypes.KVStorePrefixIterator(store, prefix)
	defer iter.Close() //nolint:errcheck // close error safe to ignore in a test.
	n := 0
	for ; iter.Valid(); iter.Next() {
		n++
	}
	return n
}

func (s *MigrationTestSuite) TestMigrate2to3_RewritesKeysAndPreservesRecords() {
	future := s.ctx.BlockTime().Add(24 * time.Hour)

	attrs := []types.Attribute{
		types.NewAttribute("provenance.kyc", s.ownerBech32, types.AttributeType_String, []byte("verified"), nil, ""),
		types.NewAttribute("provenance.kyc", s.ownerBech32, types.AttributeType_String, []byte("second-value"), &future, ""),
		types.NewAttribute("level.provenance.kyc", s.ownerBech32, types.AttributeType_String, []byte("gold"), nil, ""),
	}
	s.seedLegacy(attrs)

	// Sanity: the legacy layout is in place and the new code can't see any of it yet.
	s.Require().Equal(3, s.countPrefix(types.AttributeKeyPrefix), "legacy attribute entries before migration")
	s.Require().Equal(1, s.countPrefix(types.AttributeExpirationKeyPrefix), "legacy expiration entries before migration")

	m := keeper.NewMigrator(s.app.AttributeKeeper)
	s.Require().NoError(m.Migrate2to3(s.ctx), "Migrate2to3")

	// Every record should still be there, now readable through the new key layout.
	got, err := s.app.AttributeKeeper.GetAllAttributesAddr(s.ctx, s.ownerAddr)
	s.Require().NoError(err, "GetAllAttributesAddr after migration")
	s.Require().Len(got, 3, "attributes after migration")
	s.Assert().ElementsMatch(attrs, got, "attribute records must survive the migration unchanged")

	// Nothing should remain at any legacy key.
	store := s.ctx.KVStore(s.app.GetKey(types.StoreKey))
	for _, attr := range attrs {
		s.Assert().Nil(store.Get(legacyAttrKey(attr)), "legacy attribute key for %q/%q", attr.Name, string(attr.Value))
		s.Assert().Nil(store.Get(legacyNameAddrKey(attr.Name, attr.GetAddressBytes())), "legacy counter key for %q", attr.Name)
		if attr.ExpirationDate != nil {
			s.Assert().Nil(store.Get(legacyExpireKey(attr)), "legacy expiration key for %q", attr.Name)
		}
	}

	// ...and the entry counts should match, i.e. nothing was left behind or duplicated.
	s.Assert().Equal(3, s.countPrefix(types.AttributeKeyPrefix), "attribute entries after migration")
	s.Assert().Equal(2, s.countPrefix(types.AttributeAddrLookupKeyPrefix), "counter entries after migration (one per name/addr)")
	s.Assert().Equal(1, s.countPrefix(types.AttributeExpirationKeyPrefix), "expiration entries after migration")
}

func (s *MigrationTestSuite) TestMigrate2to3_RebuildsNameAddrCounts() {
	attrs := []types.Attribute{
		types.NewAttribute("provenance.kyc", s.ownerBech32, types.AttributeType_String, []byte("a"), nil, ""),
		types.NewAttribute("provenance.kyc", s.ownerBech32, types.AttributeType_String, []byte("b"), nil, ""),
		types.NewAttribute("level.provenance.kyc", s.ownerBech32, types.AttributeType_String, []byte("c"), nil, ""),
	}
	s.seedLegacy(attrs)

	m := keeper.NewMigrator(s.app.AttributeKeeper)
	s.Require().NoError(m.Migrate2to3(s.ctx), "Migrate2to3")

	store := s.ctx.KVStore(s.app.GetKey(types.StoreKey))
	for name, want := range map[string]uint64{"provenance.kyc": 2, "level.provenance.kyc": 1} {
		bz := store.Get(types.AttributeNameAddrKeyPrefix(name, s.ownerAddr))
		s.Require().NotNil(bz, "counter for %q should exist after migration", name)
		s.Assert().Equal(want, binary.BigEndian.Uint64(bz), "counter for %q", name)
	}

	// The counter is what AccountsByAttribute reads, so it should resolve too.
	accts, err := s.app.AttributeKeeper.AccountsByAttribute(s.ctx, "provenance.kyc")
	s.Require().NoError(err, "AccountsByAttribute")
	s.Require().Len(accts, 1, "accounts holding provenance.kyc")
	s.Assert().Equal(s.ownerAddr, accts[0], "account holding provenance.kyc")
}

func (s *MigrationTestSuite) TestMigrate2to3_RebuildsExpirationIndex() {
	past := s.ctx.BlockTime().Add(-2 * time.Hour)
	future := s.ctx.BlockTime().Add(2 * time.Hour)

	expired := types.NewAttribute("provenance.kyc", s.ownerBech32, types.AttributeType_String, []byte("stale"), &past, "")
	live := types.NewAttribute("provenance.kyc", s.ownerBech32, types.AttributeType_String, []byte("fresh"), &future, "")
	s.seedLegacy([]types.Attribute{expired, live})

	m := keeper.NewMigrator(s.app.AttributeKeeper)
	s.Require().NoError(m.Migrate2to3(s.ctx), "Migrate2to3")

	// Both expiration entries should have been rebuilt under the new key layout.
	store := s.ctx.KVStore(s.app.GetKey(types.StoreKey))
	s.Require().True(store.Has(types.AttributeExpireKey(expired)), "rebuilt expiration entry for the expired attribute")
	s.Require().True(store.Has(types.AttributeExpireKey(live)), "rebuilt expiration entry for the live attribute")

	// And the rebuilt index must actually drive the sweep: only the past-dated one goes.
	s.app.AttributeKeeper.DeleteExpiredAttributes(s.ctx, 0)

	got, err := s.app.AttributeKeeper.GetAllAttributesAddr(s.ctx, s.ownerAddr)
	s.Require().NoError(err, "GetAllAttributesAddr after the sweep")
	s.Require().Len(got, 1, "attributes left after the sweep")
	s.Assert().Equal(live, got[0], "the non-expired attribute should survive")
}

func (s *MigrationTestSuite) TestMigrate2to3_EmptyStore() {
	m := keeper.NewMigrator(s.app.AttributeKeeper)
	s.Require().NoError(m.Migrate2to3(s.ctx), "Migrate2to3 must handle empty state without error")
	s.Assert().Equal(0, s.countPrefix(types.AttributeKeyPrefix), "attribute entries")
}
