package v042_test

import (
	"testing"

	cryptotypes "github.com/tendermint/tendermint/crypto"

	"github.com/tendermint/tendermint/crypto/secp256k1"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/provenance-io/provenance/app"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	v042 "github.com/provenance-io/provenance/x/marker/legacy/v042"
	"github.com/provenance-io/provenance/x/marker/types"
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

	markers []types.MarkerAccount
}

func TestMigrateTestSuite(t *testing.T) {
	suite.Run(t, new(MigrateTestSuite))
}

func (s *MigrateTestSuite) SetupTest() {
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
	markers := []types.MarkerAccount{
		{Denom: "nhash", MarkerType: types.MarkerType_Coin, AccessControl: []types.AccessGrant{{Address: s.user1, Permissions: types.AccessListByNames("mint,burn")}}},
		{Denom: "atom", MarkerType: types.MarkerType_Coin, AccessControl: []types.AccessGrant{{Address: s.user1, Permissions: types.AccessListByNames("mint,burn,transfer")}}},
		{Denom: "security", MarkerType: types.MarkerType_RestrictedCoin, AccessControl: []types.AccessGrant{{Address: s.user1, Permissions: types.AccessListByNames("mint,burn,transfer")}}},
	}
	s.markers = markers
	err := s.InitGenesisLegacy(ctx, app)
	s.Require().NoError(err)
}

// InitGenesisLegacy sets up the key store with legacy key format (< v042)
func (s *MigrateTestSuite) InitGenesisLegacy(ctx sdk.Context, app *app.App) error {
	store := ctx.KVStore(app.GetKey(types.ModuleName))
	for _, marker := range s.markers {
		accAddr, _ := types.MarkerAddress(marker.GetDenom())
		key := v042.MarkerStoreKeyLegacy(accAddr)
		store.Set(key, accAddr.Bytes())
	}

	return nil
}

func (s *MigrateTestSuite) TestMigrateMarkerAddressKeys() {
	err := v042.MigrateMarkerAddressKeys(s.ctx, s.app.GetKey(types.ModuleName))
	s.Assert().NoError(err)
	store := s.ctx.KVStore(s.app.GetKey(types.ModuleName))
	for _, marker := range s.markers {
		// Should have removed object store locator at legacy key
		acc, _ := types.MarkerAddress(marker.GetDenom())
		key := v042.MarkerStoreKeyLegacy(acc)
		result := store.Get(key)
		s.Assert().Nil(result)

		// Should find marker from updated key
		key = types.MarkerStoreKey(acc)
		s.Assert().Equal(types.MarkerStoreKeyPrefix, key[0:1])
		s.Assert().Equal([]byte{byte(20)}, key[1:2], "length prefix should be size of address")
		s.Assert().Equal(20, len(key[2:]))
		result = store.Get(key)
		s.Assert().NotNil(result)
		s.Assert().Equal(acc.Bytes(), result)
	}
}

func (s *MigrateTestSuite) TestMigrateMarkerPermissions() {
	k := MockKeeper{markers: s.markers}
	err := v042.MigrateMarkerPermissions(s.ctx, k)
	s.Require().NoError(err)

	for _, marker := range k.markers {
		m := &marker
		switch marker.MarkerType {
		case types.MarkerType_Coin:
			s.Assert().Len(m.AddressListForPermission(types.Access_Transfer), 0, "expect no addresses with transfer permission on coins")
		case types.MarkerType_RestrictedCoin:
			s.Assert().Len(m.AddressListForPermission(types.Access_Transfer), 1, "expect one address with transfer permission on coins")
		default:
			s.Require().Fail("unknown type")
		}
	}
}

type MockKeeper struct {
	markers []types.MarkerAccount
}

func (k MockKeeper) SetMarker(_ sdk.Context, m types.MarkerAccountI) {
	updated := m.Clone()
	for i, a := range k.markers {
		if a.Denom == updated.Denom {
			k.markers[i] = *updated
		} else {
			k.markers[i] = a
		}
	}
}

func (k MockKeeper) IterateMarkers(_ sdk.Context, step func(types.MarkerAccountI) bool) {
	for i := range k.markers {
		if step(&k.markers[i]) {
			break
		}
	}
}
