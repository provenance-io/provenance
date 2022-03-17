#!/usr/bin/make -f

########################################
### Simulations

BINDIR ?= $(GOPATH)/bin
SIMAPP = ./app
DB_BACKEND ?= goleveldb
ifeq ($(DB_BACKEND),cleveldb)
  tags = -tags cleveldb
else ifeq ($(DB_BACKEND),rocksdb)
  tags = -tags rocksdb
else ifeq ($(DB_BACKEND),badgerdb)
  tags = -tags badgerdb
else ifeq ($(DB_BACKEND),boltdb)
  tags = -tags boltdb
else ifneq ($(DB_BACKEND),goleveldb)
  $(error unknown DB_BACKEND value [$(DB_BACKEND)]. Must be one of goleveldb, cleveldb, rocksdb, badgerdb, boltdb)
endif

# Bit of a hack for the runsim stuff that also works with the go test stuff.
# Basically, To test other databases, we need the -tags in the go test command (as well as the -DBBackend= flag).
# runsim takes the SimAppPkg value and supplies it to go test similar to how we do in here.
# However, runsim creates a command string, then does a split on " " to get the args.
# So this trick only works with a single tag, and the tags value needs to not be quoted.
# A similar hack is used on the test name to get the -DBBackend flag set for the test.
ifneq ($(tags),)
  SIMAPP := $(tags) $(SIMAPP)
endif

test-sim-nondeterminism:
	@echo "Running non-determinism test..."
	go test -mod=readonly $(SIMAPP) -run TestAppStateDeterminism -Enabled=true -DBBackend=$(DB_BACKEND) \
		-NumBlocks=50 -BlockSize=100 -Commit=true -Period=0 -v -timeout 24h

test-sim-custom-genesis-fast:
	@echo "Running custom genesis simulation..."
	@echo "By default, ${HOME}/.provenanced/config/genesis.json will be used."
	go test -mod=readonly $(SIMAPP) -run TestFullAppSimulation -Genesis=${HOME}/.provenanced/config/genesis.json -DBBackend=$(DB_BACKEND) \
		-Enabled=true -NumBlocks=50 -BlockSize=100 -Commit=true -Seed=99 -Period=5 -v -timeout 24h

test-sim-simple:
	@echo "Running simple module simulation..."
	go test -mod=readonly $(SIMAPP) -run TestSimple -DBBackend=$(DB_BACKEND) \
		-Enabled=true -NumBlocks=50 -BlockSize=100 -Commit=true -Seed=99 -Period=5 -v -timeout 1h

test-sim-import-export: runsim
	@echo "Running application import/export simulation. This may take several minutes..."
	$(BINDIR)/runsim -Jobs=4 -SimAppPkg='$(SIMAPP)' -ExitOnFail 30 3 'TestAppImportExport -DBBackend=$(DB_BACKEND)'

test-sim-after-import: runsim
	@echo "Running application simulation-after-import. This may take several minutes..."
	$(BINDIR)/runsim -Jobs=4 -SimAppPkg='$(SIMAPP)' -ExitOnFail 30 3 'TestAppSimulationAfterImport -DBBackend=$(DB_BACKEND)'

test-sim-custom-genesis-multi-seed: runsim
	@echo "Running multi-seed custom genesis simulation..."
	@echo "By default, ${HOME}/.provenanced/config/genesis.json will be used."
	$(BINDIR)/runsim -Genesis=${HOME}/.provenanced/config/genesis.json -SimAppPkg='$(SIMAPP)' -ExitOnFail 400 5 'TestFullAppSimulation -DBBackend=$(DB_BACKEND)'

test-sim-multi-seed-long: runsim
	@echo "Running long multi-seed application simulation. This may take awhile!"
	$(BINDIR)/runsim -Jobs=4 -SimAppPkg='$(SIMAPP)' -ExitOnFail 500 50 'TestFullAppSimulation -DBBackend=$(DB_BACKEND)'

test-sim-multi-seed-short: runsim
	@echo "Running short multi-seed application simulation. This may take awhile!"
	$(BINDIR)/runsim -Jobs=4 -SimAppPkg='$(SIMAPP)' -ExitOnFail 50 10 'TestFullAppSimulation -DBBackend=$(DB_BACKEND)'

test-sim-benchmark-invariants:
	@echo "Running simulation invariant benchmarks..."
	go test -mod=readonly -run=^$$ $(SIMAPP) -benchmem -bench=BenchmarkInvariants -DBBackend=$(DB_BACKEND) \
	-Enabled=true -NumBlocks=1000 -BlockSize=200 -Period=1 -Commit=true -Seed=57 -v -timeout 24h

.PHONY: \
test-sim-nondeterminism \
test-sim-custom-genesis-fast \
test-sim-import-export \
test-sim-after-import \
test-sim-custom-genesis-multi-seed \
test-sim-multi-seed-short \
test-sim-multi-seed-long \
test-sim-benchmark-invariants

SIM_NUM_BLOCKS ?= 500
SIM_BLOCK_SIZE ?= 200
SIM_COMMIT ?= true

test-sim-benchmark:
	@echo "Running application benchmark for numBlocks=$(SIM_NUM_BLOCKS), blockSize=$(SIM_BLOCK_SIZE). This may take awhile!"
	go test -mod=readonly -benchmem -run=^$$ $(SIMAPP) -bench ^BenchmarkFullAppSimulation$$ -DBBackend=$(DB_BACKEND)  \
		-Enabled=true -NumBlocks=$(SIM_NUM_BLOCKS) -BlockSize=$(SIM_BLOCK_SIZE) -Commit=$(SIM_COMMIT) -timeout 24h

test-sim-profile:
	@echo "Running application benchmark for numBlocks=$(SIM_NUM_BLOCKS), blockSize=$(SIM_BLOCK_SIZE). This may take awhile!"
	go test -mod=readonly -benchmem -run=^$$ $(SIMAPP) -bench ^BenchmarkFullAppSimulation$$ -DBBackend=$(DB_BACKEND) \
		-Enabled=true -NumBlocks=$(SIM_NUM_BLOCKS) -BlockSize=$(SIM_BLOCK_SIZE) -Commit=$(SIM_COMMIT) -timeout 24h -cpuprofile cpu.out -memprofile mem.out

.PHONY: test-sim-profile test-sim-benchmark