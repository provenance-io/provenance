package reward_test

import (
	"testing"
	"time"

	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/reward"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func TestEndBlockWithNoActiveRewards(t *testing.T) {
	var app *simapp.App
	var ctx sdk.Context

	now := time.Now()

	app = simapp.Setup(t)
	ctx = app.BaseApp.NewContext(false, tmproto.Header{})
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
	ctx = app.BaseApp.NewContext(false, tmproto.Header{})
	ctx = ctx.WithBlockHeight(1).WithBlockTime(now)

	ctx = ctx.WithBlockHeight(2)
	require.NotPanics(t, func() {
		reward.BeginBlocker(ctx, app.RewardKeeper)
	})
}
