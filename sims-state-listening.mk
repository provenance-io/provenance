#!/usr/bin/make -f

################################################
# Simulation tests with State Listening plugins
#
# This file is an extension for sims.mk
################################################


test-sim-nondeterminism-state-listening-file:
	@echo "Running non-determinism-state-listening-file test..."
	go test -mod=readonly $(SIMAPP) -run TestAppStateDeterminismWithStateListening -Enabled=true \
		-NumBlocks=50 -BlockSize=100 -Commit=true -Period=0 -v -timeout 24h \
		-StateListeningPlugin=file -HaltAppOnDeliveryError=true

test-sim-nondeterminism-state-listening-trace:
	@echo "Running non-determinism-state-listening-trace test..."
	go test -mod=readonly $(SIMAPP) -run TestAppStateDeterminismWithStateListening -Enabled=true \
		-NumBlocks=50 -BlockSize=100 -Commit=true -Period=0 -v -timeout 24h \
		-StateListeningPlugin=trace -HaltAppOnDeliveryError=true

SIM_DOCKER_COMPOSE_YML ?= vendor/github.com/cosmos/cosmos-sdk/plugin/plugins/kafka/docker-compose.yml

test-sim-nondeterminism-state-listening-kafka: vendor
	# This is done as a single command for the following reasons:
	# - I want the exit code to be that of the test and not the stopping of kafka.
	# - I want to fail early if kafka can't be started.
	# - I want to stop kafka command to run regardless of the exit code of the test.
	# - By default make runs each command in a new shell. That makes it impossible to store
	#     the exit code of the test command for use after the kafka command if they're not all one command.
	# I've left of the @ so that we can see the entire go test command and the docker commands.
	# They end up sepparated from where they're executed, but the echos make them identifiable in the output.
	echo "Running non-determinism-state-listening-kafka test..."; \
		echo "Starting Kafka..."; \
		docker-compose -f $(SIM_DOCKER_COMPOSE_YML) up -d zookeeper broker || exit $$?; \
		echo "Running test..."; \
		go test -mod=readonly $(SIMAPP) -run TestAppStateDeterminismWithStateListening -Enabled=true \
			-NumBlocks=50 -BlockSize=100 -Commit=true -Period=0 -v -timeout 24h \
			-StateListeningPlugin=kafka -HaltAppOnDeliveryError=false; \
		ec=$$?; \
		echo "test exited with code '$$ec'"; \
		echo "Stopping Kafka..."; \
		docker-compose -f $(SIM_DOCKER_COMPOSE_YML) down; \
		exit $$ec;

test-sim-nondeterminism-state-listening-all: \
	test-sim-nondeterminism-state-listening-file \
	test-sim-nondeterminism-state-listening-trace \
	test-sim-nondeterminism-state-listening-kafka

.PHONY: \
test-sim-nondeterminism-state-listening-all \
test-sim-nondeterminism-state-listening-file \
test-sim-nondeterminism-state-listening-trace \
test-sim-nondeterminism-state-listening-kafka
