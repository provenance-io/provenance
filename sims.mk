########################################
### Simulations
###
### Several of these are used in .github/workflows/sims.yml.
### The strings in there leave off the "test-sim" prefix.
###
### Environment Variables:
###   GO:             The command to use to execute go. Default: go
###   DB_BACKEND:     Dictates which db backend to use: goleveldb.
###                   The test-sim-nondeterminism is hard-coded to use memdb though.
###   BINDIR:         The Go bin directory, defaults to $GOPATH/bin
###   SIM_GENESIS:    Defines the path to the custom genesis file used by
###                   test-sim-custom-genesis-multi-seed and test-sim-custom-genesis-fast
###                   Default is ${HOME}/.provenanced/config/genesis.json.
###   SIM_NUM_BLOCKS: The number of blocks to use for test-sim-benchmark or test-sim-profile. Default is 500.
###   SIM_BLOCK_SIZE: The size of blocks to use for test-sim-benchmark or test-sim-profile. Default is 200.
###   SIM_COMMIT:     Whether to commit during  test-sim-benchmark or test-sim-profile. Default is true.

SIMAPP = ./app
DB_BACKEND ?= goleveldb
ifneq ($(DB_BACKEND),goleveldb)
  $(error unknown DB_BACKEND value [$(DB_BACKEND)]. Must be goleveldb)
endif

SIM_GENESIS ?= ${HOME}/.provenanced/config/genesis.json

test-sim-import-export:
	@echo "Running application import/export simulation. This may take several minutes..."
	$(GO) test -mod=readonly $(SIMAPP) -run TestAppImportExport \
	-Enabled=true -NumBlocks=30 -BlockSize=100 -DBBackend=$(DB_BACKEND) \
	-Commit=true -Seed=57 -Period=3 -v -timeout 1h

test-sim-after-import:
	@echo "Running application simulation-after-import. This may take several minutes..."
	$(GO) test -mod=readonly $(SIMAPP) -run TestAppSimulationAfterImport \
	-Enabled=true -NumBlocks=30 -BlockSize=100 -DBBackend=$(DB_BACKEND) \
	-Commit=true -Seed=57 -Period=3 -v -timeout 1h

test-sim-custom-genesis-multi-seed:
	@test -f '$(SIM_GENESIS)' || (echo "ERROR: Genesis file not found at $(SIM_GENESIS).\nRun 'provenanced init testnode' to create one, or set SIM_GENESIS=/path/to/genesis.json" && exit 1)
	@echo "Running multi-seed custom genesis simulation..."
	$(GO) test -mod=readonly $(SIMAPP) -run TestFullAppSimulation \
		-Genesis='$(SIM_GENESIS)' -Enabled=true -NumBlocks=400 -BlockSize=100 \
		-DBBackend=$(DB_BACKEND) -Commit=true -Seed=57 -Period=5 -v -timeout 24h

test-sim-multi-seed-long:
	@echo "Running long multi-seed application simulation. This may take awhile!"
	$(GO) test -mod=readonly $(SIMAPP) -run TestFullAppSimulation \
	-Enabled=true -NumBlocks=500 -BlockSize=100 -DBBackend=$(DB_BACKEND) \
	-Commit=true -Seed=57 -Period=50 -v -timeout 24h

test-sim-multi-seed-short:
	@echo "Running short multi-seed application simulation. This may take awhile!"
	$(GO) test -mod=readonly $(SIMAPP) -run TestFullAppSimulation \
	-Enabled=true -NumBlocks=50 -BlockSize=100 -DBBackend=$(DB_BACKEND) \
	-Commit=true -Seed=57 -Period=10 -v -timeout 1h


test-sim-nondeterminism:
	@echo "Running non-determinism test..."
	$(GO) test -mod=readonly $(SIMAPP) -run TestAppStateDeterminism -Enabled=true \
		-NumBlocks=50 -BlockSize=100 -Commit=true -Period=0 -v -timeout 24h

test-sim-custom-genesis-fast:
	@test -f '$(SIM_GENESIS)' || (echo "ERROR: Genesis file not found at $(SIM_GENESIS).\nRun 'provenanced init testnode' to create one, or set SIM_GENESIS=/path/to/genesis.json" && exit 1)
	@echo "Running custom genesis simulation..."
	$(GO) test -mod=readonly $(SIMAPP) -run TestFullAppSimulation -Genesis='$(SIM_GENESIS)' \
		-Enabled=true -NumBlocks=50 -BlockSize=100 -DBBackend=$(DB_BACKEND) \
		-Commit=true -Seed=99 -Period=5 -v -timeout 24h

test-sim-simple:
	@echo "Running simple module simulation..."
	$(GO) test -mod=readonly $(SIMAPP) -run TestSimple \
		-Enabled=true -NumBlocks=50 -BlockSize=100 -DBBackend=$(DB_BACKEND) \
		-Commit=true -Seed=99 -Period=5 -v -timeout 1h

test-sim-benchmark-invariants:
	@echo "Running simulation invariant benchmarks..."
	$(GO) test -mod=readonly -run=^$$ $(SIMAPP) -benchmem -bench=BenchmarkInvariants \
		-Enabled=true -NumBlocks=1000 -BlockSize=200 -DBBackend=$(DB_BACKEND) \
		-Period=1 -Commit=true -Seed=57 -v -timeout 24h

SIM_NUM_BLOCKS ?= 500
SIM_BLOCK_SIZE ?= 200
SIM_COMMIT ?= true

test-sim-benchmark:
	@echo "Running application benchmark for numBlocks=$(SIM_NUM_BLOCKS), blockSize=$(SIM_BLOCK_SIZE). This may take awhile!"
	$(GO) test -mod=readonly -benchmem -run=^$$ $(SIMAPP) -bench ^BenchmarkFullAppSimulation$$ \
		-Enabled=true -NumBlocks=$(SIM_NUM_BLOCKS) -BlockSize=$(SIM_BLOCK_SIZE) -DBBackend=$(DB_BACKEND) \
		-Commit=$(SIM_COMMIT) -timeout 24h

# Same as test-sim-benchmark except also creates files with cpu and memory profile info.
test-sim-profile:
	@echo "Running application benchmark for numBlocks=$(SIM_NUM_BLOCKS), blockSize=$(SIM_BLOCK_SIZE). This may take awhile!"
	$(GO) test -mod=readonly -benchmem -run=^$$ $(SIMAPP) -bench ^BenchmarkFullAppSimulation$$ \
		-Enabled=true -NumBlocks=$(SIM_NUM_BLOCKS) -BlockSize=$(SIM_BLOCK_SIZE) -DBBackend=$(DB_BACKEND) \
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
