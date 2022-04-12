package epoch_test

import (
	"testing"
	"time"

	simapp "github.com/provenance-io/provenance/app"
	"github.com/provenance-io/provenance/x/epoch"
	"github.com/provenance-io/provenance/x/epoch/keeper"
	"github.com/provenance-io/provenance/x/epoch/types"

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
		expectedCurrentEpochStartHeight uint64
		expectedStartHeight             uint64
		expectedCurrentEpoch            uint64
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
					StartHeight:             1,
					Duration:                (60 * 60 * 24 * 30) / 5,
					CurrentEpoch:            0,
					CurrentEpochStartHeight: uint64(ctx.BlockHeight()),
					EpochCountingStarted:    false,
				},
			},
		})

		test.fn()

		require.Equal(t, epochInfo.Identifier, "monthly")
		require.Equal(t, test.expectedCurrentEpochStartHeight, epochInfo.CurrentEpochStartHeight)
		require.Equal(t, (60*60*24*30)/5, int(epochInfo.Duration))
		require.Equal(t, test.expectedCurrentEpoch, epochInfo.CurrentEpoch)
		require.Equal(t, test.expectedStartHeight, epochInfo.StartHeight)
		require.Equal(t, epochInfo.EpochCountingStarted, true)
	}
}

func TestEpochStartingOneMonthAfterInitGenesis(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	// On init genesis, default epochs information is set
	// To check init genesis again, should make it fresh status
	epochInfos := app.EpochKeeper.AllEpochInfos(ctx)
	for _, epochInfo := range epochInfos {
		app.EpochKeeper.DeleteEpochInfo(ctx, epochInfo.Identifier)
	}

	initialBlockHeight := int64(1)
	ctx = ctx.WithBlockHeight(initialBlockHeight)

	epoch.InitGenesis(ctx, app.EpochKeeper, types.GenesisState{
		Epochs: []types.EpochInfo{
			{
				Identifier:              "monthly",
				StartHeight:             uint64(ctx.BlockHeight() + (60*60*24*30)/5),
				Duration:                (60 * 60 * 24 * 30) / 5,
				CurrentEpoch:            0,
				CurrentEpochStartHeight: uint64(initialBlockHeight),
				EpochCountingStarted:    false,
			},
		},
	})

	// epoch not started yet
	epochInfo := app.EpochKeeper.GetEpochInfo(ctx, "monthly")
	require.Equal(t, epochInfo.CurrentEpoch, uint64(0))
	require.Equal(t, epochInfo.StartHeight, uint64(initialBlockHeight+(60*60*24*30)/5))
	require.Equal(t, epochInfo.CurrentEpochStartHeight, uint64(ctx.BlockHeight()))
	require.Equal(t, epochInfo.EpochCountingStarted, false)

	// after 1 week
	ctx = ctx.WithBlockHeight((7*24*60*60)/5 + initialBlockHeight)
	epoch.BeginBlocker(ctx, app.EpochKeeper)

	// epoch not started yet
	epochInfo = app.EpochKeeper.GetEpochInfo(ctx, "monthly")
	require.Equal(t, epochInfo.CurrentEpoch, uint64(0))
	require.Equal(t, epochInfo.StartHeight, uint64(initialBlockHeight+(60*60*24*30)/5))
	require.Equal(t, epochInfo.CurrentEpochStartHeight, uint64(initialBlockHeight))
	require.Equal(t, epochInfo.EpochCountingStarted, false)

	// after 1 month
	ctx = ctx.WithBlockHeight((24*60*60*30)/5 + initialBlockHeight)
	epoch.BeginBlocker(ctx, app.EpochKeeper)

	// epoch started
	epochInfo = app.EpochKeeper.GetEpochInfo(ctx, "monthly")
	require.Equal(t, epochInfo.CurrentEpoch, uint64(1))
	require.Equal(t, epochInfo.CurrentEpochStartHeight, uint64(ctx.BlockHeight()))
	require.Equal(t, epochInfo.StartHeight, uint64(ctx.BlockHeight()))
	require.Equal(t, epochInfo.EpochCountingStarted, true)
}

// Mock to capture arguments passed into hooks
type EpochHooksMock struct {
	AfterEpochNumber  *uint64
	BeforeEpochNumber *uint64
}

// Stubs the AfterEpochEnd method to capture the epochNumber
func (h EpochHooksMock) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber uint64) {
	*h.AfterEpochNumber = epochNumber
}

// Stubs the BeforeEpochStart method to capture the epochNumber
func (h EpochHooksMock) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber uint64) {
	*h.BeforeEpochNumber = epochNumber
}

func TestBeforeAndAfterHooks(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	// Setup Mock
	// The stubbed functions will populate these variables
	var beforeEpochNumber uint64
	var afterEpochNumber uint64
	hooksMock := EpochHooksMock{
		AfterEpochNumber:  &afterEpochNumber,
		BeforeEpochNumber: &beforeEpochNumber,
	}
	app.EpochKeeper = *keeper.NewKeeper(app.AppCodec(), app.GetKey("epoch"))
	app.EpochKeeper.SetHooks(hooksMock)

	epochInfos := app.EpochKeeper.AllEpochInfos(ctx)
	now := time.Now()

	for _, epochInfo := range epochInfos {
		app.EpochKeeper.DeleteEpochInfo(ctx, epochInfo.Identifier)
	}

	initialBlockHeight := int64(1)
	ctx = ctx.WithBlockHeight(initialBlockHeight).WithBlockTime(now)
	epoch.InitGenesis(ctx, app.EpochKeeper, types.GenesisState{
		Epochs: []types.EpochInfo{
			{
				Identifier:              "minutely",
				StartHeight:             1,
				Duration:                (60) / 5,
				CurrentEpoch:            0,
				CurrentEpochStartHeight: uint64(initialBlockHeight),
				EpochCountingStarted:    false,
			},
		},
	})

	// Only the BeforeEpochStart hook should be called
	// It should have the latest epochNumber
	epoch.BeginBlocker(ctx, app.EpochKeeper)
	epochInfo := app.EpochKeeper.GetEpochInfo(ctx, "minutely")
	require.Equal(t, epochInfo.CurrentEpoch, uint64(1))
	require.Equal(t, uint64(1), beforeEpochNumber)
	require.Equal(t, uint64(0), afterEpochNumber)

	// Both hooks should be called
	// BeforeEpochStart should contain the new epochNumber
	// AfterEpochEnd should contain the previous epochNumber
	ctx = ctx.WithBlockHeight(14)
	epoch.BeginBlocker(ctx, app.EpochKeeper)
	epochInfo = app.EpochKeeper.GetEpochInfo(ctx, "minutely")
	require.Equal(t, epochInfo.CurrentEpoch, uint64(2))
	require.Equal(t, uint64(2), beforeEpochNumber)
	require.Equal(t, uint64(1), afterEpochNumber)
}
