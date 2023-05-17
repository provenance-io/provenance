package attribute_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/attribute"
	"github.com/stretchr/testify/assert"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func TestBeginBlockDeletionOfExpired(t *testing.T) {
	var app *simapp.App
	var ctx sdk.Context

	now := time.Now()

	app = simapp.Setup(t)
	ctx = app.BaseApp.NewContext(false, tmproto.Header{})
	ctx = ctx.WithBlockHeight(1).WithBlockTime(now)

	ctx = ctx.WithEventManager(sdk.NewEventManager())
	attribute.BeginBlocker(ctx, app.AttributeKeeper)
	assert.Empty(t, ctx.EventManager().Events())

}
