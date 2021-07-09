package marker_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/marker"
	"github.com/provenance-io/provenance/x/marker/types"
)

func TestBeginBlocker(t *testing.T) {
	app := app.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	testmint := &types.MarkerAccount{
		BaseAccount: &authtypes.BaseAccount{
			AccountNumber: 1,
			Address:       types.MustGetMarkerAddress("testmint").String(),
		},
		Status:      types.StatusActive,
		SupplyFixed: true,
		Denom:       "testmint",
		Supply:      sdk.NewInt(100),
	}

	app.MarkerKeeper.SetMarker(ctx, testmint)

	// Initial supply of testmint must be zero.
	require.Equal(t, app.BankKeeper.GetSupply(ctx, "testmint").Amount, sdk.NewInt(0))

	marker.BeginBlocker(ctx, abci.RequestBeginBlock{}, app.MarkerKeeper, app.BankKeeper)

	// Post begin block the supply must be 100
	require.Equal(t, app.BankKeeper.GetSupply(ctx, "testmint").Amount, sdk.NewInt(100))

	// Reset supply to a lower level
	testmint.Supply = sdk.NewInt(50)
	app.MarkerKeeper.SetMarker(ctx, testmint)

	marker.BeginBlocker(ctx, abci.RequestBeginBlock{}, app.MarkerKeeper, app.BankKeeper)

	// Post begin block the supply must be 0
	require.Equal(t, app.BankKeeper.GetSupply(ctx, "testmint").Amount, sdk.NewInt(50))

	// Cancel marker and zero out supply
	testmint.Status = types.StatusDestroyed
	require.NoError(t, app.MarkerKeeper.AdjustCirculation(ctx, testmint, sdk.NewCoin(testmint.Denom, sdk.ZeroInt())))
	app.MarkerKeeper.SetMarker(ctx, testmint)

	// Marker should still exist.
	notDeleted, err := app.MarkerKeeper.GetMarker(ctx, types.MustGetMarkerAddress("testmint"))
	require.NoError(t, err)
	require.NotNil(t, notDeleted)

	// Purges destroyed status markers
	marker.BeginBlocker(ctx, abci.RequestBeginBlock{}, app.MarkerKeeper, app.BankKeeper)

	// Marker should no longer exist.
	deleted, err := app.MarkerKeeper.GetMarker(ctx, types.MustGetMarkerAddress("testmint"))
	require.NoError(t, err)
	require.Nil(t, deleted)
}
