package keeper_test

import (
	"fmt"
	"testing"

	simapp "github.com/provenance-io/provenance/app"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	markerkeeper "github.com/provenance-io/provenance/x/marker/keeper"
	markertypes "github.com/provenance-io/provenance/x/marker/types"
)

func TestNewQuerier(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	cdc := app.LegacyAmino()
	user := testUserAddress("test")

	query := abci.RequestQuery{
		Path: "",
		Data: []byte{},
	}

	querier := markerkeeper.NewQuerier(app.MarkerKeeper, app.LegacyAmino())

	// should error for a route pointing to some other query module
	bz, err := querier(ctx, []string{"other"}, query)
	require.Error(t, err)
	require.Nil(t, bz)

	allMarkersParama := markertypes.NewQueryMarkersParams(1, 20, "testcoin", "")
	bz, errRes := cdc.MarshalJSON(allMarkersParama)
	require.Nil(t, errRes)

	query.Path = fmt.Sprintf("/custom/%s/%s", markertypes.QuerierRoute, markertypes.QueryMarkers)
	query.Data = bz

	// get all markers should not return an error even if there are no markers
	_, err = querier(ctx, []string{markertypes.QueryMarkers}, query)
	require.Nil(t, err)

	// get all holders should not return an error even if there are no of type
	_, err = querier(ctx, []string{markertypes.QueryHolders}, query)
	require.Nil(t, err)

	// create a marker account
	mac := markertypes.NewEmptyMarkerAccount("testcoin", user.String(), []markertypes.AccessGrant{*markertypes.NewAccessGrant(user,
		[]markertypes.Access{markertypes.Access_Mint})})
	require.NoError(t, app.MarkerKeeper.AddMarkerAccount(ctx, mac))

	// get all markers should not return an error
	_, err = querier(ctx, []string{markertypes.QueryMarkers}, query)
	require.Nil(t, err)

	// get all holders should not return an error
	_, err = querier(ctx, []string{markertypes.QueryHolders}, query)
	require.Nil(t, err)

	// pull marker using denom on QueryMarker endpoint
	query.Path = fmt.Sprintf("/custom/%s/%s/%s", markertypes.QuerierRoute, markertypes.QueryMarker, "testcoin")
	query.Data = []byte{}

	_, err = querier(ctx, []string{markertypes.QueryMarker, "testcoin"}, query)
	require.Nil(t, err)

	// pull marker using addess on QueryMarker endpoint
	query.Path = fmt.Sprintf("/custom/%s/%s/%s", markertypes.QuerierRoute, markertypes.QueryMarker,
		markertypes.MustGetMarkerAddress("testcoin"))
	query.Data = []byte{}

	_, err = querier(ctx, []string{markertypes.QueryMarker, markertypes.MustGetMarkerAddress("testcoin").String()},
		query)
	require.Nil(t, err)

	// pull marker using invalid denom on QueryMarker endpoint
	query.Path = fmt.Sprintf("/custom/%s/%s/%s", markertypes.QuerierRoute, markertypes.QueryMarker, "other")
	query.Data = []byte{}

	_, err = querier(ctx, []string{markertypes.QueryMarker, "other"}, query)
	require.Error(t, err)
}

func TestQuerierAccess(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	user := testUserAddress("test")
	// create a marker account
	mac := markertypes.NewEmptyMarkerAccount("testcoin", user.String(), []markertypes.AccessGrant{*markertypes.NewAccessGrant(user,
		[]markertypes.Access{markertypes.Access_Mint})})
	require.NoError(t, app.MarkerKeeper.AddMarkerAccount(ctx, mac))

	query := abci.RequestQuery{
		Path: "",
		Data: []byte{},
	}

	querier := markerkeeper.NewQuerier(app.MarkerKeeper, app.LegacyAmino())

	// pull marker access list
	query.Path = fmt.Sprintf("/custom/%s/%s/%s", markertypes.QuerierRoute, markertypes.QueryMarkerAccess, "testcoin")
	query.Data = []byte{}

	_, err := querier(ctx, []string{markertypes.QueryMarkerAccess, "testcoin"}, query)
	require.Nil(t, err)
}

func TestQuerierCoins(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	user := testUserAddress("test")
	// create a marker account
	mac := markertypes.NewEmptyMarkerAccount("testcoin", user.String(), []markertypes.AccessGrant{*markertypes.NewAccessGrant(user,
		[]markertypes.Access{markertypes.Access_Mint})})
	require.NoError(t, app.MarkerKeeper.AddMarkerAccount(ctx, mac))

	query := abci.RequestQuery{
		Path: "",
		Data: []byte{},
	}

	querier := markerkeeper.NewQuerier(app.MarkerKeeper, app.LegacyAmino())

	// pull marker account coins in escrow
	query.Path = fmt.Sprintf("/custom/%s/%s/%s", markertypes.QuerierRoute, markertypes.QueryMarkerEscrow, "testcoin")
	query.Data = []byte{}

	_, err := querier(ctx, []string{markertypes.QueryMarkerEscrow, "testcoin"}, query)
	require.Nil(t, err)
}

func TestQuerierSupply(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	user := testUserAddress("test")
	// create a marker account
	mac := markertypes.NewEmptyMarkerAccount("testcoin", user.String(), []markertypes.AccessGrant{*markertypes.NewAccessGrant(user,
		[]markertypes.Access{markertypes.Access_Mint})})
	require.NoError(t, app.MarkerKeeper.AddMarkerAccount(ctx, mac))

	query := abci.RequestQuery{
		Path: "",
		Data: []byte{},
	}

	querier := markerkeeper.NewQuerier(app.MarkerKeeper, app.LegacyAmino())

	// pull marker account total supply
	query.Path = fmt.Sprintf("/custom/%s/%s/%s", markertypes.QuerierRoute, markertypes.QueryMarkerSupply, "testcoin")
	query.Data = []byte{}

	_, err := querier(ctx, []string{markertypes.QueryMarkerSupply, "testcoin"}, query)
	require.Nil(t, err)
}
