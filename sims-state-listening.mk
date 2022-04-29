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
	@echo "Running non-determinism-state-listening-kafka test..."
	@echo "Starting Kafka..."
	docker-compose -f $(SIM_DOCKER_COMPOSE_YML) up -d zookeeper broker
	@echo "Running test..."
	-go test -mod=readonly $(SIMAPP) -run TestAppStateDeterminismWithStateListening -Enabled=true \
		-NumBlocks=50 -BlockSize=100 -Commit=true -Period=0 -v -timeout 24h \
		-StateListeningPlugin=kafka -HaltAppOnDeliveryError=false
	@echo "Stopping Kafka..."
	-docker-compose -f plugin/plugins/kafka/docker-compose.yml down

test-sim-nondeterminism-state-listening-all: \
	test-sim-nondeterminism-state-listening-file \
	test-sim-nondeterminism-state-listening-trace \
	test-sim-nondeterminism-state-listening-kafka

.PHONY: \
test-sim-nondeterminism-state-listening-all \
test-sim-nondeterminism-state-listening-file \
test-sim-nondeterminism-state-listening-trace \
test-sim-nondeterminism-state-listening-kafka
