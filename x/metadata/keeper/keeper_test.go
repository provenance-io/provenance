package keeper_test

import (
	"testing"

	"github.com/google/uuid"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	simapp "github.com/provenance-io/provenance/app"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/stretchr/testify/require"

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
