package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	simapp "github.com/provenance-io/provenance/app"

	markerkeeper "github.com/provenance-io/provenance/x/marker/keeper"
	markertypes "github.com/provenance-io/provenance/x/marker/types"
)

func TestMarkerInvariant(t *testing.T) {
	//app, ctx := createTestApp(true)
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

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
	require.NoError(t, mac.SetSupply(sdk.NewCoin(mac.Denom, sdk.OneInt())))
	require.NoError(t, app.MarkerKeeper.AddMarkerAccount(ctx, mac))

	// Initial, invariant should pass
	_, isBroken := invariantChecks(ctx)
	require.False(t, isBroken)

	// Finalize a marker, should mint supply
	require.NoError(t, app.MarkerKeeper.FinalizeMarker(ctx, user, mac.GetDenom()))
	require.NoError(t, app.MarkerKeeper.ActivateMarker(ctx, user, mac.GetDenom()))

	// After finalize-activate, invariant should pass with newly minted supply
	_, isBroken = invariantChecks(ctx)
	require.False(t, isBroken)

	require.NoError(t, app.MarkerKeeper.MintCoin(ctx, user, sdk.NewCoin(mac.GetDenom(), sdk.NewInt(1000))))

	// expect pass after mint operation
	_, isBroken = invariantChecks(ctx)
	require.False(t, isBroken)

	require.NoError(t, app.MarkerKeeper.BurnCoin(ctx, user, sdk.NewCoin(mac.GetDenom(), sdk.NewInt(100))))

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
