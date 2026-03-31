########################################
### Simulations
###
### Several of these are used in .github/workflows/sims.yml.
### The strings in there leave off the "test-sim" prefix.
###
### Environment Variables:
###   GO:             The command to use to execute go. Default: go
###   SEED:           Optional random seed for simulation tests.
###                   Usage: SEED=57 make test-sim-import-export
###   BINDIR:         The Go bin directory, defaults to $GOPATH/bin
###   SIM_GENESIS:    Defines the path to the custom genesis file used by
###                   test-sim-custom-genesis-multi-seed and test-sim-custom-genesis-fast
###                   Default is ${HOME}/.provenanced/config/genesis.json.
###   SIM_NUM_BLOCKS: The number of blocks to use for test-sim-benchmark or test-sim-profile. Default is 500.
###   SIM_BLOCK_SIZE: The size of blocks to use for test-sim-benchmark or test-sim-profile. Default is 200.
###   SIM_COMMIT:     Whether to commit during  test-sim-benchmark or test-sim-profile. Default is true.

SIMAPP = ./app

ifdef SEED 
	SEED_ARG = -Seed=$(SEED)
else
	SEED_ARG =
endif

SIM_GENESIS ?= ${HOME}/.provenanced/config/genesis.json

test-sim-import-export:
	@echo "Running application import/export simulation. This may take several minutes..."
	$(GO) test -mod=readonly -tags sims $(SIMAPP) -run TestAppImportExport \
		-NumBlocks=30 -BlockSize=100 \
		-Commit=true $(SEED_ARG) -Period=3 -v -timeout 1h

test-sim-after-import:
	@echo "Running application simulation-after-import. This may take several minutes..."
	$(GO) test -mod=readonly -tags sims $(SIMAPP) -run TestAppSimulationAfterImport \
		-NumBlocks=30 -BlockSize=100 \
		-Commit=true $(SEED_ARG) -Period=3 -v -timeout 1h

test-sim-custom-genesis-multi-seed:
	@test -f '$(SIM_GENESIS)' || (echo "ERROR: Genesis file not found at $(SIM_GENESIS).\nRun 'provenanced init testnode' to create one, or set SIM_GENESIS=/path/to/genesis.json" && exit 1)
	@echo "Running multi-seed custom genesis simulation..."
	$(GO) test -mod=readonly -tags sims $(SIMAPP) -run TestFullAppSimulation \
		-Genesis='$(SIM_GENESIS)' -NumBlocks=400 -BlockSize=100 \
		-Commit=true $(SEED_ARG) -Period=5 -v -timeout 24h

test-sim-multi-seed-long:
	@echo "Running long multi-seed application simulation. This may take awhile!"
	$(GO) test -mod=readonly -tags sims $(SIMAPP) -run TestFullAppSimulation \
		-NumBlocks=500 -BlockSize=100 \
		-Commit=true $(SEED_ARG) -Period=50 -v -timeout 24h

test-sim-multi-seed-short:
	@echo "Running short multi-seed application simulation. This may take awhile!"
	$(GO) test -mod=readonly -tags sims $(SIMAPP) -run TestFullAppSimulation \
		-NumBlocks=50 -BlockSize=100 \
		-Commit=true $(SEED_ARG) -Period=10 -v -timeout 1h


test-sim-nondeterminism:
	@echo "Running non-determinism test..."
	$(GO) test -mod=readonly -tags sims $(SIMAPP) -run TestAppStateDeterminism \
		-NumBlocks=50 -BlockSize=100 -Commit=true -Period=0 -v -timeout 24h

test-sim-custom-genesis-fast:
	@test -f '$(SIM_GENESIS)' || (echo "ERROR: Genesis file not found at $(SIM_GENESIS).\nRun 'provenanced init testnode' to create one, or set SIM_GENESIS=/path/to/genesis.json" && exit 1)
	@echo "Running custom genesis simulation..."
	$(GO) test -mod=readonly -tags sims $(SIMAPP) -run TestFullAppSimulation -Genesis='$(SIM_GENESIS)' \
		-NumBlocks=50 -BlockSize=100 \
		-Commit=true $(SEED_ARG) -Period=5 -v -timeout 24h

test-sim-simple:
	@echo "Running simple module simulation..."
	$(GO) test -mod=readonly -tags sims $(SIMAPP) -run TestSimple \
		-NumBlocks=50 -BlockSize=100 \
		-Commit=true $(SEED_ARG) -Period=5 -v -timeout 1h

test-sim-benchmark-invariants:
	@echo "Running simulation invariant benchmarks..."
	$(GO) test -mod=readonly -tags sims -run=^$$ $(SIMAPP) -benchmem -bench=BenchmarkInvariants \
		-NumBlocks=1000 -BlockSize=200 \
		-Period=1 -Commit=true $(SEED_ARG) -v -timeout 24h

SIM_NUM_BLOCKS ?= 500
SIM_BLOCK_SIZE ?= 200
SIM_COMMIT ?= true

test-sim-benchmark:
	@echo "Running application benchmark for numBlocks=$(SIM_NUM_BLOCKS), blockSize=$(SIM_BLOCK_SIZE). This may take awhile!"
	$(GO) test -mod=readonly -tags sims -benchmem -run=^$$ $(SIMAPP) -bench ^BenchmarkFullAppSimulation$$ \
		-NumBlocks=$(SIM_NUM_BLOCKS) -BlockSize=$(SIM_BLOCK_SIZE) \
		-Commit=$(SIM_COMMIT) -timeout 24h

# Same as test-sim-benchmark except also creates files with cpu and memory profile info.
test-sim-profile:
	@echo "Running application benchmark for numBlocks=$(SIM_NUM_BLOCKS), blockSize=$(SIM_BLOCK_SIZE). This may take awhile!"
	$(GO) test -mod=readonly -tags sims -benchmem -run=^$$ $(SIMAPP) -bench ^BenchmarkFullAppSimulation$$ \
		-NumBlocks=$(SIM_NUM_BLOCKS) -BlockSize=$(SIM_BLOCK_SIZE) \
		-Commit=$(SIM_COMMIT) -timeout 24h -cpuprofile cpu.out -memprofile mem.out

.PHONY: \
test-sim-nondeterminism \
test-sim-custom-genesis-fast \
test-sim-simple \
test-sim-import-export \
test-sim-after-import \
test-sim-custom-genesis-multi-seed \
test-sim-multi-seed-short \
test-sim-multi-seed-long \
test-sim-benchmark-invariants \
test-sim-profile \
test-sim-benchmark