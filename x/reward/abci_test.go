package reward_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/reward"
)

func TestEndBlockWithNoActiveRewards(t *testing.T) {
	var app *simapp.App
	var ctx sdk.Context

	now := time.Now()

	app = simapp.Setup(t)
	ctx = app.BaseApp.NewContext(false)
	ctx = ctx.WithBlockHeight(1).WithBlockTime(now)

	ctx = ctx.WithBlockHeight(2)
	require.NotPanics(t, func() {
		reward.EndBlocker(ctx, app.RewardKeeper)
	})
}

func TestBeginBlockWithNoActiveRewards(t *testing.T) {
	var app *simapp.App
	var ctx sdk.Context

	now := time.Now()

	app = simapp.Setup(t)
	ctx = app.BaseApp.NewContext(false)
	ctx = ctx.WithBlockHeight(1).WithBlockTime(now)

	ctx = ctx.WithBlockHeight(2)
	require.NotPanics(t, func() {
		reward.BeginBlocker(ctx, app.RewardKeeper)
	})
}
