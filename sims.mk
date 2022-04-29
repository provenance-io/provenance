#!/usr/bin/make -f

########################################
### Simulations
###
### Several of these are used in .github/workflows/sims.yml.
### The strings in there leave off the "test-sim" prefix.
###
### Environment Variables:
###   DB_BACKEND:     dictates which db backend to use: goleveldb, cleveldb, rocksdb, badgerdb.
###                   The test-sim-nondeterminism is hard-coded to use memdb though.
###   BINDIR:         The Go bin directory, defaults to $GOPATH/bin
###   SIM_GENESIS:    Defines the path to the custom genesis file used by
###                   test-sim-custom-genesis-multi-seed and test-sim-custom-genesis-fast
###                   Default is ${HOME}/.provenanced/config/genesis.json.
###   SIM_NUM_BLOCKS: The number of blocks to use for test-sim-benchmark or test-sim-profile. Default is 500.
###   SIM_BLOCK_SIZE: The size of blocks to use for test-sim-benchmark or test-sim-profile. Default is 200.
###   SIM_COMMIT:     Whether to commit during  test-sim-benchmark or test-sim-profile. Default is true.

BINDIR ?= $(GOPATH)/bin
SIMAPP = ./app
DB_BACKEND ?= goleveldb
ifeq ($(DB_BACKEND),cleveldb)
  tags = cleveldb
else ifeq ($(DB_BACKEND),rocksdb)
  tags = rocksdb
else ifeq ($(DB_BACKEND),badgerdb)
  tags = badgerdb
else ifneq ($(DB_BACKEND),goleveldb)
  $(error unknown DB_BACKEND value [$(DB_BACKEND)]. Must be one of goleveldb, cleveldb, rocksdb, badgerdb)
endif
# For now, the M1 macs need the dynamic tag to run at all.
# But we can only provide one tag to the runsim tests (see below).
# So if on an M1, only use goleveldb for the runsim tests.
# The non-runsim tests can still use any db backend.
RUNSIM_DB_BACKEND = $(DB_BACKEND)
ifeq ($(UNAME_S),darwin)
  ifeq ($(UNAME_M),arm64)
    # Needed on M1 macs due to kafka issue: https://github.com/confluentinc/confluent-kafka-go/issues/591#issuecomment-811705552
    # Once we no longer need the dynamic flag, a lot of this can be cleaned up so that all these tests use the same tag (if any).
    tags += dynamic
    runsim_tag = dynamic
    RUNSIM_DB_BACKEND = goleveldb
  endif
endif
ifeq ($(runsim_tag),)
  runsim_tag = $(tags)
endif

# We have to use a hack to provide -tags with the runsim stuff, but it only allows us to provide one tag.
# Runsim creates a command string, then does a split on " " to turn it into args.
# With two tags, e.g. -tags 'foo bar', you'd end up with three args, "-tags", "'foo", and "bar'", and it'll get confused.
# But we CAN provide a single tag in the -SimAppPkg value in order to trick it into including it in the `go test` commands.
# We need to provide the -DBBackend flag to the runsim tests too, and use the same hack.
SIM_APP_PKG := $(SIMAPP) -DBBackend=$(RUNSIM_DB_BACKEND)
ifneq ($(runsim_tag),)
  SIM_APP_PKG += -tags $(runsim_tag)
endif
SIMAPP += -DBBackend=$(DB_BACKEND) -tags '$(tags)'

SIM_GENESIS ?= ${HOME}/.provenanced/config/genesis.json

include sims-state-listening.mk

# Runsim Usage: runsim [flags] <blocks> <period> <testname>
# flags: [-Jobs maxprocs] [-ExitOnFail] [-Seeds comma-separated-seed-list]
#        [-Genesis file-path] [-SimAppPkg file-path] [-Github] [-Slack] [-LogObjPrefix string]

# A target with the same name is also defined in contrib/devtools/Makefile
# That one makes sure runsim is installed. This one just outputs an alert if not using the requested db backend (M1 issue).
# If you get rid of this one, you'll still want the targets below to depend on the runsim target.
runsim:
	@if [ '$(DB_BACKEND)' != '$(RUNSIM_DB_BACKEND)' ]; then printf '\n\033[93;100m Alert \033[0m Using $(RUNSIM_DB_BACKEND) instead of $(DB_BACKEND) due to M1 issue and limitations of runsim\n\n'; fi

test-sim-import-export: runsim
	@echo "Running application import/export simulation. This may take several minutes..."
	$(BINDIR)/runsim -Jobs=4 -SimAppPkg='$(SIM_APP_PKG)' -ExitOnFail 30 3 'TestAppImportExport'

test-sim-after-import: runsim
	@echo "Running application simulation-after-import. This may take several minutes..."
	$(BINDIR)/runsim -Jobs=4 -SimAppPkg='$(SIM_APP_PKG)' -ExitOnFail 30 3 'TestAppSimulationAfterImport'

test-sim-custom-genesis-multi-seed: runsim
	@echo "Running multi-seed custom genesis simulation..."
	$(BINDIR)/runsim -Genesis='$(SIM_GENESIS)' -SimAppPkg='$(SIM_APP_PKG)' -ExitOnFail 400 5 'TestFullAppSimulation'

test-sim-multi-seed-long: runsim
	@echo "Running long multi-seed application simulation. This may take awhile!"
	$(BINDIR)/runsim -Jobs=4 -SimAppPkg='$(SIM_APP_PKG)' -ExitOnFail 500 50 'TestFullAppSimulation'

test-sim-multi-seed-short: runsim
	@echo "Running short multi-seed application simulation. This may take awhile!"
	$(BINDIR)/runsim -Jobs=4 -SimAppPkg='$(SIM_APP_PKG)' -ExitOnFail 50 10 'TestFullAppSimulation'

test-sim-nondeterminism:
	@echo "Running non-determinism test..."
	go test -mod=readonly $(SIMAPP) -run TestAppStateDeterminism -Enabled=true \
		-NumBlocks=50 -BlockSize=100 -Commit=true -Period=0 -v -timeout 24h

test-sim-custom-genesis-fast:
	@echo "Running custom genesis simulation..."
	go test -mod=readonly $(SIMAPP) -run TestFullAppSimulation -Genesis='$(SIM_GENESIS)' \
		-Enabled=true -NumBlocks=50 -BlockSize=100 -Commit=true -Seed=99 -Period=5 -v -timeout 24h

test-sim-simple:
	@echo "Running simple module simulation..."
	go test -mod=readonly $(SIMAPP) -run TestSimple \
		-Enabled=true -NumBlocks=50 -BlockSize=100 -Commit=true -Seed=99 -Period=5 -v -timeout 1h

test-sim-benchmark-invariants:
	@echo "Running simulation invariant benchmarks..."
	go test -mod=readonly -run=^$$ $(SIMAPP) -benchmem -bench=BenchmarkInvariants \
	-Enabled=true -NumBlocks=1000 -BlockSize=200 -Period=1 -Commit=true -Seed=57 -v -timeout 24h

SIM_NUM_BLOCKS ?= 500
SIM_BLOCK_SIZE ?= 200
SIM_COMMIT ?= true

test-sim-benchmark:
	@echo "Running application benchmark for numBlocks=$(SIM_NUM_BLOCKS), blockSize=$(SIM_BLOCK_SIZE). This may take awhile!"
	go test -mod=readonly -benchmem -run=^$$ $(SIMAPP) -bench ^BenchmarkFullAppSimulation$$ \
		-Enabled=true -NumBlocks=$(SIM_NUM_BLOCKS) -BlockSize=$(SIM_BLOCK_SIZE) -Commit=$(SIM_COMMIT) -timeout 24h

# Same as test-sim-benchmark except also creates files with cpu and memory profile info.
test-sim-profile:
	@echo "Running application benchmark for numBlocks=$(SIM_NUM_BLOCKS), blockSize=$(SIM_BLOCK_SIZE). This may take awhile!"
	go test -mod=readonly -benchmem -run=^$$ $(SIMAPP) -bench ^BenchmarkFullAppSimulation$$ \
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
