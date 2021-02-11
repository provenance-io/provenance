package keeper_test

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	simapp "github.com/provenance-io/provenance/app"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/require"

	markertypes "github.com/provenance-io/provenance/x/marker/types"
	"github.com/provenance-io/provenance/x/metadata/types"
)

func TestMetadataScopeGetSet(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	scopeUUID := uuid.New()
	pubkey := secp256k1.GenPrivKey().PubKey()
	user := sdk.AccAddress(pubkey.Address())

	scopeID := types.ScopeMetadataAddress(scopeUUID)

	s, found := app.MetadataKeeper.GetScope(ctx, scopeID)
	require.NotNil(t, s)
	require.False(t, found)

	ns := *types.NewScope(scopeID, nil, []string{user.String()}, []string{user.String()}, "")
	require.NotNil(t, ns)
	app.MetadataKeeper.SetScope(ctx, ns)

	s, found = app.MetadataKeeper.GetScope(ctx, scopeID)
	require.True(t, found)
	require.NotNil(t, s)

}

func TestValidateUpdate(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	pubkey := secp256k1.GenPrivKey().PubKey()
	user := sdk.AccAddress(pubkey.Address().Bytes())

	pubkey2 := secp256k1.GenPrivKey().PubKey()
	user2 := sdk.AccAddress(pubkey2.Address().Bytes())

	markerAddr := markertypes.MustGetMarkerAddress("testcoin").String()
	err := app.MarkerKeeper.AddMarkerAccount(ctx, &markertypes.MarkerAccount{
		BaseAccount: &authtypes.BaseAccount{
			Address:       markerAddr,
			AccountNumber: 23,
		},
		AccessControl: []markertypes.AccessGrant{
			{
				Address:     user.String(),
				Permissions: markertypes.AccessListByNames("deposit,withdraw"),
			},
		},
		Denom:      "testcoin",
		Supply:     sdk.NewInt(1000),
		MarkerType: markertypes.MarkerType_Coin,
		Status:     markertypes.StatusActive,
	})
	require.NoError(t, err)
	app.AccountKeeper.SetAccount(ctx, app.AccountKeeper.NewAccountWithAddress(ctx, user))

	scopeUUID := uuid.New()
	scopeID := types.ScopeMetadataAddress(scopeUUID)
	err = app.MetadataKeeper.ValidateUpdate(
		ctx,
		types.Scope{},
		types.Scope{},
		[]string{})
	require.Error(t, err)
	require.Equal(t, "incorrect address length (must be at least 17, actual: 0)", err.Error())

	err = app.MetadataKeeper.ValidateUpdate(
		ctx,
		types.Scope{},
		*types.NewScope(scopeID, nil, []string{user.String()}, []string{}, ""),
		[]string{})
	require.NoError(t, err)

	changedID := types.ScopeMetadataAddress(uuid.New())
	err = app.MetadataKeeper.ValidateUpdate(
		ctx,
		*types.NewScope(scopeID, nil, []string{user.String()}, []string{}, ""),
		*types.NewScope(changedID, nil, []string{user.String()}, []string{}, ""),
		[]string{})
	require.Error(t, err, "can't change scope id in update")
	require.Equal(t, fmt.Sprintf("cannot update scope identifier. expected %s, got %s", scopeID.String(), changedID.String()), err.Error())

	err = app.MetadataKeeper.ValidateUpdate(
		ctx,
		*types.NewScope(scopeID, nil, []string{user.String()}, []string{}, ""),
		*types.NewScope(scopeID, types.ScopeSpecMetadataAddress(scopeUUID), []string{user.String()}, []string{}, ""),
		[]string{})
	require.Error(t, err)
	require.Equal(t, fmt.Sprintf("missing signature from existing owner %s; required for update", user.String()), err.Error())

	err = app.MetadataKeeper.ValidateUpdate(
		ctx,
		*types.NewScope(scopeID, nil, []string{user.String()}, []string{}, ""),
		*types.NewScope(scopeID, types.ScopeSpecMetadataAddress(scopeUUID), []string{user.String()}, []string{}, ""),
		[]string{user.String()})
	require.NoError(t, err, "no error when adding a scope spec and signed by current data owner")

	err = app.MetadataKeeper.ValidateUpdate(
		ctx,
		*types.NewScope(scopeID, nil, []string{user.String()}, []string{}, ""),
		*types.NewScope(scopeID, types.ScopeSpecMetadataAddress(scopeUUID), []string{user.String()}, []string{}, user.String()),
		[]string{user.String()})
	require.NoError(t, err, "setting a value owner should not error")

	err = app.MetadataKeeper.ValidateUpdate(
		ctx,
		*types.NewScope(scopeID, nil, []string{user.String()}, []string{}, user.String()),
		*types.NewScope(scopeID, types.ScopeSpecMetadataAddress(scopeUUID), []string{user.String()}, []string{}, user.String()),
		[]string{user.String()})
	require.NoError(t, err, "no change to value owner should not error")

	err = app.MetadataKeeper.ValidateUpdate(
		ctx,
		*types.NewScope(scopeID, nil, []string{user.String()}, []string{}, markerAddr),
		*types.NewScope(scopeID, types.ScopeSpecMetadataAddress(scopeUUID), []string{user.String()}, []string{}, user.String()),
		[]string{user.String()})
	require.NoError(t, err, "setting a new value owner should not error with withdraw permission")

	err = app.MetadataKeeper.ValidateUpdate(
		ctx,
		*types.NewScope(scopeID, nil, []string{user2.String()}, []string{}, markerAddr),
		*types.NewScope(scopeID, types.ScopeSpecMetadataAddress(scopeUUID), []string{user2.String()}, []string{}, user2.String()),
		[]string{user2.String()})
	require.Error(t, err, "setting a new value owner fails if missing withdraw permission")
	require.Equal(t, fmt.Sprintf("missing signature for %s with authority to withdraw/remove existing value owner", markerAddr), err.Error())

	err = app.MetadataKeeper.ValidateUpdate(
		ctx,
		*types.NewScope(scopeID, nil, []string{user2.String()}, []string{}, ""),
		*types.NewScope(scopeID, types.ScopeSpecMetadataAddress(scopeUUID), []string{user2.String()}, []string{}, markerAddr),
		[]string{user2.String()})
	require.Error(t, err, "setting a new value owner fails if missing deposit permission")
	require.Equal(t, fmt.Sprintf("no signatures present with authority to add scope to marker %s", markerAddr), err.Error())

	err = app.MetadataKeeper.ValidateUpdate(
		ctx,
		*types.NewScope(scopeID, nil, []string{user.String()}, []string{}, user2.String()),
		*types.NewScope(scopeID, types.ScopeSpecMetadataAddress(scopeUUID), []string{user.String()}, []string{}, user.String()),
		[]string{user.String()})
	require.Error(t, err, "setting a new value owner fails for scope owner when value owner signature is missing")
	require.Equal(t, fmt.Sprintf("missing signature from existing owner %s; required for update", user2.String()), err.Error())
}
