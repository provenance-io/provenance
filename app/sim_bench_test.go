package app

// DONTCOVER

import (
	"fmt"
	"os"
	"testing"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdksim "github.com/cosmos/cosmos-sdk/simapp"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// Profile with:
// go test -benchmem -run=^$ github.com/provenance-io/provenance/app -bench ^BenchmarkFullAppSimulation$ -Commit=true -cpuprofile cpu.out
func BenchmarkFullAppSimulation(b *testing.B) {
	b.ReportAllocs()
	config, db, dir, logger, skip, err := sdksim.SetupSimulation("goleveldb-app-sim", "Simulation")
	if err != nil {
		b.Fatalf("simulation setup failed: %s", err.Error())
	}

	if skip {
		b.Skip("skipping benchmark application simulation")
	}
	PrintConfig(config)

	defer func() {
		db.Close()
		err = os.RemoveAll(dir)
		if err != nil {
			b.Fatal(err)
		}
	}()

	app := New(logger, db, nil, true, map[int64]bool{}, DefaultNodeHome, sdksim.FlagPeriodValue, MakeEncodingConfig(), EmptyAppOptions{}, interBlockCacheOpt())

	// run randomized simulation
	_, simParams, simErr := simulation.SimulateFromSeed(
		b,
		os.Stdout,
		app.BaseApp,
		sdksim.AppStateFn(app.AppCodec(), app.SimulationManager()),
		simtypes.RandomAccounts, // Replace with own random account function if using keys other than secp256k1
		sdksim.SimulationOperations(app, app.AppCodec(), config),
		app.ModuleAccountAddrs(),
		config,
		app.AppCodec(),
	)

	// export state and simParams before the simulation error is checked
	if err = sdksim.CheckExportSimulation(app, config, simParams); err != nil {
		b.Fatal(err)
	}

	if simErr != nil {
		b.Fatal(simErr)
	}

	PrintStats(config, db)
}

func BenchmarkInvariants(b *testing.B) {
	b.ReportAllocs()
	config, db, dir, logger, skip, err := sdksim.SetupSimulation("leveldb-app-invariant-bench", "Simulation")
	if err != nil {
		b.Fatalf("simulation setup failed: %s", err.Error())
	}

	if skip {
		b.Skip("skipping benchmark application simulation")
	}

	config.AllInvariants = false
	PrintConfig(config)

	defer func() {
		db.Close()
		err = os.RemoveAll(dir)
		if err != nil {
			b.Fatal(err)
		}
	}()

	app := New(logger, db, nil, true, map[int64]bool{}, DefaultNodeHome, sdksim.FlagPeriodValue, MakeEncodingConfig(), EmptyAppOptions{}, interBlockCacheOpt())

	// run randomized simulation
	_, simParams, simErr := simulation.SimulateFromSeed(
		b,
		os.Stdout,
		app.BaseApp,
		sdksim.AppStateFn(app.AppCodec(), app.SimulationManager()),
		simtypes.RandomAccounts, // Replace with own random account function if using keys other than secp256k1
		sdksim.SimulationOperations(app, app.AppCodec(), config),
		app.ModuleAccountAddrs(),
		config,
		app.AppCodec(),
	)

	// export state and simParams before the simulation error is checked
	if err = sdksim.CheckExportSimulation(app, config, simParams); err != nil {
		b.Fatal(err)
	}

	if simErr != nil {
		b.Fatal(simErr)
	}

	PrintStats(config, db)

	ctx := app.NewContext(true, tmproto.Header{Height: app.LastBlockHeight() + 1})

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
