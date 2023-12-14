package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	simapp "github.com/provenance-io/provenance/app"
	markerkeeper "github.com/provenance-io/provenance/x/marker/keeper"
	"github.com/provenance-io/provenance/x/marker/types"
	markertypes "github.com/provenance-io/provenance/x/marker/types"
)

func TestMarkerInvariant(t *testing.T) {
	//app, ctx := createTestApp(true)
	app := simapp.Setup(t)
	ctx := app.BaseApp.NewContext(false)

	app.MarkerKeeper.SetParams(ctx, markertypes.DefaultParams())
	user := testUserAddress("test")

	// Get a reference to our invariant checks
	invariantChecks := markerkeeper.AllInvariants(app.MarkerKeeper, app.BankKeeper)
	require.NotNil(t, invariantChecks)

	// create account and check default values
	mac := markertypes.NewEmptyMarkerAccount("testcoin", user.String(),
		[]markertypes.AccessGrant{
			*markertypes.NewAccessGrant(
				user, []markertypes.Access{markertypes.Access_Burn, markertypes.Access_Mint, markertypes.Access_Withdraw}),
			*markertypes.NewAccessGrant(user, []markertypes.Access{markertypes.Access_Admin}),
		})

	require.NoError(t, mac.SetManager(user))
	require.NoError(t, mac.SetSupply(sdk.NewInt64Coin(mac.Denom, 1)))
	require.NoError(t, app.MarkerKeeper.AddMarkerAccount(ctx, mac))
	require.NoError(t, app.MarkerKeeper.SetNetAssetValue(ctx, mac, types.NewNetAssetValue(sdk.NewInt64Coin(types.UsdDenom, 1), 1), "test"))

	// Initial, invariant should pass
	_, isBroken := invariantChecks(ctx)
	require.False(t, isBroken)

	// Finalize a marker, should mint supply
	require.NoError(t, app.MarkerKeeper.FinalizeMarker(ctx, user, mac.GetDenom()))
	require.NoError(t, app.MarkerKeeper.ActivateMarker(ctx, user, mac.GetDenom()))

	// After finalize-activate, invariant should pass with newly minted supply
	_, isBroken = invariantChecks(ctx)
	require.False(t, isBroken)

	require.NoError(t, app.MarkerKeeper.MintCoin(ctx, user, sdk.NewInt64Coin(mac.GetDenom(), 1000)))

	// expect pass after mint operation
	_, isBroken = invariantChecks(ctx)
	require.False(t, isBroken)

	require.NoError(t, app.MarkerKeeper.BurnCoin(ctx, user, sdk.NewInt64Coin(mac.GetDenom(), 100)))

	// expect pass after burn operation
	_, isBroken = invariantChecks(ctx)
	require.False(t, isBroken)

	// move coin out of the marker and into a user account (recipient is empty, should go to admin)
	require.NoError(t, app.MarkerKeeper.WithdrawCoins(
		ctx, user, sdk.AccAddress{}, mac.GetDenom(), sdk.NewCoins(sdk.NewInt64Coin(mac.GetDenom(), 50))))

	// expect pass after withdraw operation
	_, isBroken = invariantChecks(ctx)
	require.False(t, isBroken)
}
