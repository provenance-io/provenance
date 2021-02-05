package keeper_test

import (
	"testing"

	"github.com/google/uuid"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	simapp "github.com/provenance-io/provenance/app"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/stretchr/testify/require"

	"github.com/provenance-io/provenance/x/metadata/keeper"
	"github.com/provenance-io/provenance/x/metadata/types"
)

func TestMetadataScopeGetSet(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	scopeUUID := uuid.New()
	pubkey := secp256k1.GenPrivKey().PubKey()
	user := sdk.AccAddress(pubkey.Address())

	scopeID := types.ScopeMetadataAddress(scopeUUID)

	k := keeper.NewKeeper(app.AppCodec(), app.GetKey(types.StoreKey), app.GetSubspace(types.ModuleName), app.AccountKeeper)

	s, found := k.GetScope(ctx, scopeID)
	require.Nil(t, s)
	require.False(t, found)

	ns := *types.NewScope(scopeID, nil, []string{user.String()}, []string{user.String()}, "")
	require.NotNil(t, ns)
	k.SetScope(ctx, ns)

	s, found = k.GetScope(ctx, scopeID)
	require.True(t, found)
	require.NotNil(t, s)

}
