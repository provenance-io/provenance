package keeper_test

import (
	"testing"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/marker/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ByteCompatTestSuite proves the collections-based keeper writes to the EXACT
// legacy KVStore keys/values, so existing mainnet state is read in place.
type ByteCompatTestSuite struct {
	suite.Suite
	app *simapp.App
	ctx sdk.Context
}

func TestByteCompatTestSuite(t *testing.T) {
	suite.Run(t, new(ByteCompatTestSuite))
}

func (s *ByteCompatTestSuite) SetupTest() {
	s.app = simapp.Setup(s.T())
	s.ctx = s.app.BaseApp.NewContext(false)
}

// rawStore returns the marker module's raw KVStore, bypassing the keeper entirely.
func (s *ByteCompatTestSuite) rawStore() storetypes.KVStore {
	return s.ctx.KVStore(s.app.GetKey(types.StoreKey))
}

// TestParamsAtLegacyKey: params must live at exactly key 0x05 with cdc.Marshal bytes.
func (s *ByteCompatTestSuite) TestParamsAtLegacyKey() {
	params := types.DefaultParams()
	s.app.MarkerKeeper.SetParams(s.ctx, params)

	raw := s.rawStore().Get(types.MarkerParamStoreKey) // 0x05
	s.Require().NotNil(raw, "params must exist at legacy key 0x05")

	want, err := s.app.AppCodec().Marshal(&params)
	s.Require().NoError(err)
	s.Assert().Equal(want, raw, "params value must be byte-identical to legacy proto marshal")
}

// TestMarkerIndexAtLegacyKey: the marker-address index must live at
// 0x02 + MustLengthPrefix(addr) — exactly types.MarkerStoreKey(addr).
func (s *ByteCompatTestSuite) TestMarkerIndexAtLegacyKey() {
	denom := "bytecompatcoin"
	addr := types.MustGetMarkerAddress(denom)
	manager := sdk.AccAddress("manager_address_____") // 20 bytes, distinct from the marker

	m := types.NewEmptyMarkerAccount(denom, manager.String(), nil)
	s.Require().NoError(s.app.MarkerKeeper.AddMarkerAccount(s.ctx, m), "AddMarkerAccount")

	s.Assert().True(s.rawStore().Has(types.MarkerStoreKey(addr)),
		"marker index must exist at legacy key 0x02+LP(addr)")
}

// TestNetAssetValueAtLegacyKey: NAV must live at 0x04 + LP(markerAddr) + denom
// with a proto-marshaled NetAssetValue as the value.
func (s *ByteCompatTestSuite) TestNetAssetValueAtLegacyKey() {
	denom := "navcompatcoin"
	addr := types.MustGetMarkerAddress(denom)
	manager := sdk.AccAddress("manager_address_____")

	m := types.NewEmptyMarkerAccount(denom, manager.String(), nil)
	s.Require().NoError(s.app.MarkerKeeper.AddMarkerAccount(s.ctx, m), "AddMarkerAccount")

	nav := types.NewNetAssetValue(sdk.NewInt64Coin("usd", 1000), 1)
	s.Require().NoError(s.app.MarkerKeeper.SetNetAssetValue(s.ctx, m, nav, "test"), "SetNetAssetValue")

	raw := s.rawStore().Get(types.NetAssetValueKey(addr, "usd"))
	s.Require().NotNil(raw, "NAV must exist at legacy key 0x04+LP(addr)+denom")

	var got types.NetAssetValue
	s.Require().NoError(s.app.AppCodec().Unmarshal(raw, &got), "NAV bytes must proto-unmarshal")
	s.Assert().Equal(nav.Price, got.Price, "NAV price")
	s.Assert().Equal(nav.Volume, got.Volume, "NAV volume")
}

// TestDenySendAtLegacyKey: deny-send entries must live at
// 0x03 + LP(markerAddr) + LP(deniedAddr), and be removed from that exact key.
func (s *ByteCompatTestSuite) TestDenySendAtLegacyKey() {
	denom := "denycompatcoin"
	markerAddr := types.MustGetMarkerAddress(denom)
	denied := sdk.AccAddress("denied_address______") // 20 bytes

	// Write through the keeper (collections path).
	s.app.MarkerKeeper.AddSendDeny(s.ctx, markerAddr, denied)

	// The entry must exist at the exact legacy key bytes.
	s.Assert().True(s.rawStore().Has(types.DenySendKey(markerAddr, denied)),
		"deny-send entry must exist at legacy key 0x03+LP(marker)+LP(denied)")

	// Behavior parity: the keeper's own read path must agree.
	s.Assert().True(s.app.MarkerKeeper.IsSendDeny(s.ctx, markerAddr, denied),
		"IsSendDeny must report true after AddSendDeny")

	// Remove and confirm it's gone from the legacy key.
	s.app.MarkerKeeper.RemoveSendDeny(s.ctx, markerAddr, denied)
	s.Assert().False(s.rawStore().Has(types.DenySendKey(markerAddr, denied)),
		"deny-send entry must be deleted from the legacy key")
	s.Assert().False(s.app.MarkerKeeper.IsSendDeny(s.ctx, markerAddr, denied),
		"IsSendDeny must report false after RemoveSendDeny")
}

// TestGenesisRoundTrip verifies that exporting and importing genesis
// state preserves the same data.
func TestGenesisRoundTrip(t *testing.T) {
	app := simapp.Setup(t)
	ctx := app.BaseApp.NewContext(false)

	denom := "roundtripcoin"
	manager := sdk.AccAddress("manager_address_____")
	m := types.NewEmptyMarkerAccount(denom, manager.String(), nil)
	require.NoError(t, app.MarkerKeeper.AddMarkerAccount(ctx, m), "AddMarkerAccount")
	require.NoError(t, app.MarkerKeeper.SetNetAssetValue(ctx, m,
		types.NewNetAssetValue(sdk.NewInt64Coin("usd", 5), 1), "test"), "SetNetAssetValue")

	gen1 := app.MarkerKeeper.ExportGenesis(ctx)

	app2 := simapp.Setup(t)
	ctx2 := app2.BaseApp.NewContext(false)
	app2.MarkerKeeper.InitGenesis(ctx2, gen1)

	gen2 := app2.MarkerKeeper.ExportGenesis(ctx2)
	require.Equal(t, gen1, gen2, "export -> import -> export must be identical")
}
