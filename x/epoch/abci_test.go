package epoch_test

import (
	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/epoch"
	"github.com/provenance-io/provenance/x/epoch/types"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func TestEpochInfoChangesBeginBlockerAndInitGenesis(t *testing.T) {
	var app *simapp.App
	var ctx sdk.Context
	var epochInfo types.EpochInfo

	now := time.Now()

	tests := []struct {
		expectedCurrentEpochStartHeight int64
		expectedStartHeight             int64
		expectedCurrentEpoch            int64
		fn                              func()
	}{
		{
			// Only advance 2 seconds, do not increment epoch
			expectedCurrentEpochStartHeight: 2,
			expectedStartHeight:             1,
			expectedCurrentEpoch:            1,
			fn: func() {
				ctx = ctx.WithBlockHeight(2)
				epoch.BeginBlocker(ctx, app.EpochKeeper)
				epochInfo = app.EpochKeeper.GetEpochInfo(ctx, "monthly")
			},
		},
		{
			expectedCurrentEpochStartHeight: 2,
			expectedStartHeight:             1,
			expectedCurrentEpoch:            1,
			fn: func() {
				ctx = ctx.WithBlockHeight(2)
				epoch.BeginBlocker(ctx, app.EpochKeeper)
				ctx = ctx.WithBlockHeight((60 * 60 * 24 * 30) / 5)
				epoch.BeginBlocker(ctx, app.EpochKeeper)
				epochInfo = app.EpochKeeper.GetEpochInfo(ctx, "monthly")
			},
		},
		{
			expectedCurrentEpochStartHeight: 535680,
			expectedStartHeight:             1,
			expectedCurrentEpoch:            2,
			fn: func() {
				ctx = ctx.WithBlockHeight(2)
				epoch.BeginBlocker(ctx, app.EpochKeeper)
				ctx = ctx.WithBlockHeight((60 * 60 * 24 * 31) / 5)
				epoch.BeginBlocker(ctx, app.EpochKeeper)
				epochInfo = app.EpochKeeper.GetEpochInfo(ctx, "monthly")
			},
		},
		{
			expectedCurrentEpochStartHeight: 535680,
			expectedStartHeight:             1,
			expectedCurrentEpoch:            2,
			fn: func() {
				ctx = ctx.WithBlockHeight(2)
				epoch.BeginBlocker(ctx, app.EpochKeeper)
				ctx = ctx.WithBlockHeight((60 * 60 * 24 * 31) / 5)
				epoch.BeginBlocker(ctx, app.EpochKeeper)
				ctx = ctx.WithBlockHeight((60 * 60 * 24 * 32) / 5)
				epoch.BeginBlocker(ctx, app.EpochKeeper)
				epochInfo = app.EpochKeeper.GetEpochInfo(ctx, "monthly")
			},
		},
		{
			expectedCurrentEpochStartHeight: 535680,
			expectedStartHeight:             1,
			expectedCurrentEpoch:            2,
			fn: func() {
				ctx = ctx.WithBlockHeight(2)
				epoch.BeginBlocker(ctx, app.EpochKeeper)
				ctx = ctx.WithBlockHeight((60 * 60 * 24 * 31) / 5)
				epoch.BeginBlocker(ctx, app.EpochKeeper)
				numBlocksSinceStart, _ := app.EpochKeeper.NumBlocksSinceEpochStart(ctx, "monthly")
				require.Equal(t, int64(0), numBlocksSinceStart)
				ctx = ctx.WithBlockHeight((60 * 60 * 24 * 32) / 5)
				epoch.BeginBlocker(ctx, app.EpochKeeper)
				epochInfo = app.EpochKeeper.GetEpochInfo(ctx, "monthly")
			},
		},
	}

	for _, test := range tests {
		app = simapp.Setup(false)
		ctx = app.BaseApp.NewContext(false, tmproto.Header{})

		// On init genesis, default epoch information is set
		// To check init genesis again, should make it fresh status
		epochInfos := app.EpochKeeper.AllEpochInfos(ctx)
		for _, epochInfo := range epochInfos {
			app.EpochKeeper.DeleteEpochInfo(ctx, epochInfo.Identifier)
		}

		ctx = ctx.WithBlockHeight(1).WithBlockTime(now)

		// check init genesis
		epoch.InitGenesis(ctx, app.EpochKeeper, types.GenesisState{
			Epochs: []types.EpochInfo{
				{
					Identifier:              "monthly",
					StartHeight:             0,
					Duration:                (60 * 60 * 24 * 30) / 5,
					CurrentEpoch:            0,
					CurrentEpochStartHeight: ctx.BlockHeight(),
					EpochCountingStarted:    false,
				},
			},
		})

		test.fn()

		require.Equal(t, epochInfo.Identifier, "monthly")
		require.Equal(t, epochInfo.CurrentEpochStartHeight, test.expectedCurrentEpochStartHeight)
		require.Equal(t, int(epochInfo.Duration), (60*60*24*30)/5)
		require.Equal(t, epochInfo.CurrentEpoch, test.expectedCurrentEpoch)
		require.Equal(t, epochInfo.StartHeight, test.expectedStartHeight)
		require.Equal(t, epochInfo.EpochCountingStarted, true)
	}
}
