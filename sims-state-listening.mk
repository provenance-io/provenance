#!/usr/bin/make -f

################################################
# Simulation tests with State Listening plugins
#
# This file is an extension for sims.mk
################################################

test-sim-nondeterminism-state-listening-trace:
	@echo "Running non-determinism-state-listening-trace test..."
	go test -mod=readonly $(SIMAPP) -run TestAppStateDeterminismWithStateListeningTrace -Enabled=true \
		-NumBlocks=50 -BlockSize=100 -Commit=true -Period=0 -v -test.v -timeout 24h;

test-sim-nondeterminism-state-listening-all: \
	test-sim-nondeterminism-state-listening-trace

.PHONY: \
test-sim-nondeterminism-state-listening-all \
test-sim-nondeterminism-state-listening-trace
