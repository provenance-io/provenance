package v042_test

import (
	"testing"

	"github.com/google/uuid"
	cryptotypes "github.com/tendermint/tendermint/crypto"

	"github.com/tendermint/tendermint/crypto/secp256k1"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/provenance-io/provenance/app"
	simapp "github.com/provenance-io/provenance/app"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	v042 "github.com/provenance-io/provenance/x/metadata/legacy/v042"
	"github.com/provenance-io/provenance/x/metadata/types"
)

type MigrateTestSuite struct {
	suite.Suite

	app *app.App
	ctx sdk.Context

	pubkey1   cryptotypes.PubKey
	user1     string
	user1Addr sdk.AccAddress

	pubkey2   cryptotypes.PubKey
	user2     string
	user2Addr sdk.AccAddress

	osLocators            []types.ObjectStoreLocator
	scopeMetaaddrs        []types.MetadataAddress
	scopeSpecMetaaddrs    []types.MetadataAddress
	contractSpecMetaaddrs []types.MetadataAddress
}

func TestMigrateTestSuite(t *testing.T) {
	suite.Run(t, new(MigrateTestSuite))
}

func (s *MigrateTestSuite) SetupTest() {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	s.app = app
	s.ctx = ctx

	s.pubkey1 = secp256k1.GenPrivKey().PubKey()
	s.user1Addr = sdk.AccAddress(s.pubkey1.Address())
	s.user1 = s.user1Addr.String()

	s.pubkey2 = secp256k1.GenPrivKey().PubKey()
	s.user2Addr = sdk.AccAddress(s.pubkey2.Address())
	s.user2 = s.user2Addr.String()

	osLocators := []types.ObjectStoreLocator{
		types.NewOSLocatorRecord(s.user1Addr, sdk.AccAddress("encryptionKey1"), "http://migration.test.user1.com"),
		types.NewOSLocatorRecord(s.user2Addr, sdk.AccAddress("encryptionKey2"), "http://migration.test.user2.com"),
	}
	s.osLocators = osLocators

	scopeMetadataAddrs := []types.MetadataAddress{
		types.ScopeMetadataAddress(uuid.New()),
		types.ScopeMetadataAddress(uuid.New()),
	}
	s.scopeMetaaddrs = scopeMetadataAddrs

	scopeSpecMetaaddrs := []types.MetadataAddress{
		types.ScopeSpecMetadataAddress(uuid.New()),
		types.ScopeSpecMetadataAddress(uuid.New()),
	}
	s.scopeSpecMetaaddrs = scopeSpecMetaaddrs

	contractSpecMetaaddrs := []types.MetadataAddress{
		types.ScopeSpecMetadataAddress(uuid.New()),
		types.ScopeSpecMetadataAddress(uuid.New()),
	}
	s.contractSpecMetaaddrs = contractSpecMetaaddrs

	var metadataData types.GenesisState
	metadataData.ObjectStoreLocators = append(metadataData.ObjectStoreLocators, osLocators...)
	err := s.InitGenesisLegacy(ctx, &metadataData, app)
	s.Require().NoError(err)
}

// InitGenesisLegacy sets up the key store with legacy key format (< v042)
func (s *MigrateTestSuite) InitGenesisLegacy(ctx sdk.Context, data *types.GenesisState, app *app.App) error {
	store := ctx.KVStore(app.GetKey(types.ModuleName))
	for _, locator := range data.ObjectStoreLocators {
		accAddr, _ := sdk.AccAddressFromBech32(locator.Owner)
		key := v042.GetOSLocatorKeyLegacy(accAddr)

		bz, err := types.ModuleCdc.Marshal(&locator)
		if err != nil {
			return err
		}
		store.Set(key, bz)
	}
	store.Set(v042.GetAddressScopeCacheKeyLegacy(s.user1Addr, s.scopeMetaaddrs[0]), []byte{0x01})
	store.Set(v042.GetAddressScopeCacheKeyLegacy(s.user2Addr, s.scopeMetaaddrs[1]), []byte{0x01})

	store.Set(v042.GetValueOwnerScopeCacheKeyLegacy(s.user1Addr, s.scopeMetaaddrs[0]), []byte{0x01})
	store.Set(v042.GetValueOwnerScopeCacheKeyLegacy(s.user2Addr, s.scopeMetaaddrs[1]), []byte{0x01})

	store.Set(v042.GetAddressScopeSpecCacheKeyLegacy(s.user1Addr, s.scopeSpecMetaaddrs[0]), []byte{0x01})
	store.Set(v042.GetAddressScopeSpecCacheKeyLegacy(s.user2Addr, s.scopeSpecMetaaddrs[1]), []byte{0x01})

	store.Set(v042.GetAddressContractSpecCacheKeyLegacy(s.user1Addr, s.contractSpecMetaaddrs[0]), []byte{0x01})
	store.Set(v042.GetAddressContractSpecCacheKeyLegacy(s.user2Addr, s.contractSpecMetaaddrs[1]), []byte{0x01})

	return nil
}

func (s *MigrateTestSuite) TestMigrateOSLocatorKeys() {
	err := v042.MigrateOSLocatorKeys(s.ctx, s.app.GetKey("metadata"))
	s.Assert().NoError(err)
	store := s.ctx.KVStore(s.app.GetKey(types.ModuleName))
	for _, locator := range s.osLocators {
		// Should have removed object store locator at legacy key
		acc, _ := sdk.AccAddressFromBech32(locator.Owner)
		key := v042.GetOSLocatorKeyLegacy(acc)
		result := store.Get(key)
		s.Assert().Nil(result)

		// Should find object store locator from updated key
		key = types.GetOSLocatorKey(acc)
		s.Assert().Equal(types.OSLocatorAddressKeyPrefix, key[0:1])
		s.Assert().Equal([]byte{byte(20)}, key[1:2], "length prefix should be size of address")
		s.Assert().Equal(20, len(key[2:]))
		result = store.Get(key)
		s.Assert().NotNil(result)
		var resultOSLocator types.ObjectStoreLocator
		err = types.ModuleCdc.Unmarshal(result, &resultOSLocator)
		s.Assert().NoError(err)
		s.Assert().Equal(locator, resultOSLocator)
	}
}

func (s *MigrateTestSuite) TestMigrateAddressScopeCacheKey() {
	err := v042.MigrateAddressScopeCacheKey(s.ctx, s.app.GetKey("metadata"))
	s.Assert().NoError(err)
	store := s.ctx.KVStore(s.app.GetKey(types.ModuleName))
	key := v042.GetAddressScopeCacheKeyLegacy(s.user1Addr, s.scopeMetaaddrs[0])
	result := store.Get(key)
	s.Assert().Nil(result)
	key = v042.GetAddressScopeCacheKeyLegacy(s.user2Addr, s.scopeMetaaddrs[1])
	result = store.Get(key)
	s.Assert().Nil(result)

	// Should find cache key with new v043 key
	key = types.GetAddressScopeCacheKey(s.user1Addr, s.scopeMetaaddrs[0])
	s.Assert().Equal(types.AddressScopeCacheKeyPrefix, key[0:1])
	s.Assert().Equal([]byte{byte(20)}, key[1:2], "length prefix should be size of address")
	s.Assert().Equal(s.user1Addr.Bytes(), key[2:22])
	s.Assert().Equal(s.scopeMetaaddrs[0].Bytes(), key[22:])
	result = store.Get(key)
	s.Assert().NotNil(result)
	key = types.GetAddressScopeCacheKey(s.user2Addr, s.scopeMetaaddrs[1])
	s.Assert().Equal(types.AddressScopeCacheKeyPrefix, key[0:1])
	s.Assert().Equal([]byte{byte(20)}, key[1:2], "length prefix should be size of address")
	s.Assert().Equal(s.user2Addr.Bytes(), key[2:22])
	s.Assert().Equal(s.scopeMetaaddrs[1].Bytes(), key[22:])
	result = store.Get(key)
	s.Assert().NotNil(result)
}

func (s *MigrateTestSuite) TestMigrateValueOwnerScopeCacheKey() {
	err := v042.MigrateValueOwnerScopeCacheKey(s.ctx, s.app.GetKey("metadata"))
	s.Assert().NoError(err)
	store := s.ctx.KVStore(s.app.GetKey(types.ModuleName))
	key := v042.GetValueOwnerScopeCacheKeyLegacy(s.user1Addr, s.scopeMetaaddrs[0])
	result := store.Get(key)
	s.Assert().Nil(result)
	key = v042.GetValueOwnerScopeCacheKeyLegacy(s.user2Addr, s.scopeMetaaddrs[1])
	result = store.Get(key)
	s.Assert().Nil(result)

	// Should find cache key with new v043 key
	key = types.GetValueOwnerScopeCacheKey(s.user1Addr, s.scopeMetaaddrs[0])
	s.Assert().Equal(types.ValueOwnerScopeCacheKeyPrefix, key[0:1])
	s.Assert().Equal([]byte{byte(20)}, key[1:2], "length prefix should be size of address")
	s.Assert().Equal(s.user1Addr.Bytes(), key[2:22])
	s.Assert().Equal(s.scopeMetaaddrs[0].Bytes(), key[22:])
	result = store.Get(key)
	s.Assert().NotNil(result)
	key = types.GetValueOwnerScopeCacheKey(s.user2Addr, s.scopeMetaaddrs[1])
	s.Assert().Equal(types.ValueOwnerScopeCacheKeyPrefix, key[0:1])
	s.Assert().Equal([]byte{byte(20)}, key[1:2], "length prefix should be size of address")
	s.Assert().Equal(s.user2Addr.Bytes(), key[2:22])
	s.Assert().Equal(s.scopeMetaaddrs[1].Bytes(), key[22:])
	result = store.Get(key)
	s.Assert().NotNil(result)
}

func (s *MigrateTestSuite) TestMigrateAddressScopeSpecCacheKey() {
	err := v042.MigrateAddressScopeSpecCacheKey(s.ctx, s.app.GetKey("metadata"))
	s.Assert().NoError(err)
	store := s.ctx.KVStore(s.app.GetKey(types.ModuleName))
	key := v042.GetAddressScopeSpecCacheKeyLegacy(s.user1Addr, s.scopeSpecMetaaddrs[0])
	result := store.Get(key)
	s.Assert().Nil(result)
	key = v042.GetAddressScopeSpecCacheKeyLegacy(s.user2Addr, s.scopeSpecMetaaddrs[1])
	result = store.Get(key)
	s.Assert().Nil(result)

	// Should find cache key with new v043 key
	key = types.GetAddressScopeSpecCacheKey(s.user1Addr, s.scopeSpecMetaaddrs[0])
	s.Assert().Equal(types.AddressScopeSpecCacheKeyPrefix, key[0:1])
	s.Assert().Equal([]byte{byte(20)}, key[1:2], "length prefix should be size of address")
	s.Assert().Equal(s.user1Addr.Bytes(), key[2:22])
	s.Assert().Equal(s.scopeSpecMetaaddrs[0].Bytes(), key[22:])
	result = store.Get(key)
	s.Assert().NotNil(result)
	key = types.GetAddressScopeSpecCacheKey(s.user2Addr, s.scopeSpecMetaaddrs[1])
	s.Assert().Equal(types.AddressScopeSpecCacheKeyPrefix, key[0:1])
	s.Assert().Equal([]byte{byte(20)}, key[1:2], "length prefix should be size of address")
	s.Assert().Equal(s.user2Addr.Bytes(), key[2:22])
	s.Assert().Equal(s.scopeSpecMetaaddrs[1].Bytes(), key[22:])
	result = store.Get(key)
	s.Assert().NotNil(result)
}

func (s *MigrateTestSuite) TestMigrateAddressContractSpecCacheKey() {
	err := v042.MigrateAddressContractSpecCacheKey(s.ctx, s.app.GetKey("metadata"))
	s.Assert().NoError(err)
	store := s.ctx.KVStore(s.app.GetKey(types.ModuleName))
	key := v042.GetAddressContractSpecCacheKeyLegacy(s.user1Addr, s.contractSpecMetaaddrs[0])
	result := store.Get(key)
	s.Assert().Nil(result)
	key = v042.GetAddressContractSpecCacheKeyLegacy(s.user2Addr, s.contractSpecMetaaddrs[1])
	result = store.Get(key)
	s.Assert().Nil(result)

	// Should find cache key with new v043 key
	key = types.GetAddressContractSpecCacheKey(s.user1Addr, s.contractSpecMetaaddrs[0])
	s.Assert().Equal(types.AddressContractSpecCacheKeyPrefix, key[0:1])
	s.Assert().Equal([]byte{byte(20)}, key[1:2], "length prefix should be size of address")
	s.Assert().Equal(s.user1Addr.Bytes(), key[2:22])
	s.Assert().Equal(s.contractSpecMetaaddrs[0].Bytes(), key[22:])
	result = store.Get(key)
	s.Assert().NotNil(result)
	key = types.GetAddressContractSpecCacheKey(s.user2Addr, s.contractSpecMetaaddrs[1])
	s.Assert().Equal(types.AddressContractSpecCacheKeyPrefix, key[0:1])
	s.Assert().Equal([]byte{byte(20)}, key[1:2], "length prefix should be size of address")
	s.Assert().Equal(s.user2Addr.Bytes(), key[2:22])
	s.Assert().Equal(s.contractSpecMetaaddrs[1].Bytes(), key[22:])
	result = store.Get(key)
	s.Assert().NotNil(result)
}
