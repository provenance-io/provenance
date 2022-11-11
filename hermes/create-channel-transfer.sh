#!/bin/bash
set -e

# Load shell variables
. ./variables.sh

### Configure the clients and connection
echo "Initiating connection handshake..."
$HERMES_BINARY -c $CONFIG_DIR create channel testing  connection-0 --port-a  transfer  --port-b  transfer

sleep 2
