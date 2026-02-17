package app

// DONTCOVER

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// Profile with:
// go test -benchmem -run=^$ github.com/provenance-io/provenance/app -bench ^BenchmarkFullAppSimulation$ -Commit=true -cpuprofile cpu.out
func BenchmarkFullAppSimulation(b *testing.B) {
	b.ReportAllocs()
	config, db, dir, logger, skip, err := setupSimulation("goleveldb-app-sim", "Simulation")
	require.NoError(b, err, "simulation setup failed")

	if skip {
		b.Skip("skipping benchmark application simulation")
	}
	printConfig(config)

	defer func() {
		require.NoError(b, db.Close())
		require.NoError(b, os.RemoveAll(dir))
	}()

	appOpts := newSimAppOpts(b)
	baseAppOpts := []func(*baseapp.BaseApp){
		interBlockCacheOpt(),
		baseapp.SetChainID(config.ChainID),
	}
	app := New(logger, db, nil, true, appOpts, baseAppOpts...)

	// run randomized simulation
	_, simParams, simErr := simulation.SimulateFromSeed(
		b,
		os.Stdout,
		app.BaseApp,
		provAppStateFn(app.AppCodec(), app.SimulationManager(), app.DefaultGenesis()),
		simtypes.RandomAccounts, // Replace with own random account function if using keys other than secp256k1
		simtestutil.SimulationOperations(app, app.AppCodec(), config),
		app.ModuleAccountAddrs(),
		config,
		app.AppCodec(),
	)

	// export state and simParams before the simulation error is checked
	err = simtestutil.CheckExportSimulation(app, config, simParams)
	require.NoError(b, err, "CheckExportSimulation")
	require.NoError(b, simErr, "SimulateFromSeed")

	printStats(config, db)
}

func BenchmarkInvariants(b *testing.B) {
	b.ReportAllocs()
	config, db, dir, logger, skip, err := setupSimulation("leveldb-app-invariant-bench", "Simulation")
	require.NoError(b, err, "simulation setup failed")

	if skip {
		b.Skip("skipping benchmark application simulation")
	}

	config.AllInvariants = false
	printConfig(config)

	defer func() {
		require.NoError(b, db.Close())
		require.NoError(b, os.RemoveAll(dir))
	}()

	appOpts := newSimAppOpts(b)
	baseAppOpts := []func(*baseapp.BaseApp){
		interBlockCacheOpt(),
		baseapp.SetChainID(config.ChainID),
	}
	app := New(logger, db, nil, true, appOpts, baseAppOpts...)

	// run randomized simulation
	_, lastBlockTime, simParams, simErr := simulation.SimulateFromSeedProv(
		b,
		os.Stdout,
		app.BaseApp,
		provAppStateFn(app.AppCodec(), app.SimulationManager(), app.DefaultGenesis()),
		simtypes.RandomAccounts, // Replace with own random account function if using keys other than secp256k1
		simtestutil.SimulationOperations(app, app.AppCodec(), config),
		app.ModuleAccountAddrs(),
		config,
		app.AppCodec(),
	)

	// export state and simParams before the simulation error is checked
	err = simtestutil.CheckExportSimulation(app, config, simParams)
	require.NoError(b, err, "CheckExportSimulation")
	require.NoError(b, simErr, "SimulateFromSeedProv")

	printStats(config, db)

	ctx := app.NewContextLegacy(true, cmtproto.Header{Height: app.LastBlockHeight() + 1, Time: lastBlockTime})

	// 3. Benchmark each invariant separately
	//
	// NOTE: We use the crisis keeper as it has all the invariants registered with
	// their respective metadata which makes it useful for testing/benchmarking.
	for _, cr := range app.CrisisKeeper.Routes() {
		cr := cr
		b.Run(fmt.Sprintf("%s/%s", cr.ModuleName, cr.Route), func(b *testing.B) {
			if res, stop := cr.Invar(ctx); stop {
				b.Fatalf(
					"broken invariant at block %d of %d\n%s",
					ctx.BlockHeight()-1, config.NumBlocks, res,
				)
			}
		})
	}
}
