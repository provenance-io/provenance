#!/bin/bash
set -e

# Load shell variables
. ./variables.sh

### Configure the clients and connection
echo "Initiating connection handshake..."
$HERMES_BINARY -c $CONFIG_DIR create channel --port-a "wasm.tp14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9s96lrg8" --port-b "wasm.tp14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9s96lrg8" testing connection-0 --order ordered --channel-version "ibc-reflect-v1"

sleep 2
