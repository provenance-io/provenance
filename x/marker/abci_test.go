package marker_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	piosimapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/marker"
	"github.com/provenance-io/provenance/x/marker/types"
)

func TestBeginBlocker(t *testing.T) {
	app := piosimapp.Setup(t)
	ctx := app.BaseApp.NewContext(false)

	testmint := &types.MarkerAccount{
		BaseAccount: &authtypes.BaseAccount{
			AccountNumber: 1,
			Address:       types.MustGetMarkerAddress("testmint").String(),
		},
		Status:      types.StatusActive,
		SupplyFixed: true,
		Denom:       "testmint",
		Supply:      sdkmath.NewInt(100),
	}

	app.MarkerKeeper.SetMarker(ctx, app.MarkerKeeper.NewMarker(ctx, testmint))

	// Initial supply of testmint must be zero.
	require.Equal(t, app.BankKeeper.GetSupply(ctx, "testmint").Amount, sdkmath.NewInt(0))

	marker.BeginBlocker(ctx, app.MarkerKeeper, app.BankKeeper)

	// Post begin block the supply must be 100
	require.Equal(t, app.BankKeeper.GetSupply(ctx, "testmint").Amount, sdkmath.NewInt(100))

	// Reset supply to a lower level
	testmint.Supply = sdkmath.NewInt(50)
	app.MarkerKeeper.SetMarker(ctx, testmint)

	marker.BeginBlocker(ctx, app.MarkerKeeper, app.BankKeeper)

	// Post begin block the supply must be 0
	require.Equal(t, app.BankKeeper.GetSupply(ctx, "testmint").Amount, sdkmath.NewInt(50))

	// Cancel marker and zero out supply
	testmint.Status = types.StatusDestroyed
	require.NoError(t, app.MarkerKeeper.AdjustCirculation(ctx, testmint, sdk.NewInt64Coin(testmint.Denom, 0)))
	app.MarkerKeeper.SetMarker(ctx, testmint)

	// Marker should still exist.
	notDeleted, err := app.MarkerKeeper.GetMarker(ctx, types.MustGetMarkerAddress("testmint"))
	require.NoError(t, err)
	require.NotNil(t, notDeleted)

	// Purges destroyed status markers
	marker.BeginBlocker(ctx, app.MarkerKeeper, app.BankKeeper)

	// Marker should no longer exist.
	deleted, err := app.MarkerKeeper.GetMarker(ctx, types.MustGetMarkerAddress("testmint"))
	require.NoError(t, err)
	require.Nil(t, deleted)
}
