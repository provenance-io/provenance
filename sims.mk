#!/usr/bin/make -f

########################################
### Simulations
###
### Several of these are used in .github/workflows/sims.yml.
### The strings in there leave off the "test-sim" prefix.
###
### Environment Variables:
###   GO:             The command to use to execute go. Default: go
###   DB_BACKEND:     Dictates which db backend to use: goleveldb, cleveldb, rocksdb, badgerdb.
###                   The test-sim-nondeterminism is hard-coded to use memdb though.
###   BINDIR:         The Go bin directory, defaults to $GOPATH/bin
###   SIM_GENESIS:    Defines the path to the custom genesis file used by
###                   test-sim-custom-genesis-multi-seed and test-sim-custom-genesis-fast
###                   Default is ${HOME}/.provenanced/config/genesis.json.
###   SIM_NUM_BLOCKS: The number of blocks to use for test-sim-benchmark or test-sim-profile. Default is 500.
###   SIM_BLOCK_SIZE: The size of blocks to use for test-sim-benchmark or test-sim-profile. Default is 200.
###   SIM_COMMIT:     Whether to commit during  test-sim-benchmark or test-sim-profile. Default is true.

GO ?= go
BINDIR ?= $(GOPATH)/bin
SIMAPP = ./app
DB_BACKEND ?= goleveldb
ifeq ($(DB_BACKEND),cleveldb)
  db_tag = cleveldb
else ifeq ($(DB_BACKEND),rocksdb)
  db_tag = rocksdb
else ifeq ($(DB_BACKEND),badgerdb)
  db_tag = badgerdb
else ifneq ($(DB_BACKEND),goleveldb)
  $(error unknown DB_BACKEND value [$(DB_BACKEND)]. Must be one of goleveldb, cleveldb, rocksdb, badgerdb)
endif

# We have to use a hack to provide -tags with the runsim stuff, but it only allows us to provide one tag.
# Runsim creates a command string, then does a split on " " to turn it into args.
# With two tags, e.g. -tags 'foo bar', you'd end up with three args, "-tags", "'foo", and "bar'", and it'll get confused.
# But we CAN provide a single tag in the -SimAppPkg value in order to trick it into including it in the `go test` commands.
# We need to provide the -DBBackend flag to the runsim tests too, and use the same hack.
SIMAPP += -DBBackend=$(DB_BACKEND)
ifneq ($(db_tag),)
  SIMAPP += -tags $(db_tag)
endif

SIM_GENESIS ?= ${HOME}/.provenanced/config/genesis.json

# Runsim Usage: runsim [flags] <blocks> <period> <testname>
# flags: [-Jobs maxprocs] [-ExitOnFail] [-Seeds comma-separated-seed-list]
#        [-Genesis file-path] [-SimAppPkg file-path] [-Github] [-Slack] [-LogObjPrefix string]

test-sim-import-export: runsim
	@echo "Running application import/export simulation. This may take several minutes..."
	$(BINDIR)/runsim -Jobs=4 -SimAppPkg='$(SIMAPP)' -ExitOnFail 30 3 'TestAppImportExport'

test-sim-after-import: runsim
	@echo "Running application simulation-after-import. This may take several minutes..."
	$(BINDIR)/runsim -Jobs=4 -SimAppPkg='$(SIMAPP)' -ExitOnFail 30 3 'TestAppSimulationAfterImport'

test-sim-custom-genesis-multi-seed: runsim
	@echo "Running multi-seed custom genesis simulation..."
	$(BINDIR)/runsim -Genesis='$(SIM_GENESIS)' -SimAppPkg='$(SIMAPP)' -ExitOnFail 400 5 'TestFullAppSimulation'

test-sim-multi-seed-long: runsim
	@echo "Running long multi-seed application simulation. This may take awhile!"
	$(BINDIR)/runsim -Jobs=4 -SimAppPkg='$(SIMAPP)' -ExitOnFail 500 50 'TestFullAppSimulation'

test-sim-multi-seed-short: runsim
	@echo "Running short multi-seed application simulation. This may take awhile!"
	$(BINDIR)/runsim -Jobs=4 -SimAppPkg='$(SIMAPP)' -ExitOnFail 50 10 'TestFullAppSimulation'

test-sim-nondeterminism:
	@echo "Running non-determinism test..."
	$(GO) test -mod=readonly $(SIMAPP) -run TestAppStateDeterminism -Enabled=true \
		-NumBlocks=50 -BlockSize=100 -Commit=true -Period=0 -v -timeout 24h

test-sim-custom-genesis-fast:
	@echo "Running custom genesis simulation..."
	$(GO) test -mod=readonly $(SIMAPP) -run TestFullAppSimulation -Genesis='$(SIM_GENESIS)' \
		-Enabled=true -NumBlocks=50 -BlockSize=100 -Commit=true -Seed=99 -Period=5 -v -timeout 24h

test-sim-simple:
	@echo "Running simple module simulation..."
	$(GO) test -mod=readonly $(SIMAPP) -run TestSimple \
		-Enabled=true -NumBlocks=50 -BlockSize=100 -Commit=true -Seed=99 -Period=5 -v -timeout 1h

test-sim-benchmark-invariants:
	@echo "Running simulation invariant benchmarks..."
	$(GO) test -mod=readonly -run=^$$ $(SIMAPP) -benchmem -bench=BenchmarkInvariants \
	-Enabled=true -NumBlocks=1000 -BlockSize=200 -Period=1 -Commit=true -Seed=57 -v -timeout 24h

SIM_NUM_BLOCKS ?= 500
SIM_BLOCK_SIZE ?= 200
SIM_COMMIT ?= true

test-sim-benchmark:
	@echo "Running application benchmark for numBlocks=$(SIM_NUM_BLOCKS), blockSize=$(SIM_BLOCK_SIZE). This may take awhile!"
	$(GO) test -mod=readonly -benchmem -run=^$$ $(SIMAPP) -bench ^BenchmarkFullAppSimulation$$ \
		-Enabled=true -NumBlocks=$(SIM_NUM_BLOCKS) -BlockSize=$(SIM_BLOCK_SIZE) -Commit=$(SIM_COMMIT) -timeout 24h

# Same as test-sim-benchmark except also creates files with cpu and memory profile info.
test-sim-profile:
	@echo "Running application benchmark for numBlocks=$(SIM_NUM_BLOCKS), blockSize=$(SIM_BLOCK_SIZE). This may take awhile!"
	$(GO) test -mod=readonly -benchmem -run=^$$ $(SIMAPP) -bench ^BenchmarkFullAppSimulation$$ \
		-Enabled=true -NumBlocks=$(SIM_NUM_BLOCKS) -BlockSize=$(SIM_BLOCK_SIZE) -Commit=$(SIM_COMMIT) -timeout 24h -cpuprofile cpu.out -memprofile mem.out

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
