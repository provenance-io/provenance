package attribute_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/attribute"
	"github.com/provenance-io/provenance/x/attribute/types"
)

func TestBeginBlockDeletionOfExpired(t *testing.T) {
	var app *simapp.App
	var ctx sdk.Context

	pubkey1 := secp256k1.GenPrivKey().PubKey()
	user1Addr := sdk.AccAddress(pubkey1.Address())
	user1 := user1Addr.String()
	now := time.Now()

	app = simapp.Setup(t)
	ctx = app.BaseApp.NewContext(false, cmtproto.Header{})
	ctx = ctx.WithBlockTime(now.Add(-3 * time.Hour))
	app.AccountKeeper.SetAccount(ctx, app.AccountKeeper.NewAccountWithAddress(ctx, user1Addr))

	ctx = ctx.WithEventManager(sdk.NewEventManager())
	attribute.BeginBlocker(ctx, app.AttributeKeeper)
	assert.Empty(t, ctx.EventManager().Events())

	past := now.Add(-2 * time.Hour)

	require.NoError(t, app.NameKeeper.SetNameRecord(ctx, "one.expire.testing", user1Addr, false), "name record should save successfully")
	require.NoError(t, app.NameKeeper.SetNameRecord(ctx, "two.expire.testing", user1Addr, false), "name record should save successfully")

	attr1 := types.NewAttribute("one.expire.testing", user1, types.AttributeType_String, []byte("test1"), nil)
	attr1.ExpirationDate = &past
	require.NoError(t, app.AttributeKeeper.SetAttribute(ctx, attr1, user1Addr))

	attr2 := types.NewAttribute("two.expire.testing", user1, types.AttributeType_String, []byte("test2"), nil)
	attr2.ExpirationDate = &past
	require.NoError(t, app.AttributeKeeper.SetAttribute(ctx, attr2, user1Addr))

	ctx = ctx.WithBlockTime(now)
	ctx = ctx.WithEventManager(sdk.NewEventManager())
	attribute.BeginBlocker(ctx, app.AttributeKeeper)
	events := ctx.EventManager().Events()
	assert.Len(t, events, 3)
	assert.Equal(t, "beginblock", events[2].Type)
	assert.Equal(t, sdk.AttributeKeyModule, events[2].Attributes[0].Key)
	assert.Equal(t, types.ModuleName, events[2].Attributes[0].Value)
	assert.Equal(t, sdk.AttributeKeyAction, events[2].Attributes[1].Key)
	assert.Equal(t, types.EventTypeDeletedExpired, events[2].Attributes[1].Value)
	assert.Equal(t, types.AttributeKeyTotalExpired, events[2].Attributes[2].Key)
	assert.Equal(t, "2", events[2].Attributes[2].Value)

}
